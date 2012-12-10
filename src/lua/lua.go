package lua

/*
#cgo CFLAGS: -I/home/hongy/src/lua-5.1.5/src
#cgo LDFLAGS: -lm -ldl -L../../lib -llua
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "callback.h"
*/
import "C"
import "fmt"
import "unsafe"
import "reflect"

type State struct {
	L *C.lua_State
}

type GoFunction func(*State) int

type callbackData struct {
	l * State
	f GoFunction
}

func LuaL_newstate() *State {
	L := C.luaL_newstate()
	state := &State{ L : L }
	return state
}

func StrRawAddress(s string) (*C.char, C.size_t) {
	b := []byte(s)
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := unsafe.Pointer(h.Data)
	size := h.Len
	return (*C.char)(data), C.size_t(size)
}

//export go_callbackFromC
func go_callbackFromC(ud interface {}) int {
	cb := ud.(callbackData)
	defer func() {
		fmt.Println("quit go_callbackFromC")
		if r := recover(); r != nil {
			L := cb.l.L
			e := fmt.Sprintf("go error: %v", r)
			str, size := StrRawAddress(e)
			C.lua_pushlstring(L, str, size)
			C.lua_error(L)
		}
	}()
	return cb.f(cb.l)
}

func (L *State) Lua_cpcall(f GoFunction) int {
	var cb interface{} = callbackData{ l : L, f : f}
	p := (*C.GoIntf)(unsafe.Pointer(&cb))
	return int(C.lua_cpcall_wrap(L.L, *p))
}

func (L *State) Lua_close() {
	C.lua_close(L.L)
}

func (L *State) Lua_gc(what, data int) int {
	return int(C.lua_gc(L.L, C.int(what), C.int(data)))
}

func (L *State) LuaL_openlibs() {
	C.luaL_openlibs(L.L);
}

func Foo(i interface{}) int {
	return 0;
}

