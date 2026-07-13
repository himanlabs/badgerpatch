package badgerpatch

import (
	"errors"
	"sync"
	"testing"
)

// --- Target Functions ---
//
// //go:noinline is required on every function you intend to patch. See
// README.md "Required: mark target functions //go:noinline" — without it
// the compiler may inline the call site and the patch silently won't apply
// there. Every target below carries it deliberately.

//go:noinline
func FetchData() (string, error) {
	return "Real Network Data", nil
}

//go:noinline
func Add(a, b int) int {
	return a + b
}

//go:noinline
func GetGreeting() string {
	return "hello from the real implementation"
}

// =====================================================================
// WITHOUT sequence: a single static replacement, same result every call
// =====================================================================

// TestStaticReturn_SingleCall patches a function to always return one
// fixed value, no state involved. This is the simplest case: use it when
// the test doesn't care how many times the function is called, only that
// it's replaced.
func TestStaticReturn_SingleCall(t *testing.T) {
	guard := StaticReturn(Add, func(a, b int) int {
		return 999 // ignores inputs entirely - proves the real Add() never runs
	})
	defer guard.Unpatch()

	if got := Add(2, 2); got != 999 {
		t.Fatalf("expected patched value 999, got %d", got)
	}
	// Calling it again confirms it's a stable static override, not a
	// one-shot - unlike the sequence tests below, every call behaves
	// identically.
	if got := Add(10, 20); got != 999 {
		t.Fatalf("expected patched value 999 on second call, got %d", got)
	}
}

// TestStaticReturn_UnpatchRestoresOriginal is the same pattern but checks
// the other half of the lifecycle: after Unpatch(), the real function must
// run again exactly as before.
func TestStaticReturn_UnpatchRestoresOriginal(t *testing.T) {
	before := GetGreeting()
	if before != "hello from the real implementation" {
		t.Fatalf("sanity check failed before patching: got %q", before)
	}

	guard := StaticReturn(GetGreeting, func() string {
		return "stubbed greeting"
	})

	if got := GetGreeting(); got != "stubbed greeting" {
		t.Fatalf("expected stubbed value while patched, got %q", got)
	}

	guard.Unpatch()

	if got := GetGreeting(); got != before {
		t.Fatalf("expected original value restored after Unpatch, got %q", got)
	}
}

// =====================================================================
// WITH sequence: a different result on each successive call
// =====================================================================

// 1. Declare sequence state globally so the replacement closure captures no
// local variables (see comment in TestApplyFuncSeq below for why that
// matters).
var testSeq *SequenceBuilder[error]

// TestApplyFuncSeq is the retry-logic case: the same function needs to
// behave differently across successive calls (fail, fail, then succeed).
// A static replacement can't express this - only a sequence can.
func TestApplyFuncSeq(t *testing.T) {
	// 2. Initialize the global sequence inside the test.
	testSeq = NewSequence[error](
		errors.New("timeout"),
		errors.New("connection reset"),
		nil,
	)

	// 3. The anonymous function accesses the global variable rather than a
	// local one. This keeps the closure itself stateless, which is
	// incidental here (Patch works fine with closures that do capture
	// locals) but is a cheap habit: it keeps the mutable state ownership
	// obviously in the SequenceBuilder rather than smeared across a
	// closure the test can't easily inspect.
	patches := ApplyFunc(FetchData, func() (string, error) {
		err := testSeq.Next()
		if err != nil {
			return "", err
		}
		return "Success on 3rd try", nil
	})
	defer patches.Reset()

	_, err1 := FetchData()    // timeout
	_, err2 := FetchData()    // connection reset
	res3, err3 := FetchData() // success

	if err1 == nil || err2 == nil || err3 != nil {
		t.Fatalf("sequence did not execute correctly: err1=%v err2=%v err3=%v", err1, err2, err3)
	}
	if res3 != "Success on 3rd try" {
		t.Fatalf("final sequence state failed, got: %s", res3)
	}
}

// TestSequenceBuilder_ExhaustionHoldsLastValue tests the SequenceBuilder
// directly, with no patching involved. This isolates one specific
// behavior: once you run past the end of the declared sequence, Next()
// should keep returning the last value indefinitely rather than panicking
// or wrapping around.
func TestSequenceBuilder_ExhaustionHoldsLastValue(t *testing.T) {
	seq := NewSequence(1, 2, 3)

	want := []int{1, 2, 3, 3, 3} // calls 4 and 5 go past the end
	for i, w := range want {
		if got := seq.Next(); got != w {
			t.Fatalf("call %d: expected %d, got %d", i+1, w, got)
		}
	}
}

// TestSequenceBuilder_ConcurrentAccess exercises the mutex inside
// SequenceBuilder directly. Run with `go test -race` - this is checking
// that concurrent Next() calls don't corrupt the `current` counter, not
// that any particular goroutine gets any particular value (that's
// inherently unordered across goroutines).
func TestSequenceBuilder_ConcurrentAccess(t *testing.T) {
	seq := NewSequence(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)

	var wg sync.WaitGroup
	results := make([]int, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = seq.Next()
		}(i)
	}
	wg.Wait()

	for _, v := range results {
		if v < 0 || v > 9 {
			t.Fatalf("got out-of-range value %d - Next() is not safe for concurrent use", v)
		}
	}
}

// =====================================================================
// Group patching: several functions patched and unpatched together
// =====================================================================

// TestPatches_GroupAppliesAndResetsAll checks the Patches/Apply/Reset
// group API - patch several unrelated functions in one batch, confirm
// every override is active, then confirm a single Reset() call restores
// every one of them.
func TestPatches_GroupAppliesAndResetsAll(t *testing.T) {
	patches := ApplyFunc(Add, func(a, b int) int { return -1 })
	Apply(patches, GetGreeting, func() string { return "grouped stub" })

	if got := Add(1, 1); got != -1 {
		t.Fatalf("expected grouped Add patch active, got %d", got)
	}
	if got := GetGreeting(); got != "grouped stub" {
		t.Fatalf("expected grouped GetGreeting patch active, got %q", got)
	}

	patches.Reset()

	if got := Add(1, 1); got != 2 {
		t.Fatalf("expected Add restored to real implementation, got %d", got)
	}
	if got := GetGreeting(); got != "hello from the real implementation" {
		t.Fatalf("expected GetGreeting restored to real implementation, got %q", got)
	}
}
