package javascript

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

const viewTestRuntime = `const panic = message => { process.stderr.write("adamic: panic: " + message + "\n"); process.exit(70); };
class AdamicClosure { constructor(code) { this.code = code; } }
const adamicTypeOf = value => value instanceof AdamicClosure ? "function" : typeof value;
const adamicCall = (value, arguments_) => value.code(value, arguments_);
const adamicClassIdentities = new WeakMap();
`

func runViewNode(test *testing.T, source, stdout, stderr string, exit int) {
	test.Helper()
	command := toolchain.Command("node", "--input-type=module")
	command.Stdin = strings.NewReader(source)
	var out, errout bytes.Buffer
	command.Stdout, command.Stderr = &out, &errout
	err := command.Run()
	code := 0
	if err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			code = failure.ExitCode()
		} else {
			test.Fatal(err)
		}
	}
	if code != exit || out.String() != stdout || errout.String() != stderr {
		test.Fatalf(
			"exit %d, stdout %q, stderr %q; want %d, %q, %q",
			code,
			out.String(),
			errout.String(),
			exit,
			stdout,
			stderr,
		)
	}
}
