package msgpack

import (
	// "io"
	"fmt"
	"unsafe"
)

var MyPrintln = fmt.Println

func PackInt8(out []byte, value int8) int {
	v := uint8(value)
	out[0] = 0xd0
	out[1] = byte(v)
	return 2
}

func PackInt16(out []byte, value int16) int {
	v := uint16(value)
	out[0] = 0xd1
	out[1] = byte(v)
	out[2] = byte(v>>8)
	return 3
}

func PackInt32(out []byte, value int32) int {
	v := uint32(value)
	out[0] = 0xd2
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func PackInt64(out []byte, value int64) int {
	v := uint64(value)
	out[0] = 0xd3
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	out[5] = byte(v>>32)
	out[6] = byte(v>>40)
	out[7] = byte(v>>48)
	out[8] = byte(v>>56)
	return 9
}

func PackUint8(out []byte, value uint8) int {
	out[0] = 0xcc
	out[1] = value
	return 2
}

func PackUint16(out []byte, value uint16) int {
	v := uint16(value)
	out[0] = 0xcd
	out[1] = byte(v)
	out[2] = byte(v>>8)
	return 3
}

func PackUint32(out []byte, value uint32) int {
	v := uint32(value)
	out[0] = 0xce
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func PackUint64(out []byte, value uint64) int {
	v := uint64(value)
	out[0] = 0xcf
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	out[5] = byte(v>>32)
	out[6] = byte(v>>40)
	out[7] = byte(v>>48)
	out[8] = byte(v>>56)
	return 9
}

func PackVarint(out []byte, value int64) int {
	// for fixnum
	if value < 128 {
		if value >= 0 {
			out[0] = byte(uint8(value))
			return 1
		}
		if value >= -32 {
			out[0] = byte( uint8(32+value) | uint8(0xe0) )
			return 1
		}
	}

	if -1<<7 <= value && value < 1<<7 {
		return PackInt8(out, int8(value))
	}

	if -1<<15 <= value && value < 1<<15 {
		return PackInt16(out, int16(value))
	}

	if -1<<31 <= value && value < 1<<31 {
		return PackInt32(out, int32(value))
	}

	return PackInt64(out, value)
}

func PackVaruint(out []byte, value uint64) int {
	// for fixnum
	if value < 128 {
		out[0] = byte(uint8(value))
		return 1
	}

	if value < 1<<8 {
		return PackUint8(out, uint8(value))
	}

	if value < 1<<16 {
		return PackUint16(out, uint16(value))
	}

	if value < 1<<32 {
		return PackUint32(out, uint32(value))
	}

	return PackUint64(out, value)
}

func PackNil(out []byte) int {
	out[0] = 0xc0
	return 1
}

func PackBoolean(out []byte, value bool) int {
	if value {
		out[0] = 0xc3
		return 1
	}

	out[0] = 0xc2
	return 0
}

func PackFloat32(out []byte, value float32) int {
	out[0] = 0xca
	v := *(*uint32)(unsafe.Pointer(&value))
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func PackFloat64(out []byte, value float64) int {
	out[0] = 0xcb
	v := *(*uint64)(unsafe.Pointer(&value))
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	out[5] = byte(v>>32)
	out[6] = byte(v>>40)
	out[7] = byte(v>>48)
	out[8] = byte(v>>56)
	return 9
}

func PackRawByteHead(out []byte, n uint32) int {
	if n < 1<<5 {
		out[0] = byte(uint8(n)&0xa0)
		return 1
	}
	if n < 1<<16 {
		out[0] = 0xda
		v := uint16(n)
		out[1] = byte(v)
		out[2] = byte(v>>8)
		return 3
	}

	out[0] = 0xdb
	v := uint32(n)
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func PackArrayHead(out []byte, n uint32) int {
	if n < 1<<4 {
		out[0] = byte(uint8(n)&0x90)
		return 1
	}
	if n < 1<<16 {
		out[0] = 0xdc
		v := uint16(n)
		out[1] = byte(v)
		out[2] = byte(v>>8)
		return 3
	}

	out[0] = 0xdd
	v := uint32(n)
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func PackMapHead(out []byte, n uint32) int {
	if n < 1<<4 {
		out[0] = byte(uint8(n)&0x80)
		return 1
	}
	if n < 1<<16 {
		out[0] = 0xde
		v := uint16(n)
		out[1] = byte(v)
		out[2] = byte(v>>8)
		return 3
	}

	out[0] = 0xdf
	v := uint32(n)
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

/*
type Packer struct {
}

func (packer *Packer) {
}
*/

