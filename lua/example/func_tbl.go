package main

import (
	"fmt"
	//"os"
	//"io"
	//"strings"
	"goinfi/lua"
)

type A struct {
}

func TestTable() {
	vm := lua.NewVM()
	defer vm.Close()
	vm.Openlibs()

	result := vm.EvalString("T = {1,2,3}; return T")
	fmt.Println("result:", result[0])

	t := result[0].(*lua.LuaTable)
	for i := 1; i <= t.Getn(); i++ {
		fmt.Printf("t[%v]=%v\n", i, t.Get(i))
	}
	ok, err := t.Set("a", "value_a")
	fmt.Println("set a result:", ok, err)
	fmt.Printf("t.a=%v\n", t.Get("a"))

	ok, err = t.Set(nil, "null")
	fmt.Println("set nil result:", ok, err)

	vm.EvalString("print('in lua:', T.a)")

	t.Release()

	ok, err = t.Set("b", "value_b")
	fmt.Println("set result:", ok, err)
}

func TestFunction() {
	vm := lua.NewVM()
	defer vm.Close()
	vm.Openlibs()

	result := vm.EvalString("return function(a, b) return a+b end")

	fn := result[0].(*lua.LuaFunction)
	result, err := fn.Call(1, 2)
	fmt.Println("call result:", result, err)
}

func main() {
	// TestTable()
	TestFunction()
}
