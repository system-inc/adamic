package oracle

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/library_json_parse_diagnostics.a",
		"internal/oracle/testdata/library_json_parse_values.a",
		"internal/oracle/testdata/library_json_parse_reviver.a",
		"internal/oracle/testdata/library_json_parse_deep.a",
		"internal/oracle/testdata/library_json_stringify_origin.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

func TestJSONParseRefusalBoundaries(t *testing.T) {
	cases := []struct{ name, source, reason string }{
		{"parsed_replacer", `JSON.stringify({a: 1}, JSON.parse('["a"]'));`, "array identity or callability"},
		{"optional_replacer", `function replacer(flag: boolean): (() => number) | undefined { return flag ? () => 7 : undefined; } JSON.stringify(1, replacer(true));`, "array identity or callability"},
		{"rebound_object", `let obj = {a: 1}; obj = {a: 2}; JSON.stringify(obj);`, "structural types can hide fields"},
		{"narrow_reviver", `JSON.parse('2', (key: string, value: 1) => value);`, "every visit"},
		{"narrow_replacer", `JSON.stringify(2, (key: string, value: 1) => value);`, "proven scalar input"},
		{"optional_toJSON", `function value(flag: boolean): (() => number) | undefined { return flag ? () => 7 : undefined; } JSON.stringify({toJSON: value(true)});`, "toJSON without proven callability"},
		{"escaping_nested", `const obj = {a: {n: 1}}; const alias = obj.a; JSON.stringify(obj);`, "structural types can hide fields"},
		{"shorthand_escape", `const obj = {a: {n: 1}}; const alias = {obj}; JSON.stringify(obj);`, "structural types can hide fields"},
		{"array_return_view", `const fn: () => readonly (number | string)[] = () => [1]; JSON.stringify({toJSON: fn});`, "literal element metadata"},
		{"parsed_descendant_space", `JSON.stringify({p: JSON.parse('{"x":1}')}, null, 2);`, "full runtime metadata"},
		{"parsed_descendant_keys", `JSON.stringify({p: JSON.parse('{"x":1,"y":2}')}, ['p', 'x']);`, "full runtime metadata"},
		{"context", `const value: string = JSON.parse('1');`, "incompatible with its contextual type"},
		{"literal_context", `const value: 1 = JSON.parse('2');`, "incompatible with its contextual type"},
		{"dynamic", `const text = '1'.repeat(2); const value: number = JSON.parse(text);`, "result's type can't be proven"},
		{"nullable_reviver", `const value: string = JSON.parse('"x"', (key: string): string | undefined => key === '' ? undefined : 'x');`, "possibly undefined"},
		{"holder", `JSON.parse('{"k":1}', (key: string, value: number) => value);`, "every visit"},
		{"dynamic_reviver", `JSON.parse('[1]'.repeat(1), (): void => {});`, "literal nesting bound"},
		{"void_reviver", `const fn: () => void = () => 7; JSON.parse('1', fn);`, "actual return is undefined"},
		{"void_replacer", `const fn: () => void = () => 7; JSON.stringify(1, fn);`, "actual return is undefined"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "refused.a")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = lower.Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want compile refusal %q, got %v", test.reason, err)
			}
		})
	}
}

// Check normal termination and leaks before the Node comparison. Temporary runtime mutants
// must fail here solely because stdout differs, rather than a crash, sanitizer or ownership bug.
func TestJSONRuntimeNodeComparison(t *testing.T) {
	for _, fixture := range []string{"parse_diagnostics", "parse_values", "parse_reviver", "stringify_origin"} {
		t.Run(fixture, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_json_"+fixture+".a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			native, binary := natively(t, program)
			if native.exitCode != 0 || len(native.stderr) != 0 {
				t.Fatalf("native must finish cleanly: exit %d stderr %q", native.exitCode, native.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Fatalf("leak check: %s", report)
			}
			t.Log("native compiled, exited zero, stderr empty, LeakSanitizer clean")
			if difference := disagreement(onNode(t, path), native); difference != "" {
				t.Fatalf("independent Node comparison: %s", difference)
			}
		})
	}
}
