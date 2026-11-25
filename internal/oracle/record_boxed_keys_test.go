package oracle

import (
	"path/filepath"
	"testing"
)

func TestRecordBoxedKeysAgainstNode(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"115-string-number-key", "p2-number-spellings", "p3-once", "p10-literal-arms",
		"extra-loop", "extra-returned-key", "extra-coalesce",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "record_boxed_keys", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node stopped: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := natively(t, program)
			for backend, got := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native} {
				if why := disagreement(truth, got); why != "" {
					t.Errorf("%s: %s: Node %+v, Adamic %+v", backend, why, truth, got)
				}
			}
			if leaked := leaks(t, program, binary); leaked != "" {
				t.Error(leaked)
			}
		})
	}
}

func TestRecordBoxedKeyPrototypeMiss(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"p4-prototype-miss", "p4-unboxed-twin"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "record_boxed_keys", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "function\n" {
				t.Fatalf("Node prototype baseline: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native} {
				const message = "adamic: panic: record member 'toString' is missing; records hold own keys only\n"
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
					t.Errorf("%s: expected the record prototype miss, got %+v", backend, got)
				}
			}
		})
	}
}

func TestRecordBoxedKeyNarrowedAwayStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, output string }{
		{"s01-narrowed-local", "via undefined\n"},
		{"s02-narrowed-field", "via undefined\n"},
		{"s03-narrowed-boolean", "via true\n"},
		{"s04-narrowed-null", "via null\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join("testdata", "record_boxed_keys", probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); disagreement(run{stdout: []byte(probe.output)}, node) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}
