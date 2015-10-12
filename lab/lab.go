package main

import "fmt"
import "unsafe"
import "strconv"



func castStr(v *string) string {
	return fmt.Sprint(uintptr(unsafe.Pointer(v)))
}

func uncastStr(s string) string {
	p, _ := strconv.ParseInt(s, 10, 64)
	return *((*string)(unsafe.Pointer(uintptr(p))))
}

func main() {
	onevar := "something"
	other := "something else"
	sa := []string{castStr(&onevar), castStr(&other)}

	for _, v := range sa {
		fmt.Printf("{{%s}}\n", v)
		fmt.Printf("%v\n", uncastStr(v))
	}

	//for _, v := range sa {
	//  vName := fmt.Sprintf("{{%s}}", v)
	//  msg = strings.Replace(msg, vName, uncastStr(v) -1)
	//}
}