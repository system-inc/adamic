package oracle

import (
	"path/filepath"
	"testing"
)

// Swapping realpath key order or dropping its entry reds realpath/sanitized, realpath/release and realpath/counted.
var jsonErrorAgrees = []string{"plain", "library_type", "watch", "realpath", "process_code", "path_code"}
var jsonErrorWrites = []string{"direct", "view", "logical"}

const jsonErrorRoot = "internal/oracle/testdata/json_errors/"
const jsonErrorWritten = "adamic: panic: NotYet: JSON.stringify an Error whose name or message was written after construction\n"

func init() {
	for _, name := range jsonErrorAgrees {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{jsonErrorRoot + name + ".a", true, false})
	}
	// Runtime stops stay outside the ordinary differential fixtures, but their counts are measured too.
	for _, name := range jsonErrorWrites {
		path := jsonErrorRoot + name + ".a"
		countFixtureSources[path] = path
		additionalFixtureCounts = append(additionalFixtureCounts, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, false, nil, false, false) }})
	}
}

// Each complete allocation is held to Node on every existing backend (#z4a9p9e).
func TestJSONErrorsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range jsonErrorAgrees {
		for backend := 0; backend < 5; backend++ {
			t.Run(name+"/"+[]string{"sanitized", "release", "javascript", "wasm32", "counted"}[backend], func(t *testing.T) { t.Parallel(); fixtureAgreesOnBackend(t, jsonErrorRoot+name+".a", backend) })
		}
	}
}

// Native stops after a program write; JavaScript must print Node's newly enumerable name.
// Skipping the sentinel check reds all three native rows; dropping the logical byte write reds logical only.
func TestJSONErrorsWrittenAfterConstruction(t *testing.T) {
	t.Parallel()
	for _, name := range jsonErrorWrites {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, jsonErrorRoot+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			want := run{stdout: []byte("{\"name\":\"Changed\"}\n")}
			if difference := disagreement(want, truth); difference != "" {
				t.Fatalf("Node: %s: exit %d stdout %q stderr %q", difference, truth.exitCode, truth.stdout, truth.stderr)
			}
			t.Run("javascript", func(t *testing.T) {
				got := onJavaScriptBackend(t, program)
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			})
			for _, backend := range []string{"sanitized", "release"} {
				t.Run(backend, func(t *testing.T) {
					var got run
					if backend == "sanitized" {
						got, _ = natively(t, program)
					} else {
						got = released(t, program)
					}
					want := run{stderr: []byte(jsonErrorWritten), exitCode: 70}
					if difference := disagreement(want, got); difference != "" {
						t.Fatalf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
					}
				})
			}
		})
	}
}
