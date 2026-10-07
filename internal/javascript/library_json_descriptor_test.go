package javascript

import (
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestJSONDescriptorsBuiltOnce(t *testing.T) {
	number := ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "number", Of: ir.Number, Expected: "number"}}}
	text := ir.JSONDecodeSchema{Nodes: []ir.JSONDecodeNode{{Kind: "string", Of: ir.String, Expected: "string"}}}
	program := &ir.Program{Source: "schemas.a", Strings: []string{"3", "text"}, Main: []ir.Statement{
		ir.WriteLine{Value: ir.JSONEncode{Value: ir.NumberConstant{Value: 3}, Schema: number}},
		ir.WriteLine{Value: ir.JSONDecode{Text: ir.StringConstant{Index: 0}, Schema: number}},
		ir.WriteLine{Value: ir.JSONEncode{Value: ir.NumberConstant{Value: 4}, Schema: number}},
		ir.WriteLine{Value: ir.JSONEncode{Value: ir.StringConstant{Index: 1}, Schema: text}},
		ir.WriteLine{Value: ir.JSONDecode{Text: ir.StringConstant{Index: 1}, Schema: text}},
	}}
	// A platform witness checks object identity across decode/encode calls.
	runtime := "const seen=new Set(); const visit=(v,s)=>{const before=seen.has(s);seen.add(s);return before?'same':'new';}; export const decodeJson=visit,encodeJson=visit; export const fileStatus=()=>{},panic=()=>{},programArguments=()=>{},readDirectory=()=>{},readTextFile=()=>{},utf8At=()=>{},utf8Length=()=>{},writeTextFile=()=>{};"
	output := JavaScriptWith(program, Options{RuntimeImport: "data:text/javascript," + runtime})
	if strings.Count(output, "const adamicJsonSchema_") != 2 || strings.Count(output, `{"nodes":`) != 2 {
		t.Fatalf("schemas not deduplicated:\n%s", output)
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("missing Node identity witness: %v", err)
	}
	result, err := exec.Command("node", "--input-type=module", "-e", output).CombinedOutput()
	if err != nil || string(result) != "new\nsame\nsame\nnew\nsame\n" {
		t.Fatalf("descriptor identities: %v\n%s", err, result)
	}
	// Reallocating even the descriptor root defeats the runtime's WeakMap.
	mutated := regexp.MustCompile(`, (adamicJsonSchema_[0-9]+)\)`).ReplaceAllString(output, ", {...$1})")
	result, err = exec.Command("node", "--input-type=module", "-e", mutated).CombinedOutput()
	if err != nil || string(result) != "new\nnew\nnew\nnew\nnew\n" {
		t.Fatalf("identity mutant failed to execute: %v\n%s", err, result)
	}
	t.Log("per-call descriptor allocation mutant caught by Node object identity")

}

func TestWithoutJSONIsByteIdentical(t *testing.T) {
	// Recorded from 0817c3d's unchanged emitter using the dedication fixture.
	program := &ir.Program{Source: "01_hello.ts", Strings: []string{"In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra"}, Main: []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}}}
	got := fmt.Sprintf("%x", sha256.Sum256([]byte(JavaScript(program))))
	const want = "1c88ae6b1f76714d5ed3252b5578ba7f4b71a2005802ca751911c02c332f9229"
	if got != want {
		t.Fatalf("non-JSON output changed: %s, want %s", got, want)
	}
}
