package badgerpatch

// StaticReturn generates a mock that purely bypasses logic with a direct override
func StaticReturn[T any](target T, mockBehavior T) *PatchGuard {
	return Patch(target, mockBehavior)
}