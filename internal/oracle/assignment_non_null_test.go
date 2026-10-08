package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Keep new Adamic fixtures honest: their explicit nullish repair lowers as .a.
// Generate temporary TypeScript controls for the newly admitted checked syntax.
func TestNonNullAssignmentCompatibility(t *testing.T) {
	for _, shape := range []string{"array_number", "array_string", "field_number", "field_string", "uint8array", "int32array", "float64array", "field_number_accessor", "field_string_accessor"} {
		t.Run(shape, func(t *testing.T) {
			fixture := filepath.Join(repository, "internal/oracle/testdata/assignment_non_null_"+shape+".a")
			bytes, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(bytes), "import { panic } from 'adamic';\n", "", 1)
			start := strings.Index(source, " const held = receiver();")
			end := strings.Index(source[start:], "\n}") + start
			target, operator := "receiver()[index()]!", "|="
			if strings.Contains(shape, "string") {
				operator = "+="
			}
			if strings.HasPrefix(shape, "field_") {
				target = "receiver().item!"
			}
			source = source[:start] + " " + target + " " + operator + " right();" + source[end:]
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			if program.NonNullChecks.Checked != 1 {
				t.Fatalf("want one held target check, got %#v", program.NonNullChecks)
			}
			node := onNode(t, path)
			native, sanitized := natively(t, program)
			for _, result := range []run{native, onJavaScriptBackend(t, program), released(t, program)} {
				if difference := disagreement(node, result); difference != "" {
					t.Fatal(difference)
				}
			}
			if leaked := leaks(t, program, sanitized); leaked != "" {
				t.Fatal(leaked)
			}
		})
	}
}

func TestNonNullAssignmentChecksBeforeRight(t *testing.T) {
	for _, source := range []string{
		"const values: number[]=[]; function right(): number {console.log('right');return 2;} console.log('before'); values[0]! |= right();",
		"const values: string[]=[]; function right(): string {console.log('right');return 'new';} console.log('before'); values[0]! += right();",
		"const value: {item:number|undefined}={item:undefined}; function right(): number {console.log('right');return 2;} console.log('before'); value.item! |= right();",
	} {
		path := filepath.Join(t.TempDir(), "main.ts")
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		node := onNode(t, path)
		if node.exitCode != 0 || !strings.Contains(string(node.stdout), "right\n") {
			t.Fatalf("unchecked Node control did not run right: %#v", node)
		}
		native, _ := natively(t, program)
		for _, result := range []run{native, onJavaScriptBackend(t, program), released(t, program)} {
			if result.exitCode != 70 || string(result.stdout) != "before\n" || !strings.Contains(string(result.stderr), "non-null assertion failed") {
				t.Fatalf("want checked target to stop before right, got %#v", result)
			}
		}
	}
}
