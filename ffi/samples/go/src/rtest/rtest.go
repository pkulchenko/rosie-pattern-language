//  -*- Mode: Go; -*-                                              
// 
//  rtest.go    Sample driver for librosie in go
// 
//  © Copyright IBM Corporation 2016.
//  LICENSE: MIT License (https://opensource.org/licenses/mit-license.html)
//  AUTHOR: Jamie A. Jennings


package main

// #cgo LDFLAGS: -L${SRCDIR}/libs ${SRCDIR}/libs/liblua.a
// #include <stdint.h>
// #include <stdarg.h>
// #include <stdlib.h>
// #include "librosie.h"
// struct string *new_string_ptr(int len, char *buf) {
//   struct string *cstr_ptr = malloc(sizeof(struct string));
//   cstr_ptr->len = (uint32_t) len;
//   uint8_t *abuf = (uint8_t *)buf; /*malloc(sizeof(uint8_t)*len);*/
//   cstr_ptr->ptr = abuf;
//   return cstr_ptr;
// }
// char *to_char_ptr(uint8_t *buf) {
//   return (char *) buf;
// }
// struct string *string_array_ref(struct stringArray a, int index) {
//   if (index > a.n) return NULL;
//   else return a.ptr[index];
// }
// #cgo CFLAGS: -std=gnu99 -I./include
import "C"

import "fmt"
import "encoding/json"
import "os"
import "strconv"

func structString_to_GoString(cstr C.struct_string) string {
	return C.GoStringN(C.to_char_ptr(cstr.ptr), C.int(cstr.len))
}

func gostring_to_structStringptr(s string) *C.struct_string {
	var cstr_ptr = C.new_string_ptr(C.int(len(s)), C.CString(s))
	return cstr_ptr
}

func main() {
	fmt.Printf("Hello, world.\n")
	
	var messages C.struct_stringArray
	
	home := gostring_to_structStringptr("/Users/jjennings/Work/Dev/rosie-pattern-language")
	engine, err := C.initialize(home, &messages)
	if engine==nil {
		fmt.Printf("Return value from initialize was NULL!")
		fmt.Printf("Err field returned by initialize was: %s\n", err)
		os.Exit(-1)
	}

	var a C.struct_stringArray
	cfg := gostring_to_structStringptr("{\"expression\":\"[:digit:]+\", \"encode\":\"json\"}")
	a, err = C.configure_engine(engine, cfg)
	retval := structString_to_GoString(*C.string_array_ref(a,0))
	fmt.Printf("Code from configure_engine: %s\n", retval)

	a, err = C.inspect_engine(engine)
	retval = structString_to_GoString(*C.string_array_ref(a,0))
	fmt.Printf("Code from inspect_engine: %s\n", retval)
	fmt.Printf("Config from inspect_engine: %s\n", structString_to_GoString(*C.string_array_ref(a,1)))
	C.free_stringArray(a)

	var foo string = "1111111111222222222211111111112222222222111111111122222222221111111111222222222211111111112222222222"
	foo_string := C.new_string_ptr(C.int(len(foo)), C.CString(foo))

	a, err = C.match(engine, foo_string, nil)
	retval = structString_to_GoString(*C.string_array_ref(a,0))
	fmt.Printf("Code from match: %s\n", retval)
	fmt.Printf("Data|false from match: %s\n", structString_to_GoString(*C.string_array_ref(a,1)))
	fmt.Printf("Leftover chars from match: %s\n", structString_to_GoString(*C.string_array_ref(a,2)))

	var r C.struct_stringArray
	var code, js_str string
	var leftover int

	foo = "1239999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999"
	foo_string = C.new_string_ptr(C.int(len(foo)), C.CString(foo))

	r = C.match(engine, foo_string, nil)
	code = structString_to_GoString(*C.string_array_ref(r,0))
	js_str = structString_to_GoString(*C.string_array_ref(r,1))
	leftover, err = strconv.Atoi(structString_to_GoString(*C.string_array_ref(r,2)))
	if code != "true" {
		fmt.Printf("Error in match: %s\n", js_str)
	} else {
		fmt.Printf("Code from match: %s\n", code)
		fmt.Printf("Data|false from match: %s\n", js_str)
		fmt.Printf("Leftover chars from match: %d\n", leftover)

		var retvals map[string]map[string]interface{}
		err = json.Unmarshal([]byte(js_str), &retvals)
		if err != nil {
			fmt.Println("JSON parse error:", err)
		}
		fmt.Printf("Match table: %s\n", retvals)
		fmt.Printf("Text from match table: %s\n", retvals["*"]["text"])
		fmt.Printf("Pos from match table: %d\n", int(retvals["*"]["pos"].(float64)))
		if retvals["*"]["subs"] != nil {
			fmt.Printf("Subs from match table: %s\n", retvals["*"]["subs"].(string))
		} else {
			fmt.Printf("No subs from match table.\n")
		}
	}
	C.free_stringArray(r)

	fmt.Printf(" done.\n");

	C.finalize(engine);


}
