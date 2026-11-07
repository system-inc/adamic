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

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func initializerWitness(test *testing.T, split bool, exceptions bool) {
	test.Helper()
	directory := test.TempDir()
	entry := filepath.Join(directory, "main.a")
	var elements []string
	for i := 0; i < 160; i++ {
		elements = append(elements, fmt.Sprintf("{value:%d}", i))
	}
	source := `import { first } from './first.a';import { second } from './second.a';import { values } from './values.a';console.log(first+second);console.log(values.length.toString());let trace='';for(const item of values){trace+=item.value.toString()+',';}console.log(trace);{const held=trace+'held';const read=()=>held;console.log(read());}`
	if exceptions {
		source += `try{const held=trace+'throw';throw new Error(held);}catch(error){console.log('caught');}finally{console.log('finally');}`
	}
	for path, text := range map[string]string{
		entry:                               source,
		filepath.Join(directory, "first.a"): `console.log('first');export const first='one';`,
		filepath.Join(
			directory,
			"second.a",
		): `import { first } from './first.a';console.log('second '+first);export const second='two';`,
		filepath.Join(
			directory,
			"values.a",
		): `export const values=[` + strings.Join(elements, ",") + `];console.log('values first');`,
	} {
		if writeError := os.WriteFile(path, []byte(text), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	checked, loadError := load.Load([]string{entry})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	initializerCompiledWitness(test, entry, program, split)
}

func initializerCompiledWitness(test *testing.T, entry string, program *ir.Program, split bool) {
	test.Helper()
	directory := test.TempDir()
	code := native.C(program)
	if code != native.C(program) {
		test.Fatal("initializer emission is not deterministic")
	}
	if strings.Count(code, " __attribute__((noinline))") < 2 {
		test.Fatal("initializer helpers are missing")
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(
		code,
		binary,
		native.Options{Split: split, Sanitize: true, Count: true, Jobs: 1},
	); buildError != nil {
		test.Fatal(buildError)
	}
	js := filepath.Join(directory, "program.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	run := func(command string, args ...string) []byte {
		test.Helper()
		timeoutContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		osCommand := exec.CommandContext(timeoutContext, command, args...)
		osCommand.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		out, outputError := osCommand.Output()
		if outputError != nil {
			test.Fatalf("%s: %v", command, outputError)
		}
		return out
	}
	truth := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	for _, actual := range [][]byte{
		run(binary),
		run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js),
	} {
		if !bytes.Equal(actual, truth) {
			test.Fatal("initializer output differs from Node")
		}
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
}
func TestInitializerStorageSplit(test *testing.T) {
	test.Parallel()
	initializerWitness(test, true, false)
}
func TestInitializerStorageWhole(test *testing.T) {
	test.Parallel()
	initializerWitness(test, false, false)
}
func TestInitializerExceptionsSplit(test *testing.T) {
	test.Parallel()
	initializerWitness(test, true, true)
}
func TestInitializerExceptionsWhole(test *testing.T) {
	test.Parallel()
	initializerWitness(test, false, true)
}

func TestInitializerTypedArrayChunks(test *testing.T) {
	test.Parallel()
	initializerFixtureWitness(test, "typed_arrays_uint16array")
}
func TestInitializerTypedArrayViewChunks(test *testing.T) {
	test.Parallel()
	initializerFixtureWitness(test, "typed_arrays_views")
}

func initializerFixtureWitness(test *testing.T, name string) {
	test.Helper()
	entry, absolutePathError := filepath.Abs("../oracle/testdata/" + name + ".a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	checked, loadError := load.Load([]string{entry})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	initializerCompiledWitness(test, entry, program, true)
}
