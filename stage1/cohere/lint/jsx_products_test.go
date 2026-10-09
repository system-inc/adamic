package lint

import "testing"

// The build phase and independently selected shards use these same recipes.
func TestProduct_jsx_membership(t *testing.T) {
	t.Parallel()
	jsxMembershipOracle(t)
}

func TestProduct_jsx_parser(t *testing.T) {
	t.Parallel()
	jsxParserOracle(t)
}

func TestProduct_jsx_tree_lowered(t *testing.T) {
	t.Parallel()
	jsxTreeLowered(t, jsxParserEntry(t))
}

func TestProduct_jsx_tree_native(t *testing.T) {
	t.Parallel()
	jsxTreeNative(t, jsxParserEntry(t), true)
}

func TestProduct_jsx_tree_setup(t *testing.T) {
	t.Parallel()
	jsxPrepareTrees(t)
}
