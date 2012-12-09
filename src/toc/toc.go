package toc

/*
#include <stdio.h>
#include <string.h>
int add(int a, int b) {
	return a+b;
}
void print(char * s, int n) {
	char ss[256];
	strncpy(ss, s, 256);
	printf("%s,%d\n", ss, n);
}
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
}

