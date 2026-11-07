package worker

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// Options keeps the ordinary four-file Worker contract when Wasm is empty.
type Options struct {
	Wasm []string
	// BuildWasm invokes the compiler's build/export/ABI pipeline on the same entry.
	BuildWasm func(entry, module, abi string, names []string) error
}

type workerFile struct{ name, text string }

//go:embed wasm/generate-crossing.mjs wasm/wtf8.mjs wasm/crossing-runtime.mjs wasm/wasi.mjs
var crossingGenerator embed.FS

// generateCrossing is the build-time seam to the ABI crossing generator.
func generateCrossing(abiPath, outputPath string) error {
	directory := filepath.Join(filepath.Dir(outputPath), "generator")
	if createDirectoryError := os.Mkdir(directory, 0755); createDirectoryError != nil {
		return createDirectoryError
	}
	for _, name := range []string{"generate-crossing.mjs", "wtf8.mjs", "crossing-runtime.mjs", "wasi.mjs"} {
		source, readFileError := crossingGenerator.ReadFile("wasm/" + name)
		if readFileError != nil {
			return readFileError
		}
		if writeError := os.WriteFile(filepath.Join(directory, name), source, 0644); writeError != nil {
			return writeError
		}
	}
	command := exec.Command("node", filepath.Join(directory, "generate-crossing.mjs"), abiPath, outputPath)
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		return fmt.Errorf("worker: crossing generation: %w\n%s", combinedOutputError, output)
	}
	if _, statError := os.Stat(outputPath); statError != nil {
		return fmt.Errorf("worker: generate-crossing.mjs did not produce %s: %w", outputPath, statError)
	}
	return nil
}

func selectedFunctions(checked *load.Program, lowered *ir.Program, names []string) (map[int]string, error) {
	entry := checked.Files()[0]
	declarations := map[string]bool{}
	for _, statement := range entry.Statements.Nodes {
		if statement.Kind == ast.KindFunctionDeclaration && statement.Name() != nil {
			declarations[statement.Name().Text()] = true
		}
	}
	selected := map[int]string{}
	for _, name := range names {
		if !declarations[name] {
			return nil, fmt.Errorf(
				"worker: refused Wasm name %q: not a top-level function in the entry; fix: name an entry function declaration",
				name,
			)
		}
		index := -1
		for candidate, function := range lowered.Functions {
			if function.Name == name && !function.Closure {
				if index >= 0 {
					return nil, fmt.Errorf(
						"worker: ambiguous Wasm name %q; fix: give entry and dependency functions distinct names",
						name,
					)
				}
				index = candidate
			}
		}
		if index < 0 {
			return nil, fmt.Errorf(
				"worker: Wasm function %q was not lowered; fix: use a concrete function declaration",
				name,
			)
		}
		if checkPureError := CheckPure(lowered, index); checkPureError != nil {
			return nil, checkPureError
		}
		selected[index] = name
	}
	return selected, nil
}

func wasmFiles(
	entry string,
	checked *load.Program,
	lowered *ir.Program,
	options Options,
) (map[int]string, string, []workerFile, error) {
	selected, selectedFunctionsError := selectedFunctions(checked, lowered, options.Wasm)
	if selectedFunctionsError != nil {
		return nil, "", nil, selectedFunctionsError
	}
	if options.BuildWasm == nil {
		return nil, "", nil, fmt.Errorf("worker: Wasm build pipeline is not configured")
	}
	scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-worker-wasm-")
	if temporaryDirectoryError != nil {
		return nil, "", nil, temporaryDirectoryError
	}
	defer os.RemoveAll(scratch)
	modulePath, abiPath := filepath.Join(scratch, "handler.wasm"), filepath.Join(scratch, "abi.json")
	if buildWasmError := options.BuildWasm(entry, modulePath, abiPath, options.Wasm); buildWasmError != nil {
		return nil, "", nil, buildWasmError
	}
	module, readError := os.ReadFile(modulePath)
	if readError != nil {
		return nil, "", nil, readError
	}
	abi, readError := os.ReadFile(abiPath)
	if readError != nil {
		return nil, "", nil, readError
	}
	var table native.ABITable
	if decodeError := json.Unmarshal(
		abi,
		&table,
	); decodeError != nil || table.Version != 1 || len(table.Exports) != len(selected) {
		return nil, "", nil, fmt.Errorf("worker: build emitted an invalid ABI table: %v", decodeError)
	}
	crossingPath := filepath.Join(scratch, "crossing.mjs")
	if generateCrossingError := generateCrossing(abiPath, crossingPath); generateCrossingError != nil {
		return nil, "", nil, generateCrossingError
	}
	indices := make([]int, 0, len(selected))
	for index := range selected {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	external := map[int]string{}
	imports := []string{}
	var state strings.Builder
	state.WriteString(`import module from './handler.wasm';
import { createCrossing } from './crossing.mjs';
import { WASIExit } from './wasi.mjs';
import { AdamicPanic } from './adamic.mjs';
let crossing;
let panicMessage = '';
export function initializeCrossing() {
 crossing ??= createCrossing(module, { logger: {
  log: line => console.log(line),
  error: line => {
   if (line.startsWith('adamic: panic: ')) panicMessage = line.slice('adamic: panic: '.length);
   else console.error(line);
  },
 } });
}
initializeCrossing();
`)
	for _, index := range indices {
		name := "wasmFunction_" + strconv.Itoa(index)
		external[index] = name
		imports = append(imports, name)
		fmt.Fprintf(
			&state,
			"export function %s(...arguments_) {\n\tpanicMessage = '';\n\ttry { return crossing[%q](...arguments_); }\n\tcatch (error) { if (error instanceof WASIExit && error.code === 70) throw new AdamicPanic(panicMessage || 'Wasm panic'); throw error; }\n}\n",
			name,
			selected[index],
		)
	}
	files := []workerFile{
		{"handler.wasm", string(module)},
		{"abi.json", string(abi)},
		{"wasm-state.mjs", state.String()},
	}
	for _, name := range []string{"crossing.mjs", "wtf8.mjs", "crossing-runtime.mjs", "wasi.mjs"} {
		contents, readError := os.ReadFile(filepath.Join(scratch, name))
		if readError != nil {
			return nil, "", nil, readError
		}
		files = append(files, workerFile{name, string(contents)})
	}
	return external, "import { " + strings.Join(imports, ", ") + " } from './wasm-state.mjs';\n", files, nil
}

func wasmBridge(handlerImport string) string {
	bridge := Bridge(handlerImport)
	bridge = strings.Replace(
		bridge,
		"export default {",
		"import { initializeCrossing } from './wasm-state.mjs';\ninitializeCrossing();\n\nexport default {",
		1,
	)
	return bridge
}
