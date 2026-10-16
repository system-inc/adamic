package printer

import (
	"github.com/system-inc/adamic/stage1/cohere/graphql/printer/oracles"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// printerCases reads the owner package's identical corpus and Go answers.
func printerCases(t *testing.T, mode string) (string, string) { return oracles.Cases(t, mode) }

const testPrinterUpstreamPreflightShards = 4

// ADAMIC_TEST_SHARD=i/n selects indices modulo n equal to i; unset runs all.
// The four fixed option modes run as top-level TestPrinterUpstreamPreflight_NNN tests.

func printerDirectory(t *testing.T, file, from, to string) string {
	t.Helper()
	return printerDirectoryAt(t, t.TempDir(), file, from, to)
}

func printerDirectoryAt(t *testing.T, directory, file, from, to string) string {
	t.Helper()
	// The port's sources, laid out under directory as they sit under stage1/cohere (printerPortFiles); a mutation
	// changes one of the printer's own.
	for _, name := range printerPortFiles {
		data, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if file != "" && name == "graphql/printer/"+file {
			if strings.Count(string(data), from) != 1 {
				t.Fatalf("mutant must match once: %s", file)
			}
			data = []byte(strings.Replace(string(data), from, to, 1))
		}
		target := filepath.Join(directory, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(directory, "graphql/printer/main.ts")
}

// The same binary's ordinary file driver is held to the batch oracle too.
func TestPrinterFileDriver(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "probe.graphql")
	if err := os.WriteFile(source, []byte("query{hello(a:1,b:2)}"), 0644); err != nil {
		t.Fatal(err)
	}
	products := printerInTreeProducts(t)
	want := "query {\n    hello(a: 1, b: 2)\n}\n"
	for _, side := range []run{onNode(t, products.source, source), execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, source)} {
		if side.exitCode != 0 || string(side.stdout) != want {
			t.Fatalf("driver: exit %d, stdout %q, stderr %q", side.exitCode, side.stdout, side.stderr)
		}
	}

}
