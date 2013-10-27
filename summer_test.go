package summer

import "testing"

func TestSimpleInject(t *testing.T) {
	type simpleStruct struct {
		MyString string `summer:"StringDependency"`
	}
	container := NewContainer()
	container.Add("injected", "StringDependency")
	s := new(simpleStruct)
	err := container.InjectInto(s)

	if s.MyString != "injected" || err != nil {
		t.Fail()
	}
}

func TestThrowsErrorForMissingDependency(t *testing.T) {
	type simpleStruct struct {
		AvailableString   string `summer:"HasIt"`
		UnavailableString string `summer:"Doesn'tHaveIt"`
	}
	container := NewContainer()
	container.Add("ok", "HasIt")
	s := new(simpleStruct)
	err := container.InjectInto(s)

	if err == nil {
		t.Fail()
	}
}

func TestThrowsErrorWhenAttemptInjectNonStruct(t *testing.T) {
	container := NewContainer()
	x := 4
	err := container.InjectInto(&x)

	if err == nil {
		t.Fail()
	}
}

type hookTestingStruct struct {
	called bool
}

func (h *hookTestingStruct) PostInjectionCallback() {
	h.called = true
}

func TestCallsPostInjectionHook(t *testing.T) {
	s := &hookTestingStruct{called: false}
	container := NewContainer()
	container.InjectInto(s)

	if !s.called {
		t.Fail()
	}
}

func TestParsesEmptyFieldTag(t *testing.T) {
	tag := parseFieldTag("")
	if tag != nil {
		t.Fail()
	}
}

func TestParsesNamedFieldTag(t *testing.T) {
	tag := parseFieldTag("nameHere")
	if tag.dependencyName != "nameHere" {
		t.Fail()
	}

	tag = parseFieldTag("nameHere,")
	if tag.dependencyName != "nameHere" {
		t.Fail()
	}
}

func TestParsesAutoInjectFieldTag(t *testing.T) {
	tag := parseFieldTag(",auto")
	if !tag.autoInject {
		t.Fail()
	}

	tag = parseFieldTag("name,auto")
	if !tag.autoInject {
		t.Fail()
	}
}

func TestAutoInjects(t *testing.T) {
	type injectedStruct struct {
		Nothing string
	}
	ij := new(injectedStruct)

	type simpleStruct struct {
		AutoString    string          `summer:",auto"`
		AutoStruct    *injectedStruct `summer:",auto"`
		AutoStructTwo injectedStruct  `summer:",auto"`
	}

	container := NewContainer()
	container.Add("ok", "")
	container.Add(ij, "")
	container.Add(*ij, "")
	s := new(simpleStruct)
	err := container.InjectInto(s)

	if err != nil || s.AutoString != "ok" || s.AutoStruct != ij {
		t.Log(err)
		t.Fail()
	}
}

func TestThrowsErrorOnMissingAutoInject(t *testing.T) {
	type simpleStruct struct {
		AutoString string `summer:",auto"`
	}
	s := new(simpleStruct)

	container := NewContainer()
	err := container.InjectInto(s)

	if err == nil {
		t.Fail()
	}
}

func TestPerformsInjections(t *testing.T) {
	type simpleStructOne struct {
		AutoString string `summer:",auto"`
	}

	type simpleStructTwo struct {
		AutoStructOne *simpleStructOne `summer:",auto"`
	}

	s1 := new(simpleStructOne)
	s2 := new(simpleStructTwo)
	h := new(hookTestingStruct)

	container := NewContainer()
	container.Add(s1, "")
	container.Add(s2, "")
	container.Add(h, "")
	container.Add("here", "")
	err := container.PerformInjections()

	if err != nil || s1.AutoString != "here" || s2.AutoStructOne != s1 || !h.called {
		t.Log(err)
		t.Fail()
	}
}

type circularStructOne struct {
	Two *circularStructTwo `summer:"2"`
}

type circularStructTwo struct {
	One *circularStructOne `summer:"1"`
}

func TestHandlesCircularDependencies(t *testing.T) {
	s1 := new(circularStructOne)
	s2 := new(circularStructTwo)

	container := NewContainer()
	container.Add(s1, "1")
	container.Add(s2, "2")
	err := container.PerformInjections()

	if err != nil || s1.Two != s2 || s2.One != s1 {
		t.Log(err)
		t.Fail()
	}
}

func TestGet(t *testing.T) {
	container := NewContainer()
	container.Add("value", "nameHere")

	value, ok := container.Get("nameHere")

	if !ok || value != "value" {
		t.Fail()
	}
}

func TestSkipsHooksOnFailedInjections(t *testing.T) {
	type missingStruct struct {
		Missing string `summer:"Missing"`
	}
	h := new(hookTestingStruct)

	container := NewContainer()
	container.Add(new(missingStruct), "")
	container.Add(h, "")
	err := container.PerformInjections()

	t.Log(err)
	if err == nil || h.called {
		t.Fail()
	}
}

func ExampleContainer_PerformInjections() {
	// All structs are set up similar to the example for InjectInto.
	type ServiceOne struct {
		StringDependency string `summer:"GiveThisString"`
	}

	type ServiceTwo struct {
		ServiceOneDependencyHere *ServiceOne `summer:",auto"`
		MyString                 string      `summer:",auto"`
	}

	// Initialize your structs however you want..
	serviceOne := new(ServiceOne)
	serviceTwo := new(ServiceTwo)

	// All injection targets and dependencies (in this case,
	// they're the same) are added in to the container.
	container := NewContainer()
	container.Add("Only string in the box", "GiveThisString")
	container.Add(serviceOne, "")
	container.Add(serviceTwo, "")

	// PerformInjections will abort if any of the injections fail,
	// and will run post-injection hooks only after all injections have
	// completed successfully.
	_ = container.PerformInjections()
}

func ExampleContainer_InjectInto() {
	// Tag any fields you want to be injected with
	// the summer tag and name of the dependency.
	// The suffix ",auto" can be used to inject by matching type.
	// Fields without tags are ignored.
	type MyService struct {
		MyDependency     string `summer:"StringDependency"`
		MyAutoDependency string `summer:",auto"`
	}
	service := new(MyService)

	// Set up your container with the Add method. The name
	// argument can be left blank if it should only be injectable by
	// type matching.
	container := NewContainer()
	container.Add("injected value", "StringDependency")

	// Check all of the struct's fields for the annotations, matching
	// them with the container's map. If any dependencies tagged in the
	// struct aren't found in the map, an error is thrown and the rest
	// of the injection process is aborted.
	_ = container.InjectInto(service)
}

func ExampleContainer_Get() {
	type MyDependency struct {
		// ...
	}

	container := NewContainer()
	container.Add(new(MyDependency), "SomeName")

	// Retrieves the named dependency from the container. Second return value
	// can be used for error-checking.
	myDependency, _ := container.Get("SomeName")

	var _ = myDependency // Ignore this
}
