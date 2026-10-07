package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
)

func decodedFields(t *testing.T, wire string) []string {
	t.Helper()
	units := utf16.Encode([]rune(wire))
	var fields []string
	for cursor := 0; cursor < len(units); {
		end := cursor
		for end < len(units) && units[end] != '\n' {
			end++
		}
		if end == len(units) {
			t.Fatal("missing frame header")
		}
		size, err := strconv.Atoi(string(utf16.Decode(units[cursor:end])))
		if err != nil || size < 0 || end+1+size > len(units) {
			t.Fatal("invalid frame length")
		}
		cursor = end + 1 + size
		fields = append(fields, string(utf16.Decode(units[end+1:cursor])))
	}
	return fields
}

func TestFactEncoding(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"", "ASCII", "世界🌍\nnext", "\x00", "-1"} {
		var got fields
		got.text(text)
		want := fmt.Sprintf("%d\n%s", len(utf16.Encode([]rune(text))), text)
		if got.String() != want {
			t.Fatalf("text framing: %q != %q", got.String(), want)
		}
	}
	for _, number := range []uint64{0, 1, 10, 9007199254740991, ^uint64(0)} {
		var got fields
		got.number(number)
		text := strconv.FormatUint(number, 10)
		want := fmt.Sprintf("%d\n%s", len(text), text)
		if got.String() != want {
			t.Fatalf("number framing: %q != %q", got.String(), want)
		}
	}
}

func TestShapeAndNameFacts(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	config := filepath.Join(directory, "tsconfig.json")
	file := filepath.Join(directory, "file.ts")
	source := "declare const value: \"世界🌍\";\nvalue;\n"
	if err := os.WriteFile(config, []byte(`{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	start := uint64(strings.LastIndex(source, "\nvalue"))
	end := start + uint64(len("\nvalue"))
	ask := func(question string) []string {
		t.Helper()
		wire, err := program.Inspect(file, start, end, "Identifier", question)
		if err != nil {
			t.Fatal(err)
		}
		return decodedFields(t, wire)
	}
	full := ask("raw-type")
	shape := ask("raw-shape")
	// This literal has one root, no rest marks and one record; field 10 is its name.
	if len(full) != 19 || len(shape) != 19 || full[7] != "1" || full[10] != `"世界🌍"` || shape[10] != "" {
		t.Fatalf("unexpected literal records: %q %q", full, shape)
	}
	for i := range full {
		if i != 1 && i != 10 && full[i] != shape[i] {
			t.Fatalf("shape changed field %d: %q != %q", i, full[i], shape[i])
		}
	}
	name := ask("name\n" + shape[5])
	if len(name) != 3 || name[0] != "1" || name[1] != "name" || name[2] != full[10] {
		t.Fatalf("wrong lazy name: %q", name)
	}
	for _, question := range []string{"name", "name\n0", "name\n01", "name\n9007199254740993", "name\nx", "name\n1\nextra", "signature-shape"} {
		if _, err := program.Inspect(file, start, end, "Identifier", question); err == nil {
			t.Fatalf("accepted invalid question %q", question)
		}
	}
	if _, err := program.Inspect(file, start, end, "NumericLiteral", "name\n"+shape[5]); err == nil {
		t.Fatal("name accepted an inexact selector")
	}
}
