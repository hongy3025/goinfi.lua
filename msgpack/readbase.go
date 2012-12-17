package msgpack

import (
	"io"
	"unsafe"
)

func decodeUint8(buf []byte) uint8 {
	return uint8(buf[0])
}

func decodeUint16(buf []byte) uint16 {
	return uint16(buf[0]) + uint16(buf[1]<<8)
}

func decodeUint32(buf []byte) uint32 {
	return uint32(buf[0]) + uint32(buf[1]<<8) + uint32(buf[2]<<16) + uint32(buf[3]<<24)
}

func decodeUint64(buf []byte) uint64 {
	return uint64(buf[0]) + uint64(buf[1]<<8) + uint64(buf[2]<<16) + uint64(buf[3]<<24) + uint64(buf[4]<<32) + uint64(buf[5]<<40) + uint64(buf[6]<<48) + uint64(buf[7] << 56)
}

func readRaw(n int, in io.Reader) (Elem, error) {
	b := make([]byte, n)
	rn, err := in.Read(b)
	if rn != n {
		return nil, err
	}
	return b, nil
}

func readUint8(in io.Reader) (uint8, error) {
	var tmp [1]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint8(buf)
	return v, nil
}

func readInt8(in io.Reader) (int8, error) {
	v, err := readUint8(in)
	if err != nil { return 0, err }
	return int8(v), nil
}

func readUint16(in io.Reader) (uint16, error) {
	var tmp [2]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint16(buf)
	return v, nil
}

func readInt16(in io.Reader) (int16, error) {
	v, err := readUint16(in)
	if err != nil { return 0, err }
	return int16(v), nil
}


func readFloat32(in io.Reader) (float32, error) {
	var tmp [4]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint32(buf)
	value := *(*float32)(unsafe.Pointer(&v))

	return value, nil
}

func readFloat64(in io.Reader) (float64, error) {
	var tmp [8]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint64(buf)
	value := *(*float64)(unsafe.Pointer(&v))

	return value, nil
}

