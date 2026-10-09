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
func TestDeepGrammar(t *testing.T) {
	t.Parallel()
	list := manifest(t, deepGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if diff := firstDifference(want, got); diff != "" {
			t.Fatal(name + ": " + diff)
		}
	}
	t.Logf("three 4096-operand cases, %d identical bytes in all builds", len(want))
}
