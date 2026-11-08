package estree

import (
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"path/filepath"
	"testing"
)

func decoratedExports() []string { return oracles.DecoratedExports() }

const testDecoratedExportsShards = 32

// Shard counts stay fixed as cases grow; keys are this file, mode and source-case identity.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every decorated export.
// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestDecoratedExports(test *testing.T) {
	finishSetup := miscStart(test)
	cases := decoratedExports()
	ids := miscIds("stage1/cohere/estree/exports_test.go:export", cases)
	oracle := miscOracle(test)
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary, script := miscBuild(test, main)
	answers := miscTextAnswers(test, oracle, cases, ".ts")
	finishSetup()
	miscRunShards(test, testDecoratedExportsShards, ids, func(test *testing.T, index int) {
		list := manifest(test, cases[index:index+1])
		want := answers[index].Data
		for name, got := range map[string][]byte{
			"Node":    onNode(test, main, "--manifest", list),
			"native":  execute(test, "", binary, "--manifest", list),
			"emitted": onNode(test, script, "--manifest", list),
		} {
			if miscCompareError := miscCompare(want, got, false); miscCompareError != nil {
				test.Fatalf("%s %s: %v", ids[index], name, miscCompareError)
			}
		}
	})
}

func TestDecoratedExportsPlantedDisagreement(test *testing.T) {
	test.Parallel()
	miscPlantedProof(test, testDecoratedExportsShards, miscIds(
		"stage1/cohere/estree/exports_test.go:export",
		decoratedExports(),
	), func(planted bool) error {
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
func TestDecoratedExportMutant(test *testing.T) {
	finishSetup := miscStart(test)
	cases := decoratedExports()
	ids := miscIds("stage1/cohere/estree/exports_test.go:export-mutant", cases)
	oracle := miscOracle(test)
	path := estreeExportMutantPath(test)
	binary, _ := miscBuild(test, path)
	answers := miscTextAnswers(test, oracle, cases, ".ts")
	finishSetup()
	miscRunShards(test, testDecoratedExportMutantShards, ids, func(test *testing.T, index int) {
		list := manifest(test, cases[index:index+1])
		want := answers[index].Data
		// A decorator before export changes the wrapper start; in cases 2 and 3
		// export precedes the decorator, so the mutated start is already correct.
		mustDisagree := index != 2 && index != 3
		for name, got := range map[string][]byte{
			"Node":   onNode(test, path, "--manifest", list),
			"native": execute(test, "", binary, "--manifest", list),
		} {
			if miscCompareError := miscCompare(want, got, mustDisagree); miscCompareError != nil {
				test.Fatalf("%s %s: %v", ids[index], name, miscCompareError)
			}
		}
	})
}

func TestDecoratedExportMutantPlantedSurvivor(test *testing.T) {
	test.Parallel()
	miscPlantedProof(test, testDecoratedExportMutantShards, miscIds(
		"stage1/cohere/estree/exports_test.go:export-mutant",
		decoratedExports(),
	), func(planted bool) error {
		got := []byte("disagree")
		if planted {
			got = []byte("agree")
		}
		return miscCompare([]byte("agree"), got, true)
	})
}

func TestDecoratedExportLibraries(test *testing.T) {
	test.Parallel()
	checkOriginalLibraries(test, decoratedExports(), 0)
}
