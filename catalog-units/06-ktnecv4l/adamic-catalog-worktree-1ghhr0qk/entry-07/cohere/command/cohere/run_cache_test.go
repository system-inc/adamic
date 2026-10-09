package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every line written through a taggingWriter starts with the tag, however the writes are split.
//
// The lines it carries are the ones a replay must not print, so a line that escaped untagged would be
// replayed as current: "graph built in 436ms" on a run that built nothing. Partial writes are the case
// worth testing, since Fprintf can hand over a line in pieces.
func TestTaggingWriterTagsEveryLine(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	writer := &taggingWriter{out: &out, tag: invocationTag, atLineStart: true}
	fmt.Fprint(writer, "graph built in ")
	fmt.Fprint(writer, "436ms\ntypes ran")
	fmt.Fprint(writer, " in 409ms\n")

	want := string(invocationTag) + "graph built in 436ms\n" + string(invocationTag) + "types ran in 409ms\n"
	if out.String() != want {
		t.Fatalf("tagged output:\n  got  %q\n  want %q", out.String(), want)
	}
}

// Without a recording, invocationOutput is the writer itself, so a run that is not recorded prints
// exactly what it printed before the run cache existed.
// Not parallel: it sets the package-level activeRunCache to nil for its length
func TestInvocationOutputIsUntouchedWithoutARecording(t *testing.T) {
	previous := activeRunCache
	activeRunCache = nil
	defer func() { activeRunCache = previous }()

	var out bytes.Buffer
	if invocationOutput(&out) != &out {
		t.Fatal("with no recording in progress, invocation lines were still wrapped")
	}
}

// The tee strips the tag on the way to the terminal and keeps tagged lines out of the replay.
//
// Both halves matter and fail differently. A tag reaching the terminal prints a control byte at the
// front of a line; a tagged line reaching the buffer is replayed as a fact about this run. This drives
// the real tee through a real pipe, since the pipe is the reason the tag travels in the stream.
func TestTeeStripsTagsFromTheTerminalAndTheReplay(t *testing.T) {
	t.Parallel()
	terminal, err := os.Create(filepath.Join(t.TempDir(), "terminal"))
	if err != nil {
		t.Fatal(err)
	}
	defer terminal.Close()

	target := terminal
	var tee teeStream
	if err := tee.start(&target); err != nil {
		t.Fatal(err)
	}
	fmt.Fprint(target, "fix: 0 of 5 files rewritten\n")
	fmt.Fprint(&taggingWriter{out: target, tag: invocationTag, atLineStart: true}, "graph built in 436ms\n")
	fmt.Fprint(target, "coverage: 460 rules\n")
	fmt.Fprint(&taggingWriter{out: target, tag: invocationTag, atLineStart: true}, "phases: fix ran in 2.9s\n  total 4.0s\n")
	fmt.Fprint(target, "  this binary was built from a modified tree\n")
	tee.stop(&target)

	if target != terminal {
		t.Fatal("stop did not restore the real stream")
	}
	shown, err := os.ReadFile(terminal.Name())
	if err != nil {
		t.Fatal(err)
	}

	wantShown := "fix: 0 of 5 files rewritten\ngraph built in 436ms\ncoverage: 460 rules\nphases: fix ran in 2.9s\n  total 4.0s\n  this binary was built from a modified tree\n"
	if string(shown) != wantShown {
		t.Errorf("the terminal did not get every line untagged:\n  got  %q\n  want %q", shown, wantShown)
	}
	wantReplayed := "fix: 0 of 5 files rewritten\ncoverage: 460 rules\n  this binary was built from a modified tree\n"
	if tee.buffer.String() != wantReplayed {
		t.Errorf("the replay holds the wrong lines:\n  got  %q\n  want %q", tee.buffer.String(), wantReplayed)
	}
	if strings.Contains(tee.buffer.String(), string(invocationTag)) || bytes.Contains(shown, invocationTag) {
		t.Error("the tag itself escaped into the terminal or the replay")
	}
}

// A provenance-tagged line reaches the terminal untagged and stays tagged in the recording, so a replay
// can say where it came from; and replayLines turns it into "label (from the cached run at T): rest"
// with no tag left anywhere.
func TestProvenanceLinesAreKeptTaggedAndReplayedWithTheirSource(t *testing.T) {
	t.Parallel()
	terminal, err := os.Create(filepath.Join(t.TempDir(), "terminal"))
	if err != nil {
		t.Fatal(err)
	}
	defer terminal.Close()

	target := terminal
	var tee teeStream
	if err := tee.start(&target); err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(&taggingWriter{out: target, tag: provenanceTag, atLineStart: true}, "fix: 0 of 5 files rewritten, 5 not formatted")
	fmt.Fprint(target, "coverage: 460 rules\n")
	tee.stop(&target)

	shown, err := os.ReadFile(terminal.Name())
	if err != nil {
		t.Fatal(err)
	}
	if string(shown) != "fix: 0 of 5 files rewritten, 5 not formatted\ncoverage: 460 rules\n" {
		t.Errorf("the terminal saw the tag or lost the line: %q", shown)
	}
	if !bytes.HasPrefix(tee.buffer.Bytes(), provenanceTag) {
		t.Fatalf("the recording lost the provenance tag, so a replay would print the line as current: %q", tee.buffer.String())
	}

	replayed := string(replayLines(tee.buffer.Bytes(), "03:41:07"))
	want := "fix (from the cached run at 03:41:07): 0 of 5 files rewritten, 5 not formatted\ncoverage: 460 rules\n"
	if replayed != want {
		t.Errorf("the replay:\n  got  %q\n  want %q", replayed, want)
	}
	if strings.Contains(replayed, string(provenanceTag)) || strings.Contains(replayed, string(invocationTag)) {
		t.Error("a tag survived into the replay")
	}
}

// Exactly two invocations may use the run cache: a bare run and `--no-fix` alone. Any other argument
// widens what the output could depend on, and one admitted by accident would be cached on a key that
// covers it but by a cache never shown to be correct for it.
// Not parallel: it sets os.Args and the package-level cacheOff, which runCacheEligible reads
func TestRunCacheEligibility(t *testing.T) {
	previous := os.Args
	defer func() { os.Args = previous }()

	for _, testCase := range []struct {
		arguments []string
		eligible  bool
	}{
		{nil, true},
		{[]string{"--no-fix"}, true},
		{[]string{"--fix"}, false},
		{[]string{"--lint"}, false},
		{[]string{"--no-fix", "--lint"}, false},
		{[]string{"source/a.ts"}, false},
	} {
		os.Args = append([]string{"cohere"}, testCase.arguments...)
		if got := runCacheEligible(); got != testCase.eligible {
			t.Errorf("cohere %v: eligible %v, want %v", testCase.arguments, got, testCase.eligible)
		}
	}

	// `--no-cache` is refused by name, so it holds even for an argument shape that is otherwise admitted.
	os.Args = []string{"cohere"}
	cacheOff = true
	defer func() { cacheOff = false }()
	if runCacheEligible() {
		t.Error("--no-cache left a bare run eligible, so a cold run could read or record the cache")
	}
}
