package estree

import (
 "strings"
 "testing"
)

// Not parallel: native.Build writes the shared user cache directory adamic/runtime
func TestSyntaxMutants(t *testing.T) {
	t.Run("mapped-constraint", func(t *testing.T) {
		list := manifest(t, syntaxGrammar())
		want := execute(t, "", goOracle(t), "--manifest", list)
		main := mutantPort(t, "convert.ts", "this.set(result, 'constraint', this.converted(this.child(parameter, 1)));", "this.set(result, 'constraint', absent());")
		binary, _ := build(t, main, true)
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if diff := firstDifference(want, got); diff == "" {
				t.Fatal(name + " mutant survived")
			} else {
				t.Log(name + ": " + diff)
			}
		}
	})
	for _, item := range []struct{ name, file, from, to, source string }{
		{"erasure-precedence", "sourceBinary.ts", "if(nextRank > lastRank ||", "if(false && nextRank > lastRank ||", "1+1 as number *2;"},
		{"reference-pragma", "pipeline.ts", "if(reference !== '')", "if(false)", "/// <reference path='missingquote.ts />\nx;"},
	} {
		t.Run(item.name, func(t *testing.T) {
			list := manifest(t, []string{item.source})
			statuses := string(execute(t, "", goOracle(t), "--audit", list, t.TempDir()))
			if !strings.Contains(statuses, `"status":"error"`) {
				t.Fatal(statuses)
			}
			main := mutantPort(t, item.file, item.from, item.to)
			binary, _ := build(t, main, true)
			for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
				if !strings.Contains(string(got), "0 Program ") {
					t.Fatal(name + " control did not accept")
				}
				t.Log(name + ": disabled check accepts Go-refused input; acceptance oracle catches it")
			}
		})
	}
}

