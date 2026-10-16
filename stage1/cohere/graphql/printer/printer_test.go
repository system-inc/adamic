package printer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

// printerCorpora holds each mode's corpus once per process: the as-Go, mutant, preflight and whitespace tests all read
// the same Go oracle's answers, so a process runs the git checks, hashes the oracle and looks the product up once per
// mode rather than once per caller (about 30 calls in a package run).
var printerCorpora sync.Map

type printerCorpus struct {
	once       sync.Once
	path, want string
}

var printerUpstream struct {
	once    sync.Once
	checked bool
}

// printerCases is one mode's corpus: the cases file and the Go oracle's answers to it.
func printerCases(t *testing.T, mode string) (string, string) {
	t.Helper()
	value, _ := printerCorpora.LoadOrStore(mode, &printerCorpus{})
	corpus := value.(*printerCorpus)
	corpus.once.Do(func() { corpus.path, corpus.want = printerCorpusProduct(t, mode) })
	// A preparation that failed inside the Once leaves the path empty for every later caller in this process.
	if corpus.path == "" {
		t.Fatalf("GraphQL %s corpus unavailable: its preparation failed earlier in this process (see that test's log)", mode)
	}
	return corpus.path, corpus.want
}

func printerCorpusProduct(t *testing.T, mode string) (string, string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// Mode and mutant shards use this fixed corpus: Cohere commit
	// 7945d102a6c18dd36adf9114a758ce646e8b2359 plus generated controls
	// from cohere_side_test.go, whose random seed is 20261006. Compare
	// live case totals in the union; guard the pinned file count separately.
	// The pin is the same for every mode, so a process checks it once.
	printerUpstream.once.Do(func() {
		files := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit,
			[]string{"internal/format/graphql"}, []string{"*_test.go"})
		if len(files) != 3 {
			t.Fatalf("pinned GraphQL corpus: %d files, want 3", len(files))
		}
		t.Log("GraphQL documents: no tracked .graphql corpus; pinned upstream test constants and generated controls")
		printerUpstream.checked = true
	})
	if !printerUpstream.checked {
		t.Fatal("pinned GraphQL corpus: its check failed earlier in this process (see that test's log)")
	}
	oracle := printerOracle(t)
	directory := printerBuild(t, printerBuildInputs{
		Name:  "GraphQL printer oracle outputs",
		Files: []string{oracle, root + "/stage1/cohere/graphql/printer/printer_test.go", root + "/stage1/cohere/graphql/printer/shards_test.go"},
		Flags: []string{"mode=" + mode, "test-run=TestAdamicPrinter"}, Toolchain: runtime.Version(),
	}, func(directory string) error {
		request, err := json.Marshal(map[string]any{"Sources": []string(nil), "Cases": directory + "/cases.txt", "Answers": directory + "/answers.txt", "Mode": mode, "Coverage": directory + "/coverage.json"})
		if err != nil {
			return err
		}
		path := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(path, request, 0644); err != nil {
			return err
		}
		command := exec.CommandContext(context.Background(), oracle, "-test.run=^TestAdamicPrinter$", "-test.count=1", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), "ADAMIC_PRINTER_REQUEST="+path)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	cases, answers := directory+"/cases.txt", directory+"/answers.txt"
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_GRAPHQL_PRINTER_KEEP"); keep != "" {
		for _, name := range []string{"cases.txt", "answers.txt", "coverage.json"} {
			data, err := os.ReadFile(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(keep, name), data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return cases, string(data)
}

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
