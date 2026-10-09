package javascript

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// Exercise actual cyclic input, separately from the finite source witness used by mutants.
func TestCheckedJSONCycleStopsAtDepth(t *testing.T) {
	script := `const adamicTypeOf=value=>typeof value;
const panic=message=>{process.stderr.write('adamic: panic: '+message+'\n');process.exit(70);};
` + checkedJSONRuntime + `
const raw={};raw.next=raw;
adamicCheckedJSON(raw,{Kind:'object',Name:'RecursiveConfig'},'raw');
`
	command := exec.Command("node", "--input-type=module", "-e", script)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	failure, ok := err.(*exec.ExitError)
	want := "adamic: panic: checked any: raw" + strings.Repeat(".next", 65) + " needs JSON value | undefined, found non-JSON recursion\n"
	if !ok || failure.ExitCode() != 70 || stdout.Len() != 0 || stderr.String() != want {
		t.Fatalf("cyclic data must stop at the depth boundary: %v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
}
