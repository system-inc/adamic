package property_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// firstKey drives a source through the parser and hands the first object-literal property's key
// node to `ask`, so every assertion below runs against a real parsed node rather than a synthesized
// one.
func firstKey(t *testing.T, source string, ask func(key *ast.Node)) {
	t.Helper()
	found := false
	rule_testing.Run(t, rule.Rule{
		Name: "key-probe",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindObjectLiteralExpression: func(node *ast.Node) {
					if found {
						return
					}
					properties := node.AsObjectLiteralExpression().Properties
					if properties == nil || len(properties.Nodes) == 0 {
						return
					}
					if key := properties.Nodes[0].Name(); key != nil {
						found = true
						ask(key)
					}
				},
			}
		},
	}, "k.ts", "declare const variable: string;\n"+source)
	if !found {
		t.Fatalf("%q: no object-literal key reached the probe", source)
	}
}

func expectName(t *testing.T, source string, accept property.Kinds, want string, wantOk bool) {
	t.Helper()
	firstKey(t, source, func(key *ast.Node) {
		got, ok := property.Name(key, accept)
		if ok != wantOk || got != want {
			t.Errorf("%q: Name = (%q, %v), want (%q, %v)", source, got, ok, want, wantOk)
		}
	})
}

// TestStaticAcceptsEverySettledSpelling covers the union set.
func TestStaticAcceptsEverySettledSpelling(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ source, want string }{
		{"({ plain: 1 });", "plain"},
		{"({ 'quoted': 1 });", "quoted"},
		{`({ "double": 1 });`, "double"},
		{"({ 42: 1 });", "42"},
		// The parser normalizes numeric keys, so two spellings of one number answer alike without
		// any work here.
		{"({ 1e1: 1 });", "10"},
		{"({ 0x10: 1 });", "16"},
		{"({ ['computedString']: 1 });", "computedString"},
		{"({ [42]: 1 });", "42"},
		{"({ [`tpl`]: 1 });", "tpl"},
		{"({ shorthand });", "shorthand"},
		{"({ method() {} });", "method"},
		{"({ get accessor() { return 1; } });", "accessor"},
	} {
		expectName(t, c.source, property.Static, c.want, true)
	}
}

// TestVariableInBracketsIsAlwaysDeclined is the one judgment this package makes rather than
// delegates to its flags.
//
// `{ [a]: 1 }` names whichever property the variable holds, which is not knowable before it runs.
// Reading the variable's spelling as the key reported `[foo]()` and `foo()` as duplicate class
// members, caught by upstream's own clean case rather than by inspection.
func TestVariableInBracketsIsAlwaysDeclined(t *testing.T) {
	t.Parallel()
	expectName(t, "({ [variable]: 1 });", property.Static, "", false)
}

// TestTextualDeclinesWhatItDoesNotAccept pins the narrow set three callers use. Each of these is a
// spelling those callers never meet, and accepting it would widen their rules silently.
func TestTextualDeclinesWhatItDoesNotAccept(t *testing.T) {
	t.Parallel()
	expectName(t, "({ plain: 1 });", property.Textual, "plain", true)
	expectName(t, "({ 'quoted': 1 });", property.Textual, "quoted", true)
	expectName(t, "({ 42: 1 });", property.Textual, "", false)
	expectName(t, "({ ['computedString']: 1 });", property.Textual, "", false)
	expectName(t, "({ [`tpl`]: 1 });", property.Textual, "", false)
}

// TestComputedPropertyNameDoesNotPanic is the guard that is load-bearing rather than defensive.
//
// Node.Text() panics on a KindComputedPropertyName rather than returning empty, so a caller that
// read the text first and filtered afterwards crashes the linter. Three of the six lifted
// implementations had no guard, because their callers never met the shape.
func TestComputedPropertyNameDoesNotPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Name panicked on a computed key: %v", recovered)
		}
	}()
	// Every accept set, including one that excludes Computed, so the early return is exercised too.
	for _, accept := range []property.Kinds{property.Static, property.Textual, property.Named} {
		firstKey(t, "({ [variable]: 1 });", func(key *ast.Node) {
			property.Name(key, accept)
			property.NameTagged(key, accept)
		})
	}
}

// TestNameTaggedSeparatesNumbersFromStrings is what one caller needs and the rest must not have.
//
// `[1.0]` and `['1.0']` are different class members while `10` and `1e1` are the same one, so
// neither the source spelling nor the cooked text answers alone.
func TestNameTaggedSeparatesNumbersFromStrings(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ source, want string }{
		{"({ plain: 1 });", "string:plain"},
		{"({ 'quoted': 1 });", "string:quoted"},
		{"({ 42: 1 });", "number:42"},
		{"({ 1.0: 1 });", "number:1"},
		{"({ '1.0': 1 });", "string:1.0"},
		{"({ 1e1: 1 });", "number:10"},
		{"({ [`tpl`]: 1 });", "string:tpl"},
	} {
		firstKey(t, c.source, func(key *ast.Node) {
			got, ok := property.NameTagged(key, property.Static)
			if !ok || got != c.want {
				t.Errorf("%q: NameTagged = (%q, %v), want (%q, true)", c.source, got, ok, c.want)
			}
		})
	}
	// The number and string spellings of one text must not collide.
	var numeric, quoted string
	firstKey(t, "({ 1.0: 1 });", func(k *ast.Node) { numeric, _ = property.NameTagged(k, property.Static) })
	firstKey(t, "({ '1.0': 1 });", func(k *ast.Node) { quoted, _ = property.NameTagged(k, property.Static) })
	if numeric == quoted {
		t.Errorf("a numeric key and a string key with the same text tagged alike: %q", numeric)
	}
}

// TestNilIsNotAName covers the guard a shared function needs and a rule-local one did not: inside a
// rule the node always came from a walk, and on a shelf any caller can pass anything.
func TestNilIsNotAName(t *testing.T) {
	t.Parallel()
	if _, ok := property.Name(nil, property.Static); ok {
		t.Error("Name(nil) reported a name")
	}
	if _, ok := property.NameTagged(nil, property.Static); ok {
		t.Error("NameTagged(nil) reported a name")
	}
	if _, ok := property.AccessedName(nil, property.Static); ok {
		t.Error("AccessedName(nil) reported a name")
	}
}

// TestAccessedName covers the member-access half, where the two spellings of one property have to
// compare as one reference and a variable subscript must not.
func TestAccessedName(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		source string
		want   string
		wantOk bool
	}{
		{"a.b;", "b", true},
		{"a['b'];", "b", true},
		{"a[`b`];", "b", true},
		{"a[0];", "0", true},
		// A variable subscript names whatever it holds, so two reads are the same reference only if
		// the variable has not changed, and nothing here can know that.
		{"a[variable];", "", false},
		{"a();", "", false},
	} {
		source := c.source
		want, wantOk := c.want, c.wantOk
		found := false
		rule_testing.Run(t, rule.Rule{
			Name: "access-probe",
			Run: func(ctx rule.Context, options any) rule.Listeners {
				check := func(node *ast.Node) {
					if found {
						return
					}
					found = true
					got, ok := property.AccessedName(node, property.Static)
					if ok != wantOk || got != want {
						t.Errorf("%q: AccessedName = (%q, %v), want (%q, %v)",
							source, got, ok, want, wantOk)
					}
				}
				return rule.Listeners{
					ast.KindPropertyAccessExpression: check,
					ast.KindElementAccessExpression:  check,
				}
			},
		}, "a.ts", "declare const a: any;\ndeclare const variable: string;\n"+source)
	}
}
