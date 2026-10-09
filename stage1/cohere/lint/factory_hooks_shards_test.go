package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testFactoryHooksUnits(t *testing.T) {
	t.Parallel()
	started := time.Now()
	directory := mutant(t, "", "")
	module := filepath.Join(directory, "rules/no-debugger/rule.ts")
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "    visit(index: number, parent: number): void {", `    ready = false;
    prepare(root: number): void { this.ready = true; console.log('prepare'); }
    finish(root: number): void { console.log('finish'); }
    visit(index: number, parent: number): void {
        if(!this.ready) { panic('visit before prepare'); }
        console.log('visit');`, 1)
	source = strings.Replace(source, "return new Rule(context);", `if(context.node(context.parents.length - 1).kind !== 'SourceFile') { panic('factory before ancestry'); }
    if(context.settings.read('number', '') !== '-2' || !context.settings.read('payload', '').includes('enabled')) { panic('structured option lost'); }
    console.log('factory');
    return new Rule(context);`, 1)
	source = "import { panic } from 'adamic';\n" + source
	if err := os.WriteFile(module, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(directory, "rules/no-debugger/rule.json")
	data, err = os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var options map[string]any
	if err := json.Unmarshal(data, &options); err != nil {
		t.Fatal(err)
	}
	options["prepare"] = "prepare"
	options["finish"] = "finish"
	write := func() {
		data, err := json.Marshal(options)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(descriptor, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	fixture := filepath.Join(t.TempDir(), "hooks.ts")
	if err := os.WriteFile(fixture, []byte("debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-debugger\t\t\tfalse\t{\"Number\":-2,\"Payload\":{\"enabled\":true}}"})
	expected := []byte("case 0\nfactory\nprepare\nvisit\nfinish\n")

	cases := factoryHookCases()
	plan := factoryHookPlan(t, cases)
	selected := factoryHookSelection(t)
	var buildTime time.Duration
	prepare := func(directory string, native bool) (module, binary string) {
		files := portFiles(t)
		for index := range files {
			files[index] = filepath.Join(directory, files[index])
		}
		var built *checkerCompilation
		dir, elapsed := factoryHookProduct(t, factoryHookInputs{Name: "lowered program and JavaScript", Files: files, Flags: []string{"EnableTSGo", "TSGoC", "JavaScript"}, Toolchain: runtime.Version()}, func(dir string) error {
			built = checkerCompile(t, directory)
			if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(built.c), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(built.javascript), 0644)
		})
		buildTime += elapsed
		module = filepath.Join(dir, "program.mjs")
		if !native {
			return
		}
		archive := ""
		if built.bridge {
			_, elapsed = factoryHookProduct(t, factoryHookInputs{Name: "sanitized checker archive", Files: files, Flags: []string{"go", "build", "-buildmode=c-archive", "-fsanitize=address,undefined"}, Toolchain: runtime.Version() + "/clang"}, func(dir string) error {
				archive = checkerArchive(t, true)
				return factoryHookCopy(archive, filepath.Join(dir, "checker.a"))
			})
			buildTime += elapsed
		}
		nativeDir, elapsed := factoryHookProduct(t, factoryHookInputs{Name: "sanitized native", Files: []string{filepath.Join(dir, "program.c"), archive}, Flags: []string{"-fsanitize=address,undefined", "-fno-sanitize-recover=all"}, Toolchain: "clang"}, func(dir string) error {
			return factoryHookCopy(checkerBinary(t, built.c, archive, true), filepath.Join(dir, "native"))
		})
		buildTime += elapsed
		binary = filepath.Join(nativeDir, "native")
		return
	}
	cases[0].directory = directory
	cases[0].module, cases[0].binary = prepare(directory, true)
	changed := mutant(t, "", "")
	if err := os.WriteFile(filepath.Join(changed, "rules/no-debugger/rule.ts"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	delete(options, "finish")
	data, err = json.Marshal(options)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(changed, "rules/no-debugger/rule.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	cases[1].directory = changed
	cases[1].module, _ = prepare(changed, false)
	t.Logf("setup: with builds=%s without builds=%s; union=%d cases; shards=%d", time.Since(started), time.Since(started)-buildTime, len(cases), len(plan))
	enumerated := 0
	for index, ids := range plan {
		enumerated++
		if !selected(index) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, id := range ids {
				c := cases[id]
				sides := []struct {
					name string
					run  execution
				}{
					{"Node", node(t, c.directory, path, false)},
					{"emitted JavaScript", runJavaScript(t, c.module, path, false)},
				}
				if !c.mutated {
					sides = append(sides, struct {
						name string
						run  execution
					}{"native", execute(t, "", c.binary, "--manifest", path)})
				}
				for _, side := range sides {
					if err := factoryHookAgreement(side.name, c.mutated, side.run.output, expected); err != nil {
						t.Fatal(err)
					}
					if c.mutated {
						t.Logf("finish-hook omission caught on %s", side.name)
					}
				}
			}
		})
	}
	if enumerated != testFactoryHooksShards {
		t.Fatalf("enumerated %d shards, want %d", enumerated, testFactoryHooksShards)
	}
}

type factoryHookCase struct {
	id, directory, module, binary string
	mutated                       bool
}

func factoryHookCases() []factoryHookCase {
	return []factoryHookCase{{id: "intact-hooks"}, {id: "finish-omission", mutated: true}}
}

func factoryHookPlan(t *testing.T, cases []factoryHookCase) [][]int {
	t.Helper()
	plan := make([][]int, testFactoryHooksShards)
	expected := map[string]bool{}
	for index, c := range cases {
		if expected[c.id] {
			t.Fatalf("repeated unsplit case %s", c.id)
		}
		expected[c.id] = true
		plan[index%len(plan)] = append(plan[index%len(plan)], index)
	}
	seen := map[string]bool{}
	total := 0
	for _, group := range plan {
		if len(group) == 0 {
			t.Fatal("empty factory hook shard")
		}
		for _, index := range group {
			c := cases[index]
			if !expected[c.id] || seen[c.id] {
				t.Fatalf("unexpected or repeated case %s", c.id)
			}
			seen[c.id] = true
			total++
		}
	}
	if total != len(cases) || len(seen) != len(expected) {
		t.Fatalf("union=%d cases/%d ids, want %d/%d", total, len(seen), len(cases), len(expected))
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("missing case %s", id)
		}
	}
	return plan
}

func factoryHookSelection(t *testing.T) func(int) bool {
	t.Helper()
	text := os.Getenv("ADAMIC_TEST_SHARD")
	if text == "" {
		return func(int) bool { return true }
	}
	parts := strings.Split(text, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD=%q, want i/n", text)
	}
	index, e1 := strconv.Atoi(parts[0])
	count, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || index < 0 || count < 1 || index >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD=%q", text)
	}
	return func(shard int) bool { return shard%count == index }
}

func factoryHookAgreement(side string, mutated bool, got, expected []byte) error {
	if bytes.HasPrefix(got, expected) == mutated {
		return fmt.Errorf("hook sequence on %s (mutant=%t): %s", side, mutated, got)
	}
	return nil
}

type factoryHookInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

// No persistent package cache: internal/buildcache is absent on this base.
// Shared builders run once per process, and callbacks write immutable products to dir.
func factoryHookProduct(t *testing.T, inputs factoryHookInputs, build func(dir string) error) (string, time.Duration) {
	t.Helper()
	dir := t.TempDir()
	started := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("build product %s: %v", inputs.Name, err)
	}
	elapsed := time.Since(started)
	t.Logf("build product %s: %s (toolchain %s)", inputs.Name, elapsed, inputs.Toolchain)
	return dir, elapsed
}

func factoryHookCopy(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	return os.WriteFile(target, data, info.Mode().Perm())
}

func TestFactoryHooksShardProbe(t *testing.T) {
	if os.Getenv("ADAMIC_FACTORY_HOOKS_PROBE") != "1" {
		return
	}
	cases := factoryHookCases()
	plan := factoryHookPlan(t, cases)
	expected := []byte("case 0\nfactory\nprepare\nvisit\nfinish\n")
	for index, group := range plan {
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, id := range group {
				// Plant a surviving finish-omission mutant in its original case only.
				if err := factoryHookAgreement("emitted JavaScript", cases[id].mutated, expected, expected); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestFactoryHooksPlantedSurvivor(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestFactoryHooksShardProbe$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_FACTORY_HOOKS_PROBE=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted surviving mutant escaped")
	}
	failures := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestFactoryHooksShardProbe/shard-") {
			failures++
			if !strings.Contains(line, "/shard-001 ") {
				t.Fatalf("wrong shard caught survivor: %s", line)
			}
		}
	}
	if failures != 1 || !bytes.Contains(output, []byte("mutant=true")) {
		t.Fatalf("want exactly shard-001 to fail: %s", output)
	}
	t.Log("planted surviving finish-omission mutant caught only by shard-001")
}
