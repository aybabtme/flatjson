package flatjson

func scanPairName(data []byte, from int) (Pos, int, *SyntaxError) {
	// scan the name
	pos, err := scanString(data, from)
	if err != nil {
		return pos, 0, syntaxErr(from, expectingNameBeforeValue, err)
	}

	// scan the separator
	i, err := scanSeparator(data, pos.To)
	if err != nil {
		return pos, 0, err
	}
	return pos, i, nil
}

func scanSeparator(data []byte, from int) (int, *SyntaxError) {
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

// scanArray according to the spec at http://www.json.org/
// but ignoring nested objects and arrays
func scanArray(data []byte, from int, cb *Callbacks) (pos Pos, err error) {
	pos.From, pos.To = -1, -1
	start := skipWhitespace(data, from)
	if len(data) == 0 || data[start] != '[' {
		return pos, syntaxErr(start, noOpeningBracketFound, nil)
	}
	i := start + 1
	for ; i < len(data); i++ {

		i = skipWhitespace(data, i)
		if i >= len(data) {
			return pos, syntaxErr(i, endOfDataNoNamePair, nil)
		}

		if data[i] == ']' {
			return Pos{start, i + 1}, nil
		}

		// decide if the value is a number, string, object, array, bool or null
		b := data[i]

		if b == '"' { // strings
			valPos, err := scanString(data, i)
			if err != nil {
				return pos, syntaxErr(i, beginStringValueButError, err)
			}

			if cb != nil && cb.OnString != nil {
				cb.OnString(String{Name: pos, Value: valPos})
			}
			i = valPos.To

		} else if b == '{' { // objects
			valPos, err := ScanObject(data, i, nil) // TODO: fix recursion
			if err != nil {
				return Pos{}, syntaxErr(i, beginObjectValueButError, err.(*SyntaxError))
			}
			i = valPos.To

		} else if b == '[' { // arrays
			valPos, err := scanArray(data, i, nil) // TODO: fix recursion
			if err != nil {
				return Pos{}, syntaxErr(i, beginArrayValueButError, err.(*SyntaxError))
			}
			i = valPos.To

		} else if b == '-' || (b >= '0' && b <= '9') { // numbers
			val, j, err := scanNumber(data, i)
			if err != nil {
				return pos, syntaxErr(i, beginNumberValueButError, err)
			}
			j = skipWhitespace(data, j)
			if j < len(data) && data[j] != ',' && data[j] != ']' {
				return pos, syntaxErr(i, malformedNumber, nil)
			}
			if cb != nil && cb.OnNumber != nil {
				cb.OnNumber(Number{Name: pos, Value: val})
			}
			i = j

		} else if i+3 < len(data) && // bool - true case
			b == 't' &&
			data[i+1] == 'r' &&
			data[i+2] == 'u' &&
			data[i+3] == 'e' {

			if cb != nil && cb.OnBoolean != nil {
				cb.OnBoolean(Bool{Name: pos, Value: true})
			}
			i += 4

		} else if i+4 < len(data) && // bool - false case
			b == 'f' &&
			data[i+1] == 'a' &&
			data[i+2] == 'l' &&
			data[i+3] == 's' &&
			data[i+4] == 'e' {

			if cb != nil && cb.OnBoolean != nil {
				cb.OnBoolean(Bool{Name: pos, Value: false})
			}
			i += 5

		} else if i+3 < len(data) && // null
			b == 'n' &&
			data[i+1] == 'u' &&
			data[i+2] == 'l' &&
			data[i+3] == 'l' {

			if cb != nil && cb.OnNull != nil {
				cb.OnNull(Null{Name: pos})
			}
			i += 4

		} else {
			return pos, syntaxErr(i, expectValueButNoKnownType, nil)
		}

		i = skipWhitespace(data, i)
		if i < len(data) {
			if data[i] == ',' {
				// more values to come
				// TODO(antoine): be kind and accept trailing commas
			} else if data[i] == ']' {
				return Pos{start, i + 1}, nil
			}
		}
	}
	return pos, syntaxErr(i, endOfDataNoClosingBracket, nil)
}
