package badgerpatch

// Patches holds a group of applied mocks
type Patches struct {
	guards []*PatchGuard
}

func NewPatches() *Patches {
	return &Patches{guards: make([]*PatchGuard, 0)}
}

func (p *Patches) Reset() {
	for _, guard := range p.guards {
		guard.Unpatch()
	}
	p.guards = nil
}

// ApplyFunc initializes a registry and applies the first mock
func ApplyFunc[T any](target T, double T) *Patches {
	p := NewPatches()
	return Apply(p, target, double)
}

// Apply chains additional mocks to the existing registry
func Apply[T any](p *Patches, target T, double T) *Patches {
	guard := Patch(target, double)
	p.guards = append(p.guards, guard)
	return p
}