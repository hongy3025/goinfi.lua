package main

import (
	"fmt"
	"time"
	"goinfi/lua"
)

type Point struct {
	X int
	Y int
}

type DoublePoint struct {
	P1 Point
	P2 Point
}

func NewDoublePoint() *DoublePoint {
	return new(DoublePoint)
}

type allMyStruct struct {
	*Point
	*DoublePoint
}

func Test() {
	vm := lua.NewVM()
	vm.Openlibs()

	vm.AddStructs(allMyStruct{})
	vm.AddFunc("test.NewDoublePoint", NewDoublePoint)

	begin := time.Now()
	_, err := vm.EvalStringWithError(`
		doublePoint = test.NewDoublePoint()
		for i=0, 1000000 do
			doublePoint.P1_X = doublePoint.P2_X
		end
	`)
	end := time.Now()
	fmt.Println(err, end.Sub(begin))
}

func main() {
	Test()
}
