package flatjson

func scanPairName(data []byte, from int) (Prefix, int, error) {
	// scan the name
	pos, err := scanString(data, from)
	pfx := newObjectKeyPrefix(pos.From, pos.To)
	if err != nil {
		return pfx, 0, syntaxErr(from, expectingNameBeforeValue, err.(*SyntaxError))
	}

	// scan the separator
	i, err := scanSeparator(data, pos.To)
	if err != nil {
		return pfx, 0, err
	}
	return pfx, i, nil
}

func scanSeparator(data []byte, from int) (int, error) {
	i := skipWhitespace(data, from)
	if i >= len(data) {
		return i, syntaxErr(i, endOfDataNoColon, nil)
	}

	if data[i] != ':' {
		return i, syntaxErr(i, noColonFound, nil)
	}
	i++
	i = skipWhitespace(data, i)
	if i >= len(data) {
		return i, syntaxErr(i, endOfDataNoValueForName, nil)
	}
	return i, nil
}

// wip

// ScanArray according to the spec at http://www.json.org/
// but ignoring nested objects and arrays
func ScanArray(data []byte, from int, cb *Callbacks) (pos Pos, found bool, err error) {
	return scanArray(data, from, nil, cb)
}

func scanArray(data []byte, from int, prefixes []Prefix, cb *Callbacks) (pos Pos, found bool, _ error) {
	pos.From, pos.To = -1, -1
	start := skipWhitespace(data, from)
	if len(data) == 0 || data[start] != '[' {
		return pos, false, syntaxErr(start, noOpeningSquareBracketFound, nil)
	}
	i := start + 1
	for index := -1; i < len(data); i++ {
		index++

		i = skipWhitespace(data, i)
		if i >= len(data) {
			return pos, false, syntaxErr(i, endOfDataNoNamePair, nil)
		}

		if data[i] == ']' {
			return Pos{start, i + 1}, true, nil
		}

		// decide if the value is a number, string, object, array, bool or null
		b := data[i]

		var (
			valPos Pos
			err    error
		)
		if b == '"' { // strings
			valPos, err = scanString(data, i)
			if err != nil {
				return pos, false, syntaxErr(i, beginStringValueButError, err.(*SyntaxError))
			}

			if cb != nil && cb.OnString != nil && cb.MaxDepth >= len(prefixes) {
				cb.OnString(prefixes, String{Name: newArrayIndexPrefix(index), Value: valPos})
			}
			i = valPos.To

		} else if b == '{' { // objects
			valPos, found, err = scanObject(data, i, append(prefixes, newArrayIndexPrefix(index)), cb) // TODO: fix recursion
			if err != nil {
				return Pos{}, found, syntaxErr(i, beginObjectValueButError, err.(*SyntaxError))
			} else if !found {
				return Pos{}, found, syntaxErr(i, expectValueButNoKnownType, nil)
			}
			i = valPos.To

		} else if b == '[' { // arrays
			valPos, found, err = scanArray(data, i, append(prefixes, newArrayIndexPrefix(index)), cb) // TODO: fix recursion
			if err != nil {
				return Pos{}, found, syntaxErr(i, beginArrayValueButError, err.(*SyntaxError))
			} else if !found {
				return Pos{}, found, syntaxErr(i, expectValueButNoKnownType, nil)
			}
			i = valPos.To

		} else if b == '-' || (b >= '0' && b <= '9') { // numbers
			val, j, err := ScanNumber(data, i)
			if err != nil {
				return pos, false, syntaxErr(i, beginNumberValueButError, err.(*SyntaxError))
			}
			j = skipWhitespace(data, j)
			if j < len(data) && data[j] != ',' && data[j] != ']' {
				return pos, false, syntaxErr(i, malformedNumber, nil)
			}
			if cb != nil && cb.OnNumber != nil && cb.MaxDepth >= len(prefixes) {
				cb.OnNumber(prefixes, Number{Name: newArrayIndexPrefix(index), Value: val})
			}
			valPos = Pos{From: i, To: j}
			i = j

		} else if i+3 < len(data) && // bool - true case
			b == 't' &&
			data[i+1] == 'r' &&
			data[i+2] == 'u' &&
			data[i+3] == 'e' {

			if cb != nil && cb.OnBoolean != nil && cb.MaxDepth >= len(prefixes) {
				cb.OnBoolean(prefixes, Bool{Name: newArrayIndexPrefix(index), Value: true})
			}
			valPos = Pos{From: i, To: i + 4}
			i += 4

		} else if i+4 < len(data) && // bool - false case
			b == 'f' &&
			data[i+1] == 'a' &&
			data[i+2] == 'l' &&
			data[i+3] == 's' &&
			data[i+4] == 'e' {

			if cb != nil && cb.OnBoolean != nil && cb.MaxDepth >= len(prefixes) {
				cb.OnBoolean(prefixes, Bool{Name: newArrayIndexPrefix(index), Value: false})
			}
			valPos = Pos{From: i, To: i + 5}
			i += 5

		} else if i+3 < len(data) && // null
			b == 'n' &&
			data[i+1] == 'u' &&
			data[i+2] == 'l' &&
			data[i+3] == 'l' {

			if cb != nil && cb.OnNull != nil && cb.MaxDepth >= len(prefixes) {
				cb.OnNull(prefixes, Null{Name: newArrayIndexPrefix(index)})
			}
			valPos = Pos{From: i, To: i + 4}
			i += 4

		} else {
			return pos, false, syntaxErr(i, expectValueButNoKnownType, nil)
		}
		if cb != nil && cb.OnRaw != nil && cb.MaxDepth >= len(prefixes) {
			cb.OnRaw(prefixes, newArrayIndexPrefix(index), valPos)
		}

		i = skipWhitespace(data, i)
		if i < len(data) {
			if data[i] == ',' {
				// more values to come
				// TODO(antoine): be kind and accept trailing commas
			} else if data[i] == ']' {
				return Pos{start, i + 1}, true, nil
			}
		}
	}
	return pos, false, syntaxErr(i, endOfDataNoClosingSquareBracket, nil)
}
