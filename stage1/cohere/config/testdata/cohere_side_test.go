// Overlaid into Go cohere's command package. The oracle calls production functions, never a replica.
package main

import (
	"fmt"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/types/program"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAdamicConfigCases(t *testing.T) {
	input := os.Getenv("ADAMIC_CONFIG_INPUT")
	if input == "" {
		t.Skip("overlay only")
	}
	content, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "settings-census\n" {
		census(t)
		return
	}
	var output strings.Builder
	start := time.Now()
	projects := 0
	for _, line := range strings.Split(string(content), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if fields[0] == "glob" {
			bit := 0
			if configuration.Match(fields[1], fields[2]) {
				bit = 1
			}
			fmt.Fprintf(&output, "glob %d\n", bit)
			continue
		}
		root := fields[0]
		fmt.Fprintf(&output, "root %s\n", root)
		found, err := discoverProjects(root, fields[1:])
		if err != nil {
			fmt.Fprintf(&output, "error %s\n", err)
			continue
		}
		fmt.Fprintf(&output, "counts %d %d %d\n", found.Submodules, found.Ignored, found.NeverDescended)
		for _, project := range found.Projects {
			fmt.Fprintf(&output, "project %s %s\n", project.Directory, project.Engine)
			projects++
		}
		for _, refused := range found.Refused {
			fmt.Fprintf(&output, "refused %s\n", refused)
		}
		for _, nested := range found.NestedRepositories {
			fmt.Fprintf(&output, "nested %s\n", nested)
		}
	}
	elapsed := time.Since(start)
	if err := os.WriteFile(os.Getenv("ADAMIC_CONFIG_OUTPUT"), []byte(output.String()), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("Go: %.6fs, %d projects, %.0f projects/s", elapsed.Seconds(), projects, float64(projects)/elapsed.Seconds())
}

func census(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"comments.json":              "{ // comment\n \"rules\": {} }",
		"trailing.json":              "{\"rules\":{},}",
		"cycle-a.json":               "{\"extends\":\"./cycle-b.json\"}",
		"cycle-b.json":               "{\"extends\":\"./cycle-a.json\"}",
		"base.json":                  "{\"rules\":{\"eqeqeq\":[\"warn\",\"always\"]},\"ignorePatterns\":[\"base/**\"]}",
		"child.json":                 "{\"extends\":\"./base.json\",\"rules\":{\"eqeqeq\":\"error\"},\"ignorePatterns\":[\"child/**\"]}",
		"tsconfig-comments.json":     "{ // JSONC\n \"compilerOptions\":{\"strict\":true,},\"include\":[\"a.ts\",],}",
		"tsconfig-cycle-a.json":      "{\"extends\":\"./tsconfig-cycle-b.json\"}",
		"tsconfig-cycle-b.json":      "{\"extends\":\"./tsconfig-cycle-a.json\"}",
		"tsconfig-base-missing.json": "{\"extends\":\"./absent.json\"}",
		"a.ts":                       "export const a = 1;",
	}
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var output strings.Builder
	for _, name := range []string{"comments", "trailing", "missing", "cycle-a", "child"} {
		read, err := configuration.Load(filepath.Join(root, name+".json"))
		if err != nil {
			fmt.Fprintf(&output, "%s: %s\n", name, strings.ReplaceAll(err.Error(), root, "ROOT"))
		} else {
			setting := read.Rules["eqeqeq"]
			fmt.Fprintf(&output, "%s: severity %s options %s ignores %s sources %s\n", name, setting.Severity, setting.Options, read.IgnorePatterns, read.Sources)
		}
	}
	for _, name := range []string{"tsconfig-comments", "missing-tsconfig", "tsconfig-cycle-a", "tsconfig-base-missing"} {
		read, err := program.ReadProjectConfig(filepath.Join(root, name+".json"))
		if err != nil {
			fmt.Fprintf(&output, "%s: %s\n", name, strings.ReplaceAll(err.Error(), root, "ROOT"))
		} else {
			fmt.Fprintf(&output, "%s: files %s\n", name, strings.ReplaceAll(strings.Join(read.FileNames, " "), root, "ROOT"))
		}
	}
	if err := os.WriteFile(os.Getenv("ADAMIC_CONFIG_OUTPUT"), []byte(strings.ReplaceAll(output.String(), root, "ROOT")), 0644); err != nil {
		t.Fatal(err)
	}
}
