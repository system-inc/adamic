package react

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// compilerRuleNames are the fifteen React Compiler rules that share upstream's one options schema.
var compilerRuleNames = []string{
	"react-hooks/config",
	"react-hooks/error-boundaries",
	"react-hooks/globals",
	"react-hooks/immutability",
	"react-hooks/incompatible-library",
	"react-hooks/no-deriving-state-in-effects",
	"react-hooks/preserve-manual-memoization",
	"react-hooks/purity",
	"react-hooks/refs",
	"react-hooks/set-state-in-effect",
	"react-hooks/set-state-in-render",
	"react-hooks/static-components",
	"react-hooks/unsupported-syntax",
	"react-hooks/use-memo",
	"react-hooks/void-use-memo",
}

// TestCompilerRulesAcceptOnlyTheEmptyObject pins ruling 2 on #e06zm4b for every one of the fifteen,
// through the decoder each one registers rather than the shared function, so a rule left
// unregistered or registered with another decoder fails by name.
func TestCompilerRulesAcceptOnlyTheEmptyObject(t *testing.T) {
	t.Parallel()

	decoders := map[string]func([]byte) (any, error){}
	for _, registration := range rule.Registered() {
		decoders[registration.Rule.Name] = registration.Decode
	}

	for _, name := range compilerRuleNames {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			decode := decoders[name]
			if decode == nil {
				t.Fatalf("%s registers no decoder, so the config refuses the {} ESLint writes", name)
			}
			for _, raw := range []string{``, `{}`, ` { } `} {
				if _, err := decode([]byte(raw)); err != nil {
					t.Errorf("%q was refused: %v", raw, err)
				}
			}
			_, err := decode([]byte(`{"environment": {"validateRefAccessDuringRender": false}}`))
			if err == nil || !strings.Contains(err.Error(), "environment") ||
				!strings.Contains(err.Error(), "React Compiler configuration") {
				t.Errorf("compiler configuration was not refused with its reason: %v", err)
			}
			for _, raw := range []string{`null`, `[]`, `"strict"`, `true`} {
				if _, err := decode([]byte(raw)); err == nil {
					t.Errorf("%s decoded; upstream's schema takes an object", raw)
				}
			}
		})
	}
}
