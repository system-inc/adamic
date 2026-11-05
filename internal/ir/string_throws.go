package ir

// A string longer than V8's longest (ADAMIC_STRING_MAX_UNITS, runtime/adamic.h) throws Node's
// RangeError "Invalid string length", which a catch takes; so does normalize given a form that isn't
// one of the four, and String.fromCodePoint given a code that isn't an integer from 0 to 0x10FFFF
// (runtime/string_build_impl.h, normalize.c and from_codes.c set it on the pending word), as the number
// formats do (number_format.go). Lowering counts the operations that can, the flow graph gives them an
// exception edge, and the emitter tests the word after each one.
//
// Every operation that can make a string longer than what it's given can reach the limit, whatever its
// operands, so each one checks: Adamic's ruling on #c3bk0fp, design (a), after Runtime Performance
// measured that sparing the ones no handler can see would spare about 4% of them on tsc. One that can't
// be longer than what it's given (slice, trim, at, a number or a boolean written out) never checks.

// StringLengthMayThrow reports whether node makes a string that can be longer than V8's longest: a
// concatenation of two or more parts (+, a template, concat, text += more), a string method that can
// lengthen its string (StringCall.MayThrow), a join, a replace through a RegExp, and JSON written out.
func StringLengthMayThrow(node any) bool {
	switch node := node.(type) {
	case Concat:
		// One part is a copy, as long as the string it's given.
		return len(node.Parts) > 1
	case StringCall:
		return node.MayThrow()
	case ArrayJoin, JSONEncode, JSONStringify:
		return true
	case RegExpCall:
		return node.Method == "replace" || node.Method == "replaceAll"
	}
	return false
}

// MayThrow reports whether a string method can throw a catchable error: each one that can make its
// string longer, past V8's longest. normalize's form that isn't one of the four throws too, but
// normalize is already here for the length, since a decomposition can lengthen any string.
func (call StringCall) MayThrow() bool {
	switch call.Method {
	case "padStart", "padEnd", "repeat", "toUpperCase", "toLowerCase", "normalize", "replace", "replaceAll":
		return true
	}
	return false
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
