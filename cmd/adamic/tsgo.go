package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func build(path, output string, arguments []string) int {
	return buildWithNative(path, output, arguments, native.Build)
}

// programRegionRequested is the Program region's opt-in. It reads the environment, not the arguments,
// so it lives outside buildWithNative and isn't a word the command matches.
func programRegionRequested() bool {
	return os.Getenv("ADAMIC_PROGRAM_REGION") == "1"
}

func buildWithExports(path, output string, arguments []string, entryFunctions bool) int {
	return buildWithNativeExports(path, output, arguments, native.Build, entryFunctions)
}

func buildWithNative(path, output string, arguments []string, buildNative func(string, string, native.Options) error) int {
	return buildWithNativeExports(path, output, arguments, buildNative, false)
}

func buildWithNativeExports(path, output string, arguments []string, buildNative func(string, string, native.Options) error, entryFunctions bool) int {
	// native links from inside its own temporary build directory and removes it, so a relative -o would name a path
	// in there and the binary would go with it: the command exited 0 and left nothing (Language, Oct 26). Abs fails
	// only when the working directory can't be read, which is the machine, so 3 (#z1ba56x's ruling).
	absoluteOutput, absoluteError := filepath.Abs(output)
	if absoluteError != nil {
		fmt.Fprintf(os.Stderr, "adamic: -o %s: %v\n", output, absoluteError)
		return 1
	}
	output = absoluteOutput
	options := native.Options{Release: true, ProgramRegion: programRegionRequested()}
	archive := ""
	explain := false
	reactor := false
	abiPath := ""
	names := []string{}
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
		case "--target":
			index++
			if index >= len(arguments) || options.Target != "" || arguments[index] != "wasm32-wasi" {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			options.Target = arguments[index]
		case "--explain-checks":
			explain = true
		case "--reactor":
			reactor = true
		case "--export", "--abi-json":
			flag := arguments[index]
			index++
			if index >= len(arguments) {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			if flag == "--export" {
				names = append(names, arguments[index])
				reactor = true
			} else {
				abiPath = arguments[index]
			}
		case "--count":
			options.Count = true
		case "--sanitize":
			options.Sanitize = true
		case "--tsgo":
			index++
			if index >= len(arguments) || archive != "" {
				fmt.Fprintln(os.Stderr, usage)
				return 2
			}
			archive = arguments[index]
		default:
			fmt.Fprintln(os.Stderr, usage)
			return 2
		}
	}
	if (reactor || abiPath != "") && options.Target != "wasm32-wasi" {
		fmt.Fprintln(os.Stderr, "adamic: export ABI requires wasm32-wasi")
		return 1
	}
	if options.Target != "" && archive != "" {
		fmt.Fprintln(os.Stderr, "adamic: --tsgo is not supported for wasm32-wasi")
		return 1
	}
	if err := native.ValidateOptions(options); err != nil {
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return 1
	}
	var program *ir.Program
	var code int
	handler := -1
	var exports []native.ABIExport
	if options.Target == "wasm32-wasi" {
		program, exports, code = compileExportsWith(path, names, entryFunctions)
		if program != nil && len(exports) == 0 && !reactor {
			// An export alias is outside the per-function crossing set (docs/wasm-abi.md), but an
			// aliased handleRequest still selects the request ABI, as it did before those exports.
			program, handler, code = compileWASI(path)
		}
		reactor = reactor || len(exports) > 0
		options.Request = reactor || handler >= 0
	} else {
		program, code = compileLibrary(path, archive != "")
	}
	if program == nil {
		return code
	}
	if explain {
		explainPredicateChecks(os.Stderr, program)
	}
	var err error
	if native.UsesTSGo(program) {
		if archive == "" {
			fmt.Fprintln(os.Stderr, "adamic: tsgo requires a linked checker archive; build with --tsgo <archive>")
			return 1
		}
		source, renderError := native.TSGoC(program)
		if renderError != nil {
			err = renderError
		} else {
			err = native.BuildTSGo(source, output, archive, options)
		}
	} else {
		var source string
		if options.Target == "wasm32-wasi" {
			if reactor {
				source, err = native.WASIExports(program, exports)
			} else {
				source, err = native.WASI(program, handler)
			}
		} else {
			source = native.C(program)
		}
		if err == nil {
			err = buildNative(source, output, options)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return 1
	}
	if abiPath != "" {
		bytes, err := json.MarshalIndent(native.ABITable{Version: 1, Exports: exports}, "", "  ")
		if err == nil {
			err = os.WriteFile(abiPath, append(bytes, '\n'), 0644)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "adamic: ABI JSON: %v\n", err)
			return 1
		}
	}
	return 0
}
