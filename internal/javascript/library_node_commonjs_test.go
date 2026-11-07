package javascript

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A bundled image's __filename and __dirname are, in the emitted ES module, the module's own path and its
// directory, as Node's path.dirname answers it: Node runs the module from a file, and what it prints is that file's
// path and the directory it's in. The imports they need come with them. Mutants: __dirname emitted as
// process.cwd(), since Node runs here from the package's directory, or node:url's import left out, and Node prints
// another line or fails to load the module.
func TestCommonJSBundleValuesAreTheModulesPath(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Source: "commonjs.ts", Main: []ir.Statement{
		ir.WriteLine{Stream: ir.Stdout, Value: ir.ProcessCall{Operation: "filename", Of: ir.String}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.ProcessCall{Operation: "dirname", Of: ir.String}},
	}}
	runtime := "export const decodeJson=()=>{},encodeJson=()=>{},fileStatus=()=>{},panic=()=>{},programArguments=()=>{},readDirectory=()=>{},readTextFile=()=>{},utf8At=()=>{},utf8Length=()=>{},writeTextFile=()=>{};"
	output := strings.Replace(JavaScript(program), "from 'adamic'", "from "+quote("data:text/javascript,"+runtime), 1)
	for _, imported := range []string{"import * as adamicNodeURL from 'node:url';\n", "import * as adamicNodePath from 'node:path';\n"} {
		if !strings.Contains(output, imported) {
			t.Fatalf("missing %q:\n%s", imported, output)
		}
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skipf("missing Node path witness: %v", err)
	}
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	module := filepath.Join(directory, "commonjs.mjs")
	if err := os.WriteFile(module, []byte(output), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := exec.Command("node", module).CombinedOutput()
	if want := module + "\n" + directory + "\n"; err != nil || string(result) != want {
		t.Fatalf("want %q, got %v: %q", want, err, result)
	}
}
