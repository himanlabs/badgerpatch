//go:build arm64

package badgerpatch

import "unsafe"

// buildJmpDirective constructs a jump for ARM architecture.
// It prioritizes a 4-byte relative branch if the distance is within ±128MB.
// If the functions are allocated further apart, it falls back to a 20-byte absolute jump.
func buildJmpDirective(to, from uintptr) []byte {
	// 1. Calculate the distance between the target and the mock
	offset := int64(to) - int64(from)

	// 2. Optimization: ARM64 relative branch ('B' instruction) is 26 bits.
	// This gives us a ±128MB range. Since Go usually packs functions together,
	// this will succeed 99% of the time, using only 4 bytes instead of 20!
	if offset >= -0x8000000 && offset < 0x8000000 {
		// Shift right by 2 to get the instruction offset (since instructions are 4-byte aligned)
		instOffset := uint32(offset>>2) & 0x03FFFFFF
		// Combine with the 'B' opcode (0x14000000)
		instruction := 0x14000000 | instOffset

		return []byte{
			byte(instruction),
			byte(instruction >> 8),
			byte(instruction >> 16),
			byte(instruction >> 24),
		}
	}

	// 3. Fallback: If distance > 128MB, use the 20-byte absolute jump using register X26
	res := make([]byte, 0, 20)
	d0d1 := to & 0xFFFF
	d2d3 := to >> 16 & 0xFFFF
	d4d5 := to >> 32 & 0xFFFF
	d6d7 := to >> 48 & 0xFFFF

	res = append(res, movImm(0b10, 0, d0d1)...)          // MOVZ x26, to[16:0]
	res = append(res, movImm(0b11, 1, d2d3)...)          // MOVK x26, to[32:16]
	res = append(res, movImm(0b11, 2, d4d5)...)          // MOVK x26, to[48:32]
	res = append(res, movImm(0b11, 3, d6d7)...)          // MOVK x26, to[64:48]
	res = append(res, []byte{0x40, 0x03, 0x1F, 0xD6}...) // BR x26 (Branch to address in x26)

	return res
}

func movImm(opc, shift int, val uintptr) []byte {
	var m uint32 = 26          // rd (Register x26)
	m |= uint32(val) << 5      // imm16
	m |= uint32(shift&3) << 21 // hw
	m |= 0b100101 << 23        // const
	m |= uint32(opc&0x3) << 29 // opc
	m |= 0b1 << 31             // sf

	res := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&res[0])) = m
	return res
}
