package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// securityRequireContextAccessFile is where the fixtures pretend to live.
const securityRequireContextAccessFile = "/repository/source/SecurityRequireContextAccess.ts"

// securityRequireContextAccessLiveOptions is the wiring api-phi-health actually runs.
//
// Taken verbatim from `BaseLintConfiguration.ts`, because the oracle every expectation below was
// measured against was driven with exactly this configuration. Inventing a simpler one would have
// made the fixtures agree with a rule nobody runs.
func securityRequireContextAccessLiveOptions() any {
	return securityRequireContextAccessOptions(`[
		{"contextKey":"AccountRequestContextKey",
		 "requiresAny":["RequireSessionAccess","WithSessionAccess"]},
		{"contextKey":"AuthenticationRequestContextKey",
		 "requiresAny":["RequireAuthenticationSession","RequireSessionAccess","WithSessionAccess"]},
		{"contextKey":"DeviceIdRequestContextKey",
		 "requiresAny":["RequireDeviceId","RequireSessionAccess","WithSessionAccess"]}
	]`)
}

// securityRequireContextAccessOptions builds options from a requirements array, through the real decoder.
//
// Through the decoder rather than by building the struct, because the decoder is where the two
// schema constraints live -- both fields required, and `requiresAny` non-empty -- and neither has
// an upstream counterpart to inherit correctness from.
func securityRequireContextAccessOptions(requirementsJson string) any {
	decoded, err := DecodeContextRequiresAccessOptions(
		[]byte(`{"requirements":` + requirementsJson + `}`))
	if err != nil {
		panic("the fixture options do not decode: " + err.Error())
	}
	return decoded
}

// securityRequireContextAccessCase is one measured row.
type securityRequireContextAccessCase struct {
	name       string
	sourceText string
	options    any
	// wantKeys is the context key each finding names, in order. nil means the case is clean.
	wantKeys []string
}

func runContextRequiresAccess(
	t *testing.T,
	testCase securityRequireContextAccessCase,
) rule_testing.Result {
	t.Helper()
	return rule_testing.RunWithOptions(t, SecurityRequireContextAccess, securityRequireContextAccessFile,
		testCase.sourceText, testCase.options)
}

// These fixtures are MEASURED rather than imported, because this rule is ours and has no upstream
// corpus.
//
// The brief warns that a fixture a porter invents encodes the same belief as the port, so it passes
// for exactly the reason the code is wrong. With no corpus that risk is live rather than
// theoretical, and the mitigation is the source repository: the original
// `SecurityRequireContextAccessRule.ts` was loaded into the ESLint Linter with the live wiring from
// `BaseLintConfiguration.ts` and driven over forty-six inputs. Every expectation below is what that
// run answered, not what the source reads like it should answer.
//
// Two of those answers contradict a careful reading of the source, and both are pinned below in
// TestContextRequiresAccessSurprises. Three more inputs turned out to be parse errors rather than
// cases, and are recorded in TestContextRequiresAccessUnmeasurableShapes.
func securityRequireContextAccessFiresCases() []securityRequireContextAccessCase {
	return []securityRequireContextAccessCase{
		{"bare: unprotected reports", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"bare: a decorator for a different key does not protect", "class A { @RequireDeviceId() m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"class decorator, wrong one", "@RequireDeviceId() class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"member-access key resolves to the property", "class A { m(@InjectRequestContext(Keys.AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"no annotation at all is NOT soft use", "class A { m(@InjectRequestContext(AccountRequestContextKey) a) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"parameter property, unprotected", "class A { constructor(@InjectRequestContext(AccountRequestContextKey) private a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"defaulted parameter, unprotected", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string = 'x') {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"two protected keys, both unprotected", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string, @InjectRequestContext(DeviceIdRequestContextKey) b: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey", "DeviceIdRequestContextKey"}},
		{"two keys, decorator satisfies only one", "class A { @RequireDeviceId() m(@InjectRequestContext(AccountRequestContextKey) a: string, @InjectRequestContext(DeviceIdRequestContextKey) b: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"same key twice reports once", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string, @InjectRequestContext(AccountRequestContextKey) b: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"import alias resolves to the exported name", "import { AccountRequestContextKey as AK } from 'x';\nclass A { m(@InjectRequestContext(AK) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"non-aliased import is unaffected", "import { AccountRequestContextKey } from 'x';\nclass A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"computed method name", "class A { [key](@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"static method, unprotected", "class A { static m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"getter with a decorated parameter is not a method definition shape", "class A { set m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
		{"authentication key, unprotected", "class A { m(@InjectRequestContext(AuthenticationRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AuthenticationRequestContextKey"}},
		{"namespaced access decorator does not match", "class A { @Access.RequireSessionAccess() m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), []string{"AccountRequestContextKey"}},
	}
}

func securityRequireContextAccessSilentCases() []securityRequireContextAccessCase {
	return []securityRequireContextAccessCase{
		{"bare: method decorator protects", "class A { @RequireSessionAccess() m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"bare: the other alternative also protects", "class A { @WithSessionAccess() m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"class decorator protects", "@RequireSessionAccess() class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"unconfigured key is ignored", "class A { m(@InjectRequestContext(SomeOtherKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"computed member access does not resolve", "class A { m(@InjectRequestContext(Keys['AccountRequestContextKey']) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: optional parameter", "class A { m(@InjectRequestContext(AccountRequestContextKey) a?: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: union with undefined", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string | undefined) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: union with null", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string | null) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: any", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: any) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: unknown", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: unknown) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: void", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: void) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"soft use: bare undefined annotation", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: undefined) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"nested union counts as soft use", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string | (number | null)) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"GraphQlFieldResolver exempts the method", "class A { @GraphQlFieldResolver() m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"parameter property, optional is soft use", "class A { constructor(@InjectRequestContext(AccountRequestContextKey) private a?: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"defaulted parameter, optional-typed is soft use", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string | undefined = undefined) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"two keys, one decorator satisfies both", "class A { @RequireSessionAccess() m(@InjectRequestContext(AccountRequestContextKey) a: string, @InjectRequestContext(DeviceIdRequestContextKey) b: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"import alias, protected", "import { AccountRequestContextKey as AK } from 'x';\nclass A { @RequireSessionAccess() m(@InjectRequestContext(AK) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"an alias TO a configured name from something else", "import { Something as AccountRequestContextKey } from 'x';\nclass A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"a different decorator entirely is ignored", "class A { m(@InjectSomethingElse(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"decorator with no argument", "class A { m(@InjectRequestContext() a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"bare decorator, not a call", "class A { m(@InjectRequestContext a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"authentication key, its own alternative", "class A { @RequireAuthenticationSession() m(@InjectRequestContext(AuthenticationRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"device key, its own alternative", "class A { @RequireDeviceId() m(@InjectRequestContext(DeviceIdRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"bare access decorator, not a call", "class A { @RequireSessionAccess m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"empty requirements: rule is inert", "class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessOptions("[]"), nil},
		{"parameter property, class-protected", "@RequireSessionAccess() class A { constructor(@InjectRequestContext(AccountRequestContextKey) private readonly a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"alias TO a configured name (recheck)", "import { Something as AccountRequestContextKey } from 'x';\nclass A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"bare access decorator, not a call (recheck)", "class A { @RequireSessionAccess m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
		{"bare access decorator on the class", "@RequireSessionAccess class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }", securityRequireContextAccessLiveOptions(), nil},
	}
}

func TestContextRequiresAccessFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range securityRequireContextAccessFiresCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			wantIds := make([]string, len(testCase.wantKeys))
			for index := range testCase.wantKeys {
				wantIds[index] = "missingProtector"
			}
			rule_testing.ExpectFindings(t, runContextRequiresAccess(t, testCase), wantIds...)
		})
	}
}

func TestContextRequiresAccessStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range securityRequireContextAccessSilentCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, runContextRequiresAccess(t, testCase))
		})
	}
}

// Which KEY each finding names, which the message-id assertions above cannot see.
//
// A method injecting two protected keys reports twice, and the two findings must name different
// keys. A rule that found the right methods through the wrong key would report the right count with
// the wrong word, and would tell the reader to add a decorator that fixes nothing.
func TestContextRequiresAccessNamesTheKey(t *testing.T) {
	t.Parallel()

	for _, testCase := range securityRequireContextAccessFiresCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runContextRequiresAccess(t, testCase)
			if len(result.Diagnostics) != len(testCase.wantKeys) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantKeys),
					len(result.Diagnostics))
			}
			for index, wantKey := range testCase.wantKeys {
				got := result.Diagnostics[index].Message.Description
				needle := "injects `" + wantKey + "`"
				if !strings.Contains(got, needle) {
					t.Errorf("finding %d rendered\n  %q\nwhich does not name the key %q",
						index, got, wantKey)
				}
			}
		})
	}
}

// Two answers that contradict a careful reading of the source, both measured.
//
// These are the rows that justify building an oracle at all. Reading the source, both look like they
// should go the other way, and a port written from the source alone would have got both wrong while
// every other fixture stayed green.
//
//	an alias TO a configured name    `import { Something as AccountRequestContextKey }` reads as
//	                                 though it should protect, since the injection site spells the
//	                                 configured key. It does NOT: the alias map is keyed by LOCAL
//	                                 name, so the configured spelling maps AWAY to `Something`.
//	                                 That is correct rather than unfortunate -- the local name is a
//	                                 different symbol that happens to share a spelling.
//	a bare access decorator          `@RequireSessionAccess` without parentheses reads as though it
//	                                 should not match, because the exemption check next to it
//	                                 accepts calls only. It DOES: the access check uses a different
//	                                 helper that also accepts a bare identifier.
func TestContextRequiresAccessSurprises(t *testing.T) {
	t.Parallel()

	cases := []securityRequireContextAccessCase{
		{
			name: "an alias TO a configured name protects nothing and is clean",
			sourceText: "import { Something as AccountRequestContextKey } from 'x';\n" +
				"class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: nil,
		},
		{
			name: "the control: the same import unaliased DOES report",
			sourceText: "import { AccountRequestContextKey } from 'x';\n" +
				"class A { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name: "a bare access decorator with no parentheses protects",
			sourceText: "class A { @RequireSessionAccess " +
				"m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: nil,
		},
		{
			name: "the control: a bare decorator of the WRONG name does not protect",
			sourceText: "class A { @RequireDeviceId " +
				"m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name: "a namespaced access decorator does NOT protect, called or not",
			sourceText: "class A { @Access.RequireSessionAccess() " +
				"m(@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name: "but a namespaced KEY does resolve, by its property",
			sourceText: "class A { " +
				"m(@InjectRequestContext(Keys.AccountRequestContextKey) a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name:       "an unannotated parameter is NOT soft use, though implicit any would be",
			sourceText: "class A { m(@InjectRequestContext(AccountRequestContextKey) a) {} }",
			options:    securityRequireContextAccessLiveOptions(),
			wantKeys:   []string{"AccountRequestContextKey"},
		},
		{
			name:       "the control: an explicit any IS soft use",
			sourceText: "class A { m(@InjectRequestContext(AccountRequestContextKey) a: any) {} }",
			options:    securityRequireContextAccessLiveOptions(),
			wantKeys:   nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runContextRequiresAccess(t, testCase)
			if len(testCase.wantKeys) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			wantIds := make([]string, len(testCase.wantKeys))
			for index := range testCase.wantKeys {
				wantIds[index] = "missingProtector"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// Three shapes OUR parser accepts and the ESLint oracle rejects, so their verdicts are unmeasured.
//
// This is a parser difference rather than a rule difference, and it is the one thing about this port
// that has no analogue in an upstream port: normally the oracle and the target parse the same
// language. Here `@typescript-eslint/parser` answers `Decorators are not valid here` for a decorated
// constructor, a decorated parameter on a plain function, and decorators inside a class expression,
// while typescript-go parses all three without a diagnostic.
//
// So this rule WILL meet them and the oracle cannot say what should happen. Rather than invent a
// verdict, the behaviour is asserted as whatever the ported logic does, with the reasoning recorded:
//
//	a decorated constructor    treated like any other member, which is what the source's own
//	                          `MethodDefinition` visitor does for constructors in ESTree.
//	a plain function           NOT a member, so no listener sees it, matching the source's visitor.
//	a class expression         its members are visited and its decorators collected, because the
//	                          source's class lookup accepts `ClassExpression` explicitly.
//
// Measured separately: none of the three occurs anywhere in api-phi-health's source today, so
// nothing turns on these answers in practice. If one appears, this test is where the decision was
// recorded rather than inferred.
func TestContextRequiresAccessUnmeasurableShapes(t *testing.T) {
	t.Parallel()

	cases := []securityRequireContextAccessCase{
		{
			name: "a decorated constructor is a member like any other",
			sourceText: "class A { @RequireSessionAccess() constructor(" +
				"@InjectRequestContext(AccountRequestContextKey) private a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: nil,
		},
		{
			name: "an unprotected constructor reports",
			sourceText: "class A { constructor(" +
				"@InjectRequestContext(AccountRequestContextKey) private a: string) {} }",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name:       "a plain function is not a member and is never checked",
			sourceText: "function m(@InjectRequestContext(AccountRequestContextKey) a: string) {}",
			options:    securityRequireContextAccessLiveOptions(),
			wantKeys:   nil,
		},
		{
			name: "a class expression's members are checked",
			sourceText: "const A = class { " +
				"m(@InjectRequestContext(AccountRequestContextKey) a: string) {} };",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: []string{"AccountRequestContextKey"},
		},
		{
			name: "and a class expression's decorators protect its members",
			sourceText: "const A = @RequireSessionAccess() class { " +
				"m(@InjectRequestContext(AccountRequestContextKey) a: string) {} };",
			options:  securityRequireContextAccessLiveOptions(),
			wantKeys: nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := runContextRequiresAccess(t, testCase)
			if len(testCase.wantKeys) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			wantIds := make([]string, len(testCase.wantKeys))
			for index := range testCase.wantKeys {
				wantIds[index] = "missingProtector"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)
		})
	}
}

// With no requirements configured the rule registers NO listener at all.
//
// This is the source's `if(requirementMap.size === 0) return {}`, and it is the property that makes
// the rule's finding count on an unwired project structurally zero rather than evidence of clean
// code. It is also why this port ships registered and unenabled: enabling it without requirements
// would look enabled while enforcing nothing.
//
// Asserted on the listener set rather than on findings, because that is the layer where the
// difference exists -- an empty requirement map matches nothing either way, so no fixture over
// findings could tell a guarded rule from an unguarded one.
func TestContextRequiresAccessRegistersNothingUnconfigured(t *testing.T) {
	t.Parallel()

	listenersFor := func(options any) rule.Listeners {
		return SecurityRequireContextAccess.Run(rule.Context{}, options)
	}

	if got := len(listenersFor(nil)); got != 0 {
		t.Errorf("nil options registered %d listeners, wanted none", got)
	}
	if got := len(listenersFor(SecurityRequireContextAccessOptions{})); got != 0 {
		t.Errorf("an empty requirement list registered %d listeners, wanted none", got)
	}
	// The control: a configured requirement registers, so the zeros above are the guard rather
	// than a rule that never registers at all.
	if got := len(listenersFor(securityRequireContextAccessLiveOptions())); got == 0 {
		t.Error("the live configuration registered no listeners")
	}
}

// The decoder, which enforces the two constraints the source states in its JSON schema.
//
// We have no schema layer, so an entry missing `contextKey` or carrying an empty `requiresAny` would
// be accepted silently. The second is the dangerous one: it reports on every injection of that key
// with a message naming no decorator to add.
func TestDecodeContextRequiresAccessOptions(t *testing.T) {
	t.Parallel()

	t.Run("nil input yields an empty list", func(t *testing.T) {
		t.Parallel()
		decoded, err := DecodeContextRequiresAccessOptions(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(decoded.(SecurityRequireContextAccessOptions).Requirements) != 0 {
			t.Error("nil input produced requirements")
		}
	})

	t.Run("a requirement with no contextKey is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeContextRequiresAccessOptions(
			[]byte(`{"requirements":[{"requiresAny":["A"]}]}`)); err == nil {
			t.Error("a requirement naming no key decoded")
		}
	})

	t.Run("an empty requiresAny is rejected", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeContextRequiresAccessOptions(
			[]byte(`{"requirements":[{"contextKey":"K","requiresAny":[]}]}`)); err == nil {
			t.Error("a requirement nothing could satisfy decoded")
		}
	})

	t.Run("the live wiring decodes", func(t *testing.T) {
		t.Parallel()
		if _, err := DecodeContextRequiresAccessOptions([]byte(
			`{"requirements":[{"contextKey":"AccountRequestContextKey",` +
				`"requiresAny":["RequireSessionAccess","WithSessionAccess"]}]}`)); err != nil {
			t.Errorf("the configuration api-phi-health runs did not decode: %v", err)
		}
	})
}

// Parameter shapes that must not take the linter down.
//
// A probe written while checking whether this rule needed the upstream `resolveDecoratedParameterNode`
// helper panicked on the third of these: `Node.Text()` raises `Unhandled case in Node.Text:
// *ast.BindingPattern` for a destructured parameter name. The walk recovers per FILE rather than per
// rule, so one such panic costs every rule its verdict on that file.
//
// This rule is safe by construction rather than by guard -- it never reads a parameter's name, only
// its decorators, its type annotation and its question token, and it reports on the enclosing
// member's name rather than the parameter's. These rows turn that argument into a measurement, since
// an argument about what a rule does not call is exactly the kind that stops being true after an
// edit.
func TestContextRequiresAccessSurvivesUnusualParameters(t *testing.T) {
	t.Parallel()

	for _, source := range []string{
		// A destructured parameter, whose name is a binding pattern rather than an identifier.
		"class A { m(@InjectRequestContext(AccountRequestContextKey) { a }: { a: string }) {} }",
		"class A { m(@InjectRequestContext(AccountRequestContextKey) [a]: string[]) {} }",
		// A rest parameter.
		"class A { m(@InjectRequestContext(AccountRequestContextKey) ...rest: string[]) {} }",
		// A computed member name, where the enclosing member's own name is not an identifier
		// either, so the report target falls back to the member itself.
		"class A { [computed](@InjectRequestContext(AccountRequestContextKey) a: string) {} }",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			// The assertion is that this returns at all. What it reports is covered above; a panic
			// here would be invisible to every other test in this file, which all use ordinary
			// parameters.
			rule_testing.RunWithOptions(t, SecurityRequireContextAccess, securityRequireContextAccessFile,
				source, securityRequireContextAccessLiveOptions())
		})
	}
}
