package lint

import "testing"

// The build phase and independently selected shards use these same recipes.
func TestProduct_LintFindsJsxMembership(test *testing.T) {
	test.Parallel()
	jsxMembershipOracle(test)
}

func TestProduct_LintFindsJsxParser(test *testing.T) {
	test.Parallel()
	jsxParserOracle(test)
}

func TestProduct_jsx_tree_lowered(test *testing.T) {
	test.Parallel()
	requireLintArtifacts(test, jsxTreeLowered(test, jsxParserEntry(test)), "program.c")
}

func TestProduct_jsx_tree_native(test *testing.T) {
	test.Parallel()
	requireLintExecutable(test, jsxTreeNative(test, jsxParserEntry(test), true))
}

func TestProduct_jsx_tree_setup(test *testing.T) {
	test.Parallel()
	requireLintJsxBundle(test, jsxPrepareTrees(test))
}
