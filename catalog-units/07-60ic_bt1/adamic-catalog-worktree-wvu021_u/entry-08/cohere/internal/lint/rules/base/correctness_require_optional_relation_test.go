package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessRequireOptionalRelationFile names the fixture file.
const correctnessRequireOptionalRelationFile = "/repository/source/Entity.ts"

// asTheRelationHarnessWroteIt transforms a fixture the way RunTyped transforms its input.
//
// rule_testing writes each typed fixture as strings.TrimSpace(contents)+"\n", so a span sliced
// from the Go literal is offset from the file the rule actually saw.
func asTheRelationHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestRelationMustBeOptional measures this port against the SOURCE rule, shape by shape.
//
// These are our rules, so there is no corpus to import and inventing fixtures would encode the
// same belief as the port. Every row below was measured by driving the real rule from
// api-phi-health over that exact source text, with its @nexus alias supplied by a loader hook.
//
// Three rows taught this port something it would not have written from the source alone: `| null`
// reports where `| null | undefined` does not, `void` reports while `any` and `unknown` do not,
// and a decorator written without parentheses does not match at all.
func TestRelationMustBeOptional(t *testing.T) {
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
			name:         "optional-question",
			why:          "the convention the rule exists to enforce, question-mark optionality",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile?: Profile; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "explicit-undefined",
			why:          "the other accepted spelling, an explicit union with undefined",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: Profile | undefined; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "bare-required",
			why:          "the ordinary violation",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: Profile; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"profile"},
			wantMessages: []string{"Relation property must be optional ('profile?: Profile') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "null-only",
			why:          "`| null` ALONE, which reports. The runtime gives undefined for an unloaded relation and never null, so this type is still lying about its load state; the source library keeps a narrower mask than its nullable sibling for exactly this",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: Profile | null; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"profile"},
			wantMessages: []string{"Relation property must be optional ('profile?: Profile | null') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "null-and-undefined",
			why:          "both, which is accepted because undefined is present",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: Profile | null | undefined; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "any-type",
			why:          "any, silent: it erases the distinction rather than answering it, and the narrow mask counts it",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: any; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "unknown-type",
			why:          "unknown, silent for the same reason",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: unknown; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "one-to-many",
			why:          "the second relation decorator in the set",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmOneToMany() items: Profile[]; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"items"},
			wantMessages: []string{"Relation property must be optional ('items?: Profile[]') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "one-to-one",
			why:          "the third",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmOneToOne() other: Profile; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"other"},
			wantMessages: []string{"Relation property must be optional ('other?: Profile') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "non-relation-decorator",
			why:          "a column decorator rather than a relation one, silent, so the set is a membership test rather than a prefix",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmColumn() name: string; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "no-decorator",
			why:          "no decorator at all, silent",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { profile: Profile; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "bare-decorator-no-parens",
			why:          "a bare `@OrmManyToOne` with no call, which is SILENT. The source helper requires a call expression with an identifier callee, and the sibling rule for declare uses a different helper that does accept the bare form; the two are deliberately not merged",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne profile: Profile; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "computed-key",
			why:          "a computed key, skipped: the source requires an Identifier key, and the message interpolates a name a computed key does not have",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() ['computed']: Profile; }\n",
			wantIds:      []string{},
			wantSpans:    []string{},
			wantMessages: []string{},
		},
		{
			name:         "declare-modifier",
			why:          "a declare modifier, which changes nothing here; this rule is about the type, not the field emit",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() declare profile: Profile; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"profile"},
			wantMessages: []string{"Relation property must be optional ('profile?: Profile') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "void-type",
			why:          "void, which REPORTS. It reads as a surprise and is deliberate: void is absent from the narrow mask even though the nullable sibling includes it",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmManyToOne() profile: void; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"profile"},
			wantMessages: []string{"Relation property must be optional ('profile?: void') because TypeORM relations are undefined when not loaded"},
		},
		{
			name:         "two-relation-decorators",
			why:          "a column decorator alongside a relation one, where the relation still decides",
			sourceText:   "declare const OrmManyToOne: any;\ndeclare const OrmOneToMany: any;\ndeclare const OrmOneToOne: any;\ndeclare const OrmColumn: any;\ndeclare class Profile {}\nclass E { @OrmColumn() @OrmManyToOne() profile: Profile; }\n",
			wantIds:      []string{"mustBeOptional"},
			wantSpans:    []string{"profile"},
			wantMessages: []string{"Relation property must be optional ('profile?: Profile') because TypeORM relations are undefined when not loaded"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, CorrectnessRequireOptionalRelation,
				correctnessRequireOptionalRelationFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheRelationHarnessWroteIt(testCase.sourceText)
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

// TestRelationMustBeOptionalRequiresTheTypedHarness pins the checker guard.
//
// The listener starts with a nil check, and the shim answers nil from a type query on a nil checker
// rather than panicking, so a rule missing that guard goes silent rather than crashing. A vacuous
// green is the more dangerous of the two because nothing announces it.
func TestRelationMustBeOptionalRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "declare const OrmManyToOne: any;\ndeclare class Profile {}\n" +
		"class E { @OrmManyToOne() profile: Profile; }\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, CorrectnessRequireOptionalRelation,
		correctnessRequireOptionalRelationFile, sourceText))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireOptionalRelation,
		correctnessRequireOptionalRelationFile, sourceText), "mustBeOptional")
}
