package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: compiler builds, sanitizers and measurements share the machine.
func TestWave15RegexPartialControlsAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE15_REGEX_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_15_regex_controls.a")
	binary := h.build(stage0, "wave15", entry, archive, false)
	oracle := volumeOracle(h, "wave15-oracle", "oracle_wave_15_regex.go")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := []string{
		"new RegExp('.', 'ii');RegExp('.', 'z');new RegExp('.', 'uv');new RegExp('.', 'zgg');",
		"declare const pattern:string;new RegExp(pattern,'i');RegExp();function f(RegExp:(a:string,b:string)=>unknown){RegExp('[','z');}",
		"const regex=/[Á]/u;",
		"const regex=/[👶🏻]/u;",
		"const regex=/[🇯🇵]/u;",
		"const regex=/[👨‍👩‍👦]/u;",
		"const regex=/[👍]/;",
		"const regex=/[👍]/u;",
		"const regex=/[abc]/;",
		"/* 世界 🌍 */\r\nconst regex=/[👍]/;\r\n",
	}
	paths := []string{}
	for i, s := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), s+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave15-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"flags", "no_invalid_regexp.a", "Duplicate flag", "Unknown flag"},
		{"suggestion", "no_misleading_character_class.a", "new Repair(literalEnd, literalEnd, 'u')", "new Repair(literalEnd, literalEnd, 'v')"},
	} {
		t.Run(change.name, func(t *testing.T) {
			previous := h.t
			h.t = t
			defer func() { h.t = previous }()
			scratch := filepath.Join(directory, change.name+"-source")
			if err := os.MkdirAll(scratch, 0755); err != nil {
				t.Fatal(err)
			}
			files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*"))
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if filepath.Ext(file) != ".a" && filepath.Ext(file) != ".ts" {
					continue
				}
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				s := string(data)
				if filepath.Base(file) == change.file {
					if strings.Count(s, change.from) != 1 {
						t.Fatal("nonunique mutant")
					}
					s = strings.Replace(s, change.from, change.to, 1)
				}
				s = strings.ReplaceAll(s, "../../typescript/", filepath.Join(repository, "stage1/typescript")+"/")
				s = strings.ReplaceAll(s, "../lint/", filepath.Join(repository, "stage1/cohere/lint")+"/")
				h.write(filepath.Join(change.name+"-source", filepath.Base(file)), s)
			}
			mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave_15_regex_controls.a"), archive, false)
			got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
			if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
				t.Fatal("mutant survived")
			}
			t.Logf("%s: exit 0, independent Go bytes catch byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		})
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','symbols-in-scope'));
`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
}
