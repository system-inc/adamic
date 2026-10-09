package program

import (
	"fmt"
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// ProjectConfig is what a tsconfig says belongs to it, read without building its program: the files it
// includes and the tsconfigs it references. Discovery reads one per project, to find solution roots and to
// give each file to the nearest tsconfig that includes it.
type ProjectConfig struct {
	// FileNames are the files the tsconfig includes, absolute and cleaned to the platform's separator.
	FileNames []string

	// References are the tsconfigs it references, absolute: a reference to a directory names the
	// tsconfig.json in it, as tsc reads one.
	References []string
}

// ReadProjectConfig reads a tsconfig the way Build does, and fails where Build would refuse it, so a
// config discovery cannot read is one the project's own run reports.
func ReadProjectConfig(configFileName string) (ProjectConfig, error) {
	configFileName = tspath.NormalizePath(configFileName)
	fileSystem := cachedvfs.From(bundled.WrapFS(osvfs.FS()))
	if !fileSystem.FileExists(tspath.RootedFilePath(configFileName)) {
		return ProjectConfig{}, fmt.Errorf("no tsconfig at %s", configFileName)
	}
	config, configErrors := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePath(configFileName), &core.CompilerOptions{}, nil, fileSystem, nil)
	if len(configErrors) > 0 {
		return ProjectConfig{}, fmt.Errorf("reading %s: %w", configFileName, joinDiagnostics(configErrors))
	}
	if config == nil {
		return ProjectConfig{}, fmt.Errorf("reading %s: the config parsed to nothing", configFileName)
	}
	if diagnostics := config.GetConfigFileParsingDiagnostics(); len(diagnostics) > 0 && !onlyOptionValueDiagnostics(diagnostics) {
		return ProjectConfig{}, fmt.Errorf("reading %s: %w", configFileName, joinDiagnostics(diagnostics))
	}
	read := ProjectConfig{}
	for _, fileName := range config.FileNames() {
		read.FileNames = append(read.FileNames, filepath.Clean(filepath.FromSlash(fileName.AsString())))
	}
	for _, reference := range config.ResolvedProjectReferencePaths() {
		read.References = append(read.References, filepath.Clean(filepath.FromSlash(reference.AsString())))
	}
	return read, nil
}
