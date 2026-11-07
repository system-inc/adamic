// Command inventory derives the porting queue from cohere's compiled rules.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	root := flag.String("root", ".", "Adamic repository root")
	compiler := flag.String("compiler", "", "pinned TypeScript checkout containing src/compiler")
	output := flag.String("output", "stage1/cohere/lint/inventory", "output directory relative to root")
	reuse := flag.String("reuse", "", "reuse inventory.json evidence, requires capture-tests=false measure=false")
	capture := flag.Bool("capture-tests", true, "execute cohere rule tests and count harness invocations")
	measure := flag.Bool("measure", true, "run every applicable Go cohere rule on both corpora")
	flag.Parse()
	absolute, absolutePathError := filepath.Abs(*root)
	must(absolutePathError)
	destination := *output
	if !filepath.IsAbs(destination) {
		destination = filepath.Join(absolute, destination)
	}
	must(os.MkdirAll(destination, 0755))
	scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-lint-inventory-")
	must(temporaryDirectoryError)
	// Keep scratch paths in the log for inspection. Never remove another worker's files.
	fmt.Println("scratch:", scratch)
	cohere := filepath.Join(absolute, "cohere")
	engine := filepath.Join(absolute, "stage1/cohere/lint/inventory/testdata/engine.go")
	virtual := filepath.Join(cohere, "adamic_inventory.go")
	overlay := filepath.Join(scratch, "build-overlay.json")
	data, encodeError := json.Marshal(map[string]any{"Replace": map[string]string{virtual: engine}})
	must(encodeError)
	must(os.WriteFile(overlay, data, 0644))
	binary := filepath.Join(scratch, "inventory-engine")
	run(cohere, filepath.Join(destination, "build.log"), nil, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	args := []string{"-root", absolute, "-output", destination, "-compiler", *compiler}
	if *reuse != "" {
		args = append(args, "-reuse", *reuse)
	}
	if *capture {
		args = append(args, "-capture-tests")
	}
	if *measure {
		args = append(args, "-measure")
	}
	run(cohere, filepath.Join(destination, "generation.log"), nil, binary, args...)
	fmt.Println("wrote", filepath.Join(destination, "inventory.json"), "and inventory.md")
}
func must(sourceError error) {
	if sourceError != nil {
		fmt.Fprintln(os.Stderr, sourceError)
		os.Exit(1)
	}
}
func run(directory, log string, environment []string, name string, args ...string) {
	file, createError := os.Create(log)
	must(createError)
	command := exec.Command(name, args...)
	command.Dir = directory
	command.Stdout = file
	command.Stderr = file
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	runError := command.Run()
	must(file.Close())
	if runError != nil {
		must(fmt.Errorf("%s: %w; read %s", name, runError, log))
	}
}
