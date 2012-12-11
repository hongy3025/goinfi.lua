package main

import "fmt"
//import "os"
import "lua"

/*
func initEnv() {
	LD_LIBRARY_PATH := os.Getenv("LD_LIBRARY_PATH")
	if len(LD_LIBRARY_PATH) > 0 {
		LD_LIBRARY_PATH = "/home/hongy/src/gosandbox/lib" + ":" + LD_LIBRARY_PATH
	} else {
		LD_LIBRARY_PATH = "/home/hongy/src/gosandbox/lib"
	}
	os.Setenv("LD_LIBRARY_PATH", LD_LIBRARY_PATH)
	fmt.Println(os.Getenv("LD_LIBRARY_PATH"))
}
*/

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
