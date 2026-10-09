package registry

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

/*
 * Every option field a rule decodes is read by the rule.
 *
 * A field that is decoded and never read is an option a config can set and that changes nothing. The
 * config loads clean, the decoder accepts the key, and the rule behaves exactly as it would without
 * it, which is silent in the expensive direction: whoever set it believes it holds. no-unused-vars
 * shipped four of them (`ignoreRestSiblings`, fixed in #6qxs15e, then `vars`, `caughtErrors` and
 * `destructuredArrayIgnorePattern`, #6esg2nx), and prefer-nullish-coalescing's doc claimed checks its
 * code never made (#10kqgs6). Strict decoding cannot see it, because the key is declared; a fixture
 * cannot see it, because nobody writes a fixture for an option they believe is wired.
 *
 * So this reads the rules tree with type information. It takes every struct field reachable from an
 * option decode, the same decodes option_decoding_test.go finds, and looks in every non-test file
 * for a selector that reads it. Two uses are not reads:
 *
 *   - the left side of a plain assignment, which is where a default is written;
 *   - the condition of an `if` whose body assigns that same field, which is the defaulting idiom
 *     (`if settings.Vars == "" { settings.Vars = "all" }`) and reads the field only to fill it.
 *
 * What it cannot see: a field read only to validate it, in a decoder that refuses a bad value and
 * then never consults the good one. That read is a real read as far as types can tell. And a field
 * reached only through reflection, of which the rules tree has none.
 *
 * A field that is decoded on purpose and never read (upstream declares it, and the decision is that
 * cohere's behavior already matches every value it can take) is named in unreadOptionFields with
 * that decision, so the gap is a recorded choice rather than an accident.
 */

// unreadOptionFields are option fields decoded and deliberately never read, keyed `Type.Field`, each
// with the reason. An entry no longer matching an unread field fails the test, so the list cannot rot.
var unreadOptionFields = map[string]string{
	// Upstream's option, accepted so a config writing it loads. ESLint's two directive checks agree on
	// every input, so the option changes nothing there either; see the field's comment.
	"NoUnusedExpressionsOptions.IgnoreDirectives": "inert upstream too: the structural directive check runs unconditionally",
	// Upstream's rules-of-hooks schema declares it and its create never reads it, only the shared
	// settings key cohere does not carry (#d21war2).
	"RulesOfHooksOptions.AdditionalHooks": "inert upstream too: the rule reads settings, never this option",
	// Captures matter only to a selector or a message template that reads them, and this port refuses
	// both, so no value of it can change a finding.
	"elementDescriptor.Capture": "read only by selector and template features this port refuses",
	// The abbreviation vocabulary is a data file both engines read. These fields are its prose for
	// whoever edits it, the why behind an entry, and no finding is built from them.
	"abbreviationVocabularyFile.About":  "the vocabulary file's own description, for its editors",
	"abbreviationEntry.Reason":          "why an entry exists, for the vocabulary's editors",
	"allowedAbbreviatedName.Reason":     "why a name is allowed, for the vocabulary's editors",
	"allowedAbbreviationSegment.Reason": "why a segment is allowed, for the vocabulary's editors",
}

// optionField is one struct field reachable from an option decode.
type optionField struct {
	key      string // Type.Field
	position string // rules-relative file and line of the field's declaration
}

// optionReadFindings is what one walk found: every reachable field, and the ones nothing reads.
type optionReadFindings struct {
	fields []optionField
	unread []optionField
}

// findUnreadOptionFields loads the packages matching patterns and reports each option field no
// non-test file reads.
func findUnreadOptionFields(t *testing.T, moduleRoot string, rulesRoot string, patterns ...string) optionReadFindings {
	t.Helper()
	config := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  moduleRoot,
	}
	loaded, err := packages.Load(config, patterns...)
	if err != nil {
		t.Fatalf("loading %v: %v", patterns, err)
	}
	if count := packages.PrintErrors(loaded); count > 0 {
		t.Fatalf("%d errors loading %v, so the walk would read a partial tree", count, patterns)
	}

	// Every field an option decode reaches, by its declaring object.
	fields := map[*types.Var]optionField{}
	nonTestFiles := func(loadedPackage *packages.Package) []*ast.File {
		files := []*ast.File{}
		for _, file := range loadedPackage.Syntax {
			if !strings.HasSuffix(loadedPackage.Fset.Position(file.Pos()).Filename, "_test.go") {
				files = append(files, file)
			}
		}
		return files
	}
	for _, loadedPackage := range loaded {
		if strings.Contains(loadedPackage.PkgPath, "/tools/") {
			continue
		}
		for _, file := range nonTestFiles(loadedPackage) {
			ast.Inspect(file, func(node ast.Node) bool {
				call, isCall := node.(*ast.CallExpr)
				if !isCall {
					return true
				}
				_, target := optionDecodeTarget(loadedPackage.TypesInfo, call)
				if target != nil {
					collectOptionFields(target, loadedPackage.Fset, rulesRoot, fields, map[types.Type]bool{})
				}
				return true
			})
		}
	}

	read := map[*types.Var]bool{}
	for _, loadedPackage := range loaded {
		for _, file := range nonTestFiles(loadedPackage) {
			markOptionReads(file, loadedPackage.TypesInfo, fields, read)
		}
	}

	findings := optionReadFindings{}
	for field, described := range fields {
		findings.fields = append(findings.fields, described)
		if !read[field] {
			findings.unread = append(findings.unread, described)
		}
	}
	sort.Slice(findings.fields, func(first, second int) bool { return findings.fields[first].key < findings.fields[second].key })
	sort.Slice(findings.unread, func(first, second int) bool { return findings.unread[first].key < findings.unread[second].key })
	return findings
}

// collectOptionFields records every exported field a decode into valueType can set, through
// pointers, lists, maps, nested structs and promoted embedded structs. A type with its own
// UnmarshalJSON is still entered: its decoder sets its fields, and the rule must read them all the
// same.
func collectOptionFields(valueType types.Type, fileSet *token.FileSet, rulesRoot string, fields map[*types.Var]optionField, seen map[types.Type]bool) {
	valueType = types.Unalias(valueType)
	if seen[valueType] {
		return
	}
	seen[valueType] = true
	switch shape := valueType.(type) {
	case *types.Pointer:
		collectOptionFields(shape.Elem(), fileSet, rulesRoot, fields, seen)
	case *types.Slice:
		collectOptionFields(shape.Elem(), fileSet, rulesRoot, fields, seen)
	case *types.Array:
		collectOptionFields(shape.Elem(), fileSet, rulesRoot, fields, seen)
	case *types.Map:
		collectOptionFields(shape.Elem(), fileSet, rulesRoot, fields, seen)
	case *types.Named:
		structure, isStruct := shape.Underlying().(*types.Struct)
		if !isStruct {
			collectOptionFields(shape.Underlying(), fileSet, rulesRoot, fields, seen)
			return
		}
		collectOptionStructFields(structure, shape.Obj().Name(), fileSet, rulesRoot, fields, seen)
	case *types.Struct:
		collectOptionStructFields(shape, "struct", fileSet, rulesRoot, fields, seen)
	}
}

func collectOptionStructFields(structure *types.Struct, owner string, fileSet *token.FileSet, rulesRoot string, fields map[*types.Var]optionField, seen map[types.Type]bool) {
	for index := range structure.NumFields() {
		field := structure.Field(index)
		tag := reflect.StructTag(structure.Tag(index)).Get("json")
		if tag == "-" {
			continue
		}
		key, _, _ := strings.Cut(tag, ",")
		if field.Anonymous() && key == "" {
			// encoding/json promotes an untagged embedded struct's fields, so each is checked on its
			// own and the embedding itself is not a field a config sets.
			collectOptionFields(field.Type(), fileSet, rulesRoot, fields, seen)
			continue
		}
		if !field.Exported() {
			continue
		}
		position := fileSet.Position(field.Pos())
		relative, err := filepath.Rel(rulesRoot, position.Filename)
		if err != nil {
			relative = position.Filename
		}
		fields[field.Origin()] = optionField{
			key:      owner + "." + field.Name(),
			position: fmt.Sprintf("%s:%d", filepath.ToSlash(relative), position.Line),
		}
		collectOptionFields(field.Type(), fileSet, rulesRoot, fields, seen)
	}
}

// markOptionReads walks one file and marks each option field it reads.
func markOptionReads(file *ast.File, info *types.Info, fields map[*types.Var]optionField, read map[*types.Var]bool) {
	// The selectors that are not reads: plain assignment targets, and the conditions of a defaulting if.
	notReads := map[*ast.SelectorExpr]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		switch statement := node.(type) {
		case *ast.AssignStmt:
			if statement.Tok == token.ASSIGN {
				for _, target := range statement.Lhs {
					if selector, isSelector := ast.Unparen(target).(*ast.SelectorExpr); isSelector {
						notReads[selector] = true
					}
				}
			}
		case *ast.IfStmt:
			assigned := map[*types.Var]bool{}
			for _, inner := range statement.Body.List {
				assignment, isAssignment := inner.(*ast.AssignStmt)
				if !isAssignment || assignment.Tok != token.ASSIGN {
					continue
				}
				for _, target := range assignment.Lhs {
					if field := selectedField(info, target); field != nil {
						assigned[field] = true
					}
				}
			}
			if len(assigned) == 0 {
				return true
			}
			ast.Inspect(statement.Cond, func(conditionNode ast.Node) bool {
				if selector, isSelector := conditionNode.(*ast.SelectorExpr); isSelector {
					if field := selectedField(info, selector); field != nil && assigned[field] {
						notReads[selector] = true
					}
				}
				return true
			})
		}
		return true
	})

	ast.Inspect(file, func(node ast.Node) bool {
		selector, isSelector := node.(*ast.SelectorExpr)
		if !isSelector || notReads[selector] {
			return true
		}
		if field := selectedField(info, selector); field != nil {
			if _, isOption := fields[field]; isOption {
				read[field] = true
			}
		}
		return true
	})
}

// selectedField is the field a selector expression selects, at its generic origin, or nil.
func selectedField(info *types.Info, expression ast.Expr) *types.Var {
	selector, isSelector := ast.Unparen(expression).(*ast.SelectorExpr)
	if !isSelector {
		return nil
	}
	selection := info.Selections[selector]
	if selection == nil || selection.Kind() != types.FieldVal {
		return nil
	}
	field, isField := selection.Obj().(*types.Var)
	if !isField {
		return nil
	}
	return field.Origin()
}

// TestEveryOptionFieldARuleDecodesIsRead is the guard over the rules tree.
func TestEveryOptionFieldARuleDecodesIsRead(t *testing.T) {
	t.Parallel()
	rulesRoot, err := filepath.Abs("../rules")
	if err != nil {
		t.Fatal(err)
	}
	findings := findUnreadOptionFields(t, "../../..", rulesRoot, "./internal/lint/rules/...")

	unread := map[string]bool{}
	for _, field := range findings.unread {
		unread[field.key] = true
		if _, decided := unreadOptionFields[field.key]; decided {
			continue
		}
		t.Errorf("%s %s is decoded from a config and nothing reads it, so setting it changes nothing. Port "+
			"what upstream does with it, or, if every value it can take already behaves as upstream does, "+
			"name it in unreadOptionFields with that decision", field.position, field.key)
	}
	for key := range unreadOptionFields {
		if !unread[key] {
			t.Errorf("unreadOptionFields names %s, which is read now or no longer decoded; remove the entry", key)
		}
	}

	// The floor is the control against a walk that loaded nothing or a matcher that stopped matching.
	t.Logf("%d option fields reachable from a decode, %d unread", len(findings.fields), len(findings.unread))
	if len(findings.fields) < 400 {
		t.Fatalf("found only %d option fields, so this walk examined too little to mean anything", len(findings.fields))
	}
}

// TestTheOptionReadGuardSeesAnUnreadField runs the same walk over a planted package, so a clean result
// above is a measurement rather than a matcher that never matches.
func TestTheOptionReadGuardSeesAnUnreadField(t *testing.T) {
	t.Parallel()
	plantedRoot, err := filepath.Abs("testdata/optionreads")
	if err != nil {
		t.Fatal(err)
	}
	findings := findUnreadOptionFields(t, "../../..", plantedRoot, "./internal/lint/registry/testdata/optionreads")

	unread := map[string]bool{}
	for _, field := range findings.unread {
		unread[field.key] = true
	}
	want := map[string]bool{
		"plantedOptions.NeverMentioned":  true,
		"plantedOptions.OnlyDefaulted":   true,
		"plantedOptions.OnlyAssigned":    true,
		"plantedNested.DeepUnread":       true,
		"plantedEmbedded.PromotedUnread": true,
		"plantedListEntry.EntryUnread":   true,
	}
	for key := range want {
		if !unread[key] {
			t.Errorf("the planted unread field %s was not found; unread: %v", key, findings.unread)
		}
	}
	for key := range unread {
		if !want[key] {
			t.Errorf("%s was reported unread, and the planted package reads it", key)
		}
	}
	// Control: the read ones are in the walk at all, so their silence above is a read and not an absence.
	seen := map[string]bool{}
	for _, field := range findings.fields {
		seen[field.key] = true
	}
	for _, key := range []string{"plantedOptions.Read", "plantedOptions.DefaultedThenRead", "plantedNested.DeepRead", "plantedEmbedded.PromotedRead", "plantedListEntry.EntryRead", "plantedSelfDecoding.Mode"} {
		if !seen[key] {
			t.Errorf("the planted read field %s was not collected, so its silence proves nothing", key)
		}
	}
}
