// Host-only persistent transport, compiled through the oracle overlay.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/lint/registry"
	"github.com/system-inc/cohere/internal/lint/rule"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
	"unicode/utf8"
)

type request struct {
	Op, Path, Root, Rule, Options string
	Source                        []byte
}
type answer struct {
	OutputRaw []byte   `json:"stdout_bytes,omitempty"`
	Reusable  bool     `json:"reusable"`
	Output    string   `json:"stdout"`
	Error     string   `json:"error,omitempty"`
	Stderr    string   `json:"stderr,omitempty"`
	Stack     string   `json:"stack,omitempty"`
	Exit      int      `json:"exit_code"`
	NS        int64    `json:"elapsed_ns"`
	Valid     bool     `json:"valid"`
	Spans     [][2]int `json:"spans,omitempty"`
	Rows      []Row    `json:"rows,omitempty"`
}

var activeRequest request

func scoreboardReadFile(path string) ([]byte, error) {
	if path == activeRequest.Path {
		return activeRequest.Source, nil
	}
	return os.ReadFile(path)
}
func handle(q request) (a answer) {
	a.Reusable = true
	start := time.Now()
	var output bytes.Buffer
	writer := bufio.NewWriter(&output)
	defer func() {
		writer.Flush()
		a.Output = output.String()
		if !utf8.Valid(output.Bytes()) {
			a.OutputRaw = append([]byte{}, output.Bytes()...)
		}
		a.NS = time.Since(start).Nanoseconds()
		if p := recover(); p != nil {
			a.Error = fmt.Sprint(p)
			a.Exit = 2
			a.Stack = string(debug.Stack())
		}
	}()
	activeRequest = q
	switch q.Op {
	case "lint":
		a.Valid = len(parse(q.Path, string(q.Source)).Diagnostics()) == 0
		row := q.Path + "\t" + q.Rule + "\t\t\t\t" + q.Options
		fmt.Fprintln(writer, "case 0")
		run(row, false, writer, nil)
	case "format":
		opts := formatoptions.Default()
		ext := filepath.Ext(strings.TrimSuffix(q.Path, ".txt"))
		if ext == ".json" || ext == ".yaml" || ext == ".yml" {
			opts = formatoptions.PrettierDefaults()
		}
		text, err := (native.Formatter{Options: opts}).Format(strings.TrimSuffix(q.Path, ".txt"), string(q.Source))
		if err != nil {
			fmt.Fprintln(writer, "error\t"+escape(err.Error()))
		} else {
			fmt.Fprintln(writer, "ok\t"+escape(text))
		}
	case "spans":
		file := parse(q.Path, string(q.Source))
		a.Valid = len(file.Diagnostics()) == 0
		if a.Valid {
			var walk func(*ast.Node)
			walk = func(n *ast.Node) {
				if n.Kind != ast.KindSourceFile {
					start := rule.TokenRange(file, n).Pos()
					end := n.End()
					if start >= 0 && end > start && end <= len(q.Source) {
						a.Spans = append(a.Spans, [2]int{start, end})
					}
				}
				n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
			}
			walk(file.AsNode())
		}
	case "valid":
		a.Valid = len(parse(q.Path, string(q.Source)).Diagnostics()) == 0
	case "census":
		a.Rows = census(q)
	case "census-reference":
		_ = registry.All()
		file := parse(q.Path, string(q.Source))
		for _, reg := range rule.Registered() {
			n, status := count(reg, file, nil, nil, q.Root)
			row := Row{Rule: reg.Rule.Name, Findings: n, Status: status}
			if n > 0 {
				row.Files = 1
			}
			a.Rows = append(a.Rows, row)
		}
	default:
		panic("unknown worker operation " + q.Op)
	}
	return
}

// Each rule gets its own context/cache and its own preorder sequence, as count
// does. Bucketing listeners changes inter-rule scheduling, never a rule's order.
// A failed listener disables only that rule and discards its partial count.
func census(q request) []Row {
	_ = registry.All()
	regs := rule.Registered()
	rows := make([]Row, len(regs))
	file := parse(q.Path, string(q.Source))
	type callback struct {
		index int
		visit func(*ast.Node)
	}
	buckets := map[ast.Kind][]callback{}
	for i, reg := range regs {
		rows[i] = Row{Rule: reg.Rule.Name, Status: "complete"}
		if len(file.Diagnostics()) > 0 {
			rows[i].Status = "parse diagnostics"
			continue
		}
		if reg.Rule.NeedsTypeChecker {
			rows[i].Status = "no program"
			continue
		}
		if reg.RequiresOptions {
			rows[i].Status = "requires options"
			continue
		}
		func() {
			defer func() {
				if p := recover(); p != nil {
					rows[i].Status = fmt.Sprintf("panic: %v", p)
					rows[i].Findings = 0
				}
			}()
			var opts any
			var err error
			switch {
			case reg.DecodeAt != nil:
				opts, err = reg.DecodeAt(nil, rule.OptionsBase{ConfigDirectory: q.Root, ProjectRoot: q.Root})
			case reg.DecodeOptionList != nil:
				opts, err = reg.DecodeOptionList(nil)
			case reg.Decode != nil:
				opts, err = reg.Decode(nil)
			}
			if err != nil {
				rows[i].Status = "option decode: " + err.Error()
				return
			}
			ctx := rule.Context{SourceFile: file, FileCache: rule.NewFileCache(), Report: func(d rule.Diagnostic) { rows[i].Findings++ }}
			for k, visit := range reg.Rule.Run(ctx, opts) {
				buckets[k] = append(buckets[k], callback{i, visit})
			}
		}()
	}
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		for _, cb := range buckets[n.Kind] {
			if rows[cb.index].Status != "complete" {
				continue
			}
			func() {
				defer func() {
					if p := recover(); p != nil {
						rows[cb.index].Status = fmt.Sprintf("panic: %v", p)
						rows[cb.index].Findings = 0
					}
				}()
				cb.visit(n)
			}()
		}
		n.ForEachChild(func(c *ast.Node) bool { walk(c); return false })
	}
	if len(file.Diagnostics()) == 0 {
		walk(file.AsNode())
	}
	for i := range rows {
		if rows[i].Findings > 0 {
			rows[i].Files = 1
		}
	}
	return rows
}
func main() {
	decoder := json.NewDecoder(bufio.NewReader(os.Stdin))
	encoder := json.NewEncoder(os.Stdout)
	for {
		var q request
		if err := decoder.Decode(&q); err == io.EOF {
			return
		} else if err != nil {
			panic(err)
		}
		if err := encoder.Encode(handle(q)); err != nil {
			panic(err)
		}
	}
}
