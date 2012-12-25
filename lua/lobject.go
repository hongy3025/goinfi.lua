package lua

/*
#cgo CFLAGS: -I lua-5.1.5/src -DLUA_USE_LINUX
#cgo linux LDFLAGS: -ldl
#cgo LDFLAGS: -lm
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	//"io"
	"fmt"
	//"unsafe"
	"reflect"
	//"strings"
	//"errors"
	"runtime"
)

type IRefLua interface {
	PushValue(state State)
}

type RefLua struct {
	VM  *VM
	Ref int
}

type Table struct {
	RefLua
}

type Function struct {
	RefLua
}

func (self *RefLua) init(state State, lobject int) {
	self.VM = state.VM
	C.lua_pushvalue(state.L, C.int(lobject))
	refvalue := C.luaL_ref(state.L, C.LUA_REGISTRYINDEX)
	self.Ref = int(refvalue)
	runtime.SetFinalizer(self, func(r *RefLua) {
		r.Release()
	})
}

func (state State) NewLuaRef(lobject int) *RefLua {
	r := new(RefLua)
	r.init(state, lobject)
	return r
}

func (state State) NewLuaTable(lobject int) *Table {
	t := new(Table)
	t.init(state, lobject)
	return t
}

func (state State) NewLuaFunction(lobject int) *Function {
	f := new(Function)
	f.init(state, lobject)
	return f
}

func (self *RefLua) PushValue(state State) {
	if self.Ref != 0 && self.VM.globalL != nil {
		C.lua_rawgeti(state.L, C.LUA_REGISTRYINDEX, C.int(self.Ref))
	}
}

//
// release reference to lua object
//
func (self *RefLua) Release() {
	if self.Ref != 0 && self.VM.globalL != nil {
		C.luaL_unref(self.VM.globalL, C.LUA_REGISTRYINDEX, C.int(self.Ref))
	}
	self.VM = nil
	self.Ref = 0
}

//
// call a lua function
//
func (fn *Function) Call(in ...interface{}) ([]interface{}, error) {
	if fn.Ref == 0 {
		return make([]interface{}, 0), fmt.Errorf("cannot call a released lua function")
	}
	L := fn.VM.globalL
	state := State{fn.VM, L}

	fn.PushValue(state)
	return callLuaFuncUtil(state, in, -1)
}

func (fn *Function) CallWith(in []interface{}, nout int) ([]interface{}, error) {
	if fn.Ref == 0 {
		return make([]interface{}, 0), fmt.Errorf("cannot call a released lua function")
	}
	L := fn.VM.globalL
	state := State{fn.VM, L}

	fn.PushValue(state)
	return callLuaFuncUtil(state, in, nout)
}

func (fn *Function) String() string {
	return fmt.Sprintf("<lua fuction @%v>", fn.Ref)
}

func (tbl *Table) Set(key interface{}, value interface{}) (bool, error) {
	if tbl.Ref == 0 {
		return false, fmt.Errorf("cannot set a released lua table")
	}
	L := tbl.VM.globalL
	state := State{tbl.VM, L}
	bottom := C.lua_gettop(L)
	defer C.lua_settop(L, bottom)

	tbl.PushValue(state)

	vkey := reflect.ValueOf(key)
	ok := state.goToLuaValue(vkey)
	if !ok {
		return false, fmt.Errorf("invalid key type for lua type: %v", vkey.Kind())
	}
	state.goToLuaValue(reflect.ValueOf(value))
	C.lua_settable(L, C.int(-3))

	return true, nil
}

func (tbl *Table) GetWithError(key interface{}) (interface{}, error) {
	if tbl.Ref == 0 {
		return nil, fmt.Errorf("cannot get a released lua table")
	}
	L := tbl.VM.globalL
	state := State{tbl.VM, L}
	bottom := C.lua_gettop(L)
	defer C.lua_settop(L, bottom)

	tbl.PushValue(state)

	vkey := reflect.ValueOf(key)
	ok := state.goToLuaValue(vkey)
	if !ok {
		return nil, fmt.Errorf("invalid key type for lua type: %v", vkey.Kind())
	}
	C.lua_gettable(L, C.int(-2))
	vvalue, err := state.luaToGoValue(-1, nil)
	if err != nil {
		return nil, err
	}

	if vvalue.IsValid() {
		return vvalue.Interface(), nil
	}
	return nil, nil
}

func (tbl *Table) Get(key interface{}) interface{} {
	v, _ := tbl.GetWithError(key)
	return v
}

func (tbl *Table) GetnWithError() (int, error) {
	if tbl.Ref == 0 {
		return 0, fmt.Errorf("cannot get lenght a released lua table")
	}
	L := tbl.VM.globalL
	state := State{tbl.VM, L}
	bottom := int(C.lua_gettop(L))
	defer C.lua_settop(L, C.int(bottom))

	tbl.PushValue(state)

	n := int(C.lua_objlen(L, C.int(-1)))
	return n, nil
}

func (tbl *Table) Getn() int {
	n, _ := tbl.GetnWithError()
	return n
}

func (tbl *Table) Foreach(fn func(key interface{}, value interface{}) bool) {
	if tbl.Ref == 0 {
		return
	}

	L := tbl.VM.globalL
	state := State{tbl.VM, L}
	bottom := C.lua_gettop(L)
	defer C.lua_settop(L, bottom)
	tbl.PushValue(state)

	ltable := C.lua_gettop(L)
	C.lua_pushnil(L)
	for {
		if 0 == C.lua_next(L, ltable) {
			return
		}

		vkey, err := state.luaToGoValue(-2, nil)
		if err != nil {
			return
		}
		vvalue, err := state.luaToGoValue(-1, nil)
		if err != nil {
			return
		}

		cont := fn(vkey.Interface(), vvalue.Interface())
		if !cont {
			return
		}

		C.lua_settop(L, -2)
	}
}

func (tbl *Table) String() string {
	return fmt.Sprintf("<lua table @%v>", tbl.Ref)
}
