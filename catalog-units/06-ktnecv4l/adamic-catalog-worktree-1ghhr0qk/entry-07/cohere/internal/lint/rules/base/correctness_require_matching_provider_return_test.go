package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// providerReturnFile is where the fixtures pretend to live.
const providerReturnFile = "/repository/source/ProviderReturn.ts"

// providerReturnPreamble declares the brand infrastructure every fixture needs.
//
// A phantom brand is the whole opt-in mechanism, so a fixture without it exercises nothing: the rule
// would find no property, decline, and every case would pass for the wrong reason. This preamble is
// a restatement of api-phi-health's own `BrandedMethodDecoratorType` and `Provider` signature, and
// the shapes were checked against the real ones rather than invented.
const providerReturnPreamble = `
type ObjectFactory<T> = { create(): T };
type BrandedMethodDecoratorType<TExpectedReturn> = MethodDecorator & {
    readonly __expectedReturnType?: TExpectedReturn;
};
declare class TypedInjectionKey<T = unknown> { private brand?: T; }
declare function Provider<T = unknown>(
    token: TypedInjectionKey<T>,
): BrandedMethodDecoratorType<T | ObjectFactory<T> | undefined>;
declare function Plain(): MethodDecorator;
declare function Loose(): BrandedMethodDecoratorType<any>;
declare function Unknowable(): BrandedMethodDecoratorType<unknown>;
declare function Strict(): BrandedMethodDecoratorType<string>;
declare const stringToken: TypedInjectionKey<string>;
declare namespace Namespace {
    function Strict(): BrandedMethodDecoratorType<string>;
}
`

// These fixtures are WRITTEN rather than imported, which inverts the usual advice.
//
// This is one of our own base rules, so there is no upstream corpus, and a fixture written beside
// the port carries the same beliefs the port does. The mitigation is that the source repository is a
// real oracle, and for this rule the oracle is unusually strong: the check ran against
// api-phi-health's ACTUAL `Provider` decorator and `TypedInjectionKey` rather than a synthetic
// stand-in, on a file seeded inside that project's tsconfig include. The original and this port
// agreed on two reporting shapes and three clean ones, and the rendered types matched character for
// character, including `string | ObjectFactory<string> | undefined`.
//
// The whole api-phi-health tree reports zero for this rule, which is a real property of that tree
// rather than a broken instrument: the same run linted 2310 files and reported 44 findings from
// other rules. A clean tree cannot validate a port, which is why the seeded file exists.
func TestProviderReturnMatchesTokenFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
	}{
		{"a return type outside the token's contract",
			"@Provider(stringToken) bad(): number { return 1; }"},
		{"a union only partly assignable",
			"@Provider(stringToken) partly(): string | number { return 'x'; }"},
		{"a concrete brand with a mismatched return",
			"@Strict() strict(): number { return 1; }"},
		{"a namespaced decorator, whose name still renders",
			"@Namespace.Strict() namespaced(): number { return 1; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn,
				providerReturnFile,
				providerReturnPreamble+"class C { "+testCase.method+" }"), "mismatch")
		})
	}
}

// The clean cases, and three of them are where this rule's real discrimination lives.
//
// The `string | undefined` row is the one that would break under the wrong undefined handling. Its
// sibling rule in api-phi-health strips `undefined` from the brand, and this one deliberately does
// not, because `Provider<T>`'s contract intrinsically includes `undefined` through its opt-out path.
// Stripping it here would reject every optional provider, which is most of them. Verified against
// the original on the seeded file: it is silent on exactly this shape.
//
// The any and unknown rows pin the other skip: a brand whose type argument is not concrete carries
// no contract, and checking it would be theatre because assignability answers yes for everything.
func TestProviderReturnMatchesTokenStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
	}{
		{"a return type inside the contract",
			"@Provider(stringToken) good(): string { return 'x'; }"},
		{"an optional provider, since undefined is part of the contract",
			"@Provider(stringToken) optional(): string | undefined { return undefined; }"},
		{"an object factory, which the contract also admits",
			"@Provider(stringToken) factory(): ObjectFactory<string> { return { create: () => 'x' }; }"},
		{"a plain MethodDecorator declares no contract",
			"@Plain() other(): number { return 1; }"},
		{"no decorator at all", "plain(): number { return 1; }"},
		{"a brand resolving to any carries no contract",
			"@Loose() loose(): number { return 1; }"},
		{"a brand resolving to unknown carries no contract",
			"@Unknowable() unknowable(): number { return 1; }"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn,
				providerReturnFile,
				providerReturnPreamble+"class C { "+testCase.method+" }"))
		})
	}
}

// A decorator on anything other than a method is not this rule's business.
//
// A decorator can sit on a class, a property, an accessor or a parameter, and all of those reach
// this listener. The original requires a MethodDefinition parent and so does this. Each of these
// carries a real brand, so they would report if the parent test were dropped.
func TestProviderReturnMatchesTokenOnlyJudgesMethods(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		classMember string
	}{
		{"a property", "@Strict() property: number = 1;"},
		{"a getter", "@Strict() get value(): number { return 1; }"},
		{"a setter", "@Strict() set value(v: number) {}"},
		// The row that actually proves the parent guard, and the only one that does.
		//
		// Written for a surviving mutant, and the first hypothesis about that survivor was wrong.
		// The three rows above pass with OR without the guard: a property, a getter and a setter
		// have no call signature, so the rule exits at the return-type lookup rather than at the
		// parent test, and dropping the guard changes nothing for them.
		//
		// A property whose TYPE is a function does have a call signature, so it is the one shape
		// that separates the two versions: with the guard it is clean, without it, it reports.
		// Measured against the original on a seeded file inside api-phi-health's tsconfig, using
		// the real Provider decorator: it reports the equivalent real method and stays silent on
		// the function-typed property, so declining is correct rather than merely convenient.
		{"a property whose type is a function, which does have a call signature",
			"@Strict() functionProperty: () => number = () => 1;"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn,
				providerReturnFile,
				providerReturnPreamble+"class C { "+testCase.classMember+" }"))
		})
	}
}

// The message renders both types and the decorator's name, and it is asserted exactly.
//
// A message-id assertion cannot see anything the format string does, and this one interpolates three
// values, so a slot filled from the wrong variable reads as a correct finding. The rendered strings
// here were taken from the ORIGINAL rule's output on the seeded file rather than derived: it printed
// `string | ObjectFactory<string> | undefined` and `number` for the same input.
//
// Asserted against literals typed here rather than against the rule's own constant, since comparing
// to the constant moves both sides together under mutation.
func TestProviderReturnMatchesTokenRendersBothTypes(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn, providerReturnFile,
		providerReturnPreamble+"class C { @Provider(stringToken) bad(): number { return 1; } }")
	rule_testing.ExpectFindings(t, result, "mismatch")

	rendered := result.Diagnostics[0].Message.Description
	for _, want := range []string{
		"@Provider",
		"`string | ObjectFactory<string> | undefined`",
		"`number`",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("expected the message to carry %s, got %q", want, rendered)
		}
	}
	// The order matters as much as the presence: a message naming the two types the wrong way round
	// tells the reader to change the wrong side, and both substrings are present either way.
	expectedIndex := strings.Index(rendered, "`string | ObjectFactory<string> | undefined`")
	returnIndex := strings.Index(rendered, "`number`")
	if expectedIndex > returnIndex {
		t.Errorf("expected the contract before the actual return type, got %q", rendered)
	}

	namespaced := rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn, providerReturnFile,
		providerReturnPreamble+"class C { @Namespace.Strict() m(): number { return 1; } }")
	rule_testing.ExpectFindings(t, namespaced, "mismatch")
	if !strings.Contains(namespaced.Diagnostics[0].Message.Description, "@Strict") {
		t.Errorf("expected a namespaced decorator to render its property name, got %q",
			namespaced.Diagnostics[0].Message.Description)
	}
}

// The finding points at the method, not at the decorator.
//
// The original reports the method's value node, which is the function carrying the return type the
// author has to change. Every assertion above is satisfied by a rule reporting the decorator
// instead, and pointing at the decorator would show the reader the line they must NOT edit.
func TestProviderReturnMatchesTokenPointsAtTheMethod(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn, providerReturnFile,
		providerReturnPreamble+"class C { @Strict() m(): number { return 1; } }")
	rule_testing.ExpectFindings(t, result, "mismatch")

	source := result.SourceFile.Text()
	reported := source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if !strings.HasPrefix(reported, "@Strict() m(): number") {
		t.Errorf("expected the finding on the method declaration, pointed at %q", reported)
	}
	if !strings.Contains(reported, "number") {
		t.Errorf("expected the reported span to include the return type the author must change, "+
			"pointed at %q", reported)
	}
}

// The typed harness is required, and a plain run must not look clean.
//
// Every judgment in this rule is a checker question, so the plain harness hands it a nil checker and
// the rule goes completely silent. That is a vacuous green rather than an obvious crash, which is
// the more dangerous of the two failure modes: a later revert to rule_testing.Run would leave every
// Fires case above passing for no reason. This asserts the guard is doing what its comment says.
func TestProviderReturnMatchesTokenNeedsTheTypedHarness(t *testing.T) {
	t.Parallel()

	source := providerReturnPreamble + "class C { @Strict() m(): number { return 1; } }"

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingProviderReturn,
		providerReturnFile, source), "mismatch")

	// The same input through the untyped harness reports nothing, because the rule declines when it
	// has no checker rather than dereferencing one.
	rule_testing.ExpectClean(t, rule_testing.Run(t, CorrectnessRequireMatchingProviderReturn,
		providerReturnFile, source))
}
