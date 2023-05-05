package flatjson

import (
	"reflect"
	"testing"
)

func TestScanObjects(t *testing.T) {
	tests := []struct {
		Name string

		Data     string
		MaxDepth int

		WantPos   Pos
		WantFound bool

		WantNumber []tnumber
		WantString []tstring
		WantBool   []tbool
		WantNull   []tnull
		WantRaw    []traw

		WantErrError  string
		WantErrOffset int
	}{

		// happy path
		// {
		// 	Name: "empty string",
		// 	Data: ``,
		// },
		// {
		// 	Name:      "empty object",
		// 	Data:      `{}`,
		// 	WantPos:   Pos{0, 2},
		// 	WantFound: true,
		// },
		// {
		// 	Name:    "simple string object",
		// 	Data:    `{"hello":"world"}`,
		// 	WantPos: Pos{0, 17},
		// 	WantString: []tstring{
		// 		{name: `"hello"`, value: `"world"`},
		// 	},
		// 	WantRaw: []traw{
		// 		{name: `"hello"`, raw: `"world"`},
		// 	},
		// 	WantFound: true,
		// },
		// {
		// 	Name:    "simple number object",
		// 	Data:    `{"hello":-49.14159e-2}`,
		// 	WantPos: Pos{0, 22},
		// 	WantNumber: []tnumber{
		// 		{name: `"hello"`, value: -49.14159e-2},
		// 	},
		// 	WantRaw: []traw{
		// 		{name: `"hello"`, raw: `-49.14159e-2`},
		// 	},
		// 	WantFound: true,
		// },
		// {
		// 	Name:    "simple true bool object",
		// 	Data:    `{"hello":true}`,
		// 	WantPos: Pos{0, 14},
		// 	WantBool: []tbool{
		// 		{name: `"hello"`, value: true},
		// 	},
		// 	WantRaw: []traw{
		// 		{name: `"hello"`, raw: `true`},
		// 	},
		// 	WantFound: true,
		// },
		// {
		// 	Name:    "simple false bool object",
		// 	Data:    `{"hello":false}`,
		// 	WantPos: Pos{0, 15},
		// 	WantBool: []tbool{
		// 		{name: `"hello"`, value: false},
		// 	},
		// 	WantRaw: []traw{
		// 		{name: `"hello"`, raw: `false`},
		// 	},
		// 	WantFound: true,
		// },
		// {
		// 	Name:    "simple null object",
		// 	Data:    `{"hello":null}`,
		// 	WantPos: Pos{0, 14},
		// 	WantNull: []tnull{
		// 		{name: `"hello"`},
		// 	},
		// 	WantRaw: []traw{
		// 		{name: `"hello"`, raw: `null`},
		// 	},
		// 	WantFound: true,
		// },

		{
			Name:    "simple composite object",
			Data:    `{"a":"1","b":2,"c":true,"d":false,"e":null,"f":{},"g":[]}`,
			WantPos: Pos{0, 57},
			WantString: []tstring{
				{name: `"a"`, value: `"1"`},
			},
			WantNumber: []tnumber{
				{name: `"b"`, value: 2},
			},
			WantBool: []tbool{
				{name: `"c"`, value: true},
				{name: `"d"`, value: false},
			},
			WantNull: []tnull{
				{name: `"e"`},
			},
			WantRaw: []traw{
				{name: `"a"`, raw: `"1"`},
				{name: `"b"`, raw: `2`},
				{name: `"c"`, raw: `true`},
				{name: `"d"`, raw: `false`},
				{name: `"e"`, raw: `null`},
				{name: `"f"`, raw: `{}`},
				{name: `"g"`, raw: `[]`},
			},
			WantFound: true,
		},

		{
			Name:     "nested composite object are flat",
			MaxDepth: 0,
			Data:     `{ "a":[{"b":true},{"c":{}}] }`,
			WantPos:  Pos{0, 29},
			WantRaw: []traw{
				{name: `"a"`, raw: `[{"b":true},{"c":{}}]`},
			},
			WantFound: true,
		},
		{
			Name:     "nested composite object goes only 1 deep",
			MaxDepth: 1,
			Data:     `{ "a":[{"b":true},{"c":{}}] }`,
			WantPos:  Pos{0, 29},
			WantRaw: []traw{
				{pfx: "a", name: `0`, raw: `{"b":true}`},
				{pfx: "a", name: `1`, raw: `{"c":{}}`},
				{name: `"a"`, raw: `[{"b":true},{"c":{}}]`},
			},
			WantFound: true,
		},
		{
			Name:     "nested composite object recurse",
			MaxDepth: 99,
			Data:     `{ "a":[{"b":true},{"c":{}}] }`,
			WantPos:  Pos{0, 29},
			WantBool: []tbool{
				{pfx: "a.0", name: `"b"`, value: true},
			},
			WantRaw: []traw{
				{pfx: "a.0", name: `"b"`, raw: `true`},
				{pfx: "a", name: `0`, raw: `{"b":true}`},
				{pfx: "a.1", name: `"c"`, raw: `{}`},
				{pfx: "a", name: `1`, raw: `{"c":{}}`},
				{name: `"a"`, raw: `[{"b":true},{"c":{}}]`},
			},
			WantFound: true,
		},
		{
			Name:     "nested composite object with whitespace",
			MaxDepth: 99,
			Data: `{
				"key":{
					"key2": {
						"deep": 2
					}
				},
				"key2":[
					"myname",
					42,
					true,
					{"is":"antoine"}
				]
			}`,
			WantPos: Pos{0, 139},
			WantNumber: []tnumber{
				{pfx: "key.key2", name: `"deep"`, value: 2},
				{pfx: "key2", name: `1`, value: 42},
			},
			WantString: []tstring{
				{pfx: "key2", name: `0`, value: `"myname"`},
				{pfx: "key2.3", name: `"is"`, value: `"antoine"`},
			},
			WantBool: []tbool{
				{pfx: "key2", name: `2`, value: true},
			},
			WantRaw: []traw{
				{pfx: "key.key2", name: `"deep"`, raw: "2"},
				{pfx: "key", name: `"key2"`, raw: "{\n\t\t\t\t\t\t\"deep\": 2\n\t\t\t\t\t}"},
				{pfx: "", name: `"key"`, raw: "{\n\t\t\t\t\t\"key2\": {\n\t\t\t\t\t\t\"deep\": 2\n\t\t\t\t\t}\n\t\t\t\t}"},
				{pfx: "key2", name: "0", raw: `"myname"`},
				{pfx: "key2", name: "1", raw: `42`},
				{pfx: "key2", name: "2", raw: `true`},
				{pfx: "key2.3", name: `"is"`, raw: `"antoine"`},
				{pfx: "key2", name: "3", raw: `{"is":"antoine"}`},
				{pfx: "", name: `"key2"`, raw: "[\n\t\t\t\t\t\"myname\",\n\t\t\t\t\t42,\n\t\t\t\t\ttrue,\n\t\t\t\t\t{\"is\":\"antoine\"}\n\t\t\t\t]"},
			},
			WantFound: true,
		},

		{
			Name: "composite object with whitespace",
			Data: `
            {
                "a" :   "1",
                "b" :   2,
                "c" :   true,
                "d" :   false,
                "e":    null,
				"f":    {},
				"g":    []
}`,
			WantPos: Pos{13, 194},
			WantString: []tstring{
				{name: `"a"`, value: `"1"`},
			},
			WantNumber: []tnumber{
				{name: `"b"`, value: 2},
			},
			WantBool: []tbool{
				{name: `"c"`, value: true},
				{name: `"d"`, value: false},
			},
			WantNull: []tnull{
				{name: `"e"`},
			},
			WantRaw: []traw{
				{name: `"a"`, raw: `"1"`},
				{name: `"b"`, raw: `2`},
				{name: `"c"`, raw: `true`},
				{name: `"d"`, raw: `false`},
				{name: `"e"`, raw: `null`},
				{name: `"f"`, raw: `{}`},
				{name: `"g"`, raw: `[]`},
			},
			WantFound: true,
		},

		{
			Name: "composite object with weird whitespace",
			Data: `
            {
                "a" :   "1"
                ,"b" :   2,
                "c" :true ,
                "d" :   false
                ,
                "e":    null,
				"f":            {},
				"g":  [    ]
}`,
			WantPos: Pos{13, 219},
			WantString: []tstring{
				{name: `"a"`, value: `"1"`},
			},
			WantNumber: []tnumber{
				{name: `"b"`, value: 2},
			},
			WantBool: []tbool{
				{name: `"c"`, value: true},
				{name: `"d"`, value: false},
			},
			WantNull: []tnull{
				{name: `"e"`},
			},
			WantRaw: []traw{
				{name: `"a"`, raw: `"1"`},
				{name: `"b"`, raw: `2`},
				{name: `"c"`, raw: `true`},
				{name: `"d"`, raw: `false`},
				{name: `"e"`, raw: `null`},
				{name: `"f"`, raw: `{}`},
				{name: `"g"`, raw: `[    ]`},
			},
			WantFound: true,
		},

		// special cases
		{
			Name: "empty object with whitespace",
			Data: `
    {

    }

`,
			WantPos:   Pos{5, 13},
			WantFound: true,
		},

		// errors
		{
			Name:          "only opening brakcet",
			Data:          `{`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 1,
		},
		{
			Name:          "single pair with no closing bracket (number)",
			Data:          `{"hello":0`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 11,
		},
		{
			Name:          "single pair with no closing bracket (bool)",
			Data:          `{"hello":true`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 14,
		},
		{
			Name:          "single pair with no closing bracket (object)",
			Data:          `{"hello":{}`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 12,
		},
		{
			Name:          "single pair with no closing bracket (array)",
			Data:          `{"hello":[]`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 12,
		},
		{
			Name:          "single pair with no closing bracket (bool and comma)",
			Data:          `{"hello":true,`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 14,
		},
		{
			Name:          "single pair with no closing bracket (object and comma)",
			Data:          `{"hello":{},`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 12,
		},
		{
			Name:          "single pair with no closing bracket (array and comma)",
			Data:          `{"hello":[],`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 12,
		},
		{
			Name:          "single pair with no closing bracket (number) and space",
			Data:          `{"hello":0 `,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 12,
		},
		{
			Name:          "single pair with no closing bracket (bool) and space",
			Data:          `{"hello":true `,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 15,
		},
		{
			Name:          "single pair with no closing bracket (object) and space",
			Data:          `{"hello": {} `,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 14,
		},
		{
			Name:          "single pair with no closing bracket (array) and space",
			Data:          `{"hello": [] `,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 14,
		},
		{
			Name:          "single pair with no closing bracket (bool and comma)",
			Data:          `{"hello":true, `,
			WantErrError:  endOfDataNoNamePair,
			WantErrOffset: 15,
		},

		{
			Name:          "missing name in name/value pair",
			Data:          `{:true, `,
			WantErrError:  expectingNameBeforeValue + ", " + reachedEndScanningCharacters,
			WantErrOffset: 1,
		},
		{
			Name:          "missing colon in name/value pair",
			Data:          `{"hello" true, `,
			WantErrError:  noColonFound,
			WantErrOffset: 9,
		},
		{
			Name:          "nothing follows the name",
			Data:          `{"hello" `,
			WantErrError:  endOfDataNoColon,
			WantErrOffset: 9,
		},
		{
			Name:          "nothing follows the colon",
			Data:          `{"hello": `,
			WantErrError:  endOfDataNoValueForName,
			WantErrOffset: 10,
		},
		{
			Name:          "malformed number value (garbage)",
			Data:          `{"hello": 7162hhhh}`,
			WantErrError:  malformedNumber,
			WantErrOffset: 10,
		},
		{
			Name:          "malformed number value (incomplete)",
			Data:          `{"hello": 7162.}`,
			WantErrError:  beginNumberValueButError + ", " + scanningForFraction + ", " + needAtLeastOneDigit,
			WantErrOffset: 10,
		},
		{
			Name:          "malformed string value (incomplete)",
			Data:          `{"hello": "world}`,
			WantErrError:  beginStringValueButError + ", " + reachedEndScanningCharacters,
			WantErrOffset: 10,
		},
		{
			Name:          "malformed array value (incomplete)",
			Data:          `{"hello": [}`,
			WantErrError:  beginArrayValueButError + ", " + expectValueButNoKnownType,
			WantErrOffset: 10,
		},
		{
			Name:          "random crap for value",
			Data:          `{"hello": lololool}`,
			WantErrError:  expectValueButNoKnownType,
			WantErrOffset: 10,
		},

		{
			Name:          "no closing bracket at end of object",
			Data:          `{"hello": "hello"   `,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 21,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			data := []byte(tt.Data)
			var gotNumber []tnumber
			onNumber := func(pfx Prefixes, v Number) {
				gotNumber = append(gotNumber, tnumber{
					pfx:   pfx.AsString(data),
					name:  v.Name.String(data),
					value: v.Value,
				})
			}
			var gotString []tstring
			onString := func(pfx Prefixes, v String) {
				gotString = append(gotString, tstring{
					pfx:   pfx.AsString(data),
					name:  v.Name.String(data),
					value: v.Value.String(data),
				})
			}
			var gotBool []tbool
			onBool := func(pfx Prefixes, v Bool) {
				gotBool = append(gotBool, tbool{
					pfx:   pfx.AsString(data),
					name:  v.Name.String(data),
					value: v.Value,
				})
			}
			var gotNull []tnull
			onNull := func(pfx Prefixes, v Null) {
				gotNull = append(gotNull, tnull{
					pfx:  pfx.AsString(data),
					name: v.Name.String(data),
				})
			}
			var gotRaw []traw
			onRaw := func(pfx Prefixes, key Prefix, value Pos) {
				v := traw{
					pfx:  pfx.AsString(data),
					name: key.String(data),
					raw:  value.String(data),
				}
				gotRaw = append(gotRaw, v)
			}

			pos, found, err := ScanObject([]byte(data), 0, &Callbacks{
				MaxDepth:  tt.MaxDepth,
				OnNumber:  onNumber,
				OnString:  onString,
				OnBoolean: onBool,
				OnNull:    onNull,
				OnRaw:     onRaw,
			})

			if tt.WantFound != found {
				t.Errorf("want found %+v", tt.WantFound)
				t.Errorf(" got found %+v", found)
			}

			// if we expect errors
			if tt.WantErrError != "" && err == nil {
				t.Errorf("want an error, got none")
			} else if tt.WantErrError != "" && err != nil {
				gotErr, _ := err.(*SyntaxError)
				wantOffset := tt.WantErrOffset
				if wantOffset != gotErr.Offset {
					t.Errorf("want err offset %d, was %d", wantOffset, gotErr.Offset)
				}
				if want, got := tt.WantErrError, gotErr.Error(); want != got {
					t.Errorf("want error: %q", want)
					t.Errorf(" got error: %q", got)
				}
			} else if err != nil {
				gotErr, _ := err.(*SyntaxError)
				t.Errorf("offset=%d", gotErr.Offset)
				t.Error(gotErr)
			} else {

				if want, got := tt.WantPos, pos; want != got {
					t.Errorf("want position %+v", want)
					t.Errorf(" got position %+v", got)
				}

				if want, got := tt.WantNumber, gotNumber; !reflect.DeepEqual(want, got) {
					t.Errorf("want number %+v", want)
					t.Errorf(" got number %+v", got)
				}

				if want, got := tt.WantString, gotString; !reflect.DeepEqual(want, got) {
					t.Errorf("want string %+v", want)
					t.Errorf(" got string %+v", got)
				}

				if want, got := tt.WantBool, gotBool; !reflect.DeepEqual(want, got) {
					t.Errorf("want bool %+v", want)
					t.Errorf(" got bool %+v", got)
				}

				if want, got := tt.WantNull, gotNull; !reflect.DeepEqual(want, got) {
					t.Errorf("want null %+v", want)
					t.Errorf(" got null %+v", got)
				}

				if want, got := tt.WantRaw, gotRaw; !reflect.DeepEqual(want, got) {
					t.Errorf("want raw %+v", want)
					t.Errorf(" got raw %+v", got)
				}
			}
		})
	}
}

type tnumber struct {
	pfx   string
	name  string
	value float64
}

type tstring struct {
	pfx   string
	name  string
	value string
}

type tbool struct {
	pfx   string
	name  string
	value bool
}

type tnull struct {
	pfx  string
	name string
}

type traw struct {
	pfx  string
	name string
	raw  string
}
