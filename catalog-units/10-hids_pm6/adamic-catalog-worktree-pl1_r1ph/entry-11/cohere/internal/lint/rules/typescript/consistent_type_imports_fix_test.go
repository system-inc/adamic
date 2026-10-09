package typescript

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// applyConsistentTypeImportsFixes applies every fix the rule proposed in one pass, back to front, the
// way ESLint's RuleTester produces `output`. Overlapping fixes from different findings are refused
// rather than resolved, which the harness does too.
func applyConsistentTypeImportsFixes(t *testing.T, result rule_testing.Result) (string, int) {
	t.Helper()
	type pending struct {
		start, end int
		text       string
	}
	fixes := []pending{}
	for _, diagnostic := range result.Diagnostics {
		for _, fix := range diagnostic.Fixes {
			fixes = append(fixes, pending{fix.Range.Pos(), fix.Range.End(), fix.Text})
		}
	}
	sort.Slice(fixes, func(first, second int) bool { return fixes[first].start > fixes[second].start })
	source := result.SourceFile.Text()
	previousStart := len(source)
	for _, fix := range fixes {
		if fix.end > previousStart {
			t.Fatalf("fixes overlap at [%d,%d)", fix.start, fix.end)
		}
		source = source[:fix.start] + fix.text + source[fix.end:]
		previousStart = fix.start
	}
	return source, len(fixes)
}

// runConsistentTypeImportsFix lints one source with the options upstream's case ran with and returns
// the source after the repair, normalised the way RunTyped writes it.
func runConsistentTypeImportsFix(t *testing.T, options string, code string) (string, string, int) {
	t.Helper()
	settings, err := DecodeConsistentTypeImportsOptions([]byte(options))
	if options == "" {
		settings, err = DefaultConsistentTypeImportsOptions(), nil
	}
	if err != nil {
		t.Fatalf("decoding %s: %v", options, err)
	}
	result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports, "consistent_type_imports.ts", code, settings)
	fixed, count := applyConsistentTypeImportsFixes(t, result)
	return strings.TrimSpace(code) + "\n", fixed, count
}

// consistentTypeImportsFixDivergences are the upstream cases whose repair this port writes
// differently on purpose, each with the output it writes instead.
//
// Only one shape diverges: upstream merges moved names into a type-only declaration that already
// exists elsewhere in the file, which is a second statement and so a second edit, and this port's
// repair is one edit inside the reported declaration. See consistent_type_imports_fix.go.
var consistentTypeImportsFixDivergences = map[int]string{
	// Upstream writes `import type { Already1 , B }` and `import type { Already2 , C, D}`, merging into
	// the two existing declarations. This writes each moved group as its own declaration beside the
	// one it came from. Both type-check and mean the same.
	16: `import type Already1Def from 'foo';
import type { Already1 } from 'foo';
import type { B } from 'foo';
import A from 'foo';
import type { C, D} from 'bar';
import { E } from 'bar';
import type { Already2 } from 'bar';
type T = { b: B; c: C; d: D };
`,
}

// TestConsistentTypeImportsFixMatchesUpstream runs every repair in upstream's corpus, 70 cases, and
// requires the result byte for byte, spacing included, except where a divergence above is recorded.
func TestConsistentTypeImportsFixMatchesUpstream(t *testing.T) {
	t.Parallel()

	ran := 0
	for _, testCase := range consistentTypeImportsUpstreamFixes {
		ran++
		t.Run(strings.ReplaceAll(strings.TrimSpace(strings.SplitN(strings.TrimSpace(testCase.code), "\n", 2)[0]), "/", "_"), func(t *testing.T) {
			t.Parallel()
			_, fixed, count := runConsistentTypeImportsFix(t, testCase.options, testCase.code)
			want := strings.TrimSpace(testCase.upstream) + "\n"
			if divergence, recorded := consistentTypeImportsFixDivergences[testCase.index]; recorded {
				want = divergence
			}
			if fixed != want {
				t.Errorf("case %d, %d fixes\n--- got\n%s--- upstream\n%s", testCase.index, count, fixed, want)
			}
		})
	}
	if ran != 70 {
		t.Fatalf("ran %d upstream cases, want 70", ran)
	}
}

// consistentTypeImportsTypeErrors runs the checker once over many sources as one program and
// returns each file's diagnostics, keyed by file name.
func consistentTypeImportsTypeErrors(t *testing.T, sources map[string]string) map[string]map[string]int {
	t.Helper()
	probe := rule.Rule{
		Name:             "test/type-check",
		NeedsTypeChecker: true,
		ProgramReads:     rule.ReadsOtherFiles,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					for _, sourceFile := range ctx.Program.SourceFiles() {
						// The harness writes fixtures under a temporary root, so a file is matched to
						// its fixture by the path the fixture named, as a suffix.
						key := ""
						for name := range sources {
							if strings.HasSuffix(sourceFile.FileName().AsString(), name) {
								key = name
							}
						}
						if key == "" {
							continue
						}
						// The checker this rule already holds, never Program.GetSemanticDiagnostics: the harness
						// holds the file's checker exclusively while a rule runs, and the program-level call waits
						// for that same lock, so it deadlocked and read as a suite too slow to finish.
						for _, diagnostic := range ctx.TypeChecker.GetDiagnostics(context.Background(), sourceFile) {
							ctx.ReportNode(node, rule.Message{
								Id:          key,
								Description: fmt.Sprintf("TS%d %s", diagnostic.Code(), diagnostic.MessageText()),
							})
						}
					}
				},
			}
		},
	}
	var subject string
	for name := range sources {
		subject = name
		break
	}
	result := rule_testing.RunTypedFiles(t, probe, sources, subject)
	byFile := map[string]map[string]int{}
	for _, diagnostic := range result.Diagnostics {
		if byFile[diagnostic.Message.Id] == nil {
			byFile[diagnostic.Message.Id] = map[string]int{}
		}
		byFile[diagnostic.Message.Id][diagnostic.Message.Description]++
	}
	return byFile
}

// consistentTypeImportsNewTypeErrors lists the diagnostics each repaired source has that its original
// did not. Fixture modules do not exist, so "no errors" is the wrong bar; "no new errors" is the one a
// repair must meet, and the error a wrong repair causes is new: TS1361, a name imported with
// `import type` used as a value.
func consistentTypeImportsNewTypeErrors(t *testing.T, originals map[string]string, repaired map[string]string) map[string][]string {
	t.Helper()
	before := consistentTypeImportsTypeErrors(t, originals)
	after := consistentTypeImportsTypeErrors(t, repaired)
	added := map[string][]string{}
	// Compared by distinct diagnostic rather than by count: splitting one declaration into two names the
	// missing fixture module twice, so "cannot find module" (TS2307, and TS2591 for a module named like
	// a Node built-in) legitimately appears once more. Any diagnostic the original never had is new.
	for file, messages := range after {
		for message := range messages {
			if before[file][message] == 0 {
				added[file] = append(added[file], message)
			}
		}
	}
	return added
}

// TestConsistentTypeImportsFixLeavesTheFileTypeChecking runs the checker on every repaired upstream
// case, and requires that the repair introduced no diagnostic and left nothing for the rule to move.
func TestConsistentTypeImportsFixLeavesTheFileTypeChecking(t *testing.T) {
	t.Parallel()

	originals := map[string]string{}
	repaired := map[string]string{}
	for _, testCase := range consistentTypeImportsUpstreamFixes {
		original, fixed, _ := runConsistentTypeImportsFix(t, testCase.options, testCase.code)
		name := fmt.Sprintf("/repository/source/case%02d.ts", testCase.index)
		originals[name] = original
		repaired[name] = fixed
		if _, _, count := runConsistentTypeImportsFix(t, testCase.options, fixed); count != 0 {
			t.Errorf("case %d: the repaired source still proposes %d fixes:\n%s", testCase.index, count, fixed)
		}
	}
	added := consistentTypeImportsNewTypeErrors(t, originals, repaired)
	for file, messages := range added {
		t.Errorf("%s: the repair added %v\n%s", file, messages, repaired[file])
	}
	// A probe that saw no program would pass for the wrong reason: the originals import modules that
	// do not exist, so the checker must have something to say about them.
	if len(consistentTypeImportsTypeErrors(t, originals)) == 0 {
		t.Fatalf("the checker reported nothing on the originals, so the comparison proved nothing")
	}
}

// TestConsistentTypeImportsFixCases covers what upstream's corpus does not: decorator metadata, the
// attributes refusal, comments, and the two fix styles on one mixed declaration.
func TestConsistentTypeImportsFixCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		options string
		files   map[string]string
		code    string
		want    string
	}{
		{
			name: "separate style splits a mixed declaration",
			code: "import { A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
			want: "import type { A} from 'foo';\nimport { B } from 'foo';\ntype T = A;\nconst b = B;\n",
		},
		{
			name:    "inline style marks the type name in place",
			options: `{"fixStyle": "inline-type-imports"}`,
			code:    "import { A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
			want:    "import { type A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
		},
		{
			name:    "a comment inside the braces survives the inline repair",
			options: `{"fixStyle": "inline-type-imports"}`,
			code:    "import { /* kept */ A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
			want:    "import { /* kept */ type A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
		},
		{
			name:    "no-type-imports removes the declaration keyword",
			options: `{"prefer": "no-type-imports"}`,
			code:    "import type { A } from 'foo';\ntype T = A;\n",
			want:    "import { A } from 'foo';\ntype T = A;\n",
		},
		{
			name:    "no-type-imports removes an inline keyword",
			options: `{"prefer": "no-type-imports"}`,
			code:    "import { type A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
			want:    "import { A, B } from 'foo';\ntype T = A;\nconst b = B;\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			original, fixed, _ := runConsistentTypeImportsFix(t, testCase.options, testCase.code)
			settings := DefaultConsistentTypeImportsOptions()
			if testCase.options != "" {
				decoded, err := DecodeConsistentTypeImportsOptions([]byte(testCase.options))
				if err != nil {
					t.Fatalf("decoding %s: %v", testCase.options, err)
				}
				settings = decoded.(ConsistentTypeImportsOptions)
			}
			result := rule_testing.RunTypedWithOptions(t, ConsistentTypeImports, "consistent_type_imports.ts", testCase.code, settings)
			rule_testing.ExpectFixedSource(t, result, testCase.want)
			name := "/repository/source/fixcase.ts"
			if added := consistentTypeImportsNewTypeErrors(t, map[string]string{name: original}, map[string]string{name: fixed}); len(added) > 0 {
				t.Errorf("the repair added diagnostics %v", added)
			}
		})
	}

	// Import attributes cannot sit on a type-only import, so a mixed declaration carrying them reports
	// and proposes nothing rather than writing `import type { A } from 'foo' with { ... }`.
	_, _, count := runConsistentTypeImportsFix(t, "", "import { A, B } from 'foo' with { type: 'json' };\ntype T = A;\nconst b = B;\n")
	if count != 0 {
		t.Errorf("a declaration with attributes proposed %d fixes, want none", count)
	}
}
