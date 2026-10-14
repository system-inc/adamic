package oracle

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// V4 checks adapter ownership and allocation counts in the same sanitized run.
// Declare that combination so Loom builds it before the consuming tests.
func init() {
	runtimeSets = append(runtimeSets, runtimeSet{[]string{"ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK"}, "v4-receiver-regexp"})
	runtimeWitnesses["v4-receiver-regexp"] = func(t *testing.T) *ir.Program {
		program, _ := v4IdentityProgram(t, v4ReceiverRegExpSource, "ok-ok\n")
		return program
	}
	runtimeOptions = append(runtimeOptions, runtimeOption{"sanitized-counted", native.Options{Sanitize: true, Count: true}, ""})
}

func TestProduct_V4SanitizedCountedRuntimes0(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized-counted", 0)
}

func TestProduct_V4SanitizedCountedRuntimes1(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized-counted", 1)
}

func TestProduct_V4SanitizedCountedRuntimes2(t *testing.T) {
	t.Parallel()
	runtimeProducts(t, "sanitized-counted", 2)
}
