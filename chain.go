package echoswg

import (
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
)

func validateChain(handlers []interface{}) ([]reflect.Type, reflect.Type, error) {
	var totalIns []reflect.Type
	var totalOuts []reflect.Type

	for _, h := range handlers {
		addTypes(h, &totalIns, &totalOuts)
	}

	uniqueIns := findUniqueTypes(totalIns, totalOuts)
	uniqueOuts := findUniqueTypes(totalOuts, totalIns)
	// if len(uniqueOuts) > 1 {
	// 	return nil, nil, fmt.Errorf("more then one unique output type: %s", uniqueOuts)
	// }
	var uniqueOut reflect.Type

	if len(uniqueOuts) > 0 {
		uniqueOut = uniqueOuts[len(uniqueOuts)-1]
	}
	return uniqueIns, uniqueOut, nil
}

func addTypes(handler interface{}, totalIns *[]reflect.Type, totalOuts *[]reflect.Type) {
	if handler == nil {
		return
	}
	handlerType := reflect.TypeOf(handler)
	// fmt.Printf("handler type >> %s\n", handlerType)
	for i := 0; i < handlerType.NumIn(); i++ {
		if handlerType.In(i).String() != "echo.Context" {
			//fmt.Printf("add to in handlerType.In(i).String()=%s\n", handlerType.In(i).String())
			*totalIns = append(*totalIns, handlerType.In(i))
		}
	}
	for i := 0; i < handlerType.NumOut(); i++ {
		if handlerType.Out(i).String() != "error" {
			*totalOuts = append(*totalOuts, handlerType.Out(i))
		}
	}
}
func getFuncName(handler interface{}) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	arr := strings.Split(fullName, ".")
	return arr[len(arr)-1]
}
func getOperationID(tag string, handlers []interface{}) string {
	var fullName = tag
	if len(handlers) == 0 {
		return fullName + fmt.Sprintf(".%d", rand.Int())
	}

	fullName += "." + strings.TrimSuffix(getFuncName(handlers[len(handlers)-1]), "-fm")
	//for _, handler := range handlers {
	//	fullName += "." + strings.TrimSuffix(getFuncName(handler), "-fm")
	//}
	return fullName
}

//func isInTypeDefined(inType reflect.Type, handler interface{}) bool {
//	if handler == nil {
//		return false
//	}
//	if inType == nil {
//		return true
//	}
//	handlerType := reflect.TypeOf(handler)
//	for i := 0; i < handlerType.NumIn(); i++ {
//		if handlerType.In(i).String() == inType.String() {
//			return true
//		}
//	}
//	return false
//}

func findUniqueTypes(source []reflect.Type, compareWith []reflect.Type) []reflect.Type {
	var uniques []reflect.Type
	for _, in := range source {
		isUnique := true
		for _, out := range compareWith {
			if in.String() == out.String() {
				isUnique = false
			}
		}
		if isUnique {
			uniques = append(uniques, in)
		}
	}
	return uniques
}
