package estree

import "testing"

func TestAcceptanceMutants(t *testing.T) {
	list := manifest(t, acceptanceGrammar())
	want := execute(t, "", goOracle(t), "--manifest", list)
	for _, item := range []struct{ name, file, from, to string }{
		{"catch-initializer", "convert.ts", "this.separated(this.child(declaration, 0), this.child(declaration, 1), 'ColonToken')", "true"},
		{"class-keyword-name", "sourceStatements.ts", "!(this.parser.peek() === 'Identifier' || this.parser.peek().endsWith('Keyword'))", "false"},
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
