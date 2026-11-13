package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Swapping realpath key order or dropping its entry reds realpath/sanitized, realpath/release and realpath/counted.
// Stopping on an ordinary message byte reds the message write rows.
var jsonErrorAgrees = []string{"plain", "library_type", "watch", "realpath", "process_code", "path_code", "message_direct", "message_view", "message_logical", "message_coded"}
var jsonErrorWrites = []string{"direct", "view", "logical", "message_absent"}

const jsonErrorRoot = "internal/oracle/testdata/json_errors/"
const jsonErrorWritten = "adamic: panic: " + ir.JSONErrorWriteStop + "\n"

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

// Both backends stop on a newly enumerable name or absent message, after Node's first bytes.
// Dropping the JavaScript check reds its four rows. Dropping either native check reds its rows:
// name reds direct, view and logical; #message reds message_absent.
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
			text := "{\"name\":\"Changed\"}\n"
			if name == "message_absent" {
				text = "{\"message\":\"y\"}\n"
			}
			if why := disagreement(run{stdout: []byte(text)}, truth); why != "" {
				t.Fatalf("Node: %s: exit %d stdout %q stderr %q", why, truth.exitCode, truth.stdout, truth.stderr)
			}
			for _, backend := range []string{"sanitized", "release", "javascript", "counted", "wasm32"} {
				t.Run(backend, func(t *testing.T) {
					var got run
					switch backend {
					case "sanitized":
						got, _ = natively(t, program)
					case "release":
						got = released(t, program)
					case "javascript":
						got = onJavaScriptBackend(t, program)
					case "counted":
						binary := filepath.Join(t.TempDir(), "counted")
						if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
							t.Fatal(err)
						}
						got = execute(t, binary)
						// The counter's diagnostic footer is instrumentation, not the program's stderr.
						if !countsLine.Match(got.stderr) {
							t.Fatalf("counted stop has no counts: %q", got.stderr)
						}
						got.stderr = countsLine.ReplaceAll(got.stderr, nil)
					case "wasm32":
						if os.Getenv("ADAMIC_TEST_WASI") != "1" || os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
							t.Skip("wasm oracle is not enabled")
						}
						got = onWASI(t, native.C(program))
					}
					want := run{stderr: []byte(jsonErrorWritten), exitCode: 70}
					if why := disagreement(want, got); why != "" {
						t.Fatalf("%s: exit %d stdout %q stderr %q", why, got.exitCode, got.stdout, got.stderr)
					}
					if !bytes.HasPrefix(truth.stdout, got.stdout) || len(truth.stdout) == len(got.stdout) {
						t.Fatal("stop stdout is not Node's strict first bytes")
					}
				})
			}
		})
	}
}
