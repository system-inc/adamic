// Package registry discovers rule-owned descriptors and generates static dispatch.
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	syntax "github.com/microsoft/TypeScript/tsc/shim/ast"
)

type Descriptor struct {
	Parent bool `json:"parent,omitempty"`
	// Node opts a rule into being handed the node it listens to, as `visit(node, index[, parent])`, rather
	// than refetching it from its index. Every rule moves to it in #93z4yv7's codemod, and then the field
	// goes: until then a rule without it keeps `visit(index[, parent])`.
	Node            bool     `json:"node,omitempty"`
	Factory         string   `json:"factory"`
	Class           string   `json:"class"`
	Name            string   `json:"name"`
	Order           int      `json:"order,omitempty"`
	Kinds           []string `json:"kinds"`
	Visit           string   `json:"visit"`
	Prepare         string   `json:"prepare,omitempty"`
	Finish          string   `json:"finish,omitempty"`
	Oracle          string   `json:"oracle"`
	UpstreamPackage string   `json:"upstreamPackage"`
	UpstreamTest    string   `json:"upstreamTest"`
	Slug            string   `json:"-"`
	Module          string   `json:"-"`
}

var identifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
var publicName = regexp.MustCompile(`^(?:@?[A-Za-z0-9_-]+/)*[A-Za-z0-9_-]+$`)
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// RuleModule requires exactly one entry module, so stale renames cannot select silently.
func RuleModule(directory string) (string, error) {
	var module string
	for _, name := range []string{"rule.ts", "rule.a"} {
		if _, statError := os.Stat(filepath.Join(directory, name)); statError == nil {
			if module != "" {
				return "", fmt.Errorf("%s: ambiguous rule.ts and rule.a", directory)
			}
			module = name
		} else if !os.IsNotExist(statError) {
			return "", statError
		}
	}
	if module == "" {
		return "", fmt.Errorf("%s: missing rule.ts or rule.a", directory)
	}
	return module, nil
}

// Witnesses are raw source files outside the module graph. Keep their script kind. A witness may sit in a
// directory under testdata, as testdata/src/utils/format.ts.txt, when its rule judges the path: the
// harness lints it at that relative path.
func Witnesses(directory string) ([]string, error) {
	var paths []string
	root := filepath.Join(directory, "testdata")
	walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			if os.IsNotExist(pathError) && path == root {
				return filepath.SkipDir
			}
			return pathError
		}
		if entry.IsDir() {
			return nil
		}
		for _, extension := range []string{"ts", "tsx", "js", "jsx"} {
			if strings.HasSuffix(path, "."+extension+".txt") {
				paths = append(paths, path)
			}
		}
		return nil
	})
	if walkError != nil {
		return nil, walkError
	}
	sort.Strings(paths)
	return paths, nil
}

// CheckWitnesses refuses a rule with no witness under its testdata/. It walks every rule's testdata/, so it isn't part
// of Discover, which lint's TestMain runs (Stale) and most of its tests run (Generate) in every process: a listing of
// every rule's testdata/ there entered the read set of every lint test after it, and a witness added to one rule moved
// every lint test's key (#s57yq6y). go run ./cmd/lint-registry checks it, and so does TestEveryRuleHasAWitness.
func CheckWitnesses(root string, descriptors []Descriptor) error {
	for _, descriptor := range descriptors {
		directory := filepath.Join(root, "rules", descriptor.Slug)
		witnesses, witnessesError := Witnesses(directory)
		if witnessesError != nil || len(witnesses) == 0 {
			return fmt.Errorf("%s: missing witness", filepath.Join(directory, "rule.json"))
		}
	}
	return nil
}

// Discover reads each rule's descriptor and the files it names, all directly in the rule's directory, and nothing
// below it: a rule's evidence/ and testdata/ are never its descriptor's inputs.
func Discover(root string) ([]Descriptor, error) {
	paths, globError := filepath.Glob(filepath.Join(root, "rules", "*", "rule.json"))
	if globError != nil {
		return nil, globError
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no rule descriptors in %s", root)
	}
	names := map[string]bool{}
	adapters := map[string]bool{}
	entries, readDirectoryError := os.ReadDir(filepath.Join(root, "rules"))
	if readDirectoryError != nil {
		return nil, readDirectoryError
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, statError := os.Stat(filepath.Join(root, "rules", entry.Name(), "rule.json")); statError != nil {
				return nil, fmt.Errorf("%s: missing rule descriptor", entry.Name())
			}
		}
	}
	var result []Descriptor
	for _, path := range paths {
		data, readError := os.ReadFile(path)
		if readError != nil {
			return nil, readError
		}
		var descriptor Descriptor
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if decodeError := decoder.Decode(&descriptor); decodeError != nil {
			return nil, fmt.Errorf("%s: %w", path, decodeError)
		}
		if decodeError := decoder.Decode(new(any)); decodeError != io.EOF {
			return nil, fmt.Errorf("%s: trailing JSON", path)
		}
		descriptor.Slug = filepath.Base(filepath.Dir(path))
		if !slugPattern.MatchString(descriptor.Slug) ||
			!publicName.MatchString(descriptor.Name) ||
			descriptor.Name == "all" ||
			names[descriptor.Name] {
			return nil, fmt.Errorf("%s: invalid or duplicate rule name %q", path, descriptor.Name)
		}
		names[descriptor.Name] = true
		if adapters[descriptor.Oracle] {
			return nil, fmt.Errorf("%s: duplicate oracle adapter %q", path, descriptor.Oracle)
		}
		adapters[descriptor.Oracle] = true
		if descriptor.Order < 0 ||
			!identifier.MatchString(descriptor.Oracle) ||
			descriptor.UpstreamPackage == "" ||
			!identifier.MatchString(descriptor.UpstreamTest) {
			return nil, fmt.Errorf("%s: invalid oracle or provenance", path)
		}
		for _, part := range strings.Split(descriptor.UpstreamPackage, "/") {
			if !slugPattern.MatchString(part) {
				return nil, fmt.Errorf("%s: invalid upstream package", path)
			}
		}
		var ruleModuleError error
		descriptor.Module, ruleModuleError = RuleModule(filepath.Dir(path))
		if ruleModuleError != nil {
			return nil, ruleModuleError
		}
		module, readError := os.ReadFile(filepath.Join(filepath.Dir(path), descriptor.Module))
		if readError != nil {
			return nil, readError
		}
		if !identifier.MatchString(descriptor.Factory) ||
			!identifier.MatchString(descriptor.Class) ||
			!strings.Contains(string(module), "export function "+descriptor.Factory+"(") ||
			!strings.Contains(string(module), "export class "+descriptor.Class+" {") {
			return nil, fmt.Errorf("%s: missing named factory or class", path)
		}
		for _, hook := range []string{descriptor.Visit, descriptor.Prepare, descriptor.Finish} {
			if hook == "" {
				continue
			}
			if !identifier.MatchString(hook) ||
				!regexp.MustCompile(`(?m)^\s+`+regexp.QuoteMeta(hook)+`\(`).Match(module) {
				return nil, fmt.Errorf("%s: missing named hook %q", path, hook)
			}
		}
		if descriptor.Visit == "" || len(descriptor.Kinds) == 0 {
			return nil, fmt.Errorf("%s: visit and kinds required", path)
		}
		seen := map[string]bool{}
		// SyntaxKind names come from the pinned upstream AST, not a second maintained list.
		kinds := map[string]bool{}
		for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
			kinds[strings.TrimPrefix(kind.String(), "Kind")] = true
		}
		for _, kind := range descriptor.Kinds {
			if !identifier.MatchString(kind) || seen[kind] || !kinds[kind] {
				return nil, fmt.Errorf("%s: invalid or duplicate kind %q", path, kind)
			}
			seen[kind] = true
		}
		adapter, parseFileError := parser.ParseFile(
			token.NewFileSet(),
			filepath.Join(filepath.Dir(path), "oracle.go"),
			nil,
			0,
		)
		if parseFileError != nil {
			return nil, parseFileError
		}
		functions := map[string]bool{}
		for _, decl := range adapter.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				functions[funcDecl.Name.Name] = true
			}
		}
		if adapter.Name.Name != "main" || !functions[descriptor.Oracle] || !functions[descriptor.Oracle+"Options"] {
			return nil, fmt.Errorf("%s: missing oracle adapter exports", path)
		}
		var mutant struct{ Name, File, From, To string }
		mutation, readError := os.ReadFile(filepath.Join(filepath.Dir(path), "mutant.json"))
		if readError != nil {
			return nil, readError
		}
		mutationDecoder := json.NewDecoder(bytes.NewReader(mutation))
		mutationDecoder.DisallowUnknownFields()
		if decodeError := mutationDecoder.Decode(&mutant); decodeError != nil {
			return nil, decodeError
		}
		if mutant.File == "" {
			mutant.File = descriptor.Module
		}
		if filepath.IsAbs(mutant.File) ||
			filepath.Clean(mutant.File) == ".." ||
			strings.HasPrefix(filepath.Clean(mutant.File), ".."+string(filepath.Separator)) ||
			!(strings.HasSuffix(mutant.File, ".ts") || strings.HasSuffix(mutant.File, ".a")) {
			return nil, fmt.Errorf("%s: mutant file must be an owned .ts or .a module", path)
		}
		if _, statError := os.Stat(filepath.Join(filepath.Dir(path), mutant.File)); statError != nil {
			return nil, statError
		}
		if mutant.Name == "" || mutant.From == "" || mutant.To == mutant.From {
			return nil, fmt.Errorf("%s: invalid owned mutant", path)
		}
		result = append(result, descriptor)
	}
	sort.Slice(result, func(left, right int) bool {
		a, b := result[left], result[right]
		if a.Order != b.Order {
			if a.Order == 0 {
				return false
			}
			if b.Order == 0 {
				return true
			}
			return a.Order < b.Order
		}
		return a.Name < b.Name
	})
	return result, nil
}

func Render(descriptors []Descriptor) (typescript, golang []byte) {
	var ts, goSource strings.Builder
	ts.WriteString(
		"// Generated by go run ./cmd/lint-registry. Do not edit: regenerate, and commit the result.\nimport type { RuleContext } from '../context.ts';\n",
	)
	goSource.WriteString(
		"// Generated by go run ./cmd/lint-registry. Do not edit: regenerate, and commit the result.\npackage main\nimport \"github.com/system-inc/cohere/internal/lint/rule\"\ntype registeredRule struct { subject rule.Rule; options func([]string) any }\nfunc registeredRules() []registeredRule { return []registeredRule{\n",
	)
	for index, descriptor := range descriptors {
		if descriptor.Module == "" {
			descriptor.Module = "rule.ts"
		}
		fmt.Fprintf(
			&ts,
			"import { %s as create%d, type %s as Rule%d } from '../rules/%s/%s';\n",
			descriptor.Factory,
			index,
			descriptor.Class,
			index,
			descriptor.Slug,
			descriptor.Module,
		)
		fmt.Fprintf(&goSource, "{%s(), %sOptions},\n", descriptor.Oracle, descriptor.Oracle)
	}
	// The generated driver is a port of cohere's fused walk (internal/types/program/walk.go): one walk per
	// file, a rule called only on the kinds it declares, and the selection decided once per file rather
	// than per node. cohere indexes a table of listeners by kind; here the table is a generated switch
	// with direct calls, which measured cheaper in native Adamic than calling through function values
	// (#93z4yv7). selectedN is whether rule N runs on this file, read once from the context.
	ts.WriteString("export class RuleSet {\n    readonly context: RuleContext;\n")
	for index := range descriptors {
		fmt.Fprintf(&ts, "    readonly rule%d: Rule%d;\n", index, index)
	}
	for index := range descriptors {
		fmt.Fprintf(&ts, "    readonly selected%d: boolean;\n", index)
	}
	ts.WriteString("    constructor(context: RuleContext")
	for index := range descriptors {
		fmt.Fprintf(&ts, ", rule%d: Rule%d", index, index)
	}
	ts.WriteString(") {\n        this.context = context;\n")
	for index := range descriptors {
		fmt.Fprintf(&ts, "        this.rule%d = rule%d;\n", index, index)
	}
	for index, descriptor := range descriptors {
		fmt.Fprintf(&ts, "        this.selected%d = context.enabled('%s');\n", index, descriptor.Name)
	}
	ts.WriteString("    }\n")
	goSource.WriteString("} }\n")
	for _, hook := range []string{"prepare", "finish"} {
		fmt.Fprintf(&ts, "    %s(root: number): void {\n", hook)
		for index, descriptor := range descriptors {
			name := descriptor.Prepare
			if hook == "finish" {
				name = descriptor.Finish
			}
			if name != "" {
				fmt.Fprintf(&ts, "    if(this.selected%d) { this.rule%d.%s(root); }\n", index, index, name)
			}
		}
		ts.WriteString("}\n")
	}
	ts.WriteString(
		"    visit(index: number, parent: number): void {\n        const node = this.context.node(index);\n        switch(node.kind) {\n",
	)
	buckets := map[string][]int{}
	for index, descriptor := range descriptors {
		for _, kind := range descriptor.Kinds {
			buckets[kind] = append(buckets[kind], index)
		}
	}
	var kinds []string
	for kind := range buckets {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		fmt.Fprintf(&ts, "        case '%s':\n", kind)
		for _, index := range buckets[kind] {
			descriptor := descriptors[index]
			arguments := "index"
			if descriptor.Node {
				arguments = "node, index"
			}
			if descriptor.Parent {
				arguments += ", parent"
			}
			fmt.Fprintf(
				&ts,
				"            if(this.selected%d) { this.rule%d.%s(%s); }\n",
				index,
				index,
				descriptor.Visit,
				arguments,
			)
		}
		ts.WriteString("            break;\n")
	}
	ts.WriteString("        default:\n            break;\n    }\n    }\n}\n")
	ts.WriteString("export function createRuleSet(context: RuleContext): RuleSet {\n")
	for index := range descriptors {
		fmt.Fprintf(&ts, "    const rule%d = create%d(context);\n", index, index)
	}
	ts.WriteString("    return new RuleSet(context")
	for index := range descriptors {
		fmt.Fprintf(&ts, ", rule%d", index)
	}
	ts.WriteString(");\n}\n")
	formatted, formatError := format.Source([]byte(goSource.String()))
	if formatError != nil {
		panic(fmt.Sprintf("registry generator produced invalid Go: %v", formatError))
	}
	return []byte(ts.String()), formatted
}

// Stale says whether root's committed .generated registry is what its descriptors render, without writing: a test
// reads the tree and never changes it (#k01pgng: a Loom runner's source is read-only, and a product that compiled an
// untracked registry keyed apart from Workshop's). A stale or missing file is named, with the command that fixes it.
func Stale(root string) ([]Descriptor, error) {
	descriptors, discoverError := Discover(root)
	if discoverError != nil {
		return nil, discoverError
	}
	ts, goSource := Render(descriptors)
	var stale []string
	for _, file := range []struct {
		name string
		data []byte
	}{{"registry.ts", ts}, {"registry.go", goSource}} {
		if existing, readError := os.ReadFile(
			filepath.Join(root, ".generated", file.name),
		); readError != nil || !bytes.Equal(existing, file.data) {
			stale = append(stale, filepath.Join(root, ".generated", file.name))
		}
	}
	if len(stale) != 0 {
		return nil, fmt.Errorf(
			"the lint registry is stale (%s): run go run ./cmd/lint-registry and commit .generated",
			strings.Join(stale, ", "),
		)
	}
	return descriptors, nil
}

// Generate validates the entire registry even when a test selects only one rule.
// Atomic replacement and unchanged-byte checks make concurrent test generation safe.
func Generate(root string) ([]Descriptor, error) {
	descriptors, discoverError := Discover(root)
	if discoverError != nil {
		return nil, discoverError
	}
	ts, goSource := Render(descriptors)
	directory := filepath.Join(root, ".generated")
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		return nil, createDirectoryError
	}
	for _, file := range []struct {
		name string
		data []byte
	}{{"registry.ts", ts}, {"registry.go", goSource}} {
		path := filepath.Join(directory, file.name)
		if existing, readError := os.ReadFile(path); readError == nil && bytes.Equal(existing, file.data) {
			continue
		}
		temporary, temporaryFileError := os.CreateTemp(directory, "registry-")
		if temporaryFileError != nil {
			return nil, temporaryFileError
		}
		if _, writeError := temporary.Write(file.data); writeError != nil {
			temporary.Close()
			os.Remove(temporary.Name())
			return nil, writeError
		}
		if closeError := temporary.Close(); closeError != nil {
			return nil, closeError
		}
		if renameError := os.Rename(temporary.Name(), path); renameError != nil {
			return nil, renameError
		}
	}
	return descriptors, nil
}
