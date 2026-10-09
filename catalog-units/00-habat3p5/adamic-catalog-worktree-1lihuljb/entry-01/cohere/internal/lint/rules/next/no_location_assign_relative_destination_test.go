package next

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const locationAssignFile = "/repository/source/Navigate.tsx"

// TestNoLocationAssignRelativeDestinationFires covers every input measured to report.
//
// There is no imported corpus for this rule: it has no oxc implementation, and the published
// `@next/eslint-plugin-next` package ships no tests. Every case below was established by running
// the reference rule through ESLint's Linter API on that exact source and reading the verdict,
// rather than by reading the reference and predicting one. The reference needs `window`,
// `location`, `document` and `self` declared as globals before it will report at all, because its
// global check reads the global scope's binding set; without them every input is silent and a
// silent probe is indistinguishable from a decline.
func TestNoLocationAssignRelativeDestinationFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare location.assign with a rooted path", "location.assign('/dashboard');\n"},
		{"a window receiver", "window.location.assign('/dashboard');\n"},
		{"a globalThis receiver", "globalThis.location.assign('/dashboard');\n"},
		{"a document receiver", "document.location.assign('/dashboard');\n"},
		{"a self receiver", "self.location.assign('/dashboard');\n"},
		{"a computed assign key", "location['assign']('/dashboard');\n"},
		{"a computed location key", "window['location'].assign('/dashboard');\n"},
		{"an href assignment", "location.href = '/dashboard';\n"},
		{"an href assignment through window", "window.location.href = '/dashboard';\n"},
		{"a plus-equals href assignment", "location.href += '/dashboard';\n"},
		{"a minus-equals href assignment", "location.href -= '/dashboard';\n"},
		{"a logical-or-equals href assignment", "location.href ||= '/dashboard';\n"},
		{"a nullish-equals href assignment", "location.href ??= '/dashboard';\n"},
		{"a computed href key", "location['href'] = '/dashboard';\n"},
		{"a path with no leading slash", "location.assign('dashboard');\n"},
		{"a dot-slash relative path", "location.assign('./dashboard');\n"},
		{"a dot-dot relative path", "location.assign('../dashboard');\n"},
		{"an empty string", "location.assign('');\n"},
		{"a fragment-only destination", "location.assign('#section');\n"},
		{"a query-only destination", "location.assign('?q=1');\n"},
		{"a template whose first span is text", "location.href = `/users/${id}`;\n"},
		{"a template opening with an interpolation", "location.href = `${base}/path`;\n"},
		{"a template with only interpolations", "location.href = `${a}${b}`;\n"},
		{"a no-substitution template", "location.assign(`/dashboard`);\n"},
		{"a concatenation whose left side is relative", "location.href = '/search?term=' + encodeURIComponent(term);\n"},
		{"a concatenation nested to the left", "location.href = '/a' + b + c;\n"},
		{"a concatenation whose left side is a template", "location.href = `/a` + x;\n"},
		{"an identifier initialized to a relative path", "const target = '/dashboard';\nlocation.assign(target);\n"},
		{"an identifier initialized to a template", "const target = `/dashboard`;\nlocation.assign(target);\n"},
		{"an identifier rewritten before the read", "let target = 'https://example.com';\ntarget = '/dashboard';\nlocation.assign(target);\n"},
		{"an identifier declared without an initializer then written", "let target;\ntarget = '/dashboard';\nlocation.assign(target);\n"},
		{"an identifier resolved through two hops", "const first = '/dashboard';\nconst second = first;\nlocation.assign(second);\n"},
		{"a violation nested inside a handler inside a component", "export function Search() {\n    const onSelect = () => {\n        window.location.href = '/support/search?term=' + encodeURIComponent(term);\n    };\n    return onSelect;\n}\n"},
		{"a violation inside a class method", "export class Nav {\n    go() {\n        window.location.href = '/home';\n    }\n}\n"},
		// A parenthesized assignment target. Our tree keeps the parenthesis as a real node where
		// the reference parser has already dropped it, so the skip at the target is load-bearing
		// here and inert there. Measured against the reference: both report.
		{"a parenthesized href target", "(location.href) = \"/dashboard\";\n"},
		{"a parenthesized href target through window", "(window.location.href) = \"/dashboard\";\n"},
		{"a parenthesized location receiver", "(location).assign('/dashboard');\n"},
		{"a parenthesized window receiver", "(window).location.href = '/dashboard';\n"},
		{"a parenthesized callee", "(location.assign)('/dashboard');\n"},
		{"a parenthesized argument", "location.assign(('/dashboard'));\n"},
		{"an optional call", "location.assign?.('/dashboard');\n"},
		{"an optional member access", "window.location?.assign('/dashboard');\n"},
		{"a use outside a block that shadowed the name", "{\n    const location = { assign() {} };\n    location.assign('/inner');\n}\nlocation.assign('/dashboard');\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NoLocationAssignRelativeDestination,
				locationAssignFile, testCase.sourceText), "noLocationAssign")
		})
	}
}

// TestNoLocationAssignRelativeDestinationStaysSilent covers every input measured not to report.
//
// The cases that matter most are the ones where the rule declines something a naive port would
// report: a value it cannot resolve, a receiver that is not the global, and the two navigation
// forms upstream simply does not handle.
func TestNoLocationAssignRelativeDestinationStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an absolute https destination", "location.assign('https://example.com');\n"},
		{"a protocol-relative destination", "location.assign('//example.com/path');\n"},
		{"an uppercase scheme", "location.assign('HTTP://example.com');\n"},
		{"a mailto scheme", "location.assign('mailto:a@example.com');\n"},
		{"a tel scheme", "location.assign('tel:5551234');\n"},
		{"a template whose first span is a scheme", "location.href = `https://${host}/x`;\n"},
		{"a concatenation whose left side is absolute", "location.href = 'https://example.com/' + path;\n"},
		{"an identifier initialized to an absolute url", "const target = 'https://example.com';\nlocation.assign(target);\n"},
		{"an identifier rewritten only after the read", "let target = 'https://example.com';\nlocation.assign(target);\ntarget = '/dashboard';\n"},
		{"an identifier declared without an initializer and written after the read", "let target;\nlocation.assign(target);\ntarget = '/dashboard';\n"},
		{"an identifier bound to a call", "const target = buildUrl();\nlocation.assign(target);\n"},
		{"a parameter", "export function go(target: string) {\n    location.assign(target);\n}\n"},
		{"a property access", "location.href = result.checkoutUrl;\n"},
		{"a call result passed directly", "location.assign(buildUrl({ id }));\n"},
		{"a location.replace call", "location.replace('/dashboard');\n"},
		{"a window.location.replace call", "window.location.replace('/dashboard');\n"},
		{"an href read", "export const current = location.href;\n"},
		{"an assign reference that is not called", "export const go = window.location.assign;\n"},
		{"a call with no arguments", "location.assign();\n"},
		{"a spread argument", "location.assign(...parts);\n"},
		{"a top receiver", "top.location.assign('/dashboard');\n"},
		{"a parent receiver", "parent.location.assign('/dashboard');\n"},
		{"a nested self under window", "window.self.location.href = '/dashboard';\n"},
		{"a location property of location", "location.location.href = '/dashboard';\n"},
		{"a module-scope binding named location", "const location = { assign(_: string) {} };\nlocation.assign('/dashboard');\n"},
		{"a parameter named location", "export function go(location: { assign(url: string): void }) {\n    location.assign('/dashboard');\n}\n"},
		{"a parameter named window", "export function go(window: { location: { href: string } }) {\n    window.location.href = '/dashboard';\n}\n"},
		// An import binds the name locally, so the receiver is not the ambient global. All four
		// import shapes were measured against the reference and all four are silent there. The
		// renamed one matters most: what binds is the local name, not the imported one.
		{"a named import of the name", "import { location } from \"./globals\";\nlocation.assign(\"/dashboard\");\n"},
		{"a default import of the name", "import location from \"./globals\";\nlocation.assign(\"/dashboard\");\n"},
		{"a namespace import of the name", "import * as window from \"./globals\";\nwindow.location.href = \"/dashboard\";\n"},
		{"an import renamed to the name", "import { anything as location } from \"./globals\";\nlocation.assign(\"/dashboard\");\n"},
		// These two must USE the shadowed name as a receiver, not merely bind it. A fixture that
		// declares the name and never navigates stays silent whatever the rule does, so it reads as
		// coverage while testing nothing; the mutation sweep found exactly that and both branches
		// survived until the use was added.
		{"a function declared with the name", "function location() {}\nlocation.assign(\"/dashboard\");\n"},
		{"a class declared with the name", "class window {}\nwindow.location.href = \"/dashboard\";\n"},
		// A catch parameter binds the name for the length of its block, which is a shape no other
		// fixture here reaches: the sweep found the catch arm surviving until these were added.
		{"a catch parameter named location", "try {\n    go();\n} catch (location) {\n    location.assign(\"/dashboard\");\n}\n"},
		{"a catch parameter named window", "try {\n    go();\n} catch (window) {\n    window.location.href = \"/dashboard\";\n}\n"},
		{"a shadow inside the block that uses it", "{\n    const location = { assign(_: string) {} };\n    location.assign('/dashboard');\n}\n"},
		{"a user object with a location property", "container.location.href = '/dashboard';\n"},
		{"a property named href on another object", "window.settings.href = '/dashboard';\n"},
		{"a differently cased method", "location.Assign('/dashboard');\n"},
		{"a differently cased receiver", "Location.assign('/dashboard');\n"},
		{"a different method on location", "location.navigate('/dashboard');\n"},
		{"a template literal used as the assign key", "location[`assign`]('/dashboard');\n"},
		{"a template literal used as the href key", "location[`href`] = '/dashboard';\n"},
		{"a non-literal computed key", "location[method]('/dashboard');\n"},
		{"a non-literal computed receiver key", "window[which].assign('/dashboard');\n"},
		{"a sequence expression callee", "(0, location.assign)('/dashboard');\n"},
		{"a conditional argument", "location.assign(condition ? '/a' : '/b');\n"},
		{"a logical argument", "location.assign(fallback || '/a');\n"},
		{"a destructuring assignment to href", "({ href: location.href } = target);\n"},
		{"a member of href", "location.href.length = 0;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoLocationAssignRelativeDestination,
				locationAssignFile, testCase.sourceText))
		})
	}
}

// TestNoLocationAssignRelativeDestinationPointsAtTheNavigation asserts where each finding lands.
//
// Message-id assertions cannot see this, and the choice is not obvious: the original reports the
// whole `CallExpression` for `assign` and the whole `AssignmentExpression` for `href`, so a finding
// covers the destination as well as the receiver. Reporting the callee instead would satisfy every
// fixture above while pointing somewhere else. The expected text below is written out rather than
// derived from the rule, so a change to what the rule reports cannot quietly move the expectation
// with it.
func TestNoLocationAssignRelativeDestinationPointsAtTheNavigation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantText   string
	}{
		{"a call", "location.assign('/dashboard');\n", "location.assign('/dashboard')"},
		{"a call through window", "window.location.assign('/x');\n", "window.location.assign('/x')"},
		{"an href assignment", "location.href = '/dashboard';\n", "location.href = '/dashboard'"},
		{
			"an href assignment through window",
			"window.location.href = '/x';\n",
			"window.location.href = '/x'",
		},
		// A compound assignment reports over the whole expression, operator included, rather than
		// over the target alone.
		{"a compound assignment", "location.href += '/x';\n", "location.href += '/x'"},
		// The parenthesized receiver is inside the reported range, which is what upstream's own
		// message text shows when it prints the receiver back with its parentheses intact.
		{"a parenthesized receiver", "(location).assign('/x');\n", "(location).assign('/x')"},
		// A concatenation is part of the navigation, so the finding covers the whole right side
		// rather than stopping at the literal that decided the verdict.
		{
			"a concatenation",
			"location.href = '/s?t=' + encodeURIComponent(t);\n",
			"location.href = '/s?t=' + encodeURIComponent(t)",
		},
		// Nested inside a function the finding must still cover only the navigation, not the
		// statement or the enclosing body.
		{
			"a nested violation",
			"export function go() {\n    window.location.href = '/x';\n}\n",
			"window.location.href = '/x'",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoLocationAssignRelativeDestination,
				locationAssignFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Fatalf("expected the finding to cover %q, got %q", testCase.wantText, reported)
			}
		})
	}
}

// TestNoLocationAssignRelativeDestinationRendersItsMessage asserts the rendered text exactly.
//
// Compared against a literal typed here rather than against the rule's own message constant. The
// constant would move with the rule under any edit, so an assertion against it can never fail and
// reads as a guard while guarding nothing.
func TestNoLocationAssignRelativeDestinationRendersItsMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, NoLocationAssignRelativeDestination,
		locationAssignFile, "location.assign('/dashboard');\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected one finding, got %d", len(result.Diagnostics))
	}
	const wantId = "noLocationAssign"
	if result.Diagnostics[0].Message.Id != wantId {
		t.Fatalf("expected id %q, got %q", wantId, result.Diagnostics[0].Message.Id)
	}
	const wantDescription = "Navigating to an internal page through `location.assign` or by writing " +
		"`location.href` tears down the running application and reloads it from the server, so " +
		"client state, the router's history entry and any in-flight data are all discarded. Use " +
		"`redirect()` in the render phase, or `useRouter().push()` in a Client Component's event " +
		"handler."
	if result.Diagnostics[0].Message.Description != wantDescription {
		t.Fatalf("expected description %q, got %q", wantDescription, result.Diagnostics[0].Message.Description)
	}
}
