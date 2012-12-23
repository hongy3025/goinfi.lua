package lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	//"io"
	"fmt"
	//"bytes"
	//"unsafe"
	//"strings"
	"reflect"
)

func mustBeMap(state State, lvalue int) *reflect.Value {
	var vmap *reflect.Value
	L := state.L
	ltype := C.lua_type(L, C.int(lvalue))
	if ltype == C.LUA_TUSERDATA {
		ref := C.clua_getGoRef(L, C.int(lvalue))
		if ref != nil {
			obj := (*refGo)(ref).obj
			objValue := reflect.ValueOf(obj)
			if objValue.Kind() == reflect.Map {
				vmap = &objValue
			}
		}
	}
	return vmap
}

func luaKeys(state State) int {
	L := state.L
	vmap := mustBeMap(state, 2)
	if vmap == nil {
		C.lua_pushnil(L)
		pushStringToLua(L, "Keys() only apply to `map'")
		return 2
	}

	vkeys := vmap.MapKeys()
	C.lua_createtable(L, C.int(len(vkeys)), 0)
	for i := 0; i < len(vkeys); i++ {
		if !state.goToLuaValue(vkeys[i]) {
			continue
		}
		C.lua_rawseti(L, C.int(-2), C.int(i+1))
	}

	return 1
}

func luaHasKey(state State) int {
	L := state.L
	vmap := mustBeMap(state, 2)
	if vmap == nil {
		C.lua_pushboolean(L, 0)
		pushStringToLua(L, "HasKey() only apply to `map'")
		return 2
	}
	key, err := state.luaToGoValue(3, nil)
	if err != nil {
		C.lua_pushboolean(L, 0)
		pushStringToLua(L, fmt.Sprintf("%v", err))
		return 2
	}
	if key.Type().AssignableTo(vmap.Type().Key()) {
		value := vmap.MapIndex(key)
		if value.IsValid() {
			C.lua_pushboolean(L, 1)
			return 1
		}
	}

	C.lua_pushboolean(L, 0)
	return 1
}

func lua_initGolangLib(vm *VM) {
	vm.AddFunc("golang.Keys", luaKeys)
	vm.AddFunc("golang.HasKey", luaHasKey)
}
