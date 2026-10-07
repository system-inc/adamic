package tailwind

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
	"testing"
	"unicode/utf16"
)

func kernelWritten(text string) string {
	var out strings.Builder
	for _, r := range text {
		if r >= 32 && r <= 126 && r != 92 {
			out.WriteRune(r)
		} else if r <= 65535 {
			fmt.Fprintf(&out, `\u%04x`, r)
		} else {
			h, l := utf16.EncodeRune(r)
			fmt.Fprintf(&out, `\u%04x\u%04x`, h, l)
		}
	}
	return out.String()
}

// Not parallel: temporarily pins the upstream deprecation version.
func TestAdamicDecisionCores(t *testing.T) {
	var rows []map[string]any
	var want strings.Builder
	add := func(row map[string]any, out string) {
		rows = append(rows, row)
		fmt.Fprintf(&want, "case %d\n%s", len(rows)-1, out)
	}
	old := tailwindVersionForDeprecations
	defer func() { tailwindVersionForDeprecations = old }()
	names := []string{"flex", "flex-shrink-", "flex-grow-2", "flex-grow-a\nb", "shrink-0", "group/name", "[font:inherit]"}
	for _, entry := range deprecations {
		names = append(names, strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(entry.Pattern.String(), "^"), "$"), "(.*)", "50"))
	}
	for _, version := range []string{"4.3.3", "4.0.0", "4.1", "3.9.2", "5.0.0", "not-a-version", "4..1"} {
		tailwindVersionForDeprecations = func() string { return version }
		for _, name := range names {
			for _, pair := range [][2]string{{"", ""}, {"hover:", ""}, {"sm:hover:", "!"}, {"", "!"}} {
				className := pair[0] + name + pair[1]
				replacement, found := deprecationFor(className)
				out := fmt.Sprintf("deprecated %t\t%s\n", found, kernelWritten(replacement))
				if found {
					message := messageDeprecatedClassIrreplaceable(className)
					if replacement != "" {
						message = messageDeprecatedClassReplaceable(className, replacement)
					}
					out += message.Id + "\t" + kernelWritten(message.Description) + "\n"
				}
				add(map[string]any{"kind": "deprecated", "name": className, "version": version}, out)
			}
		}
	}
	diagnose := func(kind, source, value string, segments []ClassSegment) {
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/fixture.ts", Path: tspath.Path("/fixture.ts")}, source, core.ScriptKindTS)
		var found []rule.Diagnostic
		ctx := rule.Context{SourceFile: file, Report: func(d rule.Diagnostic) { found = append(found, d) }}
		row := map[string]any{"kind": kind, "source": source, "value": value, "start": 0, "end": len(utf16.Encode([]rune(source)))}
		if kind == "duplicate" {
			reportDuplicates(ctx, ClassLiteral{Text: value, Range: core.NewTextRange(0, len(source))})
		} else {
			reportTemplateDuplicates(ctx, segments)
			var values []map[string]any
			for _, segment := range segments {
				values = append(values, map[string]any{"text": segment.Text, "start": segment.Range.Pos(), "end": segment.Range.End()})
			}
			row["segments"] = values
		}
		var out strings.Builder
		for _, d := range found {
			fmt.Fprintf(&out, "%d %d %s\t%s\n", d.Range.Pos(), d.Range.End(), d.Message.Id, kernelWritten(d.Message.Description))
			for _, fix := range d.Fixes {
				fmt.Fprintf(&out, "fix %d %d\t%s\n", fix.Range.Pos(), fix.Range.End(), kernelWritten(fix.Text))
			}
		}
		add(row, out.String())
	}
	for _, value := range []string{"", "flex", "flex flex", "flex flex flex", "flex gap-2 flex gap-2", " flex  flex ", "flex\t\tflex", "flex\nflex", "flex\u00a0flex", "flex ${size} flex", "a a a b b b"} {
		diagnose("duplicate", value, value, nil)
		diagnose("duplicate", "escaped "+value, value, nil)
	}
	for _, runs := range [][]string{{"flex flex ", " flex"}, {"flex ", " flex"}, {"px-", "px-"}, {"flex  flex ", " gap-2 gap-2 "}, {" flex flex ", " flex flex ", "flex "}, {"", ""}} {
		source := ""
		var segments []ClassSegment
		for _, run := range runs {
			start := len(source)
			source += run
			segments = append(segments, ClassSegment{Text: run, Range: core.NewTextRange(start, len(source))})
			source += "HOLE"
		}
		diagnose("template", source, "", segments)
	}
	for _, name := range []string{"flx", "notavariant:flex", "group", "peer/", "group/name", "hover:peer/name!", "group/\u0085", "[font:inherit]", "hover:[font:inherit]", "", "hover:", "!"} {
		known := classExistsIn(name, nil)
		add(map[string]any{"kind": "unknown", "name": name, "present": false}, fmt.Sprintf("exists %t\n", known))
		if name != "flx" && name != "notavariant:flex" && name != "hover:[font:inherit]" {
			known = classExistsIn(name, &collapse.LoadedDesignSystem{})
			add(map[string]any{"kind": "unknown", "name": name, "present": true}, fmt.Sprintf("exists %t\n", known))
		}
	}
	data, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("ADAMIC_KERNEL_ROWS"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("ADAMIC_KERNEL_WANT"), []byte(want.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("captured %d actual Go decision-core observations", len(rows))
}
