// Copyright 2013 Jerry Hongy.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lua

/*
#cgo CFLAGS: -Ilua/src
#cgo linux CFLAGS: -DLUA_USE_LINUX
#cgo linux LDFLAGS: -ldl
#cgo LDFLAGS: -lm
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "clua.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"goinfi/base"
	"io"
	"reflect"
	"strings"
	"unsafe"
)

const READ_BUFFER_SIZE = 1024 * 1024

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

type refGo struct {
	prev *refGo
	next *refGo
	vm   *VM
	obj  interface{}
}

func (self *refGo) link(head *refGo) {
	self.next = head.next
	head.next = self
	self.prev = head
	if self.next != nil {
		self.next.prev = self
	}
}

func (self *refGo) unlink() {
	self.prev.next = self.next
	if self.next != nil {
		self.next.prev = self.prev
	}
	self.prev = nil
	self.next = nil
}

type structFieldType uint

const (
	INVALID_FIELD structFieldType = iota
	DATA_FIELD
	METHOD_FIELD
)

type structField struct {
	sinfo       *structInfo
	name        string
	typ         structFieldType
	dataIndex   []int
	methodIndex int
}

type structInfo struct {
	typ    reflect.Type
	fields map[string]*structField

	cache map[*C.char]*structField
	lref  map[string]int
}

func (sinfo *structInfo) makeFieldsIndexCache(vm *VM) {
	L := vm.globalL
	sinfo.cache = make(map[*C.char]*structField, len(sinfo.fields))
	sinfo.lref = make(map[string]int, len(sinfo.fields))
	for key, field := range sinfo.fields {
		pushStringToLua(L, key)
		pstr := C.lua_tolstring(L, -1, nil)
		sinfo.cache[pstr] = field
		sinfo.lref[key] = int(C.luaL_ref(L, C.LUA_REGISTRYINDEX))
	}
}

func newStruct(typ reflect.Type) *structInfo {
	sinfo := &structInfo{typ: typ}
	sinfo.fields = make(map[string]*structField)
	return sinfo
}

type VM struct {
	globalL   *C.lua_State
	refLink   refGo
	structTbl map[reflect.Type]*structInfo
}

type State struct {
	VM *VM
	L  *C.lua_State
}

func NewVM() *VM {
	L := C.luaL_newstate()
	C.clua_initState(L)
	vm := &VM{globalL: L}
	vm.structTbl = make(map[reflect.Type]*structInfo)
	return vm
}

func (vm *VM) initLuaLib() {
	lua_initGolangLib(vm)
	lua_initPackLib(vm)
}

func (vm *VM) findStruct(typ reflect.Type) *structInfo {
	return vm.structTbl[typ]
}

func (vm *VM) addStruct(typ reflect.Type, si *structInfo) *structInfo {
	vm.structTbl[typ] = si
	return si
}

func getStructFieldValue(structValue reflect.Value, fld *structField) (value reflect.Value) {
	switch fld.typ {
	case DATA_FIELD:
		fvalue := structValue.FieldByIndex(fld.dataIndex)
		return fvalue
	case METHOD_FIELD:
		pstruct := reflect.PtrTo(structValue.Type())
		fvalue := pstruct.Method(fld.methodIndex).Func
		return fvalue
	}
	return reflect.ValueOf(nil)
}

func (state State) getStructField(structPtr reflect.Value, lkey C.int) (ret int, err error) {
	L := state.L
	vm := state.VM
	structValue := structPtr.Elem()
	t := structValue.Type()
	info := vm.findStruct(t)
	if info == nil {
		return -1, fmt.Errorf("can not index a solid struct")
	}

	ltype := int(C.lua_type(L, lkey))
	if ltype != C.LUA_TSTRING {
		return -1, fmt.Errorf("field key of struct must be a string")
	}

	//
	//key := stringFromLua(L, lkey)
	//fld, ok := info.fields[key]
	//

	// <hack> using string pstr cache
	pstr := C.lua_tolstring(L, lkey, nil)
	// </hack>
	fld, ok := info.cache[pstr]
	if !ok {
		key := stringFromLua(L, lkey)
		return -1, fmt.Errorf("not such field `%v'", key)
	}

	value := getStructFieldValue(structValue, fld)
	state.goToLuaValue(value)
	return 1, nil
}

func (state State) setStructField(structPtr reflect.Value, lkey C.int, lvalue C.int) (ret int, err error) {
	L := state.L
	vm := state.VM
	structValue := structPtr.Elem()
	t := structValue.Type()
	info := vm.findStruct(t)
	if info == nil {
		return -1, fmt.Errorf("can not assign field of a solid struct")
	}

	ltype := int(C.lua_type(L, lkey))
	if ltype != C.LUA_TSTRING {
		return -1, fmt.Errorf("field key of struct must be a string")
	}

	key := stringFromLua(L, lkey)
	fld, ok := info.fields[key]
	if !ok {
		return -1, fmt.Errorf("not such field `%v'", key)
	}

	if fld.typ != DATA_FIELD {
		return -1, fmt.Errorf("only data field is assignble, but `%v' is not !", key)
	}

	sf := t.FieldByIndex(fld.dataIndex) // StructField 
	value, err := state.luaToGoValue(int(lvalue), &sf.Type)
	if err != nil {
		return -1, err
	}
	structValue.FieldByIndex(fld.dataIndex).Set(value)

	return 0, nil
}

//export GO_unlinkObject
func GO_unlinkObject(ref unsafe.Pointer) {
	node := (*refGo)(ref)
	node.unlink()
}

//export GO_getObjectLength
func GO_getObjectLength(_L unsafe.Pointer, ref unsafe.Pointer) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refGo)(ref)
	// vm := node.vm
	// state := State{vm, L}
	v := reflect.ValueOf(node.obj)

	defer func() {
		if r := recover(); r != nil {
			pushStringToLua(L, fmt.Sprintf("%v", r))
			ret = -1
		}
	}()

	n := v.Len()
	C.lua_pushinteger(L, C.lua_Integer(n))
	return 1
}

//export GO_indexObject
func GO_indexObject(_L unsafe.Pointer, ref unsafe.Pointer, lkey C.int) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refGo)(ref)
	vm := node.vm
	state := State{vm, L}
	v := reflect.ValueOf(node.obj)
	t := v.Type()
	k := v.Kind()

	defer func() {
		if r := recover(); r != nil {
			pushStringToLua(L, fmt.Sprintf("%v", r))
			ret = -1
		}
	}()

	ltype := C.lua_type(L, lkey)
	switch k {
	case reflect.Slice:
		if ltype == C.LUA_TNUMBER {
			idx := int(C.lua_tointeger(L, lkey))
			value := v.Index(idx)
			state.goToLuaValue(value)
			return 1
		}
		panic(fmt.Sprintf("index of slice must be a number type, here got `%v'", luaTypeName(ltype)))
	case reflect.Map:
		keyType := t.Key()
		key, err := state.luaToGoValue(int(lkey), &keyType)
		if err != nil {
			panic(fmt.Sprintf("index type of map must be type `%v', %s", keyType.Kind(), err.Error()))
		}
		value := v.MapIndex(key)
		if !value.IsValid() {
			C.lua_pushnil(L)
			return 1
		}
		state.goToLuaValue(value)
		return 1
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Struct {
			ret, err := state.getStructField(v, lkey)
			if err != nil {
				panic(fmt.Sprintf("error when get field of struct, %s", err.Error()))
			}
			return ret
		}
	}

	panic(fmt.Sprintf("try to index a non-indexable go object, type `%v'", k))
	return -1
}

//export GO_newindexObject
func GO_newindexObject(_L unsafe.Pointer, ref unsafe.Pointer, lkey C.int, lvalue C.int) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refGo)(ref)
	vm := node.vm
	state := State{vm, L}
	v := reflect.ValueOf(node.obj)
	t := v.Type()
	k := v.Kind()

	defer func() {
		if r := recover(); r != nil {
			pushStringToLua(L, fmt.Sprintf("%v", r))
			ret = -1
		}
	}()

	ltype := C.lua_type(L, lkey)
	switch k {
	case reflect.Slice:
		if ltype == C.LUA_TNUMBER {
			tElem := t.Elem()
			value, err := state.luaToGoValue(int(lvalue), &tElem)
			if err != nil {
				panic(fmt.Sprintf("error when assign to slice member, %s", err.Error()))
			}
			idx := int(C.lua_tointeger(L, lkey))
			v.Index(idx).Set(value)
			return 0
		}
		panic(fmt.Sprintf("index of slice must be a number type, got `%v'", luaTypeName(ltype)))
	case reflect.Map:
		keyType := t.Key()
		key, err := state.luaToGoValue(int(lkey), &keyType)
		if err != nil {
			panic(fmt.Sprintf("index type of map must be type `%v', %s", keyType.Kind(), err.Error()))
		}
		if lvtype := C.lua_type(L, lvalue); lvtype == C.LUA_TNIL {
			v.SetMapIndex(key, reflect.Value{})
		} else {
			tElem := t.Elem()
			value, err := state.luaToGoValue(int(lvalue), &tElem)
			if err != nil {
				panic(fmt.Sprintf("error when assign to map member, %s", err.Error()))
			}
			v.SetMapIndex(key, value)
		}
		return 0
	case reflect.Ptr:
		if t.Elem().Kind() == reflect.Struct {
			_, err := state.setStructField(v, lkey, lvalue)
			if err != nil {
				panic(fmt.Sprintf("error when set field of struct, %s", err.Error()))
			}
			return 0
		}
	}

	panic(fmt.Sprintf("try to assign a non-indexable go object, type `%v'", k))
	return -1
}

//export GO_objectToString
func GO_objectToString(_L unsafe.Pointer, ref unsafe.Pointer) int {
	L := (*C.lua_State)(_L)
	node := (*refGo)(ref)
	obj := node.obj
	s := fmt.Sprintf("go object: %v at %p", reflect.TypeOf(obj).Kind(), &obj)
	pushStringToLua(L, s)
	return 1
}

func safeCall(obj reflect.Value, in []reflect.Value) (ok bool, out []reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			ok = false
			err = fmt.Errorf("%v", r)
		}
	}()
	if obj.Type().IsVariadic() {
		out = obj.CallSlice(in)
	} else {
		out = obj.Call(in)
	}
	return true, out, nil
}

func (state State) safeRawCall(objValue reflect.Value) (ret int) {
	defer func() {
		if r := recover(); r != nil {
			pushStringToLua(state.L, fmt.Sprintf("error when call raw function: %v", r))
			ret = -1
		}
	}()
	fn := objValue.Interface().(func(State) int)
	return fn(state)
}

func pushCallArgError(L *C.lua_State, idx int, err error) {
	pushStringToLua(L, fmt.Sprintf("call go func error: arg #%v,", idx)+err.Error())
}

//export GO_callObject
func GO_callObject(_L unsafe.Pointer, ref unsafe.Pointer) int {
	L := (*C.lua_State)(_L)
	node := (*refGo)(ref)
	obj := node.obj
	vm := node.vm
	state := State{vm, L}
	v := reflect.ValueOf(obj)
	k := v.Kind()

	if k != reflect.Func {
		pushStringToLua(L, fmt.Sprintf("try to call a non-function go object, type `%v'", k))
		return -1
	}

	t := v.Type()
	ningo := t.NumIn()
	if ningo == 1 {
		if t.In(0) == reflect.TypeOf(state) {
			return state.safeRawCall(v)
		}
	}

	ltop := int(C.lua_gettop(L))
	in := make([]reflect.Value, ningo)
	ilua := 2
	if t.IsVariadic() {
		for i := 0; i < ningo-1; i++ {
			tin := t.In(i)
			value, err := state.luaToGoValue(ilua, &tin)
			if err != nil {
				pushCallArgError(L, ilua, err)
				return -1
			}
			in[i] = value
			ilua++
		}
		nvarg := ltop - (ilua - 1)
		varg := reflect.MakeSlice(t.In(ningo-1), nvarg, nvarg)
		vargType := t.In(ningo - 1).Elem()
		for i := 0; i < nvarg; i++ {
			value, err := state.luaToGoValue(ilua, &vargType)
			if err != nil {
				pushCallArgError(L, ilua, err)
				return -1
			}
			if value.IsValid() {
				varg.Index(i).Set(value)
			}
			ilua++
		}
		in[ningo-1] = varg
	} else {
		for i := 0; i < ningo; i++ {
			tin := t.In(i)
			value, err := state.luaToGoValue(ilua, &tin)
			if err != nil {
				pushCallArgError(L, ilua, err)
				return -1
			}
			in[i] = value
			ilua++
		}
	}

	ok, out, err := safeCall(v, in)
	if !ok {
		pushStringToLua(L, "call go func error: "+err.Error())
		return -1
	}

	for _, value := range out {
		state.goToLuaValue(value)
	}

	return len(out)
}

func callLuaFuncUtil(state State, inv []reflect.Value, nout int) ([]interface{}, error) {
	L := state.L
	bottom := int(C.lua_gettop(L))

	var result []interface{}
	var nluaout C.int
	var nin C.int
	if nout >= 0 {
		nluaout = C.int(nout)
		result = make([]interface{}, 0, nout)
	} else {
		nluaout = C.LUA_MULTRET
		result = make([]interface{}, 0, 1)
	}
	if inv != nil {
		for _, iarg := range inv {
			state.goToLuaValue(iarg)
		}
		nin = C.int(len(inv))
	} else {
		nin = 0
	}
	ret := int(C.lua_pcall(L, nin, nluaout, 0))
	if ret != 0 {
		err := stringFromLua(L, -1)
		C.lua_settop(L, -2)
		return result, errors.New(err)
	}
	top := int(C.lua_gettop(L))
	for i := bottom; i <= top; i++ {
		value, _ := state.luaToGoValue(i, nil)
		if value.IsValid() {
			result = append(result, value.Interface())
		} else {
			result = append(result, nil)
		}
	}
	rnout := C.int(top + 1 - bottom)
	C.lua_settop(L, -rnout-1)
	return result, nil
}

func (vm *VM) EvalStringWithError(str string, arg ...interface{}) ([]interface{}, error) {
	L := vm.globalL
	state := State{vm, L}
	s, n := stringToC(str)
	bottom := C.lua_gettop(L)
	defer C.lua_settop(L, bottom)

	ret := int(C.luaL_loadbuffer(L, s, n, nil))
	if ret != 0 {
		err := stringFromLua(L, -1)
		return make([]interface{}, 0), errors.New(err)
	}

	nout := -1
	if len(arg) > 0 {
		if x, ok := arg[0].(int); ok {
			nout = x
		}
	}
	return callLuaFuncUtil(state, nil, nout)
}

func (vm *VM) EvalString(str string, arg ...interface{}) []interface{} {
	result, _ := vm.EvalStringWithError(str, arg...)
	return result
}

type loadBufferContext struct {
	reader io.Reader
	buf    []byte
}

//export GO_bufferReaderForLua
func GO_bufferReaderForLua(ud unsafe.Pointer, sz *C.size_t) *C.char {
	context := (*loadBufferContext)(ud)
	n, _ := context.reader.Read(context.buf)
	if n > 0 {
		*sz = C.size_t(n)
		return (*C.char)(unsafe.Pointer(&context.buf[0]))
	}
	return nil
}

func (vm *VM) EvalBufferWithError(reader io.Reader, arg ...interface{}) ([]interface{}, error) {
	L := vm.globalL
	state := State{vm, L}
	context := loadBufferContext{
		reader: reader,
		buf:    make([]byte, READ_BUFFER_SIZE),
	}
	bottom := C.lua_gettop(L)
	defer C.lua_settop(L, bottom)

	ret := int(C.clua_loadProxy(L, unsafe.Pointer(&context)))
	if ret != 0 {
		err := stringFromLua(L, -1)
		return make([]interface{}, 0), errors.New(err)
	}
	nout := -1
	if len(arg) > 0 {
		if x, ok := arg[0].(int); ok {
			nout = x
		}
	}
	return callLuaFuncUtil(state, nil, nout)
}

func (vm *VM) EvalBuffer(reader io.Reader, arg ...interface{}) []interface{} {
	result, _ := vm.EvalBufferWithError(reader, arg...)
	return result
}

func (vm *VM) Close() {
	C.lua_close(vm.globalL)
	vm.globalL = nil
}

func (vm *VM) Gc(what, data int) int {
	return int(C.lua_gc(vm.globalL, C.int(what), C.int(data)))
}

func (vm *VM) Openlibs() {
	C.luaL_openlibs(vm.globalL)
	vm.initLuaLib()
}

func (vm *VM) newRefNode(obj interface{}) *refGo {
	ref := new(refGo)
	ref.vm = vm
	ref.obj = obj
	ref.link(&vm.refLink)
	return ref
}

func checkFunc(fnType reflect.Type) (bool, error) {
	var state State
	foundState := 0
	nin := fnType.NumIn()
	for i := 0; i < nin; i++ {
		if fnType.In(i) == reflect.TypeOf(state) {
			foundState++
		} else if fnType.In(i) == reflect.TypeOf(&state) {
			return false, fmt.Errorf("raw function can not use `*State' as arg, instead using `State'")
		}
	}

	wrongRawFunc := false
	if foundState > 1 {
		wrongRawFunc = true
	} else if foundState == 1 {
		nout := fnType.NumOut()
		if nin != 1 || nout != 1 {
			wrongRawFunc = true
		} else {
			if fnType.Out(0).Kind() != reflect.Int {
				wrongRawFunc = true
			}
		}
	}

	if wrongRawFunc {
		return false, fmt.Errorf("raw function must be type: `func(State) int'")
	}

	return true, nil
}

func luaGetSubTable(L *C.lua_State, table C.int, key string) (bool, error) {
	pushStringToLua(L, key)
	C.lua_gettable(L, table)
	ltype := C.lua_type(L, -1)
	if ltype == C.LUA_TNIL {
		C.lua_createtable(L, 0, 0)
		// table[key] = {}
		pushStringToLua(L, key)
		C.lua_pushvalue(L, -2)
		C.lua_settable(L, table)
	}
	ltype = C.lua_type(L, -1)
	if ltype != C.LUA_TTABLE {
		C.lua_settop(L, -2)
		return false, fmt.Errorf("field `%v` exist, and it is not a table", key)
	}
	return true, nil
}

func luaPushMultiLevelTable(L *C.lua_State, path []string) (bool, error) {
	ok, _ := luaGetSubTable(L, C.LUA_GLOBALSINDEX, path[0])
	if !ok {
		return false, fmt.Errorf("field `%v` exist, and it is not a table", path[0])
	}

	for i := 1; i < len(path); i++ {
		table := C.lua_gettop(L)
		ok, _ := luaGetSubTable(L, table, path[i])
		if !ok {
			return false, fmt.Errorf("field `%v` exist, and it is not a table", strings.Join(path[:i+1], "."))
		}
	}

	return true, nil
}

func (vm *VM) AddFunc(name string, fn interface{}) (bool, error) {
	value := reflect.ValueOf(fn)
	fnType := reflect.TypeOf(fn)
	if value.Kind() != reflect.Func {
		return false, fmt.Errorf("AddFunc only add function type")
	}
	_, err := checkFunc(fnType)
	if err != nil {
		return false, err
	}
	namePath := strings.Split(name, ".")
	baseName := namePath[len(namePath)-1]
	path := namePath[:len(namePath)-1]

	L := vm.globalL
	state := State{vm, L}

	if len(path) <= 0 {
		// _G[a] = fn
		pushStringToLua(L, baseName)
		state.pushObjToLua(fn)
		C.lua_settable(vm.globalL, C.LUA_GLOBALSINDEX)
		return true, nil
	}

	// _G.a.b.c = fn
	ok, err := luaPushMultiLevelTable(L, path)
	if !ok {
		return false, err
	}
	pushStringToLua(L, baseName)
	state.pushObjToLua(fn)
	C.lua_settable(vm.globalL, -3)
	return true, nil
}

func (vm *VM) AddFuncList(prefix string, fnlist []base.KeyValue) (bool, error) {
	for _, kv := range fnlist {
		name := prefix + "." + kv.Key
		fn := kv.Value
		if ok, err := vm.AddFunc(name, fn); !ok {
			return ok, err
		}
	}
	return true, nil
}

func parseStructMembers(sinfo *structInfo, typ reflect.Type, namePath []string, indexPath []int) {
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i) // StructField
		name := sf.Name
		if name[0] >= 'A' && name[0] <= 'Z' {
			myNamePath := append(namePath, name)
			myIndexPath := append(indexPath, i)
			if sf.Type.Kind() == reflect.Struct {
				parseStructMembers(sinfo, sf.Type, myNamePath, myIndexPath)
			} else {
				fname := strings.Join(myNamePath, "_")
				fIndexPath := make([]int, len(myIndexPath))
				copy(fIndexPath, myIndexPath)
				finfo := &structField{
					sinfo:     sinfo,
					name:      fname,
					typ:       DATA_FIELD,
					dataIndex: fIndexPath,
				}
				sinfo.fields[fname] = finfo
			}
		}
	}
}

func parseStructMethods(sinfo *structInfo, typ reflect.Type) {
	stypePtr := reflect.PtrTo(sinfo.typ)
	for i := 0; i < stypePtr.NumMethod(); i++ {
		mfield := stypePtr.Method(i)
		name := mfield.Name
		if name[0] >= 'A' && name[0] <= 'Z' {
			finfo := &structField{
				sinfo:       sinfo,
				name:        name,
				typ:         METHOD_FIELD,
				methodIndex: i,
			}
			sinfo.fields[name] = finfo
		}
	}
}

func (vm *VM) AddStructList(structs interface{}) (bool, error) {
	contain := reflect.TypeOf(structs)
	for i := 0; i < contain.NumField(); i++ {
		sfield := contain.Field(i)
		if sfield.Type.Kind() != reflect.Ptr {
			continue
		}

		stype := sfield.Type.Elem()
		if stype.Kind() != reflect.Struct {
			continue
		}

		if vm.findStruct(stype) != nil {
			continue
		}

		sinfo := vm.addStruct(stype, newStruct(stype))
		namePath := make([]string, 0)
		indexPath := make([]int, 0)
		parseStructMembers(sinfo, stype, namePath, indexPath)

		parseStructMethods(sinfo, stype)

		sinfo.makeFieldsIndexCache(vm)
	}
	return true, nil
}
