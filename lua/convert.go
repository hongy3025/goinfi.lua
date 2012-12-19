package lua

/*
#cgo CFLAGS: -I lua-5.1.5/src
#cgo LDFLAGS: -L ../../../lib -llua -lm -ldl
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
	"reflect"
)

type l2gGetConvertError func(ltype C.int, gkind reflect.Kind) error

var luaT_typenames = []string {
	"nil", "boolean", "userdata", "number",
	"string", "table", "function", "userdata", "thread",
	"proto", "upval",
}

func luaTypeName(ltype C.int) string {
	return luaT_typenames[int(ltype)]
}

/*
type cstring struct {
	s *C.char
	n C.size_t
}
*/

func stringToC(str string) (*C.char, C.size_t) {
	// <HACK> get address and length of go string, to avoid two-times copy
	pstr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	s := (*C.char)(unsafe.Pointer(pstr.Data))
	n := C.size_t(pstr.Len)
	// </HACK>
	return s, n
}

func stringFromLua(L *C.lua_State, lvalue C.int) string {
	var cslen C.size_t
	cs := C.lua_tolstring(L, lvalue, &cslen)
	return C.GoStringN(cs, C.int(cslen))
}

func pushStringToLua(L *C.lua_State, str string) {
	s, n := stringToC(str)
	C.lua_pushlstring(L, s, n)
}

func pushBytesToLua(L *C.lua_State, bytes []byte) {
	h := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	data := unsafe.Pointer(h.Data)
	size := h.Len
	C.lua_pushlstring(L, (*C.char)(data), C.size_t(size))
}

func (state State) pushObjToLua(obj interface{}) {
	ref := state.VM.newRefNode(obj)
	C.clua_newGoRefUd(state.L, unsafe.Pointer(ref))
}

func (state State) goToLuaValue(value reflect.Value) {
	L := state.L
	gkind := value.Kind()
	switch gkind {
	case reflect.Bool:
		v:= value.Bool()
		if v {
			C.lua_pushboolean(L, 1)
		} else {
			C.lua_pushboolean(L, 0)
		}
		return
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := value.Int()
		C.lua_pushinteger(L, C.lua_Integer(v))
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := value.Uint()
		C.lua_pushinteger(L, C.lua_Integer(v))
		return
	// case reflect.Uintptr:
	case reflect.Float32, reflect.Float64:
		v := value.Float()
		C.lua_pushnumber(L, C.lua_Number(v))
		return

	// case reflect.Array:
	// case reflect.Complex64, reflect.Complex128:
	case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		state.pushObjToLua(value.Interface())
		return
	case reflect.String:
		v := value.String()
		pushStringToLua(L, v)
		return
	case reflect.Struct:
		objPtr := reflect.New(value.Type())
		objPtr.Elem().Set(value)
		state.pushObjToLua(objPtr.Interface())
		return
	//case reflect.UnsafePointer
	}
	C.lua_pushnil(L)
}

func (state State) luaToGoValue(_lvalue int, outType *reflect.Type) (reflect.Value, error) {
	L := state.L
	lvalue := C.int(_lvalue)
	ltype := C.lua_type(L, lvalue)
	gkind := reflect.Invalid
	if outType != nil {
		gkind = (*outType).Kind()
	}
	switch ltype {
	case C.LUA_TNIL:
		switch gkind {
		case reflect.Invalid, reflect.Func, reflect.Ptr:
			return reflect.ValueOf(nil), nil
		}
	case C.LUA_TBOOLEAN:
		switch gkind {
		case reflect.Invalid, reflect.Bool:
			cv := C.lua_toboolean(L, lvalue)
			var v bool
			if cv == 0 {
				v = false
			} else {
				v = true
			}
			return reflect.ValueOf(v), nil
		}
	//case C.LUA_TLIGHTUSERDATA:
	case C.LUA_TNUMBER:
		switch gkind {
		case reflect.Int:
			v := int(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Int8:
			v := int8(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Int16:
			v := int16(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Int32:
			v := int32(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Int64:
			v := int64(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil

		case reflect.Uint:
			v := uint(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Uint8:
			v := uint8(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Uint16:
			v := uint16(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Uint32:
			v := uint32(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil
		case reflect.Uint64:
			v := uint64(C.lua_tointeger(L, lvalue))
			return reflect.ValueOf(v), nil

		case reflect.Float32:
			v := float32(C.lua_tonumber(L, lvalue))
			return reflect.ValueOf(v), nil

		case reflect.Invalid, reflect.Float64:
			v := float64(C.lua_tonumber(L, lvalue))
			return reflect.ValueOf(v), nil
		}
	case C.LUA_TSTRING:
		switch gkind {
		case reflect.Invalid, reflect.String:
			v := stringFromLua(L, lvalue)
			return reflect.ValueOf(v), nil
		}
	//case C.LUA_TTABLE:
	//case C.LUA_TFUNCTION:
	case C.LUA_TUSERDATA:
		ref := C.clua_getGoRef(L, lvalue)
		if ref != nil {
			obj := (*refNode)(ref).obj
			objType := reflect.TypeOf(obj)
			objValue := reflect.ValueOf(obj)
			if gkind == reflect.Invalid {
				return reflect.ValueOf(obj), nil
			} else if objType == *outType {
				return reflect.ValueOf(obj), nil
			} else if objType.Kind() == reflect.Ptr {
				if objType.Elem() == *outType {
					return objValue.Elem(), nil
				}
			}
		}
	//case C.LUA_TTHREAD:
	}
	return reflect.ValueOf(nil),
		fmt.Errorf("cannot convert from lua-type `%v' to go-type `%v'",
			luaTypeName(ltype), gkind)
}

