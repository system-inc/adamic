package printer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/corpusfiles"
)

func printerCases(t *testing.T, mode string) (string, string) {
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
	command.Dir = cohere
	command.Env = append(os.Environ(), "ADAMIC_PRINTER_REQUEST="+path)
	if output, err := command.CombinedOutput(); err != nil {
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

// Not parallel: the corpus keep path is shared by the option runs.
func TestPrinterUpstreamPreflight(t *testing.T) {
	directory := os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if directory == "" {
		t.Skip("set ADAMIC_GRAPHQL_PRETTIER to an npm install of prettier@3.9.6 and graphql@17.0.2; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	script, err := filepath.Abs("testdata/prettier.mjs")
	if err != nil {
		t.Fatal(err)
	}
	embedded, err := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		t.Run(mode, func(t *testing.T) {
			cases, want := printerCases(t, mode)
			for _, side := range []struct{ name, directory, engine string }{{"npm Prettier", directory, "npm"}, {"embedded fork", embedded, "embedded"}} {
				t.Run(side.name, func(t *testing.T) {
					got := execute(t, nil, "node", script, side.directory, cases, mode, side.engine)
					if got.exitCode != 0 || len(got.stderr) > 0 {
						t.Fatalf("Prettier: exit %d, %s", got.exitCode, got.stderr)
					}
					inputs, err := os.ReadFile(cases)
					if err != nil {
						t.Fatal(err)
					}
					sources := strings.Split(string(inputs), "\n")
					a, b := strings.Split(string(got.stdout), "\n"), strings.Split(want, "\n")
					if len(a) != len(b) {
						t.Fatalf("answer count: %d vs %d", len(a), len(b))
					}
					known, unexpected, accepted, refused := 0, 0, 0, 0
					exactKnown := map[string]string{
						">":       "error\tSyntax Error: Unexpected <EOF>. (1:1)",
						"> ":      "error\tSyntax Error: Unexpected <EOF>. (1:2)",
						">\\n\\r": "error\tSyntax Error: Unexpected <EOF>. (3:1)",
						">\\r":    "error\tSyntax Error: Unexpected <EOF>. (2:1)",
					}
					for i := 0; i < len(a)-1; i++ {
						if a[i] == b[i] {
							accepted++
							continue
						}
						if strings.HasPrefix(a[i], "error\t") && strings.HasPrefix(b[i], "error\t") {
							refused++
							continue
						}
						if expected, exists := exactKnown[sources[i]]; exists && a[i] == "ok\t" && b[i] == expected {
							known++
							t.Logf("known upstream difference case %d: source %q, Go %q, Prettier %q", i, sources[i], b[i], a[i])
							continue
						}
						unexpected++
						if unexpected <= 5 {
							t.Logf("unexpected case %d: Prettier %q; Go %q", i, a[i], b[i])
						}
					}
					t.Logf("%d texts: %d byte-identical formatted, %d shared refusals, %d known whitespace differences, %d unexpected differences", len(a)-1, accepted, refused, known, unexpected)
					if unexpected != 0 {
						t.Errorf("Go cohere differs from Prettier on %d unexpected texts", unexpected)
					}
					if known != 5 {
						t.Errorf("known difference count changed: %d, recorded 5 in GAPS.md", known)
					}
				})
			}
		})
	}
}

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

// Not parallel: all option sweeps can write ADAMIC_GRAPHQL_PRINTER_KEEP.
func TestPrinterAsGoCohere(t *testing.T) {
	path := printerDirectory(t, "", "", "")
	program := lowered(t, path)
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		t.Run(mode, func(t *testing.T) {
			cases, want := printerCases(t, mode)
			nodeRun := onNode(t, path, "--cases", cases, mode)
			nativeRun, binary := natively(t, program, "--cases", cases, mode)
			backendRun := onJavaScriptBackend(t, program, "--cases", cases, mode)
			for _, side := range []struct {
				name   string
				result run
			}{{"native", nativeRun}, {"Node", nodeRun}, {"JS backend", backendRun}} {
				if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
					t.Fatalf("%s: exit %d, %s", side.name, side.result.exitCode, side.result.stderr)
				}
				if difference := firstDifference(string(side.result.stdout), want); difference != "" {
					t.Errorf("%s: %s", side.name, difference)
				}
			}
			if report := leaks(t, program, binary, "--cases", cases, mode); report != "" {
				t.Errorf("leaks: %s", report)
			}
			t.Logf("%d texts: %d formatted, %d refused", strings.Count(want, "\n"), strings.Count(want, "ok\t"), strings.Count(want, "error\t"))
		})
	}
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
