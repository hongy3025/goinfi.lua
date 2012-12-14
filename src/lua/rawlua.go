package lua

/*
#cgo CFLAGS: -I../../external/lua-5.1.5/src
#cgo LDFLAGS: -L../../lib -llua -lm -ldl
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"

func (state State) Pushstring(str string) {
	pushStringToLua(state.L, str)
}

