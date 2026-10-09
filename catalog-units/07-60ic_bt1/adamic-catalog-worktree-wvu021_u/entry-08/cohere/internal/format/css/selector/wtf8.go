package selector

// # Strings that are not well-formed until the end
//
// Upstream's strings are UTF-16 and may hold a lone surrogate: the tokenizer's backslash case takes one
// unit after the backslash, so `\` before an emoji makes a token of the backslash and the high
// surrogate, and the low surrogate starts the next word. The parser glues such pieces back together
// (splitWord merges the words, attribute() and the parentheses concatenate token values), so the pair
// is whole again in the value a node ends up with.
//
// A Go string decoded per token would turn each half into U+FFFD and never rejoin them. So inside the
// package a string is WTF-8: UTF-8, except that a surrogate code unit is written as its own three-byte
// sequence. Concatenation and byte slicing at ASCII positions keep every unit, and wellFormed, applied
// to every string on the finished tree, joins pairs into proper UTF-8. A surrogate still alone there
// becomes U+FFFD, the one place the port differs from upstream's strings (where a JSON dump would show
// a lone \udXXX escape).

import (
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
)

// unitsToWTF8 writes UTF-16 units as WTF-8: pairs as their character, lone surrogates as themselves.
func unitsToWTF8(units []uint16) string {
	buffer := make([]byte, 0, len(units))
	for index := 0; index < len(units); index++ {
		unit := rune(units[index])
		if utf16.IsSurrogate(unit) {
			if unit < 0xDC00 && index+1 < len(units) && units[index+1] >= 0xDC00 && units[index+1] <= 0xDFFF {
				buffer = utf8.AppendRune(buffer, utf16.DecodeRune(unit, rune(units[index+1])))
				index++
				continue
			}
			buffer = append(buffer, byte(0xE0|unit>>12), byte(0x80|(unit>>6)&0x3F), byte(0x80|unit&0x3F))
			continue
		}
		buffer = utf8.AppendRune(buffer, unit)
	}
	return string(buffer)
}

// wtf8ToUnits is unitsToWTF8's inverse. A byte that starts no sequence is one U+FFFD unit, as []rune
// would decode it.
func wtf8ToUnits(text string) []uint16 {
	units := make([]uint16, 0, len(text))
	for position := 0; position < len(text); {
		if position+2 < len(text) && text[position] == 0xED && text[position+1] >= 0xA0 && text[position+1] <= 0xBF &&
			text[position+2] >= 0x80 && text[position+2] <= 0xBF {
			units = append(units, uint16(0xD000|rune(text[position+1]&0x3F)<<6|rune(text[position+2]&0x3F)))
			position += 3
			continue
		}
		character, size := utf8.DecodeRuneInString(text[position:])
		units = utf16.AppendRune(units, character)
		position += size
	}
	return units
}

// wellFormed turns a WTF-8 string into UTF-8, joining surrogate pairs and replacing lone surrogates
// with U+FFFD: String.prototype.toWellFormed.
func wellFormed(text string) string {
	// Proper UTF-8 holds no surrogate, and Parse's input went through []rune, so valid text needs nothing.
	if utf8.ValidString(text) {
		return text
	}
	return string(utf16.Decode(wtf8ToUnits(text)))
}

// wellFormedTree applies wellFormed to every string on the tree, inside source, spaces and raws too.
func wellFormedTree(node *estree.Node) {
	for _, key := range node.Keys() {
		node.Set(key, wellFormedValue(node.Get(key)))
	}
}

func wellFormedValue(value any) any {
	switch typed := value.(type) {
	case string:
		return wellFormed(typed)
	case map[string]any:
		for key, each := range typed {
			typed[key] = wellFormedValue(each)
		}
		return typed
	case *estree.Node:
		wellFormedTree(typed)
		return typed
	case []*estree.Node:
		for _, child := range typed {
			wellFormedTree(child)
		}
		return typed
	}
	return value
}
