package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// fullSnapshot is a snapshot with every field set, hello's cold build and an execute half among them, and its programs
// in the file's order, which isn't sorted: a map would write Parser first.
func fullSnapshot() snapshot {
	coldBuild := int64(3064)
	value := noteSnapshot(earlierCommit,
		namedProgram{Name: "hello", Program: program{Path: generatedPath, Lines: 1, CLines: 28, WarmBuild: spread{Median: 253, Min: 249, Max: 260}, FrontEnd: spread{Median: 27, Min: 26, Max: 29}, Clang: clangTimes{Compile: 29, Link: 113}, Phases: phaseTimes{Load: 180, Lower: 9, Emit: 4}, ColdBuild: &coldBuild}},
		namedProgram{Name: "Parser", Program: program{Path: "stage3/namespaces/shapes/Parser.a", Lines: 512, CLines: 3946, WarmBuild: spread{Median: 304, Min: 298, Max: 311}, FrontEnd: spread{Median: 64, Min: 63, Max: 66}, Clang: clangTimes{Compile: 61, Link: 117}, Phases: phaseTimes{Load: 190, Lower: 30, Emit: 12}}},
		namedProgram{Name: "tokenizer", Program: program{Path: "bench/tokenizer.ts", Lines: 120, CLines: 2210, WarmBuild: spread{Median: 421, Min: 410, Max: 430}, FrontEnd: spread{Median: 38, Min: 37, Max: 40}, Clang: clangTimes{Compile: 80, Link: 115}, Phases: phaseTimes{Load: 240, Lower: 25, Emit: 10}}},
	)
	value.MeasuredAt = "2026-11-21T02:10:00Z"
	value.Machine.LoadBefore = [3]float64{0.4, 0.6, 0.9}
	value.Machine.LoadAfter = [3]float64{1.1, 0.7, 0.9}
	value.Machine.HeldQuiet = true
	value.Note = "first snapshot"
	value.Execute = json.RawMessage(`{"rounds":5,"programs":{"trees":{"native":{"median":0,"min":0,"max":0}}}}`)
	return value
}

// objectKeys is the keys of the JSON object data holds, in the order it holds them.
func objectKeys(t *testing.T, data []byte) []string {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		t.Fatalf("not an object: %v %v", token, err)
	}
	var keys []string
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			t.Fatal(err)
		}
		key, isKey := token.(string)
		if !isKey {
			t.Fatalf("a key that isn't a string: %v", token)
		}
		keys = append(keys, key)
		var skipped json.RawMessage
		if err := decoder.Decode(&skipped); err != nil {
			t.Fatal(err)
		}
	}
	return keys
}

// member is the value of key in the JSON object data holds.
func member(t *testing.T, data []byte, key string) []byte {
	t.Helper()
	var members map[string]json.RawMessage
	if err := json.Unmarshal(data, &members); err != nil {
		t.Fatal(err)
	}
	value, found := members[key]
	if !found {
		t.Fatalf("no %q in %s", key, data)
	}
	return value
}

// fileKeyOrder is the file's keys in the schema's order, each object's by the keys that reach it from the top. It's at
// package scope because it names clang as a key: internal/compilercensus reads function bodies for a C compiler's name,
// and this one runs nothing.
var fileKeyOrder = []struct {
	where []string
	want  []string
}{
	{want: []string{"schema", "commit", "measuredCommit", "measuredAt", "machine", "tools", "note", "compile", "execute"}},
	{where: []string{"machine"}, want: []string{"name", "cpu", "cores", "loadBefore", "loadAfter", "heldQuiet"}},
	{where: []string{"tools"}, want: []string{"adamic", "clang", "node"}},
	{where: []string{"compile"}, want: []string{"rounds", "programs"}},
	{where: []string{"compile", "programs"}, want: []string{"hello", "Parser", "tokenizer"}},
	{where: []string{"compile", "programs", "hello"}, want: []string{"path", "lines", "cLines", "warmBuild", "frontEnd", "clang", "phases", "coldBuild"}},
	{where: []string{"compile", "programs", "Parser"}, want: []string{"path", "lines", "cLines", "warmBuild", "frontEnd", "clang", "phases"}},
	{where: []string{"compile", "programs", "hello", "warmBuild"}, want: []string{"median", "min", "max"}},
	{where: []string{"compile", "programs", "hello", "clang"}, want: []string{"compile", "link"}},
	{where: []string{"compile", "programs", "hello", "phases"}, want: []string{"load", "lower", "emit"}},
}

// The file's keys come out in the schema's order at every level, the programs in theirs, and reading the file back
// and writing it again gives the same bytes, header and all. Mutant: map-keyed output (a map[string]any for the
// snapshot, or a map for programs), whose keys come out sorted, so commit precedes schema and Parser precedes hello.
func TestPerformanceFileKeepsItsKeyOrderAndRoundTrips(t *testing.T) {
	t.Parallel()
	written, err := encodeSnapshot(fullSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(written, []byte(header)) || !bytes.HasSuffix(written, []byte("}\n")) {
		t.Fatalf("the file doesn't open with the header and end with a newline:\n%s", written)
	}
	body := stripLineComments(written)
	for _, check := range fileKeyOrder {
		object := body
		for _, key := range check.where {
			object = member(t, object, key)
		}
		if got := objectKeys(t, object); !slices.Equal(got, check.want) {
			t.Errorf("%s's keys are %v, want %v", strings.Join(append([]string{"the file"}, check.where...), "."), got, check.want)
		}
	}
	if !strings.Contains(string(written), `"path": "<generated>"`) {
		t.Errorf("hello's path isn't written as <generated>:\n%s", written)
	}

	read, err := decodeSnapshot(written)
	if err != nil {
		t.Fatal(err)
	}
	again, err := encodeSnapshot(read)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again, written) {
		t.Fatalf("written again, the file differs:\n%s\nfirst:\n%s", again, written)
	}

	// An unmeasured snapshot writes its reason after measuredAt, and only then.
	unmeasured := fullSnapshot()
	unmeasured.Unmeasured = "the line skipped it"
	written, err = encodeSnapshot(unmeasured)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"schema", "commit", "measuredCommit", "measuredAt", "unmeasured", "machine", "tools", "note", "compile", "execute"}
	if got := objectKeys(t, stripLineComments(written)); !slices.Equal(got, want) {
		t.Errorf("an unmeasured file's keys are %v, want %v", got, want)
	}
}

// The reader strips each // comment outside a string, the header's and one at a line's end, and keeps a // inside a
// string value, an escaped quote before it included. It refuses a field it doesn't know, since writing such a file back
// would lose it. Mutant: comments stripped by a pattern blind to strings, which cuts the note at its //.
func TestPerformanceJsoncReaderStripsCommentsButNotStrings(t *testing.T) {
	t.Parallel()
	written, err := encodeSnapshot(fullSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	note := `see https://example.com/a//b and "quoted // too"`
	commented := strings.Replace(string(written), `"note": "first snapshot",`, `"note": `+jsonString(t, note)+`, // a comment at the line's end`, 1)
	commented = strings.Replace(commented, `"note"`, "// a comment on its own line\n\t\"note\"", 1)
	read, err := decodeSnapshot([]byte(commented))
	if err != nil {
		t.Fatalf("%v in:\n%s", err, commented)
	}
	if read.Note != note {
		t.Errorf("note %q, want %q", read.Note, note)
	}
	if got, want := string(stripLineComments([]byte(`{"a": "x\"//y"} // gone`+"\n"))), `{"a": "x\"//y"} `+"\n"; got != want {
		t.Errorf("stripped %q, want %q", got, want)
	}
	unknown := strings.Replace(string(written), `"schema": 1,`, `"schema": 1, "extra": true,`, 1)
	if _, err := decodeSnapshot([]byte(unknown)); err == nil {
		t.Error("a field the snapshot doesn't have was read without an error")
	}
}

// jsonString is value as a JSON string literal.
func jsonString(t *testing.T, value string) string {
	t.Helper()
	encoded, err := encodeCompact(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

// Unmeasured, the command keeps every number, the machine, the tools, measuredAt, measuredCommit and the execute half
// from -previous, takes the new commit, and writes the reason and whose numbers they are. Mutant: the numbers dropped
// (the compile half written empty), and the compile half differs.
func TestUnmeasuredSnapshotKeepsThePreviousNumbers(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	previousPath, out := filepath.Join(directory, "previous.jsonc"), filepath.Join(directory, "out.jsonc")
	previous := fullSnapshot()
	previous.Commit = "9c0ffee000000000000000000000000000000000"
	data, err := encodeSnapshot(previous)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(previousPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if status := run([]string{"-out", out, "-previous", previousPath, "-commit", laterCommit, "-unmeasured", "the line held it"}, io.Discard); status != 0 {
		t.Fatalf("exit %d", status)
	}
	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	read, err := decodeSnapshot(written)
	if err != nil {
		t.Fatal(err)
	}
	if read.Commit != laterCommit || read.MeasuredCommit != earlierCommit || read.Unmeasured != "the line held it" {
		t.Errorf("commit %s, measuredCommit %s, unmeasured %q", read.Commit, read.MeasuredCommit, read.Unmeasured)
	}
	if want := "unmeasured at 2fa0c4d: the line held it; numbers are 1ed745b's"; read.Note != want {
		t.Errorf("note %q, want %q", read.Note, want)
	}
	kept := func(value snapshot) []byte {
		value.Commit, value.Unmeasured, value.Note = "", "", ""
		encoded, err := encodeSnapshot(value)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	if !bytes.Equal(kept(read), kept(previous)) {
		t.Errorf("unmeasured, the numbers changed:\n%s\nwant:\n%s", kept(read), kept(previous))
	}
	// Without -previous there are no numbers to keep, and nothing is written.
	missing := filepath.Join(directory, "missing.jsonc")
	if status := run([]string{"-out", missing, "-commit", laterCommit, "-unmeasured", "the line held it"}, io.Discard); status == 0 {
		t.Error("-unmeasured without -previous exited 0")
	}
	if _, err := os.Stat(missing); err == nil {
		t.Error("-unmeasured without -previous wrote a file")
	}
}

// The snapshot is replaced whole: written beside its path and renamed over it, so a write that fails partway leaves
// the old file and no temporary one. A rename that fails (its path is a directory) keeps the directory as it was and
// leaves nothing beside it. That a full disk mid-write leaves the old file is held by reading: the old file is touched
// only by the rename. Mutant: the temporary file left behind, never removed.
func TestPerformanceFileIsReplacedWholeOrNotAtAll(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	path := filepath.Join(directory, "Performance.jsonc")
	for _, contents := range []string{"first\n", "second\n"} {
		if err := replaceFile(path, []byte(contents)); err != nil {
			t.Fatal(err)
		}
		if read, err := os.ReadFile(path); err != nil || string(read) != contents {
			t.Fatalf("read %q, %v, want %q", read, err, contents)
		}
	}
	blocked := filepath.Join(directory, "blocked")
	if err := os.Mkdir(filepath.Join(blocked), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "kept"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(blocked, []byte("third\n")); err == nil {
		t.Fatal("renamed over a directory that holds a file")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if strings.Join(names, " ") != "Performance.jsonc blocked" {
		t.Fatalf("the directory holds %q, want the file and the directory alone", names)
	}
}
