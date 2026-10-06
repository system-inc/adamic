package lint

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

func batch8Oracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/batch8_oracle.go")
	virtual := filepath.Join(root, "adamic_batch8_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	directory := t.TempDir()
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "oracle")
	execute(t, root, "go", "build", "-overlay="+path, "-o", binary, virtual)
	return binary
}
func batch8Node(t *testing.T, directory, path string, count bool) execution {
	t.Helper()
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path}
	if count {
		args = append(args, "--count")
	}
	return execute(t, "", "node", args...)
}
func batch8Build(t *testing.T, directory string, sanitize bool) string {
	t.Helper()
	start := time.Now()
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		t.Fatal(err)
	}
	loaded := time.Now()
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	lowering := time.Now()
	source := native.C(lowered)
	emitted := time.Now()
	binary := filepath.Join(t.TempDir(), "lint")
	if err := native.Build(source, binary, native.Options{Sanitize: sanitize}); err != nil {
		t.Fatal(err)
	}
	t.Logf("build load=%s lower=%s emit=%s clang=%s C_bytes=%d", loaded.Sub(start), lowering.Sub(loaded), emitted.Sub(lowering), time.Since(emitted), len(source))
	return binary
}
func batch8Compare(t *testing.T, oracle, binary, directory string, rows []string) []byte {
	t.Helper()
	path := manifest(t, rows)
	goAnswer := execute(t, "", oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{{"Node", batch8Node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
		if diff := difference(side.run.output, goAnswer.output); diff != "" {
			t.Fatalf("%s: %s", side.name, diff)
		}
		t.Logf("%s execution %s", side.name, side.run.duration)
	}
	t.Logf("Go, Node, sanitized native identical: %d bytes, %d cases; Go execution %s", len(goAnswer.output), len(rows), goAnswer.duration)
	return goAnswer.output
}
func batch8Spans(t *testing.T, oracle, path string) string {
	t.Helper()
	result := execute(t, "", oracle, "--jsx-spans", path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var spans []string
	for _, line := range strings.Split(strings.TrimSpace(string(result.output)), "\n") {
		if line == "empty" {
			return "empty"
		}
		if line == "" {
			continue
		}
		var start, end int
		if _, err := fmt.Sscanf(line, "%d:%d", &start, &end); err != nil {
			t.Fatal(err)
		}
		spans = append(spans, fmt.Sprintf("%d:%d", len(utf16.Encode([]rune(string(data[:start])))), len(utf16.Encode([]rune(string(data[:end]))))))
	}
	return strings.Join(spans, ",")
}
func batch8Sources(t *testing.T) []string {
	t.Helper()
	compiler := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if compiler == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout for source-corpus parity")
	}
	pin := execute(t, "", "git", "-C", compiler, "rev-parse", "HEAD")
	if strings.TrimSpace(string(pin.output)) != compilerCommit {
		t.Fatal("wrong compiler pin")
	}
	var rows []string
	for _, root := range []string{filepath.Join(compiler, "src/compiler"), filepath.Join(repository, "stage1")} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				rows = append(rows, absolute+"\tall")
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return rows
}
func batch8Key(t *testing.T, row string) string {
	t.Helper()
	fields := strings.Split(row, "\t")
	data, err := os.ReadFile(fields[0])
	if err != nil {
		t.Fatal(err)
	}
	return fields[1] + " " + fmt.Sprintf("%x", sha256.Sum256(data))
}

// Not parallel: capture creates the shared fixture scratch directory, and measurements require idle builds.
func TestBatch8(t *testing.T) {
	directory := batch8Snapshot(t)
	oracle := batch8Oracle(t)
	rows := batch8Upstream(t)
	gapsData, err := os.ReadFile("batch8_evidence/parser_gaps.json")
	if err != nil {
		t.Fatal(err)
	}
	var gaps map[string]string
	if err := json.Unmarshal(gapsData, &gaps); err != nil {
		t.Fatal(err)
	}
	var accepted []string
	adapter, blocked := 0, 0
	seenGaps := map[string]bool{}
	for _, row := range rows {
		fields := strings.Split(row, "\t")
		path, rule := fields[0], fields[1]
		spans := batch8Spans(t, oracle, path)
		if reason, known := gaps[batch8Key(t, row)]; known {
			blocked++
			seenGaps[batch8Key(t, row)] = true
			if reason == "JSX parser missing" && spans == "" {
				t.Fatalf("recorded JSX gap lost its Go JSX nodes: %s", path)
			}
			t.Logf("unsupported source %s %s: %s", rule, batch8Key(t, row), reason)
			continue
		}
		if spans != "" && rule != "react/jsx-no-comment-textnodes" {
			t.Fatalf("unrecorded JSX gap %s", batch8Key(t, row))
		}
		if spans != "" {
			adapter++
			accepted = append(accepted, path+"\t"+rule+"\t"+spans)
		} else {
			accepted = append(accepted, path+"\t"+rule)
		}
	}
	if len(seenGaps) != len(gaps) {
		t.Fatalf("gap ledger has %d keys but %d replayed", len(gaps), blocked)
	}
	t.Logf("upstream %d: %d source cases, %d Go-AST JSX listener cases, %d unsupported source cases", len(rows), len(accepted)-adapter, adapter, blocked)
	binary := batch8Build(t, directory, true)
	t.Run("jsx_gap", func(t *testing.T) { batch8JsxGap(t, oracle, binary, directory) })
	t.Run("upstream", func(t *testing.T) { batch8Compare(t, oracle, binary, directory, accepted) })
	t.Run("compiler_and_stage1", func(t *testing.T) { batch8Compare(t, oracle, binary, directory, batch8Sources(t)) })
	t.Run("mutants", func(t *testing.T) { batch8Mutants(t, oracle, directory, binary) })
	if os.Getenv("ADAMIC_LINT_BENCH") == "1" {
		t.Run("throughput", func(t *testing.T) { batch8Throughput(t, oracle, directory) })
	}
}

func batch8JsxGap(t *testing.T, oracle, binary, directory string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jsx-gap.tsx")
	source := "const view = <div>// note</div>;\n"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	goAnswer := execute(t, "", oracle, path)
	if !strings.Contains(string(goAnswer.output), "putCommentInBraces") {
		t.Fatal("Go gap control lacks its finding")
	}
	runner, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	var first []byte
	for _, side := range []struct {
		name string
		args []string
	}{{"node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path}}, {binary, []string{path}}} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		command := exec.CommandContext(ctx, side.name, side.args...)
		var stderr strings.Builder
		output, err := os.CreateTemp(t.TempDir(), "gap-stdout-")
		if err != nil {
			t.Fatal(err)
		}
		command.Stdout = output
		command.Stderr = &stderr
		err = command.Run()
		cancel()
		output.Close()
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(stderr.String(), "parser slice unsupported primary") {
			t.Fatalf("JSX parser gap changed on %s: %v %s", side.name, err, stderr.String())
		}
		if first == nil {
			first = []byte(stderr.String())
		} else if string(first) != stderr.String() {
			t.Fatal("Node/native parser refusal differs")
		}
		t.Logf("Go reports putCommentInBraces; %s refuses: %s", side.name, strings.TrimSpace(stderr.String()))
	}
}

func batch8Snapshot(t *testing.T) string {
	t.Helper()
	original, err := filepath.Abs("batch8")
	if err != nil {
		t.Fatal(err)
	}
	typescript, err := filepath.Abs("../../typescript")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	entries, err := os.ReadDir(original)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".ts") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(original, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		source := strings.ReplaceAll(string(data), "../../../typescript", typescript)
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}
