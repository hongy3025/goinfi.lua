package main

import (
	"fmt"
	"os"
	"io"
	//"strings"
	// "bytes"
	"goinfi/lua"
)

func Test1() {
	vm := lua.NewVM()
	defer vm.Close()
	vm.Openlibs()
	var EB = func(buf io.Reader) {
		_, err := vm.EvalBufferWithError(buf, 0)
		if err != nil {
			fmt.Println("> EB", err)
		}
	}
	var ES = func(s string) {
		_, err := vm.EvalStringWithError(s, 0)
		if err != nil {
			fmt.Println("> ES", err)
		}
	}

	f, err := os.Open("baselib.lua")
	if err != nil {
		f, err = os.Open("example/baselib.lua")
		if err != nil {
			fmt.Println(err)
		}
	}
	defer f.Close()
	EB(f)

	ES(`
		bin = pack.Pack(1, 2, {a=1,b=2,c=3, 1, 2})
		print('#bin', #bin)
		print(ToString({pack.Unpack(bin)}))
	`)

	ES(`
		bin = pack.Pack('key', {a=1})
		print('#bin', #bin)
		print(ToString({pack.Unpack(bin)}))
	`)

	ES(`
		bin = pack.Pack({a={b={c={d={e='depth'}}}}})
		print('#bin', #bin)
		print(ToString({pack.Unpack(bin)}))
	`)

	result, err := vm.EvalStringWithError("return 1,2,3,'aaa'", 0)
	fmt.Println(result, err)
}

func main() {
	Test1()
}

