package flatjson

import (
	"reflect"
	"testing"
)

func TestScanObjects(t *testing.T) {
	tests := []struct {
		Name string

		Data string

		WantStart int
		WantEnd   int

		WantNumber []tnumber
		WantString []tstring
		WantBool   []tbool
		WantNull   []tnull

		WantErrError  string
		WantErrOffset int
	}{

		// happy path
		{
			Name:      "empty object",
			Data:      `{}`,
			WantStart: 0,
			WantEnd:   2,
		},
		{
			Name:      "simple string object",
			Data:      `{"hello":"world"}`,
			WantStart: 0,
			WantEnd:   17,
			WantString: []tstring{
				{name: `"hello"`, value: `"world"`},
			},
		},

		// special cases
		{
			Name: "empty object with whitespace",
			Data: `
    {

    }

`,
			WantStart: 5,
			WantEnd:   13,
		},

		// errors
		{
			Name:         "empty string",
			Data:         ``,
			WantErrError: noOpeningBracketFound,
		},
		{
			Name:          "only opening brakcet",
			Data:          `{`,
			WantErrError:  endOfDataNoClosingBracket,
			WantErrOffset: 1,
		},
	}

	for _, tt := range tests {
		t.Logf("====> %s", tt.Name)

		data := []byte(tt.Data)
		var gotNumber []tnumber
		onNumber := func(v Number) {
			gotNumber = append(gotNumber, tnumber{
				name:  v.Name.String(data),
				value: v.Value,
			})
		}
		var gotString []tstring
		onString := func(v String) {
			gotString = append(gotString, tstring{
				name:  v.Name.String(data),
				value: v.Value.String(data),
			})
		}
		var gotBool []tbool
		onBool := func(v Bool) {
			gotBool = append(gotBool, tbool{
				name:  v.Name.String(data),
				value: v.Value,
			})
		}
		var gotNull []tnull
		onNull := func(v Null) {
			gotNull = append(gotNull, tnull{
				name: v.Name.String(data),
			})
		}

		start, end, err := scanObject([]byte(data), onNumber, onString, onBool, onNull)

		gotErr, _ := err.(*SyntaxError)

		// if we expect errors
		if tt.WantErrError != "" && gotErr == nil {
			t.Errorf("want an error, got none")
		} else if tt.WantErrError != "" && gotErr != nil {
			wantOffset := tt.WantErrOffset
			if wantOffset != gotErr.Offset {
				t.Errorf("want err offset %d, was %d", wantOffset, gotErr.Offset)
			}
			if want, got := tt.WantErrError, gotErr.Error(); want != got {
				t.Errorf("want error: %q", want)
				t.Errorf(" got error: %q", got)
			}
		} else if gotErr != nil {
			t.Errorf("offset=%d", gotErr.Offset)
			t.Error(gotErr)
		} else {

			if want, got := tt.WantStart, start; want != got {
				t.Errorf("want start %+v", want)
				t.Errorf(" got start %+v", got)
			}

			if want, got := tt.WantEnd, end; want != got {
				t.Errorf("want end %+v", want)
				t.Errorf(" got end %+v", got)
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
		}
	}
}

type tnumber struct {
	name  string
	value float64
}

type tstring struct {
	name  string
	value string
}

type tbool struct {
	name  string
	value bool
}

type tnull struct {
	name string
}
