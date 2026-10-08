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
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
	t.Logf("three 4096-operand cases, %d identical bytes in all builds", len(want))
}
func TestDeepMutants(t *testing.T) {
	t.Parallel()
	list := manifest(t, []string{"type T=(A); a+b+c; const t=tag`a${b}c`; a&&(b&&c);"})
	want := execute(t, "", goOracle(t), "--manifest", list)
	for _, item := range []struct{ name, file, from, to string }{
		{"binary-operator", "binaryConvert.ts", "node.set('operator', stringValue(operator));", "node.set('operator', stringValue('-'));"},
		{"postorder-alias", "postprocess.ts", "node.set(key, childValue(completed.get(value.node) ?? value.node));", "node.set(key, childValue(value.node));"},
		{"dump-property-order", "protocol.ts", "for(let index = node.properties.length - 1; index >= 0; index--)", "for(let index = 0; index < node.properties.length; index++)"},
	} {
		t.Run(item.name, func(t *testing.T) {
			t.Parallel()
			main := mutantPort(t, item.file, item.from, item.to)
			checkPort(t, main, []string{"--manifest", list}, want, true, false)
		})
	}
}
