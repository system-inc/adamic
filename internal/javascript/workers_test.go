package javascript

import (
	"context"
	"github.com/system-inc/adamic/internal/ir"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestWorkersCalleeWithoutNodeGlobals(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		name  string
		count int
	}{{"ordinary", 0}, {"argument count", 1}} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			program := &ir.Program{Source: "worker.a", Locals: []ir.Local{{Name: "count", Type: ir.Number, Function: 0}}, Functions: []ir.Function{{Name: "handle", ArgumentsCount: one.count, Returns: ir.Number, Body: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: 1}}}}}}
			runtime := "const noop=()=>{}; export {noop as encodeJson,noop as decodeJson,noop as fileStatus,noop as panic,noop as programArguments,noop as readDirectory,noop as readTextFile,noop as utf8At,noop as utf8Length,noop as writeTextFile};"
			code := JavaScriptWith(program, Options{RuntimeImport: "data:text/javascript," + runtime})
			for _, forbidden := range []string{"node:", "process.", "Buffer."} {
				if strings.Contains(code, forbidden) {
					t.Fatalf("Workers module uses unavailable %s", forbidden)
				}
			}
			code += `
globalThis.process=undefined; globalThis.Buffer=undefined;
const own={run:{code:(_closure,values)=>values[0]+1}};
console.log(adamicCall(adamicCallee(own,'run'),[41]));
const inherited=Object.create({run:(receiver,value)=>receiver.base+value}); inherited.base=2;
console.log(adamicCall(adamicCallee(inherited,'run'),[40]));
`
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			output, err := exec.CommandContext(ctx, "node", "--input-type=module", "-e", code).CombinedOutput()
			if err != nil || string(output) != "42\n42\n" {
				t.Fatalf("Workers method dispatch without Node globals: %v\n%s", err, output)
			}
		})
	}
}
