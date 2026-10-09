package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"map", "object", "sequences"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/library_runtime_tuple_" + name + ".a", true, false})
	}
}

func TestLibraryRuntimeTuples(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, stdout string }{
		{"map", "first:1\nsecond:2\nfirst:1\nsecond:2\n"},
		{"object", "first:1\nsecond:2\nfirst:1\nsecond:2\n"},
		{"sequences", "7:7\n0:9\n0:1\n0:1\ntrue\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_runtime_tuple_"+test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(test.stdout)}, truth); difference != "" {
				t.Fatal("Node: " + difference)
			}
			t.Logf("Node: exit=%d stdout=%q", truth.exitCode, truth.stdout)
			sanitized, binary := natively(t, program)
			results := map[string]run{"native release": released(t, program), "native sanitized": sanitized, "native slabs": slabbed(t, program), "JavaScript": onJavaScriptBackend(t, program)}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				results["WASI"] = onWASI(t, native.C(program))
			}
			for backend, result := range results {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, result.exitCode, result.stdout, result.stderr)
				if difference := disagreement(truth, result); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatal(report)
			}
		})
	}
}

// Keep this slice's allocation receipts runnable within the worker's 90-second limit.
func TestLibraryRuntimeTupleCounts(t *testing.T) {
	t.Parallel()
	table, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"map", "object", "sequences"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := "internal/oracle/testdata/library_runtime_tuple_" + name + ".a"
			row := counted(t, path, false, nil, false, false)
			t.Log(row)
			if !strings.Contains(string(table), row+"\n") {
				t.Fatalf("allocation counts not recorded: %s", row)
			}
		})
	}
}
