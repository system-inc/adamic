package load

import (
	"context"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
)

// The shadow changes only the JSON result declaration, never the production type.
// Surviving diagnostics stay ordinary errors, even when a JSON call is nearby.
func (p *Program) jsonContractSites(config *tsoptions.ParsedCommandLine) ([]OptionSite, error) {
	actual := p.compiler.GetSemanticDiagnostics(context.Background(), nil)
	if len(actual) == 0 {
		return nil, nil
	}
	source := *p.fs
	source.jsonAssumeString = true
	fs := cachedvfs.From(&regexpLibraryFS{FS: bundled.WrapFS(&source)})
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	shadow := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host, SingleThreaded: core.TSTrue})
	if shadow == nil {
		return nil, fmt.Errorf("load: checker built no JSON contract shadow")
	}
	remaining := diagnosticKeys(shadow.GetSemanticDiagnostics(context.Background(), nil))
	var sites []OptionSite
	for _, diagnostic := range actual {
		message := p.formatDiagnostic(diagnostic)
		if remaining[optionKey(diagnostic)] || diagnostic.File() == nil {
			continue
		}
		line, column := p.lineAndColumn(diagnostic.File(), diagnostic.Pos())
		sites = append(sites, OptionSite{File: diagnostic.File().FileName().AsString(), Line: line, Column: column, Code: int(diagnostic.Code()), Message: message, Options: []string{"JSON.stringify"}})
	}
	return sites, nil
}

func (p *Program) RequiresJSONStringifySite(node *ast.Node) bool {
	return p.requiresOptionSite(node, "JSON.stringify")
}
