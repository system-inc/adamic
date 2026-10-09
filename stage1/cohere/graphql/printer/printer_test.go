package printer

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	var sources []string
	// No .graphql documents exist in this checkout. Cases come from these
	// pinned Go test sources and the generator's explicit controls instead.
	corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit,
		[]string{"internal/format/graphql"}, []string{"*_test.go"})
	t.Log("GraphQL documents: no tracked .graphql corpus; pinned upstream test constants and generated controls")
	directory := t.TempDir()
	cases := filepath.Join(directory, "cases.txt")
	answers := filepath.Join(directory, "answers.txt")
	request, _ := json.Marshal(map[string]any{"Sources": sources, "Cases": cases, "Answers": answers, "Mode": mode, "Coverage": filepath.Join(directory, "coverage.json")})
	path := filepath.Join(directory, "request.json")
	if err = os.WriteFile(path, request, 0644); err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/cohere_side_test.go")
	generator, _ := filepath.Abs("../testdata/cohere_side_test.go")
	cohere := filepath.Join(root, "cohere")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/graphql/adamic_printer_test.go"): side, filepath.Join(cohere, "internal/format/graphql/adamic_generator_test.go"): generator}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err = os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	command := bounded(t, "go", "test", "-count=1", "-overlay="+overlayPath, "-run=^TestAdamicPrinter$", "./internal/format/graphql")
	if len(oracle) != 0 {
		command = bounded(t, oracle[0], "-test.run=^TestAdamicPrinter$", "-test.count=1", "-test.timeout=0")
	}
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PRINTER_REQUEST="+path)
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go oracle: %v\n%s", err, output)
	}
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
// The four fixed option modes run as shard-NNN and share their Go oracle build.
func TestPrinterUpstreamPreflight(t *testing.T) { printerUpstreamShards(t) }

func printerDirectory(t *testing.T, file, from, to string) string {
	t.Helper()
	directory := t.TempDir()
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

// Not parallel: the corpus keep path is shared with the option sweeps.
func TestPrinterMutants(t *testing.T) {
	cases, want := printerCases(t, "defaults")
	for _, mutation := range []struct{ name, file, from, to string }{
		{"line width ignored", "doc.ts", "this.settings.printWidth - column", "100000 - column"},
		{"end of line comment leads next node", "printer.ts", "if(own && around.following >= 0)", "if((own || end) && around.following >= 0)"},
		{"block string indentation discarded", "printer.ts", "for(const line of lines) parts.push(documents.text(line));", "for(const line of lines) parts.push(documents.text(line.trim()));"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			path := printerDirectory(t, mutation.file, mutation.from, mutation.to)
			program := lowered(t, path)
			for _, side := range []struct {
				name   string
				result run
			}{{"native", nativelyRun(t, program, "--cases", cases)}, {"Node", onNode(t, path, "--cases", cases)}} {
				if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
					t.Fatalf("%s mutant must run: exit %d, %s", side.name, side.result.exitCode, side.result.stderr)
				}
				difference := firstDifference(string(side.result.stdout), want)
				if difference == "" {
					t.Errorf("%s mutant escaped oracle", side.name)
				} else {
					t.Logf("%s caught: %s", side.name, difference)
				}
			}
		})
	}
}

// The same binary's ordinary file driver is held to the batch oracle too.
func TestPrinterFileDriver(t *testing.T) {
	t.Parallel()
	source := filepath.Join(t.TempDir(), "probe.graphql")
	if err := os.WriteFile(source, []byte("query{hello(a:1,b:2)}"), 0644); err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs("main.ts")
	program := lowered(t, path)
	want := "query {\n    hello(a: 1, b: 2)\n}\n"
	for _, side := range []run{onNode(t, path, source), nativelyRun(t, program, source)} {
		if side.exitCode != 0 || string(side.stdout) != want {
			t.Fatalf("driver: exit %d, stdout %q, stderr %q", side.exitCode, side.stdout, side.stderr)
		}
	}

}
