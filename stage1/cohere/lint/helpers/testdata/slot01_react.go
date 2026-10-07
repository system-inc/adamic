package react

// Call the real private Go helper from the oracle-only overlay.
import "github.com/microsoft/TypeScript/tsc/shim/ast"

func AdamicComponentBase(node *ast.Node) bool { return isComponentBase(node) }
