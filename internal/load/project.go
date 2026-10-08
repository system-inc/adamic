package load

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

// LoadProject uses the checker's own config parser, including inherited options and root order.
func LoadProject(path string) (*Program, error) { return loadInput(nil, nil, path) }

// LoadProjectEntry checks the entire project and selects one configured source as its runtime entry.
// Entry paths are relative to the current directory, like direct file arguments.
func LoadProjectEntry(path, entry string) (*Program, error) {
	if entry == "" {
		return nil, fmt.Errorf("load: project builds require --entry <file>; select a runtime entry from the config's files/include")
	}
	program, err := LoadProject(path)
	if err != nil {
		return nil, err
	}
	directory, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	absoluteEntry, err := filepath.Abs(entry)
	if err != nil {
		return nil, err
	}
	root, err := rootFileName(program.fs, tspath.RootedDirectoryPathFromAbsolute(directory), absoluteEntry)
	if err != nil {
		return nil, err
	}
	for _, file := range program.files {
		if file.FileName() == root {
			program.entries = []*ast.SourceFile{file}
			return program, nil
		}
	}
	return nil, fmt.Errorf("load: entry %s is outside the project's configured source file list; add it to files/include in %s or choose an entry already listed", entry, path)
}

func parseProject(path string, directory tspath.RootedDirectoryPath, fs *sourceFS) (*tsoptions.ParsedCommandLine, error) {
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(directory.ResolveFile(path), nil, nil, &projectFS{FS: fs}, nil)
	if config != nil {
		diagnostics = append(diagnostics, config.Errors...)
	}
	if len(diagnostics) != 0 {
		loaded := &Program{fs: fs}
		messages := make([]string, 0, len(diagnostics))
		for _, diagnostic := range diagnostics {
			messages = append(messages, loaded.formatDiagnostic(diagnostic))
		}
		return nil, &CheckError{Diagnostics: messages}
	}
	if config == nil {
		return nil, fmt.Errorf("load: no config at %s", path)
	}
	if len(config.ParsedConfig.ProjectReferences) != 0 {
		return nil, fmt.Errorf("load: project references are not yet supported; build each project explicitly")
	}
	adamic := false
	for _, file := range config.FileNames() {
		if _, proof := fs.adamicFile(file); proof {
			adamic = true
		}
	}
	if adamic {
		if err := projectOptions(config.CompilerOptions()); err != nil {
			return nil, err
		}
	}
	if len(config.FileNames()) == 0 {
		return nil, fmt.Errorf("load: project %s has no root files", path)
	}
	// Discover the host's actual global console declarations before choosing the prelude.
	fs.projectConsole = true
	return config, nil
}

// Omitted required options inherit Adamic's contract. Explicit weakening, including an inherited
// weakening, is refused before the program is built. Other options retain their project values.
func projectOptions(options *core.CompilerOptions) error {
	for _, option := range []struct {
		name  string
		value *core.Tristate
	}{
		{"strict", &options.Strict},
		{"noImplicitAny", &options.NoImplicitAny}, {"noImplicitThis", &options.NoImplicitThis},
		{"strictNullChecks", &options.StrictNullChecks}, {"strictFunctionTypes", &options.StrictFunctionTypes},
		{"strictBindCallApply", &options.StrictBindCallApply}, {"strictPropertyInitialization", &options.StrictPropertyInitialization},
		{"strictBuiltinIteratorReturn", &options.StrictBuiltinIteratorReturn}, {"useUnknownInCatchVariables", &options.UseUnknownInCatchVariables},
		{"alwaysStrict", &options.AlwaysStrict},
		{"noUncheckedIndexedAccess", &options.NoUncheckedIndexedAccess}, {"exactOptionalPropertyTypes", &options.ExactOptionalPropertyTypes},
		{"noImplicitReturns", &options.NoImplicitReturns}, {"noFallthroughCasesInSwitch", &options.NoFallthroughCasesInSwitch},
		{"erasableSyntaxOnly", &options.ErasableSyntaxOnly}, {"verbatimModuleSyntax", &options.VerbatimModuleSyntax},
		{"allowImportingTsExtensions", &options.AllowImportingTsExtensions}, {"noEmit", &options.NoEmit},
	} {
		if *option.value == core.TSFalse {
			return fmt.Errorf("load: compiler option %s is false; Adamic requires it to be true", option.name)
		}
		*option.value = core.TSTrue
	}
	for _, option := range []struct {
		name  string
		value *core.Tristate
	}{
		{"noCheck", &options.NoCheck}, {"skipLibCheck", &options.SkipLibCheck}, {"skipDefaultLibCheck", &options.SkipDefaultLibCheck},
		{"noResolve", &options.NoResolve}, {"noLib", &options.NoLib}, {"libReplacement", &options.LibReplacement},
	} {
		if *option.value == core.TSTrue {
			return fmt.Errorf("load: compiler option %s is true; Adamic requires it to be false", option.name)
		}
		*option.value = core.TSFalse
	}
	if options.GetAllowJS() && options.CheckJs != core.TSTrue {
		return fmt.Errorf("load: compiler option checkJs is not true; Adamic requires it to be true when allowJs enables JavaScript dependencies")
	}
	if options.Module == core.ModuleKindNone {
		options.Module = core.ModuleKindESNext
	}
	if options.Module != core.ModuleKindESNext {
		return fmt.Errorf("load: compiler option module differs; Adamic requires esnext")
	}
	if options.ModuleDetection == core.ModuleDetectionKindNone {
		options.ModuleDetection = core.ModuleDetectionKindForce
	}
	if options.ModuleDetection != core.ModuleDetectionKindForce {
		return fmt.Errorf("load: compiler option moduleDetection differs; Adamic requires force")
	}
	if options.ModuleResolution == core.ModuleResolutionKindUnknown {
		options.ModuleResolution = core.ModuleResolutionKindBundler
	}
	if options.ModuleResolution != core.ModuleResolutionKindBundler {
		return fmt.Errorf("load: compiler option moduleResolution differs; Adamic requires bundler")
	}
	if options.Target == core.ScriptTargetNone {
		options.Target = core.ScriptTargetES2024
	}
	if options.Target < core.ScriptTargetES2024 {
		return fmt.Errorf("load: compiler option target differs; Adamic requires es2024 or later")
	}
	if options.UseDefineForClassFields == core.TSFalse {
		return fmt.Errorf("load: compiler option useDefineForClassFields is false; Adamic requires it to be true")
	}
	if options.Lib == nil {
		options.Lib = []string{"lib.es2024.d.ts"}
	}
	return nil
}

// A host may supply console through libs, selected types, automatic type discovery, or explicit
// declaration roots. Names inside a module/namespace alone do not supply a global console.
func hasHostConsole(files []*ast.SourceFile) bool {
	globals := func(statements []*ast.Node) bool {
		for _, statement := range statements {
			if statement.Kind != ast.KindVariableStatement {
				continue
			}
			for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				if ast.IsIdentifier(declaration.Name()) && declaration.Name().Text() == "console" {
					return true
				}
			}
		}
		return false
	}
	for _, file := range files {
		if IsPrelude(file) || !file.IsDeclarationFile {
			continue
		}
		if !ast.IsExternalModule(file) && globals(file.Statements.Nodes) {
			return true
		}
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if ast.IsGlobalScopeAugmentation(node) {
				body := node.AsModuleDeclaration().Body
				if body != nil && body.Kind == ast.KindModuleBlock && globals(body.AsModuleBlock().Statements.Nodes) {
					return true
				}
			}
			return node.ForEachChild(visit)
		}
		if visit(file.AsNode()) {
			return true
		}
	}
	return false
}
