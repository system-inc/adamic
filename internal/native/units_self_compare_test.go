package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestSplitSelfCompareAgreesWithNode(test *testing.T) {
	test.Parallel()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	fixture, readError := os.ReadFile("../../stage3/fixtures/self-compare/sweep.a")
	if readError != nil {
		test.Fatal(readError)
	}
	// Cross the splitter's sixteen-function group boundary, as large products do.
	for index := 0; index < 16; index++ {
		fixture = append(fixture, []byte(fmt.Sprintf(
			"\nfunction splitCompare%d(value: number) { console.log(String(value === value)); }\nsplitCompare%d(Number.parseFloat(\"not a number\"));\n",
			index,
			index,
		))...)
	}
	path := filepath.Join(test.TempDir(), "sweep.a")
	if writeError := os.WriteFile(path, fixture, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	// Node executes the source fixture with its TypeScript syntax stripped.
	want, outputError := exec.CommandContext(
		timeoutContext,
		"node",
		"--input-type=module",
		"--eval",
		`import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; eval(stripTypeScriptTypes(readFileSync(process.argv[1], 'utf8')));`,
		path,
	).Output()
	if outputError != nil {
		test.Fatalf("Node: %v", outputError)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	lowered, lowerError := lower.Lower(timeoutContext, program)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	binary := filepath.Join(test.TempDir(), "program")
	// Build uses -Werror, including C's inline definition diagnostic.
	if buildError := Build(C(lowered), binary, Options{Split: true, Jobs: 2}); buildError != nil {
		test.Fatal(buildError)
	}
	got, combinedOutputError := exec.CommandContext(timeoutContext, binary).CombinedOutput()
	if combinedOutputError != nil || !bytes.Equal(got, want) {
		test.Fatalf("native: %v\ngot %q\nNode %q", combinedOutputError, got, want)
	}
}
