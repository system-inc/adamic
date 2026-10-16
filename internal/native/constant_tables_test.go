package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func constantTableWitness(t *testing.T, source, template string, split, sanitize bool) string {
	t.Helper()
	directory := t.TempDir()
	entry := filepath.Join(directory, "tables.a")
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	code := C(program)
	if C(program) != code {
		t.Fatal("constant emission is not deterministic")
	}
	if template != "" && !strings.Contains(code, template) {
		t.Fatal("missing constant template " + template)
	}
	binary := filepath.Join(directory, "tables")
	options := Options{Split: split, Sanitize: sanitize, Count: true, Jobs: 1}
	var buildError error
	if split {
		buildError = Build(code, binary, options)
	} else {
		// This remains a whole-unit control when automatic split compilation is enabled.
		buildError = BuildFlags(code, binary, options, SourceFlags(code, options))
	}
	if buildError != nil {
		t.Fatal(buildError)
	}
	run := func(name string, args ...string) (string, string) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var out, errors bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &errors
		if err := cmd.Run(); err != nil {
			t.Fatalf("%s: %v\n%s", name, err, errors.String())
		}
		return out.String(), errors.String()
	}
	truth, _ := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", entry)
	actual, stderr := run(binary)
	if actual != truth {
		t.Fatalf("native %q Node %q", actual, truth)
	}
	js := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	if actual, _ := run("node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js); actual != truth {
		t.Fatal("JavaScript differs from Node")
	}
	counts := regexp.MustCompile(`allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)`).FindStringSubmatch(stderr)
	if len(counts) != 7 {
		t.Fatalf("missing counts: %s", stderr)
	}
	n := func(i int) int { v, _ := strconv.Atoi(counts[i]); return v }
	if n(1) != n(2)+n(6) {
		t.Fatalf("unbalanced allocations: %s", stderr)
	}
	t.Log(strings.TrimSpace(stderr))
	return code
}
func constantFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/constant_tables.a")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func TestConstantTablesSplitSanitized(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const adamic_value adamic_record_table_", true, true)
}
func TestConstantTablesWholeSanitized(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const double adamic_number_table_", false, true)
}
func TestConstantTablesSplitCounted(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const adamic_value adamic_record_table_", true, false)
}
func TestConstantTablesWholeCounted(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const double adamic_number_table_", false, false)
}
func TestConstantNumbersFreshAndSignedZero(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, `function values():number[]{return [1,-0,-2,3];}const a=values();const b=values();a[0]=99;console.log(a.join(',')+':'+b.join(',')+':'+(a===b).toString());console.log((1/(b[1]??1)).toString());console.log([1e400,-1e400,5e-324,1.7976931348623157e308].join(','));`, "static const double adamic_number_table_", true, true)
}
func TestConstantRecordConstructorsStayBounded(t *testing.T) {
	t.Parallel()
	var fields, rows []string
	for field := 0; field < 24; field++ {
		fields = append(fields, fmt.Sprintf("field%d:%d", field, field))
	}
	for row := 0; row < 16; row++ {
		rows = append(rows, "{"+strings.Join(fields, ",")+"}")
	}
	code := constantTableWitness(t, "import {panic} from 'adamic';const rows=["+strings.Join(rows, ",")+"];console.log((rows[7]??panic('missing')).field23.toString());", "", true, true)
	if strings.Contains(code, "adamic_record_table_") {
		t.Fatal("wide record specialized")
	}
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range declarations {
		if d.function && strings.HasPrefix(d.name, "adamic_initialize_") {
			body := code[d.tokens[d.body].end:d.tokens[len(d.tokens)-1].start]
			if strings.Count(body, "\n") > 132 {
				t.Fatalf("unbounded constructor %s", d.name)
			}
		}
	}
}
func TestConstantRecordsKeepPlacement(t *testing.T) {
	t.Parallel()
	p := &ir.Program{GraphTypes: ir.SortedMap[int, bool]{7: true}}
	e := &emitter{program: p}
	field := ir.Field{Name: "value", Value: ir.NumberConstant{Value: 1}}
	for _, literal := range []ir.ObjectLiteral{{Record: true, Fields: []ir.Field{field}}, {ProgramRegion: true, Fields: []ir.Field{field}}, {GraphTypes: []int{7}, Fields: []ir.Field{field}}, {Fields: []ir.Field{{Name: "value", Uninitialized: true, Value: field.Value}}}, {Fields: []ir.Field{{Name: "items", Value: ir.ArrayLiteral{Element: ir.Number, ProgramRegion: true}}}}} {
		if _, ok := e.constantRecordOf(literal); ok {
			t.Fatalf("specialized placement-sensitive record %+v", literal)
		}
	}
}

func TestConstantEvaluationOrder(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, `let calls=0;function next():number{calls+=1;return calls;}const base=[1,2];console.log([next(),next(),3].join(',')+':'+calls.toString());console.log([...base,next()].join(','));`, "adamic_number_table_", true, true)
}
func TestConstantNumbersKeepPlacement(t *testing.T) {
	t.Parallel()
	e := &emitter{program: &ir.Program{GraphTypes: ir.SortedMap[int, bool]{7: true}}}
	_, ok := e.constantArray(ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 1}, ir.NumberConstant{Value: 2}}, GraphTypes: []int{7}, ProgramRegion: true})
	if !ok {
		t.Fatal("no numeric template")
	}
	s := e.out.String()
	graph, program := strings.Index(s, "adamic_graph_adopt_owned"), strings.Index(s, "adamic_program_adopt_owned")
	if graph < 0 || program < graph {
		t.Fatal("numeric table lost allocation-site adoption")
	}
}
