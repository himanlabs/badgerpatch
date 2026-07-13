//go:build windows

package badgerpatch

import (
	"syscall"
	"unsafe"
)

const PAGE_EXECUTE_READWRITE = 0x40

var procVirtualProtect = syscall.NewLazyDLL("kernel32.dll").NewProc("VirtualProtect")

func virtualProtect(lpAddress uintptr, dwSize int, flNewProtect uint32, lpflOldProtect unsafe.Pointer) error {
	ret, _, _ := procVirtualProtect.Call(
		lpAddress, uintptr(dwSize), uintptr(flNewProtect), uintptr(lpflOldProtect),
	)
	if ret == 0 {
		return syscall.GetLastError()
	}
	return nil
}

func copyToLocation(location uintptr, data []byte) {
	f := rawMemoryAccess(location, len(data))

	var oldPerms uint32
	if err := virtualProtect(location, len(data), PAGE_EXECUTE_READWRITE, unsafe.Pointer(&oldPerms)); err != nil {
		panic(err)
	}
	
	copy(f, data)

	var tmp uint32
	if err := virtualProtect(location, len(data), oldPerms, unsafe.Pointer(&tmp)); err != nil {
		panic(err)
	}
}