package lower

import (
	"fmt"
	"slices"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// refusalRule is one check refuse's walk runs on every node it visits. Rules run in ordinal order and
// the first that refuses a node is the one a program hears, so the ordinal is the first-stop reason.
// A rule registers from its own file with a free ordinal and never edits refusals.go.
type refusalRule struct {
	ordinal int
	name    string
	// check refuses node, or prunes it: no later rule sees node and the walk doesn't enter it.
	check func(l *lowering, node *ast.Node) (prune bool, err error)
}

// refusalRegistry holds the rules in ordinal order.
type refusalRegistry struct {
	rules []refusalRule
}

// refusalRules is every per-node rule refuse runs, registered by each rule's file.
var refusalRules refusalRegistry

// register adds a rule that refuses a node or lets the walk go on.
func (registry *refusalRegistry) register(ordinal int, name string, check func(*lowering, *ast.Node) error) {
	registry.insert(refusalRule{ordinal: ordinal, name: name, check: func(l *lowering, node *ast.Node) (bool, error) {
		return false, check(l, node)
	}})
}

// registerPruning adds a rule that can also stop the walk at a node without refusing it.
func (registry *refusalRegistry) registerPruning(ordinal int, name string, check func(*lowering, *ast.Node) (bool, error)) {
	registry.insert(refusalRule{ordinal: ordinal, name: name, check: check})
}

// insert keeps the rules sorted by ordinal, so their order never depends on which file's init ran
// first, and refuses a second rule at an ordinal already taken.
func (registry *refusalRegistry) insert(rule refusalRule) {
	position := sort.Search(len(registry.rules), func(index int) bool { return registry.rules[index].ordinal >= rule.ordinal })
	if position < len(registry.rules) && registry.rules[position].ordinal == rule.ordinal {
		panic(fmt.Sprintf("refusal rules %s and %s both take ordinal %d", registry.rules[position].name, rule.name, rule.ordinal))
	}
	registry.rules = slices.Insert(registry.rules, position, rule)
}

// refusalPass is what one rule of refuse's walk leaves for a later rule at the same node.
type refusalPass struct {
	// checkedCasts are the as expressions whose cast is checked at runtime (castRefusal, 1900), which
	// stringWideningRefusal (3100) lets through.
	checkedCasts map[*ast.Node]bool
}
