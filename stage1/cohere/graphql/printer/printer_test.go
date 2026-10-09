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
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

func printerCases(t *testing.T, mode string, oracle ...string) (string, string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// Mode and mutant shards use this fixed corpus: Cohere commit
	// 7945d102a6c18dd36adf9114a758ce646e8b2359 plus generated controls
	// from cohere_side_test.go, whose random seed is 20261006. Compare
	// live case totals in the union; guard the pinned file count separately.
	files := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit,
		[]string{"internal/format/graphql"}, []string{"*_test.go"})
	if len(files) != 3 {
		t.Fatalf("pinned GraphQL corpus: %d files, want 3", len(files))
	}
	t.Log("GraphQL documents: no tracked .graphql corpus; pinned upstream test constants and generated controls")
	if len(oracle) == 0 {
		oracle = []string{printerOracle(t)}
	}
	directory := printerBuild(t, printerBuildInputs{
		Name:  "GraphQL printer oracle outputs",
		Files: []string{oracle[0], root + "/stage1/cohere/graphql/printer/printer_test.go", root + "/stage1/cohere/graphql/printer/shards_test.go"},
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
		command := exec.CommandContext(context.Background(), oracle[0], "-test.run=^TestAdamicPrinter$", "-test.count=1", "-test.timeout=0")
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
	for _, group := range []struct {
		source, target string
		files          []string
	}{
		{"..", "graphql", []string{"token.ts", "characterClasses.ts", "blockString.ts", "lexer.ts", "parser.ts"}},
		{"../../json", "json", []string{"width.ts", "widthTables.ts"}},
		{".", "graphql/printer", []string{"doc.ts", "printer.ts", "main.ts"}},
	} {
		target := filepath.Join(directory, group.target)
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range group.files {
			data, err := os.ReadFile(filepath.Join(group.source, name))
			if err != nil {
				t.Fatal(err)
			}
			if group.source == "." && file == name {
				if strings.Count(string(data), from) != 1 {
					t.Fatalf("mutant must match once: %s", file)
				}
				data = []byte(strings.Replace(string(data), from, to, 1))
			}
			if err = os.WriteFile(filepath.Join(target, name), data, 0644); err != nil {
				t.Fatal(err)
			}
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
	path, _ := filepath.Abs("main.ts")
	products := preparePrinterProducts(t, path)
	want := "query {\n    hello(a: 1, b: 2)\n}\n"
	for _, side := range []run{onNode(t, path, source), execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, source)} {
		if side.exitCode != 0 || string(side.stdout) != want {
			t.Fatalf("driver: exit %d, stdout %q, stderr %q", side.exitCode, side.stdout, side.stderr)
		}
	}

}
