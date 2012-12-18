package msgpack

import (
	"testing"
	"fmt"
	"bytes"
)

var Println = fmt.Println
var TestingShort = testing.Short

func getInt64(e Elem) int64 {
	switch e.(type) {
	case int:
		return int64(e.(int))
	case int8:
		return int64(e.(int8))
	case int16:
		return int64(e.(int16))
	case int32:
		return int64(e.(int32))
	case int64:
		return e.(int64)
	case uint:
		return int64(e.(uint))
	case uint8:
		return int64(e.(uint8))
	case uint16:
		return int64(e.(uint16))
	case uint32:
		return int64(e.(uint32))
	case uint64:
		return int64(e.(uint64))
	}
	return 0
}

func getUint64(e Elem) uint64 {
	switch e.(type) {
	case int:
		return uint64(e.(int))
	case int8:
		return uint64(e.(int8))
	case int16:
		return uint64(e.(int16))
	case int32:
		return uint64(e.(int32))
	case int64:
		return e.(uint64)
	case uint:
		return uint64(e.(uint))
	case uint8:
		return uint64(e.(uint8))
	case uint16:
		return uint64(e.(uint16))
	case uint32:
		return uint64(e.(uint32))
	case uint64:
		return uint64(e.(uint64))
	}
	return 0
}

func checkVarint(t *testing.T, x int64, olen int) {
	var buf bytes.Buffer

	n, err := PackVarint(&buf, x)
	if err != nil {
		t.Errorf("pack varint, %v", err)
		return
	}

	if n != olen {
		t.Errorf("pack varint, x=%v, olen=%v, n=%v", x, olen, n)
		return
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack varint, %v", err)
		return
	}
	o := getInt64(msg.Elems[0])
	if x != o {
		t.Errorf("unpack varint, x=%v, olen=%v, o=%v", x, olen, o)
	}
}

func checkVaruint(t *testing.T, x uint64, olen int) {
	var buf bytes.Buffer

	n, err := PackVaruint(&buf, x)
	if err != nil {
		t.Errorf("pack varint, %v", err)
		return
	}

	if n != olen {
		t.Errorf("pack varuint, x=%v, olen=%v, n=%v", x, olen, n)
		return
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack varuint, %v", err)
		return
	}
	o := getUint64(msg.Elems[0])
	if x != o {
		t.Errorf("unpack varuint, x=%v, olen=%v, o=%v", x, olen, o)
	}
}

func checkRaw(t *testing.T, data[]byte, olen int) {
	var buf bytes.Buffer
	n, err := PackRaw(&buf, data)
	if err != nil {
		t.Errorf("pack raw, %v", err)
		return
	}
	if n != olen {
		t.Errorf("pack raw, n=%v, olen=%v", n, olen)
		return
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack raw, %v", err)
		return
	}
	outdata := (msg.Elems[0]).([]byte)
	bytes.Equal(data, outdata)
}

func TestMsgpack_varint(t *testing.T) {
	checkVarint(t, 0, 1)
	checkVarint(t, 127, 1)
	checkVarint(t, -1, 1)
	checkVarint(t, -32, 1)

	checkVarint(t, 1<<7, 3)
	checkVarint(t, 1<<7+1, 3)

	checkVarint(t, 1<<8-1, 3)
	checkVarint(t, 1<<8, 3)
	checkVarint(t, 1<<8+1, 3)

	checkVarint(t, 1<<15-1, 3)
	checkVarint(t, 1<<15, 5)
	checkVarint(t, 1<<15+1, 5)

	checkVarint(t, 1<<16-1, 5)
	checkVarint(t, 1<<16, 5)
	checkVarint(t, 1<<16+1, 5)

	checkVarint(t, 1<<31-1, 5)
	checkVarint(t, 1<<31, 9)
	checkVarint(t, 1<<31+1, 9)
}

func TestMsgpack_varuint(t *testing.T) {
	checkVaruint(t, 0, 1)
	checkVaruint(t, 127, 1)

	checkVaruint(t, 1<<8-1, 2)
	checkVaruint(t, 1<<8, 3)
	checkVaruint(t, 1<<8+1, 3)

	checkVaruint(t, 1<<16-1, 3)
	checkVaruint(t, 1<<16, 5)
	checkVaruint(t, 1<<16+1, 5)

	checkVaruint(t, 1<<32-1, 5)
	checkVaruint(t, 1<<32, 9)
	checkVaruint(t, 1<<32+1, 9)
}

func TestMsgpack_raw(t *testing.T) {
	s := "hello world"
	checkRaw(t, []byte(s), len(s)+1)

	d16 := make([]byte, 4097)
	checkRaw(t, d16, len(d16)+3)

	d32 := make([]byte, 65537)
	checkRaw(t, d32, len(d32)+5)
}

func TestMsgpack_array(t *testing.T) {
	var buf bytes.Buffer

	A := make([]int32, 100)
	for i:=0; i<len(A); i++ {
		A[i] = int32(i)
	}
	PackArrayHead(&buf, uint32(len(A)))
	for _, elem := range A {
		PackInt32(&buf, int32(elem))
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack array, %v", err)
		return
	}

	elems := (msg.Elems[0]).([]Elem)
	if len(elems) != len(A) {
		t.Errorf("unpack array length", len(elems), len(A))
		return
	}

	for i, elem := range elems {
		if A[i] != elem.(int32) {
			t.Errorf("unpack array, i=%v, A=%v, elem=%v", i, A[i], elem)
			return
		}
	}
}

func TestMsgpack_map(t *testing.T) {
	var buf bytes.Buffer

	K := make([]string, 100)
	V := make([]int32, 100)
	for i:=0; i<len(K); i++ {
		K[i] = fmt.Sprintf("key%v", i)
		V[i] = int32(i)
	}
	PackMapHead(&buf, uint32(len(K)))
	for i:=0; i<len(K); i++ {
		PackRaw(&buf, []byte(K[i]))
		PackInt32(&buf, int32(V[i]))
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack map, %v", err)
		return
	}

	M := (msg.Elems[0]).(*Map)
	if len(K) != int(M.Size) {
		t.Errorf("unpack map length", len(K), M.Size)
		return
	}

	for i:=0; i<int(M.Size); i++ {
		key := string(M.Elems[i].Key.([]byte))
		if K[i] != key {
			t.Errorf("unpack map key", i, K[i], key)
			return
		}
		value := M.Elems[i].Value.(int32)
		if V[i] != value {
			t.Errorf("unpack map value", i, value)
			return
		}
	}
}

func TestMsgpack_nil(t *testing.T) {
	var buf bytes.Buffer
	PackNil(&buf)

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack nil, %v", err)
		return
	}
	if !IsNil(msg.Elems[0]) {
		t.Errorf("unpack nil")
		return
	}
}

func TestMsgpack_bool(t *testing.T) {
	var buf bytes.Buffer
	PackBoolean(&buf, true)
	PackBoolean(&buf, false)

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack bool, %v", err)
		return
	}
	if true != (msg.Elems[0]).(bool) {
		t.Errorf("unpack true")
		return
	}
	if false != (msg.Elems[1]).(bool) {
		t.Errorf("unpack false")
		return
	}
}

func TestMsgpack_float32(t *testing.T) {
	var buf bytes.Buffer
	var F = []float32{1.0, 1.5, 32.5, -1.0, -1.5, -32.5}
	for _, value := range F {
		PackFloat32(&buf, value)
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack float32, %v", err)
		return
	}
	for i:=0; i<len(F); i++ {
		o := (msg.Elems[i]).(float32)
		if F[i] != o {
			t.Errorf("unpack float32", i, F[i], o)
		}
	}
}

func TestMsgpack_float64(t *testing.T) {
	var buf bytes.Buffer
	var F = []float64{1.0, 1.5, 32.5, -1.0, -1.5, -32.5}
	for _, value := range F {
		PackFloat64(&buf, value)
	}

	in := &buf
	var u Unpacker
	msg, err := u.Unpack(in)
	if err != nil {
		t.Errorf("unpack float64, %v", err)
		return
	}
	for i:=0; i<len(F); i++ {
		o := (msg.Elems[i]).(float64)
		if F[i] != o {
			t.Errorf("unpack float64", i, F[i], o)
		}
	}
}

