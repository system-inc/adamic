package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const noDupeClassMembersFile = "/repository/source/Members.ts"

func noDupeClassMembersCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoDupeClassMembersStaysSilent is upstream's eleven passing cases verbatim, plus the shapes
// that separate this wrapper from the core rule it extends.
//
// The imported half of this list cannot see the wrapper at all. Driven through the installed 8.67.0
// build, upstream's own corpus for this rule produces BYTE-IDENTICAL output from the wrapper and
// from the bare core on all twenty-one cases, because it writes no literal computed key anywhere.
// Everything below the imported cases is measured for that reason.
func TestNoDupeClassMembersStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class A {\n  foo() {}\n  bar() {}\n}\n",
		"class A {\n  static foo() {}\n  foo() {}\n}\n",
		"class A {\n  get foo() {}\n  set foo(value) {}\n}\n",
		"class A {\n  static foo() {}\n  get foo() {}\n  set foo(value) {}\n}\n",
		"class A {\n  foo() {}\n}\nclass B {\n  foo() {}\n}\n",
		"class A {\n  [foo]() {}\n  foo() {}\n}\n",
		"class A {\n  foo() {}\n  bar() {}\n  baz() {}\n}\n",
		"class A {\n  *foo() {}\n  *bar() {}\n  *baz() {}\n}\n",
		"class A {\n  get foo() {}\n  get bar() {}\n  get baz() {}\n}\n",
		"class A {\n  1() {}\n  2() {}\n}\n",
		"class Foo {\n  foo(a: string): string;\n  foo(a: number): number;\n  foo(a: any): any {}\n}\n",

		// Measured, not imported. A computed key is never a duplicate here, even when it resolves
		// to a literal the core rule can and does compare. All three report under the CORE rule and
		// are silent under the wrapper, measured on the installed build by driving both over the
		// same inputs, and silent here.
		"class A {\n  ['foo']() {}\n  ['foo']() {}\n}\n",
		"class A {\n  [1]() {}\n  [1]() {}\n}\n",

		// The case that decides HOW the filter is expressed rather than merely whether it exists.
		// The first member is computed and the second is not, so an implementation that dropped
		// findings after the core produced them would report on `foo`, whose own key is plain.
		// Upstream skips the computed member before it is keyed, so it never seeds a collision.
		// This row is the one that caught the output-filtering version of this rule.
		"class A {\n  ['foo']() {}\n  foo() {}\n}\n",

		// An unresolvable computed key, silent in the core too, so this row pins the filter is not
		// doing the work here and the core's own decline is.
		"class A {\n  [x]() {}\n  [x]() {}\n}\n",

		// A TypeScript overload set. Upstream's other filter exists for this, and our core already
		// declines it independently: a previous porter added that after measuring four false
		// positives on real source. Measured again across an overload pair, a `declare class` and
		// an `abstract class`; the wrapper and the bare core agree on all of them.
		"class A {\n  foo(a: string): void;\n  foo(a: number): void;\n  foo(a: unknown): void {}\n}\n",
		"declare class A {\n  foo(): void;\n  foo(): void;\n}\n",
		"abstract class A {\n  abstract foo(): void;\n  abstract foo(): void;\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(noDupeClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoDupeClassMembers,
				noDupeClassMembersFile, sourceText))
		})
	}
}

// noDupeClassMembersFinding is one expected finding with the layers it can be wrong at.
//
// The core rule owns the message text, so this asserts the id and the SPAN. The span is what
// separates this wrapper from a broken filter: with the computed members removed on the input side
// a finding lands on the surviving duplicate, and a port filtering the output instead lands it on a
// member upstream never reports.
type noDupeClassMembersFinding struct {
	wantSpan string
}

// TestNoDupeClassMembersFires is upstream's ten reporting cases verbatim, plus measured rows where
// a computed member sits beside a real duplicate.
func TestNoDupeClassMembersFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noDupeClassMembersFinding
	}{
		{
			sourceText: "class A {\n  foo() {}\n  foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "!class A {\n  foo() {}\n  foo() {}\n};\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  'foo'() {}\n  'foo'() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "'foo'"},
			},
		},
		{
			sourceText: "class A {\n  10() {}\n  1e1() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "1e1"},
			},
		},
		{
			sourceText: "class A {\n  foo() {}\n  foo() {}\n  foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  static foo() {}\n  static foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  foo() {}\n  get foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  set foo(value) {}\n  foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  foo;\n  foo = 42;\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			sourceText: "class A {\n  foo;\n  foo() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
		{
			// Measured: a computed duplicate and a plain one in one class. The computed pair is
			// silent and the plain pair still reports, which is what proves the removal is scoped
			// to the computed members rather than disabling the class.
			sourceText: "class A {\n  ['foo']() {}\n  ['foo']() {}\n  bar() {}\n  bar() {}\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "bar"},
			},
		},
		{
			// A duplicated property rather than a method, which upstream's wrapper also covers
			// since its listener names both member kinds.
			sourceText: "class A {\n  foo: string;\n  foo: string;\n}\n",
			wantFindings: []noDupeClassMembersFinding{
				{wantSpan: "foo"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noDupeClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoDupeClassMembers, noDupeClassMembersFile,
				testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range testCase.wantFindings {
				wantIds[position] = "noDupeClassMembers"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
			}
		})
	}
}

// TestNoDupeClassMembersRestoresTheMemberList pins the restore half of the member-list swap.
//
// It exists because a mutant deleting the restore survives every other fixture here, and the reason
// is structural rather than a coverage gap: within one `rule_testing.Run` the tree is reparsed, so
// no fixture calling the rule normally can observe a leak. What the restore protects is the case
// this package's fixtures never build, one tree shared by many rules, where a class left with its
// computed members deleted would be handed in that state to every rule visiting it afterwards.
//
// So this drives the rule's own listener and then reads the member list back off the same node,
// which is the only shape that can tell a scoped swap from a permanent one. The second row is the
// control: a class with nothing computed takes the no-swap path, and without it this test would
// pass against a rule that never swapped at all.
func TestNoDupeClassMembersRestoresTheMemberList(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		sourceText  string
		wantMembers int
	}{
		{
			name:        "a class whose members were swapped out comes back whole",
			sourceText:  "class A {\n  ['foo']() {}\n  ['foo']() {}\n  bar() {}\n  bar() {}\n}\n",
			wantMembers: 4,
		},
		{
			name:        "a class with nothing computed is never swapped in the first place",
			sourceText:  "class A {\n  foo() {}\n  foo() {}\n}\n",
			wantMembers: 2,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			observed := 0
			observer := rule.Rule{
				Name: "test/no-dupe-class-members-restore",
				Run: func(ctx rule.Context, options any) rule.Listeners {
					wrapped := NoDupeClassMembers.Run(ctx, options)
					return rule.Listeners{
						ast.KindClassDeclaration: func(node *ast.Node) {
							wrapped[ast.KindClassDeclaration](node)
							// Read AFTER the rule has run, on the same node it just handled.
							if members := node.AsClassDeclaration().Members; members != nil {
								observed = len(members.Nodes)
							}
						},
					}
				},
			}
			rule_testing.Run(t, observer, noDupeClassMembersFile, testCase.sourceText)
			if observed != testCase.wantMembers {
				t.Fatalf("members after the rule ran: expected %d, got %d", testCase.wantMembers,
					observed)
			}
		})
	}
}
