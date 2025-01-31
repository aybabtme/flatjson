package flatjson_test

import (
	"fmt"

	"github.com/omnivore/flatjson"
)

func ExampleScanObject() {
	data := []byte(`{
		"hello":["world"],
		"bonjour": {"le": "monde"}
	}`)

	flatjson.ScanObject(data, 0, &flatjson.Callbacks{
		MaxDepth: 99,
		OnNumber: func(prefixes flatjson.Prefixes, val flatjson.Number) {
			fmt.Printf("path=%s\n", prefixes.AsString(data))
			if val.Name.IsObjectKey() {
				fmt.Printf("key=%s\n", val.Name.String(data))
			} else {
				fmt.Printf("index=%d\n", val.Name.Index())
			}
			fmt.Printf("value=%f\n", val.Value)
		},
		OnString: func(prefixes flatjson.Prefixes, val flatjson.String) {
			fmt.Printf("path=%s\n", prefixes.AsString(data))
			if val.Name.IsObjectKey() {
				fmt.Printf("key=%s\n", val.Name.String(data))
			} else {
				fmt.Printf("index=%d\n", val.Name.Index())
			}
			fmt.Printf("value=%q\n", val.Value.String(data))
		},
		OnBoolean: func(prefixes flatjson.Prefixes, val flatjson.Bool) {
			fmt.Printf("path=%s\n", prefixes.AsString(data))
			if val.Name.IsObjectKey() {
				fmt.Printf("key=%s\n", val.Name.String(data))
			} else {
				fmt.Printf("index=%d\n", val.Name.Index())
			}
			fmt.Printf("value=%v\n", val.Value)
		},
		OnNull: func(prefixes flatjson.Prefixes, val flatjson.Null) {
			fmt.Printf("path=%s\n", prefixes.AsString(data))
			if val.Name.IsObjectKey() {
				fmt.Printf("key=%s\n", val.Name.String(data))
			} else {
				fmt.Printf("index=%d\n", val.Name.Index())
			}
			fmt.Printf("NULL!")
		},
	})

	// Output:
	// path=hello
	// index=0
	// value="\"world\""
	// path=bonjour
	// key="le"
	// value="\"monde\""
}
