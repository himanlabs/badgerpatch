//go:build amd64

package badgerpatch

// R11 is used to protect Go's register-based calling conventions
func buildJmpDirective(to uintptr) []byte {
	return []byte{
		0x49, 0xBB, // movabs r11, to
		byte(to), byte(to >> 8), byte(to >> 16), byte(to >> 24),
		byte(to >> 32), byte(to >> 40), byte(to >> 48), byte(to >> 56),
		0x41, 0xFF, 0xE3, // jmp r11
	}
}