// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"strings"
)

// NotYet is a construct stage 0 can't lower yet, with where it is.
type NotYet struct {
	Where string
	What  string
}

func (notYet *NotYet) Error() string {
	return fmt.Sprintf("%s: stage 0 can't lower %s yet", notYet.Where, notYet.What)
}

// Refused is something Adamic 0.1 doesn't allow at all (docs/0.1.md), as opposed to something stage 0
// hasn't learned yet. The difference is a promise, so the message keeps them apart.
type Refused struct {
	// cycle identifies the slot-cycle proof, independently of diagnostic wording.
	cycle bool
	// lyingTag is why a type predicate was refused for a comparison on a tag that can lie, for a
	// caller passing that predicate to name too.
	lyingTag string
	Where    string
	What     string
	Fix      string
}

func (refused *Refused) Error() string {
	return fmt.Sprintf("%s: Adamic 0.1 refuses %s; %s", refused.Where, refused.What, refused.Fix)
}

func (lowering *lowering) notYet(node *ast.Node, what string) error {
	return &NotYet{Where: lowering.program.Where(node), What: what}
}

// describe names a node's kind for a person: "a VariableStatement", "an ImportDeclaration".
func describe(node *ast.Node) string {
	name := strings.TrimPrefix(node.Kind.String(), "Kind")
	if strings.ContainsAny(name[:1], "AEIOU") {
		return "an " + name
	}
	return "a " + name
}
