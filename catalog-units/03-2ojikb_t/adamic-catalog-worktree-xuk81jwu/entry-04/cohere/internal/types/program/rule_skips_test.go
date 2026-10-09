package program_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/typescript"
	"github.com/system-inc/cohere/internal/types/program"
)

// An enabled rule that declines a file on a compiler option says so, once per file, and a replaying walk
// counts the same skips an uncached one does, so --coverage names the rule and the reason warm and cold
// (#pa7k7zv). no-useless-default-assignment declined every file in ahra and www while every run read
// green (#6ar414z); here the project turns strict off, which is the case that rule really declines.
func TestARuleThatDeclinesOnACompilerOptionNotesASkipPerFile(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json": strings.Replace(minimalConfig, `"strict": true`, `"strict": false`, 1),
		"defaults.ts":   "export function withDefault(value: string | undefined = undefined): string { return value ?? ''; }\n",
		"other.ts":      "export const other = 1;\n",
	})
	rules := []rule.Rule{typescript.NoUselessDefaultAssignment}
	skip := rule.SkippedNotePrefix + "strictNullChecks is off"
	want := map[string]program.RuleNotes{}
	for _, name := range []string{"defaults.ts", "other.ts"} {
		want[filepath.Join(root, name)] = program.RuleNotes{typescript.NoUselessDefaultAssignment.Name: {skip: 1}}
	}

	cold := plainWalk(t, root, rules)
	if !reflect.DeepEqual(cold.Notes, want) {
		t.Errorf("an uncached walk noted %v, want a skip on each file: %v", cold.Notes, want)
	}
	if len(cold.Diagnostics) != 0 {
		t.Errorf("a declined file reported %v", cold.Diagnostics)
	}

	_, recorded := walkAndRecord(t, root, rules, nil)
	warm, _ := walkAndRecord(t, root, rules, recorded)
	if warm.FilesReplayed == 0 {
		t.Fatal("nothing replayed, so the warm count below proves nothing")
	}
	if !reflect.DeepEqual(warm.Notes, cold.Notes) {
		t.Errorf("a replaying walk counted %v, an uncached one %v", warm.Notes, cold.Notes)
	}
}

// The same rule in a strict project judges the file and notes no skip, so a skip is the option's doing.
func TestTheSameRuleInAStrictProjectNotesNoSkip(t *testing.T) {
	t.Parallel()
	root := writeProject(t, map[string]string{
		"tsconfig.json": minimalConfig,
		"defaults.ts":   "export function withDefault(value: string | undefined = undefined): string { return value ?? ''; }\n",
	})
	result := plainWalk(t, root, []rule.Rule{typescript.NoUselessDefaultAssignment})
	if len(result.Notes) != 0 {
		t.Errorf("a strict project noted %v", result.Notes)
	}
	if len(result.Diagnostics) != 1 {
		t.Errorf("%d findings, want the one `= undefined`: %v", len(result.Diagnostics), result.Diagnostics)
	}
}
