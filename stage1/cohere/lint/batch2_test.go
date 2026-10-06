package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var batch2Names = []string{
	"no-caller",
	"no-eq-null",
	"no-empty-static-block",
	"no-proto",
	"no-script-url",
	"no-self-compare",
	"no-delete-var",
	"no-iterator",
	"no-compare-neg-zero",
	"no-async-promise-executor",
	"no-empty-pattern",
	"no-constructor-return",
	"no-multi-assign",
	"no-useless-catch",
	"no-unsafe-finally",
	"no-useless-concat",
	"no-empty-character-class",
	"no-ex-assign",
	"@typescript-eslint/no-unnecessary-type-constraint",
	"@typescript-eslint/prefer-namespace-keyword",
}

func init() {
	portFiles = append(portFiles, "no_caller.ts")
	portFiles = append(portFiles, "no_eq_null.ts")
	portFiles = append(portFiles, "no_empty_static_block.ts")
	portFiles = append(portFiles, "no_proto.ts")
	portFiles = append(portFiles, "no_script_url.ts")
	portFiles = append(portFiles, "no_self_compare.ts")
	portFiles = append(portFiles, "no_delete_var.ts")
	portFiles = append(portFiles, "no_iterator.ts")
	portFiles = append(portFiles, "no_compare_neg_zero.ts")
	portFiles = append(portFiles, "no_async_promise_executor.ts")
	portFiles = append(portFiles, "no_empty_pattern.ts")
	portFiles = append(portFiles, "no_constructor_return.ts")
	portFiles = append(portFiles, "no_multi_assign.ts")
	portFiles = append(portFiles, "no_useless_catch.ts")
	portFiles = append(portFiles, "no_unsafe_finally.ts")
	portFiles = append(portFiles, "no_useless_concat.ts")
	portFiles = append(portFiles, "no_empty_character_class.ts")
	portFiles = append(portFiles, "no_ex_assign.ts")
	portFiles = append(portFiles, "no_unnecessary_type_constraint.ts")
	portFiles = append(portFiles, "prefer_namespace_keyword.ts")
	portFiles = append(portFiles, "rule_context.ts", "batch2_registry.ts")
}
func batch2Selected(name string) bool {
	for _, selected := range batch2Names {
		if name == selected {
			return true
		}
	}
	return false
}
func batch2Extension(rule, file string) string {
	if rule == "@typescript-eslint/no-unnecessary-type-constraint" {
		return filepath.Ext(file)
	}
	return ".ts"
}
func batch2Recovery(rule, source string) string {
	if rule == "no-async-promise-executor" && source == "new Promise(@dec async () => {})" {
		return "unsupported-recovery"
	}
	if rule == "no-compare-neg-zero" && (source == "x === -0_0;" || source == "x === -00;") {
		return "recovery"
	}
	return ""
}
func captureBatch2(t *testing.T, root, overlay string) {
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/core", "-run", "Test(NoCaller|NoEqNull|NoEmptyStaticBlock|NoProto|NoScriptUrl|NoSelfCompare|NoDeleteVar|NoIterator|NoCompareNegZero|NoAsyncPromiseExecutor|NoEmptyPattern|NoConstructorReturn|NoMultiAssign|NoUselessCatch|NoUnsafeFinally|NoUselessConcat|NoEmptyCharacterClass|NoExAssign)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/typescript", "-run", "Test(NoUnnecessaryTypeConstraint|PreferNamespaceKeyword)", "-count=1", "-timeout=10m")
}

// One positive control per rule, plus decoded options and clean boundary cases.
func batch2Generated(t *testing.T) []string {
	t.Helper()
	sources := []string{
		"arguments.caller; (arguments).callee; object.callee; arguments[caller];",
		"x == (null); null != x; x === null; x == undefined;",
		"class C { static {} } class D { static { /* intentional */ } }",
		"x.__proto__; x['__proto__']; x[`__proto__`]; x[__proto__];",
		"'JAVASCRIPT:x'; `javascript:`; tag`javascript:`; 'xjavascript:';",
		"foo.bar() >= foo.bar (); x in x; x === y; /[/*]/ === /[/*]/; [] == [ ];",
		"delete (value); delete object.value;",
		"x.__iterator__; x['__iterator__']; x[__iterator__];",
		"x === -(0); -0x0 != x; x === -0n;",
		"new Promise(async () => {}); new Promise(() => {}, async () => {});",
		"const {} = x; const [] = y; function f({} = {}) {} function g({a:{}}) {}",
		"class C { constructor(){ return 1; } method(){ return 2; } \"constructor\"(){ return 3; } *constructor(){ return 4; } static constructor(){ return 5; } }",
		"let a = (b = c); x = y = z; class C { field = value = 0; }",
		"try { f(); } catch(e) { throw (e as Error)!; } try { f(); } catch(e) { throw e; } finally { g(); }",
		"function f(){ try { return 1; } finally { return 2; } } while(x) try {} finally { break; }",
		"foo + 'a' + ('b' + 'c'); 'a' +\n'b'; `a${x}` + 'b';",
		"const pattern=/[a[[]]]/v; const another=/x[]/; const good=/[\\[]/;",
		"try { f(); } catch(e) { e = x; e++; ++(e); e.prop = x; }",
		"function f<T extends any>(){} const g=<T extends unknown>()=>{};",
		"export declare /* module */ module A.B.module { module C {} } namespace D {}",
	}
	if len(sources) != len(batch2Names) {
		t.Fatal("rule controls lost")
	}
	var rows []string
	for i, source := range sources {
		path := filepath.Join(t.TempDir(), "control.ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+batch2Names[i])
		if batch2Names[i] == "no-empty-pattern" {
			rows = append(rows, path+"\tno-empty-pattern\t\t\tfalse\t{\"allowObjectPatternsAsParameters\":true}")
		}
		if batch2Names[i] == "no-multi-assign" {
			rows = append(rows, path+"\tno-multi-assign\t\t\tfalse\t{\"ignoreNonDeclaration\":true}")
		}
	}
	return rows
}

// Not parallel: upstream capture sets process-wide environment state.
func TestBatch2Controls(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := batch2Generated(t)
	for _, row := range rows {
		answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
		if string(answer.output) == "0\n" {
			t.Fatalf("positive control is inert: %s", row)
		}
	}
	compare(t, oracle, buildPort(t, directory, true), directory, manifest(t, rows))
}

// Not parallel at the parent: independent mutant children finish before timing.
func TestBatch2Mutants(t *testing.T) {
	rows := batch2Generated(t)
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for i, rule := range batch2Names {
		t.Run(rule, func(t *testing.T) {
			t.Parallel()
			stem := strings.ReplaceAll(strings.TrimPrefix(rule, "@typescript-eslint/"), "-", "_")
			from := "ctx.enabled('" + rule + "')"
			directory := mutant(t, from, "ctx.enabled('omitted-batch2-rule')", stem+".ts")
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("rule %d listener omission survived on %s", i, side.name)
				}
				t.Logf("listener omission caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}

// Check this batch independently before the combined regression gate.
func TestBatch2UpstreamAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	binary := buildPort(t, directory, true)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if !batch2Selected(fields[1]) {
			continue
		}
		if strings.HasSuffix(row, "\tunsupported-recovery") {
			t.Logf("EXPLICIT LIMIT: parser recovery is not ported for %s", row)
			checkRecoveryRefusal(t, oracle, binary, directory, row)
		} else {
			rows = append(rows, row)
		}
	}
	t.Logf("batch2 upstream compared: %d", len(rows))
	compare(t, oracle, binary, directory, manifest(t, rows))
}
