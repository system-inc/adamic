// Measurement driver: select an unchanged compiler with either runtime snapshot.
package main

import (
	"context"
	"fmt"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) != 5 {
		panic("runtime directory, source, target (native or wasi), output required")
	}
	program, loadError := load.Load([]string{os.Args[2]})
	check(loadError)
	lowered, lowerError := lower.Lower(context.Background(), program)
	check(lowerError)
	options := native.Options{}
	var source string
	compiler := "clang"
	if os.Args[3] == "wasi" {
		options.Target = "wasm32-wasi"
		options.Request = true
		handler := -1
		for i, function := range lowered.Functions {
			if function.Name == "handleRequest" {
				if handler != -1 {
					panic("ambiguous handler")
				}
				handler = i
			}
		}
		if handler < 0 {
			panic("handler absent")
		}
		var wasiError error
		source, wasiError = native.WASI(lowered, handler)
		check(wasiError)
		compiler = filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin/clang")
	} else if os.Args[3] != "native" {
		panic("unsupported target")
	} else {
		source = native.C(lowered)
	}
	library, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(os.Args[1], source, options)
	check(runtimeLibraryForSourceError)
	directory, temporaryDirectoryError := os.MkdirTemp("", "decode-service-build-")
	check(temporaryDirectoryError)
	defer os.RemoveAll(directory)
	path := filepath.Join(directory, "main.c")
	check(os.WriteFile(path, []byte(source), 0644))
	arguments := append(native.SourceFlags(source, options), "-I", filepath.Dir(library), "-o", os.Args[4])
	if !options.Request {
		arguments = append(arguments, path)
	}
	arguments = append(arguments, native.RuntimeLinkFlags(library)...)
	if options.Request {
		arguments = append(arguments, path)
	}
	arguments = append(arguments, "-lm")
	if options.Target != "" {
		arguments = append(arguments, native.WASILinkFlags(options)...)
	}
	command := exec.Command(compiler, arguments...)
	output, combinedOutputError := command.CombinedOutput()
	if combinedOutputError != nil {
		panic(fmt.Sprintf("%v: %s", combinedOutputError, output))
	}
}
func check(sourceError error) {
	if sourceError != nil {
		panic(sourceError)
	}
}
