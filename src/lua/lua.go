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
//import "fmt"
import "unsafe"
import "reflect"

type State struct {
	L *C.lua_State
}

type GoFunction func(*State) int

type callbackData struct {
	state * State
	fn GoFunction
}

func LuaL_newstate() *State {
	L := C.luaL_newstate()
	state := &State{ L : L }
	return state
}

func stringRawAddress(s string) (*C.char, C.size_t) {
	b := []byte(s)
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := unsafe.Pointer(h.Data)
	size := h.Len
	return (*C.char)(data), C.size_t(size)
}

//export go_callbackFromC
func go_callbackFromC(ud interface {}) int {
	cb := ud.(callbackData)
	/*
	defer func() {
		if r := recover(); r != nil {
			L := cb.state.L
			e := fmt.Sprintf("go error: %v", r)
			rstr, size := stringRawAddress(e)
			C.lua_pushlstring(L, rstr, size)
			C.lua_error(L)
		}
	}()
	*/
	return cb.fn(cb.state)
}

func (state *State) Cpcall(fn GoFunction) int {
	var cb interface{} = callbackData{ state : state, fn : fn }
	p := (*C.GoIntf)(unsafe.Pointer(&cb))
	return int(C.lua_cpcall_wrap(state.L, *p))
}

func (state *State) Dostring(str string) int {
	rstr, size := stringRawAddress(str)
	L := state.L
	ret := int(C.luaL_loadbuffer(L, rstr, size, rstr))
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
	rstr, size := stringRawAddress(str)
	C.lua_pushlstring(state.L, rstr, size)
}

/*
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


