package comments

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
	"strings"
	"testing"
	"unicode/utf16"
)

func leadingCapture(t *testing.T) (string, []byte) {
	t.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	file := filepath.Join(root, "internal/lint/ecmascript/comments/comments.go")
	source, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	changed := string(source)
	for _, name := range []string{"LeadingRunFor", "isAdjacentGap"} {
		anchor := "func " + name + "("
		if strings.Count(changed, anchor) != 1 {
			t.Fatal("Go instrument anchor drift", name)
		}
		changed = strings.Replace(changed, anchor, "func adamicOriginal"+strings.ToUpper(name[:1])+name[1:]+"(", 1)
	}
	dir := t.TempDir()
	side := filepath.Join(dir, "comments.go")
	if err = os.WriteFile(side, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	capture, _ := filepath.Abs("testdata/leading/capture.go.txt")
	controls, _ := filepath.Abs("testdata/leading/controls_test.go.txt")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{file: side, filepath.Join(filepath.Dir(file), "adamic_capture.go"): capture, filepath.Join(filepath.Dir(file), "adamic_capture_test.go"): controls}})
	overlayPath := filepath.Join(dir, "overlay.json")
	os.WriteFile(overlayPath, overlay, 0600)
	captures := []json.RawMessage{}
	counts := []int{}
	for _, group := range []struct{ pkg, pattern string }{{"rules/structure", "^TestReactHookRequireEffectComment"}, {"ecmascript/comments", "^Test(LeadingRunFor|AdamicAdjacentControls)"}} {
		path := filepath.Join(dir, fmt.Sprint(len(counts))+".jsonl")
		cmd := exec.Command("go", "test", "-overlay="+overlayPath, "./internal/lint/"+group.pkg, "-run", group.pattern, "-count=1", "-v", "-timeout=10m")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "ADAMIC_LEADING_CAPTURE="+path)
		data, err := cmd.CombinedOutput()
		if err != nil || bytes.Contains(data, []byte("--- SKIP:")) {
			t.Fatalf("capture %s: %v\n%s", group.pkg, err, data)
		}
		rows, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		for _, row := range bytes.Split(rows, []byte("\n")) {
			if len(row) > 0 {
				captures = append(captures, append(json.RawMessage{}, row...))
				count++
			}
		}
		if count == 0 {
			t.Fatal("zero captures", group.pkg)
		}
		counts = append(counts, count)
	}
	var want bytes.Buffer
	leading, gaps := 0, 0
	for _, raw := range captures {
		var row struct {
			Mode string
			Want json.RawMessage
		}
		if err = json.Unmarshal(raw, &row); err != nil {
			t.Fatal(err)
		}
		if row.Mode == "gap" {
			var value bool
			json.Unmarshal(row.Want, &value)
			number := 0
			if value {
				number = 1
			}
			fmt.Fprintf(&want, "gap %d\n", number)
			gaps++
			continue
		}
		var comments []struct {
			Start, End, Line, Column, EndLine int
			Block                             bool
			Text                              string
		}
		if err = json.Unmarshal(row.Want, &comments); err != nil {
			t.Fatal(err)
		}
		var ranges []string
		for _, c := range comments {
			block := 0
			if c.Block {
				block = 1
			}
			ranges = append(ranges, fmt.Sprintf("%d:%d:%d:%d:%d:%d:%s", c.Start, c.End, block, c.Line, c.Column, c.EndLine, leadingWritten(c.Text)))
		}
		fmt.Fprintf(&want, "leading %s\n", strings.Join(ranges, ","))
		leading++
	}
	data, _ := json.Marshal(captures)
	path := filepath.Join(dir, "cases.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("actual Go captures: %d consumer invocations, %d package/control invocations; %d LeadingRunFor + %d isAdjacentGap", counts[0], counts[1], leading, gaps)
	return path, want.Bytes()
}
func leadingOutputs(t *testing.T, entry, path string) []struct {
	name   string
	output []byte
} {
	t.Helper()
	entry, _ = filepath.Abs(entry)
	program, err := load.Load([]string{entry})
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
	if err = os.WriteFile(script, []byte(javascript.JavaScript(ir)), 0600); err != nil {
		t.Fatal(err)
	}
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	return []struct {
		name   string
		output []byte
	}{{"Node", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path)}, {"emitted JavaScript", run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, script, path)}, {"sanitized native", run(t, "", binary, path)}}
}
func TestLeadingHelpersConsumerCapture(t *testing.T) {
	path, want := leadingCapture(t)
	for _, side := range leadingOutputs(t, "testdata/leading/main.a", path) {
		compare(t, side.output, want)
		t.Logf("%s identical: %d Go observations", side.name, bytes.Count(want, []byte("\n")))
	}
}
func TestLeadingHelpersMutants(t *testing.T) {
	path, want := leadingCapture(t)
	root, _ := filepath.Abs("../../../../..")
	directory, _ := filepath.Abs(".")
	for _, m := range []struct{ file, from, to string }{{"leading_run_for.a", "previous.isBlock || current.isBlock", "false"}, {"is_adjacent_gap.a", "return newlines === 1;", "return newlines === 0;"}} {
		t.Run(m.file, func(t *testing.T) {
			dest := t.TempDir()
			err := filepath.WalkDir(directory, func(path string, item os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if item.IsDir() {
					return nil
				}
				if !strings.HasSuffix(path, ".a") && !strings.HasSuffix(path, ".ts") {
					return nil
				}
				relative, _ := filepath.Rel(directory, path)
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				s := string(data)
				if relative == m.file {
					if strings.Count(s, m.from) != 1 {
						t.Fatal("mutant anchor changed")
					}
					s = strings.Replace(s, m.from, m.to, 1)
				}
				s = strings.ReplaceAll(s, "../../../../../../typescript", filepath.ToSlash(filepath.Join(root, "stage1/typescript")))
				s = strings.ReplaceAll(s, "../../../../typescript", filepath.ToSlash(filepath.Join(root, "stage1/typescript")))
				s = strings.ReplaceAll(s, "../../../options_json.ts", filepath.ToSlash(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")))
				s = strings.ReplaceAll(s, "../options_json.ts", filepath.ToSlash(filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")))
				target := filepath.Join(dest, relative)
				if err = os.MkdirAll(filepath.Dir(target), 0700); err != nil {
					return err
				}
				return os.WriteFile(target, []byte(s), 0600)
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, side := range leadingOutputs(t, filepath.Join(dest, "testdata/leading/main.a"), path) {
				if bytes.Equal(side.output, want) {
					t.Fatalf("mutant survived on %s", side.name)
				}
				t.Logf("running output mutant caught on %s", side.name)
			}
		})
	}
}

func leadingWritten(text string) string {
	var result strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			result.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&result, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&result, `\u%04x\u%04x`, h, l)
		}
	}
	return result.String()
}
