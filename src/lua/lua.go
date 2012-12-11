package lua

/*
#cgo CFLAGS: -I../../external/lua-5.1.5/src
#cgo LDFLAGS: -L../../lib -llua -lm -ldl
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "golua.h"
*/
import "C"

import "fmt"
import "unsafe"
import "reflect"

type refNode struct {
	prev *refNode
	next *refNode
	state *State
	obj interface{}
}

func (self * refNode) link(head * refNode) {
	fmt.Printf("link,%v,%v\n", reflect.TypeOf(self.obj).Kind(), self.obj)
	self.next = head.next
	head.next = self
	self.prev = head
	if self.next != nil {
		self.next.prev = self
	}
}

func (self * refNode) unlink() {
	fmt.Printf("unlink,%v,%v\n", reflect.TypeOf(self.obj).Kind(), self.obj)
	self.prev.next = self.next
	if self.next != nil {
		self.next.prev = self.prev
	}
	self.prev = nil
	self.next = nil
}

type State struct {
	L *C.lua_State
	refLink refNode
}

type cstring struct {
	s *C.char
	n C.size_t
}

type GoFunc func(*State) int

type callbackData struct {
	state * State
	fn GoFunc
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func LuaL_newstate() *State {
	L := C.luaL_newstate()
	C.clua_initState(L)
	state := &State{ L : L }
	return state
}

func stringToC(s string) cstring {
	b := []byte(s)
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := unsafe.Pointer(h.Data)
	size := h.Len
	return cstring { (*C.char)(data), C.size_t(size) }
}

func goToLuaValue(L *C.lua_State, value reflect.Value) {
	switch vk := value.Kind(); vk {
		case reflect.Int:
			v := value.Int()
			C.lua_pushinteger(L, C.lua_Integer(v))
			return
	}
	C.lua_pushnil(L)
}

func luaToGoValue(L *C.lua_State, _lvalue int, goType reflect.Type) (reflect.Value, error) {
	lvalue := C.int(_lvalue)
	lt := C.lua_type(L, lvalue)
	if lt == C.LUA_TNUMBER {
		switch gk := goType.Kind(); gk {
		case reflect.Int:
			var vInt int
			v := reflect.ValueOf(&vInt).Elem()
			v.SetInt(int64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Int8:
			var vInt8 int8
			v := reflect.ValueOf(&vInt8).Elem()
			v.SetInt(int64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Int16:
			var vInt16 int16
			v := reflect.ValueOf(&vInt16).Elem()
			v.SetInt(int64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Int32:
			var vInt32 int32
			v := reflect.ValueOf(&vInt32).Elem()
			v.SetInt(int64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Int64:
			var vInt64 int64
			v := reflect.ValueOf(&vInt64).Elem()
			v.SetInt(int64(C.lua_tointeger(L, lvalue)))
			return v, nil

		case reflect.Uint:
			var vUInt uint
			v := reflect.ValueOf(&vUInt).Elem()
			v.SetUint(uint64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Uint8:
			var vUInt8 uint8
			v := reflect.ValueOf(&vUInt8).Elem()
			v.SetUint(uint64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Uint16:
			var vUInt16 uint16
			v := reflect.ValueOf(&vUInt16).Elem()
			v.SetUint(uint64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Uint32:
			var vUInt32 uint32
			v := reflect.ValueOf(&vUInt32).Elem()
			v.SetUint(uint64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Uint64:
			var vUInt64 uint64
			v := reflect.ValueOf(&vUInt64).Elem()
			v.SetUint(uint64(C.lua_tointeger(L, lvalue)))
			return v, nil

		case reflect.Float32:
			var vFloat32 float32
			v := reflect.ValueOf(&vFloat32).Elem()
			v.SetFloat(float64(C.lua_tointeger(L, lvalue)))
			return v, nil
		case reflect.Float64:
			var vFloat64 float64
			v := reflect.ValueOf(&vFloat64).Elem()
			v.SetFloat(float64(C.lua_tointeger(L, lvalue)))
			return v, nil
		}
	}
	return reflect.ValueOf(nil), nil
}


//export go_callbackFromC
func go_callbackFromC(ud interface {}) int {
	cb := ud.(callbackData)
	return cb.fn(cb.state)
}

//export go_unlinkObject
func go_unlinkObject(ref unsafe.Pointer) {
	node := (*refNode)(ref)
	node.unlink()
}

//export go_callObject
func go_callObject(ref unsafe.Pointer) int {
	node := (*refNode)(ref)
	v := reflect.ValueOf(node.obj)
	state := node.state
	L := state.L

	inlua := int(C.lua_gettop(L)) - 1
	ingo := v.Type().NumIn()
	n := min(ingo, inlua)

	in := make([]reflect.Value, n)
	for i:=0; i<n; i++ {
		tin := v.Type().In(i)
		value, err := luaToGoValue(L, i+2, tin)
		if err != nil {
			pushStringToLua(L, err.Error())
			return -1
		}
		in[i] = value
	}

	out := v.Call(in)

	for _, value := range out {
		goToLuaValue(L, value)
	}

	 return len(out)
}


func (state *State) Cpcall(fn GoFunc) int {
	var cb interface{} = callbackData{ state : state, fn : fn }
	p := (*C.GoIntf)(unsafe.Pointer(&cb))
	return int(C.clua_goPcall(state.L, *p))
}

func (state *State) Dostring(str string) int {
	cstr := stringToC(str)
	L := state.L
	ret := int(C.luaL_loadbuffer(L, cstr.s, cstr.n, nil))
	if ret != 0 {
		return ret
	}
	ret = int(C.lua_pcall(L, 0, C.LUA_MULTRET, 0))
	return ret
}

func (state *State) Close() {
	C.lua_close(state.L)
}

func pushStringToLua(L *C.lua_State, str string) {
	cstr := stringToC(str)
	C.lua_pushlstring(L, cstr.s, cstr.n)
}

func (state *State) Pushstring(str string) {
	pushStringToLua(state.L, str)
}

/*
never call this
func (state *State) Error() {
	C.lua_error(state.L)
}
*/

func (state *State) Gc(what, data int) int {
	return int(C.lua_gc(state.L, C.int(what), C.int(data)))
}

func (state *State) Openlibs() {
	C.luaL_openlibs(state.L);
}

func (state * State) newRefNode(obj interface{}) * refNode {
	ref := new(refNode)
	ref.state = state
	ref.obj = obj
	ref.link(&state.refLink)
	return ref
}

func (state *State) pushObjToLua(obj interface{}) {
	ref := state.newRefNode(obj)
	v := reflect.ValueOf(obj)
	kindStr := v.Kind().String()
	cstr := stringToC(kindStr)
	C.clua_newGoRefUd(state.L, unsafe.Pointer(ref), cstr.s, cstr.n)
}

func (state * State) NewLuaFunc(name string, fn interface{}) (bool, error) {
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func {
		return false, fmt.Errorf("fn must be a function")
	}
	state.Pushstring(name)
	state.pushObjToLua(fn)
	C.lua_settable(state.L, C.LUA_GLOBALSINDEX)

	return true, nil
}

/*
func (state * State) NewLuaModule(string mod, members []interface{}) result bool, err string {
	return true, nil
}*/

