package core

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

type call struct {
	Helper string `json:"helper"`
	Frame  string `json:"frame"`
	Args   []int  `json:"args"`
	Extra  any    `json:"extra"`
	Want   string `json:"want"`
}
type capture struct {
	Pin    string                     `json:"pin"`
	Frames map[string]json.RawMessage `json:"frames"`
	Calls  []call                     `json:"calls"`
}

func command(t *testing.T, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	f, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	cmd.Stdout = f
	if err = cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr: %s", stderr.String())
	}
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func corpus(t *testing.T, helper string) (string, []byte, int) {
	t.Helper()
	f, err := os.Open("testdata/captures.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal(err)
	}
	var c capture
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	if c.Pin != "7945d102a6c18dd36adf9114a758ce646e8b2359" {
		t.Fatal("capture pin differs")
	}
	pin := command(t, "git", "-C", "../../../../../cohere", "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin)) != c.Pin {
		t.Fatal("Go pin differs")
	}
	keys := []string{}
	for k := range c.Frames {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ids := map[string]int{}
	frames := []json.RawMessage{}
	for _, k := range keys {
		ids[k] = len(frames)
		frames = append(frames, c.Frames[k])
	}
	calls := []map[string]any{}
	var want strings.Builder
	for _, row := range c.Calls {
		if helper != "" && helper != row.Helper {
			continue
		}
		calls = append(calls, map[string]any{"helper": row.Helper, "frame": ids[row.Frame], "args": row.Args, "extra": row.Extra})
		want.WriteString(row.Want)
		want.WriteByte('\n')
	}
	if len(calls) == 0 {
		t.Fatal("empty capture set", helper)
	}
	input, err := json.Marshal(map[string]any{"frames": frames, "calls": calls})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "input.json")
	if err = os.WriteFile(path, input, 0644); err != nil {
		t.Fatal(err)
	}
	return path, []byte(want.String()), len(calls)
}
func compile(t *testing.T, entry string) (string, string) {
	t.Helper()
	program, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	ir, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "core")
	if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(t.TempDir(), "core.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); err != nil {
		t.Fatal(err)
	}
	return binary, js
}
func same(t *testing.T, got, want []byte) {
	t.Helper()
	if bytes.Equal(got, want) {
		return
	}
	a, b := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			t.Fatalf("line %d: got %q Go %q", i+1, a[i], b[i])
		}
	}
	t.Fatalf("bytes got %d Go %d", len(got), len(want))
}
func TestCapturedCalls(t *testing.T) {
	path, want, count := corpus(t, "")
	entry, _ := filepath.Abs("main.a")
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	same(t, command(t, "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	binary, js := compile(t, entry)
	same(t, command(t, "node", "--disable-warning=ExperimentalWarning", runner, js, path), want)
	same(t, command(t, binary, path), want)
	t.Logf("%d actual Go helper calls byte-identical on Node, emitted JavaScript, ASan/UBSan native", count)
}
func TestMutants(t *testing.T) {
	var changes []struct{ Helper, File, From, To string }
	b, err := os.ReadFile("testdata/mutants.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &changes); err != nil {
		t.Fatal(err)
	}
	if len(changes) != 17 {
		t.Fatal("one mutant per independent helper")
	}
	for _, change := range changes {
		t.Run(change.Helper, func(t *testing.T) {
			path, want, count := corpus(t, change.Helper)
			directory := t.TempDir()
			entries, err := os.ReadDir(".")
			if err != nil {
				t.Fatal(err)
			}
			// Keep imports relative to the repository; only owned modules are copied beside a synthetic entry.
			root, _ := filepath.Abs(".")
			for _, e := range entries {
				if !strings.HasSuffix(e.Name(), ".a") {
					continue
				}
				source, err := os.ReadFile(e.Name())
				if err != nil {
					t.Fatal(err)
				}
				s := string(source)
				if e.Name() == change.File {
					if strings.Count(s, change.From) != 1 {
						t.Fatal("mutant anchor not unique")
					}
					s = strings.Replace(s, change.From, change.To, 1)
				}
				// External relative imports resolve from their original owned module.
				for _, prefix := range []string{"../options_json.ts", "../../../../typescript/scanner/scanner.ts", "../../../../typescript/parser/parser.ts"} {
					s = strings.ReplaceAll(s, "'"+prefix+"'", "'"+filepath.Clean(filepath.Join(root, prefix))+"'")
				}
				if err = os.WriteFile(filepath.Join(directory, e.Name()), []byte(s), 0644); err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(directory, "main.a")
			runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
			node := command(t, "node", "--disable-warning=ExperimentalWarning", runner, entry, path)
			if bytes.Equal(node, want) {
				t.Fatal("mutant survived Node")
			}
			binary, js := compile(t, entry)
			nativeOutput := command(t, binary, path)
			if bytes.Equal(nativeOutput, want) {
				t.Fatal("mutant survived native")
			}
			same(t, nativeOutput, node)
			same(t, command(t, "node", "--disable-warning=ExperimentalWarning", runner, js, path), node)
			t.Logf("%s caught by output disagreement on Node, emitted JavaScript and sanitized native (%d actual calls)", change.Helper, count)
		})
	}
}
