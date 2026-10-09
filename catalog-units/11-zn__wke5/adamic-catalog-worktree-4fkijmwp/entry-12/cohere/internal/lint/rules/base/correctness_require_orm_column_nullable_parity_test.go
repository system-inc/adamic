package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessRequireOrmColumnNullableParityFile = "/repository/source/Entity.ts"

func correctnessRequireOrmColumnNullableParityCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

type correctnessRequireOrmColumnNullableParityFinding struct {
	wantId      string
	wantSpan    string
	wantMessage string
}

// TestOrmColumnNullableParityStaysSilent and its Fires twin are measured, not invented.
//
// A base rule has no upstream corpus, so every verdict below was produced by loading the real
// `CorrectnessRequireOrmColumnNullableParityRule` out of api-phi-health and driving it over the eslint interface
// against the same source text, with a real program so the type questions are answered.
//
// Three clean rows are the ones a reading of the source would get wrong: a property written with a
// question mark is already nullable, `any` and `unknown` count as nullable because the claim was
// never checked, and a non-boolean `nullable` value is a skip rather than a false.
func TestOrmColumnNullableParityStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// The same decorator on a METHOD rather than a property. Closes a mutant that accepted any
		// owner kind, which the rest of the corpus could not see because every other row is
		// already a property. Measured clean against the real rule.
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) m(): string { return \"a\"; } }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x!: string | null; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x?: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: false }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn() x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x!: any; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) @OrmManyToOne(() => B) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x!: string | undefined; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: 'items' }) x!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) [\"computed\"]!: string; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x!: unknown; }",
		"\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @SomethingElse({ nullable: true }) x!: string; }",
	}
	for index, sourceText := range cases {
		t.Run(correctnessRequireOrmColumnNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireOrmColumnNullableParity, correctnessRequireOrmColumnNullableParityFile, sourceText))
		})
	}
}

func TestOrmColumnNullableParityFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []correctnessRequireOrmColumnNullableParityFinding
	}{
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) x!: string; }",
			wantFindings: []correctnessRequireOrmColumnNullableParityFinding{
				{wantId: "columnNullableButTypeNot", wantSpan: "x", wantMessage: "Column declares 'nullable: true' but the type 'string' does not include null/undefined \u2014 a NULL row value will not match this type"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @TimestampColumn({ nullable: true }) x!: Date; }",
			wantFindings: []correctnessRequireOrmColumnNullableParityFinding{
				{wantId: "columnNullableButTypeNot", wantSpan: "x", wantMessage: "Column declares 'nullable: true' but the type 'Date' does not include null/undefined \u2014 a NULL row value will not match this type"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmJoinColumn({ nullable: true }) x!: string; }",
			wantFindings: []correctnessRequireOrmColumnNullableParityFinding{
				{wantId: "columnNullableButTypeNot", wantSpan: "x", wantMessage: "Column declares 'nullable: true' but the type 'string' does not include null/undefined \u2014 a NULL row value will not match this type"},
			},
		},
		{
			sourceText: "\ndeclare function OrmColumn(o?: any): any;\ndeclare function TimestampColumn(o?: any): any;\ndeclare function OrmJoinColumn(o?: any): any;\ndeclare function OrmManyToOne(o?: any): any;\ndeclare function OrmOneToMany(o?: any): any;\ndeclare function SerializableField(o?: any): any;\nclass A { @OrmColumn({ nullable: true }) accessor x: string = \"a\"; }",
			wantFindings: []correctnessRequireOrmColumnNullableParityFinding{
				{wantId: "columnNullableButTypeNot", wantSpan: "x", wantMessage: "Column declares 'nullable: true' but the type 'string' does not include null/undefined \u2014 a NULL row value will not match this type"},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(correctnessRequireOrmColumnNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, CorrectnessRequireOrmColumnNullableParity, correctnessRequireOrmColumnNullableParityFile, testCase.sourceText)

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

// TestOrmColumnNullableParityKeysOnTheDecoratorName pins the name-shaped behaviour.
//
// The rule reads a decorator's callee SPELLING rather than resolving it, which is what keeps the
// decorator half of this rule free of the type checker. Two consequences, both measured against the
// real rule, which agrees on both:
//
//	a LOCAL function named OrmColumn      REPORTS, it is not the real decorator
//	a NAMESPACED @N.OrmColumn call        clean, a qualified name is a different symbol
//
// The first row looks like a defect and is fidelity. The second is the shelf's `CallName` declining
// anything that is not a bare identifier, which is deliberate there and is the behaviour the real
// rule has too.
//
// Pinned because a reader meeting the first row will reasonably want to resolve the binding, and
// doing that would change what the rule is.
func TestOrmColumnNullableParityKeysOnTheDecoratorName(t *testing.T) {
	t.Parallel()

	const preamble = "declare function OrmColumn(o?: any): any;\ndeclare function SerializableField(o?: any): any;\n"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireOrmColumnNullableParity,
		correctnessRequireOrmColumnNullableParityFile,
		preamble+"function OrmColumn(o?: any): any { return null; }\nclass A { @OrmColumn({ nullable: true }) x!: string; }"),
		"columnNullableButTypeNot")

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireOrmColumnNullableParity,
		correctnessRequireOrmColumnNullableParityFile,
		preamble+"namespace N { export function OrmColumn(o?: any): any { return null; } }\nclass A { @N.OrmColumn({ nullable: true }) x!: string; }"))
}
