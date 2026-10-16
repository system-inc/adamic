package native_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func initializerWitness(t *testing.T, split bool, exceptions bool) {
	t.Helper()
	directory := t.TempDir()
	entry := filepath.Join(directory, "main.a")
	var elements []string
	for i := 0; i < 160; i++ {
		elements = append(elements, fmt.Sprintf("{value:%d}", i))
	}
	source := `import { values } from './values.a';console.log(values.length.toString());let trace='';for(const item of values){trace+=item.value.toString()+',';}console.log(trace);{const held=trace+'held';const read=()=>held;console.log(read());}`
	if exceptions {
		source += `try{const held=trace+'throw';throw new Error(held);}catch(error){console.log('caught');}finally{console.log('finally');}`
	}
	for path, text := range map[string]string{entry: source, filepath.Join(directory, "values.a"): `export const values=[` + strings.Join(elements, ",") + `];console.log('values first');`} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	checked, err := load.Load([]string{entry})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	if code != native.C(program) {
		t.Fatal("initializer emission is not deterministic")
	}
	if strings.Count(code, " __attribute__((noinline))") < 2 {
		t.Fatal("initializer helpers are missing")
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(code, binary, native.Options{Split: split, Sanitize: true, Count: true, Jobs: 1}); err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(command string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, command, args...)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		out, err := cmd.Output()
		if err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		return out
	}
	truth := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	for _, actual := range [][]byte{run(binary), run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js)} {
		if !bytes.Equal(actual, truth) {
			t.Fatal("initializer output differs from Node")
		}
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
}
func TestInitializerStorageSplit(t *testing.T)    { t.Parallel(); initializerWitness(t, true, false) }
func TestInitializerStorageWhole(t *testing.T)    { t.Parallel(); initializerWitness(t, false, false) }
func TestInitializerExceptionsSplit(t *testing.T) { t.Parallel(); initializerWitness(t, true, true) }
func TestInitializerExceptionsWhole(t *testing.T) { t.Parallel(); initializerWitness(t, false, true) }
