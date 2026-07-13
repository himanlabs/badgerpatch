package badgerpatch

import (
	"fmt"
	"reflect"
	"sync"
)

type patch struct {
	originalBytes []byte
}

var (
	lock    sync.Mutex
	patches = make(map[uintptr]patch)
)

type PatchGuard struct {
	target uintptr
}

func (g *PatchGuard) Unpatch() {
	lock.Lock()
	defer lock.Unlock()

	if p, ok := patches[g.target]; ok {
		copyToLocation(g.target, p.originalBytes)
		delete(patches, g.target)
	}
}

func Patch[T any](target T, replacement T) *PatchGuard {
	targetVal := reflect.ValueOf(target)
	replaceVal := reflect.ValueOf(replacement)

	if targetVal.Kind() != reflect.Func || replaceVal.Kind() != reflect.Func {
		panic("badgerpatch: target and replacement must be functions")
	}

	targetAddr := targetVal.Pointer()
	replaceAddr := replaceVal.Pointer()

	lock.Lock()
	defer lock.Unlock()

	if p, ok := patches[targetAddr]; ok {
		copyToLocation(targetAddr, p.originalBytes)
	}

	// buildJmpDirective is defined in architecture-specific files
	jmpCode := buildJmpDirective(replaceAddr)

	originalBytes := make([]byte, len(jmpCode))
	copy(originalBytes, rawMemoryAccess(targetAddr, len(jmpCode)))

	copyToLocation(targetAddr, jmpCode)

	// Verify the write actually landed. copyToLocation panics on OS-level
	// mprotect/VirtualProtect failures, but this catches any other
	// discrepancy (e.g. a concurrent writer) instead of returning a guard
	// that silently didn't patch anything.
	written := rawMemoryAccess(targetAddr, len(jmpCode))
	for i := range jmpCode {
		if written[i] != jmpCode[i] {
			panic(fmt.Sprintf("badgerpatch: patch verification failed at offset %d (wrote %#x, read back %#x) - target may have been modified concurrently", i, jmpCode[i], written[i]))
		}
	}

	patches[targetAddr] = patch{originalBytes: originalBytes}

	return &PatchGuard{target: targetAddr}
}