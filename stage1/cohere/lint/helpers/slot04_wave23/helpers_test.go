package stringhelpers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: these comparisons compile whole profiles and hold large output streams.
func run(t *testing.T, directory, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = directory
	log, err := os.CreateTemp(t.TempDir(), "output-")
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd.Stdout = log
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, &stderr)
	}
	data, err := os.ReadFile(log.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func write(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
func oracle(t *testing.T) string {
	root, _ := filepath.Abs("../../../../../cohere")
	here, _ := filepath.Abs("testdata")
	scratch := t.TempDir()
	script := filepath.Join(here, "overlay.py")
	run(t, "", "python3", script, root, here, scratch)
	binary := filepath.Join(t.TempDir(), "go-oracle")
	run(t, root, "go", "build", "-overlay="+filepath.Join(scratch, "overlay.json"), "-o", binary, filepath.Join(root, "adamic_slot04_wave23_oracle.go"))
	return binary
}
func compare(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q, Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("output size %d != %d", len(got), len(want))
}

type artifacts struct{ native, script string }

func build(t *testing.T, directory string) artifacts { return buildEntry(t, directory, "main.a") }
func buildEntry(t *testing.T, directory, entry string) artifacts {
	t.Helper()
	program, err := load.Load([]string{filepath.Join(directory, entry)})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(t.TempDir(), "emitted.mjs")
	write(t, script, []byte(javascript.JavaScript(ir)))
	return artifacts{binary, script}
}
func observations(t *testing.T, directory, path string) []struct {
	name   string
	output []byte
} {
	t.Helper()
	built := build(t, directory)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	return []struct {
		name   string
		output []byte
	}{
		{"Node source", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.a"), path)},
		{"sanitized native", run(t, "", built.native, path)},
		{"emitted JavaScript", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, path)},
	}
}

func TestActualGoComparisons(t *testing.T) {
	binary := oracle(t)
	directory, _ := filepath.Abs(".")
	built := build(t, directory)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	for _, fixture := range []string{"witnesses.json", "consumers.json"} {
		data, e := os.ReadFile("testdata/" + fixture)
		if e != nil {
			t.Fatal(e)
		}
		var rows []json.RawMessage
		if e = json.Unmarshal(data, &rows); e != nil {
			t.Fatal(e)
		}
		for start := 0; start < len(rows); start += 20 {
			end := start + 20
			if end > len(rows) {
				end = len(rows)
			}
			raw, e := json.Marshal(rows[start:end])
			if e != nil {
				t.Fatal(e)
			}
			path := filepath.Join(t.TempDir(), "input.json")
			write(t, path, raw)
			adapted := filepath.Join(t.TempDir(), "cases.json")
			write(t, adapted, run(t, "", binary, "--cases", path))
			want := run(t, "", binary, path)
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.a"), adapted), want)
			compare(t, run(t, "", built.native, adapted), want)
			compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, built.script, adapted), want)
			t.Logf("%s rows %d:%d: Go, source Node, sanitized native, emitted JavaScript match %d observations", fixture, start, end, bytes.Count(want, []byte("\n")))
		}
	}
}
func TestCompilingMutants(t *testing.T) {
	binary := oracle(t)
	path, _ := filepath.Abs("testdata/witnesses.json")
	want := run(t, "", binary, path)
	adapted := filepath.Join(t.TempDir(), "cases.json")
	write(t, adapted, run(t, "", binary, "--cases", path))
	mutations := []struct{ file, old, new string }{
		{"build.a", "entry.incoming=true;", "entry.incoming=false;"},
		{"build.a", "entry.reachable=true;", "entry.reachable=false;"},
		{"build.a", "if(endReachable){ops.markFinal(builder,current);}", "ops.markFinal(builder,current);"},
		{"build.a", "ops.expression(builder,root.type);", "ops.expression(builder,root.body);"},
		{"build.a", "if(root.bodyBlock)", "if(!root.bodyBlock)"},
		{"build.a", "if(root.sourceFile)", "if(false)"},
		{"build.a", "else if(root.staticBlock)", "else if(false)"},
		{"build.a", "else if(root.property)", "else if(false)"},
	}
	for _, m := range mutations {
		t.Run(m.file+m.old, func(t *testing.T) {
			root := t.TempDir()
			directory := filepath.Join(root, "slot04_wave23")
			if e := os.Mkdir(directory, 0755); e != nil {
				t.Fatal(e)
			}
			for _, name := range []string{"main.a", "build.a"} {
				b, e := os.ReadFile(name)
				if e != nil {
					t.Fatal(e)
				}
				if name == m.file {
					if strings.Count(string(b), m.old) != 1 {
						t.Fatal("anchor")
					}
					b = []byte(strings.Replace(string(b), m.old, m.new, 1))
				}
				write(t, filepath.Join(directory, name), b)
			}
			for _, name := range []string{"options_json.ts"} {
				b, e := os.ReadFile("../" + name)
				if e != nil {
					t.Fatal(e)
				}
				target := filepath.Join(root, name)
				if e = os.MkdirAll(filepath.Dir(target), 0755); e != nil {
					t.Fatal(e)
				}
				write(t, target, b)
			}
			for _, o := range observations(t, directory, adapted) {
				if bytes.Equal(o.output, want) {
					t.Fatalf("%s mutant survived", o.name)
				}
				t.Logf("%s compiling semantic mutant caught by actual-Go comparison", o.name)
			}
		})
	}
}
func missingConsumers(data []byte) ([]string, error) {
	var rows []struct{ Name string }
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	present := map[string]bool{}
	for _, row := range rows {
		name, _, _ := strings.Cut(row.Name, ":")
		present[name] = true
	}
	ledger, err := os.ReadFile("../readiness.json")
	if err != nil {
		return nil, err
	}
	var d struct {
		Remaining []struct {
			Rule    string
			Helpers []string `json:"remaining_helpers"`
		}
	}
	if err = json.Unmarshal(ledger, &d); err != nil {
		return nil, err
	}
	missing := []string{}
	for _, row := range d.Remaining {
		for _, h := range row.Helpers {
			if strings.HasSuffix(h, "control_flow_graph.Build") {
				if !present[row.Rule] {
					missing = append(missing, row.Rule)
				}
				break
			}
		}
	}
	return missing, nil
}
func TestConsumerCoverage(t *testing.T) {
	data, err := os.ReadFile("testdata/consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	missing, err := missingConsumers(data)
	if err != nil || len(missing) > 0 {
		t.Fatalf("missing %v: %v", missing, err)
	}
	var rows []struct{ Name, Source string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, row := range rows {
		name, _, _ := strings.Cut(row.Name, ":")
		names[name] = true
	}
	for name := range names {
		filtered := []struct{ Name, Source string }{}
		for _, row := range rows {
			if !strings.HasPrefix(row.Name, name+":") {
				filtered = append(filtered, row)
			}
		}
		omitted, _ := json.Marshal(filtered)
		missing, err := missingConsumers(omitted)
		if err != nil || len(missing) == 0 {
			t.Fatalf("consumer omission survived: %s: %v", name, err)
		}
		t.Logf("consumer omission caught: %s", name)
	}
}
