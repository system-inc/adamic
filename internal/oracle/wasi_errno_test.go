package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// wasiErrnoCodes are the codes node_fs_file.c's system_error_words names, each mapped by node_errno.h
// to Linux's number under ADAMIC_TARGET_WASI.
var wasiErrnoCodes = []string{"EPERM", "ENOENT", "EINTR", "EIO", "ENXIO", "EBADF", "EAGAIN", "ENOMEM", "EACCES", "EEXIST", "ENOTDIR", "EISDIR", "EINVAL", "ENFILE", "EMFILE", "EFBIG", "ENOSPC", "EROFS", "ENAMETOOLONG", "ENOSYS", "ENOTEMPTY", "ELOOP"}

// A wasm32 module's errno numbers are Linux's (Adamic on #z4a9p9e): node_errno.h maps each named code
// from wasi-libc's own numbering, where ENOENT is 44, to the number Node gives it on Linux, and any other
// code to libuv's UNKNOWN, -4094. Node on this host is the witness: each code's number is
// -os.constants.errno[code], and util.getSystemErrorName must name it back, so a table that gave a code
// another code's number fails here too. Linux's numbers are Node's on Linux alone (docs/wasm-limits.md,
// "The host's errno numbers"), so the test runs only there and says so anywhere else. Mutant,
// testdata/wasi/errno.mutants.md: the map dropped, every line differs.
func TestWASIErrnoIsNodesOnLinux(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	if runtime.GOOS != "linux" {
		t.Fatalf("the wasm32 errno table holds Linux's numbers, so it's held to Node on Linux, not %s (docs/wasm-limits.md)", runtime.GOOS)
	}
	script := "const os = require('node:os'), util = require('node:util');\n" +
		"for (const code of ['" + strings.Join(wasiErrnoCodes, "', '") + "']) {\n" +
		"  const number = -os.constants.errno[code];\n" +
		"  if (util.getSystemErrorName(number) !== code) throw new Error(code + ' is ' + number + ', named ' + util.getSystemErrorName(number));\n" +
		"  console.log(code + ' ' + number);\n" +
		"}\n" +
		"if (util.getSystemErrorName(-4094) !== 'UNKNOWN') throw new Error('-4094 is ' + util.getSystemErrorName(-4094));\n" +
		"console.log('UNKNOWN -4094');\n"
	node := execute(t, "node", "--input-type=commonjs", "--eval", script)
	if node.exitCode != 0 || len(node.stderr) != 0 {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	var lines strings.Builder
	for _, code := range wasiErrnoCodes {
		fmt.Fprintf(&lines, "    printf(\"%s %%.0f\\n\", adamic_node_errno(%s));\n", code, code)
	}
	// EXDEV is a code system_error_words doesn't name, so it reads UNKNOWN there and -4094 here.
	source := fmt.Sprintf("#include <stdio.h>\n#include %q\nint main(void) {\n%s    printf(\"UNKNOWN %%.0f\\n\", adamic_node_errno(EXDEV));\n    return 0;\n}\n",
		filepath.Join(root, "internal/native/runtime/node_errno.h"), lines.String())
	wasm := onWASI(t, source)
	if difference := disagreement(node, wasm); difference != "" {
		t.Fatalf("%s: Node %q, wasm32 %q", difference, node.stdout, wasm.stdout)
	}
	t.Logf("Node and wasm32: %q", wasm.stdout)
}
