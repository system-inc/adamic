package estree

import (
	"path/filepath"
	"testing"
)

func decoratedExports() []string {
	return []string{"@d export class C {}", "@d export default class C {}", "export @d class C {}", "export default @d class C {}", "@a @b() export abstract class C {}", "// first\n@d\nexport default class {}"}
}
func TestDecoratedExports(t *testing.T) {
	t.Parallel()
	list := manifest(t, decoratedExports())
	want := execute(t, "", goOracle(t), "--manifest", list)
	main, _ := filepath.Abs("main.ts")
	checkPort(t, main, []string{"--manifest", list}, want, false, false)
	t.Logf("%d decorated exports, %d bytes match Go", len(decoratedExports()), len(want))
}
func TestDecoratedExportMutant(t *testing.T) {
	t.Parallel()
	list := manifest(t, decoratedExports())
	want := execute(t, "", goOracle(t), "--manifest", list)
	path := mutantPort(t, "convert.ts", "this.arena.node(wrapper).start = this.start(exported);", "this.arena.node(wrapper).start = this.start(id);")
	checkPort(t, path, []string{"--manifest", list}, want, true, false)
}

func TestDecoratedExportLibraries(t *testing.T) { checkOriginalLibraries(t, decoratedExports(), 0) }
