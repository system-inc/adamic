package type_checking

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/core"
)

// TestIsStrictCompilerOptionEnabledFollowsTheCompiler pins the resolution. The rule harness writes its
// own tsconfig with `strict` set, so no fixture can leave it unset, and unset was the case that was
// wrong: Structure's tsconfig never writes `strict`, TypeScript 6 reads that as on, and this helper read
// it as off (#6ar414z).
func TestIsStrictCompilerOptionEnabledFollowsTheCompiler(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		strict core.Tristate
		option core.Tristate
		want   bool
	}{
		{"neither set, which TypeScript 6 resolves to on", core.TSUnknown, core.TSUnknown, true},
		{"strict on and the option unset", core.TSTrue, core.TSUnknown, true},
		{"strict off and the option unset", core.TSFalse, core.TSUnknown, false},
		{"an explicit true wins over strict off", core.TSFalse, core.TSTrue, true},
		{"an explicit false wins over strict on", core.TSTrue, core.TSFalse, false},
		{"an explicit false wins with strict unset", core.TSUnknown, core.TSFalse, false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := &core.CompilerOptions{Strict: testCase.strict}
			if got := IsStrictCompilerOptionEnabled(options, testCase.option); got != testCase.want {
				t.Errorf("resolved to %v, want %v", got, testCase.want)
			}
		})
	}
}
