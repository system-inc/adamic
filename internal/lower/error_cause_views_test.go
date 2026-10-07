package lower

import (
	"strings"
	"testing"
)

func TestErrorCauseClassificationsCompile(t *testing.T) {
	for _, cause := range []string{"() => 1", "new Map<string, number>()", "new Set<number>()", "null", "{value: 1}"} {
		t.Run(cause, func(t *testing.T) {
			source := `const cause = ` + cause + `; const error = new Error('x', {cause}); let held = error.cause; console.log(typeof held);`
			if cause == "null" {
				source = `const error = new Error('x', {cause: null}); let held = error.cause; console.log(typeof held);`
			}
			_, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOpaqueErrorCauseReflectionIsNotYet(t *testing.T) {
	for _, cause := range []string{"() => 1", "new Map<string, number>()", "{callback: () => 1}", "{value: null}"} {
		t.Run(cause, func(t *testing.T) {
			_, err := lowerSource(t, `function inspect(value: unknown): boolean { return typeof value === 'object' && value !== null && 'size' in value; } const error = new Error('x', {cause: `+cause+`}); console.log(String(inspect(error.cause)));`)
			if err == nil || !strings.Contains(err.Error(), "opaque Error cause") {
				t.Fatalf("want opaque Error cause refusal, got %v", err)
			}
		})
	}
}

func TestConcreteErrorCauseReflectionCompiles(t *testing.T) {
	_, err := lowerSource(t, `const error = new Error('x', {cause: {value: 1}}); const cause = error.cause; console.log(String(typeof cause === 'object' && cause !== null && 'value' in cause));`)
	if err != nil {
		t.Fatal(err)
	}
}
