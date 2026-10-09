package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessRequireOrmColumnDeclareFile names the fixture file.
const correctnessRequireOrmColumnDeclareFile = "/repository/source/Column.ts"

// asTheOrmColumnHarnessWroteIt transforms a fixture the way the harness transforms its input.
func asTheOrmColumnHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestOrmColumnRequiresDeclare measures this port against the SOURCE rule, shape by shape.
//
// No corpus exists to import, so every row was measured by driving the real rule from
// api-phi-health over that exact source text. All eight decorators in the set are exercised
// individually rather than one standing for the rest, because the set is rule-local by design and
// a missing entry is exactly the kind of thing one representative case would hide.
//
// Three rows are the reason this rule does not share a decorator helper with its sibling: a bare
// decorator reports here and is silent there, a namespaced one is silent in both, and a private
// name reports with its name rendered as a placeholder.
func TestOrmColumnRequiresDeclare(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		why          string
		sourceText   string
		wantIds      []string
		wantSpans    []string
		wantMessages []string
	}{
		{
			name:         "missing-declare",
			why:          "the ordinary violation the rule exists for",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "has-declare",
			why:          "the fix already applied, silent",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() declare name: string; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "primary-key",
			why:          "the second of the eight decorators in the set",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmPrimaryKey() id: number; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"id"},
			wantMessages: []string{"Property 'id' decorated with @OrmPrimaryKey() must use 'declare' (e.g. `declare id: ...`)"},
		},
		{
			name:         "primary-auto",
			why:          "the third",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmPrimaryAutoColumn() id: number; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"id"},
			wantMessages: []string{"Property 'id' decorated with @OrmPrimaryAutoColumn() must use 'declare' (e.g. `declare id: ...`)"},
		},
		{
			name:         "create-date",
			why:          "the fourth",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmCreateDateColumn() createdAt: Date; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"createdAt"},
			wantMessages: []string{"Property 'createdAt' decorated with @OrmCreateDateColumn() must use 'declare' (e.g. `declare createdAt: ...`)"},
		},
		{
			name:         "update-date",
			why:          "the fifth",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmUpdateDateColumn() updatedAt: Date; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"updatedAt"},
			wantMessages: []string{"Property 'updatedAt' decorated with @OrmUpdateDateColumn() must use 'declare' (e.g. `declare updatedAt: ...`)"},
		},
		{
			name:         "column-index",
			why:          "the sixth",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumnIndex() a: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"a"},
			wantMessages: []string{"Property 'a' decorated with @OrmColumnIndex() must use 'declare' (e.g. `declare a: ...`)"},
		},
		{
			name:         "column-unique",
			why:          "the seventh",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumnUnique() b: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"b"},
			wantMessages: []string{"Property 'b' decorated with @OrmColumnUnique() must use 'declare' (e.g. `declare b: ...`)"},
		},
		{
			name:         "column-unique-index",
			why:          "the eighth, so all eight are exercised rather than one standing for the rest",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumnUniqueIndex() c: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"c"},
			wantMessages: []string{"Property 'c' decorated with @OrmColumnUniqueIndex() must use 'declare' (e.g. `declare c: ...`)"},
		},
		{
			name:         "relation-decorator-not-in-set",
			why:          "a RELATION decorator, silent: that is the sibling rule's business and the two sets are kept apart on purpose",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmManyToOne() profile: string; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "no-decorator",
			why:          "no decorator at all, silent",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { name: string; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "bare-decorator-no-parens",
			why:          "a bare `@OrmColumn` with no call, which REPORTS. This is the opposite of the sibling relation rule and the reason this port cannot use the shelf's CallName, which answers empty for this shape",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "namespaced-decorator",
			why:          "a namespaced `@Orm.Column()`, silent: the source calls its helper with member-expression resolution left at the default of false, and two other rules in that library pass true while this one does not",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @Orm.Column() name: string; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "computed-key",
			why:          "a computed key, which reports with the name rendered as the literal `<computed>` while the finding still points at the key",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() ['computed']: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"'computed'"},
			wantMessages: []string{"Property '<computed>' decorated with @OrmColumn() must use 'declare' (e.g. `declare <computed>: ...`)"},
		},
		{
			name:         "two-decorators-first-matches",
			why:          "a matching decorator first, so the scan stops there",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() @OrmManyToOne() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "two-decorators-second-matches",
			why:          "a matching decorator second, which still reports and names the matching one rather than the first",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmManyToOne() @OrmColumn() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "with-initializer",
			why:          "an initializer present, which changes nothing: the initializer is precisely what overwrites the decorator's metadata, so this is the shape the rule most wants to catch",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() name: string = 'x'; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "static-property",
			why:          "a static property, which reports the same as an instance one",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() static name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "private-name",
			why:          "a PRIVATE name, which reports and renders as `<computed>` too, because a private name is its own node type rather than an identifier",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmPrimaryKey: any;\ndeclare const OrmPrimaryAutoColumn: any;\ndeclare const OrmCreateDateColumn: any;\ndeclare const OrmUpdateDateColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmColumnUnique: any;\ndeclare const OrmColumnUniqueIndex: any;\ndeclare const OrmManyToOne: any;\ndeclare const Orm: any;\nclass E { @OrmColumn() #secret: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"#secret"},
			wantMessages: []string{"Property '<computed>' decorated with @OrmColumn() must use 'declare' (e.g. `declare <computed>: ...`)"},
		},
		{
			name:         "two-matching-decorators",
			why:          "TWO decorators both in the set, where the message names the FIRST in source order. Every other row has at most one match, so the scan's early stop decides nothing there and a version taking the last match passes all nineteen",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmPrimaryKey: any;\nclass E { @OrmColumn() @OrmColumnIndex() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumn() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "two-matching-reversed",
			why:          "the same pair written the other way round, which names the other decorator; the pair together is what pins source order rather than set order or alphabetical order",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmPrimaryKey: any;\nclass E { @OrmColumnIndex() @OrmColumn() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmColumnIndex() must use 'declare' (e.g. `declare name: ...`)"},
		},
		{
			name:         "three-matching",
			why:          "three matches, so the rule is stopping at the first rather than at the second",
			sourceText:   "declare const OrmColumn: any;\ndeclare const OrmColumnIndex: any;\ndeclare const OrmPrimaryKey: any;\nclass E { @OrmPrimaryKey() @OrmColumn() @OrmColumnIndex() name: string; }\n",
			wantIds:      []string{"missingDeclare"},
			wantSpans:    []string{"name"},
			wantMessages: []string{"Property 'name' decorated with @OrmPrimaryKey() must use 'declare' (e.g. `declare name: ...`)"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, CorrectnessRequireOrmColumnDeclare,
				correctnessRequireOrmColumnDeclareFile, testCase.sourceText)
			// A row expecting nothing goes through ExpectClean rather than through ExpectFindings
			// with an empty list. The two are the same assertion to a reader and not to the
			// fixture-pair guard, which reads the call by name: a suite that only ever calls
			// ExpectFindings proves the rule can detect and never that it can stay quiet, and a
			// violation-only corpus is exactly what that guard exists to catch.
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheOrmColumnHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
				if reported.Message.Description != testCase.wantMessages[findingIndex] {
					t.Errorf("finding %d reads %q, wanted %q", findingIndex,
						reported.Message.Description, testCase.wantMessages[findingIndex])
				}
			}
		})
	}
}
