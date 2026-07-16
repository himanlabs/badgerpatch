//go:build amd64

package badgerpatch

// R11 is used to protect Go's register-based calling conventions.
// `from` is unused here: amd64's movabs+jmp is always a fixed 13-byte
// absolute sequence regardless of distance, unlike arm64 where distance
// determines which (differently-sized) encoding is safe to use.
func buildJmpDirective(to, from uintptr) []byte {
	_ = from
	return []byte{
		0x49, 0xBB, // movabs r11, to
		byte(to), byte(to >> 8), byte(to >> 16), byte(to >> 24),
		byte(to >> 32), byte(to >> 40), byte(to >> 48), byte(to >> 56),
		0x41, 0xFF, 0xE3, // jmp r11
	}
}