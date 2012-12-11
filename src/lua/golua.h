#ifndef __GOLUA_CALLBACK__
#define __GOLUA_CALLBACK__

typedef struct { void *t; void *v; } GoIntf;

typedef struct {
	void * ref;
} GoRefUd;

int clua_goPcall(lua_State *L, GoIntf cb);
void clua_initState(lua_State *L);
void clua_newGoRefUd(lua_State *L, void * ref, char *kind, size_t skind);

#endif


