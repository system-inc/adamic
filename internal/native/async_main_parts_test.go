package native

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
)

func TestAsyncMainPastPartLimitStaysWhole(t *testing.T) {
	t.Parallel()
	const statements = mainLines + 1
	program := &ir.Program{Source: "async-main.a", Strings: []string{"line"}}
	program.Main = append(program.Main, ir.Evaluate{Value: ir.PromiseValue{}})
	for range statements {
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: ir.StringConstant{Index: 0}})
	}
	if !program.HasPromises() {
		t.Fatal("missing promise")
	}
	// The same independent statements without a promise really do cross a cut.
	control := *program
	control.Main = program.Main[1:]
	if !strings.Contains(C(&control), "void adamic_initialize_chunk_") {
		t.Fatal("synchronous control was not outlined into initializer chunks")
	}
	code := C(program)
	if strings.Contains(code, "static void adamic_main_") {
		t.Fatal("promise module was split into main parts")
	}
	_, main, ok := strings.Cut(code, "int main(int argc, char **argv) {\n")
	if !ok || strings.Count(main, "\n") <= mainLines {
		t.Fatal("promise main does not exceed mainLines")
	}
	if !strings.Contains(main, "adamic_async_run();") {
		t.Fatal("promise main lost the scheduler")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	node := exec.CommandContext(ctx, "node", "--input-type=module")
	node.Stdin = strings.NewReader("Promise.resolve();\n" + strings.Repeat("console.log('line');\n", statements))
	want, err := node.CombinedOutput()
	if err != nil || string(want) != strings.Repeat("line\n", statements) {
		t.Fatalf("Node: %v %s", err, want)
	}
	got := runMainParts(t, code)
	if got.Exit != 0 || got.Stderr != "" || got.Stdout != string(want) {
		t.Fatalf("promise main differs from Node: exit %d, stderr %s", got.Exit, got.Stderr)
	}
}
