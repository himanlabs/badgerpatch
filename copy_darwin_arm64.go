//go:build darwin && arm64

package badgerpatch

/*
#include <pthread.h>

// Apple Silicon requires threads to manually toggle between Write and Execute permissions.
// We use CGO to hook into macOS's native pthread library.
void disable_jit() { pthread_jit_write_protect_np(0); }
void enable_jit() { pthread_jit_write_protect_np(1); }
*/
import "C"

// copyToLocation safely modifies memory on Apple Silicon (M-Series) Macs
func copyToLocation(location uintptr, data []byte) {
	// 1. Tell the macOS kernel to allow writing to executable memory for this thread
	C.disable_jit()
	
	// 2. Ensure we lock it back to Execute-Only when we finish
	defer C.enable_jit()

	// 3. Write our Jump instructions directly into memory
	f := rawMemoryAccess(location, len(data))
	copy(f, data)
}