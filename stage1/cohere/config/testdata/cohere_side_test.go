// Overlaid into Go cohere's command package. The oracle calls production functions, never a replica.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/types/program"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
		if fields[0] == "house" || fields[0] == "resolve" {
			path := fields[1]
			root := filepath.Dir(path)
			var loaded *configuration.Config
			var err error
			if fields[0] == "house" || len(fields) > 3 && fields[3] == "house" {
				loaded, err = configuration.LoadHouse(path, nil, configuration.HouseDetection{ReactFiles: map[string]bool{filepath.Join(root, "React.tsx"): true}, NextFiles: map[string]bool{filepath.Join(root, "Next.tsx"): true}, TailwindSkipped: "no stylesheet"})
			} else {
				loaded, err = configuration.Load(path)
			}
			if fields[0] == "house" {
				fmt.Fprintf(&output, "house %s\n", path)
			} else {
				fmt.Fprintf(&output, "resolve %s %s\n", path, fields[2])
			}
			if err != nil {
				fmt.Fprintf(&output, "error %s\n", err)
			} else if fields[0] == "house" {
				printAdamicSettings(&output, loaded)
			} else {
				resolved := loaded.Resolve(fields[2])
				bit := 0
				if resolved.Ignored {
					bit = 1
				}
				fmt.Fprintf(&output, "ignored %d %q\n", bit, resolved.IgnoredBy)
				names := []string{}
				for name := range resolved.Rules {
					names = append(names, name)
				}
				sort.Strings(names)
				for _, name := range names {
					setting := resolved.Rules[name]
					options := []string{}
					for _, option := range setting.Options {
						var compacted bytes.Buffer
						if err := json.Compact(&compacted, option); err != nil {
							t.Fatal(err)
						}
						options = append(options, compacted.String())
					}
					fmt.Fprintf(&output, "resolved-rule %q %s [%s]\n", name, setting.Severity, strings.Join(options, ","))
				}
			}
			continue
		}
		if fields[0] == "settings" {
			fmt.Fprintf(&output, "settings %s\n", fields[1])
			loaded, err := configuration.LoadFor(fields[1], fields[2:])
			if err != nil {
				fmt.Fprintf(&output, "error %s\n", err)
			} else {
				printAdamicSettings(&output, loaded)
			}
			continue
		}
		if fields[0] == "tsconfig" {
			fmt.Fprintf(&output, "tsconfig %s\n", fields[1])
			loaded, err := program.ReadProjectConfig(fields[1])
			if err != nil {
				fmt.Fprintf(&output, "error %s\n", err)
			} else {
				for _, file := range loaded.FileNames {
					fmt.Fprintf(&output, "file %q\n", file)
				}
				for _, reference := range loaded.References {
					fmt.Fprintf(&output, "reference %q\n", reference)
				}
			}
			continue
		}
		if fields[0] == "glob" {
			bit := 0
			if configuration.Match(fields[1], fields[2]) {
				bit = 1
			}
			fmt.Fprintf(&output, "glob %d\n", bit)
			continue
		}
		root := fields[0]
		patterns := fields[1:]
		if fields[0] == "configured" {
			root = fields[1]
			var err error
			patterns, err = rootIgnorePatterns(root)
			if err != nil {
				fmt.Fprintf(&output, "root %s\nerror %s\n", root, err)
				continue
			}
		}
		fmt.Fprintf(&output, "root %s\n", root)
		found, err := discoverProjects(root, patterns)
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
	t.Logf("Go: %.6fs, %d projects, %.0f projects/s; %d source files, %.0f files/s", elapsed.Seconds(), projects, float64(projects)/elapsed.Seconds(), strings.Count(output.String(), "\nfile "), float64(strings.Count(output.String(), "\nfile "))/elapsed.Seconds())
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

func printAdamicSettings(output *strings.Builder, loaded *configuration.Config) {
	fmt.Fprintf(output, "settings-root %q\n", loaded.Root)
	for _, source := range loaded.Sources {
		fmt.Fprintf(output, "source %q\n", source)
	}
	for _, pattern := range loaded.IgnorePatterns {
		fmt.Fprintf(output, "ignore %q\n", pattern)
	}
	for _, plugin := range loaded.Plugins {
		fmt.Fprintf(output, "plugin %q\n", plugin)
	}
	printRules := func(prefix string, rules map[string]configuration.RuleSetting) {
		names := make([]string, 0, len(rules))
		for name := range rules {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			setting := rules[name]
			options := []string{}
			for _, option := range setting.Options {
				var buffer bytes.Buffer
				if err := json.Compact(&buffer, option); err != nil {
					panic(err)
				}
				options = append(options, buffer.String())
			}
			fmt.Fprintf(output, "%s %q %s [%s]\n", prefix, name, setting.Severity, strings.Join(options, ","))
		}
	}
	printRules("rule", loaded.Rules)
	names := []string{}
	for name := range loaded.Departures {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		reason := loaded.Departures[name]
		fmt.Fprintf(output, "departure %q %q %q\n", name, reason.File, reason.Reason)
	}
	names = []string{}
	for name := range loaded.OffReasons {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		reason := loaded.OffReasons[name]
		fmt.Fprintf(output, "off-reason %q %q %q\n", name, reason.File, reason.Reason)
	}
	for _, override := range loaded.Overrides {
		files := []string{}
		for _, file := range override.Files {
			files = append(files, strconv.Quote(file))
		}
		fmt.Fprintf(output, "override %q %q [%s]\n", override.File, override.Reason, strings.Join(files, ","))
		printRules("override-rule", override.Rules)
	}
	version := ""
	if loaded.CohereVersion != nil {
		version = loaded.CohereVersion.Text
	}
	fmt.Fprintf(output, "cohere-version %q\n", version)
}
