// Simple DI framework for propogating dependencies across many services
// in an attempt to avoid excess boilerplate. See the Container.InjectInto example for usage.
package summer

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	summerTag     = "summer"
	tagAutoInject = "auto"
)

type PostInjector interface {
	// If your injection target conforms to this interface, Summer
	// will call this hook after injection takes place.
	PostInjectionCallback()
}

// The dependency injection container, where your dependencies can be
// named and then injected into your service's structs. This should always
// be instantiated with NewContainer.
type Container struct {
	// Holds references to all of your dependencies
	// indexed by name. Used for injection by specific name.
	dependenciesByName map[string]interface{}

	// Holds references to all of your dependencies
	// indexed by type. Used for auto injection by type.
	dependenciesByType map[reflect.Type]interface{}

	// Set of references to each dependency
	possibleInjectionSet *interfaceSet
}

func NewContainer() *Container {
	return &Container{
		dependenciesByName:   make(map[string]interface{}),
		dependenciesByType:   make(map[reflect.Type]interface{}),
		possibleInjectionSet: newInterfaceSet(),
	}
}

// Adds a new dependency to the container. The name parameter can be left
// blank, signifying that the dependency cannot be referenced explictly by
// name (and instead should be injected with the automatic mode, by type).
//
// All types that will be injected into a field expecting an interface (and not
// a pointer to a concrete struct) should be added to the container with an
// explicit name, as Summer cannot automatically inject by interface (you don't
// want to do this and cannot do this anyways, since your structs could
// implement many interfaces you're unaware of)
func (c *Container) Add(target interface{}, name string) {
	if name != "" {
		c.dependenciesByName[name] = target
	}

	// Last dependency of a specific type always takes precedence
	c.dependenciesByType[reflect.TypeOf(target)] = target

	// All unique dependencies added once, if they're injectable
	if isPointerToStruct(target) {
		c.possibleInjectionSet.Add(target)
	}
}

// Injects dependencies for every struct that has been
// added to the container. Operates as if InjectInto was called for
// all objects, with the callbacks ran after all injections take place.
//
// Errors returned are identical to InjectInto's errors.
func (c *Container) PerformInjections() error {
	var err error = nil

	c.possibleInjectionSet.EachElement(func(key interface{}) {
		if err == nil {
			if err = c.realInjectInto(key, false); err != nil {
				return
			}
		}
	})

	// Run hooks after *all* dependencies are injected successfully
	if err == nil {
		c.possibleInjectionSet.EachElement(func(key interface{}) {
			performPostInjectionHook(key)
		})
	}

	return err
}

// Injects the container's stored dependencies into the
// target implementation by examining the target's struct tags.
//
// If the target implements the PostInjector interface, post
// injection hooks are called after a successful injection.
//
// An error is returned if one of the tagged fields requests
// a named dependency missing in the container.
// An error is returned if one of the tagged fields requests
// an automatic injection and no matching type is present in the container.
// An error is also returned if the target interface{} is not
// a struct.
func (c *Container) InjectInto(target interface{}) error {
	return c.realInjectInto(target, true)
}

// Actual implementation of InjectInto. Subject to change.
func (c *Container) realInjectInto(target interface{}, performHook bool) error {
	if ok := isPointerToStruct(target); !ok {
		return errors.New("Summer: Attempted to inject into something other than a pointer-to-struct")
	}

	err := iterateFields(target, c.performInjection)
	if err != nil {
		return err
	}
	if performHook {
		performPostInjectionHook(target)
	}

	return nil
}

// Returns a named dependency from the container.
//
// When the dependency is missing from the container, the second return value
// is false.
func (c *Container) Get(name string) (interface{}, bool) {
	if dependency, ok := c.dependenciesByName[name]; ok {
		return dependency, true
	}

	return nil, false
}

func performPostInjectionHook(target interface{}) {
	if _, ok := target.(PostInjector); ok {
		target.(PostInjector).PostInjectionCallback()
	}
}

// Gets the type of target after a dereference (if necessary)
func getDereferencedType(target interface{}) reflect.Type {
	targetType := reflect.TypeOf(target)

	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	return targetType
}

func isPointerToStruct(target interface{}) bool {
	targetType := getDereferencedType(target)
	return (targetType.Kind() == reflect.Struct)
}

// struct to hold the sprawling number of arguments passed around for injection
type injectionPoint struct {
	elementType reflect.Type        // The type of the struct we're injecting into
	field       reflect.Value       // The specific instance of the struct's field we're setting
	typeField   reflect.StructField // The type's description of the field
}

// the field tag is parsed into this struct
type fieldTag struct {
	dependencyName string
	autoInject     bool
}

// Format: `summer:"dependencyName,[autoInject]"`
func parseFieldTag(rawTag string) *fieldTag {
	if rawTag == "" {
		return nil
	}

	components := strings.Split(rawTag, ",")
	shouldAutoInject := false

	if len(components) > 1 {
		shouldAutoInject = (components[1] == tagAutoInject)
	}

	return &fieldTag{
		dependencyName: components[0],
		autoInject:     shouldAutoInject,
	}
}

func (c *Container) performNamedInjection(p injectionPoint, dependencyName string) error {
	if dependency, ok := c.dependenciesByName[dependencyName]; ok {
		p.field.Set(reflect.ValueOf(dependency))
	} else {
		return errors.New(
			fmt.Sprintf("Summer: Missing required dependency %s for %s's field %s",
				dependencyName, p.elementType, p.typeField.Name))
	}

	return nil
}

func (c *Container) performAutoInjection(p injectionPoint) error {
	matchingType := p.typeField.Type
	if dependency, ok := c.dependenciesByType[matchingType]; ok {
		p.field.Set(reflect.ValueOf(dependency))
	} else {
		return errors.New(
			fmt.Sprintf("Summer: Missing autoinjected dependency %s's field %s"+
				", searched for type %s "+
				" (did you attempt to autoinject an interface?)",
				p.elementType, p.typeField.Name, matchingType))
	}

	return nil
}

func (c *Container) performInjection(p injectionPoint) error {
	tag := parseFieldTag(p.typeField.Tag.Get(summerTag))

	if tag != nil && p.field.CanSet() {
		if !tag.autoInject {
			return c.performNamedInjection(p, tag.dependencyName)
		} else {
			return c.performAutoInjection(p)
		}
	}

	return nil
}

// Iterate over all of the fields in the given (assumed) struct,
// calling the callback function for each one
func iterateFields(target interface{},
	callback func(p injectionPoint) error) error {
	element := reflect.ValueOf(target).Elem()
	elementType := element.Type()

	for index := 0; index < element.NumField(); index++ {
		ip := injectionPoint{
			field:       element.Field(index),
			typeField:   elementType.Field(index),
			elementType: elementType,
		}

		err := callback(ip)
		if err != nil {
			return err
		}
	}

	return nil
}
