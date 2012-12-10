#include <lua.h>
#include "callback.h"

static int lua_cpcall_callback(lua_State * L) {
	GoIntf *ud = (GoIntf*)lua_touserdata(L, 1);
	return go_callbackFromC(*ud);
}

int lua_cpcall_wrap(lua_State *L, GoIntf cb) {
	return lua_cpcall(L, lua_cpcall_callback, &cb);
}

