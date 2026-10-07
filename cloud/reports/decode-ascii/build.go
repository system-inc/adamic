// Measurement driver: select an unchanged compiler with either runtime snapshot.
package main

import (
	"context"
	"fmt"
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
	program, err := load.Load([]string{os.Args[2]})
	check(err)
	lowered, err := lower.Lower(context.Background(), program)
	check(err)
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
		source, err = native.WASI(lowered, handler)
		check(err)
		compiler = filepath.Join(filepath.Dir(filepath.Dir(os.Getenv("WASI_SYSROOT"))), "bin/clang")
	} else if os.Args[3] != "native" {
		panic("unsupported target")
	} else {
		source = native.C(lowered)
	}
	library, err := native.RuntimeLibrary(os.Args[1], options)
	check(err)
	directory, err := os.MkdirTemp("", "decode-service-build-")
	check(err)
	defer os.RemoveAll(directory)
	path := filepath.Join(directory, "main.c")
	check(os.WriteFile(path, []byte(source), 0644))
	arguments := append(native.Flags(options), "-I", filepath.Dir(library), "-o", os.Args[4])
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
	output, err := command.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("%v: %s", err, output))
	}
}
func check(err error) {
	if err != nil {
		panic(err)
	}
}
