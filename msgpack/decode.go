package msgpack

import (
	"io"
	"fmt"
	"unsafe"
)

var decode_Println = fmt.Println

/*
Type Chart

Type		Binary		Hex
--------------------------------------------
+FixNum		0xxxxxxx	0x00 - 0x7f
FixMap		1000xxxx	0x80 - 0x8f
FixArray	1001xxxx	0x90 - 0x9f
FixRaw		101xxxxx	0xa0 - 0xbf
nil			11000000	0xc0
reserved	11000001	0xc1
false		11000010	0xc2
true		11000011	0xc3
reserved	11000100	0xc4
reserved	11000101	0xc5
reserved	11000110	0xc6
reserved	11000111	0xc7
reserved	11001000	0xc8
reserved	11001001	0xc9
float		11001010	0xca
double		11001011	0xcb
uint 8		11001100	0xcc
uint 16		11001101	0xcd
uint 32		11001110	0xce
uint 64		11001111	0xcf
int 8		11010000	0xd0
int 16		11010001	0xd1
int 32		11010010	0xd2
int 64		11010011	0xd3
reserved	11010100	0xd4
reserved	11010101	0xd5
reserved	11010110	0xd6
reserved	11010111	0xd7
reserved	11011000	0xd8
reserved	11011001	0xd9
raw 16		11011010	0xda
raw 32		11011011	0xdb
array 16	11011100	0xdc
array 32	11011101	0xdd
map 16		11011110	0xde
map 32		11011111	0xdf
-FixNum		111xxxxx	0xe0 - 0xff
------------------------------------------
*/

type Elem interface{}
type Nil struct {}

func IsNil(elem Elem) bool {
	_, ok := elem.(Nil)
	return ok
}

type Msg struct {
	Elems []Elem
}

type Unpacker struct {
}

type ElemPair struct {
	Key Elem
	Value Elem
}

type Map struct {
	//Size uint32
	Elems []ElemPair
}

type Array struct {
	//Size uint32
	Elems []Elem
}

func newMap(n uint32) *Map {
	m := new(Map)
	//m.Size = n
	m.Elems = make([]ElemPair, 0, n)
	return m
}

func newArray(n uint32) *Array {
	a := new(Array)
	//a.Size = n
	a.Elems = make([]Elem, 0, n)
	return a
}

func newMsg() *Msg {
	msg := new(Msg)
	msg.Elems = make([]Elem, 0, 1)
	return msg
}

func (msg * Msg) append(elem Elem) {
	msg.Elems = append(msg.Elems, elem)
}

func decodeUint8(buf []byte) uint8 {
	return uint8(buf[0])
}

func decodeUint16(buf []byte) uint16 {
	return uint16(buf[0]) + uint16(buf[1])<<8
}

func decodeUint32(buf []byte) uint32 {
	return uint32(buf[0]) + uint32(buf[1])<<8 + uint32(buf[2])<<16 + uint32(buf[3])<<24
}

func decodeUint64(buf []byte) uint64 {
	return uint64(buf[0]) + uint64(buf[1])<<8 + uint64(buf[2])<<16 + uint64(buf[3])<<24 + uint64(buf[4])<<32 + uint64(buf[5])<<40 + uint64(buf[6])<<48 + uint64(buf[7])<<56
}

func unpackRaw(n uint32, in io.Reader) (Elem, error) {
	b := make([]byte, n)
	rn, err := in.Read(b)
	if rn != int(n) {
		return nil, err
	}
	return b, nil
}

func unpackUint8(in io.Reader) (uint8, error) {
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

func unpackInt8(in io.Reader) (int8, error) {
	v, err := unpackUint8(in)
	if err != nil { return 0, err }
	return int8(v), nil
}

func unpackUint16(in io.Reader) (uint16, error) {
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

func unpackInt16(in io.Reader) (int16, error) {
	v, err := unpackUint16(in)
	if err != nil { return 0, err }
	return int16(v), nil
}

func unpackUint32(in io.Reader) (uint32, error) {
	var tmp [4]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint32(buf)
	return v, nil
}

func unpackInt32(in io.Reader) (int32, error) {
	v, err := unpackUint32(in)
	if err != nil { return 0, err }
	return int32(v), nil
}

func unpackUint64(in io.Reader) (uint64, error) {
	var tmp [8]byte
	buf := tmp[:]
	nbuf := len(buf)
	n, err := in.Read(buf)
	if n != nbuf {
		return 0, err
	}
	v := decodeUint64(buf)
	return v, nil
}

func unpackInt64(in io.Reader) (int64, error) {
	v, err := unpackUint64(in)
	if err != nil { return 0, err }
	return int64(v), nil
}

func unpackFloat32(in io.Reader) (float32, error) {
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

func unpackFloat64(in io.Reader) (float64, error) {
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


func unpackMap(n uint32, in io.Reader) (Elem, error) {
	m := newMap(n)

	for i:=uint32(0); i<n; i++ {
		key, err := unpackElem(in)
		if err != nil {
			return nil, err
		}
		value, err := unpackElem(in)
		if err != nil {
			return nil, err
		}
		m.Elems = append(m.Elems, ElemPair{key, value})
	}

	return m, nil
}

func unpackArray(n uint32, in io.Reader) (Elem, error) {
	a := newArray(n)

	for i:=uint32(0); i<n; i++ {
		elem, err := unpackElem(in)
		if err != nil {
			return nil, err
		}
		a.Elems = append(a.Elems, elem)
	}

	return a, nil
}

func (u *Unpacker) reset() {
}

func (u *Unpacker) Unpack(in io.Reader) (*Msg, error) {
	msg := newMsg()
	for {
		elem, err := unpackElem(in)
		if elem != nil {
			msg.append(elem)
		}
		if err != nil {
			break
		}
	}
	return msg, nil
}

func unpackElem(in io.Reader) (Elem, error) {
	var elem Elem
	var err error
	var tmp [1]byte

	buf := tmp[:]
	n, err := in.Read(buf)
	if n == 0 {
		return nil, err
	}
	tag := uint8(buf[0])
	switch {
	case tag <= 0x7f:
		return int(tag), nil
	case tag <= 0x8f:
		return unpackMap(uint32(tag&(^uint8(0xf0))), in)
	case tag <= 0x9f:
		return unpackArray(uint32(tag&(^uint8(0xf0))), in)
	case tag <= 0xbf:
		return unpackRaw(uint32(tag&(^uint8(0xe0))), in)
	}

	switch tag {
	case 0xc0: // nil	0xc0
		elem = Nil{}
	case 0xc2: // false	0xc2
		elem = false
	case 0xc3: // true	0xc3
		elem = true
	case 0xca: // float	0xca
		elem, err = unpackFloat32(in)
	case 0xcb: // double	0xcb
		elem, err = unpackFloat64(in)
	case 0xcc: // uint 8	0xcc
		elem, err = unpackUint8(in)
	case 0xcd: // uint 16	0xcd
		elem, err = unpackUint16(in)
	case 0xce: // uint 32	0xce
		elem, err = unpackUint32(in)
	case 0xcf: // uint 64	0xcf
		elem, err = unpackUint64(in)
	case 0xd0: // int 8		0xd0
		elem, err = unpackInt8(in)
	case 0xd1: // int 16	0xd1
		elem, err = unpackInt16(in)
	case 0xd2: // int 32	0xd2
		elem, err = unpackInt32(in)
	case 0xd3: // int 64	0xd3
		elem, err = unpackInt64(in)
	case 0xda: // raw 16	0xda
		elem, err = unpackUint16(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint16))
		return unpackRaw(n, in)
	case 0xdb: // raw 32	0xdb
		elem, err = unpackUint32(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint32))
		return unpackRaw(n, in)
	case 0xdc: // array 16 0xdc
		elem, err = unpackUint16(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint16))
		return unpackArray(n, in)
	case 0xdd: // array 32 0xdd
		elem, err = unpackUint32(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint32))
		return unpackArray(n, in)
	case 0xde: // map 16	0xde
		elem, err = unpackUint16(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint16))
		return unpackMap(n, in)
	case 0xdf: // map 32	0xdf
		elem, err = unpackUint32(in)
		if err != nil { return elem, err }
		n := uint32(elem.(uint32))
		return unpackMap(n, in)
	default:
		if tag >= 0xe0 {
			elem = int(tag & (^uint8(0xe0)))-32
			return elem, err
		}
	}
	return elem, err
}

