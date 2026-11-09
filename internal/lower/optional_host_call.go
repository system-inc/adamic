package lower

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// A call through ?. to a method a Node library file declares, entry?.isFile() on a Dirent | undefined
// (#3625x53). optionalCall admits object?.method(...) when lib doesn't declare the method, and Node's
// @types files aren't lib's (load.IsNodeLibrary, not load.IsLibrary), so such a call went on to builtin,
// whose host lowering lowered it as a plain call unless it read the optional flag itself. Dirent's
// predicates (nodeFSDirectoryCall) don't: an undefined entry was a null access natively and a TypeError
// on the JavaScript backend, where Node gives undefined. So the call is NotYet unless the host lowering
// of every Node declaration it may reach is named in optionalHostCallGuarded.

// optionalHostCallGuarded names each Node host method whose own lowering of a call through ?. tests the
// receiver, evaluates it once and gives undefined where it's missing, by its registration name
// (nodeLibraryName). It's empty: of the host lowerings that read NodeFlagsOptionalChain, each refuses
// the optional call or read (Error.captureStackTrace and global.gc, process.platform, process.env,
// realpathSync.native, process.stdout.write and isTTY, fs.watchFile and unwatchFile, Stats' predicates,
// fs.watch and FSWatcher.on), and the rest don't read it. A host lowering that comes to guard its
// receiver adds its member here, with a pin that runs it on an undefined receiver against Node.
var optionalHostCallGuarded = map[string]bool{}

// optionalHostCallNotYet is what a call through ?. to a Node host method stops as when its lowering
// isn't named in optionalHostCallGuarded, followed by the members it reaches.
const optionalHostCallNotYet = "a call through ?. to a Node host method whose lowering doesn't guard an optional receiver"

// optionalHostCall judges object?.method(...) on a method lib doesn't declare, which optionalCall
// otherwise admits: NotYet when any declaration of the method, or of a root of it (each member's own,
// for a property of a union such as Stats | Dirent), is in a Node library file and that member's
// lowering isn't named in optionalHostCallGuarded. It fails closed: a root with no declaration, or a
// declaration with no file, can't be told from a Node host method's, so the call stops too.
func (l *lowering) optionalHostCall(call *ast.Node, method *ast.Symbol) error {
	unguarded := []string{}
	for _, root := range append([]*ast.Symbol{method}, l.checker.GetRootSymbols(method)...) {
		if len(root.Declarations) == 0 {
			return l.notYet(call, optionalHostCallNotYet+" (a member whose declaration can't be read; #3625x53)")
		}
		for _, declaration := range root.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			if source == nil {
				return l.notYet(call, optionalHostCallNotYet+" (a member whose declaration can't be read; #3625x53)")
			}
			if !load.IsNodeLibrary(source) {
				continue
			}
			if name := nodeLibraryName(root); !optionalHostCallGuarded[name] && !slices.Contains(unguarded, name) {
				unguarded = append(unguarded, name)
			}
		}
	}
	if len(unguarded) == 0 {
		return nil
	}
	slices.Sort(unguarded)
	return l.notYet(call, optionalHostCallNotYet+" ("+strings.Join(unguarded, ", ")+"; #3625x53)")
}
