package load

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// A compiler program has one configuration, including its imported files. Locate
// the nearest project for TypeScript roots and refuse roots from different projects
// rather than silently choosing one root's options for another.
func projectConfig(fs *sourceFS, current tspath.RootedDirectoryPath, paths []string) (*tsoptions.ParsedCommandLine, bool, error) {
	// Explicit ambient roots (for example @types/node) share the implementation
	// root's project; their package directory is not another compilation project.
	implementation := false
	for _, path := range paths {
		if filepath.Ext(path) == ".ts" && !strings.HasSuffix(path, ".d.ts") {
			implementation = true
		}
	}
	var selected tspath.RootedFilePath
	typescript := false
	for _, path := range paths {
		if filepath.Ext(path) != ".ts" || implementation && strings.HasSuffix(path, ".d.ts") {
			continue
		}
		directory := filepath.Dir(current.ResolveFile(path).AsString())
		var found tspath.RootedFilePath
		for {
			candidate := current.ResolveFile(filepath.Join(directory, "tsconfig.json"))
			if fs.FileExists(candidate) {
				found = candidate
				break
			}
			parent := filepath.Dir(directory)
			if parent == directory {
				break
			}
			directory = parent
		}
		if typescript && found != selected {
			return nil, true, fmt.Errorf("load: TypeScript roots belong to different projects (%s and %s); load them separately", selected, found)
		}
		typescript, selected = true, found
	}
	if selected == "" {
		return nil, typescript, nil
	}
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(selected, nil, nil, fs, nil)
	if len(diagnostics) > 0 {
		program := &Program{fs: fs}
		formatted := make([]string, 0, len(diagnostics))
		for _, diagnostic := range diagnostics {
			formatted = append(formatted, program.formatDiagnostic(diagnostic))
		}
		return nil, true, &CheckError{Diagnostics: formatted}
	}
	// Adamic resolves source-extension imports and emits its own backends, never
	// TypeScript output. Keep project checking and resolution options otherwise;
	// NoEmit also makes AllowImportingTsExtensions valid for emitting projects.
	options := *config.CompilerOptions()
	options.AllowImportingTsExtensions = core.TSTrue
	options.NoEmit = core.TSTrue
	config.SetCompilerOptions(&options)
	return config, true, nil
}
