package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// stopWhereNode is the why of a stop whose where is a node's: Program.Where of the node it met, which lies in the
// unit holding the node.
const stopWhereNode = "a node's where"

// stopWhereRecorded is the why of a stop whose where lowering recorded as it lowered the node: a checked view
// read's ViewWhere, or an IR use's or check's Where, each Program.Where of that node.
const stopWhereRecorded = "a node's where, recorded as lowering met the node"

// stopWhereOrFile is the why of a function with a stop at a node's where, or at one lowering recorded, but at the
// file alone, or no site, where it kept no node: that stop is unplaced, so a first-link lowering records a Gap and
// the pass's judgment abandoned (TrapJudgmentsAbandoned) until the stop takes its where through stopWhere.
const stopWhereOrFile = "a node's where, but the file alone where lowering kept no node: a Gap there"

// stopWhereFile is the why of a function whose stops name the file alone: each is unplaced, so a first-link
// lowering records a Gap and the pass's judgment abandoned (TrapJudgmentsAbandoned) until the stop takes its
// where through stopWhere.
const stopWhereFile = "the file alone: a Gap"

// stopWhereAllowed is each function a judging pass reaches whose stops name their where without stopWhere, by file
// and name, with how many and why: every other stop it makes takes its where through stopWhere (#nxb3xh1).
var stopWhereAllowed = []struct {
	function string
	stops    int
	why      string
}{
	{"arguments_length.go argumentsRefusal", 3, stopWhereNode},
	{"brand_mint.go brandMintRefusal", 1, stopWhereNode},
	{"callback_widening.go adaptArrayCallback", 1, stopWhereNode},
	{"cast.go castRoadValue", 1, stopWhereNode},
	{"cast_kind_union.go kindUnionCast", 1, stopWhereNode},
	{"cast_proof.go checkedClassCast", 1, stopWhereNode},
	{"cast_stop.go castStop", 3, stopWhereNode},
	{"census_small.go censusOverload", 3, stopWhereNode},
	{"class.go construct", 1, stopWhereNode},
	{"class.go instantiate", 1, stopWhereNode},
	{"class.go useOfThis", 3, stopWhereNode},
	{"class_features.go objectKeys", 1, stopWhereNode},
	{"class_generic_calls.go nominalTypeArguments", 1, stopWhereNode},
	{"class_inheritance.go checkMemberOverrides", 1, stopWhereNode},
	{"class_inheritance.go initializerReads", 2, stopWhereNode},
	{"class_static.go staticClassDeclaration", 1, stopWhereNode},
	{"class_static.go staticInitializationReads", 4, stopWhereNode},
	{"collections.go destructuredField", 1, stopWhereNode},
	{"cycles.go findCycles", 2, stopWhereNode},
	{"cycles.go slotsOf", 6, stopWhereNode},
	{"destructure_view.go refuseIncompatibleDestructuredView", 1, stopWhereNode},
	{"detached_own.go detachedOwnCall", 1, stopWhereNode},
	{"diagnostics.go notYet", 1, stopWhereNode},
	{"element_access_fields.go refuseViewElementReads", 1, stopWhereNode},
	{"entries_provenance.go enumerationDescriptors", 2, stopWhereNode},
	{"enum_flags.go flagWriteRefusal", 1, stopWhereNode},
	{"enumeration_accessor_reach.go checkEnumerationReachViews", 1, stopWhereNode},
	{"enums.go enumFields", 2, stopWhereNode},
	{"enums.go enumSwitch", 2, stopWhereNode},
	{"expression.go unparenthesizedValue", 1, stopWhereNode},
	{"from.go arrayFrom", 1, stopWhereNode},
	{"generic.go instantiateFunctionSignature", 2, stopWhereNode},
	{"generic.go refuseInstantiatedMutation", 1, stopWhereNode},
	{"generic_spread_copy_reads.go checkGenericSpreadCopyReads", 2, stopWhereOrFile},
	{"generic_void_reads.go refuseGenericVoidReads", 1, stopWhereNode},
	{"graph_flow.go graphFlows", 1, stopWhereOrFile},
	{"graph_records.go judgeRecordSlots", 1, stopWhereFile},
	{"graph_types.go graphAdoptionBoundary", 2, stopWhereOrFile},
	{"graph_types.go graphTypes", 3, stopWhereNode},
	{"interface_cast.go interfaceCast", 1, stopWhereNode},
	{"interface_cast.go legacyView", 3, stopWhereNode},
	{"invariance.go literalSetterRefusal", 1, stopWhereNode},
	{"invariance.go refuseElementWidening", 1, stopWhereNode},
	{"invariance.go refuseFieldWidening", 2, stopWhereNode},
	{"invariance.go wideningRefusal", 4, stopWhereNode},
	{"iteration.go planIteration", 1, stopWhereNode},
	{"iteration_consume.go destructureIterator", 1, stopWhereNode},
	{"library_for_in.go forIn", 1, stopWhereNode},
	{"library_function_expressions.go closureExpression", 1, stopWhereNode},
	{"library_json_decode.go jsonDecodeSchema", 1, stopWhereNode},
	{"library_json_decode.go jsonSchemaType", 2, stopWhereNode},
	{"library_json_stringify.go jsonCall", 1, stopWhereNode},
	{"library_method_values.go delayedArrayWritesFit", 1, stopWhereNode},
	{"library_node_fs_file.go nodeFSFile", 1, stopWhereNode},
	{"library_object.go objectCallArguments", 1, stopWhereNode},
	{"library_spread.go librarySpreadRefusal", 1, stopWhereNode},
	{"locals.go variables", 2, stopWhereNode},
	{"loose_slots.go looseCellWrite", 1, stopWhereFile},
	{"loose_slots.go looseEntryStop", 2, stopWhereFile},
	{"moves.go moveRefused", 1, stopWhereNode},
	{"never_refused.go neverExpression", 3, stopWhereNode},
	{"never_refused.go refuseAnyValue", 1, stopWhereNode},
	{"new_class_value.go newClassValue", 1, stopWhereNode},
	{"object.go arrayReduce", 1, stopWhereNode},
	{"object.go arraySort", 1, stopWhereNode},
	{"object.go arrayVisit", 2, stopWhereNode},
	{"object.go elementType", 1, stopWhereNode},
	{"object.go objectLiteral", 2, stopWhereNode},
	{"object.go property", 1, stopWhereNode},
	{"object.go propertyRead", 1, stopWhereNode},
	{"optional_widening.go optionalWideningRefusal", 1, stopWhereNode},
	{"overload_callback_use.go overloadCallbackUse", 1, stopWhereNode},
	{"overload_parse_tree_node.go overloadNodeTestCall", 1, stopWhereNode},
	{"overload_phantom_result.go overloadPhantomResult", 1, stopWhereNode},
	{"parallel.go capture", 2, stopWhereNode},
	{"parallel.go checkParallelMap", 3, stopWhereNode},
	{"parallel.go effect", 1, stopWhereNode},
	{"parallel.go function", 2, stopWhereNode},
	{"placeholder.go placeholderResetRefusal", 1, stopWhereNode},
	{"predicates.go refused", 1, stopWhereNode},
	{"predicates_proof.go predicateFailure", 1, stopWhereOrFile},
	{"predicates_proof.go predicateTagDomainRefusal", 1, stopWhereNode},
	{"program_region.go programAllocation", 1, stopWhereFile},
	{"program_region.go selectProgramMembership", 1, stopWhereFile},
	{"prototype.go objectPrototypeCallArguments", 3, stopWhereNode},
	{"prototype.go prototypeRead", 2, stopWhereNode},
	{"proven_relations.go provenRelation", 1, stopWhereNode},
	{"records.go recordAccess", 1, stopWhereNode},
	{"records.go recordKey", 1, stopWhereNode},
	{"records.go recordLiteral", 2, stopWhereNode},
	{"records_declared_keys.go declaredKeyRefusal", 2, stopWhereNode},
	{"statements.go expressionStatement", 1, stopWhereNode},
	{"statements.go statement", 2, stopWhereNode},
	{"tsgo.go tsgo", 1, stopWhereNode},
	{"unknown.go inProperty", 1, stopWhereNode},
	{"view_arrays.go checkViewArrays", 3, stopWhereOrFile},
	{"view_callable_producers.go callableProducerViewRefusal", 3, stopWhereRecorded},
	{"view_contracts.go strictViewContract", 1, stopWhereNode},
	{"view_lazy.go judgeCheckedViewReads", 8, stopWhereOrFile},
	{"view_lazy.go unemittableCheckedViewRead", 5, stopWhereRecorded},
	{"view_logical_writes.go checkViewLogicalWrites", 1, stopWhereOrFile},
	{"view_widened.go checkViewWidenedReads", 2, stopWhereRecorded},
	{"void_discards.go voidDiscards", 1, stopWhereFile},
	{"worker_environment.go refused", 1, stopWhereNode},
}

// A judging pass's stop names where it lies through stopWhere, which never names the file alone, so a first-link
// lowering places it in a unit and traps that unit, where a stop naming only the file was a Gap that gave up the
// rest of the pass: checkAccessorSpreads and checkRangeUnionReads did on tsc (#nxb3xh1). The judging passes are
// what lower.go hands firstLinkPass with rewrites false, read from its AST, and a pass's reach is every function
// of the package it calls, or hands on as a value, transitively, by name alone, so a name two functions share
// brings both in. The test fails on a NotYet or a Refused made in that reach whose Where isn't a call of
// stopWhere, outside stopWhereAllowed, and on an allowance whose count no longer holds. A pass reaches lowerBody
// through generic instantiation (findCycles' shape), so the reach is nearly the whole package, and the allowance
// holds body lowering's own stops, each at its node, beside the passes' stops that name the file alone still.
// Mutant: checkRangeUnionReads' Where put back to l.result.Source; this goes red.
func TestJudgingPassStopsNameTheirWhereThroughStopWhere(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	type declared struct {
		key  string
		file *ast.File
		body *ast.BlockStmt
	}
	files := token.NewFileSet()
	functions := map[string][]declared{}
	sources := map[string][]byte{}
	imports := map[*ast.File]map[string]bool{}
	var lowerFile *ast.File
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		source, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatal(err)
		}
		file, err := parser.ParseFile(files, entry.Name(), source, 0)
		if err != nil {
			t.Fatal(err)
		}
		sources[entry.Name()] = source
		if entry.Name() == "lower.go" {
			lowerFile = file
		}
		imports[file] = map[string]bool{}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			name := path[strings.LastIndexByte(path, '/')+1:]
			if spec.Name != nil {
				name = spec.Name.Name
			}
			imports[file][name] = true
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			functions[function.Name.Name] = append(functions[function.Name.Name], declared{key: entry.Name() + " " + function.Name.Name, file: file, body: function.Body})
		}
	}
	if lowerFile == nil {
		t.Fatal("no lower.go to read the judging passes from")
	}
	// named is the name an expression gives a function of the package, by an identifier or a selector, or "" for
	// another package's.
	named := func(file *ast.File, expression ast.Expr) string {
		switch expression := expression.(type) {
		case *ast.Ident:
			return expression.Name
		case *ast.SelectorExpr:
			if qualifier, isIdent := expression.X.(*ast.Ident); isIdent && imports[file][qualifier.Name] {
				return ""
			}
			return expression.Sel.Name
		}
		return ""
	}
	reached := map[string]bool{}
	var queue []string
	reach := func(name string) {
		if _, declares := functions[name]; declares && !reached[name] {
			reached[name] = true
			queue = append(queue, name)
		}
	}
	calls := func(file *ast.File, node ast.Node) {
		ast.Inspect(node, func(node ast.Node) bool {
			if call, isCall := node.(*ast.CallExpr); isCall {
				reach(named(file, call.Fun))
				for _, argument := range call.Args {
					reach(named(file, argument))
				}
			}
			return true
		})
	}
	passes := map[string]bool{}
	ast.Inspect(lowerFile, func(node ast.Node) bool {
		call, isCall := node.(*ast.CallExpr)
		if !isCall || named(lowerFile, call.Fun) != "firstLinkPass" || len(call.Args) != 4 {
			return true
		}
		if rewrites, isIdent := call.Args[1].(*ast.Ident); !isIdent || rewrites.Name != "false" {
			return true
		}
		if name, isLiteral := call.Args[0].(*ast.BasicLit); isLiteral {
			passes[strings.Trim(name.Value, `"`)] = true
		}
		if run, isLiteral := call.Args[3].(*ast.FuncLit); isLiteral {
			calls(lowerFile, run.Body)
		} else {
			reach(named(lowerFile, call.Args[3]))
		}
		return true
	})
	for _, pass := range []string{"checkAccessorSpreads", "checkRangeUnionReads", "checkViewWidenedReads"} {
		if !passes[pass] || !reached[pass] {
			t.Fatalf("%s isn't read as a judging pass from lower.go (passes %v), so the walk below proves nothing", pass, slices.Sorted(maps.Keys(passes)))
		}
	}
	for len(queue) > 0 {
		name := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, function := range functions[name] {
			calls(function.file, function.body)
		}
	}
	counted := map[string]int{}
	wheres := map[string][]string{}
	for name := range reached {
		for _, function := range functions[name] {
			ast.Inspect(function.body, func(node ast.Node) bool {
				literal, isLiteral := node.(*ast.CompositeLit)
				if !isLiteral {
					return true
				}
				if kind, isIdent := literal.Type.(*ast.Ident); !isIdent || kind.Name != "NotYet" && kind.Name != "Refused" {
					return true
				}
				where := "no Where"
				for _, element := range literal.Elts {
					field, isField := element.(*ast.KeyValueExpr)
					if !isField {
						continue
					}
					if key, isKey := field.Key.(*ast.Ident); !isKey || key.Name != "Where" {
						continue
					}
					if call, isCall := field.Value.(*ast.CallExpr); isCall && named(function.file, call.Fun) == "stopWhere" {
						return true
					}
					start, end := files.Position(field.Value.Pos()), files.Position(field.Value.End())
					where = string(sources[start.Filename][start.Offset:end.Offset])
				}
				counted[function.key]++
				wheres[function.key] = append(wheres[function.key], where)
				return true
			})
		}
	}
	allowed := map[string]int{}
	for _, allowance := range stopWhereAllowed {
		if _, twice := allowed[allowance.function]; twice || allowance.why == "" {
			t.Errorf("%s is allowed twice, or with no why", allowance.function)
		}
		allowed[allowance.function] = allowance.stops
	}
	var wrong []string
	keys := maps.Clone(allowed)
	maps.Copy(keys, counted)
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		if counted[key] != allowed[key] {
			wrong = append(wrong, key+": "+strconv.Itoa(counted[key])+" stops whose Where isn't stopWhere's ("+strings.Join(wheres[key], "; ")+"), allowed "+strconv.Itoa(allowed[key]))
		}
	}
	if len(wrong) > 0 {
		t.Errorf("a judging pass reaches stops whose Where isn't stopWhere's, which may place in no unit and leave the pass's judgment abandoned; give each its where through l.stopWhere (stop_where.go), or allow it in stopWhereAllowed with why:\n%s", strings.Join(wrong, "\n"))
	}
}
