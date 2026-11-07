package main

import (
	"context"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"sort"
)

func crossingType(c *checker.Checker, subjectType *checker.Type, path string, depth int) (native.ABIType, error) {
	fail := func(reason string) (native.ABIType, error) {
		return native.ABIType{}, fmt.Errorf("%s: %s", path, reason)
	}
	switch subjectType.Flags() {
	case checker.TypeFlagsNumber:
		return native.ABIType{Kind: "number"}, nil
	case checker.TypeFlagsString:
		return native.ABIType{Kind: "string"}, nil
	}
	// boolean is represented by the checker as the union of true and false.
	if subjectType.Flags()&checker.TypeFlagsBoolean != 0 || (subjectType.Flags()&checker.TypeFlagsUnion != 0 &&
		len(subjectType.Types()) == 2 &&
		subjectType.Types()[0].Flags()&checker.TypeFlagsBooleanLiteral != 0 &&
		subjectType.Types()[1].Flags()&checker.TypeFlagsBooleanLiteral != 0) {
		return native.ABIType{Kind: "boolean"}, nil
	}
	if c.IsArrayType(subjectType) {
		symbol := subjectType.Symbol()
		if symbol == nil || symbol.Name != "ReadonlyArray" {
			return fail("ABI requires a readonly array")
		}
		args := c.GetTypeArguments(subjectType)
		if len(args) != 1 {
			return fail("ABI array must have one element type")
		}
		element, crossingTypeError := crossingType(c, args[0], path+"[]", depth)
		if crossingTypeError != nil {
			return native.ABIType{}, crossingTypeError
		}
		if element.Kind != "number" && element.Kind != "string" {
			return fail("ABI arrays contain only number or string")
		}
		return native.ABIType{Kind: element.Kind + "[]"}, nil
	}
	if subjectType.Flags()&checker.TypeFlagsObject == 0 {
		return fail("type is outside ABI version 1 crossing set")
	}
	if depth >= 2 {
		return fail("record nesting exceeds depth 2")
	}
	symbol := subjectType.Symbol()
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindClassDeclaration {
				return fail("class instances are outside ABI version 1")
			}
		}
	}
	if len(c.GetSignaturesOfType(subjectType, checker.SignatureKindCall)) != 0 ||
		len(c.GetSignaturesOfType(subjectType, checker.SignatureKindConstruct)) != 0 {
		return fail("callable objects are outside ABI version 1")
	}
	if len(c.GetIndexInfosOfType(subjectType)) != 0 {
		return fail("index signatures are outside ABI version 1")
	}
	result := native.ABIType{Kind: "record", Fields: []native.ABIField{}}
	for _, field := range c.GetPropertiesOfType(subjectType) {
		fieldPath := path + "." + field.Name
		if field.Flags&ast.SymbolFlagsOptional != 0 {
			return native.ABIType{}, fmt.Errorf("%s: optional fields are outside ABI version 1", fieldPath)
		}
		if !c.IsReadonlySymbol(field) {
			return native.ABIType{}, fmt.Errorf("%s: ABI requires readonly fields", fieldPath)
		}
		for _, declaration := range field.Declarations {
			if declaration.Kind != ast.KindPropertySignature {
				return native.ABIType{}, fmt.Errorf("%s: ABI requires record data fields", fieldPath)
			}
		}
		ft, crossingTypeError := crossingType(c, c.GetTypeOfSymbol(field), fieldPath, depth+1)
		if crossingTypeError != nil {
			return native.ABIType{}, crossingTypeError
		}
		result.Fields = append(result.Fields, native.ABIField{Name: field.Name, Type: ft})
	}
	if len(result.Fields) == 0 {
		return fail("an empty object type accepts any non-nullish value, so it is not a record")
	}
	return result, nil
}

func exportSignatures(program *load.Program, names []string) ([]native.ABIExport, error) {
	return exportSignaturesWith(program, names, false)
}

// Workers may select private entry declarations without changing the public build ABI.
func exportSignaturesWith(program *load.Program, names []string, entryFunctions bool) ([]native.ABIExport, error) {
	entry := program.Files()[0]
	c, release := program.Checker(context.Background(), entry)
	defer release()
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
	}
	exports := []native.ABIExport{}
	module := c.GetSymbolAtLocation(entry.AsNode())
	if module == nil {
		if len(names) > 0 {
			return nil, fmt.Errorf("entry module has no exports")
		}
		return exports, nil
	}
	symbols := c.GetExportsOfModule(module)
	if entryFunctions {
		symbols = nil
		for _, statement := range entry.Statements.Nodes {
			if statement.Kind == ast.KindFunctionDeclaration && statement.Name() != nil {
				symbols = append(symbols, c.GetSymbolAtLocation(statement.Name()))
			}
		}
	}
	for _, symbol := range symbols {
		named := wanted[symbol.Name]
		if len(names) > 0 && !named {
			continue
		}
		if named {
			delete(wanted, symbol.Name)
		}
		if symbol.Flags&ast.SymbolFlagsAlias != 0 ||
			len(symbol.Declarations) != 1 ||
			symbol.Declarations[0].Kind != ast.KindFunctionDeclaration {
			if named {
				return nil, fmt.Errorf(
					"%s: ABI export must be an entry module export function declaration",
					symbol.Name,
				)
			}
			continue
		}
		declaration := symbol.Declarations[0]
		if ast.GetSourceFileOfNode(declaration) != entry {
			if named {
				return nil, fmt.Errorf("%s: ABI export must belong to entry module", symbol.Name)
			}
			continue
		}
		export := native.ABIExport{Name: symbol.Name, Parameters: []native.ABIField{}}
		var invalid error
		if len(declaration.TypeParameters()) > 0 {
			invalid = fmt.Errorf("%s: generic signatures are outside ABI version 1", symbol.Name)
		}
		for _, parameter := range declaration.Parameters() {
			pd := parameter.AsParameterDeclaration()
			if !ast.IsIdentifier(parameter.Name()) ||
				pd.QuestionToken != nil ||
				pd.Initializer != nil ||
				pd.DotDotDotToken != nil {
				invalid = fmt.Errorf("%s: ABI requires required named parameters", symbol.Name)
				break
			}
			abiType, crossingTypeError := crossingType(
				c,
				c.GetTypeAtLocation(parameter),
				symbol.Name+"."+parameter.Name().Text(),
				0,
			)
			if crossingTypeError != nil {
				invalid = crossingTypeError
				break
			}
			export.Parameters = append(export.Parameters, native.ABIField{Name: parameter.Name().Text(), Type: abiType})
		}
		ret := c.GetReturnTypeOfSignature(c.GetSignatureFromDeclaration(declaration))
		if ret.Flags() == checker.TypeFlagsVoid {
			export.Returns = native.ABIType{Kind: "void"}
		} else {
			var crossingTypeError error
			export.Returns, crossingTypeError = crossingType(c, ret, symbol.Name+".return", 0)
			if crossingTypeError != nil {
				invalid = crossingTypeError
			}
		}
		if invalid != nil {
			if named {
				return nil, invalid
			}
			continue
		}
		exports = append(exports, export)
	}
	for name := range wanted {
		return nil, fmt.Errorf("%s: no such entry module export", name)
	}
	sort.Slice(exports, func(left, right int) bool { return exports[left].Name < exports[right].Name })
	return exports, nil
}

func compileExports(path string, names []string) (*ir.Program, []native.ABIExport, int) {
	return compileExportsWith(path, names, false)
}

func compileExportsWith(path string, names []string, entryFunctions bool) (*ir.Program, []native.ABIExport, int) {
	checked, code := check([]string{path})
	if checked == nil {
		return nil, nil, code
	}
	exports, exportSignaturesWithError := exportSignaturesWith(checked, names, entryFunctions)
	if exportSignaturesWithError != nil {
		fmt.Fprintln(os.Stderr, "adamic: "+exportSignaturesWithError.Error())
		return nil, nil, 1
	}
	program, lowerError := lower.Lower(context.Background(), checked)
	if lowerError != nil {
		fmt.Fprintln(os.Stderr, relative("adamic: "+lowerError.Error()))
		return nil, nil, 1
	}
	for exportIndex := range exports {
		found := -1
		for functionIndex, function := range program.Functions {
			if function.Name == exports[exportIndex].Name && !function.Closure {
				if found >= 0 {
					fmt.Fprintln(os.Stderr, "adamic: ambiguous ABI function name "+function.Name)
					return nil, nil, 1
				}
				found = functionIndex
			}
		}
		if found < 0 {
			fmt.Fprintln(os.Stderr, "adamic: export was not lowered: "+exports[exportIndex].Name)
			return nil, nil, 1
		}
		exports[exportIndex].Function = found
	}
	return program, exports, 0
}
