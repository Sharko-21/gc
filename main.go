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
	Next VMObjInterface

	IsMarked bool

	HeadValue VMObjInterface
	TailValue VMObjInterface
}

type IntObj struct {
	Next VMObjInterface

	IsMarked bool
	Value    int
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
		objInterface.(*IntObj).IsMarked = true
	} else if _, ok := objInterface.(*CompositeObj); ok {
		if objInterface.(*CompositeObj).IsMarked == true {
			return
		}
		objInterface.(*CompositeObj).IsMarked = true

		mark(objInterface.(*CompositeObj).HeadValue)
		mark(objInterface.(*CompositeObj).TailValue)
	}
}

func markSweep(vm *VM) {
	obj := &vm.beginOfList

	for *obj != nil {
		if _, ok := (*obj).(*IntObj); ok {
			if (*obj).(*IntObj).IsMarked == false {
				unreached := obj
				obj = &(*unreached).(*IntObj).Next
				*unreached = nil
				//here must be free() but we don't have it in go :D
				vm.numOfObjects--
			} else {
				(*obj).(*IntObj).IsMarked = false
				obj = &(*obj).(*IntObj).Next
			}
		} else if _, ok := (*obj).(*CompositeObj); ok {
			if (*obj).(*CompositeObj).IsMarked == false {
				unreached := obj
				obj = &(*unreached).(*CompositeObj).Next
				*unreached = nil
				vm.numOfObjects--
			} else {
				(*obj).(*CompositeObj).IsMarked = false
				obj = &(*obj).(*CompositeObj).Next
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
		obj.(*IntObj).Next = vm.beginOfList
		obj.(*IntObj).IsMarked = false
	} else if _, ok := obj.(*CompositeObj); ok {
		obj.(*CompositeObj).Next = vm.beginOfList
		obj.(*CompositeObj).IsMarked = false
	}

	vm.beginOfList = obj
	vm.numOfObjects++
	return obj
}

func pushInt(vm *VM, intValue int) {
	obj := newObject(vm, &IntObj{})
	obj.(*IntObj).Value = intValue

	push(vm, obj)
}

func pushPair(vm *VM) *CompositeObj {
	obj := newObject(vm, &CompositeObj{})
	obj.(*CompositeObj).HeadValue = pop(vm)
	obj.(*CompositeObj).TailValue = pop(vm)

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
			beginOfList = &(*beginOfList).(*IntObj).Next
		} else if _, ok := (*beginOfList).(*CompositeObj); ok {
			fmt.Println("heh2", *beginOfList)
			beginOfList = &(*beginOfList).(*CompositeObj).Next
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
