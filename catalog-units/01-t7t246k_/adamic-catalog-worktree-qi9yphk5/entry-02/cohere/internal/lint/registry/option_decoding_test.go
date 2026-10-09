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
 * Every option a rule decodes is decoded strictly, and every option field declares its key.
 *
 * Go's `json.Unmarshal` drops an unknown key without a word and matches a key to its field
 * case-insensitively. Upstream's schemas refuse both, so a config with `ignorePatern` or
 * `enforceForTsTypes` loaded clean and the option did nothing. `rule.UnmarshalOptions` refuses both,
 * and `rule.DecodeOptionsInto` uses it, but a hand-written decoder chose its own call, and 89 of them
 * chose `json.Unmarshal` or a bare `json.Decoder` (#4a4yse4).
 *
 * A fixture cannot see this. A fixture hands the decoder the keys its author spelled correctly. So
 * this reads the rules tree with type information instead, and finds two things:
 *
 *   - a decode by `json.Unmarshal` or `(*json.Decoder).Decode` whose target reaches a struct. Only a
 *     struct has keys to get wrong; decoding into a string, a number, a list of them, or
 *     `json.RawMessage` stays as it is. A type with its own `UnmarshalJSON` is not looked inside,
 *     because it decides its own shape, and the decodes in its body are found on their own.
 *   - a struct field reachable from any option decode (those two, `rule.UnmarshalOptions`, and
 *     `rule.DecodeOptionsInto[T]`) with no `json` tag. Only a tag declares a key's spelling, and
 *     `rule.UnmarshalOptions` enforces case only for a field that has one.
 *
 * One shape is out of its sight: a decode whose target is typed `any`, inside a helper that takes the
 * target as a parameter. It reaches no struct as far as types can tell. There is none in the rules
 * tree now (boundaries had `strictUnmarshal`, which became `rule.UnmarshalOptions` calls), and a
 * helper like it should take the decode it means rather than wrap `json.Decoder` again.
 *
 * The tag's spelling is upstream's, checked against each installed plugin's `meta.schema` when the
 * tags were added. That check needs node_modules and so lives in the task, not here; what this
 * guards is that no field goes without a declared spelling again.
 */

// lenientOptionDecodes are the decodes whose upstream schema does not close the object, keyed by
// file and enclosing function. ESLint accepts any extra key there, so refusing one would refuse a
// config upstream loads. Each must say so in a comment at the site.
var lenientOptionDecodes = map[string]string{
	"react/forbid_dom_props.go:DecodeForbidDomPropsOptions":              "a forbid entry object sets no additionalProperties",
	"react/forbid_prop_types.go:DecodeForbidPropTypesOptions":            "the options object sets additionalProperties: true",
	"react/jsx_no_useless_fragment.go:DecodeJsxNoUselessFragmentOptions": "the options object sets no additionalProperties",
	"react/no_danger.go:DecodeNoDangerOptions":                           "the options object sets no additionalProperties",
	"react/no_unescaped_entities.go:decodeForbiddenEntity":               "a forbid entry object sets no additionalProperties",
	"react/style_prop_object.go:DecodeStylePropObjectOptions":            "the options object sets no additionalProperties",
}

// optionDecode is one call that decodes option JSON into a Go value.
type optionDecode struct {
	site     string // rules-relative file and enclosing function, the key lenientOptionDecodes uses
	position string
	call     string // Unmarshal, Decode, UnmarshalOptions or DecodeOptionsInto
	target   types.Type
}

// optionDecodingFindings is what one walk found: the decodes, the ones that are not strict, and the
// fields that declare no key.
type optionDecodingFindings struct {
	decodes   []optionDecode
	loose     []optionDecode
	untagged  []string
	lenientAt map[string]bool
}

// findOptionDecodes loads the packages matching patterns from moduleRoot and walks every decode in
// their non-test files.
func findOptionDecodes(t *testing.T, moduleRoot string, rulesRoot string, patterns ...string) optionDecodingFindings {
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

	findings := optionDecodingFindings{lenientAt: map[string]bool{}}
	untagged := map[string]bool{}
	for _, loadedPackage := range loaded {
		if strings.Contains(loadedPackage.PkgPath, "/tools/") {
			// Generators run by hand, which decode upstream's data files, not a rule's options.
			continue
		}
		for _, file := range loadedPackage.Syntax {
			fileName := loadedPackage.Fset.Position(file.Pos()).Filename
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			relative, err := filepath.Rel(rulesRoot, fileName)
			if err != nil {
				t.Fatalf("relating %s to %s: %v", fileName, rulesRoot, err)
			}
			for _, declaration := range file.Decls {
				function, isFunction := declaration.(*ast.FuncDecl)
				if !isFunction || function.Body == nil {
					continue
				}
				site := filepath.ToSlash(relative) + ":" + function.Name.Name
				ast.Inspect(function.Body, func(node ast.Node) bool {
					call, isCall := node.(*ast.CallExpr)
					if !isCall {
						return true
					}
					name, target := optionDecodeTarget(loadedPackage.TypesInfo, call)
					if target == nil {
						return true
					}
					position := loadedPackage.Fset.Position(call.Pos())
					decode := optionDecode{
						site:     site,
						position: fmt.Sprintf("%s:%d", filepath.ToSlash(relative), position.Line),
						call:     name,
						target:   target,
					}
					findings.decodes = append(findings.decodes, decode)
					reachesStruct := walkOptionType(target, loadedPackage.Fset, rulesRoot, untagged, map[types.Type]bool{})
					if reachesStruct && (name == "Unmarshal" || name == "Decode") {
						if _, lenient := lenientOptionDecodes[site]; lenient {
							findings.lenientAt[site] = true
						} else {
							findings.loose = append(findings.loose, decode)
						}
					}
					return true
				})
			}
		}
	}
	for field := range untagged {
		findings.untagged = append(findings.untagged, field)
	}
	sort.Strings(findings.untagged)
	return findings
}

// optionDecodeTarget names the decode a call makes and the type it decodes into, or returns a nil
// type for a call that is not one.
func optionDecodeTarget(info *types.Info, call *ast.CallExpr) (string, types.Type) {
	switch function := call.Fun.(type) {
	case *ast.SelectorExpr:
		object := info.Uses[function.Sel]
		if object == nil || object.Pkg() == nil {
			return "", nil
		}
		packagePath := object.Pkg().Path()
		switch {
		case packagePath == "encoding/json" && object.Name() == "Unmarshal" && len(call.Args) == 2:
			return "Unmarshal", info.TypeOf(call.Args[1])
		case packagePath == "encoding/json" && object.Name() == "Decode" && len(call.Args) == 1:
			return "Decode", info.TypeOf(call.Args[0])
		case strings.HasSuffix(packagePath, "/lint/rule") && object.Name() == "UnmarshalOptions" && len(call.Args) == 2:
			return "UnmarshalOptions", info.TypeOf(call.Args[1])
		}
	case *ast.IndexExpr:
		selector, isSelector := function.X.(*ast.SelectorExpr)
		if !isSelector {
			return "", nil
		}
		object := info.Uses[selector.Sel]
		if object != nil && object.Pkg() != nil && strings.HasSuffix(object.Pkg().Path(), "/lint/rule") && object.Name() == "DecodeOptionsInto" {
			return "DecodeOptionsInto", info.TypeOf(function.Index)
		}
	}
	return "", nil
}

// walkOptionType reports whether a decode into valueType reaches a struct, and records every field
// it reaches that declares no `json` key, as `file:line Type.Field`.
func walkOptionType(valueType types.Type, fileSet *token.FileSet, rulesRoot string, untagged map[string]bool, seen map[types.Type]bool) bool {
	valueType = types.Unalias(valueType)
	if seen[valueType] {
		return false
	}
	seen[valueType] = true
	switch shape := valueType.(type) {
	case *types.Pointer:
		return walkOptionType(shape.Elem(), fileSet, rulesRoot, untagged, seen)
	case *types.Slice:
		return walkOptionType(shape.Elem(), fileSet, rulesRoot, untagged, seen)
	case *types.Array:
		return walkOptionType(shape.Elem(), fileSet, rulesRoot, untagged, seen)
	case *types.Map:
		return walkOptionType(shape.Elem(), fileSet, rulesRoot, untagged, seen)
	case *types.Named:
		if unmarshalsItself(shape) {
			return false
		}
		structure, isStruct := shape.Underlying().(*types.Struct)
		if !isStruct {
			return walkOptionType(shape.Underlying(), fileSet, rulesRoot, untagged, seen)
		}
		walkOptionStruct(structure, shape.Obj().Name(), fileSet, rulesRoot, untagged, seen)
		return true
	case *types.Struct:
		walkOptionStruct(shape, "struct", fileSet, rulesRoot, untagged, seen)
		return true
	}
	return false
}

func walkOptionStruct(structure *types.Struct, owner string, fileSet *token.FileSet, rulesRoot string, untagged map[string]bool, seen map[types.Type]bool) {
	for index := range structure.NumFields() {
		field := structure.Field(index)
		tag := reflect.StructTag(structure.Tag(index)).Get("json")
		if tag == "-" {
			continue
		}
		key, _, _ := strings.Cut(tag, ",")
		if field.Anonymous() && key == "" {
			// encoding/json promotes an untagged embedded struct's fields, so they are checked as
			// this struct's own.
			walkOptionType(field.Type(), fileSet, rulesRoot, untagged, seen)
			continue
		}
		if !field.Exported() {
			continue
		}
		if key == "" {
			position := fileSet.Position(field.Pos())
			relative, err := filepath.Rel(rulesRoot, position.Filename)
			if err != nil {
				relative = position.Filename
			}
			untagged[fmt.Sprintf("%s:%d %s.%s", filepath.ToSlash(relative), position.Line, owner, field.Name())] = true
		}
		walkOptionType(field.Type(), fileSet, rulesRoot, untagged, seen)
	}
}

// unmarshalsItself reports whether a type has its own UnmarshalJSON, on the value or the pointer.
func unmarshalsItself(named *types.Named) bool {
	for _, receiver := range []types.Type{named, types.NewPointer(named)} {
		methods := types.NewMethodSet(receiver)
		for index := range methods.Len() {
			if methods.At(index).Obj().Name() == "UnmarshalJSON" {
				return true
			}
		}
	}
	return false
}

// TestEveryOptionDecodeIsStrictAndEveryOptionFieldDeclaresItsKey is the guard over the rules tree.
func TestEveryOptionDecodeIsStrictAndEveryOptionFieldDeclaresItsKey(t *testing.T) {
	t.Parallel()
	rulesRoot, err := filepath.Abs("../rules")
	if err != nil {
		t.Fatal(err)
	}
	findings := findOptionDecodes(t, "../../..", rulesRoot, "./internal/lint/rules/...")

	for _, decode := range findings.loose {
		t.Errorf("%s (%s) decodes options into %s with json.%s, which drops an unknown key and matches "+
			"case-insensitively. Use rule.UnmarshalOptions, or, if upstream's schema leaves this object "+
			"open, name %s in lenientOptionDecodes and say so at the site",
			decode.position, decode.site, types.TypeString(decode.target, shortPackageName), decode.call, decode.site)
	}
	for _, field := range findings.untagged {
		t.Errorf("%s is an option field with no json tag, so no spelling is declared for it and its key "+
			"matches in any case. Tag it with upstream's exact key", field)
	}
	for site := range lenientOptionDecodes {
		if !findings.lenientAt[site] {
			t.Errorf("lenientOptionDecodes names %s, which no longer makes a loose decode; remove the entry", site)
		}
	}

	// The floors are the control against a walk that loaded nothing. Measured when this landed: 258
	// decodes in all, 187 of them through rule.DecodeOptionsInto or rule.UnmarshalOptions, the rest
	// into scalars, lists of them, `json.RawMessage`, or the six lenient sites. A count far below
	// either means the pattern matched the wrong tree or the call matcher stopped matching.
	strict := 0
	for _, decode := range findings.decodes {
		if decode.call == "UnmarshalOptions" || decode.call == "DecodeOptionsInto" {
			strict++
		}
	}
	t.Logf("%d option decodes, %d through rule.UnmarshalOptions or rule.DecodeOptionsInto, %d lenient by upstream's schema",
		len(findings.decodes), strict, len(findings.lenientAt))
	if len(findings.decodes) < 200 || strict < 150 {
		t.Fatalf("found %d option decodes, %d of them strict, so this walk examined too little to mean anything",
			len(findings.decodes), strict)
	}
}

// TestTheOptionDecodingGuardSeesALooseDecodeAndAnUntaggedField runs the same walk over a planted
// package, so a clean result above is a measurement rather than a matcher that never matches.
func TestTheOptionDecodingGuardSeesALooseDecodeAndAnUntaggedField(t *testing.T) {
	t.Parallel()
	plantedRoot, err := filepath.Abs("testdata/optiondecoding")
	if err != nil {
		t.Fatal(err)
	}
	findings := findOptionDecodes(t, "../../..", plantedRoot, "./internal/lint/registry/testdata/optiondecoding")

	loose := map[string]bool{}
	for _, decode := range findings.loose {
		loose[decode.site] = true
	}
	want := map[string]bool{
		"planted.go:decodeLoose":          true,
		"planted.go:decodeLooseNested":    true,
		"planted.go:decodeWithDecoder":    true,
		"planted.go:decodeLooseAnonymous": true,
	}
	for site := range want {
		if !loose[site] {
			t.Errorf("the planted loose decode in %s was not found", site)
		}
	}
	for site := range loose {
		if !want[site] {
			t.Errorf("%s was reported loose, and the planted package decodes there strictly or into no struct", site)
		}
	}

	untagged := map[string]bool{}
	for _, finding := range findings.untagged {
		_, field, _ := strings.Cut(finding, " ")
		untagged[field] = true
	}
	for _, field := range []string{"plantedOptions.Untagged", "plantedNested.Deep", "plantedStrictOptions.AlsoUntagged"} {
		if !untagged[field] {
			t.Errorf("the planted untagged field %s was not found; found %v", field, findings.untagged)
		}
	}
	if len(untagged) != 3 {
		t.Errorf("want exactly the three planted untagged fields, and a tagged, skipped, unexported, promoted "+
			"or self-decoding one was reported too: %v", findings.untagged)
	}
}

func shortPackageName(typesPackage *types.Package) string {
	return typesPackage.Name()
}
