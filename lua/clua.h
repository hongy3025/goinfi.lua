// Copyright 2013 Jerry Hongy.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#ifndef __GOLUA_CALLBACK__
#define __GOLUA_CALLBACK__

typedef struct { void *t; void *v; } GoIntf;

typedef struct {
	void * ref;
} GoRefUd;

int clua_goPcall(lua_State *L, GoIntf cb);
void clua_initState(lua_State *L);
void clua_newGoRefUd(lua_State *L, void * ref);
void * clua_getGoRef(lua_State *L, int lv);
int clua_loadProxy(lua_State *L, void *context);

#endif


