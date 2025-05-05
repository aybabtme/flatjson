package flatjson

import (
	"errors"
	"math"
	"testing"
)

func TestScanStrings(t *testing.T) {
	tests := []struct {
		Name string

		Start int
		Data  string

		WantVal Pos

		WantErrError  string
		WantErrOffset int
	}{
		{
			Name:    "empty",
			Data:    `""`,
			WantVal: Pos{0, 2},
		},
		{
			Name:    "1 char",
			Data:    `"1"`,
			WantVal: Pos{0, 3},
		},
		{
			Name:    "short sentence",
			Data:    `"once upon a time"`,
			WantVal: Pos{0, 18},
		},
		{
			Name:    "special chars",
			Data:    `"\ \" \\ \/ \b \f \n \r \t \u1111 "`,
			WantVal: Pos{0, 35},
		},
		{
			Name:    "special chars",
			Data:    `"\ \" \\ \/ \b \f \n \r \t \u1111 "`,
			WantVal: Pos{0, 35},
		},

		{
			Name:    "empty with garbage at the end",
			Data:    `"" hjbjhbjkhbehjwb 8y97  898 \n \n `,
			WantVal: Pos{0, 2},
		},
		{
			Name:    "1 char with garbage at the end",
			Data:    `"1" hjbjhbjkhbehjwb 8y97  898 \n \n `,
			WantVal: Pos{0, 3},
		},
		{
			Name:    "short sentence with garbage at the end",
			Data:    `"once upon a time" hjbjhbjkhbehjwb 8y97  898 \n \n `,
			WantVal: Pos{0, 18},
		},
		{
			Name:    "special chars with garbage at the end",
			Data:    `"\ \" \\ \/ \b \f \n \r \t \u1111 " hjbjhbjkhbehjwb 8y97  898 \n \n `,
			WantVal: Pos{0, 35},
		},
		{
			Name:    "special chars with garbage at the end",
			Data:    `"\ \" \\ \/ \b \f \n \r \t \u1111 " hjbjhbjkhbehjwb 8y97  898 \n \n `,
			WantVal: Pos{0, 35},
		},

		// errors
		{
			Name:          "empty content",
			Data:          "",
			WantErrError:  reachedEndScanningCharacters,
			WantErrOffset: 0,
		},
		{
			Name:          "unterminated escape sequence",
			Data:          `" lol \`,
			WantErrError:  reachedEndScanningCharacters,
			WantErrOffset: 7,
		},
		{
			Name:          "bad unicode escape code",
			Data:          `" lol \u333R "`,
			WantErrError:  unicodeNotFollowHex,
			WantErrOffset: 11,
		},
		{
			Name:          "unterminated unicode escape",
			Data:          `" lol \u`,
			WantErrError:  reachedEndScanningCharacters,
			WantErrOffset: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {

			gotVal, err := scanString([]byte(tt.Data), tt.Start)

			// if we expect errors
			if tt.WantErrError != "" && err == nil {
				t.Errorf("want an error, got none")
			} else if tt.WantErrError != "" && err != nil {
				gotErr := err.(*SyntaxError)
				wantOffset := tt.WantErrOffset
				if wantOffset != gotErr.Offset {
					t.Errorf("want err offset %d, was %d", wantOffset, gotErr.Offset)
				}
				if want, got := tt.WantErrError, gotErr.Error(); want != got {
					t.Errorf("want error: %q", want)
					t.Errorf(" got error: %q", got)
				}
			} else if err != nil {
				gotErr := err.(*SyntaxError)
				t.Errorf("offset=%d", gotErr.Offset)
				t.Error(gotErr)
			}

			if want, got := tt.WantVal, gotVal; want != got {
				t.Errorf("want val %+v", want)
				t.Errorf(" got val %+v", got)
			}
		})
	}
}

func fequal(a, b float64) bool {
	if a == b {
		return true
	}

	return math.Abs(math.Abs(a-b)/math.Max(a, b)) < 0.000001
}

func TestScanNumbersErrors(t *testing.T) {
	tests := []struct {
		Name string

		Start int
		Data  string

		WantErrOffset int
		WantErrError  string
	}{
		{
			Name:         "empty string",
			Data:         "",
			WantErrError: reachedEndScanningNumber,
		},
		{
			Name:         "no digits",
			Data:         "lol",
			WantErrError: cantFindIntegerPart,
		},
		{
			Name:          "just a sign",
			Data:          "-",
			WantErrError:  reachedEndScanningNumber,
			WantErrOffset: 1,
		},
		{
			Name:          "just a sign and a dot",
			Data:          "-.",
			WantErrError:  cantFindIntegerPart,
			WantErrOffset: 1,
		},
		{
			Name:          "just a sign, a dot and an empty exponent",
			Data:          "-.e",
			WantErrError:  cantFindIntegerPart,
			WantErrOffset: 1,
		},
		{
			Name:          "just a sign, a dot and an empty signed exponent",
			Data:          "-.e-",
			WantErrError:  cantFindIntegerPart,
			WantErrOffset: 1,
		},
		{
			Name:          "just a sign, a dot and an signed exponent",
			Data:          "-.e-42",
			WantErrError:  cantFindIntegerPart,
			WantErrOffset: 1,
		},
		{
			Name:          "just a sign, a 0 and a dot",
			Data:          "-0.",
			WantErrError:  scanningForFraction + ", " + reachedEndScanningDigit,
			WantErrOffset: 3,
		},
		{
			Name:          "missing digits in fraction",
			Data:          "102.",
			WantErrError:  scanningForFraction + ", " + reachedEndScanningDigit,
			WantErrOffset: 4,
		},

		{
			Name:          "missing digits in exponent",
			Data:          "102e",
			WantErrError:  scanningForExponentSign,
			WantErrOffset: 4,
		},

		{
			Name:          "missing digits in signed exponent",
			Data:          "102e+",
			WantErrError:  scanningForExponent + ", " + reachedEndScanningDigit,
			WantErrOffset: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, _, _, _, err := ScanNumber([]byte(tt.Data), tt.Start)

			gotErr := err.(*SyntaxError)
			if tt.WantErrError != "" && gotErr == nil {
				t.Fatalf("want an error, got none")
			}

			wantOffset := tt.WantErrOffset
			if wantOffset != gotErr.Offset {
				t.Errorf("want err offset %d, was %d", wantOffset, gotErr.Offset)
			}
			if want, got := tt.WantErrError, gotErr.Error(); want != got {
				t.Errorf("want error: %q", want)
				t.Errorf(" got error: %q", got)
			}
		})
	}
}

func TestScanNumbersNoError(t *testing.T) {
	tests := []struct {
		Name string

		Start int
		Data  string

		WantF64   float64
		WantI64   int64
		WantIsInt bool
		WantEnd   int
	}{
		// debugging

		// no exponent
		{
			Name:      "integer, 0",
			Data:      "0",
			WantF64:   0,
			WantI64:   0,
			WantIsInt: true,
			WantEnd:   1,
		},
		{
			Name:      "integer, 3",
			Data:      "3",
			WantF64:   0,
			WantI64:   3,
			WantIsInt: true,
			WantEnd:   1,
		},
		{
			Name:      "integer, 42",
			Data:      "42",
			WantF64:   0,
			WantI64:   42,
			WantIsInt: true,
			WantEnd:   2,
		},
		{
			Name:      "integer, 9000",
			Data:      "9000",
			WantF64:   0,
			WantI64:   9000,
			WantIsInt: true,
			WantEnd:   4,
		},
		{
			Name:      "negative integer, -0",
			Data:      "-0",
			WantF64:   0,
			WantI64:   -0,
			WantIsInt: true,
			WantEnd:   2,
		},
		{
			Name:      "negative integer, -3",
			Data:      "-3",
			WantF64:   0,
			WantI64:   -3,
			WantIsInt: true,
			WantEnd:   2,
		},
		{
			Name:      "negative integer, -42",
			Data:      "-42",
			WantF64:   0,
			WantI64:   -42,
			WantIsInt: true,
			WantEnd:   3,
		},
		{
			Name:      "negative integer, -9000",
			Data:      "-9000",
			WantF64:   0,
			WantI64:   -9000,
			WantIsInt: true,
			WantEnd:   5,
		},
		{
			Name:    "real numbers, around 0",
			Data:    "0.14159",
			WantF64: 0.14159,
			WantEnd: 7,
		},
		{
			Name:    "real numbers, around 3",
			Data:    "3.14159",
			WantF64: 3.14159,
			WantEnd: 7,
		},
		{
			Name:    "real numbers, around 42",
			Data:    "42.14159",
			WantF64: 42.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around 9000",
			Data:    "9000.14159",
			WantF64: 9000.14159,
			WantEnd: 10,
		},
		{
			Name:    "real numbers, around -0",
			Data:    "-0.14159",
			WantF64: -0.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around -3",
			Data:    "-3.14159",
			WantF64: -3.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around -42",
			Data:    "-42.14159",
			WantF64: -42.14159,
			WantEnd: 9,
		},
		{
			Name:    "real numbers, around -9000",
			Data:    "-9000.14159",
			WantF64: -9000.14159,
			WantEnd: 11,
		},

		// with a positive exponent
		{
			Name:      "positive exponent, integer, 0",
			Data:      "0e42",
			WantF64:   0e42,
			WantI64:   0e42,
			WantIsInt: true,
			WantEnd:   4,
		},
		{
			Name:      "positive exponent, integer, 3",
			Data:      "3e42",
			WantF64:   3e42,
			WantIsInt: false, // would overflow
			WantEnd:   4,
		},
		{
			Name:      "positive exponent, integer, 42",
			Data:      "42e42",
			WantF64:   42e42,
			WantIsInt: false, // would overflow
			WantEnd:   5,
		},
		{
			Name:      "positive exponent, integer, 9000",
			Data:      "9000e42",
			WantF64:   9000e42,
			WantIsInt: false, // would overflow
			WantEnd:   7,
		},
		{
			Name:      "positive exponent, negative integer, -0",
			Data:      "-0e42",
			WantF64:   -0e42,
			WantI64:   -0e42,
			WantIsInt: true,
			WantEnd:   5,
		},
		{
			Name:      "positive exponent, negative integer, -3",
			Data:      "-3e42",
			WantF64:   -3e42,
			WantIsInt: false, // has fraction
			WantEnd:   5,
		},
		{
			Name:      "positive exponent, negative integer, -42",
			Data:      "-42e42",
			WantF64:   -42e42,
			WantIsInt: false, // has fraction
			WantEnd:   6,
		},
		{
			Name:      "positive exponent, negative integer, -9000",
			Data:      "-9000e42",
			WantF64:   -9000e42,
			WantIsInt: false, // has fraction
			WantEnd:   8,
		},
		{
			Name:    "positive exponent, real numbers, around 0",
			Data:    "0.14159e42",
			WantF64: 0.14159e42,
			WantEnd: 10,
		},
		{
			Name:    "positive exponent, real numbers, around 3",
			Data:    "3.14159e42",
			WantF64: 3.14159e42,
			WantEnd: 10,
		},
		{
			Name:    "positive exponent, real numbers, around 42",
			Data:    "42.14159e42",
			WantF64: 42.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around 9000",
			Data:    "9000.14159e42",
			WantF64: 9000.14159e42,
			WantEnd: 13,
		},
		{
			Name:    "positive exponent, real numbers, around -0",
			Data:    "-0.14159e42",
			WantF64: -0.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around -3",
			Data:    "-3.14159e42",
			WantF64: -3.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around -42",
			Data:    "-42.14159e42",
			WantF64: -42.14159e42,
			WantEnd: 12,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159e42",
			WantF64: -9000.14159e42,
			WantEnd: 14,
		},

		// positive exponent variations
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159E42",
			WantF64: -9000.14159e42,
			WantEnd: 14,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159e+42",
			WantF64: -9000.14159e42,
			WantEnd: 15,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159E+42",
			WantF64: -9000.14159e42,
			WantEnd: 15,
		},

		// with a negative exponent
		{
			Name:      "negative exponent, integer, 0",
			Data:      "0e-42",
			WantF64:   0e-42,
			WantI64:   0e-42,
			WantIsInt: true,
			WantEnd:   5,
		},
		{
			Name:    "negative exponent, integer, 3",
			Data:    "3e-42",
			WantF64: 3e-42,
			WantEnd: 5,
		},
		{
			Name:    "negative exponent, integer, 42",
			Data:    "42e-42",
			WantF64: 42e-42,
			WantEnd: 6,
		},
		{
			Name:    "negative exponent, integer, 9000",
			Data:    "9000e-42",
			WantF64: 9000e-42,
			WantEnd: 8,
		},
		{
			Name:      "negative exponent, negative integer, -0",
			Data:      "-0e-42",
			WantF64:   -0e-42,
			WantI64:   -0e-42,
			WantIsInt: true,
			WantEnd:   6,
		},
		{
			Name:    "negative exponent, negative integer, -3",
			Data:    "-3e-42",
			WantF64: -3e-42,
			WantEnd: 6,
		},
		{
			Name:    "negative exponent, negative integer, -42",
			Data:    "-42e-42",
			WantF64: -42e-42,
			WantEnd: 7,
		},
		{
			Name:    "negative exponent, negative integer, -9000",
			Data:    "-9000e-42",
			WantF64: -9000e-42,
			WantEnd: 9,
		},
		{
			Name:    "negative exponent, real numbers, around 0",
			Data:    "0.14159e-42",
			WantF64: 0.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "negative exponent, real numbers, around 3",
			Data:    "3.14159e-42",
			WantF64: 3.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "negative exponent, real numbers, around 42",
			Data:    "42.14159e-42",
			WantF64: 42.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around 9000",
			Data:    "9000.14159e-42",
			WantF64: 9000.14159e-42,
			WantEnd: 14,
		},
		{
			Name:    "negative exponent, real numbers, around -0",
			Data:    "-0.14159e-42",
			WantF64: -0.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around -3",
			Data:    "-3.14159e-42",
			WantF64: -3.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around -42",
			Data:    "-42.14159e-42",
			WantF64: -42.14159e-42,
			WantEnd: 13,
		},
		{
			Name:    "negative exponent, real numbers, around -9000",
			Data:    "-9000.14159e-42",
			WantF64: -9000.14159e-42,
			WantEnd: 15,
		},
		{
			Name:    "negative exponent with variation, real numbers, around -9000",
			Data:    "-9000.14159E-42",
			WantF64: -9000.14159e-42,
			WantEnd: 15,
		},

		// with a garbage and negative exponent
		{
			Name:      "garbage and negative exponent, integer, 0",
			Data:      "0e-42 yguhbhg2  23h23 2j3h ",
			WantF64:   0e-42,
			WantI64:   0e-42,
			WantIsInt: true,
			WantEnd:   5,
		},
		{
			Name:    "garbage and negative exponent, integer, 3",
			Data:    "3e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 3e-42,
			WantEnd: 5,
		},
		{
			Name:    "garbage and negative exponent, integer, 42",
			Data:    "42e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 42e-42,
			WantEnd: 6,
		},
		{
			Name:    "garbage and negative exponent, integer, 9000",
			Data:    "9000e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 9000e-42,
			WantEnd: 8,
		},
		{
			Name:      "garbage and negative exponent, negative integer, -0",
			Data:      "-0e-42 yguhbhg2  23h23 2j3h ",
			WantF64:   -0e-42,
			WantI64:   -0e-42,
			WantIsInt: true,
			WantEnd:   6,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -3",
			Data:    "-3e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -3e-42,
			WantEnd: 6,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -42",
			Data:    "-42e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -42e-42,
			WantEnd: 7,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -9000",
			Data:    "-9000e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -9000e-42,
			WantEnd: 9,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 0",
			Data:    "0.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 0.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 3",
			Data:    "3.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 3.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 42",
			Data:    "42.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 42.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 9000",
			Data:    "9000.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: 9000.14159e-42,
			WantEnd: 14,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -0",
			Data:    "-0.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -0.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -3",
			Data:    "-3.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -3.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -42",
			Data:    "-42.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -42.14159e-42,
			WantEnd: 13,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -9000",
			Data:    "-9000.14159e-42 yguhbhg2  23h23 2j3h ",
			WantF64: -9000.14159e-42,
			WantEnd: 15,
		},
		{
			Name:    "garbage and negative exponent with variation, real numbers, around -9000",
			Data:    "-9000.14159E-42 yguhbhg2  23h23 2j3h ",
			WantF64: -9000.14159e-42,
			WantEnd: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotF64, gotI64, gotIsInt, gotEnd, gotErr := ScanNumber([]byte(tt.Data), tt.Start)

			if gotErr != nil {
				t.Fatal(gotErr)
			}

			if want, got := tt.WantEnd, gotEnd; want != got {
				t.Errorf("want advance to %d, got %d", want, got)
			}
			if want, got := tt.WantF64, gotF64; !fequal(want, got) {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
			if want, got := tt.WantI64, gotI64; want != got {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
			if want, got := tt.WantIsInt, gotIsInt; want != got {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
		})
	}
}

func TestScanNumberF64(t *testing.T) {
	tests := []struct {
		Name string

		Start int
		Data  string

		WantF64   float64
		WantI64   int64
		WantIsInt bool
		WantEnd   int
		WantError error
	}{
		{
			Name:    "real number, 0.5",
			Data:    "0.5",
			WantF64: 0.5,
			WantEnd: 3,
		},
		{
			Name:    "real number, 1.1",
			Data:    "1.1",
			WantF64: 1.1,
			WantEnd: 3,
		},
		{
			Name:    "real number, 2.1",
			Data:    "2.1",
			WantF64: 2.1,
			WantEnd: 3,
		},
		{
			Name:    "real number, 3.7",
			Data:    "3.7",
			WantF64: 3.7,
			WantEnd: 3,
		},
		{
			Name:    "real number, 4.1",
			Data:    "4.1",
			WantF64: 4.1,
			WantEnd: 3,
		},
		{
			Name:    "real number, 5.9",
			Data:    "5.9",
			WantF64: 5.9,
			WantEnd: 3,
		},
		{
			Name:    "real number, 1.0001",
			Data:    "1.0001",
			WantF64: 1.0001,
			WantEnd: 6,
		},
		{
			Name:    "real number, 1.010020",
			Data:    "1.010020",
			WantF64: 1.010020,
			WantEnd: 8,
		},
		{
			Name:    "real number, 600.12345",
			Data:    "600.12345",
			WantF64: 600.12345,
			WantEnd: 9,
		},
		{
			Name:    "real number, 1234567890123456789.0",
			Data:    "1234567890123456789.0",
			WantF64: 1234567890123456789.0,
			WantEnd: 21,
		},
		{
			Name:      "real number, 0.785398163397448278999491",
			Data:      "0.785398163397448278999491",
			WantF64:   0.0,
			WantEnd:   21, // after decimal point, max length of scannable digits is 19 since max int64 is 9223372036854775807
			WantError: ErrScanTooLargeNumber,
		},
		{
			Name:    "real number, 10.7853981633974482789",
			Data:    "10.7853981633974482789",
			WantF64: 10.7853981633974482789,
			WantEnd: 22,
		},
		{
			Name:    "real number, 1234567890123456789.1234567890123456789",
			Data:    "1234567890123456789.1234567890123456789",
			WantF64: 1234567890123456789.1234567890123456789,
			WantEnd: 39,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotF64, gotI64, gotIsInt, gotEnd, gotErr := ScanNumber([]byte(tt.Data), tt.Start)

			if wantErr, gotErr := tt.WantError, gotErr; wantErr != nil && gotErr != nil {
				t.Logf("gotErr: %q", gotErr.Error())
				if !errors.Is(gotErr, wantErr) {
					t.Errorf("want error: %q\n", wantErr.Error())
					t.Errorf("got error: %q\n", gotErr.Error())
				}
			}

			if want, got := tt.WantEnd, gotEnd; want != got {
				t.Errorf("want advance to %d, got %d", want, got)
			}
			if want, got := tt.WantF64, gotF64; !fequal(want, got) {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
			if want, got := tt.WantI64, gotI64; want != got {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
			if want, got := tt.WantIsInt, gotIsInt; want != got {
				t.Errorf("want val %v", want)
				t.Errorf(" got val %v", got)
			}
		})
	}
}

func TestScanDigits(t *testing.T) {
	tests := []struct {
		Name string

		Start int
		Data  string

		WantI64 int64
		WantEnd int

		WantErrError string
	}{
		// all good
		{
			Name:    "only zero",
			Data:    "0",
			WantI64: 0,
			WantEnd: 1,
		},
		{
			Name:    "only three",
			Data:    "3",
			WantI64: 3,
			WantEnd: 1,
		},
		{
			Name:    "only 42",
			Data:    "42",
			WantI64: 42,
			WantEnd: 2,
		},
		{
			Name:    "zero with crap following",
			Data:    "0  \n\t junk ",
			WantI64: 0,
			WantEnd: 1,
		},
		{
			Name:    "three with crap following",
			Data:    "3  gfguhbj ",
			WantI64: 3,
			WantEnd: 1,
		},
		{
			Name:    "42 with crap following",
			Data:    "42 junk \t ",
			WantI64: 42,
			WantEnd: 2,
		},
		{
			Name:    "long number with crap following",
			Data:    "876545678191878 junk \t ",
			WantI64: 876545678191878,
			WantEnd: 15,
		},

		// errors
		{
			Name:         "not only digits for 0, start with negation",
			Data:         "-0",
			WantErrError: needAtLeastOneDigit,
		},
		{
			Name:         "not only digits for 3, start with negation",
			Data:         "-3",
			WantErrError: needAtLeastOneDigit,
		},
		{
			Name:         "not only digits for 42, start with negation",
			Data:         "-42",
			WantErrError: needAtLeastOneDigit,
		},
		{
			Name:         "letters and digits",
			Data:         "h19",
			WantErrError: needAtLeastOneDigit,
		},
		{
			Name:         "letters only",
			Data:         "aaa",
			WantErrError: needAtLeastOneDigit,
		},
		{
			Name:         "no content",
			Data:         "",
			WantErrError: reachedEndScanningDigit,
		},
	}

	for _, tt := range tests {
		t.Logf("====> %s", tt.Name)

		gotI64, gotEnd, gotErr := scanDigits([]byte(tt.Data), tt.Start)

		// if we expect errors
		if tt.WantErrError != "" && gotErr == nil {
			t.Errorf("want an error, got none")
		} else if tt.WantErrError != "" && gotErr != nil {
			wantOffset := tt.Start
			err := gotErr.(*SyntaxError)
			if wantOffset != err.Offset {
				t.Errorf("want err offset %d, was %d", wantOffset, err.Offset)
			}
			if want, got := tt.WantErrError, gotErr.Error(); want != got {
				t.Errorf("want error: %q", want)
				t.Errorf(" got error: %q", got)
			}
		} else if gotErr != nil {
			t.Error(gotErr)
		}

		if want, got := tt.WantEnd, gotEnd; want != got {
			t.Errorf("want advance to %d, got %d", want, got)
		}
		if want, got := tt.WantI64, gotI64; want != got {
			t.Errorf("want val %d", want)
			t.Errorf(" got val %d", got)
		}
	}
}

func TestSkipWhitespace(t *testing.T) {
	tests := []struct {
		Name    string
		Start   int
		Data    string
		WantEnd int
	}{
		{
			Name:    "empty string",
			Start:   0,
			Data:    "",
			WantEnd: 0,
		},
		{
			Name:    "no whitespace",
			Start:   0,
			Data:    "hello",
			WantEnd: 0,
		},
		{
			Name:    "1 space",
			Start:   0,
			Data:    " ",
			WantEnd: 1,
		},
		{
			Name:    "1 tab",
			Start:   0,
			Data:    "\t",
			WantEnd: 1,
		},
		{
			Name:    "1 newline",
			Start:   0,
			Data:    "\n",
			WantEnd: 1,
		},
		{
			Name:    "1 carriage return",
			Start:   0,
			Data:    "\r",
			WantEnd: 1,
		},
		{
			Name:    "word with many types of whitespace",
			Start:   0,
			Data:    " \r \n \t hello",
			WantEnd: 7,
		},
		{
			Name:    "offset, word with many types of whitespace",
			Start:   12,
			Data:    " \r \n \t hello  \r \n \t hello",
			WantEnd: 20,
		},

		{
			Name:    "start out of range",
			Start:   10,
			Data:    " hello",
			WantEnd: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			want, got := tt.WantEnd, skipWhitespace([]byte(tt.Data), tt.Start)
			if want != got {
				t.Errorf("want advance to %d, got %d", want, got)
			}
		})
	}
}
