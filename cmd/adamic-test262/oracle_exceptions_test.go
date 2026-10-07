package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func oracleRuntimePath(t *testing.T) string {
	t.Helper()
	runtime, err := filepath.Abs("../../oracle/adamic.mjs")
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

func TestOracleUncaughtExceptionNames(t *testing.T) {
	runtime := oracleRuntimePath(t)
	for _, name := range []string{"ReferenceError", "TypeError", "RangeError"} {
		t.Run(name, func(t *testing.T) {
			script := filepath.Join(t.TempDir(), "throw.mjs")
			source := fmt.Sprintf("import %q;\nthrow new %s('original failure');\n", runtime, name)
			if err := os.WriteFile(script, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			output, err := exec.Command("node", script).CombinedOutput()
			exit, ok := err.(*exec.ExitError)
			want := "adamic: panic: " + name + ": original failure\n"
			if !ok || exit.ExitCode() != 70 || string(output) != want {
				t.Fatalf("uncaught exception identity: want exit 70 and %q; got %v, %q", want, err, output)
			}
		})
	}
}

func TestOracleUncaughtHostErrorFormatting(t *testing.T) {
	script := filepath.Join(t.TempDir(), "host.mjs")
	source := fmt.Sprintf("import %q;\nimport {createHash} from 'node:crypto';\nconst hash = createHash('sha256'); hash.digest('hex'); hash.update('after digest');\n", oracleRuntimePath(t))
	if err := os.WriteFile(script, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", script).CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	want := "adamic: panic: Error [ERR_CRYPTO_HASH_FINALIZED]: Digest already called\n"
	if !ok || exit.ExitCode() != 70 || string(output) != want {
		t.Fatalf("host exception formatting: want exit 70 and %q; got %v, %q", want, err, output)
	}
}

func TestOracleUncaughtSyntheticErrorFormatting(t *testing.T) {
	script := filepath.Join(t.TempDir(), "synthetic.mjs")
	source := fmt.Sprintf("import %q;\nthrow {name: 'ReferenceError', message: 'original failure', toString: (receiver) => receiver.message};\n", oracleRuntimePath(t))
	if err := os.WriteFile(script, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", script).CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	want := "adamic: panic: ReferenceError: original failure\n"
	if !ok || exit.ExitCode() != 70 || string(output) != want {
		t.Fatalf("synthetic exception formatting: want exit 70 and %q; got %v, %q", want, err, output)
	}
}
