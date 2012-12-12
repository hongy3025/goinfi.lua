#include <string.h>
#include <lua.h>
#include <lauxlib.h>
#include "golua.h"

#define GO_UDATA_META_NAME "go.udata"

static void * clua_getudata(lua_State *L, int idx, const char *tname) {
	void *p = lua_touserdata(L, idx);
	if (p != NULL) {  /* value is a userdata? */
		if (lua_getmetatable(L, idx)) {  /* does it have a metatable? */
			lua_getfield(L, LUA_REGISTRYINDEX, tname);  /* get correct metatable */
			if (lua_rawequal(L, -1, -2)) {  /* does it have the correct mt? */
				lua_pop(L, 2);  /* remove both metatables */
				return p;
			}
		}
	}
	return NULL;
}

void * clua_getGoRef(lua_State *L, int idx) {
	GoRefUd * ud = (GoRefUd *)clua_getudata(L, idx, GO_UDATA_META_NAME);
	return ud->ref;
}

static void detachGoRefUd(GoRefUd * ud) {
	if(ud->ref != NULL) {
		go_unlinkObject(ud->ref);
		ud->ref = NULL;
	}
}

static int CB_cpcall(lua_State * L) {
	GoIntf *ud = (GoIntf*)lua_touserdata(L, 1);
	return go_callbackFromC(*ud);
}

static int CB__call(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = go_callObject(ud->ref);
		if (ret < 0) {
			lua_error(L);
		}
		return ret;
	}
	luaL_error(L, "try to call a detached go object");
	return 0;
}

static int CB__index(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = go_indexObject(ud->ref);
		if (ret < 0) {
			lua_error(L);
		}
	}
	luaL_error(L, "try to index a detached go object");
	return 0;
}

static int CB__newindex(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = go_newindexObject(ud->ref);
		if (ret < 0) {
			lua_error(L);
		}
	}
	luaL_error(L, "try to newindex a detached go object");
	return 0;
}

static int CB__gc(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	detachGoRefUd(ud);
	return 0;
}

int clua_goPcall(lua_State *L, GoIntf cb) {
	return lua_cpcall(L, CB_cpcall, &cb);
}

static void clua_initGoMeta(lua_State *L) {
	luaL_newmetatable(L, GO_UDATA_META_NAME);

	// t[__call]
	lua_pushliteral(L,"__call");
	lua_pushcfunction(L, &CB__call);
	lua_settable(L, -3);

	// t[__index]
	lua_pushliteral(L,"__index");
	lua_pushcfunction(L, &CB__index);
	lua_settable(L, -3);

	// t[__newindex]
	lua_pushliteral(L,"__newindex");
	lua_pushcfunction(L, &CB__newindex);
	lua_settable(L, -3);

	// t[__len]
	lua_pushliteral(L,"__len");
	lua_pushcfunction(L, &CB__len);
	lua_settable(L, -3);

	// t[__gc]
	lua_pushliteral(L,"__gc");
	lua_pushcfunction(L, &CB__gc);
	lua_settable(L, -3);

	lua_pop(L,1);
}

void clua_initState(lua_State *L) {
	clua_initGoMeta(L);
}

void clua_newGoRefUd(lua_State *L, void * ref) {
	GoRefUd * ud = (GoRefUd*)lua_newuserdata(L, sizeof(GoRefUd));
	ud->ref = ref;
	luaL_getmetatable(L, GO_UDATA_META_NAME);
	lua_setmetatable(L, -2);
}

