package main

import "fmt"
import "toc"
import "unsafe"
import "reflect"

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

func main() {
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
	for i := 0; i<ta.NumField(); i++ {
		fmt.Println("5", "field", ta.Field(i).Name)
	}
	for i := 0; i<ta.NumMethod(); i++ {
		fmt.Println("6", "method", ta.Method(i).Name)
	}
}
