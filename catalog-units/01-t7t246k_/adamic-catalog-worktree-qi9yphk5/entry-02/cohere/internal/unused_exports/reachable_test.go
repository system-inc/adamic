package unused_exports

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

// TestUnreachableFires is the half that proves the analysis can report. Each case holds code that
// genuinely cannot run, and the expected line is the first statement of the dead run.
func TestUnreachableFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		want string
	}{
		{
			name: "after return",
			code: "function f() {\n\talpha();\n\treturn;\n\tbeta();\n}",
			want: "beta();",
		},
		{
			name: "after throw",
			code: "function f() {\n\tthrow new Error('x');\n\tbeta();\n}",
			want: "beta();",
		},
		{
			name: "after break",
			code: "function f(x) {\n\twhile (x) {\n\t\tbreak;\n\t\tbeta();\n\t}\n}",
			want: "beta();",
		},
		{
			name: "after continue",
			code: "function f(x) {\n\twhile (x) {\n\t\tcontinue;\n\t\tbeta();\n\t}\n}",
			want: "beta();",
		},
		{
			name: "both branches return",
			code: "function f(x) {\n\tif (x) {\n\t\treturn 1;\n\t} else {\n\t\treturn 2;\n\t}\n\tbeta();\n}",
			want: "beta();",
		},
		{
			name: "at file scope after throw",
			code: "throw new Error('boom');\nconsole.log('never');",
			want: "console.log('never');",
		},
		{
			name: "inside an arrow function",
			code: "const f = () => {\n\treturn 1;\n\tbeta();\n};",
			want: "beta();",
		},
		{
			name: "inside a method",
			code: "class C {\n\tm() {\n\t\treturn;\n\t\tbeta();\n\t}\n}",
			want: "beta();",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceFile := parse(t, testCase.code)
			findings := FindUnreachable(sourceFile)
			if len(findings) != 1 {
				t.Fatalf("got %d findings, want 1: %s", len(findings), describe(testCase.code, findings))
			}
			got := strings.TrimSpace(testCase.code[findings[0].Range.Position:findings[0].Range.End])
			if got != testCase.want {
				t.Fatalf("reported %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestUnreachableStaysSilent is the half that proves the analysis can decline. Every case here is
// live code, and several are shapes a naive "statement after a return" scan reports wrongly.
func TestUnreachableStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
	}{
		{
			name: "return inside a branch leaves the rest reachable",
			code: "function f(x) {\n\tif (x) {\n\t\treturn;\n\t}\n\tbeta();\n}",
		},
		{
			name: "break leaves the code after the loop reachable",
			code: "function f(x) {\n\twhile (x) {\n\t\tbreak;\n\t}\n\tbeta();\n}",
		},
		{
			name: "a hoisted function declaration after a return is callable",
			code: "function f() {\n\treturn helper();\n\tfunction helper() {\n\t\treturn 1;\n\t}\n}",
		},
		{
			name: "a hoisted var after a return still declares",
			code: "function f() {\n\treturn;\n\tvar x;\n}",
		},
		{
			name: "straight line code",
			code: "function f() {\n\talpha();\n\tbeta();\n\tgamma();\n}",
		},
		{
			name: "a loop that continues still runs its body",
			code: "function f(x) {\n\twhile (x) {\n\t\tbeta();\n\t\tcontinue;\n\t}\n}",
		},
		{
			name: "try finally still runs the finally",
			code: "function f() {\n\ttry {\n\t\treturn 1;\n\t} finally {\n\t\tbeta();\n\t}\n}",
		},
		{
			// The duplicate-layout trap. The CFG lays a finally body out twice and the second copy is
			// unreachable when nothing takes the abrupt path, so a per-block scan condemns cleanup
			// code that certainly runs. This fixture is the one that caught it.
			name: "try return finally still runs the finally",
			code: "function f() {\n\ttry {\n\t\treturn 1;\n\t} finally {\n\t\tbeta();\n\t}\n}",
		},
		{
			// Same trap, catch flavour: a catch body is laid out unreachable when the try returns,
			// and it is still reachable code.
			name: "catch body beside a returning try",
			code: "function f() {\n\ttry {\n\t\treturn 1;\n\t} catch (e) {\n\t\tgamma();\n\t} finally {\n\t\tbeta();\n\t}\n}",
		},
		{
			name: "a switch case after a break in another case",
			code: "function f(x) {\n\tswitch (x) {\n\t\tcase 1:\n\t\t\tbreak;\n\t\tcase 2:\n\t\t\tbeta();\n\t}\n}",
		},
		{
			name: "an early return guard",
			code: "function f(x) {\n\tif (!x) return;\n\tbeta();\n\tgamma();\n}",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceFile := parse(t, testCase.code)
			if findings := FindUnreachable(sourceFile); len(findings) != 0 {
				t.Fatalf("reported live code as unreachable: %s", describe(testCase.code, findings))
			}
		})
	}
}

// TestUnreachableReportsOncePerDeadRun asserts a three-statement dead run reports once rather than
// three times. A single mistake should read as one finding.
func TestUnreachableReportsOncePerDeadRun(t *testing.T) {
	t.Parallel()
	sourceFile := parse(t, "function f() {\n\treturn;\n\talpha();\n\tbeta();\n\tgamma();\n}")
	findings := FindUnreachable(sourceFile)
	if len(findings) != 1 {
		t.Fatalf("got %d findings for one dead run, want 1", len(findings))
	}
}

// TestUnreachableHandlesEmptyInput asserts a nil file reports nothing rather than panicking, since
// this runs over a whole tree and one odd file must not end the report.
func TestUnreachableHandlesEmptyInput(t *testing.T) {
	t.Parallel()
	if findings := FindUnreachable(nil); len(findings) != 0 {
		t.Fatalf("a nil file reported %d findings", len(findings))
	}
}

func parse(t *testing.T, code string) *ast.SourceFile {
	t.Helper()
	return parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/unused-fixture.ts",
		PathKey:  "/unused-fixture.ts",
	}, code, core.ScriptKindTS)
}

func describe(code string, findings []Unreachable) string {
	var parts []string
	for _, finding := range findings {
		parts = append(parts, strings.TrimSpace(code[finding.Range.Position:finding.Range.End]))
	}
	return strings.Join(parts, " | ")
}
