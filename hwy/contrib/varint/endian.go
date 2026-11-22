package varint

import (
	"sync/atomic"
	"unsafe"
)

// hostBigEndian reports whether the host stores multi-byte values most
// significant byte first. The Stream-VByte encoder shuffles the in-memory bytes
// of a []uint32, which are the little-endian wire bytes only on little-endian
// hosts; big-endian hosts (s390x) take the explicit shift-based path. It is a
// variable so a test can drive the big-endian path on a little-endian host.
var hostBigEndian = func() bool {
	x := uint16(1)
	return *(*byte)(unsafe.Pointer(&x)) == 0
}()

// bePathCalls counts group encodes taken through the explicit path, so a test
// can tell which path ran (both produce the same bytes on a little-endian host).
var bePathCalls atomic.Uint64
