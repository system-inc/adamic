package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const graphQlNullableParityFile = "/repository/source/Schema.ts"

func graphQlNullableParityCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// graphQlNullableParityPreamble declares the decorators the rule keys on.
//
// The field decorator is declared as both a property AND a method decorator, which is not padding:
// the underlying decorator really is valid on both, and the rule's method arm exists for that.
const graphQlNullableParityPreamble = "declare function GraphQlArgument(n?: unknown, f?: unknown, o?: unknown): ParameterDecorator;\n" +
	"declare function GraphQlField(f?: unknown, o?: unknown): PropertyDecorator & MethodDecorator;\n" +
	"declare function GraphQlQuery(f?: unknown, o?: unknown): MethodDecorator;\n" +
	"declare function GraphQlMutation(f?: unknown, o?: unknown): MethodDecorator;\n" +
	"declare function GraphQlFieldResolver(f?: unknown, o?: unknown): MethodDecorator;\n" +
	"declare function OrmManyToOne(f?: unknown, o?: unknown): PropertyDecorator;\n" +
	"declare function SomethingElse(f?: unknown, o?: unknown): PropertyDecorator;\n" +
	"declare function GraphQlInputType(n?: unknown, o?: unknown): ClassDecorator;\n" +
	"declare function GraphQlObjectType(n?: unknown, o?: unknown): ClassDecorator;\n" +
	"declare const String: unknown;\n" +
	"class Thing { value!: string; }\n"

// TestGraphQlNullableParityStaysSilent covers the shapes that must not report.
//
// No upstream corpus exists for this rule, so these were written from the TypeScript source and then
// checked against it: the real ESLint rule was driven over api-phi-health and over a seeded file
// carrying every shape below and in the fires list. Where a verdict here disagreed with that run,
// the run won.
func TestGraphQlNullableParityStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Both directions in agreement, which is the whole point of the rule.
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  a!: string | null;\n}\n",
		"class C {\n  @GraphQlField(() => String, { nullable: false })\n  a!: string;\n}\n",
		"class C {\n  @GraphQlField(() => String)\n  a!: string;\n}\n",

		// A non-boolean `nullable` is skipped rather than read as false. These declare list-element
		// nullability, a different question this rule cannot answer from the element type, and both
		// spellings appear in the original's header comment.
		"class C {\n  @GraphQlField(() => [String], { nullable: 'items' })\n  a!: string[];\n}\n",
		"class C {\n  @GraphQlField(() => [String], { nullable: 'itemsAndList' })\n  a!: string[];\n}\n",

		// The same two spellings beside a NULLABLE type, which is what actually pins the skip. The
		// pair above cannot: a skipped decorator and one read as `nullable: false` both agree with
		// a non-nullable type, so a mutant disabling the skip survives them. Here the two paths
		// diverge, because reading the flag as false would report typeNullableButDecoratorNot.
		// Measured silent against the real ESLint rule on a seeded file.
		"class C {\n  @GraphQlField(() => [String], { nullable: 'items' })\n  a!: string[] | null;\n}\n",
		"class C {\n  @GraphQlField(() => [String], { nullable: 'itemsAndList' })\n  a?: string[];\n}\n",

		// A relation is out of scope. Its schema declares the business contract while the type
		// reflects the load state, so the two legitimately disagree and a sibling rule owns the
		// pair. Without this gate the case reports, which is the whole reason the gate exists.
		"class C {\n  @OrmManyToOne(() => Thing)\n  @GraphQlField(() => Thing, { nullable: false })\n  a?: Thing;\n}\n",

		// A decorator this rule does not target, carrying a flag that would report if it did. The
		// METHOD form is the one that pins the target-set gate: the property form is declined by
		// the method gate further down anyway, so a mutant removing the target set survives it. A
		// method with a non-target decorator reaches the return-type read and reports without the
		// gate. Found by probing, since the corpus is mine and the obvious fixture was the weak one.
		"class C {\n  @SomethingElse(() => String, { nullable: true })\n  a!: string;\n}\n",
		"class C {\n  @SomethingElse(() => String, { nullable: true })\n  m(): string {\n    return '';\n  }\n}\n",

		// An OPERATION decorator on something that is not a method. The original requires a method
		// definition before reading a return type, so these are silent however wrong the flag is,
		// and a port dropping that gate reports both. The shapes parse, so the gate is a real
		// filter rather than a guard against something impossible. Measured silent against the real
		// ESLint rule on a seeded file carrying the property form.
		"class C {\n  @GraphQlQuery(() => String, { nullable: true })\n  prop!: string;\n}\n",
		"class C {\n  @GraphQlMutation(() => String, { nullable: true })\n  get g(): string {\n    return '';\n  }\n}\n",

		// A FUNCTION-VALUED property, which is the shape that actually pins the method gate. The two
		// above cannot: a plain property and a getter have no call signature, so the nil check below
		// the gate declines them anyway and a mutant removing the gate survives. This one's type
		// DOES have a call signature, so without the gate the rule reads the function's return type
		// and reports. Probed to find it, then measured silent against the real ESLint rule.
		"class C {\n  @GraphQlQuery(() => String, { nullable: true })\n  fnProp!: () => string;\n}\n",

		// The operation decorators in agreement, including the Promise unwrap. A port comparing the
		// wrapped type would report the async one, because a Promise is never nullable.
		"class C {\n  @GraphQlQuery(() => Thing, { nullable: true })\n  async find(): Promise<Thing | null> {\n    return null;\n  }\n}\n",
		"class C {\n  @GraphQlMutation(() => Thing, { nullable: false })\n  async save(): Promise<Thing> {\n    return new Thing();\n  }\n}\n",
		"class C {\n  @GraphQlFieldResolver(() => Thing, { nullable: true })\n  resolve(): Thing | undefined {\n    return undefined;\n  }\n}\n",

		// An argument in agreement, which on an input means admitting null itself: a client can send
		// one. `id?: string` was silent here before #twm9k22 and now fires, so it moved to the fires
		// list. Both the optional and the required spelling admit null, so neither reports.
		"class C {\n  @GraphQlQuery(() => Thing)\n  find(@GraphQlArgument('id', () => String, { nullable: true }) id?: string | null): Thing {\n    void id;\n    return new Thing();\n  }\n}\n",
		"class C {\n  @GraphQlQuery(() => Thing)\n  find(@GraphQlArgument('id', () => String, { nullable: true }) id: string | null): Thing {\n    void id;\n    return new Thing();\n  }\n}\n",

		// An input class's nullable field admitting null, optional or not.
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a?: string | null;\n}\n",
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a!: string | null;\n}\n",

		// A NON-nullable input, the common shape, is not asked for null: only `nullable: true`
		// promises a client may send one. Without this pair, the input check dropping its
		// `declaredNullable` condition survives, since every other input fixture is nullable.
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String)\n  a!: string;\n}\n",
		"class C {\n  @GraphQlQuery(() => Thing)\n  find(@GraphQlArgument('id', () => String) id: string): Thing {\n    void id;\n    return new Thing();\n  }\n}\n",

		// `any` and `unknown` admit null, so an input typed with either is not asked for `| null`.
		// The same width as the parity mask, for the same reason: neither type claims anything to
		// contradict. Each pins its own flag in typeIncludesNull's mask.
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a?: any;\n}\n",
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a?: unknown;\n}\n",

		// OUTPUTS keep `?: T`, ruled for #twm9k22: a resolver's undefined serializes as null, so the
		// type says nothing false about the wire. A field outside any input class, a field of an
		// object type (which pins the input set as a set rather than as "any class decorator"),
		// and a field resolver's return (above) all stay silent.
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  a?: string;\n}\n",
		"@GraphQlObjectType()\nclass O {\n  @GraphQlField(() => String, { nullable: true })\n  a?: string;\n}\n",

		// A field on a METHOD inside an input class is not asked: a client sends properties, and a
		// computed field is output whatever class holds it.
		"@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  computed(): string | undefined {\n    return undefined;\n  }\n}\n",

		// The field decorator on a METHOD and on a GETTER, in agreement. The property arm alone
		// would go silent on both rather than report them, so these pin that the arm exists at all.
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  computed(): string | null {\n    return null;\n  }\n}\n",
		// A GETTER is silent even when its decorator and its type DISAGREE, which is a real
		// divergence from what the rule looks like it should do and was measured rather than
		// assumed. estree calls a getter a method definition, so the original reaches it, and then
		// declines: `getCallSignatures()` on a getter returns nothing, so its `if(!signature)
		// return` fires. Our parser reaches the same verdict by the same mechanism, because the
		// type at a get accessor is already the property type and carries no call signature.
		//
		// Driven against the real ESLint rule on a seeded file holding this getter beside an
		// equivalent method: the method reports and the getter does not. The method is in the fires
		// list. If somebody later decides a getter should be checked, that is a rule change in both
		// implementations rather than a bug in this one.
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  get computed(): string {\n    return '';\n  }\n}\n",
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  get computed(): string | null {\n    return null;\n  }\n}\n",

		// An argument decorator on something that is NOT a parameter. Unlike the operation
		// decorators, the argument arm has nothing below it to decline a wrong owner, so the
		// parameter kind guard is directly load-bearing: measured, ALL FOUR of these report a bogus
		// decoratorNullableButTypeNot without it, reading the property's or the class's own type as
		// if it were a parameter's.
		//
		// Silent upstream, measured on a seeded file: the original's resolveDecoratedParameterNode
		// answers null for a non-parameter owner and the rule returns.
		"class C {\n  @GraphQlArgument('a', () => String, { nullable: true })\n  a!: string;\n}\n",
		"class C {\n  @GraphQlArgument('a', () => String, { nullable: true })\n  m(): string {\n    return '';\n  }\n}\n",
		"@GraphQlArgument('a', () => String, { nullable: true })\nclass C {\n  a!: string;\n}\n",
		"class C {\n  @GraphQlArgument('a', () => String, { nullable: true })\n  get g(): string {\n    return '';\n  }\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(graphQlNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, GraphQlNullableParity,
				graphQlNullableParityFile, graphQlNullableParityPreamble+sourceText))
		})
	}
}

// TestGraphQlNullableParityFires covers the shapes that must report.
//
// Both messages interpolate the rendered type name, which is the part a reader uses to see which
// side is wrong, so the text is asserted by equality rather than by a substring.
func TestGraphQlNullableParityFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantId      string
		wantMessage string
		wantSpan    string
	}{
		// The two directions on a property. Both measured against the real ESLint rule on a seeded
		// file: same ids, same rendered text, same order.
		{
			sourceText:  "class C {\n  @GraphQlField(() => String, { nullable: true })\n  a!: string;\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'string' is not nullable",
			wantSpan:    "a",
		},
		{
			sourceText:  "class C {\n  @GraphQlField(() => String, { nullable: false })\n  a!: string | null;\n}\n",
			wantId:      "typeNullableButDecoratorNot",
			wantMessage: "Type 'string | null' is nullable but the decorator does not have 'nullable: true'",
			wantSpan:    "a",
		},

		// An ABSENT `nullable` is read as false, so an optional property reports. This is the third
		// outcome of the option read and the one a two-valued port would get wrong: it would skip
		// the decorator entirely and go silent here.
		{
			sourceText:  "class C {\n  @GraphQlField(() => String)\n  a?: string;\n}\n",
			wantId:      "typeNullableButDecoratorNot",
			wantMessage: "Type 'string | undefined' is nullable but the decorator does not have 'nullable: true'",
			wantSpan:    "a",
		},

		// `any` counts as nullable, which is the width of the shared mask rather than a detail: a
		// decorator claiming not-nullable beside a value typed any is a claim nothing checked.
		{
			sourceText:  "class C {\n  @GraphQlField(() => String, { nullable: false })\n  a!: any;\n}\n",
			wantId:      "typeNullableButDecoratorNot",
			wantMessage: "Type 'any' is nullable but the decorator does not have 'nullable: true'",
			wantSpan:    "a",
		},

		// A method return through the Promise unwrap. Without the unwrap the type is a Promise,
		// which is never nullable, so this case would report the OPPOSITE message and a port
		// missing the unwrap still looks like it works.
		{
			sourceText:  "class C {\n  @GraphQlQuery(() => Thing, { nullable: true })\n  async find(): Promise<Thing> {\n    return new Thing();\n  }\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'Thing' is not nullable",
			wantSpan:    "find",
		},
		{
			sourceText:  "class C {\n  @GraphQlMutation(() => Thing, { nullable: false })\n  async save(): Promise<Thing | null> {\n    return null;\n  }\n}\n",
			wantId:      "typeNullableButDecoratorNot",
			wantMessage: "Type 'Thing | null' is nullable but the decorator does not have 'nullable: true'",
			wantSpan:    "save",
		},
		{
			sourceText:  "class C {\n  @GraphQlFieldResolver(() => Thing, { nullable: true })\n  resolve(): Thing {\n    return new Thing();\n  }\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'Thing' is not nullable",
			wantSpan:    "resolve",
		},

		// An argument, which reports on the PARAMETER rather than on a name.
		{
			sourceText:  "class C {\n  @GraphQlQuery(() => Thing)\n  find(@GraphQlArgument('id', () => String, { nullable: true }) id: string): Thing {\n    void id;\n    return new Thing();\n  }\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'string' is not nullable",
			wantSpan:    "@GraphQlArgument('id', () => String, { nullable: true }) id: string",
		},

		// A nullable INPUT that admits undefined but not null: the client's null is what the type
		// says cannot happen. On an argument, reported on the parameter as the parity messages are.
		{
			sourceText:  "class C {\n  @GraphQlQuery(() => Thing)\n  find(@GraphQlArgument('id', () => String, { nullable: true }) id?: string): Thing {\n    void id;\n    return new Thing();\n  }\n}\n",
			wantId:      "decoratorNullableButInputExcludesNull",
			wantMessage: "Decorator declares 'nullable: true' on an input, so a client can send null, but the type 'string | undefined' does not admit null. Add '| null' so the code's checks follow what can arrive",
			wantSpan:    "@GraphQlArgument('id', () => String, { nullable: true }) id?: string",
		},
		// On an input class's field, optional and explicitly undefined alike.
		{
			sourceText:  "@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a?: string;\n}\n",
			wantId:      "decoratorNullableButInputExcludesNull",
			wantMessage: "Decorator declares 'nullable: true' on an input, so a client can send null, but the type 'string | undefined' does not admit null. Add '| null' so the code's checks follow what can arrive",
			wantSpan:    "a",
		},
		{
			sourceText:  "@GraphQlInputType()\nclass I {\n  @GraphQlField(() => Thing, { nullable: true })\n  a!: Thing | undefined;\n}\n",
			wantId:      "decoratorNullableButInputExcludesNull",
			wantMessage: "Decorator declares 'nullable: true' on an input, so a client can send null, but the type 'Thing | undefined' does not admit null. Add '| null' so the code's checks follow what can arrive",
			wantSpan:    "a",
		},
		// An input that is not nullable at all gets the parity message, not this one: the flag and
		// the type disagree outright, which is the first question asked.
		{
			sourceText:  "@GraphQlInputType()\nclass I {\n  @GraphQlField(() => String, { nullable: true })\n  a!: string;\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'string' is not nullable",
			wantSpan:    "a",
		},

		// The field decorator on a METHOD and on a GETTER, disagreeing. Their agreeing siblings are
		// in the silent list; this pair is what proves the method arm reports rather than merely
		// declining to crash.
		{
			sourceText:  "class C {\n  @GraphQlField(() => String, { nullable: true })\n  computed(): string {\n    return '';\n  }\n}\n",
			wantId:      "decoratorNullableButTypeNot",
			wantMessage: "Decorator declares 'nullable: true' but the type 'string' is not nullable",
			wantSpan:    "computed",
		},
		// A relation decorator that is NOT one of the three ORM ones does not exempt anything, so
		// the relation gate is pinned as a specific set rather than as "carries another decorator".
		{
			sourceText:  "class C {\n  @SomethingElse(() => Thing)\n  @GraphQlField(() => Thing, { nullable: false })\n  a?: Thing;\n}\n",
			wantId:      "typeNullableButDecoratorNot",
			wantMessage: "Type 'Thing | undefined' is nullable but the decorator does not have 'nullable: true'",
			wantSpan:    "a",
		},
	}
	for index, testCase := range cases {
		t.Run(graphQlNullableParityCaseName(index), func(t *testing.T) {
			t.Parallel()
			sourceText := graphQlNullableParityPreamble + testCase.sourceText
			result := rule_testing.RunTyped(t, GraphQlNullableParity, graphQlNullableParityFile,
				sourceText)

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

// TestGraphQlNullableParityRequiresTheTypedHarness pins that the rule declines rather than crashing
// without a checker, which no other test here can reach.
func TestGraphQlNullableParityRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !GraphQlNullableParity.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: one side of every comparison is a type")
	}

	sourceText := graphQlNullableParityPreamble +
		"class C {\n  @GraphQlField(() => String, { nullable: true })\n  a!: string;\n}\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, GraphQlNullableParity,
		graphQlNullableParityFile, sourceText))

	// The control, so the silence above is the guard declining rather than the rule being blind.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, GraphQlNullableParity,
		graphQlNullableParityFile, sourceText), "decoratorNullableButTypeNot")
}
