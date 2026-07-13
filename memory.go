package badgerpatch

import (
	"syscall"
	"unsafe"
)

// rawMemoryAccess creates a zero-allocation view of memory
func rawMemoryAccess(p uintptr, length int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(p)), length)
}

func pageStart(ptr uintptr) uintptr {
	return ptr &^ (uintptr(syscall.Getpagesize() - 1))
}
