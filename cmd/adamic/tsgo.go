package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func build(path, output string, arguments []string) int {
	return buildWithExports(path, output, arguments, false)
}

func buildWithExports(path, output string, arguments []string, entryFunctions bool) int {
	options := native.Options{Release: true}
	archive := ""
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
	if err := native.ValidateOptions(options); err != nil {
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return 1
	}
	if options.Target != "" && archive != "" {
		fmt.Fprintln(os.Stderr, "adamic: --tsgo is not supported for wasm32-wasi")
		return 1
	}
	var program *ir.Program
	var code int
	handler := -1
	var exports []native.ABIExport
	if options.Target == "wasm32-wasi" {
		program, exports, code = compileExportsWith(path, names, entryFunctions)
		reactor = reactor || len(exports) > 0
		options.Request = reactor
	} else {
		program, code = compileLibrary(path, archive != "")
	}
	if program == nil {
		return code
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
	} else if native.UsesApple(program) {
		err = native.BuildApple(native.C(program), output, options)
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
			err = native.Build(source, output, options)
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
