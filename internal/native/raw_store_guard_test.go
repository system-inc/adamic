package native

import (
	"cmp"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// rawStoreAllowed is every site TestRawSlotStoresAreGuarded finds unguarded today, each with why its
// slot can never be a graph member's, or why the store is graph-aware where the walk can't see it. A
// key is file:function:ordinal, the ordinal counting the function's unguarded sites in source order,
// so a line moving doesn't move a key. The list only shrinks: an entry that matches no site fails as
// stale.
var rawStoreAllowed = map[string]string{
	"class_errors.go:emitter.makeError:1":              "a fresh error object's name, a string; the error is never adopted",
	"constant_tables.go:emitter.emitConstantRecords:1": "a fresh constant-table object and its fresh number array; constant tables exclude graph literals",
	"emit_locals.go:emitter.replaceIterationCell:1":    "an async or generator frame cell, never a graph member (frames are unmovable)",
	"emit_locals.go:emitter.store:1":                   "an async frame cell borrowing its source's value, never a graph member (async frames are unmovable)",
	"emit_locals.go:emitter.store:2":                   "a counted cell, not a GraphCell, so never adopted (GraphCell spreads to every cell of an adopted environment)",
	"emit_objects.go:emitter.objectLiteral:1":          "a region value, stored only into a region-allocated literal, which is never adopted",
	"emit_records.go:emitter.recordCall:1":             "a fresh entries pair's key, a string, before the pair's adoption",
	"emit_statements.go:emitter.writeProperty:1":       "a number or boolean box behind a checked field, which graph programs refuse",
	"emit_statements.go:emitter.writeProperty:2":       "keptValue is e.keptIn(object, value), made above for every reference",
	"exceptions.go:emitter.throwNamed:1":               "a fresh error object's name, a static string; the error is never adopted",
	"foreign.go:emitter.blockFunctions:1":              "an Apple block's deliver filling its own stack array of arguments for the closure call with a retained box, never an object's slot",
	"foreign.go:emitter.blockFunctions:2":              "an Apple block's deliver filling its own stack array of arguments for the closure call with a fresh string, never an object's slot",
	"reuse.go:emitter.checkDefinedWhen:1":              "a fresh error object's name, a static string; the error is never adopted",
	"reuse.go:emitter.reused:1":                        "a reused spread bails on graph literals, and reuses only a source adamic_graph_counted says is no member",
}

// rawStoreSite is one string literal an emitter call is handed that stores a reference into a slot
// or pops one out of an array.
type rawStoreSite struct {
	file, function, key string
	position            token.Pos
	at, literal         string
	guarded             bool
}

// rawStoreHelpers are the Go helpers that store a reference graph-aware: in a graph program each
// holds, takes or drops through the graph.
var rawStoreHelpers = []string{"keptIn", "heldReferenceIn", "heldIn", "dropIn"}

// rawStorePrefixes are a store through a slot. A compound literal's `{.reference = ` is a value, not a
// store, and matches none of them.
var rawStorePrefixes = []string{"->reference = ", "].reference = ", ".value.reference = ", "%s.reference = "}

// In a graph program some objects are graph-region members, and a reference stored into a member's
// slot has to go through the graph (keptIn, heldReferenceIn, heldIn or dropIn in Go, or
// adamic_graph_hold, adamic_graph_take or adamic_graph_escape in the C), or the region's counts go
// wrong: a leak, a crash, or an early free. Three raw stores were found by hand: two in
// logical_assignment.go and one in emit_slots.go's typeof of a pop. This reads every emitter source
// file, finds each string literal handed to a call that stores a reference into a slot or pops one
// out of an array, and fails on one that's neither graph-aware nor listed in rawStoreAllowed with why,
// so the next raw store fails here instead of shipping. A listed entry that matches no site fails too,
// so the list may only shrink.
//
// Its mutants: logical_assignment.go's field store using retained(value) instead of
// e.heldReferenceIn(receiver, value) goes red as an unlisted site (M1); emit_slots.go's typeof-of-pop
// path without its escape goes red the same way (M2); and an allowlist entry for a site that no
// longer exists goes red as stale (M3).
func TestRawSlotStoresAreGuarded(t *testing.T) {
	t.Parallel()
	sites := rawStoreSites(t)
	unguarded := map[string]bool{}
	guarded := 0
	for _, site := range sites {
		if site.guarded {
			guarded++
			continue
		}
		unguarded[site.key] = true
		if _, listed := rawStoreAllowed[site.key]; !listed {
			t.Errorf("%s (%s) stores a reference into a slot, or pops one, with no graph-aware helper: %q. Store it through keptIn, heldReferenceIn, heldIn or dropIn (or adamic_graph_hold, take or escape in the C), or, if the slot can never be a graph member's, list it in rawStoreAllowed with why", site.key, site.at, site.literal)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(rawStoreAllowed)) {
		if rawStoreAllowed[key] == "" {
			t.Errorf("rawStoreAllowed lists %s without why", key)
		}
		if !unguarded[key] {
			t.Errorf("rawStoreAllowed lists %s, which matches no unguarded site: stale, remove it", key)
		}
	}
	// store's GraphCell branch, a spread literal's fields, the take loops, ArrayPop and logical
	// assignment are graph-aware today, so a reading that finds fewer has stopped seeing the emitter.
	if guarded < 5 {
		t.Fatalf("read %d guarded stores from the emitter's source; the test's reading of internal/native is out of date", guarded)
	}
}

// rawStoreSites is every slot store and pop the emitter's source hands a call, each unguarded one
// keyed as rawStoreAllowed is.
func rawStoreSites(t *testing.T) []rawStoreSite {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	var sites []rawStoreSite
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function := "(package)"
			if decl, isFunction := declaration.(*ast.FuncDecl); isFunction {
				function = rawStoreFunctionName(decl)
			}
			escaped := rawStoreEscapesPops(declaration)
			ast.Inspect(declaration, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				for _, argument := range call.Args {
					literal, isLiteral := argument.(*ast.BasicLit)
					if !isLiteral || literal.Kind != token.STRING {
						continue
					}
					text, err := strconv.Unquote(literal.Value)
					if err != nil {
						t.Fatalf("%s: %v", fileSet.Position(literal.Pos()), err)
					}
					store, pop := rawStoreSlotStore(text), strings.Contains(text, "elements[--")
					if !store && !pop {
						continue
					}
					guarded := true
					if store && !strings.Contains(text, "adamic_graph_hold(") && !strings.Contains(text, "adamic_graph_take(") && !strings.Contains(text, "adamic_graph_escape(") && !rawStoreCallsHelper(call) {
						guarded = false
					}
					if pop && !strings.Contains(text, "adamic_graph_escape") && !rawStoreMentions(call, "adamic_graph_escape") && !escaped {
						guarded = false
					}
					position := fileSet.Position(literal.Pos())
					sites = append(sites, rawStoreSite{
						file:     path,
						function: function,
						position: literal.Pos(),
						at:       position.Filename + ":" + strconv.Itoa(position.Line),
						literal:  text,
						guarded:  guarded,
					})
				}
				return true
			})
		}
	}
	// A call's literal is seen before a call nested in an earlier argument, so the ordinals are counted
	// in source order, not in the walk's.
	slices.SortStableFunc(sites, func(left, right rawStoreSite) int { return cmp.Compare(left.position, right.position) })
	ordinals := map[string]int{}
	for index := range sites {
		if sites[index].guarded {
			continue
		}
		prefix := sites[index].file + ":" + sites[index].function
		ordinals[prefix]++
		sites[index].key = prefix + ":" + strconv.Itoa(ordinals[prefix])
	}
	return sites
}

// rawStoreFunctionName is a function's name, and a method's as Receiver.Name.
func rawStoreFunctionName(decl *ast.FuncDecl) string {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return decl.Name.Name
	}
	receiver := decl.Recv.List[0].Type
	if star, isStar := receiver.(*ast.StarExpr); isStar {
		receiver = star.X
	}
	if index, isIndex := receiver.(*ast.IndexExpr); isIndex {
		receiver = index.X
	}
	if index, isIndex := receiver.(*ast.IndexListExpr); isIndex {
		receiver = index.X
	}
	if name, isName := receiver.(*ast.Ident); isName {
		return name.Name + "." + decl.Name.Name
	}
	return decl.Name.Name
}

// rawStoreSlotStore says text stores a reference through a slot: one of rawStorePrefixes, not
// followed by NULL, since a NULL store holds nothing.
func rawStoreSlotStore(text string) bool {
	for _, prefix := range rawStorePrefixes {
		for rest := text; ; {
			at := strings.Index(rest, prefix)
			if at < 0 {
				break
			}
			rest = rest[at+len(prefix):]
			if !strings.HasPrefix(rest, "NULL") {
				return true
			}
		}
	}
	return false
}

// rawStoreCallsHelper says some argument of call, at any depth, calls one of rawStoreHelpers.
func rawStoreCallsHelper(call *ast.CallExpr) bool {
	found := false
	for _, argument := range call.Args {
		ast.Inspect(argument, func(node ast.Node) bool {
			inner, isCall := node.(*ast.CallExpr)
			if !isCall {
				return !found
			}
			switch function := inner.Fun.(type) {
			case *ast.SelectorExpr:
				found = found || slices.Contains(rawStoreHelpers, function.Sel.Name)
			case *ast.Ident:
				found = found || slices.Contains(rawStoreHelpers, function.Name)
			}
			return !found
		})
	}
	return found
}

// rawStoreMentions says some argument of call, at any depth, is a string literal holding text.
func rawStoreMentions(call *ast.CallExpr, text string) bool {
	found := false
	for _, argument := range call.Args {
		ast.Inspect(argument, func(node ast.Node) bool {
			literal, isLiteral := node.(*ast.BasicLit)
			if isLiteral && literal.Kind == token.STRING && strings.Contains(literal.Value, text) {
				found = true
			}
			return !found
		})
	}
	return found
}

// rawStoreEscapesPops says a function hands its pops over through an escape: some branch of it that
// tests GraphTypes builds a literal naming adamic_graph_escape. ArrayPop's arm wraps the popped value
// that way, and typeOfReference wraps the popped slot's reference. It judges by function, not by pop,
// so evaluate's maybe pop, a number or a boolean that needs no escape, is covered by ArrayPop's branch
// beside it.
func rawStoreEscapesPops(declaration ast.Decl) bool {
	found := false
	ast.Inspect(declaration, func(node ast.Node) bool {
		branch, isBranch := node.(*ast.IfStmt)
		if !isBranch || found || !rawStoreTestsGraphTypes(branch.Cond) {
			return !found
		}
		ast.Inspect(branch.Body, func(inner ast.Node) bool {
			if literal, isLiteral := inner.(*ast.BasicLit); isLiteral && literal.Kind == token.STRING && strings.Contains(literal.Value, "adamic_graph_escape(") {
				found = true
			}
			return !found
		})
		return !found
	})
	return found
}

// rawStoreTestsGraphTypes says condition reads a GraphTypes field.
func rawStoreTestsGraphTypes(condition ast.Expr) bool {
	found := false
	ast.Inspect(condition, func(node ast.Node) bool {
		if selector, isSelector := node.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "GraphTypes" {
			found = true
		}
		return !found
	})
	return found
}
