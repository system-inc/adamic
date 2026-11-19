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
	"github.com/system-inc/adamic/internal/native_runtime"
)

func constantTableWitness(t *testing.T, source, template string, split, sanitize bool) string {
	return constantTableConfiguredWitness(t, source, template, split, sanitize, nil)
}

func constantTableConfiguredWitness(t *testing.T, source, template string, split, sanitize bool, configure func(*ir.Program)) string {
	t.Helper()
	directory := t.TempDir()
	entry := filepath.Join(directory, "tables.a")
	if err := os.WriteFile(entry, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program := namedProgram(t, entry, nil)
	if configure != nil {
		configure(program)
	}
	code := C(program)
	if C(program) != code {
		t.Fatal("constant emission is not deterministic")
	}
	if template != "" && !strings.Contains(code, template) {
		t.Fatal("missing constant template " + template)
	}
	binary := filepath.Join(directory, "tables")
	if err := Build(code, binary, native_runtime.Options{Split: split, Sanitize: sanitize, Count: true, Jobs: 1, ProgramRegion: program.ProgramRegion}); err != nil {
		t.Fatal(err)
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
	if program.ProgramRegion && n(6) == 0 {
		t.Fatal("Program fixture adopted no storage")
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
	constantTableWitness(t, constantFixture(t), "static const double adamic_numbers_", false, true)
}
func TestConstantTablesSplitCounted(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const adamic_value adamic_record_table_", true, false)
}
func TestConstantTablesWholeCounted(t *testing.T) {
	t.Parallel()
	constantTableWitness(t, constantFixture(t), "static const double adamic_numbers_", false, false)
}
func TestConstantNumbersFreshAndSignedZero(t *testing.T) {
	t.Parallel()
	padding := strings.Repeat(",0", constantNumbers-4)
	constantTableWitness(t, `function values():number[]{return [1,-0,-2,3`+padding+`];}const a=values();const b=values();a[0]=99;console.log(a.join(',')+':'+b.join(',')+':'+(a===b).toString());console.log((1/(b[1]??1)).toString());console.log([1e400,-1e400,5e-324,1.7976931348623157e308].join(','));`, "static const double adamic_numbers_", true, true)
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
	constantTableWitness(t, `let calls=0;function next():number{calls+=1;return calls;}const base=[1,2];console.log([next(),next(),3].join(',')+':'+calls.toString());console.log([...base,next()].join(','));`, "adamic_numbers_", true, true)
}
func TestConstantNumbersKeepPlacement(t *testing.T) {
	t.Parallel()
	for _, mode := range []struct{ graph, program bool }{{true, false}, {false, true}, {true, true}} {
		e := &emitter{program: &ir.Program{GraphTypes: ir.SortedMap[int, bool]{7: true}}}
		types := []int(nil)
		if mode.graph {
			types = []int{7}
		}
		_, ok := e.constantArray(ir.ArrayLiteral{Element: ir.Number, Elements: []ir.Expression{ir.NumberConstant{Value: 1}, ir.NumberConstant{Value: 2}}, GraphTypes: types, ProgramRegion: mode.program})
		if !ok {
			t.Fatal("no numeric template")
		}
		text := e.out.String()
		if strings.Contains(text, "adamic_graph_adopt_owned") != (mode.graph && !mode.program) || strings.Contains(text, "adamic_program_adopt_owned") != mode.program {
			t.Fatalf("wrong adoption graph=%t program=%t: %s", mode.graph, mode.program, text)
		}
	}
}

func TestConstantTablesLoopFreshness(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/constant_tables_fresh.a")
	if err != nil {
		t.Fatal(err)
	}
	constantTableWitness(t, string(source), "adamic_record_table_", true, true)
}

// Numeric literals cannot form a cycle by themselves. Annotate the allocation
// site directly to exercise the same graph and Program metadata the emitter reads.
func constantNumbersAdoptionWitness(t *testing.T, programMode bool) {
	t.Helper()
	code := constantTableConfiguredWitness(t, `const values=[1,2,3];values[0]=9;console.log(values.join(','));`, "adamic_numbers_", true, true, func(p *ir.Program) {
		const identity = 1 << 29
		p.GraphTypes = ir.SortedMap[int, bool]{identity: true}
		p.ProgramRegion = programMode
		if programMode {
			p.ProgramTypes = ir.SortedMap[int, bool]{identity: true}
		}
		marked := 0
		annotate := func(v ir.Expression) ir.Expression {
			if a, ok := v.(ir.ArrayLiteral); ok && a.Element == ir.Number {
				a.GraphTypes = []int{identity}
				a.ProgramRegion = programMode
				marked++
				return a
			}
			return v
		}
		for i, statement := range p.Main {
			switch s := statement.(type) {
			case ir.Declare:
				s.Value = annotate(s.Value)
				p.Main[i] = s
			case ir.Assign:
				s.Value = annotate(s.Value)
				p.Main[i] = s
			}
		}
		if marked != 1 {
			t.Fatalf("annotated %d numeric literals, want 1", marked)
		}
	})
	if strings.Contains(code, " = adamic_graph_adopt_owned(") == programMode || strings.Contains(code, " = adamic_program_adopt_owned(") != programMode {
		t.Fatal("wrong allocation-site adoption")
	}
}
func TestConstantNumbersGraphASan(t *testing.T) {
	t.Parallel()
	constantNumbersAdoptionWitness(t, false)
}

// Not parallel: it sets ADAMIC_PROGRAM_REGION with t.Setenv, which a parallel test may not.
func TestConstantNumbersProgramASan(t *testing.T) {
	t.Setenv("ADAMIC_PROGRAM_REGION", "1")
	constantNumbersAdoptionWitness(t, true)
}

func TestConstantDataSharedAcrossPaths(t *testing.T) {
	t.Parallel()
	e := &emitter{program: &ir.Program{}}
	numbers := []ir.Expression{}
	for range 16 {
		numbers = append(numbers, ir.NumberConstant{Value: 1}, ir.NumberConstant{Value: 2})
	}
	if _, ok := e.constantArray(ir.ArrayLiteral{Element: ir.Number, Elements: numbers}); !ok {
		t.Fatal("flat template not selected")
	}
	rows := []ir.Expression{}
	offsets := []string{"0"}
	for i := 0; i < 16; i++ {
		rows = append(rows, ir.ObjectLiteral{Fields: []ir.Field{{Name: "items", Value: ir.ArrayLiteral{Element: ir.Number, Elements: numbers[:2]}}}})
		offsets = append(offsets, strconv.Itoa((i+1)*2))
	}
	e.lengthTable(offsets)
	if _, ok := e.constantRecords(ir.ArrayLiteral{Element: ir.Object, Elements: rows}); !ok {
		t.Fatal("record template not selected")
	}
	declarations := strings.Join(e.declarations, "\n")
	if strings.Count(declarations, "static const double adamic_numbers_") != 1 || strings.Count(declarations, "static const size_t adamic_lengths_") != 1 {
		t.Fatal("identical immutable data has multiple names")
	}
}

func TestConstantNumbersSmallFunctionsInline(t *testing.T) {
	t.Parallel()
	zeros := func(n int) string { return strings.TrimSuffix(strings.Repeat("0,", n), ",") }
	source := fmt.Sprintf(`function pair():number[]{return [0,0];}function below():number[]{return [%s];}function boundary():number[]{return [%s];}const moduleValues=[7,8];console.log(pair().length.toString()+':'+below().length.toString()+':'+boundary().length.toString()+':'+moduleValues.join(','));`, zeros(constantNumbers-1), zeros(constantNumbers))
	var program *ir.Program
	code := constantTableConfiguredWitness(t, source, "adamic_numbers_", true, true, func(p *ir.Program) { program = p })
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	names := &emitter{program: program}
	seen := 0
	for i, f := range program.Functions {
		count, check := map[string]int{"pair": 2, "below": constantNumbers - 1, "boundary": constantNumbers}[f.Name]
		if !check {
			continue
		}
		for _, d := range declarations {
			if !d.function || d.name != names.functionName(i) {
				continue
			}
			body := code[d.tokens[d.body].end:d.tokens[len(d.tokens)-1].start]
			templated := strings.Contains(body, "adamic_initialize_number_table_copy_")
			if count < constantNumbers {
				if templated || strings.Count(body, "adamic_array_push(") != count {
					t.Fatalf("%s must use %d inline pushes", f.Name, count)
				}
			} else if !templated {
				t.Fatal("threshold literal lost its template")
			}
			seen++
		}
	}
	if seen != 3 {
		t.Fatalf("checked %d functions, want 3", seen)
	}
}

// constantRecordRows is n closed records of one layout, as a literal's source text.
func constantRecordRows(n int) string {
	rows := make([]string, n)
	for index := range rows {
		rows[index] = fmt.Sprintf("{value:%d,done:false}", index)
	}
	return "[" + strings.Join(rows, ",") + "]"
}

// constantFunctionBodies is each named function's emitted C body, its heap version and, for a
// fresh function, its region version (regionFunctionName).
func constantFunctionBodies(t *testing.T, code string, program *ir.Program, name string) (heap, region string) {
	t.Helper()
	declarations, err := splitDeclarations(code)
	if err != nil {
		t.Fatal(err)
	}
	names := &emitter{program: program}
	for index, function := range program.Functions {
		if function.Name != name {
			continue
		}
		for _, d := range declarations {
			if !d.function {
				continue
			}
			body := code[d.tokens[d.body].end:d.tokens[len(d.tokens)-1].start]
			switch d.name {
			case names.functionName(index):
				heap = body
			case names.regionFunctionName(index):
				region = body
			}
		}
	}
	return heap, region
}

// A function's literal of closed records batches into a table and one bounded loop, as a module's
// does, so a large one isn't thousands of stores in one function (#j2qkavx: objects.a's run was
// 10,296 lines, 104 s of clang for wasm32). Every record is fresh: the witness holds it to Node and
// the JavaScript backend under ASan with leak checking, counts balanced, and the mutation shows one
// call's records aren't another's.
func TestConstantRecordsInFunctions(t *testing.T) {
	t.Parallel()
	source := "function rows():{value:number,done:boolean}[]{return " + constantRecordRows(24) + ";}" +
		"const first=rows();const second=rows();const head=first[3];if(head!==undefined){head.done=true;head.value=99;}" +
		"let total=0;let done=0;for(const row of first){total+=row.value;if(row.done){done+=1;}}for(const row of second){total+=row.value;if(row.done){done+=1;}}" +
		"console.log(total.toString()+':'+done.toString()+':'+first.length.toString());"
	var program *ir.Program
	code := constantTableConfiguredWitness(t, source, "adamic_record_table_", true, true, func(p *ir.Program) { program = p })
	heap, _ := constantFunctionBodies(t, code, program, "rows")
	if !strings.Contains(heap, "adamic_initialize_record_table_copy_") {
		t.Fatal("a function's closed records kept one store a field: rows doesn't call the table copy")
	}
	if strings.Count(heap, "adamic_object_new(") != 0 {
		t.Fatal("rows still makes its records one at a time")
	}
}

// A fresh function's region version makes what it returns in the region it was handed, and
// constantRecords never runs there (Emit on #j2qkavx): the region version of build, whose literal
// holds 20 closed records, keeps ordinary emission while its heap version batches. The emitter-state
// probes pin each condition alone. Mutant (internal/native/testdata/constant-records-region.mutants.md):
// the region refusal dropped, and both the probes and the region version's body go red. No run can see
// the refusal today, since an array literal is never region-made, so this body check is the guard: if
// region.go ever makes array literals in a region, it is what fails first (Emit's should on #j2qkavx).
func TestConstantRecordsStayOutOfRegions(t *testing.T) {
	t.Parallel()
	literal := ir.ArrayLiteral{Element: ir.Object}
	for index := range 16 {
		literal.Elements = append(literal.Elements, ir.ObjectLiteral{Fields: []ir.Field{{Name: "value", Value: ir.NumberConstant{Value: float64(index)}}}})
	}
	for _, probe := range []struct {
		name  string
		state emitter
		want  bool
	}{
		// constantRecords reads no function, so these probe the region state alone; the source halves of
		// this test and TestConstantRecordsInFunctions pin a real function's emission.
		{"no region state", emitter{}, true},
		{"a fresh function's region version", emitter{inRegion: true}, false},
		{"a literal handed a region", emitter{regionLiteralDepth: 1}, false},
	} {
		state := probe.state
		state.program = &ir.Program{}
		if _, ok := state.constantRecords(literal); ok != probe.want {
			t.Errorf("%s: constantRecords took the table %t, want %t", probe.name, ok, probe.want)
		}
	}
	source := "function build():{rows:{value:number,done:boolean}[]}{return {rows:" + constantRecordRows(20) + "};}" +
		"function count(holder:{rows:{value:number,done:boolean}[]}):number{return holder.rows.length;}" +
		"let total=0;for(let round=0;round<3;round+=1){total+=count(build());}console.log(total.toString());"
	var program *ir.Program
	code := constantTableConfiguredWitness(t, source, "adamic_record_table_", true, true, func(p *ir.Program) { program = p })
	heap, region := constantFunctionBodies(t, code, program, "build")
	if region == "" {
		t.Fatal("build has no region version, so this pin holds nothing: count(build()) must hand it a region")
	}
	if strings.Contains(region, "adamic_initialize_record_table_copy_") {
		t.Fatal("build's region version batched its records into the table")
	}
	if !strings.Contains(heap, "adamic_initialize_record_table_copy_") {
		t.Fatal("build's heap version kept one store a field")
	}
}
