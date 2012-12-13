package lua

/*
#cgo CFLAGS: -I../../external/lua-5.1.5/src
#cgo LDFLAGS: -L../../lib -llua -lm -ldl
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include "golua.h"
*/
import "C"

import (
	"io"
	"fmt"
	"unsafe"
	"reflect"
	"strings"
	"errors"
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

type refNode struct {
	prev *refNode
	next *refNode
	vm *VM
	obj interface{}
}

func (self * refNode) link(head * refNode) {
	self.next = head.next
	head.next = self
	self.prev = head
	if self.next != nil {
		self.next.prev = self
	}
}

func (self * refNode) unlink() {
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
	sinfo *structInfo
	name string
	typ structFieldType
	dataIndex []int
	methodIndex int
}

type structInfo struct {
	typ reflect.Type
	fields map[string]*structField
}

func newStruct(typ reflect.Type) * structInfo {
	sinfo := & structInfo { typ : typ }
	sinfo.fields = make(map[string]*structField)
	return sinfo
}

type VM struct {
	globalL *C.lua_State
	refLink refNode
	structTbl map[reflect.Type]*structInfo
}

type State struct {
	vm *VM
	L *C.lua_State
}

func NewVM() *VM {
	L := C.luaL_newstate()
	C.clua_initState(L)
	vm := &VM{ globalL : L }
	vm.structTbl = make(map[reflect.Type]*structInfo)
	vm.initGoLib()
	return vm
}

func (vm * VM) initGoLib() {
	//vm.AddFunc("golang.keys", getMapKeys)
}

func (vm *VM) findStruct(typ reflect.Type) * structInfo {
	return vm.structTbl[typ]
}

func (vm *VM) addStruct(typ reflect.Type, si *structInfo) (*structInfo ) {
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

func (state *State) getStructField(structPtr reflect.Value, lkey C.int) (ret int, err error) {
	L := state.L
	vm := state.vm
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

	key := stringFromLua(L, lkey)
	fld, ok := info.fields[key]
	if !ok {
		return -1, fmt.Errorf("not such field `%v'", key)
	}

	value := getStructFieldValue(structValue, fld)
	state.goToLuaValue(value)
	return 1, nil
}

func (state *State) setStructField(structPtr reflect.Value, lkey C.int, lvalue C.int) (ret int, err error) {
	L := state.L
	vm := state.vm
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

//export go_unlinkObject
func go_unlinkObject(ref unsafe.Pointer) {
	node := (*refNode)(ref)
	node.unlink()
}

//export go_getObjectLength
func go_getObjectLength(_L unsafe.Pointer, ref unsafe.Pointer) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refNode)(ref)
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

//export go_indexObject
func go_indexObject(_L unsafe.Pointer, ref unsafe.Pointer, lkey C.int) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refNode)(ref)
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

//export go_newindexObject
func go_newindexObject(_L unsafe.Pointer, ref unsafe.Pointer, lkey C.int, lvalue C.int) (ret int) {
	L := (*C.lua_State)(_L)
	node := (*refNode)(ref)
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

//export go_objectToString
func go_objectToString(_L unsafe.Pointer, ref unsafe.Pointer) int {
	L := (*C.lua_State)(_L)
	node := (*refNode)(ref)
	obj := node.obj
	// vm := node.vm
	// state := State{vm, L}
	// v := reflect.ValueOf(node.obj)
	// t := v.Type()
	// k := v.Kind()
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
	out = obj.Call(in)
	return true, out, nil
}

func (state *State) safeRawCall(objValue reflect.Value) (ret int) {
	defer func() {
		if r := recover(); r != nil {
			pushStringToLua(state.L, fmt.Sprintf("error when call raw function: %v", r))
			ret = -1
		}
	}()
	fn := objValue.Interface().(func(*State) int)
	return fn(state)
}

//export go_callObject
func go_callObject(_L unsafe.Pointer, ref unsafe.Pointer) int {
	L := (*C.lua_State)(_L)
	node := (*refNode)(ref)
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
	ingo := t.NumIn()
	if ingo == 1 {
		if t.In(0) == reflect.TypeOf(&state) {
			return state.safeRawCall(v)
		}
	}

	inlua := int(C.lua_gettop(L)) - 1
	n := min(ingo, inlua)

	in := make([]reflect.Value, n)
	for i:=0; i<n; i++ {
		tin := t.In(i)
		value, err := state.luaToGoValue(i+2, &tin)
		if err != nil {
			pushStringToLua(L, fmt.Sprintf("call go func error: arg %v,", i) + err.Error())
			return -1
		}
		in[i] = value
	}

	ok, out, err := safeCall(v, in)
	if !ok {
		pushStringToLua(L, "call go func error: " + err.Error())
		return -1
	}

	for _, value := range out {
		state.goToLuaValue(value)
	}

	 return len(out)
}

func (vm *VM) ExecString(str string) (bool, error) {
	cstr := stringToC(str)
	L := vm.globalL
	ret := int(C.luaL_loadbuffer(L, cstr.s, cstr.n, nil))
	if ret != 0 {
		err := stringFromLua(L, 1)
		C.lua_settop(L, -2)
		return false, errors.New(err)
	}
	ret = int(C.lua_pcall(L, 0, 0, 0))
	if ret != 0 {
		err := stringFromLua(L, 1)
		C.lua_settop(L, -2)
		return false, errors.New(err)
	}
	return true, nil
}

type loadBufferContext struct {
	reader io.Reader
	buf []byte
}

//export go_bufferReaderForLua
func go_bufferReaderForLua(ud unsafe.Pointer, sz *C.size_t) *C.char {
	context := (*loadBufferContext)(ud)
	n, _ := context.reader.Read(context.buf)
	if n > 0 {
		*sz = C.size_t(n)
		return (*C.char)(unsafe.Pointer(&context.buf[0]))
	}
	return nil
}

const BUFFER_SIZE = 1024*1024

func (vm *VM) ExecBuffer(reader io.Reader) (bool, error) {
	L := vm.globalL
	context := loadBufferContext {
		reader : reader,
		buf : make([]byte, BUFFER_SIZE),
	 }
	ret := int(C.clua_loadProxy(L, unsafe.Pointer(&context)))
	if ret != 0 {
		err := stringFromLua(L, 1)
		C.lua_settop(L, -2)
		return false, errors.New(err)
	}
	ret = int(C.lua_pcall(L, 0, 0, 0))
	if ret != 0 {
		err := stringFromLua(L, 1)
		C.lua_settop(L, -2)
		return false, errors.New(err)
	}
	return true, nil
}

func (vm *VM) Close() {
	C.lua_close(vm.globalL)
	vm.globalL = nil
}

func (vm *VM) Gc(what, data int) int {
	return int(C.lua_gc(vm.globalL, C.int(what), C.int(data)))
}

func (vm *VM) Openlibs() {
	C.luaL_openlibs(vm.globalL);
}

func (vm *VM) newRefNode(obj interface{}) *refNode {
	ref := new(refNode)
	ref.vm = vm
	ref.obj = obj
	ref.link(&vm.refLink)
	return ref
}

func (vm *VM) AddFunc(name string, fn interface{}) (bool, error) {
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func {
		return false, fmt.Errorf("fn must be a function")
	}
	pushStringToLua(vm.globalL, name)
	state := State{ vm, vm.globalL }
	state.pushObjToLua(fn)
	C.lua_settable(vm.globalL, C.LUA_GLOBALSINDEX)

	return true, nil
}

func parseStructMembers(sinfo *structInfo, typ reflect.Type, namePath []string, indexPath []int) {
	for i:=0; i<typ.NumField(); i++ {
		sf := typ.Field(i) // StructField
		name := sf.Name
		myNamePath := append(namePath, name)
		myIndexPath := append(indexPath, i)
		if sf.Type.Kind() == reflect.Struct {
			parseStructMembers(sinfo, sf.Type, myNamePath, myIndexPath)
		} else {
			fname := strings.Join(myNamePath, "_")
			fIndexPath := make([]int, len(myIndexPath))
			copy(fIndexPath, myIndexPath)
			finfo := & structField {
				sinfo : sinfo,
				name : fname,
				typ : DATA_FIELD,
				dataIndex : fIndexPath,
			}
			sinfo.fields[fname] = finfo
		}
	}
}

func parseStructMethods(sinfo *structInfo, typ reflect.Type) {
	stypePtr := reflect.PtrTo(sinfo.typ)
	for i:=0; i<stypePtr.NumMethod(); i++ {
		mfield := stypePtr.Method(i)
		name := mfield.Name
		if name[0] >= 'A' && name[0] <= 'Z' {
			finfo := & structField {
				sinfo : sinfo,
				name : name,
				typ : METHOD_FIELD,
				methodIndex : i,
			}
			sinfo.fields[name] = finfo
		}
	}
}

func (vm *VM) AddStructs(structs interface{}) (bool, error) {
	contain := reflect.TypeOf(structs)
	for i:=0; i<contain.NumField(); i++ {
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
	}
	return true, nil
}

func (state *State) Pushstring(str string) {
	pushStringToLua(state.L, str)
}

//--------------------------------------------------------------------------------------------

/*
func getMapKeys(vm * VM, obj interface{}) LuaTable {
	// scope := vm.NewScope()
	// defer scope.Leave()

	T := vm.NewLuaTable()
	keys := reflect.ValueOf(obj).MapKeys()
	for _, keyValue := range keys {
		T.Append(keyValue.Interface())
	}
	return T
}
*/

/*
//export go_callbackFromC
func go_callbackFromC(ud interface {}) int {
	cb := ud.(callbackData)
	return cb.fn(cb.state)
}
*/

/*
type GoFunc func(*VM) int

type callbackData struct {
	vm * VM
	fn GoFunc
}
*/

/*
func (vm *VM) Cpcall(fn GoFunc) int {
	var cb interface{} = callbackData{ vm : vm, fn : fn }
	p := (*C.GoIntf)(unsafe.Pointer(&cb))
	return int(C.clua_goPcall(vm.l, *p))
}
*/

/*
func (vm *VM) Pushstring(str string) {
	pushStringToLua(vm.l, str)
}
*/

/*
never call this
func (vm *VM) Error() {
	C.lua_error(vm.l)
}
*/

