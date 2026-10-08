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
		if _, err := os.Stat(filepath.Join(directory, name)); err == nil {
			if module != "" {
				return "", fmt.Errorf("%s: ambiguous rule.ts and rule.a", directory)
			}
			module = name
		} else if !os.IsNotExist(err) {
			return "", err
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
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == root {
				return filepath.SkipDir
			}
			return err
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
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func Discover(root string) ([]Descriptor, error) {
	paths, err := filepath.Glob(filepath.Join(root, "rules", "*", "rule.json"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no rule descriptors in %s", root)
	}
	names := map[string]bool{}
	adapters := map[string]bool{}
	entries, err := os.ReadDir(filepath.Join(root, "rules"))
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(root, "rules", entry.Name(), "rule.json")); err != nil {
				return nil, fmt.Errorf("%s: missing rule descriptor", entry.Name())
			}
		}
	}
	var result []Descriptor
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var d Descriptor
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&d); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, fmt.Errorf("%s: trailing JSON", path)
		}
		d.Slug = filepath.Base(filepath.Dir(path))
		if !slugPattern.MatchString(d.Slug) || !publicName.MatchString(d.Name) || d.Name == "all" || names[d.Name] {
			return nil, fmt.Errorf("%s: invalid or duplicate rule name %q", path, d.Name)
		}
		names[d.Name] = true
		if adapters[d.Oracle] {
			return nil, fmt.Errorf("%s: duplicate oracle adapter %q", path, d.Oracle)
		}
		adapters[d.Oracle] = true
		if d.Order < 0 || !identifier.MatchString(d.Oracle) || d.UpstreamPackage == "" || !identifier.MatchString(d.UpstreamTest) {
			return nil, fmt.Errorf("%s: invalid oracle or provenance", path)
		}
		for _, part := range strings.Split(d.UpstreamPackage, "/") {
			if !slugPattern.MatchString(part) {
				return nil, fmt.Errorf("%s: invalid upstream package", path)
			}
		}
		d.Module, err = RuleModule(filepath.Dir(path))
		if err != nil {
			return nil, err
		}
		module, err := os.ReadFile(filepath.Join(filepath.Dir(path), d.Module))
		if err != nil {
			return nil, err
		}
		if !identifier.MatchString(d.Factory) || !identifier.MatchString(d.Class) || !strings.Contains(string(module), "export function "+d.Factory+"(") || !strings.Contains(string(module), "export class "+d.Class+" {") {
			return nil, fmt.Errorf("%s: missing named factory or class", path)
		}
		for _, hook := range []string{d.Visit, d.Prepare, d.Finish} {
			if hook == "" {
				continue
			}
			if !identifier.MatchString(hook) || !regexp.MustCompile(`(?m)^\s+`+regexp.QuoteMeta(hook)+`\(`).Match(module) {
				return nil, fmt.Errorf("%s: missing named hook %q", path, hook)
			}
		}
		if d.Visit == "" || len(d.Kinds) == 0 {
			return nil, fmt.Errorf("%s: visit and kinds required", path)
		}
		seen := map[string]bool{}
		// SyntaxKind names come from the pinned upstream AST, not a second maintained list.
		kinds := map[string]bool{}
		for kind := syntax.Kind(0); kind < syntax.KindCount; kind++ {
			kinds[strings.TrimPrefix(kind.String(), "Kind")] = true
		}
		for _, kind := range d.Kinds {
			if !identifier.MatchString(kind) || seen[kind] || !kinds[kind] {
				return nil, fmt.Errorf("%s: invalid or duplicate kind %q", path, kind)
			}
			seen[kind] = true
		}
		adapter, err := parser.ParseFile(token.NewFileSet(), filepath.Join(filepath.Dir(path), "oracle.go"), nil, 0)
		if err != nil {
			return nil, err
		}
		functions := map[string]bool{}
		for _, decl := range adapter.Decls {
			if f, ok := decl.(*ast.FuncDecl); ok {
				functions[f.Name.Name] = true
			}
		}
		if adapter.Name.Name != "main" || !functions[d.Oracle] || !functions[d.Oracle+"Options"] {
			return nil, fmt.Errorf("%s: missing oracle adapter exports", path)
		}
		var mutant struct{ Name, File, From, To string }
		mutation, err := os.ReadFile(filepath.Join(filepath.Dir(path), "mutant.json"))
		if err != nil {
			return nil, err
		}
		mutationDecoder := json.NewDecoder(bytes.NewReader(mutation))
		mutationDecoder.DisallowUnknownFields()
		if err := mutationDecoder.Decode(&mutant); err != nil {
			return nil, err
		}
		if mutant.File == "" {
			mutant.File = d.Module
		}
		if filepath.IsAbs(mutant.File) || filepath.Clean(mutant.File) == ".." || strings.HasPrefix(filepath.Clean(mutant.File), ".."+string(filepath.Separator)) || !(strings.HasSuffix(mutant.File, ".ts") || strings.HasSuffix(mutant.File, ".a")) {
			return nil, fmt.Errorf("%s: mutant file must be an owned .ts or .a module", path)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(path), mutant.File)); err != nil {
			return nil, err
		}
		if mutant.Name == "" || mutant.From == "" || mutant.To == mutant.From {
			return nil, fmt.Errorf("%s: invalid owned mutant", path)
		}
		witnesses, err := Witnesses(filepath.Dir(path))
		if err != nil || len(witnesses) == 0 {
			return nil, fmt.Errorf("%s: missing witness", path)
		}
		result = append(result, d)
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
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
	ts.WriteString("// Generated by lint-registry. Do not commit or edit.\nimport type { RuleContext } from '../context.ts';\n")
	goSource.WriteString("// Generated by lint-registry.\npackage main\nimport \"github.com/system-inc/cohere/internal/lint/rule\"\ntype registeredRule struct { subject rule.Rule; options func([]string) any }\nfunc registeredRules() []registeredRule { return []registeredRule{\n")
	for i, d := range descriptors {
		if d.Module == "" {
			d.Module = "rule.ts"
		}
		fmt.Fprintf(&ts, "import { %s as create%d, type %s as Rule%d } from '../rules/%s/%s';\n", d.Factory, i, d.Class, i, d.Slug, d.Module)
		fmt.Fprintf(&goSource, "{%s(), %sOptions},\n", d.Oracle, d.Oracle)
	}
	// The generated driver is a port of cohere's fused walk (internal/types/program/walk.go): one walk per
	// file, a rule called only on the kinds it declares, and the selection decided once per file rather
	// than per node. cohere indexes a table of listeners by kind; here the table is a generated switch
	// with direct calls, which measured cheaper in native Adamic than calling through function values
	// (#93z4yv7). selectedN is whether rule N runs on this file, read once from the context.
	ts.WriteString("export class RuleSet {\n    readonly context: RuleContext;\n")
	for i := range descriptors {
		fmt.Fprintf(&ts, "    readonly rule%d: Rule%d;\n", i, i)
	}
	for i := range descriptors {
		fmt.Fprintf(&ts, "    readonly selected%d: boolean;\n", i)
	}
	ts.WriteString("    constructor(context: RuleContext")
	for i := range descriptors {
		fmt.Fprintf(&ts, ", rule%d: Rule%d", i, i)
	}
	ts.WriteString(") {\n        this.context = context;\n")
	for i := range descriptors {
		fmt.Fprintf(&ts, "        this.rule%d = rule%d;\n", i, i)
	}
	for i, d := range descriptors {
		fmt.Fprintf(&ts, "        this.selected%d = context.enabled('%s');\n", i, d.Name)
	}
	ts.WriteString("    }\n")
	goSource.WriteString("} }\n")
	for _, hook := range []string{"prepare", "finish"} {
		fmt.Fprintf(&ts, "    %s(root: number): void {\n", hook)
		for i, d := range descriptors {
			name := d.Prepare
			if hook == "finish" {
				name = d.Finish
			}
			if name != "" {
				fmt.Fprintf(&ts, "    if(this.selected%d) { this.rule%d.%s(root); }\n", i, i, name)
			}
		}
		ts.WriteString("}\n")
	}
	ts.WriteString("    visit(index: number, parent: number): void {\n        const node = this.context.node(index);\n        switch(node.kind) {\n")
	buckets := map[string][]int{}
	for i, d := range descriptors {
		for _, kind := range d.Kinds {
			buckets[kind] = append(buckets[kind], i)
		}
	}
	var kinds []string
	for kind := range buckets {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		fmt.Fprintf(&ts, "        case '%s':\n", kind)
		for _, i := range buckets[kind] {
			d := descriptors[i]
			arguments := "index"
			if d.Node {
				arguments = "node, index"
			}
			if d.Parent {
				arguments += ", parent"
			}
			fmt.Fprintf(&ts, "            if(this.selected%d) { this.rule%d.%s(%s); }\n", i, i, d.Visit, arguments)
		}
		ts.WriteString("            break;\n")
	}
	ts.WriteString("        default:\n            break;\n    }\n    }\n}\n")
	ts.WriteString("export function createRuleSet(context: RuleContext): RuleSet {\n")
	for i := range descriptors {
		fmt.Fprintf(&ts, "    const rule%d = create%d(context);\n", i, i)
	}
	ts.WriteString("    return new RuleSet(context")
	for i := range descriptors {
		fmt.Fprintf(&ts, ", rule%d", i)
	}
	ts.WriteString(");\n}\n")
	formatted, err := format.Source([]byte(goSource.String()))
	if err != nil {
		panic(fmt.Sprintf("registry generator produced invalid Go: %v", err))
	}
	return []byte(ts.String()), formatted
}

// Generate validates the entire registry even when a test selects only one rule.
// Atomic replacement and unchanged-byte checks make concurrent test generation safe.
func Generate(root string) ([]Descriptor, error) {
	descriptors, err := Discover(root)
	if err != nil {
		return nil, err
	}
	ts, goSource := Render(descriptors)
	directory := filepath.Join(root, ".generated")
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}
	for _, file := range []struct {
		name string
		data []byte
	}{{"registry.ts", ts}, {"registry.go", goSource}} {
		path := filepath.Join(directory, file.name)
		if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, file.data) {
			continue
		}
		temporary, err := os.CreateTemp(directory, "registry-")
		if err != nil {
			return nil, err
		}
		if _, err = temporary.Write(file.data); err != nil {
			temporary.Close()
			os.Remove(temporary.Name())
			return nil, err
		}
		if err = temporary.Close(); err != nil {
			return nil, err
		}
		if err = os.Rename(temporary.Name(), path); err != nil {
			return nil, err
		}
	}
	return descriptors, nil
}
