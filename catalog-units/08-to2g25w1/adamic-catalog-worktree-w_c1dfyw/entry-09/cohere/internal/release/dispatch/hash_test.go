package dispatch

import (
	"os"
	"path/filepath"
	"testing"
)

// A hash that misses an input is the defect this whole package exists to prevent: it serves a
// binary built from old rules while the file on disk says the rule is enforced, and the gate prints
// green. Every test here asserts that some specific change to the inputs moves the digest.

func TestHashChangesWhenFileContentsChange(t *testing.T) {
	t.Parallel()

	// The case a filename-based hash misses entirely: a rule edited in place keeps its name.
	directory := t.TempDir()
	path := filepath.Join(directory, "rule.go")
	writeFile(t, path, "package rules\n")

	inputs := Inputs{
		GoVersion:          "go1.27.0 darwin arm64",
		TypeScriptGoCommit: "2b82831a05b6b99da279d12fecbcbc460574a85b",
		BuildFlags:         ReleaseBuildFlags,
		SourceFiles:        []string{path},
	}

	before := compute(t, inputs)

	writeFile(t, path, "package rules\n\n// an edit\n")
	after := compute(t, inputs)

	if before == after {
		t.Fatalf("editing a file left the hash at %s, so a rule change would serve a stale binary", before)
	}
}

func TestHashChangesWhenContentsChangeWithoutChangingLength(t *testing.T) {
	t.Parallel()

	// This is the test that proves the digest reads the bytes rather than a summary of them.
	//
	// It exists because the obvious version of the previous test does not. Each file is framed in
	// the digest with its name and its length, so an edit that changes the length moves the hash
	// through the length field alone — and a Compute that never hashed a single byte of content
	// still passed. A same-length edit removes every signal but the contents themselves.
	directory := t.TempDir()
	path := filepath.Join(directory, "rule.go")
	writeFile(t, path, "package rules // aaa\n")

	inputs := Inputs{SourceFiles: []string{path}}
	before := compute(t, inputs)

	writeFile(t, path, "package rules // bbb\n")
	after := compute(t, inputs)

	if before == after {
		t.Fatalf("a same-length edit left the hash at %s, so the digest is not reading file contents", before)
	}
}

func TestHashChangesWithEachInput(t *testing.T) {
	t.Parallel()

	// Each field is a thing that changes what the binary does. Missing any one of them is a
	// separate way for the cache to lie, so each gets its own assertion rather than one combined
	// check that could pass on the strength of the others.
	directory := t.TempDir()
	path := filepath.Join(directory, "rule.go")
	writeFile(t, path, "package rules\n")

	base := Inputs{
		GoVersion:          "go1.27.0 darwin arm64",
		TypeScriptGoCommit: "2b82831a05b6b99da279d12fecbcbc460574a85b",
		BuildFlags:         ReleaseBuildFlags,
		SourceFiles:        []string{path},
	}
	baseHash := compute(t, base)

	tests := []struct {
		name   string
		modify func(inputs *Inputs)
	}{
		{"a different Go toolchain", func(inputs *Inputs) { inputs.GoVersion = "go1.26.0 darwin arm64" }},
		{"a different platform", func(inputs *Inputs) { inputs.GoVersion = "go1.27.0 linux amd64" }},
		{"a different typescript-go commit", func(inputs *Inputs) { inputs.TypeScriptGoCommit = "0000000000000000000000000000000000000000" }},
		{"different build flags", func(inputs *Inputs) { inputs.BuildFlags = []string{"-trimpath"} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			changed := base
			test.modify(&changed)

			if got := compute(t, changed); got == baseHash {
				t.Fatalf("%s did not change the hash, so the cache would serve a binary built from the other one", test.name)
			}
		})
	}
}

func TestHashIsStableAcrossRuns(t *testing.T) {
	t.Parallel()

	// An unstable hash misses the cache forever, rebuilding on every invocation. That is the
	// opposite failure from staleness and just as much a defect.
	directory := t.TempDir()
	path := filepath.Join(directory, "rule.go")
	writeFile(t, path, "package rules\n")

	inputs := Inputs{
		GoVersion:          "go1.27.0 darwin arm64",
		TypeScriptGoCommit: "2b82831a05b6b99da279d12fecbcbc460574a85b",
		BuildFlags:         ReleaseBuildFlags,
		SourceFiles:        []string{path},
	}

	first := compute(t, inputs)
	for attempt := range 5 {
		if got := compute(t, inputs); got != first {
			t.Fatalf("run %d hashed to %s where the first run hashed to %s", attempt+2, got, first)
		}
	}
}

func TestHashIgnoresSourceFileOrder(t *testing.T) {
	t.Parallel()

	// The toolchain is free to list packages in whatever order it likes, and an order-sensitive
	// hash would miss the cache for reasons that have nothing to do with the code.
	directory := t.TempDir()
	first := filepath.Join(directory, "a.go")
	second := filepath.Join(directory, "b.go")
	writeFile(t, first, "package rules\n")
	writeFile(t, second, "package rules\n\n// b\n")

	forward := compute(t, Inputs{SourceFiles: []string{first, second}})
	backward := compute(t, Inputs{SourceFiles: []string{second, first}})

	if forward != backward {
		t.Fatalf("listing the same files in a different order hashed to %s and %s", forward, backward)
	}
}

func TestHashDistinguishesFilesWithSharedContent(t *testing.T) {
	t.Parallel()

	// Two files whose contents are identical must not be interchangeable with one file: the digest
	// frames each entry with its name and length so that concatenation cannot collide.
	directory := t.TempDir()
	first := filepath.Join(directory, "a.go")
	second := filepath.Join(directory, "b.go")
	writeFile(t, first, "package rules\n")
	writeFile(t, second, "package rules\n")

	one := compute(t, Inputs{SourceFiles: []string{first}})
	both := compute(t, Inputs{SourceFiles: []string{first, second}})

	if one == both {
		t.Fatal("adding a second file with identical contents did not change the hash")
	}
}

func TestComputeFailsOnUnreadableInput(t *testing.T) {
	t.Parallel()

	// Hashing around a file that was listed but cannot be read produces a stable digest for a tree
	// we could not actually see, which is staleness wearing a different hat.
	inputs := Inputs{SourceFiles: []string{filepath.Join(t.TempDir(), "absent.go")}}

	if _, err := inputs.Compute(); err == nil {
		t.Fatal("hashing a missing input succeeded, so an unreadable tree would hash as if it were fine")
	}
}

func TestCollectInputsRejectsAnEmptyFileSet(t *testing.T) {
	t.Parallel()

	// An empty input set hashes to a perfectly stable digest for a tree containing nothing, and
	// every later run hits that entry. A run that measured nothing must not look like a run that
	// found nothing.
	_, err := CollectInputs(t.TempDir(), "./cmd/does-not-exist", ReleaseBuildFlags)
	if err == nil {
		t.Fatal("collecting inputs for a package that does not exist succeeded")
	}
}

// compute hashes inputs, failing the test rather than returning an error.
func compute(t *testing.T, inputs Inputs) string {
	t.Helper()

	hash, err := inputs.Compute()
	if err != nil {
		t.Fatalf("computing the hash: %v", err)
	}
	if hash == "" {
		t.Fatal("the hash came back empty")
	}
	return hash
}

// writeFile writes contents, creating its directory, failing the test on error.
func writeFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating the directory for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}
