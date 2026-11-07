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

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestDeepGrammar(test *testing.T) {
	list := manifest(test, deepGrammar())
	want := execute(test, "", goOracle(test), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(test, main, true)
	for name, got := range map[string][]byte{
		"Node":    onNode(test, main, "--manifest", list),
		"native":  execute(test, "", binary, "--manifest", list),
		"emitted": onNode(test, script, "--manifest", list),
	} {
		if diff := firstDifference(want, got); diff != "" {
			test.Fatal(name + ": " + diff)
		}
	}
	test.Logf("three 4096-operand cases, %d identical bytes in all builds", len(want))
}
