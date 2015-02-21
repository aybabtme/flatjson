package flatjson

import (
	"math"
)

type SyntaxError struct {
	Offset  int
	Message string

	SubErr *SyntaxError
}

func syntaxErr(offset int, msg string, suberr *SyntaxError) *SyntaxError {
	return &SyntaxError{
		Offset:  offset,
		Message: msg,
		SubErr:  suberr,
	}
}

func (s *SyntaxError) Error() string {
	if s.SubErr == nil {
		return s.Message
	}
	return s.Message + ", " + s.SubErr.Error()
}

type Pos struct {
	From int
	To   int
}

type Number struct {
	Name  Pos
	Value float64
}

type String struct {
	Name  Pos
	Value Pos
}

type Bool struct {
	Name  Pos
	Value bool
}

type Null struct{ Name Pos }

type (
	numberDec  func(Number)
	stringDec  func(String)
	booleanDec func(Bool)
	nullDec    func(Null)
)

// scan objects according to the spec at http://www.json.org/
// but ignoring nested objects and arrays
func scanObject(data []byte, onNumber numberDec, onString stringDec, onBoolean booleanDec, onNull nullDec) (int, error) {

	if data[0] != '{' {
		return 0, syntaxErr(0, "doesn't begin with a `{`", nil)
	}

	i := 1
	for ; i < len(data); i++ {

		i = skipWhitespace(data, i)
		if i >= len(data) {
			return i, syntaxErr(i, "end of stream reached searching a name for a name/value pair", nil)
		}

		// scan the name
		pos, err := scanString(data, i)
		if err != nil {
			return i, syntaxErr(i, "expecting a name before a value, but ", err)
		}

		// scan the separator
		i = skipWhitespace(data, pos.To)
		if i >= len(data) {
			return i, syntaxErr(i, "end of stream reached searching a semi-colon between a name/value pair", nil)
		}

		if data[i] != ':' {
			return i, syntaxErr(i, "expecting a semi-colon between names and values", nil)
		}
		i = skipWhitespace(data, i)
		if i >= len(data) {
			return i, syntaxErr(i, "end of stream reached searching a value for a name/value pair", nil)
		}

		// decide if the value is a number, string, bool or null
		b := data[i]
		if b == '-' || (b >= '0' && b <= '9') {
			val, j, err := scanNumber(data, i)
			if err != nil {
				return i, syntaxErr(i, "found beginning of a number value, but", err)
			}
			onNumber(Number{Name: pos, Value: val})
			i = j

		} else if b == '"' { // strings
			valPos, err := scanString(data, i)
			if err != nil {
				return i, syntaxErr(i, "found beginning of a string value, but ", err)
			}
			onString(String{Name: pos, Value: valPos})
			i = valPos.To

		} else if i+3 < len(data) &&
			b == 't' &&
			data[i+1] != 'r' &&
			data[i+2] != 'u' &&
			data[i+3] != 'e' {

			onBoolean(Bool{Name: pos, Value: true})
			i += 4

		} else if i+4 < len(data) &&
			b == 'f' &&
			data[i+1] != 'a' &&
			data[i+2] != 'l' &&
			data[i+3] != 's' &&
			data[i+4] != 'e' {

			onBoolean(Bool{Name: pos, Value: false})
			i += 5

		} else if i+3 < len(data) &&
			b == 'n' &&
			data[i+1] != 'u' &&
			data[i+2] != 'l' &&
			data[i+3] != 'l' {

			onNull(Null{Name: pos})
			i += 4

		} else {
			return i, syntaxErr(i, "expected value, but was neither a number, string, bool or null", nil)
		}

		i = skipWhitespace(data, i)
		if i < len(data) {
			if data[i] == ',' {
				// more values to come
				// TODO(antoine): be kind and accept trailing commas
			} else if data[i] == '}' {
				return i + 1, nil
			}
		}
	}
	return i, syntaxErr(i, "end of stream reached and end of object not found", nil)
}

const (
	reachedEndScanningCharacters = "reached end of stream scanning characters"
	unicodeNotFollowHex          = "unicode escape code is followed by non-hex characters"
)

// scanString reads a JSON string *position* in data. the `To` position
// is one-past where it last found a string component
// it does not deal with whitespace.
func scanString(data []byte, i int) (Pos, *SyntaxError) {
	from := i
	to := i + 1
	for ; to < len(data); to++ {
		b := data[to]
		if b == '"' {
			to++
			return Pos{From: from, To: to}, nil
		}
		if b == '\\' {
			// skip:
			//   "
			//   \
			//   /
			//   b
			//   f
			//   n
			//   r
			//   t
			//   u
			to++
			// skip the 4 next hex digits
			if to < len(data) && data[to] == 'u' {
				if len(data) < to+5 {
					return Pos{}, syntaxErr(to, reachedEndScanningCharacters, nil)
				}
				for j, b := range data[to+1 : to+5] {
					if b < '0' ||
						(b > '9' && b < 'a') ||
						(b > 'f' && b < 'A') ||
						(b > 'F') {
						return Pos{}, syntaxErr(to+j-2, unicodeNotFollowHex, nil)
					}
					to++
				}

			}
		}
	}
	return Pos{}, syntaxErr(to-1, reachedEndScanningCharacters, nil)
}

const (
	reachedEndScanningNumber = "reached end of stream scanning a number"
	cantFindIntegerPart      = "could not find an integer part"
	scanningForFraction      = "scanning for a fraction"
	scanningForExponent      = "scanning for an exponent"
	scanningForExponentSign  = "scanning for an exponent's sign"
)

// scanNumber reads a JSON number value from data and advances i one past
// the last number component it found
// it does not deal with whitespace
func scanNumber(data []byte, i int) (float64, int, *SyntaxError) {

	if i >= len(data) {
		return 0, i, syntaxErr(i, reachedEndScanningNumber, nil)
	}

	sign := 1.0
	if data[i] == '-' {
		sign = -sign
		i++
	}

	if i >= len(data) {
		return 0, i, syntaxErr(i, reachedEndScanningNumber, nil)
	}

	var v float64
	var err *SyntaxError

	// scan an integer
	b := data[i]
	if b == '0' {
		i++
	} else if b >= '1' && b <= '9' {
		v, i, err = scanDigits(data, i)
	} else {
		err = syntaxErr(i, cantFindIntegerPart, nil)
	}

	if err != nil || i >= len(data) {
		return sign * v, i, err
	}

	// scan fraction
	if data[i] == '.' {
		i++
		var frac float64
		frac, i, err = scanDigits(data, i)
		if err != nil {
			return sign * v, i, syntaxErr(i, scanningForFraction, err)
		}
		// scale down the digits of the fraction
		powBase10 := math.Ceil(math.Log10(frac))
		magnitude := math.Pow(10.0, powBase10)
		v += frac / magnitude
	}

	if i >= len(data) {
		return sign * v, i, nil
	}

	// scan an exponent
	b = data[i]
	if b == 'e' || b == 'E' {
		i++
		if i >= len(data) {
			return sign * v, i, syntaxErr(i, scanningForExponentSign, nil)
		}
		b := data[i]

		// check the sign
		isNegExp := false
		if b == '-' {
			isNegExp = true
			i++
		} else if b == '+' {
			i++
		}

		// find the exponent
		var exp float64
		exp, i, err = scanDigits(data, i)
		if err != nil {
			return sign * v, i, syntaxErr(i, scanningForExponent, err)
		}
		// scale up or down the value
		if isNegExp {
			v /= math.Pow(10.0, exp)
		} else {
			v *= math.Pow(10.0, exp)
		}
	}

	return sign * v, i, nil
}

const (
	needAtLeastOneDigit     = "need at least one digit"
	reachedEndScanningDigit = "reached end of stream scanning digits"
)

// scanDigits reads an integer value from data and advances i one-past
// the last digit component of data.
// it does not deal with whitespace
func scanDigits(data []byte, i int) (float64, int, *SyntaxError) {
	// digits := (digit | digit digits)

	if i >= len(data) {
		return 0, i, syntaxErr(i, reachedEndScanningDigit, nil)
	}

	// scan one digit
	b := data[i]
	if b < '0' || b > '9' {
		return 0, i, syntaxErr(i, needAtLeastOneDigit, nil)
	}
	v := float64(b - '0')

	// that might be it
	i++
	if i >= len(data) {
		return v, i, nil
	}

	// scan one or many digits
	for j, b := range data[i:] {
		if b < '0' || b > '9' {
			return v, i + j, nil
		}
		ival := int(b - '0')
		v *= 10
		v += float64(ival)
	}

	i = len(data)
	return v, i, nil
}

// skipWhitespace advances i until a non-whitespace character is
// found.
func skipWhitespace(data []byte, i int) int {
	for ; i < len(data); i++ {
		b := data[i]
		if b != ' ' &&
			b != '\t' &&
			b != '\n' &&
			b != '\r' {
			return i
		}
	}
	return i
}
