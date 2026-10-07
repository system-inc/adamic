package main

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/native"
)

func build(path, output string, arguments []string) int {
	options := native.Options{}
	archive := ""
	for index := 0; index < len(arguments); index++ {
		switch arguments[index] {
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
	program, code := compileLibrary(path, archive != "")
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
	} else {
		err = native.Build(native.C(program), output, options)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "adamic: %v\n", err)
		return 1
	}
	return 0
}
