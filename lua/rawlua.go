package lua

/*
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"

func (state State) Pushstring(str string) {
	pushStringToLua(state.L, str)
}
