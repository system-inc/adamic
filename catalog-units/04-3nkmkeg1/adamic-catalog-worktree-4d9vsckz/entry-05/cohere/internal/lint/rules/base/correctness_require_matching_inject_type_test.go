package base

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessRequireMatchingInjectTypeFile = "/repository/source/Injection.ts"

func correctnessRequireMatchingInjectTypeCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestInjectTypeMatchesParameterStaysSilent covers every shape the original declines.
//
// One of Kirk's own rules, so there is no corpus and every case is invented, which is the situation
// the brief warns makes a fixture worthless as evidence: it encodes the same belief as the port. The
// mitigation is that the original is executable. Each source below was written into
// api-phi-health, linted through the real rule with a real program, and the verdict recorded here is
// what the original produced.
//
// A type-aware rule needs the program, so these went through the project service on a real file
// rather than through a string, and the probe captured non-rule messages alongside the findings. It
// found no instrument noise on any of the twelve, which is worth saying because the sibling rule's
// first measurement produced three false zeros that read exactly like clean results.
func TestInjectTypeMatchesParameterStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// The correct pairing. This row is what the undefined-strip exists for: the brand field is
		// declared optional, so without stripping it the brand reads as `Service | undefined`, which is
		// not assignable to a parameter typed `Service`, and this case reports. Probed before the rule
		// was written; the rule would have inverted without it.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private right: Service) {} }\n",
		// A plain ParameterDecorator has no brand to read, so there is no contract and nothing to check.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@PlainInject() private anything: Other) {} }\n",
		// A brand of `any` means the decorator carried no concrete contract, which is a site not yet
		// migrated to a typed injection key. Checking it would noise on every legacy call.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<any>() private loose: Other) {} }\n",
		// The same for `unknown`, which the original tests in the same flag comparison.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<unknown>() private loose: Other) {} }\n",
		// Assignability runs brand-to-parameter, so a SUBtype resolving into a wider parameter is fine.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<SubService>() private s: Service) {} }\n",
		// A parameter with no annotation at all. Its type is implicitly `any`, which everything is
		// assignable to, so the comparison passes rather than the rule declining.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private u) {} }\n",
		// A union parameter that includes the resolved type.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private u: Service | Other) {} }\n",
		// A branded decorator on something that is NOT a parameter. The contract only means anything
		// on a parameter, and no other case here places one elsewhere, so a mutant dropping the
		// parameter-kind test survived the whole suite. All three placements measured clean against
		// the original.
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nexport class C { @TypedInject<Service>() field: Other = new Other(); }\n",
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nexport class C { @TypedInject<Service>() method(): Other { return new Other(); } }\n",
		"type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\n@TypedInject<Service>() export class C { x = 1; }\n",
	}
	for index, sourceText := range cases {
		t.Run(correctnessRequireMatchingInjectTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingInjectType,
				correctnessRequireMatchingInjectTypeFile, sourceText))
		})
	}
}

// TestInjectTypeMatchesParameterFires covers the shapes the original reports.
//
// Every row asserts the rendered message as well as the span. The message names BOTH types by their
// printed names and the decorator by its own, so it is where a port comparing the wrong pair, or
// printing the namespace instead of the member, shows itself. The span is the parameter rather than
// the decorator, which is the original's choice and is the half a reader changes.
func TestInjectTypeMatchesParameterFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText  string
		wantSpan    string
		wantMessage string
	}{
		// The motivating shape.
		{
			sourceText:  "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private wrong: Other) {} }\n",
			wantSpan:    "wrong: Other",
			wantMessage: "@TypedInject resolves to 'Service' but parameter is typed as 'Other'",
		},
		// The other direction of the same comparison: a supertype resolving into a narrower parameter
		// is a real mismatch. This row is what keeps the assignability arguments from being swapped,
		// which the matched case alone cannot see.
		{
			sourceText:  "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private s: SubService) {} }\n",
			wantSpan:    "s: SubService",
			wantMessage: "@TypedInject resolves to 'Service' but parameter is typed as 'SubService'",
		},
		// A defaulted parameter. Our parser puts the default on the parameter node, so all three
		// parameter shapes are one node kind here; the original consolidates four drifting copies of
		// this resolution and records that one read the wrong node's type for exactly this shape.
		{
			sourceText:  "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@TypedInject<Service>() private d: Other = new Other()) {} }\n",
			wantSpan:    "d: Other = new Other()",
			wantMessage: "@TypedInject resolves to 'Service' but parameter is typed as 'Other'",
		},
		// A plain method parameter rather than a constructor parameter property.
		{
			sourceText:  "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { method(@TypedInject<Service>() plain: Other) {} }\n",
			wantSpan:    "plain: Other",
			wantMessage: "@TypedInject resolves to 'Service' but parameter is typed as 'Other'",
		},
		// A NAMESPACED decorator, which prints the property name rather than the namespace. The
		// message says `@inject` and not `@Namespace`, which no other row can distinguish.
		{
			sourceText:  "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\ndeclare function PlainInject(): ParameterDecorator;\ndeclare const Namespace: { inject<T>(): TypedParameterDecorator<T> };\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nclass SubService extends Service { extra = true; }\nexport class C { constructor(@Namespace.inject<Service>() private n: Other) {} }\n",
			wantSpan:    "n: Other",
			wantMessage: "@inject resolves to 'Service' but parameter is typed as 'Other'",
		},
	}
	for index, testCase := range cases {
		t.Run(correctnessRequireMatchingInjectTypeCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, CorrectnessRequireMatchingInjectType,
				correctnessRequireMatchingInjectTypeFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "mismatch")

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			diagnostic := result.Diagnostics[0]
			gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
			if gotSpan != testCase.wantSpan {
				t.Fatalf("span: expected %q, got %q", testCase.wantSpan, gotSpan)
			}
			if diagnostic.Message.Description != testCase.wantMessage {
				t.Fatalf("message: expected %q, got %q", testCase.wantMessage,
					diagnostic.Message.Description)
			}
		})
	}
}

// TestInjectTypeMatchesParameterRequiresTheTypedHarness pins the nil-checker guard.
//
// A mutant removing that guard survives every other fixture here, and structurally so: the guard
// prevents a PANIC and no findings assertion can see one. It matters more than usual for a rule in
// this package, because the walk recovers per FILE rather than per rule, so one nil dereference
// costs every rule its verdict on that whole file.
//
// The untyped harness hands the rule a nil checker, so silence there is the guard working. The
// second half is the control: without something that fires, a rule which can never report at all
// satisfies the first half.
func TestInjectTypeMatchesParameterRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "type TypedParameterDecorator<TResolved> = ParameterDecorator & {\n    readonly __resolvedType?: TResolved;\n};\ndeclare function TypedInject<T>(): TypedParameterDecorator<T>;\nclass Service { serviceMarker = 1; }\nclass Other { otherMarker = 'x'; }\nexport class C { constructor(@TypedInject<Service>() private wrong: Other) {} }\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, CorrectnessRequireMatchingInjectType,
		correctnessRequireMatchingInjectTypeFile, sourceText))

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, CorrectnessRequireMatchingInjectType,
		correctnessRequireMatchingInjectTypeFile, sourceText), "mismatch")
}
