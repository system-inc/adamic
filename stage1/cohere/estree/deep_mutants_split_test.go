package estree

import "testing"

func TestDeepMutants(t *testing.T) {
	list := manifest(t, []string{"type T=(A); a+b+c; const t=tag`a${b}c`; a&&(b&&c);"})
	want := execute(t, "", goOracle(t), "--manifest", list)
	for _, item := range []struct{ name, file, from, to string }{
		{"binary-operator", "binaryConvert.ts", "node.set('operator', stringValue(operator));", "node.set('operator', stringValue('-'));"},
		{"postorder-alias", "postprocess.ts", "node.set(key, childValue(completed.get(value.node) ?? value.node));", "node.set(key, childValue(value.node));"},
		{"dump-property-order", "protocol.ts", "for(let index = node.properties.length - 1; index >= 0; index--)", "for(let index = 0; index < node.properties.length; index++)"},
	} {
		t.Run(item.name, func(t *testing.T) {
			main := mutantPort(t, item.file, item.from, item.to)
			binary, _ := build(t, main, true)
			for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
				if diff := firstDifference(want, got); diff == "" {
					t.Fatal(name + " mutant survived")
				} else {
					t.Log(name + ": " + diff)
				}
			}
		})
	}
}
