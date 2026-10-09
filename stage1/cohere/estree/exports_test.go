package estree

import (
	"path/filepath"
	"testing"
)

func decoratedExports() []string {
	return []string{"@d export class C {}", "@d export default class C {}", "export @d class C {}", "export default @d class C {}", "@a @b() export abstract class C {}", "// first\n@d\nexport default class {}"}
}

const testDecoratedExportsShards = 6

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every decorated export.
func TestDecoratedExports(t *testing.T) {
	finishSetup := miscStart(t)
	cases := decoratedExports()
	ids := miscIDs("export", len(cases))
	if len(cases) != testDecoratedExportsShards {
		t.Fatal("decorated export enumeration changed")
	}
	oracle := miscOracle(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := miscBuild(t, main)
	finishSetup()
	miscRunShards(t, testDecoratedExportsShards, ids, func(t *testing.T, i int) {
		list := manifest(t, cases[i:i+1])
		want := execute(t, "", oracle, "--manifest", list)
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
			if err := miscCompare(want, got, false); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	})
}

func TestDecoratedExportsPlantedDisagreement(t *testing.T) {
	miscPlantedProof(t, testDecoratedExportsShards, miscIDs("export", len(decoratedExports())), func(planted bool) error {
		got := []byte("agree")
		if planted {
			got = []byte("disagree")
		}
		return miscCompare([]byte("agree"), got, false)
	})
}

const testDecoratedExportMutantShards = 6

// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every decorated export mutant case.
func TestDecoratedExportMutant(t *testing.T) {
	finishSetup := miscStart(t)
	cases := decoratedExports()
	if len(cases) != testDecoratedExportMutantShards {
		t.Fatal("decorated export mutant enumeration changed")
	}
	ids := miscIDs("export-mutant", len(cases))
	oracle := miscOracle(t)
	path := mutantPort(t, "convert.ts", "this.arena.node(wrapper).start = this.start(exported);", "this.arena.node(wrapper).start = this.start(id);")
	binary, _ := miscBuild(t, path)
	finishSetup()
	miscRunShards(t, testDecoratedExportMutantShards, ids, func(t *testing.T, i int) {
		list := manifest(t, cases[i:i+1])
		want := execute(t, "", oracle, "--manifest", list)
		// A decorator before export changes the wrapper start; in cases 2 and 3
		// export precedes the decorator, so the mutated start is already correct.
		mustDisagree := i != 2 && i != 3
		for name, got := range map[string][]byte{"Node": onNode(t, path, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if err := miscCompare(want, got, mustDisagree); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	})
}

func TestDecoratedExportMutantPlantedSurvivor(t *testing.T) {
	miscPlantedProof(t, testDecoratedExportMutantShards, miscIDs("export-mutant", len(decoratedExports())), func(planted bool) error {
		got := []byte("disagree")
		if planted {
			got = []byte("agree")
		}
		return miscCompare([]byte("agree"), got, true)
	})
}

func TestDecoratedExportLibraries(t *testing.T) { checkOriginalLibraries(t, decoratedExports(), 0) }
