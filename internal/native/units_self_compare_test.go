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

func TestSplitSelfCompareAgreesWithNode(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	fixture, err := os.ReadFile("../../stage3/fixtures/self-compare/sweep.a")
	if err != nil {
		t.Fatal(err)
	}
	// Cross the splitter's sixteen-function group boundary, as large products do.
	for index := 0; index < 16; index++ {
		fixture = append(fixture, []byte(fmt.Sprintf("\nfunction splitCompare%d(value: number) { console.log(String(value === value)); }\nsplitCompare%d(Number.parseFloat(\"not a number\"));\n", index, index))...)
	}
	path := filepath.Join(t.TempDir(), "sweep.a")
	if err := os.WriteFile(path, fixture, 0644); err != nil {
		t.Fatal(err)
	}
	// Node executes the source fixture with its TypeScript syntax stripped.
	want, err := exec.CommandContext(ctx, "node", "--input-type=module", "--eval", `import {readFileSync} from 'node:fs'; import {stripTypeScriptTypes} from 'node:module'; eval(stripTypeScriptTypes(readFileSync(process.argv[1], 'utf8')));`, path).Output()
	if err != nil {
		t.Fatalf("Node: %v", err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(ctx, program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "program")
	// Build uses -Werror, including C's inline definition diagnostic.
	if err := Build(C(lowered), binary, Options{Split: true, Jobs: 2}); err != nil {
		t.Fatal(err)
	}
	got, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("native: %v\ngot %q\nNode %q", err, got, want)
	}
}
