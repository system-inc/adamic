package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// header opens every Performance.jsonc. It is fixed, so a landing's diff shows only the numbers that moved.
const header = "// Performance.jsonc: Adamic's performance at this commit, rewritten by the line in every landing commit (#hed1mkb).\n" +
	"// Git history is the series. Never edit it by hand. Times are integer milliseconds: the median of the rounds, with\n" +
	"// their min and max. cmd/adamic-performance writes it; nothing that builds or tests reads it.\n"

// schemaVersion is the file's schema, which a reader refuses when it is another.
const schemaVersion = 1

// A snapshot is the whole file. Its fields are in the file's order, and every one is written, so the keys come out the
// same way in every commit and git's diff of two snapshots is the numbers alone.
type snapshot struct {
	Schema         int    `json:"schema"`
	Commit         string `json:"commit"`
	MeasuredCommit string `json:"measuredCommit"`
	MeasuredAt     string `json:"measuredAt"`
	// Unmeasured is why this commit's numbers are MeasuredCommit's, and empty when they are its own.
	Unmeasured string  `json:"unmeasured,omitempty"`
	Machine    machine `json:"machine"`
	Tools      tools   `json:"tools"`
	Note       string  `json:"note"`
	Compile    compile `json:"compile"`
	// Execute is the execute half's, carried through as it was read: this command doesn't measure it yet.
	Execute json.RawMessage `json:"execute"`
}

// A machine is what the numbers were measured on, since numbers without it aren't worth reading.
type machine struct {
	Name       string     `json:"name"`
	CPU        string     `json:"cpu"`
	Cores      int        `json:"cores"`
	LoadBefore [3]float64 `json:"loadBefore"`
	LoadAfter  [3]float64 `json:"loadAfter"`
	HeldQuiet  bool       `json:"heldQuiet"`
}

// tools are the versions a change of would move every number: the Go that built the adamic binary, the first line of
// clang's --version, and Node's, left out where there is none.
type tools struct {
	Adamic string `json:"adamic"`
	Clang  string `json:"clang"`
	Node   string `json:"node,omitempty"`
}

// compile is the compile half: how long each program takes to build and to reach C.
type compile struct {
	Rounds   int      `json:"rounds"`
	Programs programs `json:"programs"`
}

// A program is one program's compile numbers, every time in milliseconds.
type program struct {
	// Path is the source's path from the repository's root, or generatedPath for one this command writes.
	Path  string `json:"path"`
	Lines int    `json:"lines"`
	// CLines is how many lines of C adamic c prints for it.
	CLines    int        `json:"cLines"`
	WarmBuild spread     `json:"warmBuild"`
	FrontEnd  spread     `json:"frontEnd"`
	Clang     clangTimes `json:"clang"`
	Phases    phaseTimes `json:"phases"`
	// ColdBuild is the one build with every cache empty, which only hello has: the runtime archive's compile is in it.
	ColdBuild *int64 `json:"coldBuild,omitempty"`
}

// A spread is the median of the rounds with their min and max, so a comparison can ask whether two ranges overlap.
type spread struct {
	Median int64 `json:"median"`
	Min    int64 `json:"min"`
	Max    int64 `json:"max"`
}

// clangTimes are the medians of the warm rounds' C compile and link, as the build log reports what clang ran for.
type clangTimes struct {
	Compile int64 `json:"compile"`
	Link    int64 `json:"link"`
}

// phaseTimes are the medians of the warm rounds' compiler phases, as internal/buildlog's lines report each: loading
// the program through typescript-go, lowering it to IR, and emitting its C.
type phaseTimes struct {
	Load  int64 `json:"load"`
	Lower int64 `json:"lower"`
	Emit  int64 `json:"emit"`
}

// A namedProgram is one entry of programs.
type namedProgram struct {
	Name    string
	Program program
}

// programs is the compile half's object of programs, kept in its order: a Go map would write them sorted, and a
// program's place in the file would move whenever one was added.
type programs []namedProgram

// lookup is the named program, and whether there is one.
func (list programs) lookup(name string) (program, bool) {
	for _, entry := range list {
		if entry.Name == name {
			return entry.Program, true
		}
	}
	return program{}, false
}

// MarshalJSON writes the programs as one object, keyed by name in their order.
func (list programs) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteByte('{')
	for index, entry := range list {
		if index > 0 {
			buffer.WriteByte(',')
		}
		name, err := encodeCompact(entry.Name)
		if err != nil {
			return nil, err
		}
		value, err := encodeCompact(entry.Program)
		if err != nil {
			return nil, err
		}
		buffer.Write(name)
		buffer.WriteByte(':')
		buffer.Write(value)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

// UnmarshalJSON reads the object of programs in the order it holds them. A name twice is an error, as is a field no
// program has, so a file this command can't write back as it was is refused rather than read short.
func (list *programs) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return fmt.Errorf("programs isn't an object")
	}
	read := programs{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		name, isName := token.(string)
		if !isName {
			return fmt.Errorf("a program's name isn't a string: %v", token)
		}
		if _, seen := read.lookup(name); seen {
			return fmt.Errorf("the program %q is named twice", name)
		}
		var value program
		if err := decoder.Decode(&value); err != nil {
			return fmt.Errorf("program %q: %w", name, err)
		}
		read = append(read, namedProgram{Name: name, Program: value})
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	*list = read
	return nil
}

// encodeCompact is value as JSON with < > and & as they are: encoding/json's default escapes them, and generatedPath
// would read <generated>.
func encodeCompact(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
}

// encodeSnapshot is the file's bytes: the header, then the snapshot indented with tabs, then a newline. An empty
// execute half is written as {}, never null, so the next run reads it back as the same object.
func encodeSnapshot(value snapshot) ([]byte, error) {
	if len(value.Execute) == 0 {
		value.Execute = json.RawMessage("{}")
	}
	var buffer bytes.Buffer
	buffer.WriteString(header)
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// decodeSnapshot reads a Performance.jsonc: its line comments stripped, then the snapshot, with nothing unknown in it
// and nothing after it, at this schema.
func decodeSnapshot(data []byte) (snapshot, error) {
	decoder := json.NewDecoder(bytes.NewReader(stripLineComments(data)))
	decoder.DisallowUnknownFields()
	var value snapshot
	if err := decoder.Decode(&value); err != nil {
		return snapshot{}, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return snapshot{}, errors.New("something follows the snapshot's object")
	}
	if value.Schema != schemaVersion {
		return snapshot{}, fmt.Errorf("schema %d, and this command reads %d", value.Schema, schemaVersion)
	}
	return value, nil
}

// stripLineComments is data with every // comment outside a string taken out up to its line's end, the newline kept,
// so what's left is JSON and an error's offset is still on the line it was. A scanner rather than a pattern, because a
// string holds // as text (a URL, or a note quoting a path), and its escapes decide where it ends.
func stripLineComments(data []byte) []byte {
	stripped := make([]byte, 0, len(data))
	inString, escaped := false, false
	for index := 0; index < len(data); index++ {
		character := data[index]
		switch {
		case inString:
			switch {
			case escaped:
				escaped = false
			case character == '\\':
				escaped = true
			case character == '"':
				inString = false
			}
		case character == '"':
			inString = true
		case character == '/' && index+1 < len(data) && data[index+1] == '/':
			for index < len(data) && data[index] != '\n' {
				index++
			}
			if index < len(data) {
				stripped = append(stripped, '\n')
			}
			continue
		}
		stripped = append(stripped, character)
	}
	return stripped
}
