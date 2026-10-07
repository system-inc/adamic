package estree

import (
	"path/filepath"
	"testing"
)

func decoratedExports() []string {
	return []string{"@d export class C {}", "@d export default class C {}", "export @d class C {}", "export default @d class C {}", "@a @b() export abstract class C {}", "// first\n@d\nexport default class {}"}
}
func TestDecoratedExports(t *testing.T) {
	list := manifest(t, decoratedExports())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
		if d := firstDifference(want, got); d != "" {
			t.Fatal(name + ": " + d)
		}
	}
	t.Logf("%d decorated exports, %d bytes match Go", len(decoratedExports()), len(want))
}
func TestDecoratedExportMutant(t *testing.T) {
	list := manifest(t, decoratedExports())
	want := execute(t, "", goOracle(t), "--manifest", list)
	path := mutantPort(t, "convert.ts", "this.arena.node(wrapper).start = this.start(exported);", "this.arena.node(wrapper).start = this.start(id);")
	binary, _ := build(t, path, true)
	for name, got := range map[string][]byte{"Node": onNode(t, path, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
		if d := firstDifference(want, got); d == "" {
			t.Fatal(name + " mutant survived")
		} else {
			t.Log(name + ": " + d)
		}
	}
}

func TestDecoratedExportLibraries(t *testing.T) { checkOriginalLibraries(t, decoratedExports(), 0) }
