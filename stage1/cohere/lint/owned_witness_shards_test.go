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

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

type ownedWitnessCase struct {
	id, row    string
	descriptor registry.Descriptor
}

func ownedWitnessCases(t *testing.T, directory string) []ownedWitnessCase {
	t.Helper()
	var cases []ownedWitnessCase
	for _, d := range prepareRegistry(t, directory) {
		paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
		if err != nil {
			t.Fatal(err)
		}
		rows := ownedWitnessRows(t, directory, d)
		if len(paths) != len(rows) {
			t.Fatalf("%s: %d paths, %d rows", d.Slug, len(paths), len(rows))
		}
		for index, row := range rows {
			relative, err := filepath.Rel(directory, paths[index])
			if err != nil {
				t.Fatal(err)
			}
			cases = append(cases, ownedWitnessCase{filepath.ToSlash(relative), row, d})
		}
	}
	return cases
}

// Validate the entire plan before applying a box selector: every original identity
// must occur exactly once, including on boxes which execute only part of the plan.
func planOwnedWitnesses(cases []ownedWitnessCase, count int) ([][]ownedWitnessCase, error) {
	if count < 1 || len(cases) < count {
		return nil, fmt.Errorf("%d witnesses cannot fill %d shards", len(cases), count)
	}
	original := make(map[string]bool, len(cases))
	plan := make([][]ownedWitnessCase, count)
	for index, witness := range cases {
		if original[witness.id] {
			return nil, fmt.Errorf("repeated unsplit witness %s", witness.id)
		}
		original[witness.id] = true
		plan[index%count] = append(plan[index%count], witness)
	}
	if err := checkOwnedWitnessUnion(cases, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func checkOwnedWitnessUnion(cases []ownedWitnessCase, plan [][]ownedWitnessCase) error {
	want := make(map[string]bool, len(cases))
	for _, witness := range cases {
		if want[witness.id] {
			return fmt.Errorf("repeated unsplit witness %s", witness.id)
		}
		want[witness.id] = true
	}
	seen := make(map[string]bool, len(cases))
	total := 0
	for _, group := range plan {
		for _, witness := range group {
			total++
			if !want[witness.id] {
				return fmt.Errorf("unexpected witness %s", witness.id)
			}
			if seen[witness.id] {
				return fmt.Errorf("repeated witness %s", witness.id)
			}
			seen[witness.id] = true
		}
	}
	if total != len(cases) {
		return fmt.Errorf("union has %d witnesses, want %d", total, len(cases))
	}
	for id := range want {
		if !seen[id] {
			return fmt.Errorf("missing witness %s", id)
		}
	}
	return nil
}

func ownedShardSelection(t *testing.T) func(int) bool {
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

func ownedWitnessAgreement(id, side string, got, want []byte) error {
	if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%s on %s: %s", id, side, diff)
	}
	return nil
}

// Preserve the live typed checker, transcript replay, all runtime comparisons,
// and the sanitized native execution of compareWithJavaScript.
func compareOwnedWitness(t *testing.T, oracle, binary, directory, path, module, id string) {
	t.Helper()
	want := execute(t, "", oracle, "--manifest", path)
	text, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var sides []struct {
		name string
		run  execution
	}
	if bytes.HasPrefix(text, []byte("program ")) {
		prefix := filepath.Join(t.TempDir(), "transcript")
		native := execute(t, "", binary, "--manifest", path, "--record", prefix)
		runner := filepath.Join(repository, "oracle/node.mjs")
		sides = append(sides, struct {
			name string
			run  execution
		}{"native", native},
			struct {
				name string
				run  execution
			}{"Node", execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", path, "--replay", prefix)},
			struct {
				name string
				run  execution
			}{"emitted JavaScript", execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, "--manifest", path, "--replay", prefix)})
	} else {
		sides = append(sides, struct {
			name string
			run  execution
		}{"Node", node(t, directory, path, false)},
			struct {
				name string
				run  execution
			}{"emitted JavaScript", runJavaScript(t, module, path, false)},
			struct {
				name string
				run  execution
			}{"native", execute(t, "", binary, "--manifest", path)})
	}
	for _, side := range sides {
		if err := ownedWitnessAgreement(id, side.name, side.run.output, want.output); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOwnedWitnessUnionRejectsMissingAndRepeated(t *testing.T) {
	cases := ownedWitnessCases(t, ".")
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
	}
	missing := append([][]ownedWitnessCase(nil), plan...)
	missing[0] = missing[0][1:]
	if checkOwnedWitnessUnion(cases, missing) == nil {
		t.Fatal("missing witness survived")
	}
	repeated := append([][]ownedWitnessCase(nil), plan...)
	repeated[0] = append(append([]ownedWitnessCase(nil), repeated[0]...), cases[0])
	if checkOwnedWitnessUnion(cases, repeated) == nil {
		t.Fatal("repeated witness survived")
	}
}

// The child uses the real corpus plan and the production comparison, with a
// disagreement planted in precisely one case. It needs no build products.
func TestOwnedWitnessShardProbe(t *testing.T) {
	if os.Getenv("ADAMIC_OWNED_WITNESS_PROBE") != "1" {
		return
	}
	cases := ownedWitnessCases(t, ".")
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
	}
	for index, group := range plan {
		t.Run(fmt.Sprintf("shard-%03d", index), func(t *testing.T) {
			t.Parallel()
			for _, witness := range group {
				want := []byte("case 0\n")
				got := want
				if witness.id == cases[0].id {
					got = []byte("case 0\nplanted disagreement\n")
				}
				if err := ownedWitnessAgreement(witness.id, "native", got, want); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestOwnedWitnessPlantedDisagreement(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "-test.run=^TestOwnedWitnessShardProbe$", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_OWNED_WITNESS_PROBE=1")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted disagreement survived")
	}
	failures := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestOwnedWitnessShardProbe/shard-") {
			failures++
			if !strings.Contains(line, "/shard-000 ") {
				t.Fatalf("wrong shard caught disagreement: %s", line)
			}
		}
	}
	if failures != 1 || !bytes.Contains(output, []byte("planted disagreement")) {
		t.Fatalf("want exactly shard-000 to fail: %s", output)
	}
	t.Log("planted native disagreement caught only by shard-000")
}

// ownedProductInputs describes the build inputs beside each callback. There is
// deliberately no persistent package cache: internal/buildcache is not on this
// base. The existing shared helpers build once per process; shards only consume
// these immutable copies. Each callback writes its products into dir.
type ownedProductInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func ownedBuildProduct(t *testing.T, inputs ownedProductInputs, build func(dir string) error) (string, time.Duration) {
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

func ownedCopyProduct(source, target string) error {
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

func ownedBuildProducts(t *testing.T, directory string) (oracle, binary, module string, builds time.Duration) {
	t.Helper()
	files := portFiles(t)
	toolchain := runtime.Version()
	oracleDir, elapsed := ownedBuildProduct(t, ownedProductInputs{
		Name: "Go oracle", Files: files, Flags: []string{"go", "build", "-overlay"}, Toolchain: toolchain,
	}, func(dir string) error { return ownedCopyProduct(goOracle(t), filepath.Join(dir, "oracle")) })
	builds += elapsed
	oracle = filepath.Join(oracleDir, "oracle")
	var lowered *checkerCompilation
	loweredDir, elapsed := ownedBuildProduct(t, ownedProductInputs{
		Name: "lowered C and emitted JavaScript", Files: files, Flags: []string{"EnableTSGo", "TSGoC", "JavaScript"}, Toolchain: toolchain,
	}, func(dir string) error {
		lowered = checkerCompile(t, directory)
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(lowered.c), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(lowered.javascript), 0644)
	})
	builds += elapsed
	module = filepath.Join(loweredDir, "program.mjs")
	archive := ""
	if lowered.bridge {
		_, elapsed = ownedBuildProduct(t, ownedProductInputs{
			Name: "sanitized checker archive", Files: files, Flags: []string{"go", "build", "-buildmode=c-archive", "-fsanitize=address,undefined", "-fno-sanitize-recover=all"}, Toolchain: toolchain + "/clang",
		}, func(dir string) error {
			archive = checkerArchive(t, true)
			return ownedCopyProduct(archive, filepath.Join(dir, "checker.a"))
		})
		builds += elapsed
	}
	nativeDir, elapsed := ownedBuildProduct(t, ownedProductInputs{
		Name: "sanitized native binary", Files: []string{filepath.Join(loweredDir, "program.c"), archive}, Flags: []string{"-fsanitize=address,undefined", "-fno-sanitize-recover=all"}, Toolchain: "clang",
	}, func(dir string) error {
		return ownedCopyProduct(checkerBinary(t, lowered.c, archive, true), filepath.Join(dir, "native"))
	})
	builds += elapsed
	binary = filepath.Join(nativeDir, "native")
	return
}
