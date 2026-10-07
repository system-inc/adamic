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
	if err := os.Mkdir(directory, 0755); err != nil {
		return err
	}
	for _, name := range []string{"generate-crossing.mjs", "wtf8.mjs", "crossing-runtime.mjs", "wasi.mjs"} {
		source, err := crossingGenerator.ReadFile("wasm/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, name), source, 0644); err != nil {
			return err
		}
	}
	command := exec.Command("node", filepath.Join(directory, "generate-crossing.mjs"), abiPath, outputPath)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("worker: crossing generation: %w\n%s", err, output)
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("worker: generate-crossing.mjs did not produce %s: %w", outputPath, err)
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
			return nil, fmt.Errorf("worker: refused Wasm name %q: not a top-level function in the entry; fix: name an entry function declaration", name)
		}
		index := -1
		for candidate, function := range lowered.Functions {
			if function.Name == name && !function.Closure {
				if index >= 0 {
					return nil, fmt.Errorf("worker: ambiguous Wasm name %q; fix: give entry and dependency functions distinct names", name)
				}
				index = candidate
			}
		}
		if index < 0 {
			return nil, fmt.Errorf("worker: Wasm function %q was not lowered; fix: use a concrete function declaration", name)
		}
		if err := CheckPure(lowered, index); err != nil {
			return nil, err
		}
		selected[index] = name
	}
	return selected, nil
}

func wasmFiles(entry string, checked *load.Program, lowered *ir.Program, options Options) (map[int]string, string, []workerFile, error) {
	selected, err := selectedFunctions(checked, lowered, options.Wasm)
	if err != nil {
		return nil, "", nil, err
	}
	if options.BuildWasm == nil {
		return nil, "", nil, fmt.Errorf("worker: Wasm build pipeline is not configured")
	}
	scratch, err := os.MkdirTemp("", "adamic-worker-wasm-")
	if err != nil {
		return nil, "", nil, err
	}
	defer os.RemoveAll(scratch)
	modulePath, abiPath := filepath.Join(scratch, "handler.wasm"), filepath.Join(scratch, "abi.json")
	if err := options.BuildWasm(entry, modulePath, abiPath, options.Wasm); err != nil {
		return nil, "", nil, err
	}
	module, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, "", nil, err
	}
	abi, err := os.ReadFile(abiPath)
	if err != nil {
		return nil, "", nil, err
	}
	var table native.ABITable
	if err := json.Unmarshal(abi, &table); err != nil || table.Version != 1 || len(table.Exports) != len(selected) {
		return nil, "", nil, fmt.Errorf("worker: build emitted an invalid ABI table: %v", err)
	}
	crossingPath := filepath.Join(scratch, "crossing.mjs")
	if err := generateCrossing(abiPath, crossingPath); err != nil {
		return nil, "", nil, err
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
		fmt.Fprintf(&state, "export function %s(...arguments_) {\n\tpanicMessage = '';\n\ttry { return crossing[%q](...arguments_); }\n\tcatch (error) { if (error instanceof WASIExit && error.code === 70) throw new AdamicPanic(panicMessage || 'Wasm panic'); throw error; }\n}\n", name, selected[index])
	}
	files := []workerFile{{"handler.wasm", string(module)}, {"abi.json", string(abi)}, {"wasm-state.mjs", state.String()}}
	for _, name := range []string{"crossing.mjs", "wtf8.mjs", "crossing-runtime.mjs", "wasi.mjs"} {
		contents, err := os.ReadFile(filepath.Join(scratch, name))
		if err != nil {
			return nil, "", nil, err
		}
		files = append(files, workerFile{name, string(contents)})
	}
	return external, "import { " + strings.Join(imports, ", ") + " } from './wasm-state.mjs';\n", files, nil
}

func wasmBridge(handlerImport string) string {
	bridge := Bridge(handlerImport)
	bridge = strings.Replace(bridge, "export default {", "import { initializeCrossing } from './wasm-state.mjs';\ninitializeCrossing();\n\nexport default {", 1)
	return bridge
}
