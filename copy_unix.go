//go:build !windows && !(darwin && arm64)

package badgerpatch

import "syscall"

func copyToLocation(location uintptr, data []byte) {
	pageSize := syscall.Getpagesize()
	start := pageStart(location)
	length := int(pageStart(location+uintptr(len(data)))-start) + pageSize

	page := rawMemoryAccess(start, length)

	if err := syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC); err != nil {
		panic(err)
	}

	f := rawMemoryAccess(location, len(data))
	copy(f, data)

	if err := syscall.Mprotect(page, syscall.PROT_READ|syscall.PROT_EXEC); err != nil {
		panic(err)
	}
}