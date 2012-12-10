package main

import "fmt"
import "golua"

func cpmain(l * golua.State) int {
	fmt.Printf("hello")
	return 0
}

func main() {
	L := golua.LuaL_newstate()
	L.Lua_cpcall(cpmain)
	L.Lua_close()
}
