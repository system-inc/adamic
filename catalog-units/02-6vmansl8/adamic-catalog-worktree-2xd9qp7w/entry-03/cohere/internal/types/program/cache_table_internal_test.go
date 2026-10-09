package program

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var internalTestIdentity = CacheTableIdentity{SelfCommit: "test", CompilerCommit: "test", GoToolchain: "go-test", Platform: "test/test"}

// encodeRaw writes a header and a body exactly as given, so a test can produce what no current encoder
// would: an old version, a wrong magic, another commit, a list index past its table.
func encodeRaw(t *testing.T, header cacheTableHeader, body any) []byte {
	t.Helper()
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(header); err != nil {
		t.Fatal(err)
	}
	if err := encoder.Encode(body); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func currentHeader(section string) cacheTableHeader {
	return cacheTableHeader{Magic: cacheTableMagic, Version: cacheTableVersion, Identity: internalTestIdentity, Section: section}
}

// writeRaw puts a raw file into a table's directory.
func writeRaw(t *testing.T, directory string, name string, contents []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), contents, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCacheTableRefusesAnotherFormat covers the header fields no other build of this code can get wrong,
// and so only a raw write can: whatever the version is now, never a number someone must remember to
// update, the magic, and the section a file says it holds.
func TestCacheTableRefusesAnotherFormat(t *testing.T) {
	t.Parallel()
	for name, header := range map[string]cacheTableHeader{
		"an older version":  {Magic: cacheTableMagic, Version: cacheTableVersion - 1, Identity: internalTestIdentity, Section: cacheTableTypesSection},
		"a newer version":   {Magic: cacheTableMagic, Version: cacheTableVersion + 1, Identity: internalTestIdentity, Section: cacheTableTypesSection},
		"another magic":     {Magic: "something else", Version: cacheTableVersion, Identity: internalTestIdentity, Section: cacheTableTypesSection},
		"another section":   {Magic: cacheTableMagic, Version: cacheTableVersion, Identity: internalTestIdentity, Section: cacheTableFormatSection},
		"another toolchain": {Magic: cacheTableMagic, Version: cacheTableVersion, Identity: CacheTableIdentity{SelfCommit: "test", CompilerCommit: "test", GoToolchain: "other", Platform: "test/test"}, Section: cacheTableTypesSection},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			writeRaw(t, directory, cacheTableFile(cacheTableTypesSection), encodeRaw(t, header, &TypesSection{Version: typesSectionVersion}))
			table, err := ReadCacheTable(directory, internalTestIdentity, CacheTableSections{Types: true}, PathAnchor{})
			if !errors.Is(err, ErrCacheTableUnreadable) || table.Types != nil {
				t.Fatalf("read %s: section %v, error %v", name, table.Types, err)
			}
		})
	}
}

// TestCacheTableRefusesAListIndexPastItsTable pins the one check gob cannot make for us. The rule lists
// are interned by hand, and an index trusted past its table would hand an entry another file's rules.
func TestCacheTableRefusesAListIndexPastItsTable(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"Rules", "TypedRules", "Listening", "ShapedRules"} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			entry := lintCacheWireEntry{Path: "/a.ts"}
			reflect.ValueOf(&entry).Elem().FieldByName(field).SetInt(1)
			wire := &lintCacheWire{Version: lintCacheVersion, Lists: [][]string{{"no-debugger"}}, Entries: []lintCacheWireEntry{entry}}
			directory := t.TempDir()
			writeRaw(t, directory, cacheTableFile(cacheTableFindingsSection), encodeRaw(t, currentHeader(cacheTableFindingsSection), wire))
			table, err := ReadCacheTable(directory, internalTestIdentity, CacheTableSections{Findings: true}, PathAnchor{})
			if !errors.Is(err, ErrCacheTableUnreadable) || table.Findings != nil {
				t.Fatalf("an entry whose %s index is past the list table was read: %v", field, err)
			}
		})
	}
}

// TestCacheTableDropsAFindingsSectionOfAnotherMeaning: the findings section keeps its own version, for
// changes in what an entry means rather than in its shape. One of another version is dropped, which is a
// miss, and the run beside it is kept, since each is proven by its own key.
func TestCacheTableDropsAFindingsSectionOfAnotherMeaning(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeRaw(t, directory, cacheTableRunFile("--no-fix"),
		encodeRaw(t, currentHeader(cacheTableRunSection), cacheTableRunBody{Invocation: "--no-fix", Run: &RunCache{Key: "kept"}}))
	writeRaw(t, directory, cacheTableFile(cacheTableFindingsSection), encodeRaw(t, currentHeader(cacheTableFindingsSection),
		&lintCacheWire{Version: lintCacheVersion - 1, Lists: [][]string{{}}, Entries: []lintCacheWireEntry{{Path: "/a.ts"}}}))
	table, err := ReadCacheTable(directory, internalTestIdentity, CacheTableSections{Runs: []string{"--no-fix"}, Findings: true}, PathAnchor{})
	if err != nil {
		t.Fatalf("reading: %v", err)
	}
	if table.Findings != nil {
		t.Error("a findings section of another version was used")
	}
	if table.Runs["--no-fix"] == nil {
		t.Error("the run was thrown away with the findings")
	}
}

// A file written by another cohere commit is discarded alone (#45ekc65): planted beside current files, it
// takes nothing else with it, and the error names it. The format record from another commit is the one file
// kept, for its own key to decide. A run file answering for another invocation than its name is dropped too.
func TestAFileFromAnotherCohereCommitIsDiscardedAlone(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	current := &CacheTable{
		Runs:       map[string]*RunCache{"": {Key: "bare"}, "--no-fix": {Key: "no-fix"}},
		Findings:   &LintCache{Version: lintCacheVersion, Entries: []LintCacheEntry{{Path: "/a.ts"}}},
		Signatures: map[string]SignatureEntry{"/a.ts": {Version: "1"}},
		Types:      &TypesSection{Version: typesSectionVersion},
	}
	if err := WriteCacheTable(directory, current, internalTestIdentity, EveryCacheTableSection, PathAnchor{}); err != nil {
		t.Fatal(err)
	}
	other := internalTestIdentity
	other.SelfCommit = "another"
	otherHeader := func(section string) cacheTableHeader {
		header := currentHeader(section)
		header.Identity = other
		return header
	}
	writeRaw(t, directory, cacheTableFile(cacheTableSignaturesSection),
		encodeRaw(t, otherHeader(cacheTableSignaturesSection), map[string]SignatureEntry{"/a.ts": {Version: "stale"}}))
	writeRaw(t, directory, cacheTableFile(cacheTableFormatSection),
		encodeRaw(t, otherHeader(cacheTableFormatSection), &FormatSection{Key: "formatter"}))

	table, err := ReadCacheTable(directory, internalTestIdentity, EveryCacheTableSection, PathAnchor{})
	if !errors.Is(err, ErrCacheTablePartlyKept) || !strings.Contains(err.Error(), cacheTableFile(cacheTableSignaturesSection)) {
		t.Fatalf("the planted file was not reported as another commit's: %v", err)
	}
	if table.Signatures != nil {
		t.Error("another commit's signatures were used")
	}
	if table.Formatted == nil || table.Formatted.Key != "formatter" {
		t.Error("another commit's format record was not kept for its key to decide")
	}
	if table.Runs[""] == nil || table.Runs["--no-fix"] == nil || table.Findings == nil || table.Types == nil {
		t.Errorf("a current file went with the planted one: runs %v, findings %v, types %v", table.Runs, table.Findings != nil, table.Types != nil)
	}

	// A run file under the wrong name is refused, whatever it holds.
	if err := os.Rename(filepath.Join(directory, cacheTableRunFile("")), filepath.Join(directory, cacheTableRunFile("--fix"))); err != nil {
		t.Fatal(err)
	}
	table, err = ReadCacheTable(directory, internalTestIdentity, CacheTableSections{Runs: []string{"--fix"}}, PathAnchor{})
	if !errors.Is(err, ErrCacheTableUnreadable) || table.Runs["--fix"] != nil {
		t.Fatalf("a run recorded for the bare invocation answered for --fix: %v", err)
	}
}

// The single file the table was before format 8 is removed on sight and never read: a table.gob of any
// content is gone after a read, and the read reports nothing found.
func TestTheSingleFileTableIsRemovedUnread(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeRaw(t, directory, legacyCacheTableFile, []byte("whatever an older cohere wrote"))
	_, err := ReadCacheTable(directory, internalTestIdentity, EveryCacheTableSection, PathAnchor{})
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("reading a directory holding only table.gob: %v", err)
	}
	if _, statError := os.Stat(filepath.Join(directory, legacyCacheTableFile)); !errors.Is(statError, os.ErrNotExist) {
		t.Errorf("table.gob is still there: %v", statError)
	}
}

// pinnedCacheTableShapes is the encoded shape each format version stands for. gob decodes a renamed,
// added or removed field without error, into a zero value, so a shape that changes under an unchanged
// version is a silent partial read of every old table. Changing a field means bumping cacheTableVersion
// and pinning the new shape under it, and this test says so when it is forgotten.
var pinnedCacheTableShapes = map[int]string{
	1:  "3db631dcfd27db51bf5c9e29a4a0ac41a7ba30dc90aeb8712037e057a516effb",
	2:  "30d879299d079a880679687f46936674d9ecd124a5916a1ea76ff210d6e67873",
	3:  "c01be7eecead0ab5c95964ed45cb149bcfd0883b2130a21e39d8d25e3c1d3622",
	4:  "f215adaa07ccb0ee6926835db75c04bc8d9b001003eefac84388b95e1f85e9e9",
	5:  "850508d7af486cc0b05eab3dd93ac299e60b678a00ff30e8702f98210505949f",
	6:  "cc48115e6624c13b4d674cf5c0d4c47da536f02516afdd5575344321ce77dc2b",
	7:  "efe906781bbaf6e01d9c195c2b08f8cb9f5e855fdb445bfc5b2edcdada8d4ada",
	8:  "b75bba7c7483bb48bb6fc5c8e2d2b38a3d337fbddb67d95ce423069807772e5e",
	9:  "51a0ac42299c64e9c01ee4daae5e05482437993bec1ae5cb730a7dda7e7f592b",
	10: "295b6d55fc54d1269471e40676b04b3876baac62a8779c23d79e0c9f1118481c",
	11: "05b274b78356fdf614f2b8b60faa41a4fdeaa9522627a8244ff3a4e9e1314507",
	12: "9205583b30639c3716e0f12554904157381c15dfabc355233523722aa3c50594",
	13: "92cec94e75987f5af469a8fe4e611c4e95ee9d56a41ad9003db764354922a60c",
	14: "5c81e8d748ea0cf87759a0e7de3d205cbc70a6ab930d09b7a827d34e2dd9eb4b",
	15: "5b2612f8b8bd8ec1aefe9b94d18dd81bf2a05115282a3c2e80edc8b6fe40abd5",
}

func TestCacheTableShapeIsPinnedToItsVersion(t *testing.T) {
	t.Parallel()
	shape := describeShape(reflect.TypeOf(cacheTableHeader{})) + "\n" + describeShape(reflect.TypeOf(cacheTableBodies{}))
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte(shape)))
	pinned, found := pinnedCacheTableShapes[cacheTableVersion]
	if !found || pinned != sum {
		t.Fatalf("the cache table's encoded shape is %s under version %d, pinned as %q.\n"+
			"If a field changed, bump cacheTableVersion and pin %s under the new version: an old table would "+
			"otherwise decode into zero values without an error.\nThe shape:\n%s",
			sum, cacheTableVersion, pinned, sum, shape)
	}
}

// describeShape renders a type the way gob sees it: exported fields by name and type, recursively.
func describeShape(subject reflect.Type) string {
	var out strings.Builder
	var walk func(reflect.Type, map[reflect.Type]bool)
	walk = func(subject reflect.Type, seen map[reflect.Type]bool) {
		switch subject.Kind() {
		case reflect.Pointer:
			out.WriteString("*")
			walk(subject.Elem(), seen)
		case reflect.Slice:
			out.WriteString("[]")
			walk(subject.Elem(), seen)
		case reflect.Array:
			fmt.Fprintf(&out, "[%d]", subject.Len())
			walk(subject.Elem(), seen)
		case reflect.Map:
			out.WriteString("map[")
			walk(subject.Key(), seen)
			out.WriteString("]")
			walk(subject.Elem(), seen)
		case reflect.Struct:
			if seen[subject] {
				out.WriteString(subject.Name())
				return
			}
			seen[subject] = true
			out.WriteString(subject.Name() + "{")
			for index := range subject.NumField() {
				field := subject.Field(index)
				if !field.IsExported() {
					continue
				}
				out.WriteString(field.Name + " ")
				walk(field.Type, seen)
				out.WriteString("; ")
			}
			out.WriteString("}")
		default:
			out.WriteString(subject.Kind().String())
		}
	}
	walk(subject, map[reflect.Type]bool{})
	return out.String()
}

// Rule lists are interned by their length and their two ends, and that key is only a filter: two lists that
// share it and differ in the middle stay two lists, and every entry resolves back to exactly its own.
func TestInterningKeepsListsThatShareTheirEndsApart(t *testing.T) {
	t.Parallel()
	cache := &LintCache{Version: lintCacheVersion}
	lists := [][]string{
		{"a", "b", "c"},
		{"a", "x", "c"},
		{"a", "b", "c"},
		{},
		nil,
		{"a", "c"},
	}
	for index, list := range lists {
		cache.Entries = append(cache.Entries, LintCacheEntry{Path: fmt.Sprintf("/f%d.ts", index), Rules: list})
	}
	wire := cache.wire()
	if len(wire.Lists) != 4 {
		t.Errorf("%d distinct lists interned from 4 distinct ones: %v", len(wire.Lists), wire.Lists)
	}
	back, err := wire.cache()
	if err != nil {
		t.Fatal(err)
	}
	for index, list := range lists {
		if strings.Join(back.Entries[index].Rules, ",") != strings.Join(list, ",") {
			t.Errorf("entry %d came back with %v, want %v", index, back.Entries[index].Rules, list)
		}
	}
}
