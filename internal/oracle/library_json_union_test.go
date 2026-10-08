package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{
		"internal/oracle/testdata/json_stringify_unions.a", true, false,
	})
}

// Boxing a number as text is memory-safe and completes normally. Only source Node
// can decide that the selected union arm must serialize as a JSON number.
func TestJSONStringifyOracleCatchesWrongUnionArm(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "mutant.a")
	source := `function show(value: string | number | null): void {
        console.log(JSON.stringify(value) ?? 'missing');
    }
    show(42); show('built'.repeat(2)); show(null);`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	call := program.Functions[0].Body[0].(ir.WriteLine).Value.(ir.Coalesce).Value.(ir.JSONStringify)
	number := len(program.Strings)
	program.Strings = append(program.Strings, "number")
	call.Value = ir.Conditional{
		Condition: ir.Binary{Operator: ir.Equal, Left: ir.TypeOf{Value: call.Value}, Right: ir.StringConstant{Index: number}},
		WhenTrue:  ir.Box{Value: ir.NumberToString{Value: ir.Narrow{Value: call.Value, To: ir.Number}}},
		WhenNot:   call.Value,
	}
	write := program.Functions[0].Body[0].(ir.WriteLine)
	missing := write.Value.(ir.Coalesce)
	missing.Value = call
	write.Value = missing
	program.Functions[0].Body[0] = write
	native, binary := natively(t, program)
	if native.exitCode != 0 || len(native.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly: exit %d stderr %q", native.exitCode, native.stderr)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatalf("mutant leaked: %s", report)
	}
	if difference := disagreement(onNode(t, path), native); difference != "stdout differs" {
		t.Fatalf("want Node alone to catch the wrong arm, got %q", difference)
	}
	t.Log("wrong union arm compiled, exited 0, leaked nothing; only Node caught stdout differs")
}
