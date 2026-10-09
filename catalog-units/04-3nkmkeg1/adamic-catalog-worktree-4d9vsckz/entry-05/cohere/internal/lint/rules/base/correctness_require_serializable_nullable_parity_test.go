package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessRequireSerializableNullableParityFile = "/repository/source/Entity.ts"

func correctnessRequireSerializableNullableParityCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

type correctnessRequireSerializableNullableParityFinding struct {
	wantId      string
	wantSpan    string
	wantMessage string
}

// TestSerializableNullableParityStaysSilent and its Fires twin are measured, not invented.
//
// Same method as the ORM sibling: the real rule was loaded out of api-phi-health and driven over
// the eslint interface with a real program.
//
// The row worth reading twice is a bare `@SerializableField()` on a nullable property, which
// REPORTS. An absent `optional` is read as false rather than skipped, so absent and explicitly-false
// are the same claim. Reading the source suggests otherwise, because the guard it writes tests for
// a null that absent never produces.
func TestSerializableNullableParityStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// A DIFFERENT decorator carrying the same option key, and the same decorator on a METHOD
		// rather than a property. Each closes a mutant the rest of the corpus could not see: one
		// that accepts any decorator name, and one that accepts any owner kind. Both measured
		// clean against the real rule.
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ optional: true }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true }) m(): string { return \"a\"; } }",
		// A non-boolean `optional` on a NULLABLE property. These two are what separate the skip
		// branch from the rest of the rule: with skip ignored, an absent-or-false optional beside a
		// nullable type reports, so only a nullable property can see the difference. The corpus
		// first written here had `optional: 'items'` on a non-nullable property only, which both the
		// rule and a skip-ignoring mutant decline, and the mutant survived. Measured clean.
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: 'items' }) x?: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: 'items' }) x!: string | null; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true }) x?: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: false }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true, defaultValue: \"a\" }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: false, defaultValue: \"a\" }) x?: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField() x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true }) x!: string | null; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: 'items' }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true }) [\"c\"]!: string; }",
	}
	for index, sourceText := range cases {
		t.Run(correctnessRequireSerializableNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireSerializableNullableParity, correctnessRequireSerializableNullableParityFile, sourceText))
		})
	}
}

func TestSerializableNullableParityFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []correctnessRequireSerializableNullableParityFinding
	}{
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: true }) x!: string; }",
			wantFindings: []correctnessRequireSerializableNullableParityFinding{
				{wantId: "decoratorOptionalButTypeNot", wantSpan: "x", wantMessage: "@SerializableField declares 'optional: true' but type 'string' does not include null/undefined"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: false }) x?: string; }",
			wantFindings: []correctnessRequireSerializableNullableParityFinding{
				{wantId: "typeOptionalButDecoratorNot", wantSpan: "x", wantMessage: "Type 'string | undefined' is nullable but @SerializableField does not declare 'optional: true'"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField() x?: string; }",
			wantFindings: []correctnessRequireSerializableNullableParityFinding{
				{wantId: "typeOptionalButDecoratorNot", wantSpan: "x", wantMessage: "Type 'string | undefined' is nullable but @SerializableField does not declare 'optional: true'"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SerializableField({ optional: false }) x!: any; }",
			wantFindings: []correctnessRequireSerializableNullableParityFinding{
				{wantId: "typeOptionalButDecoratorNot", wantSpan: "x", wantMessage: "Type 'any' is nullable but @SerializableField does not declare 'optional: true'"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(correctnessRequireSerializableNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, CorrectnessRequireSerializableNullableParity, correctnessRequireSerializableNullableParityFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position, want := range testCase.wantFindings {
				wantIds[position] = want.wantId
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage, diagnostic.Message.Description)
				}
				if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
					t.Fatalf("finding %d: expected no repair, got %d fixes and %d suggestions",
						position, len(diagnostic.Fixes), len(diagnostic.Suggestions))
				}
			}
		})
	}
}

// TestSerializableNullableParityKeysOnTheDecoratorName pins the name-shaped behaviour.
//
// The same property as the column sibling, measured the same way and agreeing with the real rule on
// both rows:
//
//	a LOCAL function named SerializableField   REPORTS
//	a NAMESPACED @N.SerializableField call     clean
//
// It is a property of every base rule that keys on an identifier's spelling rather than resolving
// it, and it is written down in both files because each reads as a defect on its own.
func TestSerializableNullableParityKeysOnTheDecoratorName(t *testing.T) {
	t.Parallel()

	const preamble = "declare function OrmColumn(o?: any): any;\ndeclare function SerializableField(o?: any): any;\n"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireSerializableNullableParity,
		correctnessRequireSerializableNullableParityFile,
		preamble+"function SerializableField(o?: any): any { return null; }\nclass A { @SerializableField({ optional: true }) x!: string; }"),
		"decoratorOptionalButTypeNot")

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireSerializableNullableParity,
		correctnessRequireSerializableNullableParityFile,
		preamble+"namespace N { export function SerializableField(o?: any): any { return null; } }\nclass A { @N.SerializableField({ optional: true }) x!: string; }"))
}
