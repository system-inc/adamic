// Package worker generates a runnable Cloudflare Worker from a checked Adamic handler.
package worker

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

//go:embed utf8.mjs
var utf8Runtime string

//go:embed json_decode.mjs
var jsonDecodeRuntime string

//go:embed json_encode.mjs
var jsonEncodeRuntime string

//go:embed runtime.mjs
var workerRuntime string

// CheckSignature requires the prelude's actual types, not structurally compatible lookalikes.
// The bridge is generated only after this check, before any output directory is touched.
func CheckSignature(ctx context.Context, program *load.Program) error {
	_, err := checkedWorkerEnvironment(ctx, program)
	return err
}

func checkedWorkerEnvironment(ctx context.Context, program *load.Program) ([]string, error) {
	entry := program.Files()[0]
	types, release := program.Checker(ctx, entry)
	defer release()
	var exported *ast.Symbol
	for _, symbol := range types.GetExportsOfModule(entry.AsNode().Symbol()) {
		if symbol.Name == "handle" && types.GetTypeOnlyAliasDeclaration(symbol) == nil {
			exported = symbol
			break
		}
	}
	found := "no exported handle"
	refuse := func() error {
		return fmt.Errorf("%s: refused: worker handle must be function handle(request: HttpRequest[, environment: Environment]): HttpResponse or async function handle(request: HttpRequest[, environment: Environment]): Promise<HttpResponse>, using the exact types from 'adamic/http'; Environment must be a program-declared interface of required readonly string properties; found %s\n  fix: import type { HttpRequest, HttpResponse } from 'adamic/http'; export a synchronous handle returning HttpResponse, or an async handle returning Promise<HttpResponse>; optionally add environment: Environment with required readonly string fields", program.FileName(entry), found)
	}
	if exported == nil {
		return nil, refuse()
	}
	target := exported
	if target.Flags&ast.SymbolFlagsAlias != 0 {
		target = types.GetAliasedSymbol(target)
	}
	functionType := types.GetTypeOfSymbolAtLocation(target, entry.AsNode())
	found = "handle: " + types.TypeToString(functionType)
	signatures := types.GetSignaturesOfType(functionType, checker.SignatureKindCall)
	if len(signatures) != 1 || len(target.Declarations) != 1 {
		return nil, refuse()
	}
	declaration := target.Declarations[0]
	if declaration.Kind != ast.KindFunctionDeclaration || len(declaration.TypeParameters()) != 0 || (len(declaration.Parameters()) < 1 || len(declaration.Parameters()) > 2) {
		return nil, refuse()
	}
	for _, node := range declaration.Parameters() {
		parameter := node.AsParameterDeclaration()
		if parameter.QuestionToken != nil || parameter.DotDotDotToken != nil || parameter.Initializer != nil {
			return nil, refuse()
		}
	}
	runtime := types.TryFindAmbientModule("adamic/http")
	if runtime == nil {
		return nil, fmt.Errorf("worker: missing adamic/http prelude module")
	}
	var requestType, responseType *checker.Type
	for _, symbol := range types.GetExportsOfModule(runtime) {
		switch symbol.Name {
		case "HttpRequest":
			requestType = types.GetDeclaredTypeOfSymbol(symbol)
		case "HttpResponse":
			responseType = types.GetDeclaredTypeOfSymbol(symbol)
		}
	}
	if requestType == nil || responseType == nil {
		return nil, fmt.Errorf("worker: missing HttpRequest or HttpResponse in the adamic prelude")
	}
	signature := signatures[0]
	result := types.GetReturnTypeOfSignature(signature)
	if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) {
		if !workerLibraryPromise(result) {
			return nil, refuse()
		}
		arguments := types.GetTypeArguments(result)
		if len(arguments) != 1 {
			return nil, refuse()
		}
		result = arguments[0]
	}
	if len(signature.Parameters()) != len(declaration.Parameters()) || types.GetTypeOfSymbolAtLocation(signature.Parameters()[0], declaration) != requestType || result != responseType {
		return nil, refuse()
	}
	if len(signature.Parameters()) == 2 {
		annotation := declaration.Parameters()[1].AsParameterDeclaration().Type
		if annotation != nil && annotation.Kind == ast.KindTypeReference && len(annotation.TypeArguments()) != 0 {
			return nil, environmentRefusal("environment", "generic interfaces are not supported")
		}
		return workerEnvironmentFields(types, types.GetTypeOfSymbolAtLocation(signature.Parameters()[1], declaration))
	}
	return nil, nil
}

// The Promise must come from the bundled library, not a structural thenable or a
// program declaration named Promise. Its single payload is checked by identity.
func workerLibraryPromise(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return false
	}
	symbol := proven.Symbol()
	return symbol != nil && symbol.Name == "Promise" && len(symbol.Declarations) != 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// refuseDroppedExports refuses each value export of the entry that lowering didn't keep. Lowering keeps
// only functions and closures (recordExports), and the handler module exports exactly that list, so a
// class, a value that isn't a closure, an enum or a namespace would vanish from the Worker's module where
// the source's module has it. A type-only export has no value in either module and stays out.
func refuseDroppedExports(program *load.Program, lowered *ir.Program) error {
	kept := map[string]bool{}
	for _, exported := range lowered.Exports {
		kept[exported.Name] = true
	}
	entry := program.Files()[0]
	types, release := program.Checker(context.Background(), entry)
	defer release()
	exports := types.GetExportsOfModule(entry.AsNode().Symbol())
	sort.Slice(exports, func(left, right int) bool { return exports[left].Name < exports[right].Name })
	for _, exported := range exports {
		if kept[exported.Name] || types.GetTypeOnlyAliasDeclaration(exported) != nil {
			continue
		}
		target := exported
		if target.Flags&ast.SymbolFlagsAlias != 0 {
			target = types.GetAliasedSymbol(target)
		}
		if target.Flags&ast.SymbolFlagsValue == 0 || len(target.Declarations) == 0 {
			continue
		}
		declaration := target.Declarations[0]
		if target.ValueDeclaration != nil {
			declaration = target.ValueDeclaration
		}
		kind := "a " + declaration.Kind.String()
		switch declaration.Kind {
		case ast.KindClassDeclaration:
			kind = "a class"
		case ast.KindEnumDeclaration:
			kind = "an enum"
		case ast.KindModuleDeclaration:
			kind = "a namespace"
		case ast.KindVariableDeclaration:
			kind = "a let value that isn't a closure"
			if declaration.Parent != nil && declaration.Parent.Flags&ast.NodeFlagsConst != 0 {
				kind = "a const value that isn't a closure"
			}
		case ast.KindFunctionDeclaration:
			kind = "a function lowering kept no body for"
		}
		source := ast.GetSourceFileOfNode(declaration)
		site := declaration
		if name := declaration.Name(); name != nil {
			site = name
		}
		line, column := scanner.GetLineAndCharacterOfPosition(source, scanner.GetTokenPosOfNode(site, source, false))
		return fmt.Errorf("%s:%d:%d: refused: worker export %s is %s, and a Worker's module exports only functions and closures\n  fix: export a function that returns it, or stop exporting it", program.FileName(source), line+1, column+1, exported.Name, kind)
	}
	return nil
}

// Bridge imports a handler module; both the compiled handler and stripped source use this code.
// Its fixed request and response layout follows the signature validated by CheckSignature.
func Bridge(handlerImport string) string {
	return `// Generated by adamic worker. Do not edit.
import { handle } from ` + strconv.Quote(handlerImport) + `;
import { AdamicPanic } from './adamic.mjs';

export default {
	async fetch(request, env, context) {
		const url = new URL(request.url);
		const input = {
			method: request.method,
			url: request.url,
			path: url.pathname,
			query: Array.from(url.searchParams, ([name, value]) => ({ name, value })),
			headers: Array.from(request.headers, ([name, value]) => ({ name, value })),
			body: await request.text(),
		};
		let output;
		try {
			output = handle(input);
		} catch (error) {
			if (!(error instanceof AdamicPanic)) throw error;
			console.error(error.message);
			return new Response('internal error', { status: 500 });
		}
		const headers = new Headers();
		for (const field of output.headers) headers.append(field.name, field.value);
		return new Response(output.body, { status: output.status, headers });
	},
};
`
}

// Keep the synchronous bridge byte-for-byte unchanged, including its scheduling.
func asyncBridge(handlerImport string) (string, error) {
	return awaitHandlerCall(Bridge(handlerImport))
}

func awaitHandlerCall(bridge string) (string, error) {
	const needle = "output = handle(input);"
	if count := strings.Count(bridge, needle); count != 1 {
		return "", fmt.Errorf("worker: refused async bridge: expected exactly one %q, found %d; fix: restore the generated handler call", needle, count)
	}
	return strings.Replace(bridge, needle, "output = await handle(input);", 1), nil
}

// Refuse before lowering or choosing a split: an async entry never becomes a
// JavaScript leftover beside selected Wasm functions in this first async head.
func refuseAsyncWasm(program *load.Program, asyncHandle bool, names []string) error {
	if len(names) == 0 {
		return nil
	}
	if asyncHandle {
		return fmt.Errorf("worker: refused --wasm with async handle; fix: omit --wasm for an async handle")
	}
	for _, statement := range program.Files()[0].Statements.Nodes {
		if statement.Kind != ast.KindFunctionDeclaration || statement.Name() == nil || !ast.HasSyntacticModifier(statement, ast.ModifierFlagsAsync) {
			continue
		}
		for _, name := range names {
			if statement.Name().Text() == name {
				return fmt.Errorf("worker: refused Wasm function %q: async functions cannot cross the Wasm boundary; fix: keep it in JavaScript and omit it from --wasm", name)
			}
		}
	}
	return nil
}

// Build checks and lowers before writing anything, including on a signature refusal.
func Build(entry, output string) error {
	return BuildWith(entry, output, Options{})
}

func BuildWith(entry, output string, options Options) error {
	program, err := load.Load([]string{entry})
	if err != nil {
		return err
	}
	environment, err := checkedWorkerEnvironment(context.Background(), program)
	if err != nil {
		return err
	}
	types, release := program.Checker(context.Background(), program.Files()[0])
	asyncHandle := false
	genericExport := ""
	for _, exported := range types.GetExportsOfModule(program.Files()[0].AsNode().Symbol()) {
		if types.GetTypeOnlyAliasDeclaration(exported) != nil {
			continue
		}
		target := exported
		if target.Flags&ast.SymbolFlagsAlias != 0 {
			target = types.GetAliasedSymbol(target)
		}
		for _, declaration := range target.Declarations {
			if exported.Name == "handle" {
				asyncHandle = ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync)
			}
			if declaration.Kind == ast.KindFunctionDeclaration && len(declaration.TypeParameters()) != 0 {
				if genericExport == "" {
					genericExport = exported.Name
				}
			}
		}
	}
	release()
	if err := refuseAsyncWasm(program, asyncHandle, options.Wasm); err != nil {
		return err
	}
	if genericExport != "" {
		return fmt.Errorf("worker: exported generic function %s is not yet supported; fix: export a concrete function", genericExport)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	if err := refuseDroppedExports(program, lowered); err != nil {
		return err
	}
	hasHandle := false
	for _, exported := range lowered.Exports {
		if exported.Name == "handle" {
			hasHandle = true
		}
	}
	if !hasHandle {
		return fmt.Errorf("worker: checked handle was lost during lowering")
	}
	backend := javascript.Options{ModuleExports: true, RuntimeImport: "./adamic.mjs"}
	var wasm []workerFile
	var imports string
	bridge := Bridge("./handler.mjs")
	if environment != nil {
		bridge, err = environmentBridge("./handler.mjs", environment, asyncHandle)
	} else if asyncHandle {
		bridge, err = asyncBridge("./handler.mjs")
	}
	if err != nil {
		return err
	}
	if len(options.Wasm) != 0 {
		backend.ExternalFunctions, imports, wasm, err = wasmFiles(entry, program, lowered, options)
		if err != nil {
			return err
		}
		bridge, err = initializeWasmBridge(bridge)
		if err != nil {
			return err
		}
	}
	handler := imports + javascript.JavaScriptWith(lowered, backend)
	name := strings.TrimSuffix(filepath.Base(entry), filepath.Ext(entry))
	configuration := fmt.Sprintf("name = %s\nmain = \"worker.mjs\"\ncompatibility_date = \"2026-10-07\"\n", strconv.Quote(name))
	if len(options.Wasm) != 0 {
		configuration += "\n[[rules]]\ntype = \"CompiledWasm\"\nglobs = [\"**/*.wasm\"]\nfallthrough = false\n"
	}
	runtime := strings.Replace(workerRuntime, "/* UTF8_RUNTIME */", strings.Replace(utf8Runtime, "export function createUtf8", "function createUtf8", 1), 1)
	const decoderStub = "// The JavaScript backend imports decodeJson. Its Workers implementation is a separate unit, so until it\n// lands a call panics, loudly, rather than decoding with code the oracle also runs.\nexport function decodeJson() { panic('decodeJson: not yet available on Workers'); }"
	if strings.Count(runtime, decoderStub) != 1 {
		return fmt.Errorf("worker: decodeJson runtime assembly marker missing")
	}
	runtime = strings.Replace(runtime, decoderStub, jsonDecodeRuntime, 1)
	const encoderStub = "// Keep the backend's encoder import linkable until the Workers JSON runtime unit lands.\nexport function encodeJson() { panic('encodeJson: not yet available on Workers'); }"
	if strings.Count(runtime, encoderStub) != 1 {
		return fmt.Errorf("worker: encodeJson runtime assembly marker missing")
	}
	runtime = strings.Replace(runtime, encoderStub, jsonEncodeRuntime, 1)
	files := []workerFile{
		{"handler.mjs", handler},
		{"adamic.mjs", runtime},
		{"worker.mjs", bridge},
		{"wrangler.toml", configuration},
	}
	files = append(files, wasm...)
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	for _, file := range files {
		if err := os.WriteFile(filepath.Join(output, file.name), []byte(file.text), 0644); err != nil {
			return err
		}
	}
	return nil
}
