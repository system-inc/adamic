package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"

	"github.com/system-inc/adamic/internal/native/runtimeproducts"
)

func init() {
	runtimeWitnesses["v4-receiver-regexp"] = func(t *testing.T) *ir.Program {
		program, _ := v4IdentityProgram(t, v4ReceiverRegExpSource, "ok-ok\n")
		return program
	}
}

func TestProduct_V4SanitizedCountedRuntimes0(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized-counted", 0)
}

func TestProduct_V4SanitizedCountedRuntimes1(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized-counted", 1)
}

func TestProduct_V4SanitizedCountedRuntimes2(t *testing.T) {
	t.Parallel()
	runtimeproducts.Build(t, "sanitized-counted", 2)
}
