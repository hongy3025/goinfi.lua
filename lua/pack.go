package lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	"io"
	"fmt"
	"bytes"
	"unsafe"
	//"strings"
	P "goinfi/msgpack"
)

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
	bytes := C.GoBytes(unsafe.Pointer(cs), C.int(cslen))
	return P.PackRaw(out, bytes)
}

func packLuaNil(out io.Writer, state State, object int) (n int, err error) {
	return 0, nil
}

func packLuaTable(out io.Writer, state State, object int, depth int) (n int, err error) {
	return 0, nil
}

func _packLuaObject(out io.Writer, state State, object int, depth int) (n int, err error) {
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
	return _packLuaObject(out, state, object, 0)
}

func PackLuaObjects(out io.Writer, state State, from int, to int) (ok bool, err error) {
	L := state.L
	top := C.lua_gettop(L)
	ok = true
	err = nil
	for index:=from; index<=to; index++ {
		_, err = PackLuaObject(out, state, index)
		if err != nil {
			ok = false
			break
		}
	}
	C.lua_settop(L, top)
	return ok, err
}

func luaPackToString(state State) int {
	var out bytes.Buffer
	L := state.L
	from := 1
	to := int(C.lua_gettop(state.L))
	ok, err := PackLuaObjects(&out, state, from, to)
	if !ok {
		C.lua_pushnil(L)
		pushStringToLua(L, fmt.Sprintf("%v", err))
		return 2
	}
	pushBytesToLua(L, out.Bytes())
	return 1;
}

func pack_initLua(vm *VM) {
	vm.AddFunc("pack.Pack", luaPackToString)
}
