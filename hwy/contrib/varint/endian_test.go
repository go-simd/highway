package varint

import (
	"bytes"
	"testing"
)

// Golden vector from the Stream-VByte format: values are stored as
// little-endian, minimal-length byte runs. ctrl bits: v0 4B, v1 2B, v2 1B, v3 3B.
func TestStreamVByte32GroupGoldenBytes(t *testing.T) {
	values := []uint32{0x01020304, 0x0506, 0x07, 0x080910}
	wantCtrl := byte(3 | 1<<2 | 0<<4 | 2<<6)
	wantData := []byte{0x04, 0x03, 0x02, 0x01, 0x06, 0x05, 0x07, 0x10, 0x09, 0x08}

	check := func(name string, fn func([]uint32, []byte) (byte, int)) {
		t.Helper()
		dst := make([]byte, 16)
		ctrl, n := fn(values, dst)
		if ctrl != wantCtrl || n != len(wantData) || !bytes.Equal(dst[:n], wantData) {
			t.Errorf("%s: ctrl=%#x n=%d data=% x; want ctrl=%#x data=% x", name, ctrl, n, dst[:n], wantCtrl, wantData)
		}
	}
	check("BaseEncodeStreamVByte32Group", BaseEncodeStreamVByte32Group)
	// The explicit path used on big-endian hosts must agree on every host.
	check("encodeGroupScalarInto", encodeGroupScalarInto)

	// Decode the golden bytes back, padded as the SIMD decoder requires.
	data := append(append([]byte{}, wantData...), make([]byte, 16)...)
	got := make([]uint32, 4)
	BaseDecodeStreamVByte32GroupSIMD(wantCtrl, data, got)
	for i := range values {
		if got[i] != values[i] {
			t.Errorf("decode[%d] = %#x, want %#x", i, got[i], values[i])
		}
	}
}
