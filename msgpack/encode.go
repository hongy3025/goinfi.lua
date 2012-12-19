package msgpack

import (
	"io"
	"fmt"
	"unsafe"
)

const MBUFSIZE = 10

var MyPrintln = fmt.Println

func packInt8(out []byte, value int8) int {
	v := uint8(value)
	out[0] = 0xd0
	out[1] = byte(v)
	return 2
}

func packInt16(out []byte, value int16) int {
	v := uint16(value)
	out[0] = 0xd1
	out[1] = byte(v)
	out[2] = byte(v>>8)
	return 3
}

func packInt32(out []byte, value int32) int {
	v := uint32(value)
	out[0] = 0xd2
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func packInt64(out []byte, value int64) int {
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

func packUint8(out []byte, value uint8) int {
	out[0] = 0xcc
	out[1] = value
	return 2
}

func packUint16(out []byte, value uint16) int {
	v := uint16(value)
	out[0] = 0xcd
	out[1] = byte(v)
	out[2] = byte(v>>8)
	return 3
}

func packUint32(out []byte, value uint32) int {
	v := uint32(value)
	out[0] = 0xce
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func packUint64(out []byte, value uint64) int {
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

func packVarint(out []byte, value int64) int {
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
		return packInt8(out, int8(value))
	}

	if -1<<15 <= value && value < 1<<15 {
		return packInt16(out, int16(value))
	}

	if -1<<31 <= value && value < 1<<31 {
		return packInt32(out, int32(value))
	}

	return packInt64(out, value)
}

func packVaruint(out []byte, value uint64) int {
	// for fixnum
	if value < 128 {
		out[0] = byte(uint8(value))
		return 1
	}

	if value < 1<<8 {
		return packUint8(out, uint8(value))
	}

	if value < 1<<16 {
		return packUint16(out, uint16(value))
	}

	if value < 1<<32 {
		return packUint32(out, uint32(value))
	}

	return packUint64(out, value)
}

func packNil(out []byte) int {
	out[0] = 0xc0
	return 1
}

func packBoolean(out []byte, value bool) int {
	if value {
		out[0] = 0xc3
		return 1
	}

	out[0] = 0xc2
	return 1
}

func packFloat32(out []byte, value float32) int {
	out[0] = 0xca
	v := *(*uint32)(unsafe.Pointer(&value))
	out[1] = byte(v)
	out[2] = byte(v>>8)
	out[3] = byte(v>>16)
	out[4] = byte(v>>24)
	return 5
}

func packFloat64(out []byte, value float64) int {
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

func packRawHead(out []byte, n uint32) int {
	if n < 1<<5 {
		out[0] = byte(uint8(n)|0xa0)
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

func packArrayHead(out []byte, n uint32) int {
	if n < 1<<4 {
		out[0] = byte(uint8(n)|0x90)
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

func packMapHead(out []byte, n uint32) int {
	if n < 1<<4 {
		out[0] = byte(uint8(n)|0x80)
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

func PackInt8(out io.Writer, value int8) (int, error) {
	var buf [MBUFSIZE]byte
	n := packInt8(buf[:], value)
	return out.Write(buf[:n])
}

func PackInt16(out io.Writer, value int16) (int, error) {
	var buf [MBUFSIZE]byte
	n := packInt16(buf[:], value)
	return out.Write(buf[:n])
}

func PackInt32(out io.Writer, value int32) (int, error) {
	var buf [MBUFSIZE]byte
	n := packInt32(buf[:], value)
	return out.Write(buf[:n])
}

func PackInt64(out io.Writer, value int64) (int, error) {
	var buf [MBUFSIZE]byte
	n := packInt64(buf[:], value)
	return out.Write(buf[:n])
}

func PackUint8(out io.Writer, value uint8) (int, error) {
	var buf [MBUFSIZE]byte
	n := packUint8(buf[:], value)
	return out.Write(buf[:n])
}

func PackUint16(out io.Writer, value uint16) (int, error) {
	var buf [MBUFSIZE]byte
	n := packUint16(buf[:], value)
	return out.Write(buf[:n])
}

func PackUint32(out io.Writer, value uint32) (int, error) {
	var buf [MBUFSIZE]byte
	n := packUint32(buf[:], value)
	return out.Write(buf[:n])
}

func PackUint64(out io.Writer, value uint64) (int, error) {
	var buf [MBUFSIZE]byte
	n := packUint64(buf[:], value)
	return out.Write(buf[:n])
}

func PackVarint(out io.Writer, value int64) (int, error) {
	var buf [MBUFSIZE]byte
	n := packVarint(buf[:], value)
	return out.Write(buf[:n])
}

func PackVaruint(out io.Writer, value uint64) (int, error) {
	var buf [MBUFSIZE]byte
	n := packVaruint(buf[:], value)
	return out.Write(buf[:n])
}

func PackNil(out io.Writer) (int, error) {
	var buf [MBUFSIZE]byte
	n := packNil(buf[:])
	return out.Write(buf[:n])
}

func PackBoolean(out io.Writer, value bool) (int, error) {
	var buf [MBUFSIZE]byte
	n := packBoolean(buf[:], value)
	return out.Write(buf[:n])
}

func PackFloat32(out io.Writer, value float32) (int, error) {
	var buf [MBUFSIZE]byte
	n := packFloat32(buf[:], value)
	return out.Write(buf[:n])
}

func PackFloat64(out io.Writer, value float64) (int, error) {
	var buf [MBUFSIZE]byte
	n := packFloat64(buf[:], value)
	return out.Write(buf[:n])
}

func PackRaw(out io.Writer, data []byte) (int, error) {
	var buf [MBUFSIZE]byte
	n := packRawHead(buf[:], uint32(len(data)))
	res, err := out.Write(buf[:n])
	if err != nil {
		return res, err
	}
	bn, err := out.Write(data)
	return n+bn, err
}

func PackArrayHead(out io.Writer, size uint32) (int, error) {
	var buf [MBUFSIZE]byte
	n := packArrayHead(buf[:], size)
	return out.Write(buf[:n])
}

func PackMapHead(out io.Writer, size uint32) (int, error) {
	var buf [MBUFSIZE]byte
	n := packMapHead(buf[:], size)
	return out.Write(buf[:n])
}

