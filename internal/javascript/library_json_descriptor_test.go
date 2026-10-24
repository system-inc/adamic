package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestJSONDescriptorsBuiltOnce(t *testing.T) {
	t.Parallel()
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
	output := strings.Replace(JavaScript(program), "from 'adamic'", "from "+quote("data:text/javascript,"+runtime), 1)
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

// A program that never decodes or encodes JSON carries nothing of it but the two names in the import of
// adamic: no schema descriptor and no call. Stated as a property rather than a pinned hash of the whole
// output, so a change to the rest of the emitter never reds it.
func TestWithoutJSONCarriesOnlyTheImport(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "01_hello.ts", Strings: []string{"In dedication, with gratitude, to Kenneth Lane Thompson, whose work we stand on. - Kirk and Ahra"}, Main: []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}}}
	code := JavaScript(program)
	if strings.Contains(code, "adamicJsonSchema_") {
		t.Errorf("a program without JSON carries a schema descriptor:\n%s", code)
	}
	for _, name := range []string{"encodeJson", "decodeJson"} {
		if count := strings.Count(code, name); count != 1 {
			t.Errorf("%s appears %d times in a program without JSON, want once, in the import:\n%s", name, count, code)
		}
	}
}
