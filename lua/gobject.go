package lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	"fmt"
	"reflect"
	"goinfi/base"
)

var theNullKeyValue base.KeyValue
var typeOfKeyValue reflect.Type = reflect.TypeOf(theNullKeyValue)

var theNullSliceKeyValue []base.KeyValue = make([]base.KeyValue, 0)
var typeOfSliceKeyValue reflect.Type = reflect.TypeOf(theNullSliceKeyValue)

func sizeOfLuaTable(L *C.lua_State, ltable int) int {
	var size int = 0
	C.lua_pushnil(L)
	for {
		if 0 == C.lua_next(L, C.int(ltable)) {
			break
		}
		size++
		C.lua_settop(L, -2) // pop 1
	}
	return size
}

func (state *State) luaTableToKeyValues(ltable int) (value reflect.Value, err error) {
	var vvalue reflect.Value
	L := state.L

	size := sizeOfLuaTable(L, ltable)
	result := make([]base.KeyValue, 0, size)

	C.lua_pushnil(L)
	for {
		if 0 == C.lua_next(L, C.int(ltable)) {
			break
		}

		lvalue := int(C.lua_gettop(L))
		lkey := lvalue - 1

		vkey, err := state.luaToGoValue(lkey, nil)
		if err != nil {
			C.lua_settop(L, -3) // pop 2
			break
		}

		if C.LUA_TTABLE == C.lua_type(L, C.int(lvalue)) {
			vvalue, err = state.luaTableToKeyValues(lvalue)
		} else {
			vvalue, err = state.luaToGoValue(lvalue, nil)
		}
		if err != nil {
			C.lua_settop(L, -3) // pop 2
			break
		}

		key := vkey.Interface()
		var skey string
		if s, ok := key.(string); ok {
			skey = s
		} else {
			skey = fmt.Sprint(key)
		}
		value := vvalue.Interface()

		result = append(result, base.KeyValue{skey, value})

		C.lua_settop(L, -2) // pop 1
	}

	return reflect.ValueOf(result), err
}
