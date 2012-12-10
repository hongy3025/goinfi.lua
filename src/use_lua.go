package main

import "fmt"
import "lua"

func main() {
	L := lua.LuaL_newstate()

	L.Openlibs()
	L.Cpcall(func(l * lua.State) int {
		fmt.Println("hello world")
		L.Dostring("print(0);error('haha'); print(1)")
		return 0
	})
	L.Close()
}
