package vec

import (
	"bytes"
	"math"
	"testing"
)

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
