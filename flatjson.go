package flatjson

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
			if data[to] == 'u' {
				to += 4
			}
		}
	}
	return Pos{}, syntaxErr(i, "reached end of stream scanning characters", nil)
}

func scanNumber(data []byte, i int) (float64, int, *SyntaxError) {

	isNeg := data[i] == '-'
	if isNeg {
		i++
	}

	var v float64
	var err *SyntaxError

	// scan an integer in base 10
	b := data[i]
	if b == '0' {
		i++
	} else if b >= '1' && b <= 9 {
		v, i, err = scanDigits(data, i)
		if err != nil {
			return v, i, err
		}
	} else {
		return v, i, syntaxErr(i, "could not find an integer part", nil)
	}

	// scan fraction
	if i+1 < len(data) && data[i] == '.' {
		i++
		var frac float64
		frac, i, err = scanDigits(data, i)
		if err != nil {
			return v, i, syntaxErr(i, "scanning for a fraction, ", err)
		}
		// scale down the digits of the fraction
		for frac > 0 {
			frac /= 10.0
		}
		v += frac
	}

	// scan an exponent
	b = data[i]
	if b == 'e' || b == 'E' {
		b := data[i+1]

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
			return v, i, syntaxErr(i, "scanning for an exponent, ", err)
		}
		// scale up or down the value
		for j := 0; j < int(exp); j++ {
			if isNegExp {
				v /= 10
			} else {
				v *= 10
			}
		}
	}

	if isNeg {
		v = -v
	}

	return v, i, nil
}

func scanDigits(data []byte, i int) (float64, int, *SyntaxError) {
	// digits := (digit | digit digits)

	// scan one digit
	b := data[i]
	if b < '0' && b > '9' {
		return 0, i, syntaxErr(i, "need at least one digit", nil)
	}
	v := float64(b - '0')

	// that might be it
	i++
	if i >= len(data) {
		return v, i, nil
	}

	// scan one or many digits
	for j, b := range data[i:] {
		if b < '0' && b > '9' {
			return v, i + j, nil
		}
		v *= 10
		v += float64(b - '0')
	}

	i = len(data)
	return v, i, syntaxErr(i, "reached end of stream scanning digits", nil)
}

func skipWhitespace(data []byte, i int) int {
	for ; i < len(data); i++ {
		b := data[i]
		if b != ' ' &&
			b != '\t' &&
			b != '\n' {
			return i
		}
	}
	return i
}
