package bdjson

import (
	"fmt"
	"testing"
)

func TestSimpleJson(t *testing.T) {
	js, err := NewJson([]byte(`{
		"test": { 
			"string_slice": ["asdf", "ghjk", "zxcv"],
			"silce": [1, "2", 3],
			"silcewithsubs": [{"subkeyone": 1},
			{"subkeytwo": 2, "subkeythree": 3}],
			"int": 10,
			"float": 5.150,
			"bignum": 9223372036854775807,
			"string": "simplejson",
			"bool": true 
		},
		"test1": { 
			"string_slice": "sb"
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(js)

	val, ok := js.CheckGet("test1")
	if ok {
		fmt.Println(val)
	}

	silce, err := js.Get("test").Get("silce").Slice()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(silce)

	aws := js.Get("test").Get("silcewithsubs")
	fmt.Println(aws)

	awsval, err := aws.GetIndex(0).Get("subkeyone").Int()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(awsval)

	awsval, err = aws.GetIndex(1).Get("subkeytwo").Int()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(awsval)

	awsval, err = aws.GetIndex(1).Get("subkeythree").Int()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(awsval)

	i, err := js.Get("test").Get("int").Int()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(i)

	f, err := js.Get("test").Get("float").Float64()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(f)

	s, err := js.Get("test").Get("string").String()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)

	b, err := js.Get("test").Get("bool").Bool()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(b)

	strs, err := js.Get("test").Get("string_slice").StringSlice()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(strs[0])

	gp, err := js.GetLot("test", "string").String()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gp)

	gp1, err := js.GetLot("test", "int").Int()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(gp1)

	js.Set("test", "setTest")
	fmt.Println(js.Get("test"))
}
