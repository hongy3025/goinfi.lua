package toc

/*
#include "toc.h"
 */
import "C"
import "unsafe"
import "reflect"

func Add(a int, b int) int {
	return int(C.add(C.int(a), C.int(b)))
}

func Print(s string) {
	b := []byte(s)
	h := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := unsafe.Pointer(h.Data)
	size := h.Len
	C.print((*C.char)(data), C.int(size))
	if true {
		import "os"
	}
}

