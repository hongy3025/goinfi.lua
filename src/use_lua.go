package main

import "fmt"
import "lua"

func main() {
	L := lua.LuaL_newstate()
	L.LuaL_openlibs()
	L.Lua_cpcall(func(l * lua.State) int {
		fmt.Println("hello world")
		return 0
	})
	L.Lua_close()
}
