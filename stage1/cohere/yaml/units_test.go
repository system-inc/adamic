package yaml

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// There is no internal/buildcache on this base. Each builder writes a fresh
// product directory once per run, outside parallel logic, shared by its units.
// Keep the inputs beside the builder so Product can replace this fallback when
// the shared internal/buildcache helper lands; this is not a package cache.
type yamlProductInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func yamlProduct(t *testing.T, inputs yamlProductInputs, build func(dir string) error) string {
	t.Helper()
	if inputs.Name == "" || len(inputs.Files) == 0 || inputs.Toolchain == "" {
		t.Fatal("incomplete build inputs")
	}
	// Expand directories into explicit files; buildcache keys must cover transitive
	// checker, lowering/emitter, runtime and upstream Go sources as well as adapters.
	var files []string
	for _, input := range inputs.Files {
		err := filepath.WalkDir(input, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && entry.Name() == ".git" {
				return filepath.SkipDir
			}
			if !entry.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Strings(files)
	inputs.Files = files
	switch inputs.Toolchain {
	case "go":
		inputs.Toolchain = strings.TrimSpace(string(run(t, "", nil, "go", "version")))
	case "clang":
		inputs.Toolchain = strings.Join(strings.Fields(string(run(t, "", nil, "clang", "--version"))), " ")
	case "adamic":
		inputs.Toolchain = runtime.Version() + "/" + runtime.GOOS + "/" + runtime.GOARCH
	}
	dir := t.TempDir()
	started := time.Now()
	err := build(dir)
	t.Logf("build cold: %s %.6fs (toolchain=%s flags=%q input-files=%d)", inputs.Name, time.Since(started).Seconds(), inputs.Toolchain, inputs.Flags, len(inputs.Files))
	if err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	return dir
}

func yamlBuildCommand(dir, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := childguard.Run(command, childguard.Options{}); err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, stderr.Bytes())
	}
	if stderr.Len() != 0 {
		return fmt.Errorf("%s stderr: %s", name, stderr.Bytes())
	}
	return nil
}

type yamlUnit struct {
	name string
	ids  []string
}

func yamlValidateUnion(enumeration []string, units []yamlUnit) error {
	wanted := make(map[string]bool, len(enumeration))
	for _, id := range enumeration {
		if wanted[id] {
			return fmt.Errorf("repeated unsplit case %s", id)
		}
		wanted[id] = true
	}
	seen := make(map[string]bool, len(enumeration))
	names := make(map[string]bool, len(units))
	for _, unit := range units {
		if unit.name == "" || names[unit.name] {
			return fmt.Errorf("missing or repeated unit name %q", unit.name)
		}
		names[unit.name] = true
		for _, id := range unit.ids {
			if !wanted[id] {
				return fmt.Errorf("%s: unexpected case %s", unit.name, id)
			}
			if seen[id] {
				return fmt.Errorf("%s: repeated case %s", unit.name, id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(wanted) {
		for _, id := range enumeration {
			if !seen[id] {
				return fmt.Errorf("missing case %s (union %d, unsplit %d)", id, len(seen), len(wanted))
			}
		}
	}
	return nil
}

func yamlSelectedUnits(t *testing.T, count int) []bool {
	t.Helper()
	i, n := 0, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		pieces := strings.Split(value, "/")
		if len(pieces) != 2 {
			t.Fatalf("ADAMIC_TEST_SHARD must be zero-based i/n, got %q", value)
		}
		var err error
		i, err = strconv.Atoi(pieces[0])
		if err != nil {
			t.Fatal(err)
		}
		n, err = strconv.Atoi(pieces[1])
		if err != nil {
			t.Fatal(err)
		}
		if n <= 0 || i < 0 || i >= n {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	selected := make([]bool, count)
	for index := range selected {
		selected[index] = index%n == i
	}
	return selected
}

func yamlShardName(index, count int) string {
	width := max(3, len(strconv.Itoa(max(0, count-1))))
	return fmt.Sprintf("shard-%0*d", width, index)
}

func yamlRejectSurvivor(unit, side string, actual, expected []byte) error {
	if bytes.Equal(actual, expected) {
		return fmt.Errorf("%s: %s missed mutant", unit, side)
	}
	return nil
}

func TestYAMLUnitUnionRejectsMissingAndRepeatedCases(t *testing.T) {
	enumeration := []string{"case-0", "case-1"}
	for _, units := range [][]yamlUnit{
		{{name: "unit-0", ids: []string{"case-0"}}},
		{{name: "unit-0", ids: []string{"case-0", "case-1"}}, {name: "unit-1", ids: []string{"case-1"}}},
	} {
		if err := yamlValidateUnion(enumeration, units); err == nil {
			t.Fatal("invalid union accepted")
		}
	}
}

// Plant one surviving mutant/case in the same rejection path used by the live
// units. Both execution sides must report only the unit owning that case.
func TestFormatterMutantUnitCatchesPlantedSurvivor(t *testing.T) {
	const planted = 3
	wanted := []byte("ok\tplanted single case\n")
	for _, side := range []string{"native", "Node"} {
		caught := 0
		for index := range formatterMutants {
			unit := yamlShardName(index, len(formatterMutants))
			actual := []byte("ok\tkilled mutant\n")
			if index == planted {
				actual = append([]byte(nil), wanted...)
			}
			err := yamlRejectSurvivor(unit, side, actual, wanted)
			if err != nil {
				caught++
				if index != planted || !strings.Contains(err.Error(), unit) {
					t.Fatalf("wrong unit caught planted survivor: %v", err)
				}
				t.Logf("planted survivor case-0 caught by %s (%s): %v", unit, side, err)
			}
		}
		if caught != 1 {
			t.Fatalf("%s: planted survivor caught by %d units, want exactly 1", side, caught)
		}
	}
}

// Enforce Kirk's unit wall after product preparation and parallel scheduling.
func yamlUnitClock(t *testing.T) func() {
	started := time.Now()
	return func() {
		elapsed := time.Since(started)
		t.Logf("unit logic wall: %.6fs", elapsed.Seconds())
		if elapsed > 30*time.Second {
			t.Errorf("invalid test unit: %s ran for %s, maximum 30s", t.Name(), elapsed)
		}
	}
}

func yamlFormatterProducts(t *testing.T) (binary, emitted, entry, runner string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err = filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("*.ts")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, filepath.Join(root, "internal"), filepath.Join(root, "cohere"), filepath.Join(root, "go.mod"))
	lowered := yamlProduct(t, yamlProductInputs{
		Name: "format-lowered", Files: files, Flags: []string{"main.ts", "C", "JavaScript"}, Toolchain: "adamic",
	}, func(dir string) error {
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "format.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "format.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	options := native.Options{Sanitize: true}
	product := yamlProduct(t, yamlProductInputs{
		Name: "format-sanitized", Files: []string{filepath.Join(lowered, "format.c"), filepath.Join(root, "internal/native")}, Flags: native.Flags(options), Toolchain: "clang",
	}, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "format.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "format"), options)
	})
	return filepath.Join(product, "format"), filepath.Join(lowered, "format.mjs"), entry, filepath.Join(root, "oracle/node.mjs")
}

const (
	yamlFormatterCaseUnitSize  = 512
	yamlFileDriverCaseUnitSize = 1
)

type yamlCaseUnit struct {
	yamlUnit
	start, end int
}

func yamlCaseUnits(count, size int) ([]yamlCaseUnit, error) {
	if size <= 0 || count < 0 {
		return nil, fmt.Errorf("invalid case partition count=%d size=%d", count, size)
	}
	enumeration := make([]string, count)
	for index := range enumeration {
		enumeration[index] = fmt.Sprintf("case-%05d", index)
	}
	var result []yamlCaseUnit
	var units []yamlUnit
	for start := 0; start < count; start += size {
		end := min(start+size, count)
		unit := yamlUnit{name: yamlShardName(len(result), (count+size-1)/size)}
		for index := start; index < end; index++ {
			unit.ids = append(unit.ids, fmt.Sprintf("case-%05d", index))
		}
		result = append(result, yamlCaseUnit{yamlUnit: unit, start: start, end: end})
		units = append(units, unit)
	}
	if err := yamlValidateUnion(enumeration, units); err != nil {
		return nil, err
	}
	return result, nil
}

func yamlCompareCaseOutputs(unit, side string, actual, expected []byte) error {
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("%s: %s: %s", unit, side, firstDifference(actual, expected))
	}
	return nil
}

func TestYAMLFormatterCaseUnitCatchesPlantedDisagreement(t *testing.T) {
	for _, partition := range []struct {
		name                 string
		count, size, planted int
	}{
		{"formatter", 10026, yamlFormatterCaseUnitSize, 517},
		{"file-driver", 50, yamlFileDriverCaseUnitSize, 5},
	} {
		t.Run(partition.name, func(t *testing.T) {
			units, err := yamlCaseUnits(partition.count, partition.size)
			if err != nil {
				t.Fatal(err)
			}
			for _, side := range []string{"native ASan/UBSan/LSan", "Node", "emitted JavaScript"} {
				caught := 0
				for _, unit := range units {
					var wanted, actual bytes.Buffer
					for index := unit.start; index < unit.end; index++ {
						fmt.Fprintf(&wanted, "ok\tcase-%d\n", index)
						if index == partition.planted {
							actual.WriteString("ok\tplanted disagreement\n")
						} else {
							fmt.Fprintf(&actual, "ok\tcase-%d\n", index)
						}
					}
					err := yamlCompareCaseOutputs(unit.name, side, actual.Bytes(), wanted.Bytes())
					if err != nil {
						caught++
						if partition.planted < unit.start || partition.planted >= unit.end || !strings.Contains(err.Error(), unit.name) {
							t.Fatalf("wrong unit caught planted disagreement: %v", err)
						}
						t.Logf("planted case-%05d caught by %s (%s)", partition.planted, unit.name, side)
					}
				}
				if caught != 1 {
					t.Fatalf("%s disagreement caught by %d units, want 1", side, caught)
				}
			}
		})
	}
}
