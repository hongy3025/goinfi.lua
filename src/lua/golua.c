#include <string.h>
#include <lua.h>
#include <lauxlib.h>
#include "golua.h"

static int CB_cpcall(lua_State * L) {
	GoIntf *ud = (GoIntf*)lua_touserdata(L, 1);
	return go_callbackFromC(*ud);
}

static void * lua_getudata(lua_State *L, int ud, const char *tname) {
	void *p = lua_touserdata(L, ud);
	if (p != NULL) {  /* value is a userdata? */
		if (lua_getmetatable(L, ud)) {  /* does it have a metatable? */
			lua_getfield(L, LUA_REGISTRYINDEX, tname);  /* get correct metatable */
			if (lua_rawequal(L, -1, -2)) {  /* does it have the correct mt? */
				lua_pop(L, 2);  /* remove both metatables */
				return p;
			}
		}
	}
	return NULL;
}

static int CB_funcCall(lua_State * L) {
	GoRefUd * ud = (GoRefUd*)lua_touserdata(L, 1);
	if (ud == NULL) {
		return 0;
	}
	return go_callObject(ud->ref);
}

static int CB_gc(lua_State * L) {
	return 0;
}

/*
static GoRefUd * checkGoRefUd(lua_State* L, int index)
{
	unsigned int* fid = (unsigned int*)luaL_checkudata(L,index,"GoLua.GoFunction");
	luaL_argcheck(L, fid != NULL, index, "'GoFunction' expected");
	return fid;
}
*/

int LuaCpcallWrap(lua_State *L, GoIntf cb) {
	return lua_cpcall(L, CB_cpcall, &cb);
}

void lua_initstate(lua_State *L) {
	luaL_newmetatable(L,"lua.GoFunc");

	// t[__call]
	lua_pushliteral(L,"__call");
	lua_pushcfunction(L, &CB_funcCall);
	lua_settable(L,-3);

	// t[__gc]
	lua_pushliteral(L,"__gc");
	lua_pushcfunction(L, &CB_gc);
	lua_settable(L,-3);

	lua_pop(L,1);
}

static void makeMetaName(char * metaName, char * kind, size_t skind) {
	char * p = metaName;
	strcpy(p, "lua.");
	p += strlen(p);
	memcpy(p, kind, skind);
	p[skind] = '\0';
}

void newGoRefUd(lua_State *L, void * ref, char *kind, size_t skind) {
	char metaName[256];
	makeMetaName(metaName, kind, skind);
	GoRefUd * ud = (GoRefUd*)lua_newuserdata(L, sizeof(GoRefUd));
	ud->ref = ref;
	luaL_getmetatable(L, metaName);
	lua_setmetatable(L, -2);
}

