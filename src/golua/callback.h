#ifndef __GOLUA_CALLBACK__
#define __GOLUA_CALLBACK__

typedef struct { void *t; void *v; } GoIntf;
int lua_cpcall_wrap(lua_State *L, GoIntf cb);

#endif


