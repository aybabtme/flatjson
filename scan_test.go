package flatjson

import (
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
			_, _, err := ScanNumber([]byte(tt.Data), tt.Start)

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

		WantVal float64
		WantEnd int
	}{
		// no exponent
		{
			Name:    "integer, 0",
			Data:    "0",
			WantVal: 0,
			WantEnd: 1,
		},
		{
			Name:    "integer, 3",
			Data:    "3",
			WantVal: 3,
			WantEnd: 1,
		},
		{
			Name:    "integer, 42",
			Data:    "42",
			WantVal: 42,
			WantEnd: 2,
		},
		{
			Name:    "integer, 9000",
			Data:    "9000",
			WantVal: 9000,
			WantEnd: 4,
		},
		{
			Name:    "negative integer, -0",
			Data:    "-0",
			WantVal: -0,
			WantEnd: 2,
		},
		{
			Name:    "negative integer, -3",
			Data:    "-3",
			WantVal: -3,
			WantEnd: 2,
		},
		{
			Name:    "negative integer, -42",
			Data:    "-42",
			WantVal: -42,
			WantEnd: 3,
		},
		{
			Name:    "negative integer, -9000",
			Data:    "-9000",
			WantVal: -9000,
			WantEnd: 5,
		},
		{
			Name:    "real numbers, around 0",
			Data:    "0.14159",
			WantVal: 0.14159,
			WantEnd: 7,
		},
		{
			Name:    "real numbers, around 3",
			Data:    "3.14159",
			WantVal: 3.14159,
			WantEnd: 7,
		},
		{
			Name:    "real numbers, around 42",
			Data:    "42.14159",
			WantVal: 42.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around 9000",
			Data:    "9000.14159",
			WantVal: 9000.14159,
			WantEnd: 10,
		},
		{
			Name:    "real numbers, around -0",
			Data:    "-0.14159",
			WantVal: -0.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around -3",
			Data:    "-3.14159",
			WantVal: -3.14159,
			WantEnd: 8,
		},
		{
			Name:    "real numbers, around -42",
			Data:    "-42.14159",
			WantVal: -42.14159,
			WantEnd: 9,
		},
		{
			Name:    "real numbers, around -9000",
			Data:    "-9000.14159",
			WantVal: -9000.14159,
			WantEnd: 11,
		},

		// with a positive exponent
		{
			Name:    "positive exponent, integer, 0",
			Data:    "0e42",
			WantVal: 0e42,
			WantEnd: 4,
		},
		{
			Name:    "positive exponent, integer, 3",
			Data:    "3e42",
			WantVal: 3e42,
			WantEnd: 4,
		},
		{
			Name:    "positive exponent, integer, 42",
			Data:    "42e42",
			WantVal: 42e42,
			WantEnd: 5,
		},
		{
			Name:    "positive exponent, integer, 9000",
			Data:    "9000e42",
			WantVal: 9000e42,
			WantEnd: 7,
		},
		{
			Name:    "positive exponent, negative integer, -0",
			Data:    "-0e42",
			WantVal: -0e42,
			WantEnd: 5,
		},
		{
			Name:    "positive exponent, negative integer, -3",
			Data:    "-3e42",
			WantVal: -3e42,
			WantEnd: 5,
		},
		{
			Name:    "positive exponent, negative integer, -42",
			Data:    "-42e42",
			WantVal: -42e42,
			WantEnd: 6,
		},
		{
			Name:    "positive exponent, negative integer, -9000",
			Data:    "-9000e42",
			WantVal: -9000e42,
			WantEnd: 8,
		},
		{
			Name:    "positive exponent, real numbers, around 0",
			Data:    "0.14159e42",
			WantVal: 0.14159e42,
			WantEnd: 10,
		},
		{
			Name:    "positive exponent, real numbers, around 3",
			Data:    "3.14159e42",
			WantVal: 3.14159e42,
			WantEnd: 10,
		},
		{
			Name:    "positive exponent, real numbers, around 42",
			Data:    "42.14159e42",
			WantVal: 42.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around 9000",
			Data:    "9000.14159e42",
			WantVal: 9000.14159e42,
			WantEnd: 13,
		},
		{
			Name:    "positive exponent, real numbers, around -0",
			Data:    "-0.14159e42",
			WantVal: -0.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around -3",
			Data:    "-3.14159e42",
			WantVal: -3.14159e42,
			WantEnd: 11,
		},
		{
			Name:    "positive exponent, real numbers, around -42",
			Data:    "-42.14159e42",
			WantVal: -42.14159e42,
			WantEnd: 12,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159e42",
			WantVal: -9000.14159e42,
			WantEnd: 14,
		},

		// positive exponent variations
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159E42",
			WantVal: -9000.14159e42,
			WantEnd: 14,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159e+42",
			WantVal: -9000.14159e42,
			WantEnd: 15,
		},
		{
			Name:    "positive exponent, real numbers, around -9000",
			Data:    "-9000.14159E+42",
			WantVal: -9000.14159e42,
			WantEnd: 15,
		},

		// with a negative exponent
		{
			Name:    "negative exponent, integer, 0",
			Data:    "0e-42",
			WantVal: 0e-42,
			WantEnd: 5,
		},
		{
			Name:    "negative exponent, integer, 3",
			Data:    "3e-42",
			WantVal: 3e-42,
			WantEnd: 5,
		},
		{
			Name:    "negative exponent, integer, 42",
			Data:    "42e-42",
			WantVal: 42e-42,
			WantEnd: 6,
		},
		{
			Name:    "negative exponent, integer, 9000",
			Data:    "9000e-42",
			WantVal: 9000e-42,
			WantEnd: 8,
		},
		{
			Name:    "negative exponent, negative integer, -0",
			Data:    "-0e-42",
			WantVal: -0e-42,
			WantEnd: 6,
		},
		{
			Name:    "negative exponent, negative integer, -3",
			Data:    "-3e-42",
			WantVal: -3e-42,
			WantEnd: 6,
		},
		{
			Name:    "negative exponent, negative integer, -42",
			Data:    "-42e-42",
			WantVal: -42e-42,
			WantEnd: 7,
		},
		{
			Name:    "negative exponent, negative integer, -9000",
			Data:    "-9000e-42",
			WantVal: -9000e-42,
			WantEnd: 9,
		},
		{
			Name:    "negative exponent, real numbers, around 0",
			Data:    "0.14159e-42",
			WantVal: 0.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "negative exponent, real numbers, around 3",
			Data:    "3.14159e-42",
			WantVal: 3.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "negative exponent, real numbers, around 42",
			Data:    "42.14159e-42",
			WantVal: 42.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around 9000",
			Data:    "9000.14159e-42",
			WantVal: 9000.14159e-42,
			WantEnd: 14,
		},
		{
			Name:    "negative exponent, real numbers, around -0",
			Data:    "-0.14159e-42",
			WantVal: -0.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around -3",
			Data:    "-3.14159e-42",
			WantVal: -3.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "negative exponent, real numbers, around -42",
			Data:    "-42.14159e-42",
			WantVal: -42.14159e-42,
			WantEnd: 13,
		},
		{
			Name:    "negative exponent, real numbers, around -9000",
			Data:    "-9000.14159e-42",
			WantVal: -9000.14159e-42,
			WantEnd: 15,
		},
		{
			Name:    "negative exponent with variation, real numbers, around -9000",
			Data:    "-9000.14159E-42",
			WantVal: -9000.14159e-42,
			WantEnd: 15,
		},

		// with a garbage and negative exponent
		{
			Name:    "garbage and negative exponent, integer, 0",
			Data:    "0e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 0e-42,
			WantEnd: 5,
		},
		{
			Name:    "garbage and negative exponent, integer, 3",
			Data:    "3e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 3e-42,
			WantEnd: 5,
		},
		{
			Name:    "garbage and negative exponent, integer, 42",
			Data:    "42e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 42e-42,
			WantEnd: 6,
		},
		{
			Name:    "garbage and negative exponent, integer, 9000",
			Data:    "9000e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 9000e-42,
			WantEnd: 8,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -0",
			Data:    "-0e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -0e-42,
			WantEnd: 6,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -3",
			Data:    "-3e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -3e-42,
			WantEnd: 6,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -42",
			Data:    "-42e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -42e-42,
			WantEnd: 7,
		},
		{
			Name:    "garbage and negative exponent, negative integer, -9000",
			Data:    "-9000e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -9000e-42,
			WantEnd: 9,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 0",
			Data:    "0.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 0.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 3",
			Data:    "3.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 3.14159e-42,
			WantEnd: 11,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 42",
			Data:    "42.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 42.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around 9000",
			Data:    "9000.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: 9000.14159e-42,
			WantEnd: 14,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -0",
			Data:    "-0.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -0.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -3",
			Data:    "-3.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -3.14159e-42,
			WantEnd: 12,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -42",
			Data:    "-42.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -42.14159e-42,
			WantEnd: 13,
		},
		{
			Name:    "garbage and negative exponent, real numbers, around -9000",
			Data:    "-9000.14159e-42 yguhbhg2  23h23 2j3h ",
			WantVal: -9000.14159e-42,
			WantEnd: 15,
		},
		{
			Name:    "garbage and negative exponent with variation, real numbers, around -9000",
			Data:    "-9000.14159E-42 yguhbhg2  23h23 2j3h ",
			WantVal: -9000.14159e-42,
			WantEnd: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			gotVal, gotEnd, gotErr := ScanNumber([]byte(tt.Data), tt.Start)

			if gotErr != nil {
				t.Fatal(gotErr)
			}

			if want, got := tt.WantEnd, gotEnd; want != got {
				t.Errorf("want advance to %d, got %d", want, got)
			}
			if want, got := tt.WantVal, gotVal; !fequal(want, got) {
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

		WantVal float64
		WantEnd int

		WantErrError string
	}{
		// all good
		{
			Name:    "only zero",
			Data:    "0",
			WantVal: 0,
			WantEnd: 1,
		},
		{
			Name:    "only three",
			Data:    "3",
			WantVal: 3,
			WantEnd: 1,
		},
		{
			Name:    "only 42",
			Data:    "42",
			WantVal: 42,
			WantEnd: 2,
		},
		{
			Name:    "zero with crap following",
			Data:    "0  \n\t junk ",
			WantVal: 0,
			WantEnd: 1,
		},
		{
			Name:    "three with crap following",
			Data:    "3  gfguhbj ",
			WantVal: 3,
			WantEnd: 1,
		},
		{
			Name:    "42 with crap following",
			Data:    "42 junk \t ",
			WantVal: 42,
			WantEnd: 2,
		},
		{
			Name:    "long number with crap following",
			Data:    "876545678191878 junk \t ",
			WantVal: 876545678191878,
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

		gotVal, gotEnd, gotErr := scanDigits([]byte(tt.Data), tt.Start)

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
		if want, got := tt.WantVal, gotVal; want != got {
			t.Errorf("want val %f", want)
			t.Errorf(" got val %f", got)
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
