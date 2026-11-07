// statecopy generates typed copies of lowering-owned mutable state for the overlay.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type generator struct {
	irTypes   map[string]ast.Expr
	types     map[string]ast.Expr
	needed    map[string]bool
	sortedMap *ast.TypeSpec
	positions *token.FileSet
	// The unexported fields of exported IR structs, as Type.field, each one named in irPrivateFields.
	privateIRFields []string
}

// How both latent copies treat each unexported field of an exported IR struct. latentCopyIR copies an IR
// value by reflection, which can't write an unexported field, so every such field is named here with what the
// copy does instead, and loadGenerator refuses one that isn't (#fwbnwn3: methodTables arrived Oct 20 and every
// full-mode unit that copied a Program panicked on it until this table existed).
type privateIRField int

const (
	// privateNil: a cache a census never fills; the copy accepts it only nil, and refuses a populated one.
	privateNil privateIRField = iota
	// privateEmpty: a cache the IR rebuilds from exported fields when it's nil; the copy starts it empty.
	privateEmpty
	// privateShared: an identity compared by pointer; the copy shares the owning pointer, whose value nothing
	// writes after it's built, rather than making a struct it can't give that identity.
	privateShared
)

var irPrivateFields = map[string]privateIRField{
	"Program.argumentFacts":       privateNil,    // Backend-only derivation; a census must not populate it.
	"Program.methodTables":        privateEmpty,  // Rebuilt from the copied method declarations.
	"BoundaryStringTable.indexes": privateEmpty,  // Rebuilt from copied Values.
	"GeneratedType.identity":      privateShared, // Immutable owner identity, compared by pointer.
}

// Only audited program facts are listed. Every unlisted field keeps the normal
// typed unit snapshot; adding a lowering field does not require a carry decision.
// Maps remain owned copies in each unit and statement snapshot.
type loweringFieldPolicy uint8

const programFact loweringFieldPolicy = 1

var loweringFieldPolicies = map[string]loweringFieldPolicy{
	"entriesRecords":             programFact,
	"entriesProgramRecords":      programFact,
	"entriesProgramRecordsKnown": programFact,
}

func printed(n ast.Node) string {
	var b bytes.Buffer
	if nodeError := format.Node(&b, token.NewFileSet(), n); nodeError != nil {
		panic(nodeError)
	}
	return b.String()
}
func (g *generator) copy(t ast.Expr, value string) string {
	switch t := t.(type) {
	case *ast.Ident:
		if _, ok := g.types[t.Name].(*ast.StructType); ok {
			g.needed[t.Name] = true
			return "latentCopy_" + t.Name + "(" + value + ", seen)"
		}
		if underlying, ok := g.types[t.Name]; ok {
			switch underlying.(type) {
			case *ast.ArrayType, *ast.MapType:
				panic("state copy: mutable named container " + t.Name + " requires an explicit copier")
			}
		}
		return value
	case *ast.SelectorExpr:
		if printed(t.X) == "ir" {
			return "latentCopyIR(" + value + ", seen)"
		}
		if printed(t.X) != "ast" && printed(t.X) != "checker" {
			panic("state copy: foreign state type " + printed(t) + " requires an explicit copier")
		}
		return value
	case *ast.StarExpr:
		// The opaque mapper belongs to the checker; copying its empty facade loses identity.
		if printed(t.X) == "typeMapper" {
			return value
		}
		if name, ok := t.X.(*ast.Ident); ok {
			if _, ok := g.types[name.Name].(*ast.StructType); ok {
				g.needed[name.Name] = true
				return "latentCopyPointer_" + name.Name + "(" + value + ", seen)"
			}
		}
		if name, ok := t.X.(*ast.SelectorExpr); ok && printed(name.X) == "ir" {
			return "latentCopyIR(" + value + ", seen)"
		}
		if name, ok := t.X.(*ast.SelectorExpr); ok &&
			(printed(name.X) == "ast" || printed(name.X) == "checker" || printed(name) == "load.Program") {
			return value
		}
		panic("state copy: foreign or unrecognized pointer " + printed(t) + " requires an explicit copier")
	case *ast.ArrayType:
		if t.Len != nil {
			panic("state copy: fixed array " + printed(t) + " requires an explicit copier")
		}
		ty := printed(t)
		return "func(value " +
			ty +
			") " +
			ty +
			" {if value==nil{return nil}; result:=make(" +
			ty +
			",len(value)); for i,item:=range value {result[i]=" +
			g.copy(t.Elt, "item") +
			"}; return result}(" +
			value +
			")"
	case *ast.MapType:
		ty := printed(t)
		return g.copyMap(ty, t.Key, t.Value, value)
	case *ast.IndexListExpr:
		selector, ok := t.X.(*ast.SelectorExpr)
		if ok &&
			printed(selector.X) == "ir" &&
			selector.Sel.Name == "SortedMap" &&
			len(t.Indices) == 2 &&
			g.sortedMap != nil {
			definition, ok := g.sortedMap.Type.(*ast.MapType)
			if !ok || printed(definition.Key) != "K" || printed(definition.Value) != "V" {
				panic("state copy: ir.SortedMap definition is no longer map[K]V")
			}
			return g.copyMap(printed(t), t.Indices[0], t.Indices[1], value)
		}
		panic("state copy: unsupported field type " + printed(t))
	case *ast.InterfaceType:
		panic("state copy: anonymous interface requires an explicit copier")
	default:
		panic("state copy: unsupported field type " + printed(t))
	}
}
func (g *generator) copyField(owner string, id *ast.Ident, field *ast.Field) (result string) {
	defer func() {
		if failure := recover(); failure != nil {
			panic(fmt.Sprintf("%s: %s.%s: %v", g.positions.Position(field.Pos()), owner, id.Name, failure))
		}
	}()
	return g.copy(field.Type, "value."+id.Name)
}
func (g *generator) copyMap(ty string, key, item ast.Expr, value string) string {
	return "func(value " +
		ty +
		") " +
		ty +
		" {if value==nil{return nil}; result:=make(" +
		ty +
		",len(value)); for key,item:=range value {result[" +
		g.copy(key, "key") +
		"]=" +
		g.copy(item, "item") +
		"}; return result}(" +
		value +
		")"
}
func loadGenerator(repository, lowerInput string) generator {
	g := generator{
		irTypes:   map[string]ast.Expr{},
		types:     map[string]ast.Expr{},
		needed:    map[string]bool{"lowering": true},
		positions: token.NewFileSet(),
	}

	irFiles, globError := filepath.Glob(filepath.Join(repository, "internal/ir/*.go"))
	if globError != nil {
		panic(globError)
	}
	irPositions := token.NewFileSet()
	for _, path := range irFiles {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, parseFileError := parser.ParseFile(irPositions, path, nil, 0)
		if parseFileError != nil {
			panic(parseFileError)
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					spec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					g.irTypes[spec.Name.Name] = spec.Type
					if spec.Name.Name == "SortedMap" {
						g.sortedMap = spec
					}
					// An unexported struct is reached only through an unexported field, which this table names.
					structure, isStruct := spec.Type.(*ast.StructType)
					if !isStruct || !spec.Name.IsExported() {
						continue
					}
					for _, field := range structure.Fields.List {
						for _, id := range field.Names {
							if id.IsExported() {
								continue
							}
							key := spec.Name.Name + "." + id.Name
							if _, ok := irPrivateFields[key]; !ok {
								panic(fmt.Sprintf(
									"%s: state copy: unexported IR field %s isn't named in statecopy's irPrivateFields; say how a latent copy treats it",
									irPositions.Position(id.Pos()),
									key,
								))
							}
							g.privateIRFields = append(g.privateIRFields, key)
						}
					}
				}
			}
		}
	}
	g.checkIRFields(irPositions, irSpecFields)
	paths, globError := filepath.Glob(filepath.Join(repository, "internal/lower/*.go"))
	if globError != nil {
		panic(globError)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		input := path
		if filepath.Base(path) == "lower.go" {
			input = lowerInput
		}
		file, parseFileError := parser.ParseFile(g.positions, input, nil, 0)
		if parseFileError != nil {
			panic(parseFileError)
		}
		for _, decl := range file.Decls {
			if group, ok := decl.(*ast.GenDecl); ok {
				for _, spec := range group.Specs {
					if spec, ok := spec.(*ast.TypeSpec); ok {
						g.types[spec.Name.Name] = spec.Type
					}
				}
			}
		}
	}
	return g
}
func generate(repository, lowerInput string) []byte {
	g := loadGenerator(repository, lowerInput)
	if _, ok := g.types["lowering"].(*ast.StructType); !ok {
		panic("state copy: lowering: expected a struct")
	}
	var body strings.Builder
	done := map[string]bool{}
	for {
		names := []string{}
		for name := range g.needed {
			if !done[name] {
				names = append(names, name)
			}
		}
		sort.Strings(names)
		if len(names) == 0 {
			break
		}
		for _, name := range names {
			done[name] = true
			fields := g.types[name].(*ast.StructType).Fields.List
			fmt.Fprintf(&body, "func latentCopy_%s(value %s, seen map[any]any) %s {result:=value;", name, name, name)
			for _, field := range fields {
				if len(field.Names) == 0 {
					panic("state copy: " + name + ": embedded field unsupported")
				}
				for _, id := range field.Names {
					var expr string
					if name == "lowering" &&
						id.Name == "impliedTargetQuery" &&
						printed(field.Type) == "func(*ast.Node) *checker.Type" {
						// Production leaves this test query nil. A closure needs an audited copier.
						fmt.Fprint(
							&body,
							"if value.impliedTargetQuery != nil {panic(\"latent state copy: populated impliedTargetQuery is unsupported\")};",
						)
						expr = "value." + id.Name
					} else if name == "lowering" &&
						id.Name == "sourceIdentities" &&
						printed(field.Type) == "map[*ast.Node]ir.SourceIdentity" {
						// Shared on purpose: a cache of identities that are functions of the program alone, one map for
						// every lowering of a latent run (lower.go.txt's latentRun), so a copy reads and fills the same
						// one.
						expr = "value." + id.Name
					} else {
						expr = g.copyField(name, id, field)
					}
					if expr != "value."+id.Name {
						fmt.Fprintf(&body, "result.%s=%s;", id.Name, expr)
					}
				}
			}
			body.WriteString("return result}\n")
			fmt.Fprintf(
				&body,
				"func latentCopyPointer_%s(value *%s, seen map[any]any) *%s {if value==nil{return nil}; if prior,ok:=seen[value];ok{return prior.(*%s)};result:=new(%s);seen[value]=result; *result=latentCopy_%s(*value,seen);return result}\n",
				name,
				name,
				name,
				name,
				name,
				name,
			)
		}
	}
	// Capture/install only the fields explicitly marked as carried program facts.
	// The ordinary typed snapshot above retains these scalars through rollback.
	facts := []string{}
	for _, field := range g.types["lowering"].(*ast.StructType).Fields.List {
		for _, id := range field.Names {
			if loweringFieldPolicies[id.Name] == programFact {
				if ty := printed(field.Type); ty != "bool" && ty != "map[*ast.Node]bool" {
					panic("state copy: program fact " + id.Name + " needs an audited carry")
				}
				facts = append(facts, id.Name)
			}
		}
	}
	sort.Strings(facts)
	if len(facts) > 0 {
		body.WriteString("type latentProgramFacts struct {")
		for _, name := range facts {
			for _, field := range g.types["lowering"].(*ast.StructType).Fields.List {
				for _, id := range field.Names {
					if id.Name == name {
						fmt.Fprintf(&body, "%s %s;", name, printed(field.Type))
					}
				}
			}
		}
		body.WriteString(
			"}\nfunc latentCaptureProgramFacts(value *lowering) latentProgramFacts { return latentProgramFacts{",
		)
		for _, field := range facts {
			fmt.Fprintf(&body, "%s:value.%s,", field, field)
		}
		body.WriteString("}}\nfunc (value *lowering) latentApplyProgramFacts(facts latentProgramFacts) {")
		for _, field := range facts {
			fmt.Fprintf(&body, "value.%s=facts.%s;", field, field)
		}
		body.WriteString("}\n")
	}
	// Both copies' answers for unexported IR fields, from irPrivateFields, so the table has one home.
	keys := make([]string, 0, len(irPrivateFields))
	for key := range irPrivateFields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	body.WriteString(
		"// latentPrivateIRField is whether the full copy leaves an unexported IR field at its zero value.\n",
	)
	body.WriteString("func latentPrivateIRField(v reflect.Value, index int) bool {field:=v.Type().Field(index).Name;")
	for _, key := range keys {
		owner, field, _ := strings.Cut(key, ".")
		test := "v.Type()==reflect.TypeFor[ir." + owner + "]()&&field==" + fmt.Sprintf("%q", field)
		switch irPrivateFields[key] {
		case privateNil:
			fmt.Fprintf(
				&body,
				"if %s {if !v.Field(index).IsNil(){panic(%q)}; return true};",
				test,
				"latent state copy: populated IR cache "+key+" is unsupported",
			)
		case privateEmpty:
			fmt.Fprintf(&body, "if %s {return true};", test)
		}
	}
	body.WriteString("return false}\n")
	body.WriteString(
		"// latentSharedIRPointer is whether the full copy shares an IR pointer rather than copying what it points to.\n",
	)
	body.WriteString("func latentSharedIRPointer(t reflect.Type) bool {")
	for _, key := range keys {
		if irPrivateFields[key] == privateShared {
			owner, _, _ := strings.Cut(key, ".")
			fmt.Fprintf(&body, "if t==reflect.TypeFor[*ir.%s]() {return true};", owner)
		}
	}
	body.WriteString("return false}\n")
	g.specFunctionCopy(&body)
	g.specMutableFields(&body)
	// Only imports actually mentioned in generated field types are needed.
	var header strings.Builder
	header.WriteString("package lower\nimport(\n\"reflect\"\n")
	for _, pair := range [][2]string{
		{"ast", "github.com/microsoft/TypeScript/tsc/shim/ast"},
		{"checker", "github.com/microsoft/TypeScript/tsc/shim/checker"},
		{"ir", "github.com/system-inc/adamic/internal/ir"},
	} {
		if strings.Contains(body.String(), pair[0]+".") {
			fmt.Fprintf(&header, "%q\n", pair[1])
		}
	}
	header.WriteString(")\n")
	text, formatError := format.Source([]byte(header.String() + body.String()))
	if formatError != nil {
		panic(formatError)
	}
	return text
}
func main() {
	if len(os.Args) != 4 {
		panic("usage: statecopy repository overlaid-lower.go output.go")
	}
	text := generate(os.Args[1], os.Args[2])
	if writeError := os.WriteFile(os.Args[3], text, 0600); writeError != nil {
		panic(writeError)
	}
}
