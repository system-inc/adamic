package estree

import (
	"path/filepath"
	"testing"
)

// Products these tests read that no TestProduct_ built, so Workshop's combined proof (Oct 10) found each missing on the
// reading checkout: the Go oracle's answers for the recovery cases, the recovery port and its sourceLookahead mutant,
// the JSX mutant's source, port and answers, and the native recipe proof's two builds. Each is built here exactly as
// its reader asks for it, so Workshop builds it once and the reader only reads it.

func TestProduct_RecoveredExpressionsAnswer(t *testing.T) {
	t.Parallel()
	recoveryAnswer(t, goOracle(t), manifest(t, recoveredExpressions()), "--manifest")
}

func TestProduct_RecoveredGrammarAnswer(t *testing.T) {
	t.Parallel()
	recoveryAnswer(t, goOracle(t), manifest(t, recoveredGrammar()), "--manifest")
}

func TestProduct_RecoveryLibraryGapsAnswers(t *testing.T) {
	t.Parallel()
	for _, source := range recoveryLibraryGaps() {
		recoveryAnswer(t, goOracle(t), manifest(t, []string{source}), "--manifest")
	}
}

func TestProduct_RecoveryPort(t *testing.T) {
	t.Parallel()
	main, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	recoveryBuild(t, main, true)
}

func TestProduct_RecoveredExpressionMutantPort(t *testing.T) {
	t.Parallel()
	recoveryBuild(t, recoveredExpressionMutantPath(t), true)
}

func TestProduct_JSXMutantSource(t *testing.T) { t.Parallel(); jsxMutantPath(t) }
func TestProduct_JSXMutantPort(t *testing.T)   { t.Parallel(); miscBuild(t, jsxMutantPath(t)) }
func TestProduct_JSXMutantAnswers(t *testing.T) {
	t.Parallel()
	miscTextAnswers(t, miscOracle(t), jsxCases(), ".tsx")
}

func TestProduct_RecoveryNativeRecipe(t *testing.T) {
	t.Parallel()
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	recoveryNativeRecipe(t, setup)
}
