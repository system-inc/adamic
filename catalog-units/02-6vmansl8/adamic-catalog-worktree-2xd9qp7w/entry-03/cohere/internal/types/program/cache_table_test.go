package program_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/types/program"
)

// sampleCacheTable holds both sections, so a test that loses one cannot pass by looking only at the other.
func sampleCacheTable() *program.CacheTable {
	table := program.NewCacheTable()
	table.Runs["--no-fix"] = &program.RunCache{
		Version: 3,
		Key:     "a-key",
		Inputs: []program.RunCacheInput{
			{Path: "/project/source/a.ts", Exists: true, Size: 20, ModifiedNanoseconds: 1},
			{Path: "/project/source", Directory: true, Exists: true, ModifiedNanoseconds: 2},
			{Path: "/project/missing.json"},
		},
		Output:                  []byte("lint: 1 findings\n"),
		Errors:                  []byte("note: something\n"),
		RecordedUnixNanoseconds: 42,
		ExitCode:                1,
	}
	table.Findings = sampleLintCache()
	table.Formatted = &program.FormatSection{
		Key:     "format-key",
		Entries: map[string]program.FormatEntry{"/project/source/a.ts": {Sum: "sum", Options: "options", Size: 20, ModifiedNanoseconds: 7}},
	}
	return table
}

// readEverySection reads every section of the table in directory.
func readEverySection(t *testing.T, directory string, identity program.CacheTableIdentity) (*program.CacheTable, error) {
	t.Helper()
	return program.ReadCacheTable(directory, identity, program.EveryCacheTableSection, program.PathAnchor{})
}

// TestCacheTableRoundTripsEverySection is the table as a whole, through its files and back.
func TestCacheTableRoundTripsEverySection(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	original := sampleCacheTable()
	original.Signatures = map[string]program.SignatureEntry{"/project/source/a.ts": {Version: "1", Signature: "s", Syntax: "x"}}
	original.Types = &program.TypesSection{Version: 2, GlobalsClean: true}
	if err := program.WriteCacheTable(directory, original, testIdentity, program.EveryCacheTableSection, program.PathAnchor{}); err != nil {
		t.Fatalf("writing: %v", err)
	}
	decoded, err := readEverySection(t, directory, testIdentity)
	if err != nil {
		t.Fatalf("reading: %v", err)
	}

	run := decoded.Runs["--no-fix"]
	if run == nil {
		t.Fatal("the recorded run did not survive")
	}
	if run.Key != "a-key" || run.ExitCode != 1 || run.RecordedUnixNanoseconds != 42 || run.Version != 3 {
		t.Errorf("run fields: key %q, exit %d, recorded %d, version %d", run.Key, run.ExitCode, run.RecordedUnixNanoseconds, run.Version)
	}
	if string(run.Output) != "lint: 1 findings\n" || string(run.Errors) != "note: something\n" {
		t.Errorf("run streams: %q / %q", run.Output, run.Errors)
	}
	if len(run.Inputs) != 3 || !run.Inputs[1].Directory || run.Inputs[2].Exists || run.Inputs[0].Size != 20 {
		t.Errorf("run inputs: %+v", run.Inputs)
	}
	if decoded.Findings == nil || len(decoded.Findings.Entries) != 2 {
		t.Fatalf("the findings section did not survive beside the run: %+v", decoded.Findings)
	}
	if decoded.Formatted == nil || decoded.Formatted.Key != "format-key" ||
		decoded.Formatted.Entries["/project/source/a.ts"] != (program.FormatEntry{Sum: "sum", Options: "options", Size: 20, ModifiedNanoseconds: 7}) {
		t.Errorf("the format section did not survive: %+v", decoded.Formatted)
	}
	if decoded.Signatures["/project/source/a.ts"] != original.Signatures["/project/source/a.ts"] {
		t.Errorf("the signatures did not survive: %+v", decoded.Signatures)
	}
	if decoded.Types == nil || decoded.Types.Version != 2 || !decoded.Types.GlobalsClean {
		t.Errorf("the types section did not survive: %+v", decoded.Types)
	}
}

// A write replaces the sections it names and no other (#45ekc65): the format record written alone leaves
// every other file byte for byte, and a section named but nil in the table is not written, so a run with
// nothing to say about a section never erases what another run left there.
func TestAWriteReplacesOnlyTheSectionsItNames(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := program.WriteCacheTable(directory, sampleCacheTable(), testIdentity, program.EveryCacheTableSection, program.PathAnchor{}); err != nil {
		t.Fatal(err)
	}
	snapshot := func() map[string]string {
		t.Helper()
		files := map[string]string{}
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			contents, err := os.ReadFile(filepath.Join(directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			files[entry.Name()] = string(contents)
		}
		return files
	}
	before := snapshot()

	formatOnly := &program.CacheTable{Formatted: &program.FormatSection{Key: "another-key"}}
	if err := program.WriteCacheTable(directory, formatOnly, testIdentity, program.EveryCacheTableSection, program.PathAnchor{}); err != nil {
		t.Fatal(err)
	}
	after := snapshot()
	if len(after) != len(before) {
		t.Fatalf("the write changed which files exist: %d before, %d after", len(before), len(after))
	}
	changed := []string{}
	for name := range before {
		if before[name] != after[name] {
			changed = append(changed, name)
		}
	}
	if len(changed) != 1 || changed[0] != "format.gob" {
		t.Errorf("writing the format record alone changed %v, want only format.gob", changed)
	}
	read, err := readEverySection(t, directory, testIdentity)
	if err != nil || read.Formatted.Key != "another-key" || read.Runs["--no-fix"] == nil || read.Findings == nil {
		t.Errorf("after writing the format record alone: format %+v, run %v, findings %v, error %v",
			read.Formatted, read.Runs["--no-fix"] != nil, read.Findings != nil, err)
	}
}

// TestCacheTableDiscardsWhatItCannotTrust is every way a file can be wrong, each of which must be a
// discard with a reason rather than a section assembled out of the wrong bytes. A finding pointing at the
// wrong rule and range is worse than no finding, because it reads as a real result.
func TestCacheTableDiscardsWhatItCannotTrust(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	if err := program.WriteCacheTable(source, sampleCacheTable(), testIdentity, program.CacheTableSections{Findings: true}, program.PathAnchor{}); err != nil {
		t.Fatalf("writing: %v", err)
	}
	valid, err := os.ReadFile(filepath.Join(source, "findings.gob"))
	if err != nil {
		t.Fatal(err)
	}
	otherCompiler := testIdentity
	otherCompiler.CompilerCommit = "another"
	otherToolchain := testIdentity
	otherToolchain.GoToolchain = "go-other"
	otherPlatform := testIdentity
	otherPlatform.Platform = "other/other"

	cases := []struct {
		name     string
		buffer   []byte
		identity program.CacheTableIdentity
	}{
		{"empty", nil, testIdentity},
		{"truncated in the header", valid[:10], testIdentity},
		{"truncated in the body", valid[:len(valid)/2], testIdentity},
		{"one byte short", valid[:len(valid)-1], testIdentity},
		{"trailing bytes", append(append([]byte{}, valid...), 0, 0, 0, 0), testIdentity},
		{"not a table at all", []byte("{\"version\":5}"), testIdentity},
		{"written against another compiler commit", valid, otherCompiler},
		{"written by another Go toolchain", valid, otherToolchain},
		{"written for another platform", valid, otherPlatform},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			if err := os.WriteFile(filepath.Join(directory, "findings.gob"), testCase.buffer, 0o600); err != nil {
				t.Fatal(err)
			}
			read, err := program.ReadCacheTable(directory, testCase.identity, program.CacheTableSections{Findings: true}, program.PathAnchor{})
			if read.Findings != nil {
				t.Fatalf("read without discarding into %d entries; the reader cannot detect this and every "+
					"clean read from it is vacuous", len(read.Findings.Entries))
			}
			if !errors.Is(err, program.ErrCacheTableUnreadable) {
				t.Errorf("error should be ErrCacheTableUnreadable so a caller runs without it rather than failing: %v", err)
			}
		})
	}
}

// Files from another cohere commit keep only their format record, whose own key decides whether it still
// holds, and drop the runs and findings, whose keys name the binary and could never match anyway.
func TestCacheTableFromAnotherCohereCommitKeepsOnlyTheFormatRecord(t *testing.T) {
	t.Parallel()
	otherCommit := testIdentity
	otherCommit.SelfCommit = "another"
	directory := t.TempDir()
	if err := program.WriteCacheTable(directory, sampleCacheTable(), otherCommit, program.EveryCacheTableSection, program.PathAnchor{}); err != nil {
		t.Fatal(err)
	}

	table, err := readEverySection(t, directory, testIdentity)
	if !errors.Is(err, program.ErrCacheTablePartlyKept) {
		t.Fatalf("a table from another cohere commit should be partly kept, got %v", err)
	}
	if errors.Is(err, program.ErrCacheTableUnreadable) {
		t.Error("a partly kept table was reported as wholly discarded")
	}
	if table.Formatted == nil || table.Formatted.Key != "format-key" || len(table.Formatted.Entries) != 1 {
		t.Errorf("the format record did not survive: %+v", table.Formatted)
	}
	if len(table.Runs) != 0 || table.Findings != nil {
		t.Errorf("runs or findings from another cohere commit were kept: %d runs, findings %v", len(table.Runs), table.Findings != nil)
	}
	table.Runs["bare"] = nil
}

// TestReadCacheTableAlwaysReturnsATable pins the caller's contract: a table to use, and an error that says
// whether its emptiness is worth reporting. Nothing on disk is a first run; anything else is a discard.
func TestReadCacheTableAlwaysReturnsATable(t *testing.T) {
	t.Parallel()
	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		table, err := readEverySection(t, filepath.Join(t.TempDir(), "absent"), testIdentity)
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("a missing table should say it is missing, got %v", err)
		}
		if errors.Is(err, program.ErrCacheTableUnreadable) {
			t.Error("a missing table was reported as a discard, so every first run would print a warning")
		}
		if table == nil || table.Runs == nil || table.Findings != nil {
			t.Fatalf("a missing table should come back empty and usable, got %+v", table)
		}
		table.Runs["bare"] = nil
	})

	t.Run("corrupt", func(t *testing.T) {
		t.Parallel()
		directory := t.TempDir()
		if err := os.WriteFile(filepath.Join(directory, "findings.gob"), []byte("{not a table"), 0o600); err != nil {
			t.Fatal(err)
		}
		table, err := readEverySection(t, directory, testIdentity)
		if !errors.Is(err, program.ErrCacheTableUnreadable) {
			t.Errorf("a corrupt file should be a discard, got %v", err)
		}
		if table == nil || len(table.Runs) != 0 || table.Findings != nil {
			t.Fatalf("a corrupt table should come back empty, got %+v", table)
		}
	})
}

// TestCacheTableSurvivesDisk round-trips through the real filesystem, because the write path is where a
// cache is truncated or half-visible, and asserts the directory holds only the finished files: a leftover
// temporary means a failure path forgot to clean up, and it accumulates silently across runs.
func TestCacheTableSurvivesDisk(t *testing.T) {
	t.Parallel()
	directory := filepath.Join(t.TempDir(), "nested")
	original := sampleCacheTable()

	if err := program.WriteCacheTable(directory, original, testIdentity, program.EveryCacheTableSection, program.PathAnchor{}); err != nil {
		t.Fatalf("writing: %v", err)
	}
	read, err := readEverySection(t, directory, testIdentity)
	if err != nil {
		t.Fatalf("reading back: %v", err)
	}

	// The interpolated description is the one thing in a finding nothing can rebuild, so it is the one
	// worth asserting survived a real write.
	entry, hit := read.Findings.Lookup(original.Findings.Entries[0].Path, original.Findings.Entries[0].ContentHash,
		original.Findings.Key, original.Findings.Entries[0].Rules)
	if !hit {
		t.Fatal("an entry written to disk came back as a miss")
	}
	if entry.Findings[1].MessageDescription != original.Findings.Entries[0].Findings[1].MessageDescription {
		t.Errorf("description did not survive disk: %q", entry.Findings[1].MessageDescription)
	}
	if read.Runs["--no-fix"] == nil {
		t.Error("the run did not survive disk")
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	if len(names) != 3 || names[0] != "findings.gob" || names[1] != "format.gob" || !strings.HasPrefix(names[2], "run-") {
		t.Errorf("the directory holds %v, want exactly findings.gob, format.gob and one run file", names)
	}
}
