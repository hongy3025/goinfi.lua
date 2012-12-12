package main

import (
	"fmt"
	//"toc"
	"unsafe"
	"reflect"
	//"strconv"
)

type I interface {
	X() int
	Y() int
}

type A struct {
	x int "x cord of A"
	y int "y cord of A"
}

func (a A) X() int {
	return a.x
}

func (a A) Y() int {
	return a.y
}

func Foo(a, b int, i I, c...int) (x int, y int) {
	return 0, 0
}

var OnlyType = unsafe.Pointer(nil)

func DumpFunc(fn interface{}) {
	t := reflect.TypeOf(fn)
	fmt.Println("t", t.Kind(), t.IsVariadic(), t.Name(), t.NumIn(), t.NumOut())
	for i:=0; i<t.NumIn(); i++ {
		p := t.In(i)
		fmt.Println("in", p)
		if p.Kind() == reflect.Interface {
			fmt.Println("NumMethod", p.Name(), p.NumMethod())
		}
	}

	for i:=0; i<t.NumOut(); i++ {
		fmt.Println("out", t.Out(i))
	}
	v := reflect.ValueOf(Foo)
	fmt.Println("v", v.CanInterface(), v.CanSet())
}

func DumpInterface(i interface{}) {
	t := reflect.TypeOf(i)
	fmt.Println(t.Kind().String())
}

func add(a, b int) (int, int) {
	return a+b, a*b
}

func TestCall() {
	v := reflect.ValueOf(add)
	n := 2
	arg := make([]reflect.Value, n)
	for i:=0; i<n; i++ {
		arg[i] = reflect.ValueOf(i)
	}
	out := v.Call(arg)
	for i:=0; i<len(out); i++ {
		fmt.Println("call result", out[i].Int())
	}
}

type Point struct {
	X, Y int
}

func (point * Point) Add(other Point) *Point {
	point.X += other.X
	point.Y += other.Y
	return point
}

type Rect struct {
	Left, Top, Width, Height int
}

type allMyStruct struct {
	*Point
	*Rect
}

func AddStructs(structs interface {}) {
	contain := reflect.TypeOf(structs)
	for i:=0; i<contain.NumField(); i++ {
		stru := contain.Field(i)
		fmt.Println("Struct", stru.Name)
		stype := stru.Type.Elem()
		for j:=0; j<stype.NumField(); j++ {
			sstru := stype.Field(j)
			fmt.Println("\tfield", sstru.Name, sstru.Type)
		}
		stypePtr := reflect.PtrTo(stype)
		for j:=0; j<stypePtr.NumMethod(); j++ {
			m := stypePtr.Method(j)
			fmt.Println("\tmethod", m.Name, m.Type, m.Type.NumIn(), m.Type.NumOut())
		}
	}
}

func main() {
	AddStructs(allMyStruct{})
	//p1 := Point{1, 1}
	//p2 := Point{2, 2}
	//p1.Add(p2)
	//fmt.Println(p1.X, p1.Y)
	/*
	toc.Print("hello world")
	a := A{}
	fmt.Println("unsafe.Sizeof(a))", unsafe.Sizeof(a))
	fmt.Println("unsafe.Sizeof(a.x))", unsafe.Sizeof(a.x))
	fmt.Println("unsafe.Offsetof(a.y))", unsafe.Offsetof(a.y))
	fmt.Println("1", reflect.ValueOf(a).Kind())
	ta := reflect.TypeOf(a)
	fmt.Println("2", "ta", ta)
	fmt.Println("3", "is struct?", ta.Kind() == reflect.Struct)
	fmt.Println("4", "fields of A", ta.NumField())
	for i:=0; i<ta.NumField(); i++ {
		fmt.Println("5", "field", ta.Field(i).Name)
	}
	for i:=0; i<ta.NumMethod(); i++ {
		fmt.Println("6", "method", ta.Method(i).Name)
	}
	m := make(map[interface{}] int)
	fmt.Println("5", (*A)(OnlyType))
	m[a] = m[a] + 1
	*/
	//DumpFunc(Foo)
	// a := A{}
	//DumpFunc(A.X)
	// a := A{}
	// var i I = a
	// DumpInterface(i)
	//	TestCall()
	//f := 0.1
	//fmt.Println(strconv.FormatFloat(f, 'f', 0, 64))
	//var i interface{}
	//M := make(map[reflect.Type]int)
	//i := MyPoint{}
	//j := MyPoint{}
	//ti := reflect.TypeOf(i)
	//tj := reflect.TypeOf(j)
	//fmt.Println(ti==tj)
	//M[ti] = 1
	//v, ok := M[tj]
	//fmt.Println(v, ok)
}

