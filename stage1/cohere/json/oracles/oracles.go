// Package oracles owns json's compiler-free Go reference products.
// Both test binaries resolve relative paths from stage1/cohere/json.
package oracles

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

const repository = "../../.."

var oracleProducer bool

// EnableProductProduction is shared with the parent json tests.
func EnableProductProduction() { oracleProducer = true }
func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	t.Helper()
	return buildcache.Product(t, inputs, func(dir string) error {
		if !oracleProducer {
			return fmt.Errorf("settle json/oracles owners first: %s", inputs.Name)
		}
		return build(dir)
	})
}
func stockOracle(t *testing.T, output, pkg string, args []string, mutations []buildcache.Mutation, test bool) string {
	t.Helper()
	if oracleProducer {
		if !test {
			return buildcache.GoBuildIn(t, "cohere", output, pkg, args)
		}
		if len(mutations) == 0 {
			return buildcache.GoTest(t, "cohere", output, pkg, args)
		}
		return buildcache.GoTestMutated(t, "cohere", output, pkg, args, mutations)
	}
	var inputs buildcache.Inputs
	var err error
	if test {
		inputs, err = buildcache.GoTestInputs("cohere", output, pkg, args, mutations)
	} else {
		inputs, err = buildcache.GoBuildInInputs("cohere", output, pkg, args, mutations)
	}
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, output, inputs, "stage1/cohere/json/oracles owner", func(string) error { return fmt.Errorf("settle json/oracles owners first") })
	return filepath.Join(directory, output)
}

type ComparisonCase struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

type ComparisonAnswer struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}

// SameAnswer is shared with the parent json tests.
func SameAnswer(left, right ComparisonAnswer) bool {
	if (left.Error == "") != (right.Error == "") {
		return false
	}
	return left.Error != "" || left.Output == right.Output
}

func escape(text string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(text)
}

// Protocol is shared with the parent json tests.
func Protocol(cases []ComparisonCase, answers []ComparisonAnswer) (string, string) {
	var inputs, outputs strings.Builder
	for index, item := range cases {
		fmt.Fprintf(&inputs, "%s\t%s\n", item.Name, escape(item.Text))
		value := answers[index]
		if value.Error != "" {
			fmt.Fprintf(&outputs, "error\t%s\n", escape(value.Error))
		} else {
			fmt.Fprintf(&outputs, "ok\t%s\n", escape(value.Output))
		}
	}
	return inputs.String(), outputs.String()
}

// ParseAnswers is shared with the parent json tests.
func ParseAnswers(Protocol string, count int) ([]ComparisonAnswer, error) {
	lines := strings.Split(strings.TrimSuffix(Protocol, "\n"), "\n")
	if len(lines) != count {
		return nil, fmt.Errorf("Go answered %d/%d cases", len(lines), count)
	}
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	answers := make([]ComparisonAnswer, count)
	for index, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Go case %d malformed answer", index)
		}
		value := unescape.Replace(parts[1])
		switch parts[0] {
		case "ok":
			answers[index].Output = value
		case "error":
			answers[index].Error = value
		default:
			return nil, fmt.Errorf("Go case %d unknown answer kind %q", index, parts[0])
		}
	}
	return answers, nil
}

var upstreamParityOracleBuild struct {
	once sync.Once
	path string
}

// AuditOracle is shared with the parent json tests.
func AuditOracle(t *testing.T) string {
	t.Helper()
	upstreamParityOracleBuild.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		side, err := filepath.Abs("testdata/cohere_side_test.go")
		if err != nil {
			t.Fatal(err)
		}
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/internal/format/javascript/adamic_json_audit_test.go"): side}})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			t.Fatal(err)
		}
		upstreamParityOracleBuild.path = stockOracle(t, "oracle-oracle-owner-v1", "./internal/format/javascript", []string{"-overlay=" + path}, nil, true)
	})
	if upstreamParityOracleBuild.path == "" {
		t.Fatal("upstream oracle preparation failed")
	}
	return upstreamParityOracleBuild.path
}

// FormatterOracle is shared with the parent json tests.
func FormatterOracle(t *testing.T) string {
	t.Helper()

	// cohere's command/formatter_comparison with testdata/cohere_driver.go laid over its main, keyed by what go
	// says the build compiles (GoBuildIn). The overlay names this checkout's paths, so it sits in a temporary
	// directory: it builds the product and is never part of it (#tqrqx60).
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("testdata/cohere_driver.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): source}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	return stockOracle(t, "go-cohere-oracle-owner-v1", "./command/formatter_comparison", []string{"-overlay=" + path}, nil, false)
}

type PrinterMutation struct{ Name, File, From, To string }

var printerMutations = []PrinterMutation{
	{"missing final newline", "print_json.go", `return concatIn(path, print("node", nil), hardline)`, `return print("node", nil)`},
	{"missing property space", "print_json.go", `print("key", nil), ": ", print("value", nil)`, `print("key", nil), ":", print("value", nil)`},
	{"wrong filename parser", "printer.go", `return "json-stringify"`, `return "json"`},
}

// PrinterMutations is shared with the parent json tests.
func PrinterMutations() []PrinterMutation { return append([]PrinterMutation(nil), printerMutations...) }

// PrinterMutant is shared with the parent json tests.
func PrinterMutant(t *testing.T, planted []buildcache.Mutation) string {
	t.Helper()
	cohere, err := filepath.Abs("../../../cohere")
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	writeJSON(t, overlayPath, map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/javascript/adamic_json_audit_test.go"): side}})
	return stockOracle(t, "javascript.test-oracle-owner-v1", "./internal/format/javascript", []string{"-overlay=" + overlayPath}, planted, true)
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
}

// ClassOrderAnswerInputs is shared with the parent json tests.
func ClassOrderAnswerInputs(oracle []byte, item ComparisonCase) buildcache.Inputs {
	return buildcache.Inputs{
		Name:  "json class-order Go answer oracle-owner-v1",
		Files: []string{"stage1/cohere/json/class_order_grain30_test.go", "stage1/cohere/json/port_test.go"},
		Flags: []string{fmt.Sprintf("Go oracle binary sha256 %x", sha256.Sum256(oracle)), fmt.Sprintf("case %s sha256 %x", item.Name, sha256.Sum256([]byte(item.Text))), "--cases"},
	}
}

func buildJSONClassOrderAnswer(oracle string, item ComparisonCase) func(string) error {
	return func(directory string) error {
		input, _ := Protocol([]ComparisonCase{item}, make([]ComparisonAnswer, 1))
		cases := filepath.Join(directory, "cases.txt")
		if err := os.WriteFile(cases, []byte(input), 0644); err != nil {
			return err
		}
		reference, err := executeResult(nil, oracle, "--cases", cases)
		if err != nil {
			return err
		}
		if reference.exitCode != 0 || len(reference.stderr) != 0 {
			return fmt.Errorf("Go oracle: exit %d stderr %s", reference.exitCode, reference.stderr)
		}
		if _, err := ParseAnswers(string(reference.stdout), 1); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "answer.txt"), reference.stdout, 0644)
	}
}

var jsonClassOrderAnswer struct {
	once   sync.Once
	answer string
}

// ClassOrderAnswer is shared with the parent json tests.
func ClassOrderAnswer(t *testing.T) string {
	t.Helper()
	jsonClassOrderAnswer.once.Do(func() {
		oracle := FormatterOracle(t)
		binary, err := os.ReadFile(oracle)
		if err != nil {
			t.Fatal(err)
		}
		all := CorpusCases(t)
		var items []ComparisonCase
		for _, item := range all {
			if item.Name == "cohere/internal/lint/rules/tailwind/collapse/testdata/classorder_fixtures.json" {
				items = append(items, item)
			}
		}
		if len(items) != 1 {
			t.Fatalf("dedicated fixture count %d, want 1", len(items))
		}
		directory := oracleProduct(t, ClassOrderAnswerInputs(binary, items[0]), buildJSONClassOrderAnswer(oracle, items[0]))
		stored, err := os.ReadFile(filepath.Join(directory, "answer.txt"))
		if err != nil {
			t.Fatal(err)
		}
		jsonClassOrderAnswer.answer = string(stored)
	})
	if jsonClassOrderAnswer.answer == "" {
		t.Fatal("the class-order Go answer failed earlier in this process")
	}
	return jsonClassOrderAnswer.answer
}

var jsonGuard = childguard.Options{Stall: 5 * time.Minute}

type execution struct {
	stdout, stderr []byte
	exitCode       int
}

func executeResult(_ []string, binary string, args ...string) (execution, error) {
	command := toolchain.CommandIn("", binary, args...)
	command.Env = nil
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, jsonGuard)
	var exitError interface{ ExitCode() int }
	if err != nil && !errors.As(err, &exitError) {
		return execution{}, err
	}
	return execution{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}, nil
}
