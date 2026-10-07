package regexp

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"
)

var adamicTrace strings.Builder

func adamicIdentity(r rune, size int, ctx escapeContext) (decodedEscape, error) {
	fmt.Fprintf(&adamicTrace, "identity:%d:%d:%t:%t:%d:%t;", r, size, ctx.unicode, ctx.inClass, ctx.groups, ctx.named)
	return identityEscape(r, size, ctx)
}
func adamicIndexByte(s string, b byte) int {
	fmt.Fprintf(&adamicTrace, "brace:%d;", len(s))
	return strings.IndexByte(s, b)
}
func adamicDigit(s string, i int) bool {
	fmt.Fprintf(&adamicTrace, "digit:%d;", i)
	return isDecimalDigit(s, i)
}
func adamicDecimal(s string, i int) (int, int) {
	fmt.Fprintf(&adamicTrace, "decimal:%d;", i)
	return decimalEscape(s, i)
}
func adamicOctal(s string, i int) (rune, int) {
	fmt.Fprintf(&adamicTrace, "octal:%d;", i)
	return decodeLegacyOctal(s, i)
}
func adamicDecodeRune(s string) (rune, int) {
	fmt.Fprintf(&adamicTrace, "decode:%d;", len(s))
	return utf8.DecodeRuneInString(s)
}

type AdamicSource struct {
	Bytes []int `json:"bytes"`
	Sizes []int `json:"sizes"`
}

func AdamicSourceOf(s string) AdamicSource {
	v := AdamicSource{Bytes: []int{}, Sizes: []int{}}
	for i := 0; i < len(s); i++ {
		v.Bytes = append(v.Bytes, int(s[i]))
		_, size := utf8.DecodeRuneInString(s[i:])
		v.Sizes = append(v.Sizes, size)
	}
	v.Sizes = append(v.Sizes, 0)
	return v
}
func AdamicObserve(mode, source string, index int, u, c bool, groups int, named bool) ([]int, string) {
	adamicTrace.Reset()
	ctx := escapeContext{unicode: u, inClass: c, groups: groups, named: named}
	deps := []int{0, 0, 0, 0, 0, -1}
	if mode == "property" {
		deps[5] = strings.IndexByte(source[index+1:], '}')
	}
	if mode == "class" {
		body, neg, width, e := readClass(source)
		message := ""
		if e != nil {
			message = e.Error()
		}
		return deps, fmt.Sprintf("%d|%t|%s|%s|%s", width, neg, hex.EncodeToString([]byte(body)), message, adamicTrace.String())
	}
	var value decodedEscape
	var e error
	if mode == "numeric" {
		d, dw := decimalEscape(source, index)
		o, ow := decodeLegacyOctal(source, index)
		next := 0
		if isDecimalDigit(source, index+1) {
			next = 1
		}
		deps = []int{d, dw, int(o), ow, next, -1}
		value, e = decodeNumericEscape(source, index, ctx)
	} else {
		value, e = decodePropertyEscape(source, index, 1, ctx)
	}
	identity, ie := identityEscape(rune(source[index]), 1, ctx)
	errorText := ""
	if ie != nil {
		errorText = ie.Error()
	}
	if mode == "property" {
		deps = append(deps, int(identity.kind), int(identity.set), int(identity.r), identity.width)
		if identity.negated {
			deps = append(deps, 1)
		} else {
			deps = append(deps, 0)
		}
	}
	_ = errorText
	message := ""
	if e != nil {
		message = e.Error()
	}
	return deps, fmt.Sprintf("%d,%d,%d,%d,%t|%s|%s", value.kind, value.set, value.r, value.width, value.negated, message, adamicTrace.String())
}
func AdamicIdentityError(source string, index int, u, c bool, groups int, named bool) string {
	_, e := identityEscape(rune(source[index]), 1, escapeContext{unicode: u, inClass: c, groups: groups, named: named})
	if e != nil {
		return e.Error()
	}
	return ""
}
