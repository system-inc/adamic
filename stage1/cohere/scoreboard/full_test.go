package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSweepStatesAndMutant(t *testing.T) {
	t.Parallel()
	if ruleFamily("@next/next/no-img-element") != "next" || ruleFamily("@typescript-eslint/no-explicit-any") != "typescript" {
		t.Fatal("registered and unported namespace families differ")
	}
	for _, test := range []struct {
		a, b         WorkAnswer
		format       bool
		state, cause string
	}{
		{WorkAnswer{Output: "ok\tx\n"}, WorkAnswer{Output: "ok\tx\n"}, true, "agree", ""},
		{WorkAnswer{Output: "ok\tx\n"}, WorkAnswer{Output: "ok\tmutant\n"}, true, "diverge", ""},
		{WorkAnswer{Output: "error\tinvalid\n"}, WorkAnswer{Output: "error\tinvalid\n"}, true, "blocked", "formatter refusal"},
		{WorkAnswer{Output: "skipped typed no program\n"}, WorkAnswer{Output: "skipped typed no program\n"}, false, "blocked", "checker fact"},
		{WorkAnswer{Error: "invalid corpus x"}, WorkAnswer{}, false, "blocked", "parser"},
		{WorkAnswer{}, WorkAnswer{Exit: 70}, false, "blocked", "execution fault"},
		{WorkAnswer{Error: "worker timeout", Exit: -1}, WorkAnswer{}, true, "blocked", "execution fault"},
	} {
		state, cause := stateCode(test.a, test.b, test.format)
		if state != test.state || cause != test.cause {
			t.Fatalf("state=%s cause=%s; expected %s %s", state, cause, test.state, test.cause)
		}
	}
	total := bump(RuleTotal{}, "diverge", "", 1, 0)
	if total.Diverge != 1 || total.Checks != 1 || total.Findings != 1 {
		t.Fatal("streamed mutant disappeared")
	}
}
func TestReductionSubsequence(t *testing.T) {
	t.Parallel()
	if !subsequence("//", "const x = 1;\n// comment") || subsequence("//", "* block *") || !subsequence("😀", "x😀y") {
		t.Fatal("cached witness is not a source deletion")
	}
}
func TestHostAdapterAnchors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"missing", "anchor anchor"} {
		if _, err := replaceOnce(source, "anchor", "replacement"); err == nil {
			t.Fatal("accepted missing or ambiguous source boundary")
		}
	}
	if got, err := replaceOnce("before anchor after", "anchor", "replacement"); err != nil || got != "before replacement after" {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestFailurePhaseAndOutputTrace(t *testing.T) {
	t.Parallel()
	trace := "scoreboard-phase\tparser\nscoreboard-log\t\"range 0 1 id \\t\\t\\t0 1\\n\"\n"
	a := WorkAnswer{Error: "worker timeout"}
	applyTrace(&a, trace, true)
	if a.Phase != "parser" || a.Output != "range 0 1 id \t\t\t0 1\n" || a.Stderr != "" {
		t.Fatalf("trace lost failure facts: %+v", a)
	}
	b := WorkAnswer{Error: "exception", Phase: "parser", Output: "already captured\n"}
	applyTrace(&b, "scoreboard-phase\trule\n"+trace, false)
	if b.Phase != "parser" || b.Output != "already captured\n" {
		t.Fatalf("cleanup rewrote captured failure: %+v", b)
	}
}

func TestByteTransportCannotNormalizeInvalidUTF8(t *testing.T) {
	t.Parallel()
	original := WorkAnswer{Output: string([]byte{'x', 0xff, '\n'})}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded WorkAnswer
	if err = json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Output != original.Output {
		t.Fatal("JSON transport silently replaced oracle stdout bytes")
	}
	mutant := WorkAnswer{Output: "x\ufffd\n"}
	if state, _ := stateCode(decoded, mutant, false); state != "diverge" {
		t.Fatal("UTF-8 replacement mutant survived")
	}
}

func TestStderrCollectorDoesNotRestoreEarlierRequest(t *testing.T) {
	t.Parallel()
	var collector lockedBuffer
	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() { _, err := io.Copy(&collector, reader); done <- err }()
	defer reader.Close()
	defer writer.Close()
	take := func(want string) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		var got string
		for time.Now().Before(deadline) {
			got += collector.take()
			if got != "" {
				break
			}
			time.Sleep(time.Millisecond)
		}
		if got != want {
			t.Fatalf("stderr crossed request boundaries: got %q, want %q", got, want)
		}
	}
	writer.Write([]byte("first\n"))
	take("first\n")
	writer.Write([]byte("second\n"))
	take("second\n")
	writer.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestCleanCopiesPreserveMeasuredBoundaries(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	out := filepath.Join(root, "clean")
	source := filepath.Join(root, "receipts-00-0000.jsonl.gz")
	original := Receipt{Index: 7, Path: "x.ts", SHA: "source-hash", GoFormat: WorkAnswer{Output: "ok\t\"stderr\":\\\\\"value\"\n", Exit: 0, NS: 123}, NodeFormat: WorkAnswer{Output: "ok\t\"stderr\":\\\\\"value\"\n", Stderr: "host trace with \\\\ and \"quote\"\n", Exit: 0, NS: 456}}
	f, err := os.Create(source)
	if err != nil {
		t.Fatal(err)
	}
	z := gzip.NewWriter(f)
	if err = json.NewEncoder(z).Encode(original); err != nil {
		t.Fatal(err)
	}
	z.Close()
	f.Close()
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("python3", "clean_receipts.py", root, out)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("cleanup: %s %v", result, err)
	}
	after, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("cleanup changed original archive")
	}
	f, err = os.Open(filepath.Join(out, filepath.Base(source)))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var cleaned Receipt
	if err = json.NewDecoder(reader).Decode(&cleaned); err != nil {
		t.Fatal(err)
	}
	original.NodeFormat.Stderr = ""
	if !reflect.DeepEqual(original, cleaned) {
		t.Fatalf("cleanup altered measured fields: %+v %+v", original, cleaned)
	}
}

func TestLiteralSkipTextCannotBlockARealAnswer(t *testing.T) {
	t.Parallel()
	a := WorkAnswer{Output: "fixed\tconst message = 'skipped typed no program';\\u000a\n"}
	if status, cause := stateCode(a, a, false); status != "agree" || cause != "" {
		t.Fatalf("literal content was mistaken for coverage metadata: %s %s", status, cause)
	}
}

func TestByteAuditSentinel(t *testing.T) {
	t.Parallel()
	if ambiguousBytes(WorkAnswer{Output: "clean"}) {
		t.Fatal("lossless stdout marked ambiguous")
	}
	if !ambiguousBytes(WorkAnswer{Output: "\ufffd"}) {
		t.Fatal("replacement mutant escaped byte audit")
	}
	if ambiguousBytes(WorkAnswer{Output: "\ufffd", OutputRaw: []byte{0xff}}) {
		t.Fatal("explicit raw bytes marked ambiguous")
	}
}
