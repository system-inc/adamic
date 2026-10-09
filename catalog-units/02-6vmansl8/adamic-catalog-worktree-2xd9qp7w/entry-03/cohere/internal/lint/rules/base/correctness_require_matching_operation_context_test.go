package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const graphQlOperationContextFile = "/repository/source/Resolving.ts"

func graphQlOperationContextCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// graphQlOperationContextPreamble declares the decorators and the context class the rule keys on.
//
// Every fixture needs them, and they have to be real declarations rather than `any`: the rule reads
// the parameter type's SYMBOL NAME, so a context typed through an unresolved import becomes an error
// type whose symbol is not `GraphQlOperationContext`, and every case would report wrongBaseType for
// a reason that has nothing to do with what it was written to test.
//
// That is not hypothetical. The first oracle run against api-phi-health reported wrongBaseType on a
// parameter that plainly was an operation context, because the seeded file imported the type from
// the decorator's module rather than its own. The two live in different files there.
const graphQlOperationContextPreamble = "declare class GraphQlOperationContext<K> { key: K; }\n" +
	"declare function InjectGraphQlOperationContext(): ParameterDecorator;\n" +
	"declare function GraphQlQuery(fn?: unknown): MethodDecorator;\n" +
	"declare function GraphQlMutation(fn?: unknown): MethodDecorator;\n" +
	"declare function GraphQlFieldResolver(fn?: unknown): MethodDecorator;\n" +
	"declare function SomethingElse(fn?: unknown): MethodDecorator;\n" +
	"class ThingOne { one!: string; }\n" +
	"class ThingTwo { two!: string; }\n" +
	"class Derived extends ThingOne { extra!: number; }\n"

// TestGraphQlOperationContextMatchesReturnStaysSilent covers the shapes that must not report.
//
// There is no upstream corpus for this rule: it is ours, and the TypeScript implementation is the
// spec. So these cases were written from that source and then CHECKED against it, by driving the
// real ESLint rule over api-phi-health and over seeded files. Where a case's verdict here disagreed
// with that run, the run won.
func TestGraphQlOperationContextMatchesReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// The matching shape, which is what the one real call site in api-phi-health looks like.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// The three unwrap paths, each of which must leave the comparison equal. A port that
		// compared the wrapped types would report all three.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<ThingOne[]> {\n    void c;\n    return [];\n  }\n}\n",
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<Readonly<ThingOne[]>> {\n    void c;\n    return [];\n  }\n}\n",
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<ThingOne | null> {\n    void c;\n    return null;\n  }\n}\n",

		// The unwrap runs on BOTH sides, so a wrapper on the parameter generic is stripped too.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne[]>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// The scope gate. A method carrying no GraphQL operation decorator is out of this rule's
		// territory however wrong its generic is, which the original states in a comment and which
		// nothing else in the fixture set would catch.
		"class R {\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
		"class R {\n  @SomethingElse(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// A different injection decorator entirely: the rule keys on the name and nothing else.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@SomethingElse() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// A context with no type argument at all. The original returns early rather than reporting,
		// because there is no generic to compare and guessing one would invent a finding.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<never>): Promise<never> {\n    void c;\n    throw new Error('x');\n  }\n}\n",

		// `Readonly<T>` over a NON-array, which is the only shape that keeps the alias symbol and so
		// the only shape that reaches the alias arm of the unwrap. Probed directly: the checker
		// resolves `Readonly<T[]>` to `ReadonlyArray<T>` with symbol `ReadonlyArray` and NO alias at
		// all, so the array arm handles it and the readonly fixture above never exercised the arm it
		// was written for. A mutant deleting that arm survived until these two cases existed.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<Readonly<ThingOne>> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<Readonly<ThingOne>>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// A union of two non-nullish members on BOTH sides. The union arm collapses only when
		// exactly one member survives the nullish filter, so a real two-member union is left intact
		// and compares equal to itself. Its asymmetric sibling reports and is in the fires list; the
		// pair is what pins the `== 1` rather than `>= 1`.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne | ThingTwo>): Promise<ThingOne | ThingTwo> {\n    void c;\n    return new ThingOne();\n  }\n}\n",

		// The decorator on something that is NOT a parameter, inside a local class declared in an
		// operation method. This is the one shape that reaches past the enclosing-method walk with a
		// non-parameter owner, and it is what makes the parameter kind guard load-bearing rather
		// than defensive.
		//
		// The obvious placements do not test it. A decorator on a class, a method, a property or an
		// accessor at the top level all reach this listener with a non-parameter parent, measured,
		// and all four are then declined by the enclosing-method walk because a class member has no
		// enclosing method. Only a local class INSIDE a method finds one, and without the guard the
		// rule reports a bogus mismatch on its property and a bogus wrongBaseType on its method.
		//
		// Silent upstream, measured: the original's resolveDecoratedParameterNode answers null for
		// a non-parameter owner, which is the same job this guard does. Both shapes below.
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(): Promise<ThingOne> {\n    class Inner {\n      @InjectGraphQlOperationContext()\n      p: GraphQlOperationContext<ThingTwo> = null as never;\n    }\n    void Inner;\n    return new ThingOne();\n  }\n}\n",
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(): Promise<ThingOne> {\n    class Inner {\n      @InjectGraphQlOperationContext()\n      m(): GraphQlOperationContext<ThingTwo> {\n        return null as never;\n      }\n    }\n    void Inner;\n    return new ThingOne();\n  }\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(graphQlOperationContextCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, GraphQlOperationContextMatchesReturn,
				graphQlOperationContextFile, graphQlOperationContextPreamble+sourceText))
		})
	}
}

// TestGraphQlOperationContextMatchesReturnFires covers the shapes that must report, with the
// rendered message text asserted by equality.
//
// Both messages interpolate a type name through the checker and the mismatch message interpolates
// two, so a substring predicate could not tell a correct finding from one naming the wrong types.
func TestGraphQlOperationContextMatchesReturnFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantId      string
		wantMessage string
		wantSpan    string
	}{
		// The plain mismatch, measured against the real ESLint rule on a seeded file in
		// api-phi-health: same id, same rendered text, same single finding.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingTwo> but the method returns 'ThingOne'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>",
		},

		// A SUBTYPE, which is why the comparison is bidirectional. `Derived` is assignable to
		// `ThingOne` but not the reverse, so one direction alone would let this through, and a
		// resolver reading a Derived selection set while returning a ThingOne is the defect.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<Derived>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<Derived> but the method returns 'ThingOne'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<Derived>",
		},

		// The other direction of the same pair, so the bidirectionality is pinned from both sides
		// rather than from one.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => Derived)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<Derived> {\n    void c;\n    return new Derived();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingOne> but the method returns 'Derived'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>",
		},

		// The wrong base type, reported under its own id. A sibling rule owns this judgment through
		// the decorator's branded type, and this rule reports it anyway rather than going silent on
		// the shape that most needs saying something.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: ThingTwo): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "wrongBaseType",
			wantMessage: "Parameter decorated with @InjectGraphQlOperationContext must be typed as GraphQlOperationContext<...>, got 'ThingTwo'",
			wantSpan:    "@InjectGraphQlOperationContext() c: ThingTwo",
		},

		// The other two operation decorators, so the scope set is pinned rather than assumed from
		// the one everybody writes.
		{
			sourceText:  "class R {\n  @GraphQlMutation(() => ThingOne)\n  async save(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingTwo> but the method returns 'ThingOne'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>",
		},
		{
			sourceText:  "class R {\n  @GraphQlFieldResolver(() => ThingOne)\n  async resolve(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingTwo> but the method returns 'ThingOne'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>",
		},

		// A non-async operation, so the Promise unwrap is not what makes the rule work.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => ThingOne)\n  find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): ThingOne {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingTwo> but the method returns 'ThingOne'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>",
		},

		// A two-member union against a single type. Taking the FIRST non-nullish member rather than
		// requiring the ONLY one would collapse this union to `ThingOne` and call it a match, which
		// is a real selection-set difference reported as agreement. Measured against the ESLint rule
		// on a seeded file: same finding, same rendered text.
		{
			sourceText:  "class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>): Promise<ThingOne | ThingTwo> {\n    void c;\n    return new ThingOne();\n  }\n}\n",
			wantId:      "mismatch",
			wantMessage: "@InjectGraphQlOperationContext parameter is GraphQlOperationContext<ThingOne> but the method returns 'ThingOne | ThingTwo'",
			wantSpan:    "@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingOne>",
		},
	}
	for index, testCase := range cases {
		t.Run(graphQlOperationContextCaseName(index), func(t *testing.T) {
			t.Parallel()
			sourceText := graphQlOperationContextPreamble + testCase.sourceText
			result := rule_testing.RunTyped(t, GraphQlOperationContextMatchesReturn,
				graphQlOperationContextFile, sourceText)

			rule_testing.ExpectFindings(t, result, testCase.wantId)

			// RunTyped writes the fixture trimmed, so the span slice is against that text.
			onDisk := strings.TrimSpace(sourceText) + "\n"
			diagnostic := result.Diagnostics[0]

			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}
			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
		})
	}
}

// TestGraphQlOperationContextMatchesReturnRequiresTheTypedHarness pins that the rule declines rather
// than crashing without a checker, which no other test here can reach.
func TestGraphQlOperationContextMatchesReturnRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !GraphQlOperationContextMatchesReturn.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: both sides of its comparison are types")
	}

	sourceText := graphQlOperationContextPreamble +
		"class R {\n  @GraphQlQuery(() => ThingOne)\n  async find(@InjectGraphQlOperationContext() c: GraphQlOperationContext<ThingTwo>): Promise<ThingOne> {\n    void c;\n    return new ThingOne();\n  }\n}\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, GraphQlOperationContextMatchesReturn,
		graphQlOperationContextFile, sourceText))

	// The control, so the silence above is the guard declining rather than the rule being blind.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, GraphQlOperationContextMatchesReturn,
		graphQlOperationContextFile, sourceText), "mismatch")
}
