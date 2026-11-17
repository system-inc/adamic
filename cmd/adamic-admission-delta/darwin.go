package main

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/adamic/internal/platform"
)

// A platform stop has two independent keys: our parse of the revision's modules,
// and the compiler's typed stop. Neither a diagnostic's words nor a path grants it.
type platformJudge struct {
	root, base, head string
	revisions        map[string]*revisionImports
}

type revisionImports struct {
	root, revision string
	files          map[string]bool
	imports        map[string][]string
}

func (j *platformJudge) revision(sha string) (*revisionImports, error) {
	if r := j.revisions[sha]; r != nil {
		return r, nil
	}
	listing := execute(j.root, time.Minute, "git", "ls-tree", "-r", "--name-only", "-z", sha)
	if listing.Exit != 0 || listing.Error != "" {
		return nil, fmt.Errorf("darwin revision %s: %s %s", sha, listing.Error, listing.Stderr)
	}
	r := &revisionImports{root: j.root, revision: sha, files: map[string]bool{}, imports: map[string][]string{}}
	for _, file := range strings.Split(listing.Stdout, "\x00") {
		r.files[file] = true
	}
	j.revisions[sha] = r
	return r, nil
}

func (r *revisionImports) specifiers(file string) ([]string, error) {
	if imports, ok := r.imports[file]; ok {
		return imports, nil
	}
	source := execute(r.root, time.Minute, "git", "show", r.revision+":"+file)
	if source.Exit != 0 || source.Error != "" {
		return nil, fmt.Errorf("darwin source %s:%s: %s %s", r.revision, file, source.Error, source.Stderr)
	}
	name := "/" + file
	if strings.HasSuffix(name, ".a") {
		name += ".ts"
	}
	kind := core.ScriptKindTS
	if strings.HasSuffix(file, ".tsx") {
		kind = core.ScriptKindTSX
	}
	if strings.HasSuffix(file, ".js") || strings.HasSuffix(file, ".mjs") || strings.HasSuffix(file, ".cjs") {
		kind = core.ScriptKindJS
	}
	if strings.HasSuffix(file, ".jsx") {
		kind = core.ScriptKindJSX
	}
	parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(name), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(name))}, source.Stdout, kind)
	// The parser recovers through syntax errors. They are not a third key: a
	// typed platform stop leaves the whole Apple row for the darwin witnesses.
	var imports []string
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		var specifier *ast.Node
		switch node.Kind {
		case ast.KindImportDeclaration, ast.KindExportDeclaration:
			specifier = node.ModuleSpecifier()
		case ast.KindImportEqualsDeclaration:
			reference := node.AsImportEqualsDeclaration().ModuleReference
			if reference.Kind == ast.KindExternalModuleReference {
				specifier = reference.AsExternalModuleReference().Expression
			}
		case ast.KindCallExpression:
			call := node.AsCallExpression()
			if call.Expression.Kind == ast.KindImportKeyword && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				specifier = call.Arguments.Nodes[0]
			}
		case ast.KindImportType:
			argument := node.AsImportTypeNode().Argument
			if argument.Kind == ast.KindLiteralType {
				specifier = argument.AsLiteralTypeNode().Literal
			}
		}
		if specifier != nil && specifier.Kind == ast.KindStringLiteral {
			imports = append(imports, specifier.Text())
		}
		node.ForEachChild(visit)
		return false
	}
	visit(parsed.AsNode())
	r.imports[file] = imports
	return imports, nil
}

func (r *revisionImports) local(from, specifier string) string {
	file := path.Clean(path.Join(path.Dir(from), specifier))
	if file == ".." || strings.HasPrefix(file, "../") {
		return ""
	}
	// Match the loader's source spellings: X.a is visible as X.a.ts, never
	// guessed from an extensionless X. A real X.a.ts wins over that alias.
	candidates := []string{file}
	switch {
	case strings.HasSuffix(file, ".a"):
		candidates = []string{strings.TrimSuffix(file, ".a") + ".d.a.ts", file + ".ts", file}
	case strings.HasSuffix(file, ".a.ts"):
		candidates = []string{file, strings.TrimSuffix(file, ".ts")}
	case strings.HasSuffix(file, ".jsx"):
		stem := strings.TrimSuffix(file, ".jsx")
		candidates = []string{stem + ".tsx", stem + ".ts", stem + ".d.ts", file, stem + ".js"}
	case strings.HasSuffix(file, ".js"):
		stem := strings.TrimSuffix(file, ".js")
		candidates = []string{stem + ".ts", stem + ".tsx", stem + ".d.ts", file, stem + ".jsx"}
	case strings.HasSuffix(file, ".mjs"):
		stem := strings.TrimSuffix(file, ".mjs")
		candidates = []string{stem + ".mts", stem + ".d.mts", file}
	case strings.HasSuffix(file, ".cjs"):
		stem := strings.TrimSuffix(file, ".cjs")
		candidates = []string{stem + ".cts", stem + ".d.cts", file}
	case path.Ext(file) == "":
		candidates = []string{file + ".ts", file + ".tsx", file + ".d.ts", file + ".js", file + ".jsx", file + "/index.ts", file + "/index.tsx", file + "/index.d.ts", file + "/index.js", file + "/index.jsx"}
	}
	for _, candidate := range candidates {
		if r.files[candidate] {
			return candidate
		}
	}
	return ""
}

func (r *revisionImports) apple(file string) (bool, error) {
	seen := map[string]bool{}
	var walk func(string) (bool, error)
	walk = func(file string) (bool, error) {
		if seen[file] || !r.files[file] {
			return false, nil
		}
		seen[file] = true
		imports, err := r.specifiers(file)
		if err != nil {
			return false, err
		}
		for _, specifier := range imports {
			if strings.HasPrefix(specifier, "apple/") {
				return true, nil
			}
			if strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") {
				if local := r.local(file, specifier); local != "" {
					found, err := walk(local)
					if found || err != nil {
						return found, err
					}
				}
			}
		}
		return false, nil
	}
	return walk(file)
}

func platformStopped(o observation) bool {
	return o.Exit == platform.MissingExitCode && o.Error == ""
}

// stop makes the whole row unjudged, including observations already recorded.
// Only a judged base can move into darwin. The lazy native build is needed
// only when head stopped natively and the base file already reaches apple/.
func (j *platformJudge) stop(record *entry, o observation, baseBuild func() observation) (bool, error) {
	if !platformStopped(o) {
		return false, nil
	}
	head, err := j.revision(j.head)
	if err != nil {
		return false, err
	}
	apple, err := head.apple(record.Path)
	if err != nil {
		return false, err
	}
	record.AppleImport = apple
	if !apple {
		record.Class = "platform-stop-without-apple"
		record.PlatformReason = "exit 5 without a reachable apple/ import at head"
		return true, nil
	}
	base, err := j.revision(j.base)
	if err != nil {
		return false, err
	}
	if base.files[record.Path] {
		apple, err := base.apple(record.Path)
		if err != nil {
			return false, err
		}
		if apple {
			baseStopped := platformStopped(record.Base)
			if baseBuild != nil {
				observed := baseBuild()
				record.BaseNativeCompile = &observed
				baseStopped = platformStopped(observed)
			}
			if !baseStopped {
				record.Class = "moved-into-darwin"
				record.PlatformReason = "a reachable apple/ import was judged at base"
				return true, nil
			}
		}
	}
	record.Class = "darwin"
	record.VoidReason = "darwin: unjudged here; judged by the darwin witnesses"
	return true, nil
}

// A program leaving the platform stop is judged normally. Its recorded base
// observation stays intact; a typed base stop with the first key is not a crash.
func (j *platformJudge) leaving(record *entry) error {
	if !platformStopped(record.Base) {
		return nil
	}
	base, err := j.revision(j.base)
	if err != nil {
		return err
	}
	apple, err := base.apple(record.Path)
	if err != nil {
		return err
	}
	if apple {
		switch compileClass(record.Head) {
		case "accepted":
			record.Class = "newly-accepted"
		case "refused":
			record.Class = "refused-by-both"
		}
	}
	return nil
}
