package summer

// Type used for a set of dependencies pending injection, used
// to avoid circular dependency problems
type interfaceSet struct {
	set map[interface{}]bool
}

func newInterfaceSet() *interfaceSet {
	return &interfaceSet{set: make(map[interface{}]bool)}
}

// Returns false if the item was already part of the set, otherwise true
func (s *interfaceSet) Add(target interface{}) bool {
	_, present := s.set[target]
	s.set[target] = true
	return !present
}

func (s *interfaceSet) EachElement(callback func(key interface{})) {
	for key, _ := range s.set {
		callback(key)
	}
}
