package native

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

// Runtime evidence stays runnable while the separate @types/node loader lands.
// Build IR directly, and compare both backends with independent source on Node.
func TestNodeBufferRuntimeWithoutDeclarations(t *testing.T) {
	t.Parallel()
	program := &ir.Program{}
	call := func(name string, encoding int, returns ir.Type, arguments ...ir.Expression) ir.Expression {
		return ir.NodeBufferCall{Function: name, Encoding: encoding, Returns: returns, Arguments: arguments}
	}
	text := func(value string) ir.Expression {
		index := len(program.Strings)
		program.Strings = append(program.Strings, value)
		return ir.StringConstant{Index: index}
	}
	bytes := func(values []int) ir.Expression {
		elements := make([]ir.Expression, len(values))
		for index, value := range values {
			elements[index] = ir.NumberConstant{Value: float64(value)}
		}
		return call("buffer_copy", 0, ir.Array, ir.ArrayLiteral{Elements: elements, Element: ir.Number})
	}
	stringify := func(buffer ir.Expression, encoding int) ir.Expression {
		return call("buffer_string", encoding, ir.String, buffer, ir.NumberConstant{Value: 0}, ir.NumberConstant{Value: math.Inf(1)})
	}
	print := func(value ir.Expression) {
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: value})
	}
	var source strings.Builder
	source.WriteString("import { Buffer } from 'node:buffer'; import { createHash } from 'node:crypto';\n")
	for _, value := range []string{"", "YQ==", " Y W\nJ\tj = = ", "YWJj====", "-__v", "Y!W@J#j"} {
		print(stringify(call("buffer_from", 2, ir.Array, text(value)), 3))
		quoted, _ := json.Marshal(value)
		fmt.Fprintf(&source, "console.log(Buffer.from(%s,'base64').toString('hex'));\n", quoted)
	}
	for _, value := range []string{"", "abc", "héllo 🌍", "a\x00b"} {
		print(stringify(call("buffer_from", 0, ir.Array, text(value)), 2))
		quoted, _ := json.Marshal(value)
		fmt.Fprintf(&source, "console.log(Buffer.from(%s,'utf8').toString('base64'));\n", quoted)
	}
	for _, values := range [][]int{{0xe0, 0x80, 0x80}, {0xe2, 0x82}, {0xf0, 0x90, 0x80, 0x41}, {0xed, 0xa0, 0x80}, {0xf4, 0x90, 0x80, 0x80}} {
		print(stringify(call("buffer_from", 0, ir.Array, stringify(bytes(values), 0)), 3))
		encoded, _ := json.Marshal(values)
		fmt.Fprintf(&source, "console.log(Buffer.from(Buffer.from(%s).toString('utf8'),'utf8').toString('hex'));\n", encoded)
	}
	random := rand.New(rand.NewSource(0x12345678))
	for sample := 0; sample < 32; sample++ {
		values := []int{0, 0xd8, 0xff, 0xdf, 0, 0xdc, 0, 0xd8}
		for index := 0; index < sample*2+1; index++ {
			values = append(values, random.Intn(256))
		}
		decoded := stringify(bytes(values), 1)
		print(stringify(call("buffer_from", 1, ir.Array, decoded), 3))
		print(call("hash_digest", 0, ir.String, call("hash_update", 0, ir.Object, call("hash_new", 0, ir.Object), decoded)))
		encoded, _ := json.Marshal(values)
		fmt.Fprintf(&source, "{ const text=Buffer.from(%s).toString('utf16le'); console.log(Buffer.from(text,'utf16le').toString('hex')); console.log(createHash('sha256').update(text).digest('hex')); }\n", encoded)
	}
	directory := t.TempDir()
	oracle := filepath.Join(directory, "oracle.mjs")
	if err := os.WriteFile(oracle, []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	want := runWithInput(t, "", "node", runner, oracle)
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", "node", runner, generated); got != want {
		t.Fatalf("JavaScript stdout differs from Node: got %q, want %q", got, want)
	}
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(directory, "native")
		if err := Build(C(program), binary, Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		if got := runWithInput(t, "", binary); got != want {
			t.Fatalf("native stdout differs from Node (sanitize %v): got %q, want %q", sanitize, got, want)
		}
	}
}
