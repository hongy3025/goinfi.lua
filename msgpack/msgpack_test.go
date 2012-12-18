package msgpack

import (
	"testing"
	"fmt"
	"bytes"
	"reflect"
)

var Println = fmt.Println
var TestingShort = testing.Short

func getBuf() []byte {
	return make([]byte, 4096)
}

func checkVarint(t *testing.T, x int64, xlen int) {
	ubuf := getBuf()
	var buf []byte
	var n int

	buf = ubuf[:]
	n = PackVarint(buf, x)
	if n != xlen {
		t.Errorf("pack varint, x=%v, xlen=%v, n=%v", x, xlen, n)
	}
	buf = buf[:n]
	in := bytes.NewReader(buf)

	var u Unpacker
	msg, _ := u.Unpack(in)
	o := reflect.ValueOf(msg.Elems[0]).Int()
	if x != o {
		t.Errorf("unpack varint, x=%v, xlen=%v, o=%v", x, xlen, o)
	}
}

func TestUnpackMsg_varint(t *testing.T) {
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

