// Package oracles owns cssstrings Go reference products without linking the compiler.
// Both test binaries resolve paths from the parent package directory.
package oracles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/toolchain"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const repository = "../../.."

var oracleProducer bool

func EnableProductProduction() { oracleProducer = true }

type Corpus struct{ Texts, Paths, Raw []string }

func Enumerate(t *testing.T) Corpus { c := enumerateStrings(t); return Corpus{c.texts, c.paths, c.raw} }

type AnswerCase struct {
	File  string
	Input []byte
}

func BatchRanges(count int) [][2]int {
	var ranges [][2]int
	for lo := 0; lo < count; lo += 512 {
		ranges = append(ranges, [2]int{lo, min(lo+512, count)})
	}
	return ranges
}
func BatchCases(corpus Corpus) []AnswerCase {
	var cases []AnswerCase
	for _, r := range BatchRanges(len(corpus.Texts)) {
		cases = append(cases, AnswerCase{fmt.Sprintf("batch-%06d-%06d.txt", r[0], r[1]-1), Input(corpus.Texts[r[0]:r[1]])})
	}
	return cases
}
func RawCases(corpus Corpus) []AnswerCase {
	var cases []AnswerCase
	for i, text := range corpus.Raw {
		for _, p := range []string{"d", "s"} {
			cases = append(cases, AnswerCase{fmt.Sprintf("raw-%06d-%s.txt", i, p), []byte(p + stringsEncode.Replace(text) + "\n")})
		}
	}
	return cases
}
func AnswerProductInputs(name, identity, tool string, files []string, cases []AnswerCase) Inputs {
	hash := sha256.New()
	for _, c := range cases {
		fmt.Fprintf(hash, "%s %d\n", c.File, len(c.Input))
		hash.Write(c.Input)
	}
	if name == "Go batch oracle answers" || name == "Go raw oracle answers" {
		name += " oracle-owner-v1"
		files = append(append([]string{}, files...), "oracles/oracles.go")
	}
	return Inputs{Name: name, Files: append([]string{"setup_products_test.go"}, files...), Flags: []string{identity, fmt.Sprintf("cases-sha256=%x", hash.Sum(nil))}, Toolchain: tool}
}
func runOracle(t *testing.T, binary string, args []string, input []byte, path string, bounded bool) ([]byte, error) {
	if err := os.WriteFile(path, input, 0644); err != nil {
		return nil, err
	}
	defer os.Remove(path)
	var ctx context.Context
	var cancel context.CancelFunc
	if bounded {
		ctx, cancel = context.WithTimeout(context.Background(), 90*time.Second)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()
	command := toolchain.CommandContextIn(ctx, "", binary, append(append([]string{}, args...), path)...)
	command.Env = nil
	var stdout, stderr []byte
	var err error
	if bounded {
		// One case reads through childguard.Capture, to its end with no timer deciding when: exec's copying under
		// a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the
		// oracle exited, so a case that had finished could fail its product under load (#n3gk22h's shape). Its 90 s
		// deadline ends a hung case before childguard.Run's windows (two minutes of silence, thirty before the first
		// byte) could.
		stdout, stderr, err = childguard.Capture(ctx, command)
	} else {
		// The whole corpus keeps childguard.Run's two-minute stall guard, shorter than the 600 s Loom gives a
		// product unit, since the oracle prints a line per case. With no WaitDelay exec reads both streams to their
		// end, and the oracle starts no process that could hold them open after it exits.
		var output, diagnostics bytes.Buffer
		command.Stdout, command.Stderr = &output, &diagnostics
		err = childguard.Run(command, childguard.Options{})
		stdout, stderr = output.Bytes(), diagnostics.Bytes()
	}
	if err != nil || len(stderr) != 0 {
		return nil, fmt.Errorf("oracle on %s: %v; %s", filepath.Base(path), err, stderr)
	}
	return stdout, nil
}
func FullAnswers(t *testing.T, binary string, args []string, data []byte, identity string) []byte {
	t.Helper()
	in := CacheInputs(t, Inputs{Name: "Go oracle answers oracle-owner-v1", Files: []string{"testdata/library.mjs", "oracles/oracles.go"}, Flags: []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))}, Toolchain: buildcache.Tool("node", "--version")})
	dir := buildcache.Product(t, in, func(dir string) error {
		if !oracleProducer {
			return fmt.Errorf("settle the Go full-answer owner first")
		}
		out, err := runOracle(t, binary, args, data, filepath.Join(dir, "input.txt"), false)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), out, 0644)
	})
	out, err := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

type answerProduct struct {
	once      sync.Once
	directory string
	err       error
}

var batchAnswers, rawAnswers answerProduct

func answersDirectory(t *testing.T, state *answerProduct, name, binary, identity string, cases []AnswerCase) string {
	t.Helper()
	state.once.Do(func() {
		state.err = fmt.Errorf("answer product: %s stopped before its lookup; its log has the failure", t.Name())
		in := CacheInputs(t, AnswerProductInputs(name, identity, runtime.Version(), nil, cases))
		state.directory, state.err = buildcache.Get(in, func(dir string) error {
			if !oracleProducer {
				return fmt.Errorf("settle the Go answer owner first")
			}
			pending := make(chan AnswerCase)
			failures := make(chan error, len(cases))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for c := range pending {
						out, err := runOracle(t, binary, nil, c.Input, filepath.Join(dir, c.File+".input"), true)
						if err == nil {
							err = os.WriteFile(filepath.Join(dir, c.File), out, 0644)
						}
						if err != nil {
							failures <- err
						}
					}
				}()
			}
			for _, c := range cases {
				pending <- c
			}
			close(pending)
			workers.Wait()
			close(failures)
			return <-failures
		})
		if state.err != nil {
			state.err = fmt.Errorf("%s: %w", in.Name, state.err)
		}
	})
	if state.err != nil {
		t.Fatal(state.err)
	}
	return state.directory
}
func BatchAnswers(t *testing.T, binary, identity string, cases []AnswerCase) string {
	return answersDirectory(t, &batchAnswers, "Go batch oracle answers", binary, identity, cases)
}
func RawAnswers(t *testing.T, binary, identity string, cases []AnswerCase) string {
	return answersDirectory(t, &rawAnswers, "Go raw oracle answers", binary, identity, cases)
}

type stringsCorpus struct{ texts, paths, raw []string }

func enumerateStrings(t *testing.T) stringsCorpus {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	paths := corpusfiles.Upstream(t, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{"internal/format/css/testdata/prettier", "internal/lint/rules/tailwind"}, []string{"*.css", "*.scss", "*.less"})
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		texts = append(texts, string(data))
	}
	files := len(texts)
	alphabet := []string{"a", "'", "\"", "\\", "\n", "😀"}
	var generate func(string, int)
	generate = func(s string, n int) {
		texts = append(texts, s)
		if n > 0 {
			for _, c := range alphabet {
				generate(s+c, n-1)
			}
		}
	}
	generate("", 5)
	texts = append(texts, "\r\t\u2028\u2029", strings.Repeat("'😀\\\"x' ", 10000), "'never closed\\", `"\'"`, "\x00'null'")
	raw := append(append([]string{}, texts[:files]...), "", "a", "a\n", "a\r\n\n", "'😀'", "\"\\'\"")
	return stringsCorpus{texts: texts, paths: paths, raw: raw}
}

func Printer(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs("testdata/go_driver.go")
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_stage_one/main.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/css/adamic_stage_one.go"): bridge,
		mainPath: driver,
	}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	const output = "go-printer-oracle-owner-v1"
	args := []string{"-overlay=" + overlayPath}
	if oracleProducer {
		return buildcache.GoBuildIn(t, "cohere", output, mainPath, args)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, mainPath, args, nil)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, output, inputs, "stage1/cohere/cssstrings/oracles TestProduct_CSSStringsGoOracle", func(string) error { return fmt.Errorf("settle the printer owner first") })
	return filepath.Join(directory, output)
}

func Input(texts []string) []byte {
	var input strings.Builder
	for _, text := range texts {
		for _, preference := range []string{"d", "s"} {
			input.WriteString(preference + stringsEncode.Replace(text) + "\n")
		}
	}
	return []byte(input.String())
}

var stringsEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func CacheInputs(t *testing.T, inputs Inputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	keyed := buildcache.Inputs{Name: "cssstrings " + inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, buildcache.Tool("go", "version"), runtime.GOOS + "/" + runtime.GOARCH}}
	keyed.Files = append(keyed.Files, "stage1/cohere/cssstrings/shards_test.go", "stage1/cohere/cssstrings/port_test.go")
	keyed.Flags = append(keyed.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))

	for _, name := range inputs.Files {
		absolute, err := filepath.Abs(name)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			logical := filepath.Base(name)
			if filepath.Ext(name) == ".ts" {
				logical = filepath.Base(filepath.Dir(name)) + "/" + logical
			}
			keyed.Flags = append(keyed.Flags, fmt.Sprintf("generated-input=%s:%x", logical, sha256.Sum256(content)))
		} else {
			keyed.Files = append(keyed.Files, filepath.ToSlash(relative))
		}
	}
	// A lowering build runs the compiler linked into this test binary. Its key names the compiler by its sources
	// (compilerinputs.Lowering, the Go release) rather than by this binary's bytes, which differ wherever the test binary
	// was built, so Workshop's product and a runner's ask are one key (#nm31pcn).

	return keyed
}

type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

func Identity(t *testing.T, path string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(path, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("oracle-content-sha256=%x", hash.Sum(nil))
}
