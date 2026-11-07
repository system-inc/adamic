package native

import (
	"os"
	"path/filepath"
	"strings"
)

// Only runtime construction needs the pattern compiler and Unicode tables.
// Literal programs share the ordinary runtime without parsing these sources.
func regexpCompilerSource(source string) bool {
	return strings.Contains(source, "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n")
}

func regexpCompilerLibrary(source string, options Options) (string, error) {
	flags := append(featureFlags(source), "-DADAMIC_REGEXP_RUNTIME_COMPILER=1")
	files, flags, compiler, version, runtimeInputsError := runtimeInputs("", options, flags)
	if runtimeInputsError != nil {
		return "", runtimeInputsError
	}
	// The component reads runtime headers but compiles only its own units.
	kept := files[:0]
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") || strings.HasPrefix(file.name, "regexp_compile") {
			kept = append(kept, file)
		}
	}
	return runtimeLibraryFromInputs(kept, flags, compiler, version, options, true, "native-regexp-compiler")
}

// The archive is named relative to the build directory so debug maps and
// sanitizer metadata never record the machine's cache location.
func regexpCompilerLinkFlags(source string, options Options, directory string) ([]string, error) {
	if !regexpCompilerSource(source) {
		return nil, nil
	}
	library, regexpCompilerLibraryError := regexpCompilerLibrary(source, options)
	if regexpCompilerLibraryError != nil {
		return nil, regexpCompilerLibraryError
	}
	const name = "regexp-compiler"
	if symlinkError := os.Symlink(filepath.Dir(library), filepath.Join(directory, name)); symlinkError != nil {
		return nil, symlinkError
	}
	archive := filepath.Join(name, filepath.Base(library))
	if options.Target == "wasm32-wasi" {
		return []string{"-Xlinker", "--whole-archive", archive, "-Xlinker", "--no-whole-archive"}, nil
	}
	return RuntimeLinkFlags(archive), nil
}
