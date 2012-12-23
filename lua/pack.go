package lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"io"
	"unsafe"
	//"strings"
	P "goinfi/msgpack"
	"reflect"
)

const MAX_PACK_DEPTH = 50

func packLuaNumber(out io.Writer, state State, object int) (n int, err error) {
	value := float64(C.lua_tonumber(state.L, C.int(object)))
	i := int64(value)
	if float64(i) == value {
		return P.PackVarint(out, i)
	}
	return P.PackFloat64(out, value)
}

func packLuaBoolean(out io.Writer, state State, object int) (n int, err error) {
	b := int(C.lua_toboolean(state.L, C.int(object)))
	if b == 0 {
		return P.PackBoolean(out, false)
	}
	return P.PackBoolean(out, true)
}

func packLuaString(out io.Writer, state State, object int) (n int, err error) {
	var cslen C.size_t
	cs := C.lua_tolstring(state.L, C.int(object), &cslen)

	// <HACK> pretend a []byte slice to avoid intermediate buffer
	slhead := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cs)),
		Len:  int(cslen), Cap: int(cslen),
	}
	pslice := (*[]byte)(unsafe.Pointer(&slhead))
	// </HACK>

	return P.PackRaw(out, *pslice)
}

func packLuaNil(out io.Writer, state State, object int) (n int, err error) {
	return P.PackNil(out)
}

func packLuaTable(out io.Writer, state State, object int, depth int) (n int, err error) {
	depth++
	L := state.L

	if depth > MAX_PACK_DEPTH {
		return 0, fmt.Errorf("pack too depth, depth=%v", depth)
	}

	n = 0
	err = nil
	var mapSize int = 0
	C.lua_pushnil(L)
	for {
		if 0 == C.lua_next(L, C.int(object)) {
			break
		}
		mapSize++
		C.lua_settop(L, -2) // pop 1
	}

	var ni int
	ni, err = P.PackMapHead(out, uint32(mapSize))
	n += ni
	if err != nil {
		return
	}

	C.lua_pushnil(L)
	for {
		if 0 == C.lua_next(L, C.int(object)) {
			break
		}

		top := int(C.lua_gettop(L))
		// key
		ni, err = packLuaObject(out, state, top-1, depth)
		n += ni
		if err != nil {
			C.lua_settop(L, -3) // pop 2
			return
		}
		// value
		ni, err = packLuaObject(out, state, top, depth)
		n += ni
		if err != nil {
			C.lua_settop(L, -3) // pop 2
			return
		}
		C.lua_settop(L, -2) // removes value, keeps key for next iteration
	}

	return
}

func packLuaObject(out io.Writer, state State, object int, depth int) (n int, err error) {
	L := state.L
	ltype := C.lua_type(L, C.int(object))
	switch ltype {
	case C.LUA_TNUMBER:
		n, err = packLuaNumber(out, state, object)
	case C.LUA_TBOOLEAN:
		n, err = packLuaBoolean(out, state, object)
	case C.LUA_TSTRING:
		n, err = packLuaString(out, state, object)
	case C.LUA_TNIL:
		n, err = packLuaNil(out, state, object)
	case C.LUA_TTABLE:
		n, err = packLuaTable(out, state, object, depth)
	case C.LUA_TUSERDATA:
		fallthrough
	case C.LUA_TTHREAD:
		fallthrough
	case C.LUA_TLIGHTUSERDATA:
		fallthrough
	default:
		typeName := luaTypeName(ltype)
		return n, fmt.Errorf("cannot pack lua type `%v'", typeName)
	}
	return n, err
}

func PackLuaObject(out io.Writer, state State, object int) (n int, err error) {
	return packLuaObject(out, state, object, 0)
}

func PackLuaObjects(out io.Writer, state State, from int, to int) (ok bool, err error) {
	L := state.L
	top := C.lua_gettop(L)
	ok = true
	err = nil
	for object := from; object <= to; object++ {
		_, err = PackLuaObject(out, state, object)
		if err != nil {
			ok = false
			break
		}
	}
	C.lua_settop(L, top)
	return ok, err
}

func luaPackToString(state State) int {
	// arg 1 is func udata itself
	var out bytes.Buffer
	L := state.L
	from := 2
	to := int(C.lua_gettop(state.L))
	ok, err := PackLuaObjects(&out, state, from, to)
	if !ok {
		C.lua_pushnil(L)
		pushStringToLua(L, fmt.Sprintf("%v", err))
		return 2
	}
	pushBytesToLua(L, out.Bytes())
	return 1
}

func unpackMapToLua(state State, m *P.Map) {
	L := state.L
	n := len(m.Elems)

	C.lua_createtable(L, 0, 0)

	for i := 0; i < n; i++ {
		key := m.Elems[i].Key
		value := m.Elems[i].Value

		// key
		switch key.(type) {
		case int, int8, int32, int64, uint, uint8, uint32, uint64, float32, float64, string:
			UnpackObjectToLua(state, key)
		case []byte:
			bytes := key.([]byte)
			pushBytesToLua(L, bytes)
		default:
			continue
		}

		// value
		UnpackObjectToLua(state, value)

		C.lua_settable(L, -3)
	}
}

func unpackArrayToLua(state State, a *P.Array) {
	L := state.L
	n := len(a.Elems)
	C.lua_createtable(L, C.int(n), 0)
	for i := 0; i < n; i++ {
		UnpackObjectToLua(state, a.Elems[i])
		C.lua_rawseti(L, C.int(-2), C.int(i+1))
	}
}

func UnpackObjectToLua(state State, elem P.Elem) {
	L := state.L
	switch elem.(type) {
	case *P.Map:
		m := elem.(*P.Map)
		unpackMapToLua(state, m)
	case *P.Array:
		a := elem.(*P.Array)
		unpackArrayToLua(state, a)
	case []byte:
		bytes := elem.([]byte)
		pushBytesToLua(L, bytes)
	default:
		value := reflect.ValueOf(elem)
		state.goToLuaValue(value)
	}
}

func UnpackToLua(in io.Reader, state State) int {
	L := state.L
	var n int = 0
	var unpack P.Unpacker
	msg, err := unpack.Unpack(in)
	if err != nil {
		C.lua_pushinteger(L, C.lua_Integer(n))
		pushStringToLua(L, fmt.Sprintf("unpack error: %v", err))
		return 2
	}
	n = len(msg.Elems)
	C.lua_pushinteger(L, C.lua_Integer(n))
	for _, elem := range msg.Elems {
		UnpackObjectToLua(state, elem)
	}
	return n + 1
}

func luaUnpackFromString(state State) int {
	// arg 1 is func udata itself
	var cslen C.size_t
	cs := C.lua_tolstring(state.L, C.int(2), &cslen)

	// <HACK> pretend a []byte slice to avoid intermediate buffer
	slhead := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cs)),
		Len:  int(cslen), Cap: int(cslen),
	}
	pslice := (*[]byte)(unsafe.Pointer(&slhead))
	// </HACK>

	reader := bytes.NewReader(*pslice)
	return UnpackToLua(reader, state)
}

func lua_initPackLib(vm *VM) {
	vm.AddFunc("pack.Pack", luaPackToString)
	vm.AddFunc("pack.Unpack", luaUnpackFromString)
}
