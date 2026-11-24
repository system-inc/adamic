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
			native, _ := natively(t, program)
			for backend, got := range map[string]run{"javascript": onJavaScriptBackend(t, program), "native": native} {
				if why := disagreement(truth, got); why != "" {
					t.Errorf("%s: %s: Node %+v, Adamic %+v", backend, why, truth, got)
				}
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
