package native

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

type nodeBufferHostObservation struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Exit   int    `json:"exit"`
}

func nodeBufferHostRun(t *testing.T, command string, arguments ...string) nodeBufferHostObservation {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	process := exec.CommandContext(ctx, command, arguments...)
	process.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	process.Cancel = func() error { return syscall.Kill(-process.Process.Pid, syscall.SIGKILL) }
	process.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	process.Stdout = &stdout
	process.Stderr = &stderr
	err := process.Run()
	code := 0
	if err != nil {
		if status, ok := err.(*exec.ExitError); ok {
			code = status.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return nodeBufferHostObservation{stdout.String(), stderr.String(), code}
}

// This is deliberately a component test, not a claim that the unmodified host
// fixture passes the checker. Its Node observation is the actual fixture and
// status.json, while IR isolates this unit from the pending shared loader/fs.
func nodeBufferHostCompare(t *testing.T, file string, program *ir.Program) {
	t.Helper()
	bucket, err := filepath.Abs("../../stage3/fixtures/host")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(bucket, "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		File string                    `json:"file"`
		Node nodeBufferHostObservation `json:"node"`
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var expected nodeBufferHostObservation
	found := false
	for _, row := range rows {
		if row.File == file {
			expected = row.Node
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("no host observation for %s", file)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	if got := nodeBufferHostRun(t, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(bucket, file)); got != expected {
		t.Fatalf("fixture Node observation differs from status.json: got %+v, want %+v", got, expected)
	}
	directory := t.TempDir()
	js := filepath.Join(directory, "component.mjs")
	if err = os.WriteFile(js, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := nodeBufferHostRun(t, "node", "--disable-warning=ExperimentalWarning", runner, js); got != expected {
		t.Fatalf("JavaScript component differs from status.json: got %+v, want %+v", got, expected)
	}
	generated := C(program)
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "component")
		if err = Build(generated, binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := nodeBufferHostRun(t, binary); got != expected {
			t.Fatalf("native component differs from status.json (sanitize %v): got %+v, want %+v", sanitize, got, expected)
		}
	}
}

func TestNodeBufferHostSHA256Component(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "21_createHash.a", Strings: []string{"", "abc", "héllo 🌍", "a\x00b"}}
	for index := range program.Strings {
		hash := ir.NodeBufferCall{Function: "hash_new", Returns: ir.Object}
		update := ir.NodeBufferCall{Function: "hash_update", Returns: ir.Object, Arguments: []ir.Expression{hash, ir.StringConstant{Index: index}}}
		digest := ir.NodeBufferCall{Function: "hash_digest", Returns: ir.String, Arguments: []ir.Expression{update}}
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: digest})
	}
	nodeBufferHostCompare(t, "21_createHash.a", program)
}

func TestNodeBufferHostFallbackComponent(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../stage3/fixtures/host/22_createHash_fallback.a")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), "export function generateDjb2Hash")
	if start < 0 {
		t.Fatal("upstream fallback function absent")
	}
	end := strings.Index(string(source)[start:], "\n}\n")
	if end < 0 {
		t.Fatal("upstream fallback function boundary absent")
	}
	function := string(source)[start : start+end+3]
	// Preserve the upstream function verbatim; only its scratch-directory driver
	// is replaced. The absent-crypto selection remains a full-fixture dependency.
	path := filepath.Join(t.TempDir(), "fallback.a")
	driver := function + "\nfor (const text of ['', 'abc', 'héllo 🌍', 'x'.repeat(40)]) { console.log(generateDjb2Hash(text)); }\n"
	if err = os.WriteFile(path, []byte(driver), 0o644); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	nodeBufferHostCompare(t, "22_createHash_fallback.a", program)
}

func TestNodeBufferHostUTF16Components(t *testing.T) {
	t.Parallel()
	for _, file := range []string{"02_readFile_utf16le.a", "03_readFile_utf16be.a"} {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("../../stage3/fixtures/host", file))
			if err != nil {
				t.Fatal(err)
			}
			match := regexp.MustCompile(`const cases = (\[\[.*\]\]);`).FindStringSubmatch(string(source))
			if len(match) != 2 {
				t.Fatal("fixture byte cases absent")
			}
			literal := regexp.MustCompile(`0x[0-9a-fA-F]+`).ReplaceAllStringFunc(match[1], func(hex string) string {
				value, err := strconv.ParseInt(hex, 0, 64)
				if err != nil {
					t.Fatal(err)
				}
				return strconv.FormatInt(value, 10)
			})
			var cases [][]int
			if err = json.Unmarshal([]byte(literal), &cases); err != nil {
				t.Fatal(err)
			}
			program := &ir.Program{Source: file}
			number := func(value float64) ir.Expression { return ir.NumberConstant{Value: value} }
			for _, values := range cases {
				base := len(program.Locals)
				program.Locals = append(program.Locals, ir.Local{Type: ir.Array, Function: -1}, ir.Local{Type: ir.Number, Function: -1}, ir.Local{Type: ir.Number, Function: -1}, ir.Local{Type: ir.Number, Function: -1})
				buffer := ir.Read{Local: base, Of: ir.Array}
				length := ir.Read{Local: base + 1, Of: ir.Number}
				index := ir.Read{Local: base + 2, Of: ir.Number}
				temporary := ir.Read{Local: base + 3, Of: ir.Number}
				elements := make([]ir.Expression, len(values))
				for at, value := range values {
					elements[at] = number(float64(value))
				}
				body := []ir.Statement{ir.Declare{Local: base, Value: ir.NodeBufferCall{Function: "buffer_copy", Returns: ir.Array, Arguments: []ir.Expression{ir.ArrayLiteral{Elements: elements, Element: ir.Number}}}}}
				if strings.Contains(file, "utf16be") {
					next := ir.Binary{Operator: ir.Add, Left: index, Right: number(1)}
					at := func(position ir.Expression) ir.Expression {
						return ir.Unwrap{Value: ir.ArrayIndex{Array: buffer, Index: position, Element: ir.Number}}
					}
					store := func(position, value ir.Expression) ir.Statement {
						return ir.Evaluate{Value: ir.NodeBufferCall{Function: "buffer_set", Returns: ir.Number, Arguments: []ir.Expression{buffer, position, value}}}
					}
					body = append(body, ir.Declare{Local: base + 1, Value: ir.Binary{Operator: ir.BitAnd, Left: ir.Length{Array: buffer}, Right: ir.Unary{Operator: ir.BitNot, Operand: number(1)}}}, ir.Declare{Local: base + 2, Value: number(0)},
						ir.Loop{Condition: ir.Binary{Operator: ir.Less, Left: index, Right: length}, Body: []ir.Statement{ir.Declare{Local: base + 3, Value: at(index)}, store(index, at(next)), store(next, temporary)}, Update: []ir.Statement{ir.Assign{Local: base + 2, Value: ir.Binary{Operator: ir.Add, Left: index, Right: number(2)}}}})
				}
				decoded := ir.NodeBufferCall{Function: "buffer_string", Encoding: 1, Returns: ir.String, Arguments: []ir.Expression{buffer, number(2), number(math.Inf(1))}}
				body = append(body, ir.WriteLine{Stream: ir.Stdout, Value: ir.JSONStringify{Value: decoded, Schema: &ir.JSONSchema{Kind: "string"}}})
				program.Main = append(program.Main, ir.Block{Body: body})
			}
			nodeBufferHostCompare(t, file, program)
		})
	}
}
