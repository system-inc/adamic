package main

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func build(path, output string, arguments []string) int {
	options := native.Options{}
	archive := ""
	explain := false
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
	if options.Target == "wasm32-wasi" {
		program, handler, code = compileWASI(path)
		options.Request = handler >= 0
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
			source, err = native.WASI(program, handler)
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
	if explain && len(ir.InsertedChecks(program)) != 0 {
		explainChecks(program, os.Stderr)
	}
	return 0
}
