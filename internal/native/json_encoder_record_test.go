package native

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Main's record harness supplies a borrowed ordered slot view with no heap shape.
// Compile a renamed copy of production C so the mutant never edits the checkout.
func jsonEncoderRecordView(test *testing.T, wasm bool) {
	test.Helper()
	options := Options{Sanitize: !wasm, Count: true}
	if wasm {
		if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
			test.Skip("set ADAMIC_ORACLE_WASI=1")
		}
		options = Options{Target: "wasm32-wasi", Count: true}
	}
	encoder, err := os.ReadFile("runtime/json_stringify.c")
	if err != nil {
		test.Fatal(err)
	}
	source := strings.ReplaceAll(string(encoder), "adamic_json_stringify", "probe_adamic_json_stringify")
	// Keep the real header name; rename only the exported entry points.
	source = strings.ReplaceAll(source, "probe_adamic_json_stringify.h", "json_stringify.h")
	harness := strings.ReplaceAll(recordHarness(test), "adamic_json_stringify(", "probe_adamic_json_stringify(")
	marker := "size_t slots = object->shape == NULL ? schema->count : object->shape->count;"
	if strings.Count(source, marker) != 1 {
		test.Fatal("record view mutation anchor missing")
	}
	for _, mutant := range []bool{false, true} {
		code := source
		if mutant {
			code = strings.Replace(
				code,
				marker,
				"if (object->shape == NULL) { ascii(w, \"{}\"); return true; }\n"+marker,
				1,
			)
		}
		binary := filepath.Join(test.TempDir(), "record-view")
		if err := Build("#define _POSIX_C_SOURCE 200809L\n"+code+"\n"+harness, binary, options); err != nil {
			test.Fatal(err)
		}
		modes := []string{"semantics", "two-indices", "prototypes", "iteration"}
		if mutant {
			modes = modes[:2]
		}
		for _, mode := range modes {
			var stdout, stderr string
			var err error
			if wasm {
				stdout, stderr, err = recordNodeRun(
					"--disable-warning=ExperimentalWarning",
					"../../oracle/wasi.mjs",
					binary,
					mode,
				)
			} else {
				stdout, stderr, err = recordRun(binary, mode)
			}
			if err != nil {
				test.Fatalf("%s mutant=%v must finish cleanly: %v\n%s", mode, mutant, err, stderr)
			}
			recordCheckCounts(test, stderr)
			want := recordNode(test, mode)
			if !mutant && stdout != want {
				test.Fatalf("%s: record view differs from Node at byte %d", mode, recordDifference(stdout, want))
			}
			if mutant && stdout == want {
				test.Fatalf("%s: Node did not catch dropped record path", mode)
			}
		}
		if mutant {
			test.Log(
				"dropped record path: clean build, exit, sanitizers and balanced counts; caught only by Node output",
			)
		}
	}
}

func TestJSONEncoderRecordView_000(test *testing.T) {
	test.Parallel()
	jsonEncoderRecordView(test, false)
}
func TestJSONEncoderRecordView_001(test *testing.T) {
	test.Parallel()
	jsonEncoderRecordView(test, true)
}
