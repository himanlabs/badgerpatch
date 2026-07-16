// Package badgerpatch rewrites a function's compiled machine code at runtime
// so calls to it jump to a replacement instead. This is a test-double /
// mocking technique, NOT a general-purpose production mechanism.
//
// IMPORTANT CAVEATS - read before using:
//
//  1. Inlining: if the Go compiler inlines a call to your target function
//     (common for small, simple functions), the call site never goes through
//     the function's machine code at all, so the patch has no effect at that
//     call site. Either:
//     - add `//go:noinline` above the target function's definition, or
//     - build/test with `-gcflags="all=-l"` to disable inlining globally.
//     Patch() cannot detect or fix this on its own since it only has
//     visibility into the target function's own bytes, not caller call sites.
//
//  2. Concurrency: patching overwrites several bytes of executable memory.
//     It is not atomic. Only call Patch/Unpatch when you can guarantee no
//     other goroutine is concurrently calling the target function (e.g.
//     during single-threaded test setup, before spawning workers).
//
//  3. Scope: this package is intended to be imported only from _test.go
//     files, to replace collaborators in unit tests. Do not wire it into
//     real request-handling / business logic paths.
package badgerpatch

import (
	"fmt"
	"reflect"
	"runtime"
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

// Patch forces compile-time type matching between target and replacement
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

	// buildJmpDirective is defined in architecture-specific files. `targetAddr`
	// is passed as `from` because on arm64 the size of the safest encoding
	// depends on the distance between target and replacement (see
	// jmp_arm64.go).
	jmpCode := buildJmpDirective(replaceAddr, targetAddr)

	// Safety check: refuse to write past the target function's own
	// boundary into whatever code follows it. Without this, a target
	// function shorter than the jump sequence gets silently corrupted at
	// its tail, along with whatever function happens to sit next in
	// memory - exactly what happened when a 16-byte function was patched
	// with what was then a 20-byte arm64 sequence (see CHANGELOG.md).
	// runtime.FuncForPC reports which Go function a given address belongs
	// to; if the last byte we're about to write lands in a *different*
	// function than the one we started at, the write would overflow.
	endAddr := targetAddr + uintptr(len(jmpCode)) - 1
	endFn := runtime.FuncForPC(endAddr)
	if endFn == nil || endFn.Entry() != targetAddr {
		neighbor := "an unknown/non-Go address"
		if endFn != nil {
			neighbor = endFn.Name()
		}
		panic(fmt.Sprintf(
			"badgerpatch: target function %s is too small to patch safely on this architecture (patch needs %d bytes, but the function ends before that, spilling into %s) - this function cannot be used as a Patch()/ApplyFunc() target",
			targetVal.Type().String(), len(jmpCode), neighbor,
		))
	}

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