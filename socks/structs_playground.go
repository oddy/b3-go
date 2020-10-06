package main

import (
	"fmt"
	"reflect"
)

type argh struct {
	Zz int `b3.type:"UTF8" b3.tag:"1"`
	Yy int
}

func roger(qq interface{}) {
	fmt.Println(qq)
	fmt.Println(reflect.TypeOf(qq))
	fmt.Println(reflect.TypeOf(qq).Kind())
	fmt.Println(reflect.ValueOf(qq))
	fmt.Println(reflect.ValueOf(qq).Kind())
	fmt.Println(reflect.ValueOf(qq).Elem())
	fmt.Println(reflect.ValueOf(qq).Elem().Kind())

	x := reflect.ValueOf(qq).Elem()
	fmt.Println("---")

	typ := reflect.TypeOf(qq).Elem()
	fmt.Println(typ)

	for i := 0 ; i< x.NumField() ; i++ {
		val := x.Field(i)
		fmt.Println(val)

		fmt.Println(val.IsValid())
		fmt.Println(val.CanSet())

		// fmt.Println(val.Tag.Get("b3.type"))  // this val doesn't have a Tag member.
		// we have to get them from the struct TYPE's fields, not the struct VALUE's fields.
		tfield := typ.Field(i)
		fmt.Println(" tfield ",tfield)
		fmt.Println(" tfield tag ",tfield.Tag)
		var x, ok = tfield.Tag.Lookup("b3.type")
		fmt.Println(" tfield tag ok ",ok)
		fmt.Println(" tfield tag x  ",x)
		fmt.Println(" tfield tag get b3 type ",tfield.Tag.Get("b3.type"))

		val999 := reflect.ValueOf(999)
		val.Set(val999)
	}
}

func _main() {
	x := argh{1111,2222}
	v := 8

	fmt.Println(x)
	fmt.Println("==========")

	roger(&x)

	fmt.Println("==========")
	fmt.Println(x)

	_ = v
	_ = x
}
