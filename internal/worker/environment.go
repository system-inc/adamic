package worker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func environmentRefusal(field, reason string) error {
	return fmt.Errorf("worker: refused environment field %q: %s; found an unsupported environment contract\n  fix: use a program-declared interface with required readonly string properties", field, reason)
}

// No checker objects escape this validation. The ordered string names are bridge
// metadata, not a new IR representation or an environment host object.
func workerEnvironmentFields(types *checker.Checker, proven *checker.Type) ([]string, error) {
	if proven.Flags()&checker.TypeFlagsObject == 0 || proven.Symbol() == nil {
		return nil, environmentRefusal("environment", "type must be a program-declared interface")
	}
	if alias := proven.Alias(); alias != nil {
		for _, d := range alias.Symbol().Declarations {
			if len(d.TypeParameters()) != 0 {
				return nil, environmentRefusal("environment", "generic interfaces are not supported")
			}
		}
	}
	symbol := proven.Symbol()
	if len(symbol.Declarations) == 0 {
		return nil, environmentRefusal("environment", "type must be a program-declared interface")
	}
	for _, d := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(d)
		if d.Kind != ast.KindInterfaceDeclaration || load.IsLibrary(file) || load.IsPrelude(file) || load.IsNodeLibrary(file) {
			return nil, environmentRefusal("environment", "type must be a program-declared interface")
		}
		if len(d.TypeParameters()) != 0 {
			return nil, environmentRefusal("environment", "generic interfaces are not supported")
		}
	}
	if len(types.GetIndexInfosOfType(proven)) != 0 {
		return nil, environmentRefusal("[index]", "index signatures are not supported")
	}
	if len(types.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 || len(types.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
		return nil, environmentRefusal("[call]", "call and construct signatures are not supported")
	}
	fields := make([]string, 0)
	for _, field := range types.GetPropertiesOfType(proven) {
		for _, d := range field.Declarations {
			if d.Kind != ast.KindPropertySignature {
				return nil, environmentRefusal(field.Name, "must be a data property, not a method or accessor")
			}
			if d.Name() == nil || (d.Name().Kind != ast.KindIdentifier && d.Name().Kind != ast.KindStringLiteral && d.Name().Kind != ast.KindNumericLiteral) {
				return nil, environmentRefusal(field.Name, "must have a literal string field name")
			}
		}
		if field.Flags&ast.SymbolFlagsOptional != 0 {
			return nil, environmentRefusal(field.Name, "must be required, not optional")
		}
		if !types.IsReadonlySymbol(field) {
			return nil, environmentRefusal(field.Name, "must be readonly")
		}
		if types.GetTypeOfSymbol(field) != types.GetStringType() {
			return nil, environmentRefusal(field.Name, "must have type exactly string")
		}
		fields = append(fields, field.Name)
	}
	return fields, nil
}

// Only this path receives an environment. The one-parameter bridges retain their
// exact head-1 bytes. Validate before calling user code, logging names only.
func environmentBridge(handlerImport string, fields []string, async bool) (string, error) {
	bridge := Bridge(handlerImport)
	quoted := make([]string, len(fields))
	for i, name := range fields {
		quoted[i] = strconv.Quote(name)
	}
	declarations := "const environmentFields = [" + strings.Join(quoted, ", ") + "];\n\nexport default {"
	validation := `        const environment = {};
        for (const name of environmentFields) {
            const value = env?.[name];
            if (typeof value !== 'string') {
                console.error('environment binding ' + JSON.stringify(name) + ' is missing or not a string');
                return new Response('internal error', { status: 500 });
            }
            Object.defineProperty(environment, name, { value, enumerable: true });
        }
` + "\t\tlet output;"
	call := "handle(input, environment)"
	if async {
		call = "await " + call
	}
	for _, replacement := range []struct{ needle, value string }{
		{"export default {", declarations},
		{"\t\tlet output;", validation},
		{"output = handle(input);", "output = " + call + ";"},
	} {
		if strings.Count(bridge, replacement.needle) != 1 {
			return "", fmt.Errorf("worker: refused environment bridge: expected exactly one %q; fix: restore the generated bridge template", replacement.needle)
		}
		bridge = strings.Replace(bridge, replacement.needle, replacement.value, 1)
	}
	return bridge, nil
}
