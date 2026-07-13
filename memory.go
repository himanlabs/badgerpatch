package badgerpatch

import (
	"unsafe"
)

// rawMemoryAccess creates a zero-allocation view of memory
func rawMemoryAccess(p uintptr, length int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), length)
}

