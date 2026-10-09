package policy

import (
	"bytes"
	"fmt"
	"strings"
)

// The Swift engine compiles policy in rather than reading it at run time: a released engine has no
// checkout beside it, and a file the front door hands it is one more thing that can go missing between
// the two. So each policy file Swift reads becomes a generated Swift source holding the file's bytes
// verbatim, which the engine parses exactly as it would have parsed the file.
//
// The generated files are checked in, written by `go run ./policy/tools/generate`, rather than made by a
// SwiftPM build plugin. Measured 2026-10-03 (#2cqemcd) on a probe package that compiles in
// Abbreviations.json: the plugin worked, reading the file from outside the package, and cost 4.1s on a
// cold build (its generator is a second executable) and 0.19s on every no-op build. A checked-in file
// costs neither, builds the same under every build system and toolchain the release supports, and can be
// read and searched like any other source. What it costs is a copy that can go stale, so both engines
// test it: TestSwiftFilesAreCurrent here, and the Swift tests compare the compiled-in bytes with the file
// on disk.

// SwiftFile is one generated Swift source, at its path from the module root.
type SwiftFile struct {
	Path     string
	Contents []byte
}

// swiftStringIndentation is where the house format puts a multi-line string's lines inside a static
// property of a type.
const swiftStringIndentation = "        "

// swiftDirectory is where the generated sources live, from the module root.
const swiftDirectory = "swift/Sources/CohereSwift/Policy/"

// swiftPolicyFiles are the policy the Swift engine compiles in, each with the type that carries it and
// what writes that type's source.
var swiftPolicyFiles = []struct {
	source   string
	typeName string
	generate func() ([]byte, error)
}{
	{"Abbreviations.json", "PolicyAbbreviations", func() ([]byte, error) {
		return swiftSource("Abbreviations.json", "PolicyAbbreviations", abbreviationsFile)
	}},
	{"messages/", "RuleMessages", func() ([]byte, error) {
		return swiftMessagesSource(Messages, swiftMessagesDigest(Messages))
	}},
}

// SwiftFiles are the generated sources for all the policy the Swift engine compiles in.
func SwiftFiles() ([]SwiftFile, error) {
	files := make([]SwiftFile, 0, len(swiftPolicyFiles))
	for _, policyFile := range swiftPolicyFiles {
		contents, err := policyFile.generate()
		if err != nil {
			return nil, fmt.Errorf("policy/%s: %w", policyFile.source, err)
		}
		files = append(files, SwiftFile{Path: swiftDirectory + policyFile.typeName + ".generated.swift", Contents: contents})
	}
	return files, nil
}

// swiftSource is a Swift enum whose `file` is the policy file's text, byte for byte.
//
// The text sits in a multi-line raw string, so nothing in it is escaped and the generated file reads as
// the JSON it carries. Two things could make the string differ from the file, and each is ruled out:
//   - A raw string ends at `"""` followed by its own number of `#`, and treats `\` followed by that many as
//     an escape. The delimiter takes one more `#` than the longest run in the text, so neither can occur.
//   - A multi-line literal reads every line ending as `\n`. A file with a carriage return is refused rather
//     than changed.
//
// Every line is written in the house format, so the engine's own format check passes the file as
// generated: each line of text indented to the closing delimiter, which Swift strips from every line, and
// an empty line left empty, which Swift allows. The line break before the delimiter is not part of the
// string, so one is written after the text, and the text ends as the file ends, final newline or not.
func swiftSource(name string, typeName string, contents []byte) ([]byte, error) {
	if bytes.ContainsRune(contents, '\r') {
		return nil, fmt.Errorf("it holds a carriage return, which a Swift multi-line string would read as a line feed")
	}
	delimiter := strings.Repeat("#", longestRun(contents, '#')+1)

	var source bytes.Buffer
	fmt.Fprintf(&source, "/*\n Generated from cohere's policy/%s by `go run ./policy/tools/generate`. Do not edit it here:\n", name)
	source.WriteString(" edit the JSON and run that again. While the two differ, a test in each engine fails.\n */\n")
	fmt.Fprintf(&source, "enum %s {\n", typeName)
	fmt.Fprintf(&source, "    /* policy/%s, byte for byte. */\n", name)
	fmt.Fprintf(&source, "    static let file = %s\"\"\"\n", delimiter)
	for index, line := range strings.Split(string(contents), "\n") {
		if index > 0 {
			source.WriteString("\n")
		}
		if line != "" {
			source.WriteString(swiftStringIndentation + line)
		}
	}
	fmt.Fprintf(&source, "\n%s\"\"\"%s\n}\n", swiftStringIndentation, delimiter)
	return source.Bytes(), nil
}

// longestRun is the most consecutive copies of one byte anywhere in the contents.
func longestRun(contents []byte, character byte) int {
	longest, current := 0, 0
	for _, each := range contents {
		if each != character {
			current = 0
			continue
		}
		current++
		longest = max(longest, current)
	}
	return longest
}
