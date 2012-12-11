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

	L.Cpcall(func(l * lua.State) int {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("there is an error")
			}
		}()
		fmt.Println("hello world")
		//panic("aaa")
		// L.Dostring("print(0);error('haha'); print(1)")
		return 0
	})
}
*/

func main() {
	fmt.Println("begin")

	L := lua.LuaL_newstate()
	defer L.Close()

	L.Openlibs()

	L.NewLuaFunc("foo", func() {
		fmt.Println("this is function foo")
	})

	L.NewLuaFunc("myadd", func(a, b int) int {
		fmt.Println("ab", a, b)
		return a+b
	})

	L.Dostring("print('foo', pcall(function() foo() end))")
	L.Dostring("print('myadd', pcall(function() print(myadd(1, 2)) end))")
}
