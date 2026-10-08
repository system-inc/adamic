package load

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// nearestProject follows tsc's upward tsconfig search. Explicit .a roots keep
// Adamic's options. A standalone .ts file keeps the existing Adamic fallback.
func nearestProject(file string) string {
	for directory := filepath.Dir(file); ; directory = filepath.Dir(directory) {
		name := filepath.Join(directory, "tsconfig.json")
		if info, err := os.Stat(name); err == nil && !info.IsDir() {
			return name
		}
		if parent := filepath.Dir(directory); parent == directory {
			return ""
		}
	}
}

func projectOptionsForRoots(fs *sourceFS, roots []tspath.RootedFilePath) (*core.CompilerOptions, string, error) {
	var project string
	var adamic, standalone bool
	for _, root := range roots {
		if _, isAdamic := fs.adamicFile(root); isAdamic {
			adamic = true
			continue
		}
		owner := nearestProject(root.AsString())
		if owner == "" {
			standalone = true
			continue
		}
		if project != "" && project != owner {
			return nil, "", fmt.Errorf("load: separate checker ownership is not implemented for projects %s and %s", project, owner)
		}
		project = owner
	}
	if project == "" {
		return compilerOptions(), "", nil
	}
	if adamic || standalone {
		return nil, "", fmt.Errorf("load: mixed project .ts and Adamic-option files need separate checker ownership, not implemented yet")
	}
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(project), &core.CompilerOptions{}, nil, fs, nil)
	formatter := &Program{fs: fs}
	if config != nil {
		diagnostics = append(diagnostics, config.GetConfigFileParsingDiagnostics()...)
	}
	if len(diagnostics) != 0 {
		messages := make([]string, 0, len(diagnostics))
		for _, diagnostic := range diagnostics {
			messages = append(messages, formatter.formatDiagnostic(diagnostic))
		}
		return nil, "", &CheckError{Diagnostics: messages}
	}
	if config == nil {
		return nil, "", fmt.Errorf("load: %s parsed to no project", project)
	}
	// Explicit command-line roots join the project in the option audit too.
	// auditProjectOptions receives these roots from Load, so each is checked under
	// the same declarations and options as the production program.
	options := config.CompilerOptions().Clone()
	// Undefined and callback variance must retain their meaning in Adamic's
	// representations even when a project disables these soundness rules.
	options.StrictNullChecks = core.TSTrue
	options.StrictFunctionTypes = core.TSTrue
	return options, project, nil
}
