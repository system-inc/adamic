package lint

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var batch4Names = []string{
	"default-case-last",
	"default-param-last",
	"for-direction",
	"guard-for-in",
	"max-classes-per-file",
	"max-depth",
	"max-lines",
	"max-nested-callbacks",
	"grouped-accessor-pairs",
	"@typescript-eslint/default-param-last",
	"@typescript-eslint/ban-tslint-comment",
	"@typescript-eslint/init-declarations",
	"base/boundary-no-global-container",
	"nexus/consistency-no-stuttering-name",
	"base/consistency-no-hand-built-declared-error",
	"base/consistency-require-pagination-argument-name",
	"base/correctness-require-orm-column-declare",
	"nexus/consistency-no-for-in",
	"nexus/consistency-no-screaming-snake-case",
	"nexus/consistency-no-utils-folder",
}
var batch4Files = []string{
	"default_case_last.ts",
	"default_param_last.ts",
	"for_direction.ts",
	"guard_for_in.ts",
	"max_classes_per_file.ts",
	"max_depth.ts",
	"max_lines.ts",
	"max_nested_callbacks.ts",
	"grouped_accessor_pairs.ts",
	"ts_default_param_last.ts",
	"ts_ban_tslint_comment.ts",
	"ts_init_declarations.ts",
	"base_boundary_no_global_container.ts",
	"nexus_consistency_no_stuttering_name.ts",
	"base_consistency_no_hand_built_declared_error.ts",
	"base_consistency_require_pagination_argument_name.ts",
	"base_correctness_require_orm_column_declare.ts",
	"nexus_consistency_no_for_in.ts",
	"nexus_consistency_no_screaming_snake_case.ts",
	"nexus_consistency_no_utils_folder.ts",
}

func init() {
	portFiles = append(portFiles, batch4Files...)
	portFiles = append(portFiles, "batch4_registry.ts", "batch4_helpers.ts", "batch4_messages.ts")
}
func batch4Selected(name string) bool {
	for _, selected := range batch4Names {
		if name == selected {
			return true
		}
	}
	return false
}
func captureBatch4(t *testing.T, root, overlay string) {
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/core", "-run", "Test(DefaultCaseLast|DefaultParamLast|ForDirection|GuardForIn|MaxClassesPerFile|MaxDepth|MaxLines|MaxNestedCallbacks|GroupedAccessorPairs)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/typescript", "-run", "Test(DefaultParamLast|BanTslintComment|InitDeclarations)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/base", "-run", "Test(NoGlobalContainer|NoHandBuiltDeclaredError|PaginationDecorator|OrmColumnRequiresDeclare)", "-count=1", "-timeout=10m")
	execute(t, root, "go", "test", "-overlay="+overlay, "./internal/lint/rules/nexus", "-run", "Test(ConsistencyNoForIn|ConsistencyNoStutteringName|ConsistencyNoScreamingSnakeCase|ConsistencyNoUtilsFolder)", "-count=1", "-timeout=10m")
}
func batch4Generated(t *testing.T) []string {
	t.Helper()
	cases := []struct{ rule, source, options, file string }{
		{"default-case-last", "switch(x){default: f(); case 1: break;}", "", "source/control.ts"},
		{"default-param-last", "function f(a=1,b){}", "", "source/control.ts"},
		{"for-direction", "for(let i=0;i<10;i--){}", "", "source/control.ts"},
		{"guard-for-in", "for(const key in object) use(key);", "", "source/control.ts"},
		{"max-classes-per-file", "class A{} class B{}", "", "source/control.ts"},
		{"max-depth", "if(a){if(b){if(c){if(d){if(e) f();}}}}", "", "source/control.ts"},
		{"max-lines", "a();\nb();\nc();", "{\"Maximum\":1}", "source/control.ts"},
		{"max-nested-callbacks", "outer(()=>inner(()=>f()));", "{\"Maximum\":1}", "source/control.ts"},
		{"grouped-accessor-pairs", "class C { get a(){} b(){} set a(v){} }", "", "source/control.ts"},
		{"@typescript-eslint/default-param-last", "function f(a=1,b){}", "", "source/control.ts"},
		{"@typescript-eslint/ban-tslint-comment", "// tslint:disable\nlet x=1;", "", "source/control.ts"},
		{"@typescript-eslint/init-declarations", "let x; let y=1;", "{\"Mode\":\"always\"}", "source/control.ts"},
		{"base/boundary-no-global-container", "getGlobalContainer();", "", "source/control.ts"},
		{"nexus/consistency-no-stuttering-name", "outcome.outcome; result.result; outcome?.outcome; user.user;", "", "source/control.ts"},
		{"base/consistency-no-hand-built-declared-error", "new BaseError(\"x\",{identifier:\"Broken\"});", "", "source/control.ts"},
		{"base/consistency-require-pagination-argument-name", "class C{m(@GraphQlArgument(\"bad\") p: PaginationInput){}}", "", "source/control.ts"},
		{"base/correctness-require-orm-column-declare", "class C{@OrmColumn() id: number;}", "", "source/control.ts"},
		{"nexus/consistency-no-for-in", "for(const key in object){}", "", "source/control.ts"},
		{"nexus/consistency-no-screaming-snake-case", "const MAX_RETRY_COUNT=3; export const HTTP_TIMEOUT=5;", "", "source/control.ts"},
		{"nexus/consistency-no-utils-folder", "let x=1;", "", "utils/control.ts"},
	}
	var rows []string
	for _, c := range cases {
		path := filepath.Join(t.TempDir(), c.file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(c.source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+c.rule+"\t\t\tfalse\t"+c.options)
	}
	// Cross-family controls cover option defaults, filename gates and overlapping findings.
	for _, c := range []struct{ source, file, rule, options string }{
		{"if(a) {}", "source/negative.ts", "max-depth", `{"Maximum":-1}`},
		{"payload.payload; result.result;", "source/custom.ts", "nexus/consistency-no-stuttering-name", `{"GenericNames":["payload"]}`},
		{"const EXTERNAL_NAME=1; const EXTERNAL_NAME_OLD=2;", "source/allow.ts", "nexus/consistency-no-screaming-snake-case", `{"allow":["EXTERNAL_NAME"]}`},
		{"L: for(let i=0;i<10;i--) {}", "utils/overlap.ts", "all", `{"Maximum":0}`},
		{"class A {} class B {}", "source/classes.ts", "all", ""},
	} {
		path := filepath.Join(t.TempDir(), c.file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(c.source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, path+"\t"+c.rule+"\t\t\tfalse\t"+c.options)
	}
	return rows
}
func TestBatch4Controls(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	rows := batch4Generated(t)
	for _, row := range rows {
		answer := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
		if string(answer.output) == "0\n" {
			t.Fatalf("inert positive control: %s", row)
		}
	}
	path := manifest(t, rows)
	want := execute(t, "", oracle, "--manifest", path).output
	if diff := difference(node(t, directory, path, false).output, want); diff != "" {
		t.Fatalf("Node preflight: %s", diff)
	}
	compare(t, oracle, buildPort(t, directory, true), directory, manifest(t, rows))
}

// Four isolated sanitized builds fit the available cores and memory.
// Each subtest owns its copied sources and executable.
func TestBatch4Mutants(t *testing.T) {
	rows := batch4Generated(t)
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for i, rule := range batch4Names {
		t.Run(rule, func(t *testing.T) {
			t.Parallel()
			from := "ctx.enabled('" + rule + "')"
			directory := mutant(t, from, "false && "+from, batch4Files[i])
			binary := buildPort(t, directory, true)
			for _, side := range []struct {
				name string
				run  execution
			}{{"Node", node(t, directory, path, false)}, {"native", execute(t, "", binary, "--manifest", path)}} {
				if bytes.Equal(side.run.output, want) {
					t.Fatalf("omission survived on %s", side.name)
				}
				t.Logf("compiled listener omission caught on %s: %s", side.name, difference(side.run.output, want))
			}
		})
	}
}
func TestBatch4UpstreamAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if batch4Selected(fields[1]) {
			rows = append(rows, row)
		}
	}
	t.Logf("batch4 upstream compared: %d", len(rows))
	compare(t, goOracle(t), buildPort(t, directory, true), directory, manifest(t, rows))
}

// Only malformed comments that Go's rule tests deliberately run through recovery.
// Compare findings on these inputs; a fixed parseable source cannot be demanded.
func batch4Recovery(rule, source string) string {
	if rule == "@typescript-eslint/ban-tslint-comment" {
		switch strings.TrimSpace(source) {
		case "/*", "/*/", "/**", "/* tslint:disable":
			return "recovery"
		}
	}

	return ""
}
func TestBatch4NodeUpstreamAgree(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if batch4Selected(fields[1]) && !strings.HasSuffix(row, "\tunsupported-recovery") {
			rows = append(rows, row)
		}
	}
	path := manifest(t, rows)
	want := execute(t, "", goOracle(t), "--manifest", path).output
	got := node(t, directory, path, false).output
	if diff := difference(got, want); diff != "" {
		t.Fatal(diff)
	}
	t.Logf("Go and Node identical: %d bytes", len(want))
}

func TestBatch4RecoveryCensus(t *testing.T) {
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if batch4Selected(fields[1]) {
			rows = append(rows, row)
		}
	}
	result := execute(t, "", goOracle(t), "--manifest", manifest(t, rows), "--parse-errors")
	t.Logf("Go parser recovery census:\n%s", result.output)
}
