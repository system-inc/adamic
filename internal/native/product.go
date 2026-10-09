package native

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildProduct builds a native cache product in its private miss directory.
// Paths in diagnostics and debug information are relative so rebuilding in a
// different directory produces identical bytes for the shared store's audit.
// The directory must be empty; the completed product contains only the binary.
func BuildProduct(source, directory string, options Options) error {
	if err := ValidateOptions(options); err != nil {
		return err
	}
	if options.Target != "" || options.Request || options.Split || os.Getenv("ADAMIC_NATIVE_SPLIT") == "1" {
		return fmt.Errorf("native: reproducible products require an unsplit native target")
	}
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		return err
	}
	features := featureFlags(source)
	flags := sourceFlags(source, options)
	flags = append(flags, "-ffile-prefix-map="+directory+"=.", "-fdebug-compilation-dir=.", "-I", ".")
	main := filepath.Join(directory, "main.c")
	if err := os.WriteFile(main, []byte(source), 0o644); err != nil {
		return err
	}
	arguments := append(flags, "main.c")
	var scratch []string
	scratch = append(scratch, main)
	for _, file := range files {
		if file.name == "regexp_replace.c" && !slicesContain(features, "-DADAMIC_REGEXP_REPLACE_CALLBACK=1") {
			continue
		}
		if file.name == "node_host.c" && !slicesContain(features, "-DADAMIC_NODE_HOST=1") {
			continue
		}
		path := filepath.Join(directory, file.name)
		if err := os.WriteFile(path, file.contents, 0o644); err != nil {
			return err
		}
		scratch = append(scratch, path)
		if strings.HasSuffix(file.name, ".c") {
			arguments = append(arguments, file.name)
		}
	}
	arguments = append(arguments, "-lm", "-o", "port")
	command := exec.Command(compilerName(options), arguments...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("native: reproducible clang build: %w\n%s", err, output)
	}
	for _, path := range scratch {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}
