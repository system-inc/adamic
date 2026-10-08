package decoratorshelpers

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func run(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	f, err := os.CreateTemp(t.TempDir(), "out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	if err = childguard.Run(cmd, childguard.Options{}); err != nil {
		t.Fatalf("%s: %v: %s", name, err, &stderr)
	}
	if stderr.Len() != 0 {
		t.Fatalf("%s: %s", name, &stderr)
	}
	data, err := os.ReadFile(f.Name())
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
func corpus(t *testing.T, name string) (string, []byte, int) {
	t.Helper()
	f, e := os.Open("testdata/" + name + ".json.gz")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	data, e := io.ReadAll(z)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "cases.json")
	write(t, path, data)
	var x struct{ Calls []struct{ Want string } }
	if e = json.Unmarshal(data, &x); e != nil {
		t.Fatal(e)
	}
	var want strings.Builder
	for _, c := range x.Calls {
		want.WriteString(c.Want + "\n")
	}
	return path, []byte(want.String()), len(x.Calls)
}
func build(t *testing.T, directory string) (string, string) {
	t.Helper()
	p, e := load.Load([]string{filepath.Join(directory, "main.a")})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if e = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	script := filepath.Join(t.TempDir(), "emitted.mjs")
	write(t, script, []byte(javascript.JavaScript(ir)))
	return binary, script
}
func outputs(t *testing.T, directory, path, binary, script string) [][]byte {
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	return [][]byte{run(t, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.a"), path), run(t, "node", "--disable-warning=ExperimentalWarning", runner, script, path), run(t, binary, path)}
}
func TestGoAgreement(t *testing.T) {
	directory, _ := filepath.Abs(".")
	binary, script := build(t, directory)
	for _, name := range []string{"calls", "controls"} {
		t.Run(name, func(t *testing.T) {
			path, want, count := corpus(t, name)
			for i, out := range outputs(t, directory, path, binary, script) {
				if !bytes.Equal(out, want) {
					a, b := strings.Split(string(out), "\n"), strings.Split(string(want), "\n")
					for j := 0; j < len(a) && j < len(b); j++ {
						if a[j] != b[j] {
							t.Fatalf("backend %d call %d got %q Go %q", i, j, a[j], b[j])
						}
					}
					t.Fatalf("backend %d output size", i)
				}
				t.Logf("backend %d: %d Go observations", i, count)
			}
		})
	}
}
func TestCaptureCoverage(t *testing.T) {
	var coverage struct {
		GoPin          string
		GoSourceSha256 string
		Calls          struct {
			Calls   int
			Symbols map[string]int
		}
		Controls struct {
			Calls   int
			Symbols map[string]int
		}
	}
	b, e := os.ReadFile("testdata/coverage.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &coverage); e != nil {
		t.Fatal(e)
	}
	root, _ := filepath.Abs("../../../../../cohere")
	pin := strings.TrimSpace(string(run(t, "git", "-C", root, "rev-parse", "HEAD")))
	if pin != coverage.GoPin {
		t.Fatal("Go pin drift")
	}
	source, e := os.ReadFile(filepath.Join(root, "internal/lint/ecmascript/decorators/decorators.go"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(source)) != coverage.GoSourceSha256 {
		t.Fatal("Go source drift")
	}
	for _, name := range []string{"CallName", "HasDecoratorInSet", "Of"} {
		if coverage.Calls.Symbols[name] <= 0 || coverage.Controls.Symbols[name] <= 0 {
			t.Fatal("missing helper input", name)
		}
	}
	for _, entry := range []struct {
		Name  string
		Count int
	}{{"calls", coverage.Calls.Calls}, {"controls", coverage.Controls.Calls}} {
		_, _, count := corpus(t, entry.Name)
		if count != entry.Count {
			t.Fatal("capture count drift")
		}
	}
}
func TestMutants(t *testing.T) {
	mutants := []struct{ Name, File, From, To string }{{"call_name_identifier_inverted", "call_name.a", "callee.kind!=='Identifier'", "callee.kind==='Identifier'"}, {"of_decorator_inverted", "of.a", "child.kind==='Decorator'", "child.kind!=='Decorator'"}, {"membership_inverted", "has_decorator_in_set.a", "names.includes(callName(nodes,decorator))", "!names.includes(callName(nodes,decorator))"}}
	for _, m := range mutants {
		t.Run(m.Name, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "decorators")
			if e := os.MkdirAll(directory, 0755); e != nil {
				t.Fatal(e)
			}
			files, e := filepath.Glob("*.a")
			if e != nil {
				t.Fatal(e)
			}
			for _, file := range files {
				b, e := os.ReadFile(file)
				if e != nil {
					t.Fatal(e)
				}
				if file == m.File {
					if bytes.Count(b, []byte(m.From)) != 1 {
						t.Fatal("mutant anchor", m.Name)
					}
					b = bytes.Replace(b, []byte(m.From), []byte(m.To), 1)
				}
				write(t, filepath.Join(directory, file), b)
			}
			b, e := os.ReadFile("../options_json.ts")
			if e != nil {
				t.Fatal(e)
			}
			write(t, filepath.Join(filepath.Dir(directory), "options_json.ts"), b)
			binary, script := build(t, directory)
			caught := []bool{false, false, false}
			for _, name := range []string{"calls", "controls"} {
				path, want, _ := corpus(t, name)
				for i, out := range outputs(t, directory, path, binary, script) {
					if !bytes.Equal(out, want) {
						caught[i] = true
					}
				}
			}
			for i, yes := range caught {
				if !yes {
					t.Fatalf("mutant survived backend %d", i)
				}
				t.Logf("caught backend %d", i)
			}
		})
	}
}
