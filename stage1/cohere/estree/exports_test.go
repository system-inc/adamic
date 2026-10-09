package estree

import (
	"path/filepath"
	"testing"
)

func decoratedExports() []string {
	return []string{"@d export class C {}", "@d export default class C {}", "export @d class C {}", "export default @d class C {}", "@a @b() export abstract class C {}", "// first\n@d\nexport default class {}"}
}

const testDecoratedExportsShards = 32

// Shard counts stay fixed as cases grow; keys are this file, mode and source-case identity.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every decorated export.
// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestDecoratedExports(t *testing.T) {
	finishSetup := miscStart(t)
	cases := decoratedExports()
	ids := miscIDs("stage1/cohere/estree/exports_test.go:export", cases)
	oracle := miscOracle(t)
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary, script := miscBuild(t, main)
	answers := miscTextAnswers(t, oracle, cases, ".ts")
	finishSetup()
	miscRunShards(t, testDecoratedExportsShards, ids, func(t *testing.T, i int) {
		list := manifest(t, cases[i:i+1])
		want := answers[i].Data
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": onNode(t, script, "--manifest", list)} {
			if err := miscCompare(want, got, false); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	})
}

func TestDecoratedExportsPlantedDisagreement(t *testing.T) {
	t.Parallel()
	miscPlantedProof(t, testDecoratedExportsShards, miscIDs("stage1/cohere/estree/exports_test.go:export", decoratedExports()), func(planted bool) error {
		got := []byte("agree")
		if planted {
			got = []byte("disagree")
		}
		return miscCompare([]byte("agree"), got, false)
	})
}

const testDecoratedExportMutantShards = 32

// Shard counts stay fixed as cases grow; keys are this file, mode and source-case identity.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every decorated export mutant case.
// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestDecoratedExportMutant(t *testing.T) {
	finishSetup := miscStart(t)
	cases := decoratedExports()
	ids := miscIDs("stage1/cohere/estree/exports_test.go:export-mutant", cases)
	oracle := miscOracle(t)
	path := miscMutant(t, "convert.ts", "this.arena.node(wrapper).start = this.start(exported);", "this.arena.node(wrapper).start = this.start(id);")
	binary, script := miscBuild(t, path)
	answers := miscTextAnswers(t, oracle, cases, ".ts")
	finishSetup()
	miscRunShards(t, testDecoratedExportMutantShards, ids, func(t *testing.T, i int) {
		list := manifest(t, cases[i:i+1])
		want := answers[i].Data
		// A decorator before export changes the wrapper start; in cases 2 and 3
		// export precedes the decorator, so the mutated start is already correct.
		mustDisagree := i != 2 && i != 3
		for name, got := range map[string][]byte{"Node": onNode(t, path, "--manifest", list), "native": execute(t, "", binary, "--manifest", list), "emitted": mutantEmittedOutput(t, path, script, "--manifest", list)} {
			if err := miscCompare(want, got, mustDisagree); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	})
}

func TestDecoratedExportMutantPlantedSurvivor(t *testing.T) {
	t.Parallel()
	miscPlantedProof(t, testDecoratedExportMutantShards, miscIDs("stage1/cohere/estree/exports_test.go:export-mutant", decoratedExports()), func(planted bool) error {
		got := []byte("disagree")
		if planted {
			got = []byte("agree")
		}
		return miscCompare([]byte("agree"), got, true)
	})
}

func TestDecoratedExportLibraries(t *testing.T) {
	t.Parallel()
	checkOriginalLibraries(t, decoratedExports(), 0)
}
