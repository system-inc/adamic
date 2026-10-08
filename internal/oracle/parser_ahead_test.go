package oracle

import (
	"path/filepath"
	"testing"
)

// These pins describe real train boundaries. Their success is not parser acceptance.
func TestParserAheadFactoryBoundaries(t *testing.T) {
	for _, row := range []struct{ name, stdout, message string }{
		{"complete", "ready\n", "field read failed: <write>.text is not initialized; expected string, found missing"},
		{"escaped", "undefined\n", "field read failed: node.text is not initialized; expected string, found missing"},
	} {
		t.Run(row.name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories", row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != row.stdout || len(truth.stderr) != 0 {
				t.Fatalf("source Node %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + row.message + "\n")}
			actual, _ := nativelyUncached(t, program)
			for backend, got := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s", backend, difference)
				}
			}
			t.Log("pending: staged field allocation and writes; this is a boundary pin, not native acceptance")
		})
	}
}

// This existing complete-layout path is a control, not staged-construction credit.
func TestParserAheadFactoryCompleteControl(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/parser-ahead/factories/already-complete.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\n" || len(truth.stderr) != 0 {
		t.Fatalf("source Node %+v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, binary := nativelyUncached(t, program)
	for _, got := range []run{actual, onJavaScriptBackend(t, program)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatal(d)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	changed := 0
	for i, s := range program.Strings {
		if s == "ready" {
			program.Strings[i] = "wrong"
			changed++
		}
	}
	if changed != 1 {
		t.Fatalf("payload mutant changed %d constants", changed)
	}
	mutant, _ := nativelyUncached(t, program)
	for _, got := range []run{mutant, onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || string(got.stdout) != "wrong\n" {
			t.Fatalf("mutant must execute normally: %+v", got)
		}
		if disagreement(truth, got) == "" {
			t.Fatal("factory payload mutant survived")
		}
	}
	t.Log("factory payload mutant caught in native and JavaScript by source Node stdout")
}
