package lint

import "testing"

// The build phase and independently selected shards use these same recipes.
func TestProduct_LintFindsJSXMembership(t *testing.T) {
	t.Parallel()
	jsxMembershipOracle(t)
}

func TestProduct_LintFindsJSXParser(t *testing.T) {
	t.Parallel()
	jsxParserOracle(t)
}

func TestProduct_jsx_tree_lowered(t *testing.T) {
	t.Parallel()
	requireLintArtifacts(t, jsxTreeLowered(t, jsxParserEntry(t)), "program.c")
}

func TestProduct_jsx_tree_native(t *testing.T) {
	t.Parallel()
	requireLintExecutable(t, jsxTreeNative(t, jsxParserEntry(t), true))
}

func TestProduct_jsx_tree_setup(t *testing.T) {
	t.Parallel()
	requireLintJSXBundle(t, jsxPrepareTrees(t))
}
