// Copyright 2025 go-highway Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hwy

import (
	"math/bits"
	"unsafe"
)

// This file provides bit manipulation operations for integer vectors.
// These are pure Go (scalar) implementations that work with any integer type.
// When SIMD implementations are available, they will be used via dispatch.

// PopCount counts the number of set bits (1s) in each lane.
func PopCount[T Integers](v Vec[T]) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = popCount(v.data[i])
	}
	return Vec[T]{data: result}
}

// popCount counts set bits for a single value.
func popCount[T Integers](val T) T {
	switch any(val).(type) {
	case int8:
		return T(bits.OnesCount8(uint8(any(val).(int8))))
	case uint8:
		return T(bits.OnesCount8(any(val).(uint8)))
	case int16:
		return T(bits.OnesCount16(uint16(any(val).(int16))))
	case uint16:
		return T(bits.OnesCount16(any(val).(uint16)))
	case int32:
		return T(bits.OnesCount32(uint32(any(val).(int32))))
	case uint32:
		return T(bits.OnesCount32(any(val).(uint32)))
	case int64:
		return T(bits.OnesCount64(uint64(any(val).(int64))))
	case uint64:
		return T(bits.OnesCount64(any(val).(uint64)))
	default:
		return 0
	}
}

// LeadingZeroCount counts the number of leading zero bits in each lane.
func LeadingZeroCount[T Integers](v Vec[T]) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = leadingZeroCount(v.data[i])
	}
	return Vec[T]{data: result}
}

// leadingZeroCount counts leading zeros for a single value.
func leadingZeroCount[T Integers](val T) T {
	switch any(val).(type) {
	case int8:
		return T(bits.LeadingZeros8(uint8(any(val).(int8))))
	case uint8:
		return T(bits.LeadingZeros8(any(val).(uint8)))
	case int16:
		return T(bits.LeadingZeros16(uint16(any(val).(int16))))
	case uint16:
		return T(bits.LeadingZeros16(any(val).(uint16)))
	case int32:
		return T(bits.LeadingZeros32(uint32(any(val).(int32))))
	case uint32:
		return T(bits.LeadingZeros32(any(val).(uint32)))
	case int64:
		return T(bits.LeadingZeros64(uint64(any(val).(int64))))
	case uint64:
		return T(bits.LeadingZeros64(any(val).(uint64)))
	default:
		return 0
	}
}

// TrailingZeroCount counts the number of trailing zero bits in each lane.
func TrailingZeroCount[T Integers](v Vec[T]) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = trailingZeroCount(v.data[i])
	}
	return Vec[T]{data: result}
}

// trailingZeroCount counts trailing zeros for a single value.
func trailingZeroCount[T Integers](val T) T {
	switch any(val).(type) {
	case int8:
		v := uint8(any(val).(int8))
		if v == 0 {
			return T(8)
		}
		return T(bits.TrailingZeros8(v))
	case uint8:
		v := any(val).(uint8)
		if v == 0 {
			return T(8)
		}
		return T(bits.TrailingZeros8(v))
	case int16:
		v := uint16(any(val).(int16))
		if v == 0 {
			return T(16)
		}
		return T(bits.TrailingZeros16(v))
	case uint16:
		v := any(val).(uint16)
		if v == 0 {
			return T(16)
		}
		return T(bits.TrailingZeros16(v))
	case int32:
		v := uint32(any(val).(int32))
		if v == 0 {
			return T(32)
		}
		return T(bits.TrailingZeros32(v))
	case uint32:
		v := any(val).(uint32)
		if v == 0 {
			return T(32)
		}
		return T(bits.TrailingZeros32(v))
	case int64:
		v := uint64(any(val).(int64))
		if v == 0 {
			return T(64)
		}
		return T(bits.TrailingZeros64(v))
	case uint64:
		v := any(val).(uint64)
		if v == 0 {
			return T(64)
		}
		return T(bits.TrailingZeros64(v))
	default:
		return 0
	}
}

// RotateRight rotates the bits in each lane to the right by the specified count.
func RotateRight[T Integers](v Vec[T], count int) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = rotateRight(v.data[i], count)
	}
	return Vec[T]{data: result}
}

// rotateRight rotates bits right for a single value.
func rotateRight[T Integers](val T, count int) T {
	switch any(val).(type) {
	case int8:
		v := uint8(any(val).(int8))
		return T(int8(bits.RotateLeft8(v, -count)))
	case uint8:
		v := any(val).(uint8)
		return T(bits.RotateLeft8(v, -count))
	case int16:
		v := uint16(any(val).(int16))
		return T(int16(bits.RotateLeft16(v, -count)))
	case uint16:
		v := any(val).(uint16)
		return T(bits.RotateLeft16(v, -count))
	case int32:
		v := uint32(any(val).(int32))
		return T(int32(bits.RotateLeft32(v, -count)))
	case uint32:
		v := any(val).(uint32)
		return T(bits.RotateLeft32(v, -count))
	case int64:
		v := uint64(any(val).(int64))
		return T(int64(bits.RotateLeft64(v, -count)))
	case uint64:
		v := any(val).(uint64)
		return T(bits.RotateLeft64(v, -count))
	default:
		return val
	}
}

// ReverseBits reverses the bit order in each lane.
func ReverseBits[T Integers](v Vec[T]) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = reverseBits(v.data[i])
	}
	return Vec[T]{data: result}
}

// reverseBits reverses bit order for a single value.
func reverseBits[T Integers](val T) T {
	switch any(val).(type) {
	case int8:
		v := uint8(any(val).(int8))
		return T(int8(bits.Reverse8(v)))
	case uint8:
		v := any(val).(uint8)
		return T(bits.Reverse8(v))
	case int16:
		v := uint16(any(val).(int16))
		return T(int16(bits.Reverse16(v)))
	case uint16:
		v := any(val).(uint16)
		return T(bits.Reverse16(v))
	case int32:
		v := uint32(any(val).(int32))
		return T(int32(bits.Reverse32(v)))
	case uint32:
		v := any(val).(uint32)
		return T(bits.Reverse32(v))
	case int64:
		v := uint64(any(val).(int64))
		return T(int64(bits.Reverse64(v)))
	case uint64:
		v := any(val).(uint64)
		return T(bits.Reverse64(v))
	default:
		return val
	}
}

// HighestSetBitIndex returns the index of the highest set bit in each lane.
// Returns -1 for lanes with value 0.
// For a value with bit pattern ...001xxx, returns the position of the leftmost 1.
// This is equivalent to floor(log2(x)) for non-zero values.
func HighestSetBitIndex[T Integers](v Vec[T]) Vec[T] {
	result := make([]T, len(v.data))
	for i := 0; i < len(v.data); i++ {
		result[i] = highestSetBitIndex(v.data[i])
	}
	return Vec[T]{data: result}
}

// highestSetBitIndex returns the index of the highest set bit for a single value.
// Returns the maximum value of the type (all bits set) if the value is 0.
// For signed types this is -1, for unsigned types this is the max value.
func highestSetBitIndex[T Integers](val T) T {
	// A single arithmetic path instead of an eight-way type switch: the switch
	// made the s390x compiler fail with "width not calculated: [2048]uint8"
	// when instantiated from a test binary (see the go-simd-highway-ice notes).
	// A zero lane gives bits.Len64(0)-1 == -1, which converts to -1 for signed
	// types and to all-ones for unsigned ones, as documented above.
	u := uint64(val)
	if width := unsafe.Sizeof(val) * 8; width < 64 {
		u &= (uint64(1) << width) - 1 // drop the sign extension of negative lanes
	}
	return T(bits.Len64(u) - 1)
}
