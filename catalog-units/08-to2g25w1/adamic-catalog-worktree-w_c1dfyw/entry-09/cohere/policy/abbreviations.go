package policy

import (
	"bytes"
	_ "embed"
)

// Abbreviations.json is the vocabulary the naming rules judge with: nexus's
// consistency-no-abbreviated-identifier in Go, and cohere-swift's of the same name in Swift. The Go rule
// reads and checks it (internal/lint/rules/nexus/abbreviation_vocabulary.go), refusing an unknown key and
// panicking at init on a file it refuses, since the words, their forms and their order are that rule's to
// interpret. The Swift engine compiles in the same bytes (SwiftFiles), so a released engine carries the
// words with nothing handed to it at run time.
//
//go:embed Abbreviations.json
var abbreviationsFile []byte

// AbbreviationsFile is Abbreviations.json exactly as embedded. A copy, so a caller that writes into it
// cannot change what the next caller reads.
func AbbreviationsFile() []byte {
	return bytes.Clone(abbreviationsFile)
}
