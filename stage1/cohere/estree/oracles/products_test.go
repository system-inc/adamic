package oracles

import (
	"fmt"
	"testing"
)

func boundedAnswers(t *testing.T) {
	inputs := []ComparisonInput{}
	for _, item := range BoundedCases(t) {
		if item.Audit {
			inputs = append(inputs, item)
		}
	}
	CachedComparisonAnswers(t, ParserOracle(t), inputs)
}
func textAnswers(t *testing.T, oracle string, texts []string, extension string) {
	inputs := make([]ComparisonInput, len(texts))
	for i, text := range texts {
		inputs[i] = ComparisonInput{ID: fmt.Sprint(i), Text: text, Filename: "0000" + extension}
	}
	CachedComparisonAnswers(t, oracle, inputs)
}
func libraryGapsAnswers(t *testing.T) {
	for _, source := range RecoveryLibraryGaps() {
		CachedRecoveryAnswer(t, ParserOracle(t), manifest(t, []string{source}), "--manifest")
	}
}

func TestProduct_EstreeGoOracle(t *testing.T) { t.Parallel(); ParserOracle(t) }

func TestProduct_AcceptanceOracle(t *testing.T) { t.Parallel(); AcceptanceMutantsOracle(t) }

func TestProduct_scalar_edges_go_oracle(t *testing.T) { t.Parallel(); ScalarEdgesOracle(t) }

func TestProduct_SyntaxMutantsGoOracle(t *testing.T) { t.Parallel(); SyntaxMutantsOracle(t) }

func TestProduct_ThreePortOracle(t *testing.T) { t.Parallel(); ThreePortMutantsOracle(t) }

func TestProduct_DeepMutantsFixture(t *testing.T) { t.Parallel(); DeepMutantsFixture(t) }

func TestProduct_EstreeBoundedAnswers(t *testing.T) { t.Parallel(); boundedAnswers(t) }

func TestProduct_EstreeExportAnswers(t *testing.T) {
	t.Parallel()
	textAnswers(t, ParserOracle(t), DecoratedExports(), ".ts")
}

func TestProduct_JSXMutantAnswers(t *testing.T) {
	t.Parallel()
	textAnswers(t, ParserOracle(t), JSXCases(), ".tsx")
}

func TestProduct_RecoveredExpressionsAnswer(t *testing.T) {
	t.Parallel()
	CachedRecoveryAnswer(t, ParserOracle(t), manifest(t, RecoveredExpressions()), "--manifest")
}

func TestProduct_RecoveredGrammarAnswer(t *testing.T) {
	t.Parallel()
	CachedRecoveryAnswer(t, ParserOracle(t), manifest(t, RecoveredGrammar()), "--manifest")
}

func TestProduct_RecoveryLibraryGapsAnswers(t *testing.T) { t.Parallel(); libraryGapsAnswers(t) }

func TestProduct_RecoveryMutantsAnswer(t *testing.T) {
	t.Parallel()
	CachedRecoveryAnswer(t, ScalarEdgesOracle(t), manifest(t, RecoveredGrammar()), "--manifest")
}
