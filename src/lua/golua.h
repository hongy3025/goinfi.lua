#ifndef __GOLUA_CALLBACK__
#define __GOLUA_CALLBACK__

typedef struct { void *t; void *v; } GoIntf;

typedef struct {
	void * ref;
} GoRefUd;

int LuaCpcallWrap(lua_State *L, GoIntf cb);
void lua_initstate(lua_State *L);
void newGoRefUd(lua_State *L, void * ref, char *kind, size_t skind);

#endif


