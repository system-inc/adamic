// Package oracles owns cssnumbers Go reference products without linking the compiler.
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

func Enumerate(test *testing.T) Corpus {
	c := enumerateNumbers(test)
	return Corpus{c.texts, c.paths, c.raw}
}

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
		cases = append(
			cases,
			AnswerCase{fmt.Sprintf("batch-%06d-%06d.txt", r[0], r[1]-1), Input(corpus.Texts[r[0]:r[1]])},
		)
	}
	return cases
}
func RawCases(corpus Corpus) []AnswerCase {
	var cases []AnswerCase
	for i, text := range corpus.Raw {
		for _, p := range []string{"d", "s"} {
			cases = append(
				cases,
				AnswerCase{fmt.Sprintf("raw-%06d-%s.txt", i, p), []byte(p + numbersEncode.Replace(text) + "\n")},
			)
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
	return Inputs{
		Name:      name,
		Files:     append([]string{"setup_products_test.go"}, files...),
		Flags:     []string{identity, fmt.Sprintf("cases-sha256=%x", hash.Sum(nil))},
		Toolchain: tool,
	}
}
func runOracle(test *testing.T, binary string, args []string, input []byte, path string, bounded bool) ([]byte, error) {
	if writeError := os.WriteFile(path, input, 0644); writeError != nil {
		return nil, writeError
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
func FullAnswers(test *testing.T, binary string, args []string, data []byte, identity string) []byte {
	test.Helper()
	in := CacheInputs(test, Inputs{
		Name:      "Go oracle answers oracle-owner-v1",
		Files:     []string{"testdata/library.mjs", "oracles/oracles.go"},
		Flags:     []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))},
		Toolchain: buildcache.Tool("node", "--version"),
	})
	dir := buildcache.Product(test, in, func(dir string) error {
		if !oracleProducer {
			return fmt.Errorf("settle the Go full-answer owner first")
		}
		out, runOracleError := runOracle(test, binary, args, data, filepath.Join(dir, "input.txt"), false)
		if runOracleError != nil {
			return runOracleError
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), out, 0644)
	})
	out, readError := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	return out
}

type answerProduct struct {
	once      sync.Once
	directory string
	err       error
}

var batchAnswers, rawAnswers answerProduct

func answersDirectory(test *testing.T, state *answerProduct, name, binary, identity string, cases []AnswerCase) string {
	test.Helper()
	state.once.Do(func() {
		state.err = fmt.Errorf("answer product: %s stopped before its lookup; its log has the failure", test.Name())
		in := CacheInputs(test, AnswerProductInputs(name, identity, runtime.Version(), nil, cases))
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
						out, err := runOracle(test, binary, nil, c.Input, filepath.Join(dir, c.File+".input"), true)
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
		test.Fatal(state.err)
	}
	return state.directory
}
func BatchAnswers(test *testing.T, binary, identity string, cases []AnswerCase) string {
	return answersDirectory(test, &batchAnswers, "Go batch oracle answers", binary, identity, cases)
}
func RawAnswers(test *testing.T, binary, identity string, cases []AnswerCase) string {
	return answersDirectory(test, &rawAnswers, "Go raw oracle answers", binary, identity, cases)
}

type numbersCorpus struct{ texts, paths, raw []string }

func enumerateNumbers(test *testing.T) numbersCorpus {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var texts []string
	paths := corpusfiles.Upstream(
		test,
		filepath.Join(root, "cohere"),
		corpusfiles.CohereCommit,
		[]string{"internal/format/css/testdata/prettier", "internal/lint/rules/tailwind"},
		[]string{"*.css", "*.scss", "*.less"},
	)
	for _, path := range paths {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		texts = append(texts, string(data))
	}
	files := len(texts)
	alphabet := []string{"a", "0", "1", ".", "e", "-", "\"", "\\"}
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
	texts = append(
		texts,
		"\r\t\u2028\u2029",
		strings.Repeat("'😀\\\"x' ", 10000),
		"'never closed\\",
		`"\'"`,
		"\x00'null'",
	)
	for _, mantissa := range []string{"0", "01", ".0", ".00100", "1.", "1.000", "12.34000"} {
		for _, exponent := range []string{"", "e0", "E+000", "e-000", "e+001", "e-002", "E00020", "e+", "e-"} {
			for _, unit := range strings.Split("|n|EM|Q|HZ|kHZ|px|cQmAX|unknown|foo|é|Σ|İ", "|") {
				for _, prefix := range []string{"", "a", "$x", "@foo", "-", "+", "😀", "'", "\""} {
					texts = append(texts, prefix+mantissa+exponent+unit)
				}
			}
		}
	}
	for _, unit := range strings.Split(
		"em|rem|ex|rex|cap|rcap|ch|rch|ic|ric|lh|rlh|vw|svw|lvw|dvw|vh|svh|lvh|dvh|vi|svi|lvi|dvi|vb|svb|lvb|dvb|vmin|svmin|lvmin|dvmin|vmax|svmax|lvmax|dvmax|cm|mm|Q|in|pt|pc|px|deg|grad|rad|turn|s|ms|Hz|kHz|dpi|dpcm|dppx|x|cqw|cqh|cqi|cqb|cqmin|cqmax|fr",
		"|",
	) {
		texts = append(texts, ".1000E+002"+strings.ToUpper(unit))
	}
	texts = append(texts, cssAuditNumberInputs...)
	texts = append(
		texts,
		"'1.000px'",
		"1.000unknown",
		"a1.000px",
		"😀1.000px",
		strings.Repeat(".000100E-002KHZ ", 10000),
	)
	raw := append(append([]string{}, texts[:files]...), "", "a", "a\n", "a\r\n\n", "'😀'", "\"\\'\"")
	return numbersCorpus{texts: texts, paths: paths, raw: raw}
}

var cssAuditNumberInputs = []string{"\u00801.000px"}

func Printer(test *testing.T) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	bridge, _ := filepath.Abs("testdata/bridge.go")
	driver, _ := filepath.Abs("testdata/go_driver.go")
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_stage_one/main.go")
	overlay, encodeError := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/css/adamic_stage_one.go"): bridge,
		mainPath: driver,
	}})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	overlayPath := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "go-printer-oracle-owner-v1"
	args := []string{"-overlay=" + overlayPath}
	if oracleProducer {
		return buildcache.GoBuildIn(test, "cohere", output, mainPath, args)
	}
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("cohere", output, mainPath, args, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		output,
		inputs,
		"stage1/cohere/cssnumbers/oracles TestProduct_CSSNumbersGoOracle",
		func(string) error { return fmt.Errorf("settle the printer owner first") },
	)
	return filepath.Join(directory, output)
}

func Input(texts []string) []byte {
	var input strings.Builder
	for _, text := range texts {
		for _, preference := range []string{"d", "s"} {
			input.WriteString(preference + numbersEncode.Replace(text) + "\n")
		}
	}
	return []byte(input.String())
}

var numbersEncode = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func CacheInputs(test *testing.T, inputs Inputs) buildcache.Inputs {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	keyed := buildcache.Inputs{
		Name:      "cssnumbers " + inputs.Name,
		Flags:     append([]string{}, inputs.Flags...),
		Toolchain: []string{inputs.Toolchain, buildcache.Tool("go", "version"), runtime.GOOS + "/" + runtime.GOARCH},
	}
	keyed.Files = append(
		keyed.Files,
		"stage1/cohere/cssnumbers/shards_test.go",
		"stage1/cohere/cssnumbers/port_test.go",
	)
	keyed.Flags = append(keyed.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))

	for _, name := range inputs.Files {
		absolute, absolutePathError := filepath.Abs(name)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		relative, relativePathError := filepath.Rel(root, absolute)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			content, readError := os.ReadFile(name)
			if readError != nil {
				test.Fatal(readError)
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
	// A lowering or native build runs the compiler linked into this test binary. Its key names the compiler by its
	// sources (numbersCompilerFiles, the Go release) rather than by this binary's bytes, which differ wherever the test
	// binary was built, so Workshop's product and a runner's ask are one key (#nm31pcn).

	return keyed
}

type Inputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain string
}

func Identity(test *testing.T, path string) string {
	test.Helper()
	hash := sha256.New()
	walkError := filepath.WalkDir(path, func(name string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		content, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		relative, relativePathError := filepath.Rel(path, name)
		if relativePathError != nil {
			return relativePathError
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return fmt.Sprintf("oracle-content-sha256=%x", hash.Sum(nil))
}
