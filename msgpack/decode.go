package msgpack

import (
	"io"
)

/*
Type Chart

Type	Binary	Hex
--------------------------------------------
Positive FixNum	0xxxxxxx	0x00 - 0x7f
FixMap	1000xxxx	0x80 - 0x8f
FixArray	1001xxxx	0x90 - 0x9f
FixRaw	101xxxxx	0xa0 - 0xbf
nil	11000000	0xc0
reserved	11000001	0xc1
false	11000010	0xc2
true	11000011	0xc3
reserved	11000100	0xc4
reserved	11000101	0xc5
reserved	11000110	0xc6
reserved	11000111	0xc7
reserved	11001000	0xc8
reserved	11001001	0xc9
float	11001010	0xca
double	11001011	0xcb
uint 8	11001100	0xcc
uint 16	11001101	0xcd
uint 32	11001110	0xce
uint 64	11001111	0xcf
int 8	11010000	0xd0
int 16	11010001	0xd1
int 32	11010010	0xd2
int 64	11010011	0xd3
reserved	11010100	0xd4
reserved	11010101	0xd5
reserved	11010110	0xd6
reserved	11010111	0xd7
reserved	11011000	0xd8
reserved	11011001	0xd9
raw 16	11011010	0xda
raw 32	11011011	0xdb
array 16	11011100	0xdc
array 32	11011101	0xdd
map 16	11011110	0xde
map 32	11011111	0xdf
Negative FixNum	111xxxxx	0xe0 - 0xff

*/

type Elem interface{}

type Msg struct {
	elems []Elem
}

type decodeState struct {
	// tagBuf [10]byte
}

type ElemPair struct {
	Key Elem
	Value Elem
}

type Map struct {
	Size int
	Elems []ElemPair
}

func newMap(n int) *Map {
	m := new(Map)
	m.Size = n
	m.Elems = make([]ElemPair, 0, n)
	return m
}

func newMsg() *Msg {
	msg := new(Msg)
	msg.elems = make([]Elem, 1)
	return msg
}

func (msg * Msg) append(elem Elem) *Msg {
	msg.elems = append(msg.elems, elem)
	return msg
}
func (state *decodeState) reset() {
	// state.tagBuf = state.underBuf
}

func (state *decodeState) readMsg(in io.Reader) (*Msg, error) {
	msg := newMsg()
	for {
		elem, err := state.readElem(in)
		if err != nil {
			return msg, err
		}
		msg.append(elem)
	}
	return nil, nil
}

func (state *decodeState) readMap(n int, in io.Reader) (Elem, error) {
	m := newMap(n)

	for i:=0; i<n; i++ {
		key, err := state.readElem(in)
		if err != nil {
			return nil, err
		}
		value, err := state.readElem(in)
		if err != nil {
			return nil, err
		}
		m.Elems = append(m.Elems, ElemPair{key, value})
	}

	return m, nil
}

func (state *decodeState) readArray(n int, in io.Reader) (Elem, error) {
	a := make([]Elem, 0, n)

	for i:=0; i<n; i++ {
		member, err := state.readElem(in)
		if err != nil {
			return nil, err
		}
		a = append(a, member)
	}

	return a, nil
}

func (state *decodeState) readElem(in io.Reader) (Elem, error) {
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
		return state.readMap(int(tag>>4), in)
	case tag <= 0x9f:
		return state.readArray(int(tag>>4), in)
	case tag <= 0xbf:
		return readRaw(int(tag>>3), in)
	}

	switch tag {
	// nil		0xc0
	case 0xc0:
		return nil, nil
	// false	0xc2
	case 0xc2:
		return false, nil
	// true	0xc3
	case 0xc3:
		return true, nil
	// float	0xca
	case 0xca:
		return readFloat32(in)
	// double	0xcb
	case 0xcb:
		return readFloat64(in)
	// uint 8	0xcc
	case 0xcc:
		value, err := readUint8(in)
		if err != nil { return nil, err }
		return value, nil
	// uint 16	0xcd
	// uint 32	0xce
	// uint 64	0xcf
	// int 8	0xd0
	// int 16	0xd1
	// int 32	0xd2
	// int 64	0xd3
	// raw 16	0xda
	case 0xda:
		n, err := readUint16(in)
		if err != nil { return nil, err }
		return readRaw(int(n), in)
	// raw 32	0xdb
	// array 16 0xdc
	// array 32 0xdd
	// map 16	0xde
	// map 32	0xdf
	}

	return 0, nil
}

