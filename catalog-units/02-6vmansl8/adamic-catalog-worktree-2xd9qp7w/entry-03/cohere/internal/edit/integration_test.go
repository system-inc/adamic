package edit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

// proposeFromRules runs real rules over the current text and collects their fixes.
//
// Every synthetic test in this package hands the engine proposals a test author wrote, which proves
// the engine's arithmetic and proves nothing about whether a real rule's ranges line up with it.
// This closes that: the offsets come from the parser, through a shipped rule, into the engine, and
// the assertion is on the resulting file.
//
// Driving real rules this way is what surfaced the trivia defect these tests now pin. Reasoning
// about ranges would not have found it, because every range involved was individually defensible.
func proposeFromRules(rules ...rule.Rule) Propose {
	return func(fileName string, text string) ([]Proposal, error) {
		rooted := fileName
		if !tspath.IsRootedDiskPath(rooted) {
			rooted = "/" + strings.TrimPrefix(rooted, "/")
		}
		rooted = tspath.NormalizePath(rooted)

		scriptKind := core.ScriptKindTS
		if strings.HasSuffix(rooted, ".tsx") {
			scriptKind = core.ScriptKindTSX
		}

		sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
			FileName: tspath.RootedFilePath(rooted), PathKey: tspath.PathKey(rooted),
		}, text, scriptKind)
		if sourceFile == nil {
			return nil, nil
		}

		var diagnostics []rule.Diagnostic
		for _, subject := range rules {
			currentRule := subject
			context := rule.Context{
				SourceFile: sourceFile,
				Report: func(diagnostic rule.Diagnostic) {
					diagnostic.RuleName = currentRule.Name
					diagnostics = append(diagnostics, diagnostic)
				},
			}
			listeners := currentRule.Run(context, nil)
			walkForRules(sourceFile.AsNode(), listeners)
		}

		return ProposalsFrom(diagnostics), nil
	}
}

func walkForRules(node *ast.Node, listeners rule.Listeners) {
	if node == nil {
		return
	}
	if listener, isListening := listeners[node.Kind]; isListening {
		listener(node)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		walkForRules(child, listeners)
		return false
	})
}

// A real rule's fix must land at the right bytes and produce a file that still parses. This is the
// end-to-end proof: the rule computes ranges against a real parse, and the engine applies them.
func TestARealRuleFixLandsCorrectly(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	fileName := filepath.Join(directory, "imports.ts")
	source := "import * as NodeFileSystem from 'fs';\n" +
		"import * as NodePath from 'path';\n" +
		"import * as NodeOperatingSystem from 'os';\n" +
		"\n" +
		"export const roots = [NodeFileSystem, NodePath, NodeOperatingSystem];\n"

	if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := FixFile(fileName, proposeFromRules(nexus.ImportRequireNodeNamespace), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Changed {
		t.Fatalf("the rule proposes fixes for three bare specifiers and none landed: %+v", result)
	}

	written, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	got := string(written)

	want := "import * as NodeFileSystem from 'node:fs';\n" +
		"import * as NodePath from 'node:path';\n" +
		"import * as NodeOperatingSystem from 'node:os';\n" +
		"\n" +
		"export const roots = [NodeFileSystem, NodePath, NodeOperatingSystem];\n"

	if got != want {
		t.Fatalf("the rewrite is wrong:\n  want %q\n  got  %q", want, got)
	}

	// Three fixes to one file, and every one measured against text the earlier ones moved. This is
	// the offset case in its real form rather than its synthetic one.
	if len(result.Applied) != 3 {
		t.Fatalf("expected 3 fixes applied, got %d", len(result.Applied))
	}

	if parses, reason := Parses(fileName, got); !parses {
		t.Fatalf("the rewritten file does not parse: %s", reason)
	}
}

// The same rule must reach a fixpoint: after its fixes land, re-running proposes nothing.
//
// A rule whose fix does not actually silence its own finding would loop to the pass ceiling, and
// that is worth catching here rather than discovering it on a whole tree.
func TestARealRuleConvergesInTwoPasses(t *testing.T) {
	t.Parallel()
	source := "import * as NodeFileSystem from 'fs';\n"

	result, err := FixText("imports.ts", source, proposeFromRules(nexus.ImportRequireNodeNamespace), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Converged {
		t.Fatalf("a real rule did not converge: %+v", result)
	}
	// One pass to apply, one to confirm nothing is left.
	if result.Passes != 2 {
		t.Fatalf("expected 2 passes, got %d", result.Passes)
	}
	if result.Text != "import * as NodeFileSystem from 'node:fs';\n" {
		t.Fatalf("unexpected result: %q", result.Text)
	}
}

// A rule's fix must replace a node's own text and leave the trivia before it alone.
//
// This fixture found a live defect rather than confirming a known-good behavior, and it is kept
// because the defect is easy to reintroduce. `node.Loc.Pos()` in typescript-go is the position
// before leading trivia, so the range for the specifier in `import * as X from 'fs'` spans
// `" 'fs'"` — the space included. The original `rule.ReplaceNode` used `node.Loc` directly, so a
// rule replacing that node also deleted the whitespace, producing `from'node:fs'`.
//
// The reason it matters more than its cosmetics, and the reason it belongs in this package's tests
// rather than only in the rule package's: the corrupted output still parses. The refusal guard here
// catches a rewrite that breaks syntax, and this one does not break syntax — it silently ate a
// space, a newline, or an entire comment, and every downstream check stayed green. A helper handing
// the engine a range wider than the rule meant is damage no validation downstream can detect. The
// parse guard is necessary and not sufficient, and this fixture is the standing proof of that gap.
//
// Fixed in internal/rule/rule.go, which now trims through `TokenRange`. The three shapes below are
// the ones measured as broken before the fix.
func TestReplaceNodeMustNotEatLeadingTrivia(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "space before the specifier",
			source: "import * as NodeFileSystem from 'fs';\n",
			want:   "import * as NodeFileSystem from 'node:fs';\n",
		},
		{
			name:   "newline before the specifier",
			source: "import * as NodeFileSystem from\n\t'fs';\n",
			want:   "import * as NodeFileSystem from\n\t'node:fs';\n",
		},
		{
			name:   "comment before the specifier",
			source: "import * as NodeFileSystem from /* pinned */ 'fs';\n",
			want:   "import * as NodeFileSystem from /* pinned */ 'node:fs';\n",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result, err := FixText("imports.ts", testCase.source, proposeFromRules(nexus.ImportRequireNodeNamespace), DefaultMaxPasses)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Text != testCase.want {
				t.Fatalf("trivia was consumed by the fix:\n  want %q\n  got  %q", testCase.want, result.Text)
			}
		})
	}
}

// A rule that reports without a fix must leave the file alone. Most rules are in this shape
// deliberately — the namespace case in this very rule declines to fix because rewriting the import
// without re-rooting its references through scope would leave the file uncompilable.
func TestAFindingWithNoFixChangesNothing(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	fileName := filepath.Join(directory, "named.ts")
	// A named import of a builtin: reported, and deliberately not fixed.
	source := "import { readFileSync } from 'node:fs';\n\nexport const read = readFileSync;\n"

	if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := FixFile(fileName, proposeFromRules(nexus.ImportRequireNodeNamespace), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Changed {
		t.Fatalf("a finding with no fix caused a rewrite")
	}

	onDisk, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != source {
		t.Fatalf("the file changed:\n  want %q\n  got  %q", source, onDisk)
	}
}

// A rule whose proposed fix breaks syntax must be refused, with the file left untouched.
//
// The synthetic tests in this package prove the guard works on ranges a test author wrote. This
// proves it fires on a proposal that arrived through the real rule interface, which is the path a
// genuinely wrong rule would take. A rule cannot damage a file it has no way to write to, and this
// is what that sentence actually means in practice.
var breakingRule = rule.Rule{
	Name: "probe-break-syntax",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				body := node.AsFunctionDeclaration().Body
				if body == nil {
					return
				}
				ctx.Report(rule.Diagnostic{
					Range:      node.Loc,
					Message:    rule.Message{Id: "break"},
					SourceFile: ctx.SourceFile,
					Fixes:      []rule.Fix{ctx.ReplaceNode(body, "{ return (1; }")},
				})
			},
		}
	},
}

func TestARuleProposingBrokenSyntaxIsRefused(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	fileName := filepath.Join(directory, "probe.ts")
	source := "export function alpha() { return 1; }\n"

	if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := FixFile(fileName, proposeFromRules(breakingRule), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Changed {
		t.Fatalf("a syntax-breaking fix from a real rule was written")
	}

	onDisk, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != source {
		t.Fatalf("the file was modified:\n  want %q\n  got  %q", source, onDisk)
	}

	refused := false
	for _, rejection := range result.Rejected {
		if strings.HasPrefix(rejection.Reason, ReasonParseFailure) {
			refused = true
		}
	}
	if !refused {
		t.Fatalf("no parse-failure refusal was recorded: %+v", result.Rejected)
	}
}

// nonSilencingRule wraps a case clause's statements in braces and never stops proposing it.
//
// This is the shape a real rule takes when its fix does not silence its own finding: the finding is
// re-reported against the rewritten text every pass, so the fix lands again, and again, until the
// budget runs out. It is written as a probe rather than found in the catalog because no shipped rule
// does this today — every ReportNodeWithFixes site emits one fix that removes its own trigger. The
// guard is worth having precisely because it is unreachable now: it costs nothing today and the
// first rule to get this wrong will be caught by the engine rather than by a developer reading a
// file full of nested braces.
var nonSilencingRule = rule.Rule{
	Name: "probe-never-silences",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCaseBlock: func(node *ast.Node) {
				caseBlock := node.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}
				for _, clauseNode := range caseBlock.Clauses.Nodes {
					clause := clauseNode.AsCaseOrDefaultClause()
					if clause == nil || clause.Statements == nil {
						continue
					}
					statements := clause.Statements.Nodes
					if len(statements) == 0 {
						continue
					}
					ctx.Report(rule.Diagnostic{
						Range:      clauseNode.Loc,
						Message:    rule.Message{Id: "wrap"},
						SourceFile: ctx.SourceFile,
						Fixes: []rule.Fix{
							ctx.InsertBefore(statements[0], "{ "),
							ctx.InsertAfter(statements[len(statements)-1], " }"),
						},
					})
				}
			},
		}
	},
}

// A file that exhausts the pass budget is left exactly as it was found.
//
// The engine used to keep the last pass that parsed, reasoning that it was a valid file. It is: ten
// passes of this rule produce `{ { { ... } } }`, which parses and is nobody's code. Parsing and
// being finished are different properties, and only the first one was ever checked.
func TestAFileThatDoesNotConvergeIsNotWritten(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	fileName := filepath.Join(directory, "switch.ts")
	source := "switch (x) {\n\tcase 1:\n\t\tconst a = 1;\n\t\tbreak;\n}\n"

	if err := os.WriteFile(fileName, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := FixFile(fileName, proposeFromRules(nonSilencingRule), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Converged {
		t.Fatalf("the probe rule cannot converge; the test proves nothing if it did")
	}
	if result.Changed {
		t.Fatalf("a file that never reached a fixpoint was reported as changed")
	}
	if len(result.Applied) != 0 {
		t.Fatalf("fixes from a discarded run were reported as applied: %d", len(result.Applied))
	}

	onDisk, err := os.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if string(onDisk) != source {
		t.Fatalf("the file was rewritten:\n  want %q\n  got  %q", source, string(onDisk))
	}
	if result.Text != source {
		t.Fatalf("the result text is not the original:\n  want %q\n  got  %q", source, result.Text)
	}
}

// The refusal names the rule that would not settle, not only the file it landed on.
//
// A reader handed a filename goes and reads code that is not the problem. The rule is the defect.
func TestNonConvergenceNamesTheRule(t *testing.T) {
	t.Parallel()
	source := "switch (x) {\n\tcase 1:\n\t\tconst a = 1;\n\t\tbreak;\n}\n"

	result, err := FixText("switch.ts", source, proposeFromRules(nonSilencingRule), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.UnconvergedRules) != 1 || result.UnconvergedRules[0] != "probe-never-silences" {
		t.Fatalf("expected the rule named, got %v", result.UnconvergedRules)
	}

	// The rejection survives the refusal rather than being replaced by it.
	found := false
	for _, rejection := range result.Rejected {
		if strings.Contains(rejection.Reason, ReasonPassesReached) {
			found = true
		}
	}
	if !found {
		t.Fatalf("the pass-budget rejection was dropped: %+v", result.Rejected)
	}

	summary := Summarize([]FileResult{result})
	if len(summary.UnconvergedRules) != 1 || summary.UnconvergedRules[0] != "probe-never-silences" {
		t.Fatalf("the summary does not name the rule: %v", summary.UnconvergedRules)
	}
	if len(summary.FilesNotConverged) != 1 {
		t.Fatalf("the summary does not name the file: %v", summary.FilesNotConverged)
	}
}

// siblingOverlapRule reports one diagnostic whose two fixes overlap each other.
//
// This is the constructed case for the last unobserved half of the fix contract. ProposalsFrom
// flattens a diagnostic's fixes into independent proposals, so overlap resolution judges siblings
// the same way it judges rivals. Nothing shipped can reach this: every ReportNodeWithFixes site
// emits exactly one fix. It is built because the alternative is a documented property nobody has
// watched happen, and a property stated by code and never observed is a claim, not a guarantee.
var siblingOverlapRule = rule.Rule{
	Name: "probe-sibling-overlap",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				start, end := node.Loc.Pos(), node.Loc.End()
				ctx.Report(rule.Diagnostic{
					Range:      node.Loc,
					Message:    rule.Message{Id: "sib"},
					SourceFile: ctx.SourceFile,
					Fixes: []rule.Fix{
						rule.ReplaceRange(ctx.SourceFile.Loc.WithPos(start).WithEnd(end), `"A"`),
						rule.ReplaceRange(ctx.SourceFile.Loc.WithPos(start+1).WithEnd(end), `B"`),
					},
				})
			},
		}
	},
}

// Two fixes from one diagnostic are judged separately, and one can land while its sibling is refused.
//
// The engine documents this at ProposalsFrom and resolveOverlaps; this is the observation. The file
// is written with half of an edit that no rule proposed as a whole, and it parses, so the refusal
// guard has no reason to fire. That combination is the reason the multi-fix hazard is worth a
// warning rather than a shrug: the failure is silent by construction.
func TestSiblingFixesFromOneDiagnosticHalfApply(t *testing.T) {
	t.Parallel()
	source := "const a = \"xy\";\n"

	result, err := FixText("p.ts", source, proposeFromRules(siblingOverlapRule), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Applied) != 1 {
		t.Fatalf("expected exactly one sibling to land, got %d", len(result.Applied))
	}

	refusedAsOverlap := 0
	for _, rejection := range result.Rejected {
		if strings.Contains(rejection.Reason, ReasonOverlap) {
			refusedAsOverlap++
		}
	}
	if refusedAsOverlap == 0 {
		t.Fatalf("no sibling was refused as an overlap: %+v", result.Rejected)
	}

	if !strings.Contains(result.Text, `"A"`) {
		t.Fatalf("the surviving half did not land: %q", result.Text)
	}
	if strings.Contains(result.Text, `B"`) {
		t.Fatalf("both siblings landed, so this no longer tests a half-application: %q", result.Text)
	}

	// The half that survived leaves valid syntax, which is exactly why the parse guard cannot help.
	if parses, reason := Parses("p.ts", result.Text); !parses {
		t.Fatalf("expected the half-applied result to parse, got %s", reason)
	}
}
