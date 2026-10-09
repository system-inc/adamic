package config

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func loaderCases(t *testing.T) string {
	t.Helper()
	var input strings.Builder
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, project := range []string{root, filepath.Join(root, "cohere"), filepath.Join(root, "cohere/TypeScript")} {
		fmt.Fprintf(&input, "settings\t%s\n", filepath.Join(project, "CohereSettings.json"))
		fmt.Fprintf(&input, "house\t%s\n", filepath.Join(project, "CohereSettings.json"))
		err := filepath.WalkDir(project, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && path != project && (entry.Name() == ".git" || entry.Name() == "node_modules" || entry.Name() == "testdata" || entry.Name() == ".cache" || entry.Name() == ".build") {
				return filepath.SkipDir
			}
			if !entry.IsDir() && entry.Name() == "tsconfig.json" {
				fmt.Fprintf(&input, "tsconfig\t%s\n", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	tree := t.TempDir()
	write := func(name, text string) string {
		t.Helper()
		path := filepath.Join(tree, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	settings := map[string]string{
		"empty": "{}", "null": "null", "comments": "{ // comment\n \"rules\":{} }", "trailing": "{\"rules\":{},}",
		"unknown": "{\"unimplemented\":true}", "unknown-override": "{\"overrides\":[{\"files\":[\"*.ts\"],\"excludedFiles\":[]}]}",
		"base":    "{\"rules\":{\"eqeqeq\":[\"warn\",\"always\",{\"null\":\"ignore\"}]},\"ignorePatterns\":[\"base/**\"],\"plugins\":[\"react\"]}",
		"child":   "{\"extends\":\"./base.json\",\"rules\":{\"eqeqeq\":\"error\"},\"ignorePatterns\":[\"child/**\"],\"plugins\":[\"react\",\"@typescript-eslint\"]}",
		"cycle-a": "{\"extends\":\"./cycle-b.json\"}", "cycle-b": "{\"extends\":\"./cycle-a.json\"}",
		"absent-base": "{\"extends\":\"./absent.json\"}", "set-typescript": "{\"extends\":\"cohere:typescript\"}",
		"set-composed": "{\"extends\":[\"cohere:react\",\"cohere:next\",\"cohere:tailwind\"]}", "set-adamic": "{\"extends\":\"cohere:adamic\"}",
		"set-style": "{\"extends\":\"cohere:style\"}", "old-house": "{\"extends\":\"cohere:house\"}",
		"bad-set": "{\"extends\":\"cohere:absent\"}", "empty-extends": "{\"extends\":[\"\"]}", "wrong-extends": "{\"extends\":7}",
		"reason":     "{\"rules\":{\"eqeqeq\":\"off\"},\"reasons\":{\"eqeqeq\":\"  Chosen off.  \"}}",
		"bad-reason": "{\"reasons\":{\"eqeqeq\":\"Something.\"}}", "empty-reason": "{\"reasons\":{\"eqeqeq\":\" \"}}",
		"departure":     "{\"extends\":\"./base.json\",\"rules\":{\"eqeqeq\":\"off\"},\"departures\":{\"eqeqeq\":\"Chosen difference.\"}}",
		"bad-departure": "{\"extends\":\"./base.json\",\"departures\":{\"eqeqeq\":\"Same.\"}}",
		"scoped":        "{\"extends\":\"./base.json\",\"overrides\":[{\"files\":[\"**/*.test.ts\"],\"rules\":{\"eqeqeq\":[\"off\",\"smart\"]},\"reason\":\" tests \"}]}",
		"invalid-rule":  "{\"rules\":{\"eqeqeq\":9}}", "invalid-severity": "{\"rules\":{\"eqeqeq\":\"typo\"}}", "empty-rule": "{\"rules\":{\"eqeqeq\":[]}}",
		"null-rule": "{\"rules\":{\"eqeqeq\":null}}", "wrong-rules": "{\"rules\":[]}", "wrong-plugins": "{\"plugins\":[9]}",
		"pinned": "{\"cohere\":\"^1.2.3\"}",
	}
	for index, text := range []string{
		"", "{", "{rules:{}}", "{\"rules\": [}", "{\"ignorePatterns\":[\"\\q\"]}", "{\"ignorePatterns\":[\"\\u00xx\"]}", "{\"ignorePatterns\":[\"line\n\"]}",
		"{\"rules\":{}} tail", "{\"rules\":{\"a\":[\"warn\",1.]}}", "{\"rules\":{\"a\":[\"warn\",1e]}}", "{\"rules\":{\"a\":[\"warn\",truX]}}", "{\"rules\":{\"a\":[\"warn\",01]}}",
		`{"ignorePatterns":["\u0000","\u007f","\u0085","\u2028","\u200d","\ud800","\ud800\udc00"]}`,
		`{"rules":{"a":"warn"},"rules":{"b":"error"}}`, `{"overrides":{}}`, `{"overrides":[7]}`, `{"overrides":[{"files":7}]}`, `{"overrides":[{"files":[9]}]}`, `{"overrides":[{"reason":7}]}`, `{"reasons":{"a":7}}`,
		`{"cohere":"*"}`, `{"cohere":""}`, `{"cohere":"^1.2"}`, `{"cohere":"^01.2.3"}`, `{"cohere":"1.2.3 ||"}`, `{"cohere":"1.2.3-"}`, `{"cohere":" >=1.2.3 <2.0.0 "}`,
	} {
		settings[fmt.Sprintf("syntax-%02d", index)] = text
	}
	write("alias-base.json", `{"rules":{"nexus/no-shadow":["warn","option"]}}`)
	write("alias-child.json", `{"extends":"./alias-base.json","rules":{"@typescript-eslint/no-shadow":"error"}}`)
	write("twin-base.json", `{"rules":{"@typescript-eslint/no-invalid-this":"error"}}`)
	write("twin-child.json", `{"extends":"./twin-base.json","rules":{"no-invalid-this":"warn"}}`)
	for _, name := range []string{"alias-child", "twin-child"} {
		fmt.Fprintf(&input, "settings\t%s\tno-shadow\t@typescript-eslint/no-shadow\tno-invalid-this\t@typescript-eslint/no-invalid-this\n", filepath.Join(tree, name+".json"))
	}
	for name, text := range settings {
		write(name+".json", text)
	}
	names := make([]string, 0, len(settings))
	for name := range settings {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&input, "settings\t%s\n", filepath.Join(tree, name+".json"))
	}
	fmt.Fprintf(&input, "settings\t%s\n", filepath.Join(tree, "missing.json"))
	for _, name := range []string{"empty", "child", "reason", "scoped", "departure", "set-typescript"} {
		path := filepath.Join(tree, name+".json")
		if name != "set-typescript" {
			fmt.Fprintf(&input, "house\t%s\n", path)
		}
		for _, file := range []string{"source/a.ts", "source/a.test.ts", "React.tsx", "Next.tsx", "base/a.ts", "child/a.ts"} {
			for _, mode := range []string{"settings", "house"} {
				fmt.Fprintf(&input, "resolve\t%s\t%s\t%s\n", path, filepath.Join(tree, file), mode)
			}
		}
	}
	random := rand.New(rand.NewSource(20261006))
	for number := 0; number < 80; number++ {
		rule := []string{"off", "warn", "error", "deny", "allow", "WARNING"}[random.Intn(6)]
		extra := []string{"", ",\"ignorePatterns\":[\"one/**\",\"two/*\"]", ",\"plugins\":[\"react\",\"react-hooks\"]"}[random.Intn(3)]
		name := fmt.Sprintf("generated-%d.json", number)
		path := write(name, fmt.Sprintf(`{"extends":"./base.json","rules":{"eqeqeq":%q}%s}`, rule, extra))
		fmt.Fprintf(&input, "settings\t%s\n", path)
	}
	for _, name := range []string{"src/a.ts", "src/a.tsx", "src/a.d.ts", "src/a.js", "src/a.jsx", "src/b.ts", "src/b.d.ts", "src/b.js", "src/x.cts", "src/x.d.cts", "src/x.cjs", "src/x.mts", "src/x.d.mts", "src/x.mjs", "src/z.min.js", "src/keep.min.js", "src/é.ts", "src/😀.ts", "src/.hidden.ts", "src/deep/c.ts", "src/deep/data.json", "src/data.json", "src/code.a", "src/a.a", "src/node_modules/n.ts", "src/.hidden/h.ts", "src/UPPER.TS", "out/a.ts"} {
		write(name, "export const value=1;\n")
	}
	projects := map[string]string{
		"default": "{}", "comments": "{ // JSONC\n\"include\":[\"src\",],\"compilerOptions\":{\"allowJs\":true,},}",
		"order": "{\"include\":[\"src/deep\",\"src/*.ts\"]}", "exclude": "{\"include\":[\"src\"],\"exclude\":[\"src/deep\",\"src/b.*\"]}",
		"literal":           "{\"files\":[\"src/b.js\",\"missing.ts\",\"src/a.ts\",\"src/a.ts\"],\"include\":[\"src\"],\"exclude\":[\"src/**\"]}",
		"base":              "{\"include\":[\"../src\"],\"compilerOptions\":{\"outDir\":\"../out\",\"allowJs\":true}}",
		"child":             "{\"extends\":\"./base.json\",\"exclude\":[\"../src/b.*\"]}",
		"base-no-extension": "{\"extends\":\"./base\"}",
		"cycle-a":           "{\"extends\":\"./cycle-b.json\"}", "cycle-b": "{\"extends\":\"./cycle-a.json\"}", "missing-base": "{\"extends\":\"./absent.json\"}",
		"empty": "{\"files\":[]}", "solution": "{\"files\":[],\"references\":[{\"path\":\"../src\"},{\"path\":\"./base.json\"}]}",
		"json":              "{\"compilerOptions\":{\"resolveJsonModule\":true},\"include\":[\"../src/**/*.json\",\"../src/**/*.ts\"]}",
		"foreign-extension": `{ "sourceExtensions": [".a"], "include": ["../src"] }`,
		"literal-foreign":   `{ "files": ["../src/code.a"] }`,
		"extension-base":    `{ "sourceExtensions": [".a"], "include": ["../src"] }`,
		"extension-child":   `{ "extends": "./extension-base.json" }`,
		"extension-clear":   `{ "extends": "./extension-base.json", "sourceExtensions": [] }`,
		"extension-invalid": `{ "sourceExtensions": ["a", ".", ".ts", ".TS", ".json", ".a", ".a"], "include": ["../src"] }`,
		"extension-types":   `{ "sourceExtensions": [7, ".a", null], "include": ["../src"] }`,
		"extension-wrong":   `{ "sourceExtensions": true, "include": ["../src"] }`,
		"extension-null":    `{ "sourceExtensions": null, "include": ["../src"] }`,
		"unknown-option":    "{\"compilerOptions\":{\"doesNotExist\":true,\"target\":\"nonsense\"},\"include\":[\"../src\"]}",
		"template":          "{\"include\":[\"${configDir}/../src\"]}", "wildcard-question": "{\"include\":[\"../src/?.ts\"]}",
		"hidden-explicit":       "{\"include\":[\"../src/.hidden/*.ts\",\"../src/.hidden.ts\"]}",
		"package-explicit":      "{\"include\":[\"../src/node_modules/*.ts\"]}",
		"min-explicit":          "{\"compilerOptions\":{\"allowJs\":true},\"include\":[\"../src/*.min.js\"]}",
		"all-json-not-explicit": "{\"compilerOptions\":{\"resolveJsonModule\":true},\"include\":[\"../src\"]}",
		"invalid-recursive":     `{ "include": ["**"] }`,
		"invalid-parent":        `{ "include": ["../src/**/*.ts"], "exclude": ["**/../src"] }`,
		"no-input":              "{\"include\":[\"absent/*.ts\"]}",
	}
	for name, text := range projects {
		write("configs/"+name+".json", text)
	}
	names = nil
	for name := range projects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(&input, "tsconfig\t%s\n", filepath.Join(tree, "configs", name+".json"))
	}
	fmt.Fprintf(&input, "tsconfig\t%s\n", filepath.Join(tree, "missing-tsconfig.json"))
	for number := 0; number < 80; number++ {
		include := []string{"../src", "../src/*.ts", "../src/**/*.ts", "../src/deep", "../src/*.*"}[random.Intn(5)]
		exclude := []string{"../src/b.*", "../src/deep", "../out", "nothing"}[random.Intn(4)]
		allow := random.Intn(2) == 0
		js := random.Intn(2) == 0
		path := write(fmt.Sprintf("configs/generated-%d.json", number), fmt.Sprintf(`{"compilerOptions":{"allowJs":%t,"resolveJsonModule":%t},"include":[%q],"exclude":[%q]}`, allow, js, include, exclude))
		fmt.Fprintf(&input, "tsconfig\t%s\n", path)
	}
	inputPath := filepath.Join(t.TempDir(), "cases")
	if err := os.WriteFile(inputPath, []byte(input.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return inputPath
}
func TestLoadersMatchGoCohere(t *testing.T) {
	t.Parallel()
	input := loaderCases(t)
	want := goCohere(t, input)

	directory := filepath.Dir(portDirectory(t, "", "", ""))
	path := filepath.Join(directory, "loader_main.ts")
	program := lowered(t, path)
	result, binary := natively(t, program, input)
	for _, side := range []struct {
		name   string
		result run
	}{{"native", result}, {"Node", onNode(t, path, input)}, {"backend", onJavaScriptBackend(t, program, input)}} {
		if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
			t.Fatalf("%s: exit %d, stderr %s", side.name, side.result.exitCode, side.result.stderr)
		}

		if difference := firstDifference(string(side.result.stdout), want); difference != "" {
			t.Errorf("%s: %s", side.name, difference)
			os.WriteFile(filepath.Join(t.TempDir(), "actual"), side.result.stdout, 0644)
		}
	}
	if report := leaks(t, program, binary, input); report != "" {
		t.Error(report)
	}
	if t.Failed() {
		t.Fatal("baseline must agree before mutants")
	}
	t.Logf("%d settings, %d tsconfigs, %d source files", strings.Count(want, "settings "), strings.Count(want, "tsconfig "), strings.Count(want, "\nfile "))
	for _, mutant := range []struct{ name, file, from, to string }{
		{"source extension opt-in", "tsconfig.ts", "if (!extensions.includes(suffix))", "if (false)"},
		{"preset origin", "sets.ts", `\"cohere:style\"`, `\"cohere:house\"`},
		{"strict JSON comments", "settings.ts", "parseJson(text, false)", "parseJson(text, true)"},
		{"inherited rule options", "settings.ts", "setting = { severity: setting.severity, options: inherited.options };", "setting = { severity: setting.severity, options: [] };"},
		{"tsconfig excludes", "tsconfig.ts", "if (exclude.matches(path, false))", "if (false)"},
	} {
		t.Run("catches "+mutant.name, func(t *testing.T) {
			path := filepath.Join(filepath.Dir(portDirectory(t, mutant.file, mutant.from, mutant.to)), "loader_main.ts")
			program := lowered(t, path)
			sides := []struct {
				name   string
				result run
			}{{"native", nativelyRun(t, program, input)}, {"Node", onNode(t, path, input)}, {"backend", onJavaScriptBackend(t, program, input)}}
			if mutant.name == "preset origin" {
				for _, side := range sides[1:] {
					if string(side.result.stdout) != string(sides[0].result.stdout) {
						t.Fatal("preset mutant must agree across Adamic and Node before Go catches it")
					}
				}
				t.Log("all three port executions agree; only the independent Go reference catches the preset mutant")
			}
			for _, side := range sides {
				if side.result.exitCode != 0 || len(side.result.stderr) > 0 {
					t.Fatalf("%s mutant must execute successfully: exit %d stderr %s", side.name, side.result.exitCode, side.result.stderr)
				}
				difference := firstDifference(string(side.result.stdout), want)
				if difference == "" {
					t.Fatal("mutant survived")
				}
				t.Logf("%s caught: %s", side.name, difference)
			}
		})
	}
	if os.Getenv("ADAMIC_CONFIG_TIMING") != "" {
		content, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		var onlyTsconfig strings.Builder
		for _, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(line, "tsconfig\t") {
				fmt.Fprintln(&onlyTsconfig, line)
			}
		}
		tsconfigInput := filepath.Join(t.TempDir(), "tsconfig-cases")
		if err := os.WriteFile(tsconfigInput, []byte(onlyTsconfig.String()), 0644); err != nil {
			t.Fatal(err)
		}
		answers := goCohere(t, tsconfigInput)
		binary := filepath.Join(t.TempDir(), "timed")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		for round := 0; round < 3; round++ {
			started := time.Now()
			result := execute(t, nil, binary, tsconfigInput)
			elapsed := time.Since(started)
			if result.exitCode != 0 || len(result.stderr) > 0 || string(result.stdout) != answers {
				t.Fatal("timed native differs from Go")
			}
			t.Logf("tsconfig native round %d: %.6fs, %d files, %.0f files/s", round, elapsed.Seconds(), strings.Count(answers, "\nfile "), float64(strings.Count(answers, "\nfile "))/elapsed.Seconds())
		}
	}

}
