package lint

import (
	"bytes"
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

func testDecodedOptionsUnits(t *testing.T) {
	t.Parallel()
	started := time.Now()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(t.TempDir(), "catch.ts")
	if err := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}"})
	cases := decodedOptionsCases()
	plan := decodedOptionsPlan(t, cases)
	selected := decodedOptionsSelection(t)
	var buildTime time.Duration
	oracleDir, elapsed := decodedOptionsProduct(t, decodedOptionsInputs{Name: "Go oracle", Files: portFiles(t), Flags: []string{"go", "build", "-overlay"}, Toolchain: runtime.Version()}, func(dir string) error {
		return decodedOptionsCopy(goOracle(t), filepath.Join(dir, "oracle"))
	})
	buildTime += elapsed
	oracle := filepath.Join(oracleDir, "oracle")
	prepare := func(directory string, native bool) (module, binary string) {
		files := portFiles(t)
		for index := range files {
			files[index] = filepath.Join(directory, files[index])
		}
		var built *checkerCompilation
		dir, elapsed := decodedOptionsProduct(t, decodedOptionsInputs{Name: "lowered program and JavaScript", Files: files, Flags: []string{"EnableTSGo", "TSGoC", "JavaScript"}, Toolchain: runtime.Version()}, func(dir string) error {
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
			_, elapsed = decodedOptionsProduct(t, decodedOptionsInputs{Name: "sanitized checker archive", Files: files, Flags: []string{"go", "build", "-buildmode=c-archive", "-fsanitize=address,undefined"}, Toolchain: runtime.Version() + "/clang"}, func(dir string) error {
				archive = checkerArchive(t, true)
				return decodedOptionsCopy(archive, filepath.Join(dir, "checker.a"))
			})
			buildTime += elapsed
		}
		nativeDir, elapsed := decodedOptionsProduct(t, decodedOptionsInputs{Name: "sanitized native", Files: []string{filepath.Join(dir, "program.c"), archive}, Flags: []string{"-fsanitize=address,undefined", "-fno-sanitize-recover=all"}, Toolchain: "clang"}, func(dir string) error {
			return decodedOptionsCopy(checkerBinary(t, built.c, archive, true), filepath.Join(dir, "native"))
		})
		buildTime += elapsed
		binary = filepath.Join(nativeDir, "native")
		return
	}
	cases[0].directory = directory
	cases[0].module, cases[0].binary = prepare(directory, true)
	changed := mutant(t, "'allowemptycatch'", "'ignored-allowemptycatch'", "main.ts")
	cases[1].directory = changed
	cases[1].module, _ = prepare(changed, false)
	t.Logf("setup: with builds=%s without builds=%s; union=%d cases; shards=%d", time.Since(started), time.Since(started)-buildTime, len(cases), len(plan))
	enumerated := 0
	for index, group := range plan {
		enumerated++
		if !selected(index) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, id := range group {
				c := cases[id]
				if !c.mutated {
					compareWithJavaScript(t, oracle, c.binary, c.directory, path, c.module)
					count := execute(t, "", oracle, "--manifest", path, "--count")
					if string(count.output) != "0\n" {
						t.Fatal("JSON catch option did not override the legacy default")
					}
					continue
				}
				want := execute(t, "", oracle, "--manifest", path).output
				for _, side := range []struct {
					name string
					run  execution
				}{
					{"Node", node(t, c.directory, path, false)},
					{"emitted JavaScript", runJavaScript(t, c.module, path, false)},
				} {
					if err := decodedOptionsAgreement(side.name, true, side.run.output, want); err != nil {
						t.Fatal(err)
					}
					t.Logf("ignored decoded-option mutant caught on %s: %s", side.name, difference(side.run.output, want))
				}
			}
		})
	}
	if enumerated != testDecodedOptionsAndMutantShards {
		t.Fatalf("enumerated %d shards, want %d", enumerated, testDecodedOptionsAndMutantShards)
	}
}

type decodedOptionsCase struct {
	id, directory, module, binary string
	mutated                       bool
}

func decodedOptionsCases() []decodedOptionsCase {
	return []decodedOptionsCase{{id: "decoded-option"}, {id: "ignored-option", mutated: true}}
}

func decodedOptionsPlan(t *testing.T, cases []decodedOptionsCase) [][]int {
	t.Helper()
	plan := make([][]int, testDecodedOptionsAndMutantShards)
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
			t.Fatal("empty decoded option shard")
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

func decodedOptionsSelection(t *testing.T) func(int) bool {
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

func decodedOptionsAgreement(side string, mutated bool, got, want []byte) error {
	if bytes.Equal(got, want) == mutated {
		return fmt.Errorf("ignored decoded-option mutant survived on %s (mutant=%t)", side, mutated)
	}
	return nil
}

type decodedOptionsInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

// No persistent package cache: internal/buildcache is absent on this base.
// Shared builders run once per process, and callbacks write immutable products to dir.
func decodedOptionsProduct(t *testing.T, inputs decodedOptionsInputs, build func(dir string) error) (string, time.Duration) {
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

func decodedOptionsCopy(source, target string) error {
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

func TestDecodedOptionsAndMutantShardProbe(t *testing.T) {
	if os.Getenv("ADAMIC_DECODED_OPTIONS_PROBE") != "1" {
		return
	}
	cases := decodedOptionsCases()
	plan := decodedOptionsPlan(t, cases)
	expected := []byte("case 0\n")
	for index, group := range plan {
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, id := range group {
				// Plant a surviving ignored-option mutant in its original case only.
				if err := decodedOptionsAgreement("emitted JavaScript", cases[id].mutated, expected, expected); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestDecodedOptionsAndMutantPlantedSurvivor(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestDecodedOptionsAndMutantShardProbe$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_DECODED_OPTIONS_PROBE=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted surviving mutant escaped")
	}
	failures := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestDecodedOptionsAndMutantShardProbe/shard-") {
			failures++
			if !strings.Contains(line, "/shard-001 ") {
				t.Fatalf("wrong shard caught survivor: %s", line)
			}
		}
	}
	if failures != 1 || !bytes.Contains(output, []byte("mutant=true")) {
		t.Fatalf("want exactly shard-001 to fail: %s", output)
	}
	t.Log("planted surviving ignored-option mutant caught only by shard-001")
}
