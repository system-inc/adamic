package oracles

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

func boundedAnswers(test *testing.T) {
	inputs := []ComparisonInput{}
	for _, item := range BoundedCases(test) {
		if item.Audit {
			inputs = append(inputs, item)
		}
	}
	CachedComparisonAnswers(test, ParserOracle(test), inputs)
}
func textAnswers(test *testing.T, oracle string, texts []string, extension string) {
	inputs := make([]ComparisonInput, len(texts))
	for i, text := range texts {
		inputs[i] = ComparisonInput{ID: fmt.Sprint(i), Text: text, Filename: "0000" + extension}
	}
	CachedComparisonAnswers(test, oracle, inputs)
}
func libraryGapsAnswers(test *testing.T) {
	for _, source := range RecoveryLibraryGaps() {
		CachedRecoveryAnswer(test, ParserOracle(test), manifest(test, []string{source}), "--manifest")
	}
}

func TestProduct_EstreeGoOracle(test *testing.T) { test.Parallel(); ParserOracle(test) }

func TestProduct_AcceptanceOracle(test *testing.T) { test.Parallel(); AcceptanceMutantsOracle(test) }

func TestProduct_scalar_edges_go_oracle(test *testing.T) {
	test.Parallel()
	binary := ScalarEdgesOracle(test)
	// The product row proves the oracle exists and runs: it answers an empty manifest (#n1g5604).
	manifest := filepath.Join(test.TempDir(), "empty-manifest")
	if err := os.WriteFile(manifest, nil, 0600); err != nil {
		test.Fatal(err)
	}
	buildcache.RequireArtifacts(
		test,
		filepath.Dir(binary),
		buildcache.Artifact{Name: filepath.Base(binary), Executable: true, Arguments: []string{"--manifest", manifest}},
	)
}

func TestProduct_SyntaxMutantsGoOracle(test *testing.T) { test.Parallel(); SyntaxMutantsOracle(test) }

func TestProduct_ThreePortOracle(test *testing.T) { test.Parallel(); ThreePortMutantsOracle(test) }

func TestProduct_DeepMutantsFixture(test *testing.T) { test.Parallel(); DeepMutantsFixture(test) }

func TestProduct_EstreeBoundedAnswers(test *testing.T) { test.Parallel(); boundedAnswers(test) }

func TestProduct_EstreeExportAnswers(test *testing.T) {
	test.Parallel()
	textAnswers(test, ParserOracle(test), DecoratedExports(), ".ts")
}

func TestProduct_JSXMutantAnswers(test *testing.T) {
	test.Parallel()
	textAnswers(test, ParserOracle(test), JSXCases(), ".tsx")
}

func TestProduct_RecoveredExpressionsAnswer(test *testing.T) {
	test.Parallel()
	CachedRecoveryAnswer(test, ParserOracle(test), manifest(test, RecoveredExpressions()), "--manifest")
}

func TestProduct_RecoveredGrammarAnswer(test *testing.T) {
	test.Parallel()
	CachedRecoveryAnswer(test, ParserOracle(test), manifest(test, RecoveredGrammar()), "--manifest")
}

func TestProduct_RecoveryLibraryGapsAnswers(test *testing.T) {
	test.Parallel()
	libraryGapsAnswers(test)
}

func TestProduct_RecoveryMutantsAnswer(test *testing.T) {
	test.Parallel()
	CachedRecoveryAnswer(test, ScalarEdgesOracle(test), manifest(test, RecoveredGrammar()), "--manifest")
}
