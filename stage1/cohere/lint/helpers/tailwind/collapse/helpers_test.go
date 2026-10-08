package collapse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func repository(t *testing.T) string {
	t.Helper()
	p, e := filepath.Abs("../../../../../..")
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func run(t *testing.T, dir string, env []string, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	c.Env = append(os.Environ(), env...)
	var out, errout bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errout
	if e := c.Run(); e != nil {
		t.Fatalf("%s %v: %v\n%s\n%s", name, args, e, out.String(), errout.String())
	}
	if errout.Len() != 0 {
		t.Fatalf("stderr: %s", errout.String())
	}
	return out.Bytes()
}
func write(t *testing.T, p string, b []byte) {
	t.Helper()
	if e := os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
}
func backends(t *testing.T, entry string) [][]string {
	t.Helper()
	entry, _ = filepath.Abs(entry)
	p, e := load.Load([]string{entry})
	if e != nil {
		t.Fatal(e)
	}
	ir, e := lower.Lower(context.Background(), p)
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "native")
	if e = native.Build(native.C(ir), bin, native.Options{Sanitize: true}); e != nil {
		t.Fatal(e)
	}
	js := filepath.Join(dir, "emitted.mjs")
	write(t, js, []byte(javascript.JavaScript(ir)))
	runner := filepath.Join(repository(t), "oracle/node.mjs")
	return [][]string{{"node", "--disable-warning=ExperimentalWarning", runner, entry}, {"node", "--disable-warning=ExperimentalWarning", runner, js}, {bin}}
}
func oracle(t *testing.T, stem string) string {
	t.Helper()
	root := filepath.Join(repository(t), "cohere")
	virtual := filepath.Join(root, "adamic_collapse_oracle.go")
	main, _ := filepath.Abs("testdata/" + stem + "_oracle.go")
	side, _ := filepath.Abs("testdata/" + stem + "_exports.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: main, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_helper_exports.go"): side}})
	dir := t.TempDir()
	path := filepath.Join(dir, "overlay.json")
	write(t, path, overlay)
	bin := filepath.Join(dir, "oracle")
	run(t, root, nil, "go", "build", "-overlay="+path, "-o", bin, virtual)
	return bin
}
func capturedActions(t *testing.T) (string, []byte, int) {
	t.Helper()
	root := filepath.Join(repository(t), "cohere")
	sourcePath := filepath.Join(root, "internal/lint/rules/tailwind/collapse/design_system.go")
	source, e := os.ReadFile(sourcePath)
	if e != nil {
		t.Fatal(e)
	}
	anchor := "return builds.count\n}"
	if strings.Count(string(source), anchor) != 2 {
		t.Fatal("counter capture anchor changed")
	}
	dir := t.TempDir()
	side := filepath.Join(dir, "design_system.go")
	write(t, side, []byte(strings.ReplaceAll(string(source), anchor, "adamicCounterCapture(builds.count)\n\treturn builds.count\n}")))
	capture := filepath.Join(dir, "capture.go")
	write(t, capture, []byte(`package tailwind
import("fmt";"os")
func adamicCounterCapture(count int) { f,e:=os.OpenFile(os.Getenv("ADAMIC_COLLAPSE_CAPTURE"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if e!=nil {panic(e)};defer f.Close();fmt.Fprintln(f,count) }
`))
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{sourcePath: side, filepath.Join(root, "internal/lint/rules/tailwind/collapse/adamic_counter_capture.go"): capture}})
	path := filepath.Join(dir, "overlay.json")
	write(t, path, overlay)
	calls := filepath.Join(dir, "calls.txt")
	run(t, root, []string{"ADAMIC_COLLAPSE_CAPTURE=" + calls}, "go", "test", "-overlay="+path, "./internal/lint/rules/tailwind", "-count=1", "-timeout=10m")
	want, e := os.ReadFile(calls)
	if e != nil {
		t.Fatal(e)
	}
	lines := strings.Split(strings.TrimSpace(string(want)), "\n")
	if len(lines) < 200 {
		t.Fatal("upstream counter capture unexpectedly small")
	}
	previous := 0
	var actions strings.Builder
	for _, line := range lines {
		value, e := strconv.Atoi(line)
		if e != nil {
			t.Fatal(e)
		}
		switch value {
		case previous:
			actions.WriteString("read\n")
		case previous + 1:
			actions.WriteString("build\n")
		default:
			t.Fatalf("counter capture lost an increment: %d -> %d", previous, value)
		}
		previous = value
	}
	input := filepath.Join(dir, "actions.txt")
	write(t, input, []byte(actions.String()))
	return input, want, len(lines)
}
func TestCounterGoAgreement(t *testing.T) {
	input, want, count := capturedActions(t)
	goOracle := oracle(t, "counter")
	if got := run(t, "", nil, goOracle, input); !bytes.Equal(got, want) {
		t.Fatal("capture replay disagrees with unchanged Go counter")
	}
	for i, c := range backends(t, "main.a") {
		got := run(t, "", nil, c[0], append(c[1:], input)...)
		if !bytes.Equal(got, want) {
			t.Fatalf("backend %d disagrees with Go", i)
		}
	}
	t.Logf("%d actual upstream counter calls match Go on Node, emitted JavaScript and sanitized native; decimal output byte-identical", count)
}
func TestCounterMutants(t *testing.T) {
	input := filepath.Join(t.TempDir(), "actions.txt")
	write(t, input, []byte("read\nbuild\nread\nbuild\nread\n"))
	want := run(t, "", nil, oracle(t, "counter"), input)
	for _, m := range []struct{ file, from, to string }{{"next_build_count.a", "buildCounter.value++;", "buildCounter.value += 2;"}, {"builds_so_far.a", "return buildCounter.value;", "return buildCounter.value + 1;"}} {
		t.Run(m.file, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"main.a", "counter_state.a", "next_build_count.a", "builds_so_far.a"} {
				data, e := os.ReadFile(name)
				if e != nil {
					t.Fatal(e)
				}
				if name == m.file {
					if strings.Count(string(data), m.from) != 1 {
						t.Fatal("mutant anchor changed")
					}
					data = []byte(strings.Replace(string(data), m.from, m.to, 1))
				}
				write(t, filepath.Join(dir, name), data)
			}
			for i, c := range backends(t, filepath.Join(dir, "main.a")) {
				got := run(t, "", nil, c[0], append(c[1:], input)...)
				if bytes.Equal(got, want) {
					t.Fatal("compiled mutant survived")
				}
				t.Logf("compiled semantic mutant %s caught on %s by Go byte comparison", m.file, []string{"Node", "emitted JavaScript", "sanitized native"}[i])
			}
		})
	}
}
func TestStylesheetByteInputBoundary(t *testing.T) {
	goOracle := oracle(t, "file")
	commands := backends(t, "testdata/file-input.a")
	for _, sample := range []struct {
		name  string
		value []byte
		gap   bool
	}{{"ASCII", []byte("A"), false}, {"Unicode", []byte("é😀"), false}, {"invalid80", []byte{0x80}, true}, {"invalid81", []byte{0x81}, true}, {"invalidFF", []byte{0xff}, true}} {
		t.Run(sample.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "theme.css")
			css := append([]byte("@theme { --color-probe: "), sample.value...)
			css = append(css, []byte("; }")...)
			write(t, file, css)
			want := run(t, "", nil, goOracle, file)
			var expected strings.Builder
			for _, b := range sample.value {
				fmt.Fprintf(&expected, "%d\n", b)
			}
			if string(want) != expected.String() {
				t.Fatal("Go loader did not preserve original bytes")
			}
			for i, c := range commands {
				got := run(t, "", nil, c[0], append(c[1:], file)...)
				if sample.gap {
					if bytes.Equal(got, want) {
						t.Fatal("known input gap unexpectedly closed")
					}
					t.Logf("BLOCKED stylesheetCollector.loadFile on %s: Go bytes %q; readTextFile bytes %q (clean exit)", []string{"Node", "emitted JavaScript", "sanitized native"}[i], strings.TrimSpace(string(want)), strings.TrimSpace(string(got)))
				} else if !bytes.Equal(got, want) {
					t.Fatal("valid UTF-8 control differs")
				}
			}
		})
	}
}
