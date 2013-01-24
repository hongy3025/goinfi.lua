// Copyright 2013 Jerry Hongy.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
