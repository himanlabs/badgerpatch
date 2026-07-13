//go:build arm64

package badgerpatch

import "unsafe"

func buildJmpDirective(to uintptr) []byte {
	res := make([]byte, 0, 24)
	d0d1 := to & 0xFFFF
	d2d3 := to >> 16 & 0xFFFF
	d4d5 := to >> 32 & 0xFFFF
	d6d7 := to >> 48 & 0xFFFF

	res = append(res, movImm(0b10, 0, d0d1)...)          // MOVZ x26, to[16:0]
	res = append(res, movImm(0b11, 1, d2d3)...)          // MOVK x26, to[32:16]
	res = append(res, movImm(0b11, 2, d4d5)...)          // MOVK x26, to[48:32]
	res = append(res, movImm(0b11, 3, d6d7)...)          // MOVK x26, to[64:48]
	res = append(res, []byte{0x40, 0x03, 0x1F, 0xD6}...) // BR x26 (jump directly to the address we just built)

	return res
}

func movImm(opc, shift int, val uintptr) []byte {
	var m uint32 = 26
	m |= uint32(val) << 5
	m |= uint32(shift&3) << 21
	m |= 0b100101 << 23
	m |= uint32(opc&0x3) << 29
	m |= 0b1 << 31

	res := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&res[0])) = m
	return res
}