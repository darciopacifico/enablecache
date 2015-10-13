package aop

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/darciopacifico/enablecache/cache"
	"regexp"
	"runtime"
	"strconv"
)

//regex that substitute the start of full function name
var RGX_FUNCNAME = regexp.MustCompile(`(.*\/)`)

//check if is array
func isMany(vType reflect.Type) bool {
	return vType.Kind() == reflect.Array || vType.Kind() == reflect.Slice
}

//check if some value is an array
func isValMany(value reflect.Value) bool {
	return isMany(value.Type())
}

//return function name
func GetFunctionName(i interface{}) string {

	reflect.ValueOf(i).String()

	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	return string(RGX_FUNCNAME.ReplaceAll([]byte(fullName), []byte{}))
}

//Retrieve ttl value, if interfaca implements cache.ExposeTTL
func discoverTTL(valRet interface{}, defaultTTL int) int {

	//switch valRel.(type)
	exposeTTL, isExposeTTL := valRet.(cache.ExposeTTL)

	ttl := defaultTTL

	if isExposeTTL {
		ttl = exposeTTL.GetTtl()
	}
	return ttl
}

//panic if is not a pointer
func mustBePointer(someValues ...interface{}) {
	for _, value := range someValues {
		if reflect.ValueOf(value).Kind() != reflect.Ptr {
			log.Error("Is not a pointer!!")
			panic(errors.New("Paramater needs to be a pointer!"))
		}
	}
}

//recursivelly iterate over a type until find a non array type
func getArrayInnerType(arrType reflect.Type) reflect.Type {
	if isMany(arrType) {
		return getArrayInnerType(arrType.Elem())
	} else {
		return arrType
	}
}

//return the the out types for a function
func getInOutTypes(someFunction reflect.Type) ([]reflect.Type, []reflect.Type) {
	if someFunction.Kind() == reflect.Ptr {
		someFunction = someFunction.Elem()
	}

	//recover the return types
	numOut := someFunction.NumOut()
	outTypes := make([]reflect.Type, numOut)
	for i := 0; i < numOut; i++ {
		outTypes[i] = someFunction.Out(i)
	}

	//recover the input types
	numIn := someFunction.NumIn()
	inTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		inTypes[i] = someFunction.In(i)
	}
	return inTypes, outTypes
}

func fromWrappedToArray(wrappedResponseArray reflect.Value) []reflect.Value {

	hotResponseLen := wrappedResponseArray.Len()
	hotReturnedValues := make([]reflect.Value, hotResponseLen)
	for i := 0; i < hotResponseLen; i++ {
		hotReturnedValues[i] = wrappedResponseArray.Index(i)
	}

	return hotReturnedValues
}

func fromArrayToWrapped(arrayVales []reflect.Value, typeDestinySlice reflect.Type) reflect.Value {
	wrappedArray := reflect.MakeSlice(typeDestinySlice, len(arrayVales), len(arrayVales))

	for index, value := range arrayVales {
		wrappedArray.Index(index).Set(value)
	}

	return wrappedArray
}

//recursivelly iterate over a type until find a non array type
func getArrayInnerTypes(arrTypes []reflect.Type) []reflect.Type {
	arrInnTypes := make([]reflect.Type, len(arrTypes))

	for i, arrType := range arrTypes {
		arrInnTypes[i] = getArrayInnerType(arrType)
	}

	return arrInnTypes
}

// Try to convert a int value to string. if is not a integer raise error
func valIntToString(value reflect.Value) (string, error) {

	var strVal string

	switch value.Type().Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		strVal = strconv.Itoa(int(value.Int()))
	default:
		log.Error("Error trying to convert value to string!", value)
		return "", errors.New("Error trying to convert value to string!" + value.String())
	}

	return strVal, nil
}

func mustBeCompatible(a, b reflect.Type) {
	if !a.ConvertibleTo(b) || !a.AssignableTo(b) {
		panic(errors.New(fmt.Sprintf(
			"Types '%v' and '%v' is not compatile to each other! "+
				"It is no possible to make a swap function that "+
				"return or receive different kinds of objects!", a.Name(), b.Name())))
	}
}
