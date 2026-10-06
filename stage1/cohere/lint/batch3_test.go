package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

var batch3Controls = []struct{ rule, source, options string }{
	{"one-var", "var first=1; var second=2; var third; {let local=1; let other=2;} function nested(){var inside; var more;}\n", ""},
	{"one-var", "export declare let first, second, third; const one=1, two=2; for(let first=0, second=1; first<2; first++){}", `{"Var":{"Uninitialized":"Never","Initialized":"Never"},"Let":{"Uninitialized":"Never","Initialized":"Never"},"Const":{"Uninitialized":"Never","Initialized":"Never"}}`},
	{"nexus/consistency-no-ambiguous-identifier", "const n=1; const _=2; const x=3; items.sort((a,b)=>a-b); try{work();}catch(e){report(e);} const handleClick=(e)=>run(e); this.e; external.e;", ""},
	{"nexus/consistency-no-abbreviated-identifier", "const ctxRequest=1, durationMs=2, originalInitCwd=3, maxAgentsBytes=4; interface ExampleProperties {maxBootBytes:number} function invoke(...args:string[]){use(args);} const props=1;", ""},
	{"@typescript-eslint/no-non-null-assertion", "foo!; foo! /* . comment */ .bar; foo![key]; foo!(); foo!?.bar; foo!.bar=1; delete foo!.bar; [foo!.bar]; let value!:number;", ""},
	{"@typescript-eslint/prefer-enum-initializers", "enum Direction {Left=10, Right, 'quoted', Other=2, Last} const enum OtherDirection {Up,Down}\n", ""},
	{"nexus/consistency-no-long-line-comment", "// alpha\n// beta\n// gamma\n// delta\n// epsilon\nlet value=1;\n  // one\n  // two\n  // three\n  // four\n  // five\n", ""},
	{"nexus/consistency-no-long-line-comment", "// one\n// two\n// @ts-ignore\n// three\n// four\n// five\n", `{"MaximumLineCount":2}`},
	{"nexus/consistency-no-multiline-arrow-function", "const calculate=(value:number)=>{\n return value+1;\n}; React.useMemo(()=>({value:1})); target.addEventListener('click',()=>work()); const inherited=()=>{\n return arguments;\n};", ""},
	{"prefer-destructuring", "const foo=object.foo; let index=array[0]; foo=object.foo; const same=(receiver.foo); const other=object['other']; const comments=object./* keep */comments; const wrapped=(left,right).wrapped;", ""},
	{"nexus/consistency-no-single-line-jsdoc", "/** Describes this. */\nconst value=1;\n/** https://example */\n/** Inline. */call();\n/**\n * Multiline.\n */\n", ""},
	{"nexus/consistency-no-shouting", "// NEVER DO THIS REALLY LOUDLY PLEASE\n// JSON USD README are allowed\n// `NEVER` \"REALLY\" 'DO NOT' TODO\n/*\n * @example\n * NEVER SHOUT\n * @returns REALLY important\n */\n", ""},
	{"nexus/consistency-no-shouting", "// NEVER REALLY LOUDLY PLEASE\n", `{"Allow":["NEVER","REALLY"]}`},
}

func batch3Generated(t *testing.T) []string {
	t.Helper()
	var rows []string
	controls := append([]struct{ rule, source, options string }(nil), batch3Controls...)
	for _, ending := range []string{"\n", "\r\n", "\r", "\u2028", "\u2029"} {
		controls = append(controls, struct{ rule, source, options string }{"nexus/consistency-no-single-line-jsdoc", "/**" + ending + " * Describes this." + ending + " */" + ending + "const value=1;", ""})
		controls = append(controls, struct{ rule, source, options string }{"nexus/consistency-no-long-line-comment", "// one" + ending + "// two" + ending + "// three" + ending + "// four" + ending + "// five" + ending, ""})
	}
	controls = append(controls, struct{ rule, source, options string }{"one-var", "function foo() { var bar, baz; var a = true; var b = false; var c, d;}", `{"Var":{"Uninitialized":"Always","Initialized":"Never"},"SeparateRequires":false}`})
	controls = append(controls, struct{ rule, source, options string }{"nexus/consistency-no-ambiguous-identifier", "const onſHandler=(e)=>run(e); const onKHandler=(e)=>run(e); const ONİ=(e)=>run(e);", ""})
	controls = append(controls, struct{ rule, source, options string }{"nexus/consistency-no-shouting", "/* \n * `é😀NEVER` REALLY\n * 'USERS NEVER' 'NEVER'd users' 'NOT NULL'\n * \"é😀 REALLY\n * NEVER\" LOUDLY\n */", ""})
	controls = append(controls, struct{ rule, source, options string }{"one-var", "var first /* = misleading */: number; var second: number; var third = 1; var fourth = 2;", `{"Var":{"Uninitialized":"Always","Initialized":"Consecutive"}}`})
	controls = append(controls, struct{ rule, source, options string }{"nexus/consistency-no-ambiguous-identifier", "const {n /* : misleading */ = 1, e: named} = value; const typed /* = */: number = 1;", ""})
	for index, control := range controls {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("control-%d.ts", index))
		if err := os.WriteFile(path, []byte(control.source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+control.rule+"\t\t\t\t"+control.options)
	}
	return rows
}

// Not parallel: native builds are bounded and sanitizer checks precede timing.
func TestBatch3Controls(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	compare(t, goOracle(t), buildPort(t, directory, true), directory, manifest(t, batch3Generated(t)))
}

// Each dispatch mutant executes normally, then loses its selected family's
// positive-control findings. Compile errors and sanitizer failures do not count.
func TestBatch3Mutants(t *testing.T) {
	path := manifest(t, batch3Generated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	seen := map[string]bool{}
	for _, control := range batch3Controls {
		if seen[control.rule] {
			continue
		}
		seen[control.rule] = true
		t.Run(control.rule, func(t *testing.T) {
			targets := map[string]string{
				"one-var": "rules/one_var.ts",
				"nexus/consistency-no-ambiguous-identifier":     "rules/no_ambiguous_identifier.ts",
				"nexus/consistency-no-abbreviated-identifier":   "rules/no_abbreviated_identifier.ts",
				"@typescript-eslint/no-non-null-assertion":      "rules/no_non_null_assertion.ts",
				"@typescript-eslint/prefer-enum-initializers":   "rules/prefer_enum_initializers.ts",
				"nexus/consistency-no-single-line-jsdoc":        "rules/no_single_line_jsdoc.ts",
				"nexus/consistency-no-long-line-comment":        "rules/no_long_line_comment.ts",
				"nexus/consistency-no-shouting":                 "rules/no_shouting.ts",
				"nexus/consistency-no-multiline-arrow-function": "rules/no_multiline_arrow_function.ts",
				"prefer-destructuring":                          "rules/prefer_destructuring.ts",
			}
			from := "context.enabled('" + control.rule + "')"
			directory := mutant(t, from, "context.enabled('omitted-batch3-rule')", targets[control.rule])
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s survived on %s", control.rule, side.name)
				}
				t.Logf("%s caught on %s: %s", control.rule, side.name, difference(side.run.output, want))
			}
		})
	}
}

func TestJsxParserGap(t *testing.T) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/jsx_parser.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_jsx_parser_gap.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(directory, "oracle")
	execute(t, root, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
	if answer := execute(t, "", oracle); string(answer.output) != "true\n" {
		t.Fatalf("Go JSX answer: %q", answer.output)
	}
	path, err := filepath.Abs("gaps/7_jsx_parse.ts")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "gap")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, answer := range []execution{execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, path), execute(t, "", binary)} {
		if string(answer.output) != "false\n" {
			t.Fatalf("JSX parser gap changed: %q", answer.output)
		}
	}
	t.Log("Go TSX parser prints true; Node and sanitized native stage1 parser print false")
}

// These mutants leave the legacy first repair record intact. The new complete
// protocol must catch a changed second edit and a missing third suggestion.
func TestBatch3RepairMutants(t *testing.T) {
	path := manifest(t, batch3Generated(t))
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, change := range []struct{ name, file, from, to string }{
		{"second optional-chain edit changed", "rules/no_non_null_assertion.ts", "edits.push(new RepairEdit(dot, dot + 1, '?.'));", "edits.push(new RepairEdit(dot, dot + 1, '!!'));"},
		{"third enum suggestion omitted", "rules/prefer_enum_initializers.ts", "for(const value of values)", "for(const value of values.slice(0, 2))"},
		{"fix cycle leaves partial source", "lint.ts", "return this.source;", "return current;"},
	} {
		t.Run(change.name, func(t *testing.T) {
			directory := mutant(t, change.from, change.to, change.file)
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("%s survived on %s", change.name, side.name)
				}
				t.Logf("%s caught on %s: %s", change.name, side.name, difference(side.run.output, want))
			}
		})
	}
}
