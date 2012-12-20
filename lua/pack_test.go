package lua

import (
	"testing"
)

func TestLua_pack(t *testing.T) {
	r := NewRunner(t)
	defer r.End()

	var result []interface{}
	var expect []interface{}

	// number

	result = r.E(`
		data = pack.Pack(0)
		n, t = pack.Unpack(data)
		return t
	`)
	r.AssertEqual(result[0], float64(0))

	result = r.E(`
		data = pack.Pack(-1.5)
		n, t = pack.Unpack(data)
		return t
	`)
	r.AssertEqual(result[0], -1.5)

	// string
	result = r.E(`
		data = pack.Pack('hello world')
		n, t = pack.Unpack(data)
		return t
	`)
	r.AssertEqual(result[0], "hello world")

	// num and string
	result = r.E(`
		data = pack.Pack(1, 'hello world', 3)
		n, a, b, c = pack.Unpack(data)
		return a, b, c
	`)
	expect = []interface{}{ 1.0, "hello world", 3.0 }
	r.AssertEqual(result, expect)

	// lua array
	result = r.E(`
		data = pack.Pack({'foo', 'bar', 100 })
		n, t = pack.Unpack(data)
		return t[1], t[2], t[3]
	`)
	expect = []interface{}{ "foo", "bar", 100.0 }
	r.AssertEqual(result, expect)

	// lua table mix array
	result = r.E(`
		data = pack.Pack(1, 2, {a=1,b=2,c=3, 1, 2})
		t = { pack.Unpack(data) }
		return t[2], t[3], t[4].a, t[4].b, t[4].c, t[4][1], t[4][2]
	`)
	expect = []interface{}{ 1.0, 2.0, 1.0, 2.0, 3.0, 1.0, 2.0 }
	r.AssertEqual(result, expect)

	// depth table
	result = r.E(`
		data = pack.Pack({a={b={c={d={e='depth'}}}}})
		n, t = pack.Unpack(data)
		return t.a.b.c.d.e
	`)
	r.AssertEqual(result[0], "depth")
}

