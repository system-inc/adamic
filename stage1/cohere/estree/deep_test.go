package estree

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func deepGrammar() []string {
	var numeric, quoted, logical strings.Builder
	numeric.WriteString("type T=(A);\n")
	logical.WriteString("(")
	for index := 0; index < 4096; index++ {
		if index > 0 {
			numeric.WriteString("+")
			quoted.WriteString("+")
			logical.WriteString("&&")
		}
		fmt.Fprint(&numeric, index)
		quoted.WriteString("'x'")
		logical.WriteString("x")
	}
	numeric.WriteString(";")
	quoted.WriteString(";")
	logical.WriteString(");")
	return []string{numeric.String(), quoted.String(), logical.String()}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestDeepGrammar(t *testing.T) {
	list := manifest(t, deepGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
	t.Logf("three 4096-operand cases, %d identical bytes in all builds", len(want))
}
