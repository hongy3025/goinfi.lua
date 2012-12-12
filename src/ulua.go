package main

import (
	"fmt"
	//"os"
	"lua"
)

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

type Point struct {
	X	int
	Y	int
}

func (p *Point) SumXY() int {
	return p.X + p.Y
}

type Rect struct {
	Left	int
	Top		int
	Width	int
	Height	int
}

type allMyStruct struct {
	*Point
	*Rect
}

func NewPoint(x, y int) *Point {
	return &Point{x, y}
}

func main() {
	fmt.Println("begin")

	L := lua.LuaL_newstate()
	defer L.Close()

	L.Openlibs()

	L.AddStructs(allMyStruct{})

	L.AddFunc("NewPoint", NewPoint)

	L.Dostring("function P(fn) print('call', pcall(fn)) end")
	L.Dostring("P(function() pt = NewPoint(1, 2) end) ")
	L.Dostring("P(function() print(pt.X, pt.Y) end) ")
	L.Dostring("P(function() print(pt.SumXY) end) ")
	L.Dostring("P(function() print(pt:SumXY()) end) ")

	/*
	L.AddFunc("foo", func() {
		fmt.Println("this is function foo")
	})

	L.AddFunc("myadd", func(a, b int) int {
		return a+b
	})

	L.AddFunc("myconcat", func(a, b string) string {
		return a + "," + b
	})

	L.AddFunc("get2d", func() Point {
		return Point {10, 10}
	})

	L.AddFunc("add2d", func(a *Point, b *Point) Point {
		return Point { a.X+b.X, a.Y+b.Y }
	})

	L.Dostring("print('foo', pcall(function() foo() end))")
	L.Dostring("print('myadd', pcall(function() print('result=', myadd(1, 2)) end))")
	L.Dostring("print('myconcat', pcall(function() print(myconcat('1', '2')) end))")

	L.Dostring("print('add2d', pcall(function() print(add2d(get2d(), get2d())) end))")
	*/
}
