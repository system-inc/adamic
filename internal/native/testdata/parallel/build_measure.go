// Build a runtime C measurement harness against the compiler in the current checkout.
package main

import (
	"flag"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	source := flag.String("source", "", "C source")
	runtime := flag.String("runtime", "", "runtime source snapshot (optional)")
	output := flag.String("o", "", "output binary")
	flag.Parse()
	data, err := os.ReadFile(*source)
	if err != nil {
		panic(err)
	}
	if *runtime == "" {
		if err := native.Build(string(data), *output, native.Options{}); err != nil {
			panic(err)
		}
		return
	}
	library, err := native.RuntimeLibrary(*runtime, native.Options{})
	if err != nil {
		panic(err)
	}
	directory, err := os.MkdirTemp("", "adamic-measure-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(directory)
	sourcePath := filepath.Join(directory, "main.c")
	if err := os.WriteFile(sourcePath, data, 0600); err != nil {
		panic(err)
	}
	arguments := append(native.Flags(native.Options{}), "-I", filepath.Dir(library), sourcePath, "-o", *output)
	arguments = append(arguments, native.RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
		panic(string(output) + err.Error())
	}
}
