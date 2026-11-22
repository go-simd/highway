package varint

import "unsafe"

// hostBigEndian reports whether the host stores multi-byte values most
// significant byte first. The Stream-VByte encoder shuffles the in-memory bytes
// of a []uint32, which are the little-endian wire bytes only on little-endian
// hosts; big-endian hosts (s390x) take the explicit shift-based path.
var hostBigEndian = func() bool {
	x := uint16(1)
	return *(*byte)(unsafe.Pointer(&x)) == 0
}()
