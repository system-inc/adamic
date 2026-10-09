package unused_exports

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Island is a cluster of declarations that are dead together.
//
// The cluster is the finding, not the individual line. An abandoned export drags a private world
// behind it — the helper only it called, the table only that helper read — and a flat list prints
// those as three unrelated findings in three places. Grouping them says the thing a reader actually
// needs to know: this is one abandoned direction, and here is all of it.
type Island struct {
	// Members are the dead declarations, ordered so the exported ones come first. The export is the
	// entry point a reader recognizes, and the private members are the world behind it.
	Members []DeadDeclaration

	// Exported is how many members are exported, which is what separates an abandoned public
	// direction from a private tangle nobody ever reached.
	Exported int
}

// DeadDeclaration is one declaration the closure could not mark.
type DeadDeclaration struct {
	FileName string
	Name     string
	Kind     string
	Range    TextSpan
	Exported bool

	// Intentional carries a `cohere-keep` written above the declaration.
	Intentional intentionalReason
}

// declarationRecord is one declaration in the program, with the symbol the closure marks it by.
type declarationRecord struct {
	symbol *ast.Symbol
	dead   DeadDeclaration
}

// closure computes what is alive by marking the roots and propagating along reference edges until
// nothing changes.
//
// # What a root is
//
// Three things, and getting any of them wrong is the failure mode that matters:
//
//   - A reference from module top level. Its referrer is nil, which the index records as a genuine
//     root rather than as an absent edge. Module-level code runs when the module loads, so anything
//     it touches is alive.
//   - Every declaration in a file the root set spares. A Next.js `page.tsx` is loaded by the
//     framework with no import anywhere, so its exports and everything they reach are alive.
//   - Every declaration whose referrer is a symbol this inventory never saw. A reference from a
//     `.d.ts`, from a node_modules file, or from a construct this analysis does not model produces
//     a referrer with no record, and treating that as dead would condemn code on the strength of a
//     gap in our own inventory.
//
// # Why this is opt-in
//
// The transitive claim is strictly stronger than the flat one and it FAILS DIFFERENTLY. A wrong
// root in the flat report mis-reports one file. A wrong root here cascades: everything that was
// alive only through that root goes dark at once, and the report gets dramatically bigger in a way
// that looks like a discovery rather than like a bug. Keeping the flat view as the default means
// both numbers can be read against each other, and a reader who has seen them agree on the shape of
// the tree can then trust the bigger one. That is worth one flag.
//
// # Convergence
//
// A worklist rather than a repeat-until-stable sweep, so termination is structural rather than
// hoped for: a symbol is pushed only on the transition from unmarked to marked, so each is pushed at
// most once and the loop is bounded by the number of declarations. There is no fixpoint to miss and
// no pass count to tune. The pass count is reported anyway — as the longest chain the marking
// followed — because that number is the one that says whether the graph is deep or the closure quit
// early.
func computeIslands(index *referenceIndex, inventory []declarationRecord, rootSymbols map[*ast.Symbol]bool) ([]Island, int) {
	if len(inventory) == 0 {
		return nil, 0
	}

	// Every symbol this analysis has a declaration for. A referrer outside this set is a reference
	// from somewhere we do not model, and is treated as a root rather than as nothing.
	known := make(map[*ast.Symbol]bool, len(inventory))
	for _, record := range inventory {
		known[record.symbol] = true
	}

	// Forward edges: marking a symbol alive must mark what IT references. The index is keyed the
	// other way, so it is inverted once here rather than searched repeatedly.
	reaches := make(map[*ast.Symbol][]*ast.Symbol, len(index.referencedBy))
	for target, referrers := range index.referencedBy {
		for _, referrer := range referrers {
			if referrer == nil {
				// Module top level. Not an edge — a root, handled below.
				continue
			}
			reaches[referrer] = append(reaches[referrer], target)
		}
	}

	alive := make(map[*ast.Symbol]bool, len(inventory))
	depth := make(map[*ast.Symbol]int, len(inventory))
	var worklist []*ast.Symbol

	markRoot := func(symbol *ast.Symbol) {
		if symbol == nil || alive[symbol] {
			return
		}
		alive[symbol] = true
		depth[symbol] = 0
		worklist = append(worklist, symbol)
	}

	// Root one: anything referenced from module top level, or referenced by something outside the
	// inventory.
	for target, referrers := range index.referencedBy {
		for _, referrer := range referrers {
			if referrer == nil || !known[referrer] {
				markRoot(target)
				break
			}
		}
	}

	// Root two: every declaration the root set spared, plus anything the caller declared a root.
	for symbol := range rootSymbols {
		markRoot(symbol)
	}

	longest := 0
	for len(worklist) > 0 {
		current := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, target := range reaches[current] {
			if alive[target] {
				continue
			}
			alive[target] = true
			depth[target] = depth[current] + 1
			if depth[target] > longest {
				longest = depth[target]
			}
			worklist = append(worklist, target)
		}
	}

	// Everything the marking never reached is dead. Grouping it into islands is a connected-component
	// walk over the SAME edges, ignoring direction: two dead declarations belong together when one
	// references the other, whichever way the reference points, because a reader deleting one will be
	// looking at the other.
	deadBySymbol := make(map[*ast.Symbol]DeadDeclaration, len(inventory))
	var deadSymbols []*ast.Symbol
	for _, record := range inventory {
		if alive[record.symbol] {
			continue
		}
		if _, already := deadBySymbol[record.symbol]; already {
			continue
		}
		deadBySymbol[record.symbol] = record.dead
		deadSymbols = append(deadSymbols, record.symbol)
	}

	neighbours := make(map[*ast.Symbol][]*ast.Symbol, len(deadSymbols))
	for target, referrers := range index.referencedBy {
		if _, targetDead := deadBySymbol[target]; !targetDead {
			continue
		}
		for _, referrer := range referrers {
			if referrer == nil {
				continue
			}
			if _, referrerDead := deadBySymbol[referrer]; !referrerDead {
				continue
			}
			if referrer == target {
				// A symbol referencing only itself. Not an edge to anywhere, and deliberately not
				// treated as one: a recursive function nothing else calls is dead, and letting the
				// self-edge form a two-member island would print it as a cluster of one thing twice.
				continue
			}
			neighbours[target] = append(neighbours[target], referrer)
			neighbours[referrer] = append(neighbours[referrer], target)
		}
	}

	seen := make(map[*ast.Symbol]bool, len(deadSymbols))
	var islands []Island
	for _, start := range deadSymbols {
		if seen[start] {
			continue
		}
		var members []DeadDeclaration
		stack := []*ast.Symbol{start}
		seen[start] = true
		for len(stack) > 0 {
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			members = append(members, deadBySymbol[current])
			for _, neighbour := range neighbours[current] {
				if seen[neighbour] {
					continue
				}
				seen[neighbour] = true
				stack = append(stack, neighbour)
			}
		}

		exported := 0
		for _, member := range members {
			if member.Exported {
				exported++
			}
		}
		// Exported members first: the export is the name a reader recognizes and the private members
		// are the world behind it, so leading with the export is what makes the cluster legible.
		sort.SliceStable(members, func(left, right int) bool {
			if members[left].Exported != members[right].Exported {
				return members[left].Exported
			}
			if members[left].FileName != members[right].FileName {
				return members[left].FileName < members[right].FileName
			}
			return members[left].Name < members[right].Name
		})
		islands = append(islands, Island{Members: members, Exported: exported})
	}

	// Largest first. The reader's question is "what is the biggest thing I abandoned", and a report
	// sorted by filename buries a twelve-member cluster between two singletons.
	sort.SliceStable(islands, func(left, right int) bool {
		if len(islands[left].Members) != len(islands[right].Members) {
			return len(islands[left].Members) > len(islands[right].Members)
		}
		return islands[left].Members[0].FileName < islands[right].Members[0].FileName
	})

	return islands, longest
}

// fileDeclarations inventories every declaration in one file, exported or not.
//
// The closure needs the private ones as much as the public ones: `ScoreWindow` is unused, so it is
// dead; `normalizeDecay` is referenced but only by `ScoreWindow`, so it is dead too; `DecayTable`
// after it. Only the first of those three is an export, and a report that inventoried exports alone
// would print one line where the truth is a cluster of three.
//
// This deliberately does NOT use the checker. `node.Symbol()` is the same symbol pointer the index
// is keyed on, established by the same probe that established the referrer path, and reaching for a
// checker call per declaration would put back exactly the cost the referrer path avoided.
func fileDeclarations(file *ast.SourceFile, text string) []declarationRecord {
	fileName := file.FileName()
	var records []declarationRecord

	add := func(name *ast.Node, owner *ast.Node, kind string, exported bool, span *ast.Node) {
		if name == nil || owner == nil {
			return
		}
		symbol := owner.Symbol()
		if symbol == nil {
			// No symbol means the binder did not bind it, and marking a declaration dead on the
			// strength of a binding we never got is how live code gets condemned.
			return
		}
		records = append(records, declarationRecord{
			symbol: symbol,
			dead: DeadDeclaration{
				FileName:    fileName.AsString(),
				Name:        name.Text(),
				Kind:        kind,
				Range:       TextSpan{Position: span.Pos(), End: span.End()},
				Exported:    exported,
				Intentional: findIntentionalMarker(leadingText(text, span)),
			},
		})
	}

	for _, statement := range file.Statements.Nodes {
		exported := hasExportModifier(statement)
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			add(statement.Name(), statement, "function", exported, statement)
		case ast.KindClassDeclaration:
			add(statement.Name(), statement, "class", exported, statement)
		case ast.KindInterfaceDeclaration:
			add(statement.Name(), statement, "interface", exported, statement)
		case ast.KindTypeAliasDeclaration:
			add(statement.Name(), statement, "type", exported, statement)
		case ast.KindEnumDeclaration:
			add(statement.Name(), statement, "enum", exported, statement)
		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				continue
			}
			for _, declaration := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				name := declaration.Name()
				// Only a plain identifier, matching the flat report's rule: a destructuring pattern
				// binds several names whose liveness this does not separate.
				if name != nil && name.Kind == ast.KindIdentifier {
					add(name, declaration, "constant", exported, statement)
				}
			}
		}
	}
	return records
}

// collectRootSymbols marks every top-level declaration in a spared file as a closure root.
//
// A file the root set spares is alive by convention — a Next.js route, a test, a locale loaded by a
// computed `import()` — so anything it declares is alive and anything it reaches is alive. This is
// the single most load-bearing input to the closure: on this tree 1,228 files are spared, and
// omitting their declarations would darken most of the application in one stroke while looking like
// a discovery.
func collectRootSymbols(file *ast.SourceFile, into map[*ast.Symbol]bool) {
	for _, statement := range file.Statements.Nodes {
		switch statement.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration:
			if symbol := statement.Symbol(); symbol != nil {
				into[symbol] = true
			}
		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				continue
			}
			for _, declaration := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				if symbol := declaration.Symbol(); symbol != nil {
					into[symbol] = true
				}
			}
		}
	}
}
