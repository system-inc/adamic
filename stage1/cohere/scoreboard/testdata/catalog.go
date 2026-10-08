// Compiled through an overlay in cohere, so all internal APIs remain the oracle's.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
	"os"
	"path/filepath"
	"strings"
)

type Snapshot struct {
	Root string `json:"root"`
}

type Manifest struct {
	Snapshots []Snapshot                 `json:"snapshots"`
	Files     []string                   `json:"files"`
	Program   string                     `json:"program"`
	Options   map[string]json.RawMessage `json:"options"`
}
type Row struct {
	Rule     string `json:"rule"`
	Findings int    `json:"findings"`
	Files    int    `json:"files"`
	Status   string `json:"status"`
}

func kind(path string) core.ScriptKind {
	switch filepath.Ext(strings.TrimSuffix(path, ".txt")) {
	case ".js":
		return core.ScriptKindJS
	case ".jsx":
		return core.ScriptKindJSX
	case ".tsx":
		return core.ScriptKindTSX
	}
	return core.ScriptKindTS
}
func lintFile(path string) bool {
	switch filepath.Ext(strings.TrimSuffix(path, ".txt")) {
	case ".ts", ".a", ".js", ".jsx", ".tsx":
		return true
	}
	return false
}
func count(reg rule.Registration, file *ast.SourceFile, graph *program.Graph, raw []byte, root string) (n int, status string) {
	status = "complete"
	defer func() {
		if p := recover(); p != nil {
			n = 0
			status = fmt.Sprintf("panic: %v", p)
		}
	}()
	if len(file.Diagnostics()) > 0 {
		return 0, "parse diagnostics"
	}
	if reg.Rule.NeedsTypeChecker && graph == nil {
		return 0, "no program"
	}
	if reg.RequiresOptions && len(raw) == 0 {
		return 0, "requires options"
	}
	var options any
	var err error
	switch {
	case reg.DecodeAt != nil:
		options, err = reg.DecodeAt(raw, rule.OptionsBase{ConfigDirectory: root, ProjectRoot: root})
	case reg.DecodeOptionList != nil:
		options, err = reg.DecodeOptionList(raw)
	case reg.Decode != nil:
		options, err = reg.Decode(raw)
	}
	if err != nil {
		return 0, "option decode: " + err.Error()
	}
	ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { n++ }}
	if graph != nil {
		checker, release := graph.CheckerForFile(context.Background(), file)
		defer release()
		ctx.TypeChecker = checker
		ctx.Program = rule.ViewProgram(graph.Program, file, reg.Rule)
	}
	listeners := reg.Rule.Run(ctx, options)
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if f := listeners[node.Kind]; f != nil {
			f(node)
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(file.AsNode())
	return
}
func main() {
	if os.Args[1] == "--valid" {
		path := os.Args[2]
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(path), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(path))}, string(data), kind(path))
		if len(file.Diagnostics()) != 0 {
			os.Exit(1)
		}
		return
	}

	if os.Args[1] == "--format" {
		path := os.Args[2]
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		opts := formatoptions.Default()
		if filepath.Ext(path) == ".json" || filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml" {
			opts = formatoptions.PrettierDefaults()
		}
		text, err := (native.Formatter{Options: opts}).Format(strings.TrimSuffix(path, ".txt"), string(data))
		if err != nil {
			fmt.Println("error\t" + escape(err.Error()))
		} else {
			fmt.Println("ok\t" + escape(text))
		}
		return
	}
	_ = registry.All()
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var m Manifest
	if err = json.Unmarshal(data, &m); err != nil {
		panic(err)
	}
	root := os.Args[2]
	var graph *program.Graph
	if m.Program != "" {
		config := m.Program
		if !filepath.IsAbs(config) {
			config = filepath.Join(root, config)
		}
		graph, err = program.Build(program.Options{ConfigFileName: config, CurrentDirectory: filepath.Dir(config), SingleThreaded: true})
		if err != nil {
			panic(err)
		}
	}
	registrations := rule.Registered()
	rows := make([]Row, len(registrations))
	for i, r := range registrations {
		rows[i] = Row{Rule: r.Rule.Name, Status: "complete"}
	}
	for _, path := range m.Files {
		if !lintFile(path) {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		data, err = os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(path), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(path))}, string(data), kind(path))
		if graph != nil {
			file = graph.Program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
			if file == nil {
				panic("program omits source " + path)
			}
		}
		base := root
		for _, snapshot := range m.Snapshots {
			if strings.HasPrefix(path, filepath.Clean(snapshot.Root)+string(filepath.Separator)) {
				base = snapshot.Root
				break
			}
		}
		for i, reg := range registrations {
			n, status := count(reg, file, graph, m.Options[reg.Rule.Name], base)
			rows[i].Findings += n
			if n > 0 {
				rows[i].Files++
			}
			if status != "complete" {
				rows[i].Status = status
			}
		}
	}
	if err = json.NewEncoder(os.Stdout).Encode(rows); err != nil {
		panic(err)
	}
}
func escape(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(s)
}
