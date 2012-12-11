package main

import "fmt"
//import "toc"
import "unsafe"
import "reflect"

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


func main() {
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
	TestCall()
}
