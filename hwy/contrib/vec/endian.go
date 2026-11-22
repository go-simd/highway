package vec

import (
	"encoding/binary"
	"math"
	"unsafe"
)

// hostBigEndian reports whether the host stores multi-byte values most
// significant byte first. The float encoders reinterpret memory as bytes, which
// only yields the documented little-endian wire format on little-endian hosts;
// on big-endian hosts (s390x) they take the explicit paths below.
var hostBigEndian = func() bool {
	x := uint16(1)
	return *(*byte)(unsafe.Pointer(&x)) == 0
}()

func encodeFloat32sLE(dst []byte, src []float32) {
	for i, f := range src {
		binary.LittleEndian.PutUint32(dst[i*4:], math.Float32bits(f))
	}
}

func decodeFloat32sLE(dst []float32, src []byte) {
	for i := range dst {
		dst[i] = math.Float32frombits(binary.LittleEndian.Uint32(src[i*4:]))
	}
}

func encodeFloat64sLE(dst []byte, src []float64) {
	for i, f := range src {
		binary.LittleEndian.PutUint64(dst[i*8:], math.Float64bits(f))
	}
}

func decodeFloat64sLE(dst []float64, src []byte) {
	for i := range dst {
		dst[i] = math.Float64frombits(binary.LittleEndian.Uint64(src[i*8:]))
	}
}
