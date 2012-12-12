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

import (
	"fmt"
	"unsafe"
	"reflect"
)

type refNode struct {
	prev *refNode
	next *refNode
	state *State
	obj interface{}
}

func (self * refNode) link(head * refNode) {
	//fmt.Printf("link,%v,%v\n", reflect.TypeOf(self.obj).Kind(), self.obj)
	self.next = head.next
	head.next = self
	self.prev = head
	if self.next != nil {
		self.next.prev = self
	}
}

func (self * refNode) unlink() {
	//fmt.Printf("unlink,%v,%v\n", reflect.TypeOf(self.obj).Kind(), self.obj)
	self.prev.next = self.next
	if self.next != nil {
		self.next.prev = self.prev
	}
	self.prev = nil
	self.next = nil
}

type field struct {
	ti *typeInfo
	idx int
	isdata bool
}

type typeInfo struct {
	typ reflect.Type
	fields map[string]field
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

func (state *State) findTypeInfo(typ reflect.Type) * typeInfo {
}

func indexStruct(state *State, structPtr reflect.Value, lidx C.int) (ret int, err error) {
	L := state.L
	structValue := structPtr.Elem()
	t := structValue.Type()
	info := state.findTypeInfo(t)
	if info == nil {
		return -1, fmt.Errorf("can not index a solid struct")
	}
	ltype := int(C.lua_type(L, lidx))
	if ltype != C.LUA_TSTRING {
		return -1, fmt.Errorf("field key of struct must be a string")
	}
	key := stringFromLua(L, lidx)
	fld := info.findFieldByName(key)
	if fld == nil {
		return -1, fmt.Errorf("not such field `%v'", key)
	}
	value := getStructFieldValue(structValue, fld)
	state.goToLuaValue(value)
	return 1
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

//export go_indexObject
func go_indexObject(ref unsafe.Pointer, lidx C.int) ret int {
	node := (*refNode)(ref)
	state := node.state
	L := state.L
	v := reflect.ValueOf(node.obj)
	t := v.Type()
	k := v.Kind()

	defer func() {
		if err := recover(); err != nil {
			pushStringToLua(L, err.Error())
			ret = -1
		}
	}()

	/*
	if k == reflect.Ptr && v.Type().Elem().Kind() == reflect.Struct {
		indexable = true
	}
	*/

	ltype := int(C.lua_type(L, lidx))
	switch k {
	case reflect.Slice:
		if ltype == C.LUA_TNUMBER {
			idx := int(C.lua_tointeger(L, lidx))
			value = v.Index(idx).Interface()
			state.goToLuaValue(value)
			return 1
		}
		panic(fmt.Sprintf("index of slice must be a number type, here got `%v'", luaTypeName(ltype)))
	case reflect.Map:
		keyType := t.Key()
		key, err := state.luaToGoValue(int(lidx), keyType)
		if err != nil {
			panic(fmt.Sprintf("index type of map must be type `%v', %s", keyType.Kind(), err.Error()))
		}
		value := v.MapIndex(key)
		if !value.IsValid() {
			C.lua_pushnil(L)
			C.lua_pushboolean(L, 0)
			return 2
		}
		state.goToLuaValue(value)
		C.lua_pushboolean(L, 1)
		return 2
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Struct {
			ret, err := indexStruct(state, v, lidx)
			if err != nil {
				panic(fmt.Sprintf("get field of struct fail, %s", err.Error()))
			}
			return ret
		}
	}

	panic(fmt.Sprintf("try to index a non-indexable go object, type `%v'", k))
}

//export go_callObject
func go_callObject(ref unsafe.Pointer) int {
	node := (*refNode)(ref)
	state := node.state
	L := state.L

	v := reflect.ValueOf(node.obj)
	k := v.Kind()
	if k != reflect.Func {
		pushStringToLua(L, fmt.Sprintf("try to call a non-function go object, type `%v'", k))
		return -1
	}

	inlua := int(C.lua_gettop(L)) - 1
	ingo := v.Type().NumIn()
	n := min(ingo, inlua)

	in := make([]reflect.Value, n)
	for i:=0; i<n; i++ {
		tin := v.Type().In(i)
		value, err := state.luaToGoValue(i+2, tin)
		if err != nil {
			pushStringToLua(L, fmt.Sprintf("call go func: arg %v,", i) + err.Error())
			return -1
		}
		in[i] = value
	}

	out := v.Call(in)

	for _, value := range out {
		state.goToLuaValue(value)
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

func (state *State) AddFunc(name string, fn interface{}) (bool, error) {
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func {
		return false, fmt.Errorf("fn must be a function")
	}
	state.Pushstring(name)
	state.pushObjToLua(fn)
	C.lua_settable(state.L, C.LUA_GLOBALSINDEX)

	return true, nil
}

func (state *State) AddStructs(structs interface{}) (bool, error) {
	contain := reflect.TypeOf(structs)
}

/*
func (state * State) NewLuaModule(string mod, members []interface{}) result bool, err string {
	return true, nil
}*/

