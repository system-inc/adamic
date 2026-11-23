package native

import (
	"cmp"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// unionNarrowProof is why narrow's reads past its kind test are safe where the narrow isn't Checked.
const unionNarrowProof = "a Checked narrow stops on the line above unless the heap's kind is the member's. An unchecked one stands on lowering's proof that the union is this member (ir.Narrow: by typeof, === or assignment): lower's narrowed_union_member tests typeof, and Array.isArray where the member excludes an array, before its narrow, and to_primitive.go narrows only under adamic_union_kind's arms for its member. Adamic's (b1) keeps that proof true of the layout: a type whose value can be an array or a box is held as Array or Union, never as the Object, String or scalar member this reads"

// unionLayoutListed is every Union layout site TestUnionLayoutReadsAreKindTested finds without a kind
// test before it on every path. tested is the Go condition under which one is, as the walk renders it,
// empty when none is; reason is why the read is safe where none is. A key is file:function:ordinal, the
// ordinal counting the function's Union layout sites in source order, so a line moving doesn't move a
// key. An entry that matches no such site fails as stale.
var unionLayoutListed = map[string]unionLayoutListing{
	"class_accessors.go:emitter.accessorCall:1": {
		reason: "the checker typed this accessor read as number, so the getter the descriptor names returns a number; one held as Union hands it in a box, which the descriptor's `->type`, compared with Union's tag, picks out before this reads the box",
	},
	"class_accessors.go:emitter.accessorCall:2": {
		reason: "the checker typed this accessor read as boolean, so the getter the descriptor names returns a boolean; one held as Union hands it in a box, which the descriptor's `->type`, compared with Union's tag, picks out before this reads the box",
	},
	"emit_branches.go:emitter.coalesce:1": {
		reason: "the checker typed value ?? fallback as number, and this read is used only past the present test (`!= NULL && != &adamic_null`), so the present union is a number and this reads its box",
	},
	"emit_branches.go:emitter.coalesce:2": {
		reason: "the checker typed value ?? fallback as boolean, and this read is used only past the present test (`!= NULL && != &adamic_null`), so the present union is a boolean and this reads its box",
	},
	"emit_branches.go:emitter.coalesce:3": {
		reason: "the checker typed value ?? fallback as number | undefined, and this read is used only past the present test, so the present union is a number and this reads its box",
	},
	"emit_branches.go:emitter.coalesce:4": {
		reason: "the checker typed value ?? fallback as boolean | undefined, and this read is used only past the present test, so the present union is a boolean and this reads its box",
	},
	"emit_branches.go:emitter.coalesce:5": {
		reason: "the checker typed value ?? fallback as the reference type Of, and this cast is used only past the present test; by Adamic's (b1) a value whose heap kind can differ from Of's is held as Union, never as Of, so a present value of type Of has Of's layout",
	},
	"taste.go:logicalConverted:1": {
		reason: "the checker typed this && or || as number, and logicalValue reads the left only on the branch that keeps it, so the union it keeps is a number and this reads its box",
	},
	"taste.go:logicalConverted:2": {
		reason: "the checker typed this && or || as boolean, and logicalValue reads the left only on the branch that keeps it, so the union it keeps is a boolean and this reads its box",
	},
	"taste.go:logicalConverted:3": {
		reason: "the checker typed this && as number | undefined: undefined is the `NULL` tested first, and the falsy union kept otherwise is a number, whose box this reads",
	},
	"taste.go:logicalConverted:4": {
		reason: "the checker typed this && as boolean | undefined: undefined is the `NULL` tested first, and the falsy union kept otherwise is a boolean, whose box this reads",
	},
	"taste.go:logicalConverted:5": {
		reason: "the checker typed this && or || as the reference type to, so the union left it keeps is a to; by Adamic's (b1) a value whose heap kind can differ from to's is held as Union, never as to, so it has to's layout",
	},
	"union.go:converted:1": {
		reason: "no caller reaches this case with a Union today: coalesce and logicalConverted read a Union themselves before they fall back to converted, an array search's SearchFirst is its element, Union or the element's Maybe (lowering's arraySearchContract), and every other caller converts to Union or Weak",
	},
	"union.go:emitter.narrow:1": {
		reason: "the record test is written into the Checked narrow's own line after `->kind != adamic_kind_object ||`, so C's short circuit reads the record flag only off an object",
	},
	"union.go:emitter.narrow:2": {
		reason: "the record test is written into the Checked narrow's own line after `->kind != adamic_kind_object ||`, so C's short circuit reads the record flag only off an object",
	},
	"union.go:emitter.narrow:3": {
		tested: "narrow.Checked",
		reason: unionNarrowProof,
	},
	"union.go:emitter.narrow:4": {
		tested: "narrow.Checked",
		reason: unionNarrowProof,
	},
	"union.go:emitter.narrow:5": {
		tested: "narrow.Checked",
		reason: unionNarrowProof,
	},
	"union.go:emitter.narrow:6": {
		tested: "narrow.Checked",
		reason: unionNarrowProof,
	},
	"union.go:emitter.narrow:7": {
		tested: "narrow.Checked",
		reason: unionNarrowProof,
	},
}

// unionLayoutListing is one entry of unionLayoutListed.
type unionLayoutListing struct {
	tested string
	reason string
}

// unionLayoutSite is one string literal the emitter's source writes a Union operand's layout read with.
type unionLayoutSite struct {
	file, function, key string
	position            token.Pos
	at, literal         string
	tested              string
	dominated           bool
}

// unionLayoutLiteral is a string literal in a function, with the nodes it sits in, outermost first.
type unionLayoutLiteral struct {
	literal   *ast.BasicLit
	text      string
	ancestors []ast.Node
}

// unionLayoutName is a Go variable a function makes hold a C Union value, read after it's made in the
// block that makes it.
type unionLayoutName struct {
	name  string
	scope ast.Node
	after token.Pos
}

// unionLayoutCasts are a cast to a heap's layout: a number's or a boolean's box, an object, an array,
// a string, a closure or a map. adamic_heap is the Union's own type, which has no layout to read.
var unionLayoutCasts = regexp.MustCompile(`_box \*\)|adamic_(object|array|string|closure|map) \*\)`)

// unionLayoutTypedCast is a cast to a type the same call spells through cType.
var unionLayoutTypedCast = regexp.MustCompile(`\(%s\)\(?%s`)

// unionLayoutKindTest is a kind test: a heap's or a view snapshot's kind compared, or switched on.
var unionLayoutKindTest = regexp.MustCompile(`(->|\.)kind\s*[!=]=|switch \(%s->kind\)`)

// Adamic's ruling (b1) on #hae423v: ir.Object promises its value is an adamic object, and a value
// whose heap kind can be something else at run time is held as ir.Union, whose every consumer asks the
// kind before it reads a layout. This reads every emitter source file and finds each string literal
// that casts to a heap layout (unionLayoutCasts, or a cast through cType) where the operand is a
// Union: in the body of a branch that tests a type `== ir.Union`, in a case of ir.Union, in a call
// handed ir.Union (a C tag test against it), in a function taking an ir.Narrow, or in a call naming a
// variable own, snapshot or converted made a Union. Each such site must have a kind test emitted before
// it on every Go path through its function, or be listed in unionLayoutListed with the condition under
// which one is and why the read is safe where none is. Before means earlier in the function's source
// and outside any branch that doesn't also hold the site. It reads the Go, not the C, so it can't tell
// which value a kind test asks about; a site it calls tested is one a reader can check against the
// test above it.
//
// A new unlisted site fails, so the next Union cast is classified when it's written. A listed site that
// no longer exists, or is now tested on every path, fails as stale, and a listed site whose condition
// moved fails too. Its mutant: union.go's Checked narrow without its kind test leaves narrow's reads,
// listed as tested when narrow.Checked, tested on no path, and goes red.
func TestUnionLayoutReadsAreKindTested(t *testing.T) {
	t.Parallel()
	sites := unionLayoutSites(t)
	found := map[string]bool{}
	dominated := 0
	for _, site := range sites {
		found[site.key] = true
		listing, listed := unionLayoutListed[site.key]
		if site.dominated {
			dominated++
			t.Logf("%s (%s): kind-tested on every path", site.key, site.at)
			if listed {
				t.Errorf("%s (%s) is kind-tested on every path now, so unionLayoutListed's entry for it is stale: remove it", site.key, site.at)
			}
			continue
		}
		t.Logf("%s (%s): kind-tested %s", site.key, site.at, unionLayoutTested(site.tested))
		if !listed {
			t.Errorf("%s (%s) casts or reads a layout off a Union, kind-tested %s: %q. Test the heap's kind before it on every path, or, if lowering proves which arm it is, list it in unionLayoutListed with tested %q and the proof", site.key, site.at, unionLayoutTested(site.tested), site.literal, site.tested)
			continue
		}
		if listing.tested != site.tested {
			t.Errorf("%s (%s) is listed as kind-tested %s, but the walk finds it kind-tested %s: %q", site.key, site.at, unionLayoutTested(listing.tested), unionLayoutTested(site.tested), site.literal)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(unionLayoutListed)) {
		if unionLayoutListed[key].reason == "" {
			t.Errorf("unionLayoutListed lists %s without why", key)
		}
		if !found[key] {
			t.Errorf("unionLayoutListed lists %s, which matches no Union layout site: stale, remove it", key)
		}
	}
	// toBoolean's box reads under its switch on the kind and readUnionViewAt's casts past its snapshot's
	// kind tests are tested today, so a reading that finds fewer has stopped seeing the emitter.
	if dominated < 6 {
		t.Fatalf("read %d kind-tested Union layout sites from the emitter's source; the test's reading of internal/native is out of date", dominated)
	}
}

// unionLayoutTested says when a kind test comes before a site, for a failure's message.
func unionLayoutTested(tested string) string {
	if tested == "" {
		return "on no path"
	}
	return "when " + tested
}

// unionLayoutSites is every Union layout site in the emitter's source, keyed as unionLayoutListed is.
func unionLayoutSites(t *testing.T) []unionLayoutSite {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	var sites []unionLayoutSite
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			sites = append(sites, unionLayoutFunctionSites(t, fileSet, path, function)...)
		}
	}
	slices.SortStableFunc(sites, func(left, right unionLayoutSite) int { return cmp.Compare(left.position, right.position) })
	ordinals := map[string]int{}
	for index := range sites {
		prefix := sites[index].file + ":" + sites[index].function
		ordinals[prefix]++
		sites[index].key = prefix + ":" + strconv.Itoa(ordinals[prefix])
	}
	return sites
}

// unionLayoutFunctionSites is function's Union layout sites, each with whether a kind test comes before
// it on every path, and if not, when one does.
func unionLayoutFunctionSites(t *testing.T, fileSet *token.FileSet, path string, function *ast.FuncDecl) []unionLayoutSite {
	t.Helper()
	var literals []unionLayoutLiteral
	var names []unionLayoutName
	var stack []ast.Node
	ast.Inspect(function, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		switch node := node.(type) {
		case *ast.BasicLit:
			if node.Kind == token.STRING {
				text, err := strconv.Unquote(node.Value)
				if err != nil {
					t.Fatalf("%s: %v", fileSet.Position(node.Pos()), err)
				}
				literals = append(literals, unionLayoutLiteral{literal: node, text: text, ancestors: slices.Clone(stack)})
			}
		case *ast.AssignStmt:
			if len(node.Lhs) != 0 && len(node.Rhs) == 1 && unionLayoutMakesUnion(node.Rhs[0]) {
				if name, isName := node.Lhs[0].(*ast.Ident); isName && name.Name != "_" {
					names = append(names, unionLayoutName{name: name.Name, scope: unionLayoutScope(stack), after: node.End()})
				}
			}
		}
		stack = append(stack, node)
		return true
	})
	narrows := unionLayoutTakesNarrow(function)
	var kindTests []unionLayoutLiteral
	for _, literal := range literals {
		if unionLayoutKindTest.MatchString(literal.text) {
			kindTests = append(kindTests, literal)
		}
	}
	var sites []unionLayoutSite
	for _, literal := range literals {
		call := unionLayoutCall(literal.ancestors)
		at := unionLayoutCastAt(literal.text, call)
		if at < 0 || (!narrows && !unionLayoutScoped(literal, call, names)) {
			continue
		}
		dominated, tested := unionLayoutDominance(literal, at, kindTests)
		position := fileSet.Position(literal.literal.Pos())
		sites = append(sites, unionLayoutSite{
			file:      path,
			function:  rawStoreFunctionName(function),
			position:  literal.literal.Pos(),
			at:        position.Filename + ":" + strconv.Itoa(position.Line),
			literal:   literal.text,
			tested:    tested,
			dominated: dominated,
		})
	}
	return sites
}

// unionLayoutCastAt is where text casts to a heap layout, or -1 where it doesn't.
func unionLayoutCastAt(text string, call *ast.CallExpr) int {
	if at := unionLayoutCasts.FindStringIndex(text); at != nil {
		return at[0]
	}
	if call != nil && slices.ContainsFunc(call.Args, unionLayoutIsCType) {
		if at := unionLayoutTypedCast.FindStringIndex(text); at != nil {
			return at[0]
		}
	}
	return -1
}

// unionLayoutScoped says literal's operand is a Union: it's in the body of a branch that tests a type
// `== ir.Union`, or in a case of ir.Union or ir.Narrow, or its call is handed ir.Union or names a
// variable made a Union above it.
func unionLayoutScoped(literal unionLayoutLiteral, call *ast.CallExpr, names []unionLayoutName) bool {
	for index, ancestor := range literal.ancestors {
		switch ancestor := ancestor.(type) {
		case *ast.IfStmt:
			if index+1 < len(literal.ancestors) && literal.ancestors[index+1] == ancestor.Body && unionLayoutTestsUnion(ancestor.Cond) {
				return true
			}
		case *ast.CaseClause:
			for _, expression := range ancestor.List {
				if unionLayoutIsIR(expression, "Union") || unionLayoutIsIR(expression, "Narrow") || unionLayoutTestsUnion(expression) {
					return true
				}
			}
		}
	}
	if call == nil {
		return false
	}
	for _, argument := range call.Args {
		if unionLayoutIsIR(argument, "Union") || unionLayoutMentions(argument, literal, names) {
			return true
		}
	}
	return false
}

// unionLayoutDominance says whether a kind test comes before literal's cast, at at, on every path, and
// if not, the conditions under which one does, empty when none does.
func unionLayoutDominance(literal unionLayoutLiteral, at int, kindTests []unionLayoutLiteral) (bool, string) {
	if test := unionLayoutKindTest.FindStringIndex(literal.text); test != nil && test[0] < at {
		return true, ""
	}
	shared := map[ast.Node]bool{}
	for _, ancestor := range literal.ancestors {
		shared[ancestor] = true
	}
	var conditions []string
	for _, test := range kindTests {
		if test.literal.Pos() >= literal.literal.Pos() {
			continue
		}
		guards := unionLayoutGuards(test.ancestors, shared)
		if len(guards) == 0 {
			return true, ""
		}
		conditions = append(conditions, strings.Join(guards, " && "))
	}
	slices.Sort(conditions)
	return false, strings.Join(slices.Compact(conditions), " || ")
}

// unionLayoutGuards is the conditions of the branches a kind test sits in that don't also hold the site,
// whose ancestors are shared.
func unionLayoutGuards(ancestors []ast.Node, shared map[ast.Node]bool) []string {
	var guards []string
	for index, ancestor := range ancestors {
		var branch ast.Node
		if index+1 < len(ancestors) {
			branch = ancestors[index+1]
		}
		switch ancestor := ancestor.(type) {
		case *ast.IfStmt:
			if branch == nil || shared[branch] {
				continue
			}
			if branch == ancestor.Body {
				guards = append(guards, types.ExprString(ancestor.Cond))
			} else if branch == ancestor.Else {
				guards = append(guards, "!("+types.ExprString(ancestor.Cond)+")")
			}
		case *ast.CaseClause:
			if shared[ancestor] {
				continue
			}
			if ancestor.List == nil {
				guards = append(guards, "default")
				continue
			}
			cases := []string{}
			for _, expression := range ancestor.List {
				cases = append(cases, types.ExprString(expression))
			}
			guards = append(guards, "case "+strings.Join(cases, ", "))
		case *ast.ForStmt:
			if branch == ancestor.Body && !shared[branch] {
				guards = append(guards, "a for loop runs")
			}
		case *ast.RangeStmt:
			if branch == ancestor.Body && !shared[branch] {
				guards = append(guards, "range "+types.ExprString(ancestor.X)+" is not empty")
			}
		case *ast.FuncLit:
			if !shared[ancestor] {
				guards = append(guards, "a func literal runs")
			}
		}
	}
	return guards
}

// unionLayoutCall is the call nearest around a literal, or nil.
func unionLayoutCall(ancestors []ast.Node) *ast.CallExpr {
	for index := len(ancestors) - 1; index >= 0; index-- {
		if call, isCall := ancestors[index].(*ast.CallExpr); isCall {
			return call
		}
	}
	return nil
}

// unionLayoutScope is the block or case nearest around a statement, whose rest reads what it declares.
func unionLayoutScope(ancestors []ast.Node) ast.Node {
	for index := len(ancestors) - 1; index >= 0; index-- {
		switch ancestors[index].(type) {
		case *ast.BlockStmt, *ast.CaseClause:
			return ancestors[index]
		}
	}
	return nil
}

// unionLayoutMakesUnion says expression makes a Go string naming a C Union value: e.own or e.snapshot
// of ir.Union, or converted to ir.Union.
func unionLayoutMakesUnion(expression ast.Expr) bool {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return false
	}
	switch function := call.Fun.(type) {
	case *ast.SelectorExpr:
		return (function.Sel.Name == "own" || function.Sel.Name == "snapshot") && len(call.Args) != 0 && unionLayoutIsIR(call.Args[0], "Union")
	case *ast.Ident:
		return function.Name == "converted" && len(call.Args) == 3 && unionLayoutIsIR(call.Args[1], "Union")
	}
	return false
}

// unionLayoutMentions says argument names a variable made a Union before literal, in a block literal
// sits in. A selector's field is no variable, so e.value doesn't name value.
func unionLayoutMentions(argument ast.Expr, literal unionLayoutLiteral, names []unionLayoutName) bool {
	found := false
	var visit func(ast.Node) bool
	visit = func(node ast.Node) bool {
		if found {
			return false
		}
		switch node := node.(type) {
		case *ast.SelectorExpr:
			ast.Inspect(node.X, visit)
			return false
		case *ast.Ident:
			for _, name := range names {
				if name.name == node.Name && name.after < literal.literal.Pos() && slices.Contains(literal.ancestors, name.scope) {
					found = true
				}
			}
		}
		return !found
	}
	ast.Inspect(argument, visit)
	return found
}

// unionLayoutTestsUnion says expression compares something `== ir.Union`.
func unionLayoutTestsUnion(expression ast.Expr) bool {
	found := false
	ast.Inspect(expression, func(node ast.Node) bool {
		if binary, isBinary := node.(*ast.BinaryExpr); isBinary && binary.Op == token.EQL && (unionLayoutIsIR(binary.X, "Union") || unionLayoutIsIR(binary.Y, "Union")) {
			found = true
		}
		return !found
	})
	return found
}

// unionLayoutTakesNarrow says function takes an ir.Narrow, whose Value is a Union by the IR's own word.
func unionLayoutTakesNarrow(function *ast.FuncDecl) bool {
	for _, field := range function.Type.Params.List {
		if unionLayoutIsIR(field.Type, "Narrow") {
			return true
		}
	}
	return false
}

// unionLayoutIsIR says expression is ir.name.
func unionLayoutIsIR(expression ast.Expr, name string) bool {
	selector, isSelector := expression.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != name {
		return false
	}
	qualifier, isIdent := selector.X.(*ast.Ident)
	return isIdent && qualifier.Name == "ir"
}

// unionLayoutIsCType says expression is a call to cType, the C type a cast spells.
func unionLayoutIsCType(expression ast.Expr) bool {
	call, isCall := expression.(*ast.CallExpr)
	if !isCall {
		return false
	}
	function, isIdent := call.Fun.(*ast.Ident)
	return isIdent && function.Name == "cType"
}
