package main

import (
	"fmt"
	"os"
	"lua"
	"io"
	"strings"
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

type DoublePoint struct {
	P1	Point
	P2	Point
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
	*DoublePoint
}

func NewPoint(x, y int) *Point {
	return &Point{x, y}
}

func NewDoublePoint() *DoublePoint {
	return new(DoublePoint)
}

func NewIntSlice() []int {
	return []int{1,2,3,4}
}

func NewStrIntMap() map[string]int {
	return map[string]int{
		"a" : 1, "b" : 2, "c" : 3,
	}
}

func Test1() {
	L := lua.LuaL_newstate()
	defer L.Close()

	L.Openlibs()

	L.AddStructs(allMyStruct{})

	L.AddFunc("NewPoint", NewPoint)
	L.AddFunc("NewDoublePoint", NewDoublePoint)
	L.AddFunc("NewIntSlice", NewIntSlice)
	L.AddFunc("NewStrIntMap", NewStrIntMap)

	var ES = func(s string) {
		ok, err := L.ExecString(s)
		fmt.Println("> ES", ok, err)
	}
	ES("pt = NewPoint(1, 2)")
	ES("print(pt.X, pt.Y)")
	ES("print(pt.SumXY)")

	var EB = func(buf io.Reader) {
		ok, err := L.ExecBuffer(buf)
		fmt.Println("> EB", ok, err)
	}
	b := strings.NewReader("print(pt:SumXY())")
	EB(b)

	f, err := os.Open("test1.lua")
	if err != nil {
		fmt.Println(err)
	} else {
		defer f.Close()
		EB(f)
	}

	ES("dp = NewDoublePoint(); print('dp.P1.X', dp.P1_X)")

	ES("slice = NewIntSlice(); print('slice[0]', #slice, slice[0])")

	ES("map = NewStrIntMap(); print('map[a]', #map, map['c'], map['x'])")
}

func main() {
	Test1()
}

func oldTest() {
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
