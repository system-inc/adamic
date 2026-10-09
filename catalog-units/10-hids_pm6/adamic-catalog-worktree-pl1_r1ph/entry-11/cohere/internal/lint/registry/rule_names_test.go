package registry

import (
	"strings"
	"testing"
)

// registeredRuleNames is every name in the live catalog.
func registeredRuleNames() []string {
	rules := All()
	names := make([]string, 0, len(rules))
	for _, registered := range rules {
		names = append(names, registered.Name)
	}
	return names
}

// TestRegisteredNamesMatchTheirUpstreamSpelling refuses a namespace this tree does not recognize.
//
// Rules used to register bare, and `bareRuleName` stripped any prefix before matching, so a config
// key spelled `bogusplugin/no-alert` resolved exactly like the correct one -- measured, same six
// findings, no warning. ESLint is not that forgiving: an unknown rule key is fatal there, so a
// wrong prefix killed the whole run twice in one day while `s l --linter both` kept printing a
// comparison against a linter that had linted nothing.
//
// Registering the real upstream name removes the translation step. One name is right everywhere,
// and a wrong prefix now fails here rather than in ESLint.
//
// A plugin rule must carry its plugin's namespace; an ESLint core rule must carry none, because
// upstream has none. Today that is 272 namespaced against 146 bare, and both halves are correct:
// the shape of a name is decided by what upstream calls it, never by a house preference for one
// form. Anything else is the inconsistency this guard exists to catch.
//
// This test once lived beside the parity suite and was argued in its terms, because a namespaced
// registration was invisible to the parity count. That catalog is gone, and the property is not:
// `cohere --rules` prints these names and `configuration.settingFor` splits them on the `/`
// boundary, so a name that lies about its plugin still reaches a user and still fails a config
// lookup. The guard outlived the instrument that motivated it.
func TestRegisteredNamesMatchTheirUpstreamSpelling(t *testing.T) {
	t.Parallel()
	registered := registeredRuleNames()
	if len(registered) == 0 {
		t.Fatal("registry.All() returned no rules, so this test checked nothing")
	}

	// The namespaces a rule may carry. A rule with no slash is an ESLint core rule or one of this
	// tree's own, both of which are correct bare.
	knownNamespaces := map[string]bool{
		"@typescript-eslint": true,
		// Adamic's soundness rules (#drbrp8c), named by Adamic's own list in policy/RuleNaming.json, so
		// one rule has one name in both projects.
		"adamic": true,
		// `base` is api-phi-health's own lint layer, the same kind of namespace as structure and
		// nexus below: rules this organization wrote rather than ported from a plugin. It is not an
		// ESLint plugin prefix, so nothing translates it on the way out.
		"base": true,
		// eslint-plugin-boundaries, which Base's config loads to keep its layers pointing one way.
		"boundaries":         true,
		"react":              true,
		"react-hooks":        true,
		"@next/next":         true,
		"better-tailwindcss": true,
		"structure":          true,
		"nexus":              true,
		// @eslint-community/eslint-plugin-eslint-comments, whose rules are about directive comments.
		// Its plugin key is the two-segment `@eslint-community/eslint-comments`, which is the prefix
		// its documentation configures and the one ESLint would resolve if the plugin were installed.
		"@eslint-community/eslint-comments": true,
	}

	for _, name := range registered {
		slash := strings.LastIndex(name, "/")
		if slash < 0 {
			continue
		}
		namespace := name[:slash]
		if knownNamespaces[namespace] {
			continue
		}
		t.Errorf("rule %q carries the namespace %q, which is not one this tree recognizes; a rule "+
			"registers under its real upstream name, and an unrecognized prefix is the spelling "+
			"that reaches ESLint and kills the run", name, namespace)
	}
}
