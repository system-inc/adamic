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

// environmentField is one declared binding as bridge metadata: its name, and whether it is the
// registered KV namespace, which the bridge checks as an object, or else a string.
type environmentField struct {
	name string
	kv   bool
}

// Ordered names and kinds are bridge metadata. The proven type stays within the
// borrowed checker lifetime, through the environment boundary check; neither
// checker facts nor an environment host object escape into the IR.
func workerEnvironmentFields(types *checker.Checker, proven *checker.Type) ([]environmentField, error) {
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
	fields := make([]environmentField, 0)
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
		// The prelude symbol's identity admits a binding, never a program interface's spelling; the
		// bridge checks it is a KV namespace before the handler runs.
		if load.WorkerKV(types, types.GetTypeOfSymbol(field)) {
			fields = append(fields, environmentField{name: field.Name, kv: true})
			continue
		}
		if types.GetTypeOfSymbol(field) != types.GetStringType() {
			return nil, environmentRefusal(field.Name, "must have type exactly string")
		}
		fields = append(fields, environmentField{name: field.Name})
	}
	return fields, nil
}

// environmentKVValidation checks each KV binding before the handler runs: an object, not null,
// whose get, put and delete are functions, as Workers' own binding is. Anything else answers the
// string fields' 500, its log naming the binding and never its value. It's a check, never a
// contract: typeof get === 'function' proves no return type, and what holds get's value is
// adamicWorkersTextOrNull on every admitted get (docs/axioms/workers.md).
const environmentKVValidation = `        for (const name of environmentKVFields) {
            const value = env?.[name];
            if (value === null || typeof value !== 'object' || typeof value.get !== 'function' || typeof value.put !== 'function' || typeof value.delete !== 'function') {
                console.error('environment binding ' + JSON.stringify(name) + ' is missing or not a KV namespace');
                return new Response('internal error', { status: 500 });
            }
            Object.defineProperty(environment, name, { value, enumerable: true });
        }
`

// Only this path receives an environment. The one-parameter bridges retain their
// exact head-1 bytes. Validate before calling user code, logging names only. An
// environment of strings alone keeps its bytes; a KV binding adds its own list and
// a check that it is an object whose get, put and delete are functions, answering
// the string fields' named 500, and the handler receives Workers' binding itself.
func environmentBridge(handlerImport string, fields []environmentField, async bool) (string, error) {
	bridge := Bridge(handlerImport)
	quoted, bindings := []string{}, []string{}
	for _, field := range fields {
		if field.kv {
			bindings = append(bindings, strconv.Quote(field.name))
		} else {
			quoted = append(quoted, strconv.Quote(field.name))
		}
	}
	declarations := "const environmentFields = [" + strings.Join(quoted, ", ") + "];\n\nexport default {"
	if len(bindings) != 0 {
		declarations = "const environmentFields = [" + strings.Join(quoted, ", ") + "];\nconst environmentKVFields = [" + strings.Join(bindings, ", ") + "];\n\nexport default {"
	}
	validation := `        const environment = {};
        for (const name of environmentFields) {
            const value = env?.[name];
            if (typeof value !== 'string') {
                console.error('environment binding ' + JSON.stringify(name) + ' is missing or not a string');
                return new Response('internal error', { status: 500 });
            }
            Object.defineProperty(environment, name, { value, enumerable: true });
        }
`
	if len(bindings) != 0 {
		validation += environmentKVValidation
	}
	validation += "\t\tlet output;"
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
