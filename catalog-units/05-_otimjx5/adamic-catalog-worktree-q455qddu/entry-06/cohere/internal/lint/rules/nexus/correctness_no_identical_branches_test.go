package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoIdenticalBranchesFile = "/repository/source/CorrectnessNoIdenticalBranches.ts"

const correctnessNoIdenticalBranchesJsxFile = "/repository/source/CorrectnessNoIdenticalBranches.tsx"

func correctnessNoIdenticalBranchesSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// correctnessNoIdenticalBranchesKlingApi models `modules/kling/KlingApi.ts` in ahra on 2026-10-02,
// trimmed to the two payload builders that carry the bug, at `:156` and `:187`. The `modelMode`
// argument is the text written at both sites: `'pro'` before the fix, `'std'` on the false side
// after it (Kling's two model modes).
func correctnessNoIdenticalBranchesKlingApi(falseMode string) string {
	return correctnessNoIdenticalBranchesSource(
		"type Argument = { name: string; value: string };",
		"type Payload = { type: string; arguments: Argument[] };",
		"export function buildPayload(",
		"    uploadedUrl: string | undefined,",
		"    modelVersion: string,",
		"    highQuality: boolean,",
		"    hasTailImage: boolean,",
		"    duration: string,",
		"    prompt: string,",
		"    aspectRatio: string,",
		"): Payload {",
		"    let payload: Payload;",
		"    const is3 = modelVersion === '3.0';",
		"    if(uploadedUrl) {",
		"        const modelType = is3 ? 'm2v_aio2video' : highQuality ? 'm2v_img2video_hq' : 'm2v_img2video';",
		"        const effectiveDuration = is3 ? (hasTailImage ? '5' : '7') : duration;",
		"        payload = {",
		"            type: modelType,",
		"            arguments: [",
		"                { name: 'prompt', value: prompt },",
		"                { name: 'duration', value: effectiveDuration },",
		"                { name: 'kling_version', value: modelVersion },",
		"                ...(is3",
		"                    ? [",
		"                          { name: 'imageCount', value: '1' },",
		"                          { name: 'camera_control_enabled', value: 'false' },",
		"                          { name: 'model_mode', value: highQuality ? 'pro' : "+falseMode+" },",
		"                      ]",
		"                    : [{ name: 'tail_image_enabled', value: hasTailImage ? 'true' : 'false' }]),",
		"                { name: 'biz', value: 'klingai' },",
		"            ],",
		"        };",
		"    }",
		"    else {",
		"        const modelType = is3 ? 'm2v_aio2video' : highQuality ? 'm2v_txt2video_hq' : 'm2v_txt2video';",
		"        payload = {",
		"            type: modelType,",
		"            arguments: [",
		"                { name: 'prompt', value: prompt },",
		"                { name: 'duration', value: is3 ? '7' : duration },",
		"                { name: 'aspect_ratio', value: aspectRatio },",
		"                ...(is3",
		"                    ? [",
		"                          { name: 'imageCount', value: '1' },",
		"                          { name: 'camera_control_enabled', value: 'false' },",
		"                          { name: 'model_mode', value: highQuality ? 'pro' : "+falseMode+" },",
		"                      ]",
		"                    : []),",
		"                { name: 'biz', value: 'klingai' },",
		"            ],",
		"        };",
		"    }",
		"    return payload;",
		"}",
	)
}

// correctnessNoIdenticalBranchesReported is the source text each finding points at, in order.
func correctnessNoIdenticalBranchesReported(result rule_testing.Result, sourceText string) []string {
	reported := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		reported = append(reported, sourceText[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	return reported
}

func correctnessNoIdenticalBranchesExpectSpans(t *testing.T, result rule_testing.Result, sourceText string, wantSpans []string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, correctnessNoIdenticalBranchesId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	reported := correctnessNoIdenticalBranchesReported(result, sourceText)
	for index := range wantSpans {
		if reported[index] != wantSpans[index] {
			t.Fatalf("finding %d points at\n%s\nwant\n%s", index, reported[index], wantSpans[index])
		}
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Message.Description != correctnessNoIdenticalBranchesMessage().Description {
			t.Fatalf("message is %q", diagnostic.Message.Description)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// The real site, before the fix: both payload builders, one finding each, on the ternary itself.
// The other ternaries in the file (`is3 ? ... : highQuality ? ... : ...`, `hasTailImage ? '5' : '7'`)
// differ and stay silent beside it.
func TestCorrectnessNoIdenticalBranchesFiresOnKlingApi(t *testing.T) {
	t.Parallel()

	sourceText := correctnessNoIdenticalBranchesKlingApi("'pro'")
	result := rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesFile, sourceText)
	correctnessNoIdenticalBranchesExpectSpans(t, result, sourceText, []string{
		"highQuality ? 'pro' : 'pro'",
		"highQuality ? 'pro' : 'pro'",
	})
}

// The same file after the fix: the false side asks for the standard mode, and nothing fires.
func TestCorrectnessNoIdenticalBranchesStaysSilentOnFixedKlingApi(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesFile, correctnessNoIdenticalBranchesKlingApi("'std'"))
	rule_testing.ExpectClean(t, result)
}

func TestCorrectnessNoIdenticalBranchesFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		lines     []string
		wantSpans []string
	}{
		{"an if and else doing the same thing", []string{
			"declare function start(): void;",
			"export function run(ready: boolean) {",
			"    if(ready) {",
			"        start();",
			"    }",
			"    else {",
			"        start();",
			"    }",
			"}",
		}, []string{"if(ready) {\n        start();\n    }\n    else {\n        start();\n    }"}},
		// Structure, not text: line breaks, spacing and comments inside the branches do not count.
		{"branches differing only in whitespace and comments", []string{
			"declare function save(key: string, value: number): void;",
			"export function run(ready: boolean, count: number) {",
			"    if(ready) {",
			"        // keep it",
			"        save('count', count + 1);",
			"    }",
			"    else {",
			"        save( 'count' ,",
			"            count+1 /* again */ );",
			"    }",
			"}",
		}, []string{"if(ready) {\n        // keep it\n        save('count', count + 1);\n    }\n    else {\n        save( 'count' ,\n            count+1 /* again */ );\n    }"}},
		{"a braced branch against the same statement unbraced", []string{
			"declare function start(): void;",
			"export function run(ready: boolean) {",
			"    if(ready) { start(); } else start();",
			"}",
		}, []string{"if(ready) { start(); } else start();"}},
		{"branches that only return", []string{
			"export function run(ready: boolean) {",
			"    if(ready) return;",
			"    else return;",
			"}",
		}, []string{"if(ready) return;\n    else return;"}},
		{"branches that only continue, inside a loop", []string{
			"export function run(items: number[]) {",
			"    for(const item of items) {",
			"        if(item > 1) { continue; } else { continue; }",
			"    }",
			"}",
		}, []string{"if(item > 1) { continue; } else { continue; }"}},
		// Every branch of the chain is the same, so the finding covers the whole chain once.
		{"an else-if chain whose every branch is the same", []string{
			"declare function start(): void;",
			"export function run(a: boolean, b: boolean) {",
			"    if(a) { start(); }",
			"    else if(b) { start(); }",
			"    else { start(); }",
			"}",
		}, []string{"if(a) { start(); }\n    else if(b) { start(); }\n    else { start(); }"}},
		// Only `b` is ignored, so the finding starts at its `if`.
		{"an else-if chain whose last condition is ignored", []string{
			"declare function start(): void;",
			"declare function stop(): void;",
			"export function run(a: boolean, b: boolean) {",
			"    if(a) { stop(); }",
			"    else if(b) { start(); }",
			"    else { start(); }",
			"}",
		}, []string{"if(b) { start(); }\n    else { start(); }"}},
		{"a ternary chain whose every branch is the same", []string{
			"export function run(a: boolean, b: boolean) {",
			"    return a ? 'pro' : b ? 'pro' : 'pro';",
			"}",
		}, []string{"a ? 'pro' : b ? 'pro' : 'pro'"}},
		{"a ternary chain whose last condition is ignored", []string{
			"export function run(a: boolean, b: boolean) {",
			"    return a ? 'std' : (b ? 'pro' : 'pro');",
			"}",
		}, []string{"b ? 'pro' : 'pro'"}},
		{"parentheses around one ternary branch", []string{
			"export function run(a: boolean, count: number) {",
			"    return a ? (count + 1) : count + 1;",
			"}",
		}, []string{"a ? (count + 1) : count + 1"}},
		// The type control: the branches read identifiers the condition does not narrow, so their
		// types agree and the finding stands.
		{"identifiers the condition does not narrow", []string{
			"declare function format(value: string, width: number): string;",
			"export function run(ready: boolean, label: string, width: number) {",
			"    return ready ? format(label, width) : format(label, width);",
			"}",
		}, []string{"ready ? format(label, width) : format(label, width)"}},
		{"an empty array written two ways", []string{
			"export function run(ready: boolean) {",
			"    return ready ? [] : [ ];",
			"}",
		}, []string{"ready ? [] : [ ]"}},
		{"object literals formatted differently", []string{
			"export function run(ready: boolean, prompt: string) {",
			"    return ready",
			"        ? { name: 'prompt', value: prompt }",
			"        : {",
			"              name: 'prompt',",
			"              value: prompt",
			"          };",
			"}",
		}, []string{"ready\n        ? { name: 'prompt', value: prompt }\n        : {\n              name: 'prompt',\n              value: prompt\n          }"}},
		// ImportRequirePathAliasRule.ts:158 in ahra's nexus. The condition narrows a discriminated
		// union, but the branch only reads a property, and reading a property of the union compiles
		// exactly when reading it of each member does, so the condition really is ignored.
		{"a narrowed union read through a property", []string{
			"type ImportDeclaration = { type: 'ImportDeclaration'; source: { value: string } };",
			"type ImportExpression = { type: 'ImportExpression'; source: { value: unknown } };",
			"export function check(node: ImportDeclaration | ImportExpression) {",
			"    const source = node.type === 'ImportDeclaration' ? node.source : node.source;",
			"    return source;",
			"}",
		}, []string{"node.type === 'ImportDeclaration' ? node.source : node.source"}},
		// LocalStorageService.ts:223 in Structure (phi and connected). The field is a `string`, so
		// the first branch narrows it to `never` and the second already sees the whole type: the
		// merged copy is the second branch, and the number check is dead.
		{"a condition that narrows one branch to never", []string{
			"interface Item { expiresAt?: string }",
			"export function read(parsed: Item) {",
			"    let expiresAt: Date | undefined;",
			"    if(parsed.expiresAt) {",
			"        expiresAt =",
			"            typeof parsed.expiresAt === 'number'",
			"                ? new Date(parsed.expiresAt)",
			"                : new Date(parsed.expiresAt);",
			"    }",
			"    return expiresAt;",
			"}",
		}, []string{"typeof parsed.expiresAt === 'number'\n                ? new Date(parsed.expiresAt)\n                : new Date(parsed.expiresAt)"}},
		// The same `never` shape reaching a compared position, an operand of `*`: the first branch
		// cannot run, the second sees the whole `number`, and the merged copy is the second branch.
		{"an operator operand narrowed to never on one side", []string{
			"export function run(count: number) {",
			"    return typeof count === 'string' ? count * 2 : count * 2;",
			"}",
		}, []string{"typeof count === 'string' ? count * 2 : count * 2"}},
		// The operator guard's control: `===` accepts any union, so a narrowed operand of it does not
		// make the branches different.
		{"a narrowed operand of an operator that accepts unions", []string{
			"export function run(value: string | number) {",
			"    return typeof value === 'string' ? value === 'a' || value === 1 : value === 'a' || value === 1;",
			"}",
		}, []string{"typeof value === 'string' ? value === 'a' || value === 1 : value === 'a' || value === 1"}},
		// The unary operators accept `number | bigint` (checked with tsc), so a narrowed operand of
		// `++` does not make the branches different.
		{"a narrowed operand of a unary increment", []string{
			"export function run(value: number | bigint) {",
			"    if(typeof value === 'bigint') { value++; } else { value++; }",
			"    return value;",
			"}",
		}, []string{"if(typeof value === 'bigint') { value++; } else { value++; }"}},
		// The same resolved signature with a fresh object literal argument on each side: the two
		// literals have two types, and the rule does not compare argument types that cannot matter.
		{"a call with an object literal argument", []string{
			"declare function save(options: { key: string; value: number }): void;",
			"export function run(ready: boolean) {",
			"    if(ready) { save({ key: 'count', value: 1 }); } else { save({ key: 'count', value: 1 }); }",
			"}",
		}, []string{"if(ready) { save({ key: 'count', value: 1 }); } else { save({ key: 'count', value: 1 }); }"}},
		// The rare chain where an `if` holds an `if` and its `else` is the same `if`: `a` is
		// ignored, and so the whole statement is the finding.
		{"an if whose else is the same if as its branch", []string{
			"declare function p(): void;",
			"declare function q(): void;",
			"export function run(a: boolean, c: boolean) {",
			"    if(a) { if(c) p(); else q(); } else if(c) p(); else q();",
			"}",
		}, []string{"if(a) { if(c) p(); else q(); } else if(c) p(); else q();"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := correctnessNoIdenticalBranchesSource(testCase.lines...)
			result := rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesFile, sourceText)
			correctnessNoIdenticalBranchesExpectSpans(t, result, sourceText, testCase.wantSpans)
		})
	}
}

func TestCorrectnessNoIdenticalBranchesStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// Each condition still chooses between `start` and `stop`: this is `if (a || b)` written
		// long, a style question rather than an ignored condition.
		{"two equal branches followed by a different one", []string{
			"declare function start(): void;",
			"declare function stop(): void;",
			"export function run(a: boolean, b: boolean) {",
			"    if(a) { start(); }",
			"    else if(b) { start(); }",
			"    else { stop(); }",
			"}",
		}},
		// With no final `else`, nothing runs when neither holds, so both conditions decide.
		{"an else-if chain with no final else", []string{
			"declare function start(): void;",
			"export function run(a: boolean, b: boolean) {",
			"    if(a) { start(); }",
			"    else if(b) { start(); }",
			"}",
		}},
		{"equal branches that are not adjacent", []string{
			"declare function start(): void;",
			"declare function stop(): void;",
			"export function run(a: boolean, b: boolean, c: boolean) {",
			"    if(a) { start(); }",
			"    else if(b) { stop(); }",
			"    else if(c) { start(); }",
			"}",
		}},
		{"an if with no else", []string{
			"declare function start(): void;",
			"export function run(ready: boolean) {",
			"    if(ready) { start(); }",
			"    start();",
			"}",
		}},
		// Empty branches are `no-empty`'s finding, and commented ones carry the only difference an
		// empty branch can carry.
		{"empty branches", []string{
			"export function run(ready: boolean) {",
			"    if(ready) {} else {}",
			"    if(ready) ; else ;",
			"    if(ready) {",
			"        // handled by the caller",
			"    }",
			"    else {",
			"        // nothing to do",
			"    }",
			"}",
		}},
		// The operator is a field the child walk never visits; the gap scan is what tells these apart.
		{"a different unary operator", []string{
			"export function run(ready: boolean) {",
			"    return ready ? -1 : +1;",
			"}",
		}},
		{"plain and optional member access", []string{
			"export function run(ready: boolean, item: { name: string }) {",
			"    return ready ? item.name : item?.name;",
			"}",
		}},
		{"const against let", []string{
			"declare function use(value: number): void;",
			"export function run(ready: boolean) {",
			"    if(ready) { const value = 1; use(value); } else { let value = 1; use(value); }",
			"}",
		}},
		// Leaves are compared as written: a scanner cannot tell a regular expression from a division,
		// and these strings hold text that looks like a comment or doubled whitespace.
		{"literals that differ only where a rescan would blur them", []string{
			"export function run(ready: boolean) {",
			"    const pattern = ready ? /a b/ : /a  b/;",
			"    const text = ready ? 'a /* b */ c' : 'a  c';",
			"    const template = ready ? `a /* b */ ${pattern}` : `a  ${pattern}`;",
			"    return [text, template];",
			"}",
		}},
		// A decision, not a limit to work around: the tokens differ, and nothing is forgiven past
		// whitespace and comments.
		{"single and double quotes", []string{
			"export function run(ready: boolean) {",
			"    return ready ? 'pro' : \"pro\";",
			"}",
		}},
		{"a trailing comma", []string{
			"export function run(ready: boolean, prompt: string) {",
			"    return ready ? { value: prompt } : { value: prompt, };",
			"}",
		}},
		{"a branch that differs inside a nested call", []string{
			"declare function save(key: string, value: number): void;",
			"export function run(ready: boolean, count: number) {",
			"    if(ready) { save('count', count + 1); } else { save('count', count - 1); }",
			"}",
		}},
		// The branches are the same tokens and different code: `value` is a string in one and a
		// number in the other, so each call resolves a different overload, and `format(value)` on
		// the union would not type-check. The condition is load-bearing.
		{"the condition narrows a type the branches read", []string{
			"declare function format(value: string): string;",
			"declare function format(value: number): string;",
			"export function run(value: string | number) {",
			"    const ternary = typeof value === 'string' ? format(value) : format(value);",
			"    if(typeof value === 'string') { use(format(value)); } else { use(format(value)); }",
			"    return ternary;",
			"}",
			"declare function use(text: string): void;",
		}},
		// A method on a union of receivers: `slice` is a different method on each side. tsc happens
		// to accept `value.slice(0)` on the union, but some union method calls do not resolve, and the
		// rule does not try to tell which, so a different resolved signature is enough to stay silent.
		{"the condition picks the method a call resolves to", []string{
			"export function run(value: string | number[]) {",
			"    return Array.isArray(value) ? value.slice(0) : value.slice(0);",
			"}",
		}},
		// `*` rejects `number | bigint` while accepting each alone (TS2365 under tsc), so the
		// merged copy would not type-check.
		{"the condition narrows the operand of an arithmetic operator", []string{
			"export function run(value: number | bigint) {",
			"    return typeof value === 'bigint' ? value * value : value * value;",
			"}",
		}},
		{"the condition narrows the operand of a relational operator", []string{
			"export function run(value: string | number) {",
			"    return typeof value === 'number' ? value < 3 : value < 3;",
			"}",
		}},
		// A union of tuples cannot be spread into a call (TS2556 under tsc), so the spread's type is
		// compared even when the signature is the same.
		{"the condition narrows a spread argument", []string{
			"declare function record(kind: 's' | 'n', value: string | number): void;",
			"export function run(entry: ['s', string] | ['n', number]) {",
			"    if(entry[0] === 's') { record(...entry); } else { record(...entry); }",
			"}",
		}},
		// A type-level conditional distributes over a union, so identical branches still do work.
		{"a conditional type", []string{
			"export type Boxed<T> = T extends unknown ? { value: T } : { value: T };",
		}},
		// `switch` is declined by decision; see the rule's doc comment.
		{"a switch whose every case is the same", []string{
			"declare function start(): void;",
			"export function run(kind: string) {",
			"    switch(kind) {",
			"        case 'a': start(); break;",
			"        default: start(); break;",
			"    }",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesFile, correctnessNoIdenticalBranchesSource(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// JSX text is not code: doubled spaces render, and `/* */` inside it is text. A scanner run over it
// as code would call both pairs equal.
func TestCorrectnessNoIdenticalBranchesJsx(t *testing.T) {
	t.Parallel()

	silent := correctnessNoIdenticalBranchesSource(
		"export function Label(props: { ready: boolean }) {",
		"    const first = props.ready ? <p>a  b</p> : <p>a b</p>;",
		"    const second = props.ready ? <p>a /* b */ c</p> : <p>a c</p>;",
		"    return <div>{first}{second}</div>;",
		"}",
	)
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesJsxFile, silent))

	firing := correctnessNoIdenticalBranchesSource(
		"export function Label(props: { ready: boolean }) {",
		"    return props.ready ? <p className='label'>Ready</p> : <p className='label' >Ready</p>;",
		"}",
	)
	result := rule_testing.RunTyped(t, CorrectnessNoIdenticalBranches, correctnessNoIdenticalBranchesJsxFile, firing)
	correctnessNoIdenticalBranchesExpectSpans(t, result, firing, []string{
		"props.ready ? <p className='label'>Ready</p> : <p className='label' >Ready</p>",
	})
}

// The rule declares the checker and declines a file without one, rather than reporting without the
// narrowing guard.
func TestCorrectnessNoIdenticalBranchesDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()
	if !CorrectnessNoIdenticalBranches.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker")
	}
	if listeners := CorrectnessNoIdenticalBranches.Run(rule.Context{}, nil); listeners != nil {
		t.Errorf("with no checker the rule must register no listeners, got %d", len(listeners))
	}
}
