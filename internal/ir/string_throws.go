package ir

// normalize given a form that isn't one of the four, and String.fromCodePoint given a code that isn't an
// integer from 0 to 0x10FFFF, throw Node's RangeError, which a catch takes (runtime/normalize.c and
// from_codes.c set it on the pending word), as the number formats do (number_format.go). Lowering counts
// the calls that can, the flow graph gives them an exception edge, and the emitter tests the word after
// each one.

// MayThrow reports whether a string method can throw a catchable error: normalize, unless its form is a
// constant naming one of the four. strings are the program's constants, one of which names the form.
// repeat's failure is still a panic natively (lower's libraryFailure names it), so it isn't one here.
func (call StringCall) MayThrow(strings []string) bool {
	return call.Method == "normalize" && len(call.Arguments) > 0 && !IsNormalizationForm(call.Arguments[0], strings)
}

// MayThrow reports whether String.fromCodePoint can be given a code it rejects: unless every code is
// written out as a constant integer from 0 to 0x10FFFF. A spread's codes are known only as it runs.
// fromCharCode takes any number, so it never throws.
func (call StringFromCodes) MayThrow() bool {
	if !call.CodePoints {
		return false
	}
	if call.Spread != nil {
		return true
	}
	for _, code := range call.Codes {
		if !ConstantWithin(code, 0, 0x10ffff) {
			return true
		}
	}
	return false
}

// IsNormalizationForm reports whether a value is a string constant naming one of normalize's four forms.
func IsNormalizationForm(value Expression, strings []string) bool {
	constant, isConstant := value.(StringConstant)
	if !isConstant {
		return false
	}
	switch strings[constant.Index] {
	case "NFC", "NFD", "NFKC", "NFKD":
		return true
	}
	return false
}
