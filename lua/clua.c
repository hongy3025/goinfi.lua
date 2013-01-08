#include <string.h>
#include <lua.h>
#include <lauxlib.h>
#include "clua.h"
#include "_cgo_export.h"

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
		GO_unlinkObject(ud->ref);
		ud->ref = NULL;
	}
}

static int CB__call(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = GO_callObject(L, ud->ref);
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
		int ret = GO_indexObject(L, ud->ref, 2);
		if (ret < 0) {
			lua_error(L);
		}
		return ret;
	}
	luaL_error(L, "try to index a detached go object");
	return 0;
}

static int CB__newindex(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = GO_newindexObject(L, ud->ref, 2, 3);
		if (ret < 0) {
			lua_error(L);
		}
		return ret;
	}
	luaL_error(L, "try to newindex a detached go object");
	return 0;
}

static int CB__len(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = GO_getObjectLength(L, ud->ref);
		if (ret < 0) {
			lua_error(L);
		}
		return ret;
	}
	luaL_error(L, "try to get length of a detached go object");
	return 0;
}

static int CB__tostring(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud->ref != NULL) {
		int ret = GO_objectToString(L, ud->ref);
		if (ret < 0) {
			lua_error(L);
		}
		return ret;
	}
	lua_pushfstring(L, "detached go object at %p", ud);
	return 1;
}

static int CB__gc(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	detachGoRefUd(ud);
	return 0;
}

static const char * clua_goBufferReader(lua_State *L, void *ud, size_t *sz) {
	return GO_bufferReaderForLua(ud, sz);
}

int clua_loadProxy(lua_State *L, void *context) {
	return lua_load(L, clua_goBufferReader, context, NULL);
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

	// t[__tostring]
	lua_pushliteral(L,"__tostring");
	lua_pushcfunction(L, &CB__tostring);
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

