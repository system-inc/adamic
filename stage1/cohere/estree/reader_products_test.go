package estree

import (
	"path/filepath"
	"testing"
)

// Products these tests read that no TestProduct_ built, so Workshop's combined proof (Oct 10) found each missing on the
// reading checkout: the Go oracle's answers for the recovery cases, the recovery port and its sourceLookahead mutant,
// the JSX mutant's source, port and answers, and the native recipe proof's two builds. Each is built here exactly as
// its reader asks for it, so Workshop builds it once and the reader only reads it.

func TestProduct_EstreeFindsRecoveredExpressionsAnswer(test *testing.T) {
	test.Parallel()
	recoveryAnswer(test, goOracle(test), manifest(test, recoveredExpressions()), "--manifest")
}

func TestProduct_EstreeFindsRecoveredGrammarAnswer(test *testing.T) {
	test.Parallel()
	recoveryAnswer(test, goOracle(test), manifest(test, recoveredGrammar()), "--manifest")
}

func TestProduct_EstreeFindsRecoveryLibraryGapsAnswers(test *testing.T) {
	test.Parallel()
	for _, source := range recoveryLibraryGaps() {
		recoveryAnswer(test, goOracle(test), manifest(test, []string{source}), "--manifest")
	}
}

func TestProduct_RecoveryPort(test *testing.T) {
	test.Parallel()
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	recoveryBuild(test, main, true)
}

func TestProduct_RecoveredExpressionMutantPort(test *testing.T) {
	test.Parallel()
	recoveryBuild(test, recoveredExpressionMutantPath(test), true)
}

func TestProduct_JSXMutantSource(test *testing.T) { test.Parallel(); jsxMutantPath(test) }
func TestProduct_JSXMutantPort(test *testing.T) {
	test.Parallel()
	miscBuild(test, jsxMutantPath(test))
}
func TestProduct_EstreeFindsJSXMutantAnswers(test *testing.T) {
	test.Parallel()
	miscTextAnswers(test, miscOracle(test), jsxCases(), ".tsx")
}

func TestProduct_RecoveryNativeRecipe(test *testing.T) {
	test.Parallel()
	setup := beginRecoverySetup(test)
	defer setup.report(test)
	recoveryNativeRecipe(test, setup)
}
