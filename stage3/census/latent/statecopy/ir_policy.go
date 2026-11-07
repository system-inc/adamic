package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"
)

type irCopyPolicy uint8

const (
	copiedIR irCopyPolicy = iota // Assignment owns all scalar data, including immutable aggregates.
	deepIR                       // Own mutable descendants; shared identities and empty caches obey irPrivateFields.
	sharedIR                     // Preserve an audited immutable owning pointer's identity.
)

type irFieldPolicy struct {
	Type string
	Copy irCopyPolicy
	Why  string
}

// Matches latentSpecImmutable's scalar rule. Named types are resolved against
// actual declarations so a nested type becoming mutable cannot retain a copied policy.
func irImmutable(t ast.Expr, types map[string]ast.Expr, visiting map[string]bool) bool {
	switch t := t.(type) {
	case *ast.Ident:
		if underlying, ok := types[t.Name]; ok {
			if visiting[t.Name] {
				return false
			}
			visiting[t.Name] = true
			defer delete(visiting, t.Name)
			return irImmutable(underlying, types, visiting)
		}
		switch t.Name {
		case "bool", "string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128", "byte", "rune":
			return true
		}
	case *ast.ArrayType:
		return t.Len != nil && irImmutable(t.Elt, types, visiting)
	case *ast.StructType:
		for _, field := range t.Fields.List {
			if len(field.Names) == 0 {
				return false
			}
			for _, id := range field.Names {
				if !id.IsExported() {
					return false
				}
			}
			if !irImmutable(field.Type, types, visiting) {
				return false
			}
		}
		return true
	}
	return false
}

func (generator *generator) checkIRFields(positions *token.FileSet, policies map[string]irFieldPolicy) {
	for owner, expression := range generator.irTypes {
		structure, ok := expression.(*ast.StructType)
		if !ok || !ast.IsExported(owner) {
			continue
		}
		for _, field := range structure.Fields.List {
			if len(field.Names) == 0 {
				panic(fmt.Sprintf(
					"%s: state copy: IR %s embedded field %s needs an explicit policy",
					positions.Position(field.Pos()),
					owner,
					printed(field.Type),
				))
			}
			for _, id := range field.Names {
				if !id.IsExported() {
					continue
				} // irPrivateFields is checked by loadGenerator.
				key := owner + "." + id.Name
				policy, ok := policies[key]
				fail := func(reason string) {
					panic(fmt.Sprintf("%s: state copy: IR field %s: %s", positions.Position(id.Pos()), key, reason))
				}
				if !ok {
					fail("isn't named in statecopy's irSpecFields; say how a speculative copy treats it")
				}
				if policy.Type != printed(field.Type) {
					fail("type changed from " + policy.Type + " to " + printed(field.Type) + "; review its copy policy")
				}
				if policy.Why == "" {
					fail("copy policy has no reason")
				}
				immutable := irImmutable(field.Type, generator.irTypes, map[string]bool{})
				switch policy.Copy {
				case copiedIR:
					if !immutable {
						fail("copied policy would alias mutable state; use a deep copy or audited shared pointer")
					}
				case deepIR:
					if immutable {
						fail("deep policy names an immutable value; review its copy policy")
					}
				case sharedIR:
					if printed(field.Type) != "*GeneratedType" ||
						irPrivateFields["GeneratedType.identity"] != privateShared {
						fail("shared policy lacks an audited immutable pointer identity")
					}
				default:
					fail("invalid copy policy")
				}
			}
		}
	}
}

// Generate the fast path from the same checked inventory used for every IR
// struct. No runtime field-name list can lag behind this copier.
func (generator *generator) specFunctionCopy(body *strings.Builder) {
	structure, ok := generator.irTypes["Function"].(*ast.StructType)
	if !ok {
		return
	} // Small generator fixtures need not declare Function.
	body.WriteString(
		"func latentSpecCopyFunction(value ir.Function, seen map[any]any) ir.Function { var result ir.Function;",
	)
	for _, field := range structure.Fields.List {
		for _, id := range field.Names {
			if !id.IsExported() {
				continue
			} // Audited private caches start zero.
			expression := "value." + id.Name
			if irSpecFields["Function."+id.Name].Copy == deepIR {
				expression = "latentCopyIR(" + expression + ",seen)"
			}
			fmt.Fprintf(body, "result.%s=%s;", id.Name, expression)
		}
	}
	body.WriteString("return result}\n")
}

// General speculative reflection uses generated field indexes too. Only non-IR
// census-local values fall back to runtime structural classification.
func (generator *generator) specMutableFields(body *strings.Builder) {
	names := []string{}
	for name, expression := range generator.irTypes {
		if _, ok := expression.(*ast.StructType); ok && ast.IsExported(name) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	body.WriteString("func latentSpecIRMutableFields(t reflect.Type) ([]int,bool) {switch t {")
	for _, name := range names {
		fmt.Fprintf(body, "case reflect.TypeFor[ir.%s]():return []int{", name)
		index := 0
		for _, field := range generator.irTypes[name].(*ast.StructType).Fields.List {
			for _, id := range field.Names {
				if !id.IsExported() || irSpecFields[name+"."+id.Name].Copy == deepIR {
					fmt.Fprintf(body, "%d,", index)
				}
				index++
			}
		}
		body.WriteString("},true;")
	}
	body.WriteString("};return nil,false}\n")
}
