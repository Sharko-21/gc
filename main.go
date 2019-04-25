package main

import (
	"fmt"
	"runtime/debug"
)

const (
	STACK_MAX_SIZE int = 256
	ICGT           int = 8
)

type VMObjInterface interface{}

type CompositeObj struct {
	next VMObjInterface

	isMarked bool

	headValue VMObjInterface
	tailValue VMObjInterface
}

type IntObj struct {
	next VMObjInterface

	isMarked bool
	value    int
}

type VM struct {
	stack     [STACK_MAX_SIZE]VMObjInterface
	stackSize int

	numOfObjects    int
	maxNumofObjects int

	beginOfList VMObjInterface
}

func createVM() VM {
	vm := new(VM)
	vm.stackSize = 0
	vm.numOfObjects = 0
	vm.maxNumofObjects = ICGT
	vm.beginOfList = nil

	return *vm
}

func markAll(vm *VM) {
	for i := 0; i < vm.stackSize; i++ {
		mark(vm.stack[i])
	}
}

func mark(objInterface VMObjInterface) {
	if _, ok := objInterface.(*IntObj); ok {
		objInterface.(*IntObj).isMarked = true
	} else if _, ok := objInterface.(*CompositeObj); ok {
		if objInterface.(*CompositeObj).isMarked == true {
			return
		}
		objInterface.(*CompositeObj).isMarked = true

		mark(objInterface.(*CompositeObj).headValue)
		mark(objInterface.(*CompositeObj).tailValue)
	}
}

func markSweep(vm *VM) {
	obj := &vm.beginOfList

	for *obj != nil {
		if _, ok := (*obj).(*IntObj); ok {
			if (*obj).(*IntObj).isMarked == false {
				unreached := obj
				obj = &(*unreached).(*IntObj).next
				*unreached = nil
				//here must be free() but we don't have it in go :D
				vm.numOfObjects--
			} else {
				(*obj).(*IntObj).isMarked = false
				obj = &(*obj).(*IntObj).next
			}
		} else if _, ok := (*obj).(*CompositeObj); ok {
			if (*obj).(*CompositeObj).isMarked == false {
				unreached := obj
				obj = &(*unreached).(*CompositeObj).next
				*unreached = nil
				vm.numOfObjects--
			} else {
				(*obj).(*CompositeObj).isMarked = false
				obj = &(*obj).(*CompositeObj).next
			}
		}
	}
}

func push(vm *VM, obj VMObjInterface) {
	if _, ok := obj.(*IntObj); ok {
		vm.stack[vm.stackSize] = obj.(*IntObj)
	} else if _, ok := obj.(*CompositeObj); ok {
		vm.stack[vm.stackSize] = obj.(*CompositeObj)
	}
	vm.stackSize++
}

func pop(vm *VM) VMObjInterface {
	vm.stackSize--
	stackElem := vm.stack[vm.stackSize]
	vm.stack[vm.stackSize] = nil
	return stackElem
}

func newObject(vm *VM, obj VMObjInterface) VMObjInterface {
	if vm.numOfObjects == vm.maxNumofObjects {
		gc(vm)
	}

	if _, ok := obj.(*IntObj); ok {
		obj.(*IntObj).next = vm.beginOfList
		obj.(*IntObj).isMarked = false
	} else if _, ok := obj.(*CompositeObj); ok {
		obj.(*CompositeObj).next = vm.beginOfList
		obj.(*CompositeObj).isMarked = false
	}

	vm.beginOfList = obj
	vm.numOfObjects++
	return obj
}

func pushInt(vm *VM, intValue int) {
	obj := newObject(vm, &IntObj{})
	obj.(*IntObj).value = intValue

	push(vm, obj)
}

func pushPair(vm *VM) *CompositeObj {
	obj := newObject(vm, &CompositeObj{})
	obj.(*CompositeObj).headValue = pop(vm)
	obj.(*CompositeObj).tailValue = pop(vm)

	push(vm, obj)
	return obj.(*CompositeObj)
}

func gc(vm *VM) {
	numOfObjects := vm.numOfObjects

	markAll(vm)
	markSweep(vm)

	vm.maxNumofObjects = vm.numOfObjects * 2
	fmt.Printf("Collected %d objects, left %d\n", numOfObjects-vm.numOfObjects, vm.numOfObjects)
}

func freeVM(vm *VM) {
	vm.stackSize = 0
	gc(vm)
}

func printList(vm *VM) {
	beginOfList := &vm.beginOfList

	for *beginOfList != nil {
		if _, ok := (*beginOfList).(*IntObj); ok {
			fmt.Println("heh", *beginOfList)
			beginOfList = &(*beginOfList).(*IntObj).next
		} else if _, ok := (*beginOfList).(*CompositeObj); ok {
			fmt.Println("heh2", *beginOfList)
			beginOfList = &(*beginOfList).(*CompositeObj).next
		}
	}
}

func firstTest() {
	fmt.Println("1: Objects on the stack are preserved.")
	vm := createVM()
	pushInt(&vm, 1)
	pushInt(&vm, 2)

	gc(&vm)
	freeVM(&vm)
}

func secondTest() {
	fmt.Println("2: Unreached objects are collected.")
	vm := createVM()
	pushInt(&vm, 1)
	pushInt(&vm, 2)
	pop(&vm)
	pop(&vm)
	pushInt(&vm, 3)
	pushInt(&vm, 4)

	gc(&vm)
	freeVM(&vm)
}

func thirdTest() {
	fmt.Println("3: Reach the nested objects.")
	vm := createVM()
	pushInt(&vm, 1)
	pushInt(&vm, 2)
	pushPair(&vm)
	pushInt(&vm, 3)
	pushInt(&vm, 4)

	pushPair(&vm)
	pushPair(&vm)

	gc(&vm)
	freeVM(&vm)
}

func main() {
	debug.SetGCPercent(-1)
	firstTest()
	secondTest()
	thirdTest()
}
