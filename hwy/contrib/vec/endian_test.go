package vec

import (
	"bytes"
	"math"
	"testing"
)

// The Base* functions take the explicit little-endian path exactly when the host
// is big-endian. Forcing hostBigEndian drives that path on a little-endian host.
func TestEndianPathSelection(t *testing.T) {
	real := hostBigEndian
	defer func() { hostBigEndian = real }()
	f32 := []float32{1, 2, 3}
	f64 := []float64{1, 2}
	b := make([]byte, 24)
	calls := map[string]func(){
		"EncodeFloat32s":    func() { BaseEncodeFloat32s(b, f32) },
		"DecodeFloat32s":    func() { BaseDecodeFloat32s(make([]float32, 3), b) },
		"EncodeFloat64s":    func() { BaseEncodeFloat64s(b, f64) },
		"DecodeFloat64s":    func() { BaseDecodeFloat64s(make([]float64, 2), b) },
		"fb EncodeFloat32s": func() { BaseEncodeFloat32s_fallback(b, f32) },
		"fb DecodeFloat32s": func() { BaseDecodeFloat32s_fallback(make([]float32, 3), b) },
		"fb EncodeFloat64s": func() { BaseEncodeFloat64s_fallback(b, f64) },
		"fb DecodeFloat64s": func() { BaseDecodeFloat64s_fallback(make([]float64, 2), b) },
	}
	for name, fn := range calls {
		hostBigEndian = true
		before := bePathCalls.Load()
		fn()
		if bePathCalls.Load() != before+1 {
			t.Errorf("%s: big-endian host did not take the explicit path", name)
		}
		if !real { // on a real big-endian host the unsafe path would be wrong
			hostBigEndian = false
			before = bePathCalls.Load()
			fn()
			if bePathCalls.Load() != before {
				t.Errorf("%s: little-endian host took the explicit path", name)
			}
		}
	}
}

// Byte-level golden vectors: the wire format is IEEE-754 little-endian.
func TestFloatEncodingGoldenBytes(t *testing.T) {
	f32 := []float32{1, -2, math.Float32frombits(0x01020304)}
	g32 := []byte{0x00, 0x00, 0x80, 0x3f, 0x00, 0x00, 0x00, 0xc0, 0x04, 0x03, 0x02, 0x01}
	f64 := []float64{1, math.Float64frombits(0x0102030405060708)}
	g64 := []byte{0, 0, 0, 0, 0, 0, 0xf0, 0x3f, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}

	enc32 := map[string]func([]byte, []float32){
		"BaseEncodeFloat32s": BaseEncodeFloat32s, "fallback": BaseEncodeFloat32s_fallback, "LE": encodeFloat32sLE,
	}
	for name, fn := range enc32 {
		dst := make([]byte, len(g32))
		fn(dst, f32)
		if !bytes.Equal(dst, g32) {
			t.Errorf("%s: % x, want % x", name, dst, g32)
		}
	}
	dec32 := map[string]func([]float32, []byte){
		"BaseDecodeFloat32s": BaseDecodeFloat32s, "fallback": BaseDecodeFloat32s_fallback, "LE": decodeFloat32sLE,
	}
	for name, fn := range dec32 {
		dst := make([]float32, len(f32))
		fn(dst, g32)
		for i := range f32 {
			if math.Float32bits(dst[i]) != math.Float32bits(f32[i]) {
				t.Errorf("%s decode[%d] = %#x", name, i, math.Float32bits(dst[i]))
			}
		}
	}
	enc64 := map[string]func([]byte, []float64){
		"BaseEncodeFloat64s": BaseEncodeFloat64s, "fallback": BaseEncodeFloat64s_fallback, "LE": encodeFloat64sLE,
	}
	for name, fn := range enc64 {
		dst := make([]byte, len(g64))
		fn(dst, f64)
		if !bytes.Equal(dst, g64) {
			t.Errorf("%s: % x, want % x", name, dst, g64)
		}
	}
	dec64 := map[string]func([]float64, []byte){
		"BaseDecodeFloat64s": BaseDecodeFloat64s, "fallback": BaseDecodeFloat64s_fallback, "LE": decodeFloat64sLE,
	}
	for name, fn := range dec64 {
		dst := make([]float64, len(f64))
		fn(dst, g64)
		for i := range f64 {
			if math.Float64bits(dst[i]) != math.Float64bits(f64[i]) {
				t.Errorf("%s decode[%d] = %#x", name, i, math.Float64bits(dst[i]))
			}
		}
	}
}
