package micromark

// characterReference is micromark-core-commonmark/lib/character-reference.js.
var characterReference = &Construct{Name: "characterReference", Tokenize: tokenizeCharacterReference}

func tokenizeCharacterReference(self *Self, effects *Effects, ok State, nok State) State {
	size := 0
	var max int
	var test func(code Code) bool
	// Upstream compares the function itself, `test === asciiAlphanumeric`; Go cannot compare funcs, so
	// open records the named kind alongside test.
	named := false

	var start, open, numeric, value State

	// Start of character reference.
	//
	//	> | a&amp;b
	//	     ^
	//	> | a&#123;b
	//	     ^
	//	> | a&#x9;b
	//	     ^
	start = func(code Code) State {
		effects.Enter(TypeCharacterReference, nil)
		effects.Enter(TypeCharacterReferenceMarker, nil)
		effects.Consume(code)
		effects.Exit(TypeCharacterReferenceMarker)
		return open
	}

	// After `&`, at `#` for numeric references or alphanumeric for named
	// references.
	//
	//	> | a&amp;b
	//	      ^
	//	> | a&#123;b
	//	      ^
	//	> | a&#x9;b
	//	      ^
	open = func(code Code) State {
		if code == CodeNumberSign {
			effects.Enter(TypeCharacterReferenceMarkerNumeric, nil)
			effects.Consume(code)
			effects.Exit(TypeCharacterReferenceMarkerNumeric)
			return numeric
		}

		effects.Enter(TypeCharacterReferenceValue, nil)
		max = characterReferenceNamedSizeMax
		test = asciiAlphanumeric
		named = true
		return value(code)
	}

	// After `#`, at `x` for hexadecimals or digit for decimals.
	//
	//	> | a&#123;b
	//	       ^
	//	> | a&#x9;b
	//	       ^
	numeric = func(code Code) State {
		if code == CodeUppercaseX || code == CodeLowercaseX {
			effects.Enter(TypeCharacterReferenceMarkerHexadecimal, nil)
			effects.Consume(code)
			effects.Exit(TypeCharacterReferenceMarkerHexadecimal)
			effects.Enter(TypeCharacterReferenceValue, nil)
			max = characterReferenceHexadecimalSizeMax
			test = asciiHexDigit
			return value
		}

		effects.Enter(TypeCharacterReferenceValue, nil)
		max = characterReferenceDecimalSizeMax
		test = asciiDigit
		return value(code)
	}

	// After markers (`&#x`, `&#`, or `&`), in value, before `;`.
	//
	// The character reference kind defines what and how many characters are
	// allowed.
	//
	//	> | a&amp;b
	//	      ^^^
	//	> | a&#123;b
	//	       ^^^
	//	> | a&#x9;b
	//	        ^
	value = func(code Code) State {
		if code == CodeSemicolon && size != 0 {
			token := effects.Exit(TypeCharacterReferenceValue)

			if named {
				// `!decodeNamedCharacterReference(...)`: false (unknown) and an empty decoding are both
				// falsy.
				decoded, present := DecodeNamedCharacterReference(self.SliceSerialize(token, false))
				if !present || decoded == "" {
					return nok(code)
				}
			}

			// To do: `markdown-rs` uses a different name:
			// `CharacterReferenceMarkerSemi`.
			effects.Enter(TypeCharacterReferenceMarker, nil)
			effects.Consume(code)
			effects.Exit(TypeCharacterReferenceMarker)
			effects.Exit(TypeCharacterReference)
			return ok
		}

		if test(code) {
			// `size++ < max`: only reached when test passes (short-circuit), and the increment happens
			// whether or not the comparison passes.
			below := size < max
			size++
			if below {
				effects.Consume(code)
				return value
			}
		}

		return nok(code)
	}

	return start
}
