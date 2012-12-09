// Created by cgo - DO NOT EDIT

package toc

import "unsafe"

import "syscall"

import _ "runtime/cgo"

type _ unsafe.Pointer

func _Cerrno(dst *error, x int) { *dst = syscall.Errno(x) }
type _Ctype_void [0]byte

func _Cfunc_print_me()
