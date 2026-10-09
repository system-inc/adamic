package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Bound the complete process group, including compiler grandchildren. Unix
// process groups work on both Linux and macOS without an external timeout tool.
func typeSymbolExec(limit time.Duration, name string, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command, cancel
}

func typeSymbolCommand(directory string, name string, args ...string) error {
	command, cancel := typeSymbolExec(90*time.Second, name, args...)
	defer cancel()
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", command, err, output)
	}
	return nil
}

func typeSymbolMust(h *harness, label, name string, args ...string) result {
	command, cancel := typeSymbolExec(90*time.Second, name, args...)
	defer cancel()
	return h.must(label, command)
}

func typeSymbolOracle(h *harness) string {
	return typeAwareOverlayOracle(h, "type-symbol-oracle-build", "type-symbol-oracle", "adamic_type_symbol_oracle.go", "oracle_volume.go", "oracle-overlay.json", false)
}

func typeSymbolFileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// The callback writes only its product directory. Flags key fetched build inputs
// by bytes, not by cache location; Files cover transitive TypeScript imports.
func typeSymbolNativeSpec(repository, stage0, name, entry, archive string, sanitize bool) (buildcache.Inputs, func(string) error, error) {
	var files []string
	for _, root := range []string{"stage1/typescript", "stage1/cohere/lint", "stage1/cohere/typeaware"} {
		if err := filepath.WalkDir(filepath.Join(repository, root), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				files = append(files, relative)
			}
			return nil
		}); err != nil {
			return buildcache.Inputs{}, nil, err
		}
	}
	var flags []string
	for _, input := range []struct{ name, path string }{{"stage0", stage0}, {"archive", archive}, {"entry", entry}} {
		sum, err := typeSymbolFileHash(input.path)
		if err != nil {
			return buildcache.Inputs{}, nil, err
		}
		flags = append(flags, input.name+"="+sum)
	}
	flags = append(flags, fmt.Sprintf("sanitize=%t", sanitize))
	for _, key := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		flags = append(flags, key+"="+os.Getenv(key))
	}
	inputs := buildcache.Inputs{Name: "typeaware-" + name, Files: files, Flags: flags, Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool("clang", "-print-search-dirs"), buildcache.Tool("clang", "-print-resource-dir")}}
	build := func(directory string) error {
		args := []string{"build", entry, "-o", filepath.Join(directory, "native"), "--tsgo", archive}
		if sanitize {
			args = append(args, "--sanitize")
		}
		started := time.Now()
		err := typeSymbolCommand(repository, stage0, args...)
		fmt.Printf("typeSymbol-build %s cold=%.6fs\n", name, time.Since(started).Seconds())
		return err
	}
	return inputs, build, nil
}

// This generated control set is enumerated live; no total is hard-coded.
// ADAMIC_TEST_SHARD=i/n (zero based) selects shard 000; unset runs it.
const testVolumeTypeSymbolShards = 1

func typeSymbolCases() []string {
	cases := append([]string(nil), volumeControls()...)
	for i := range cases {
		cases[i] += "\nexport {};\n"
	}
	return append(cases, "declare const console: {log():void};\n", "const detached=console.log;\nexport {};\n")
}

func typeSymbolIDs(shard int) []int {
	var ids []int
	for i := range typeSymbolCases() {
		if i%testVolumeTypeSymbolShards == shard {
			ids = append(ids, i)
		}
	}
	return ids
}

func TestVolumeTypeSymbolUnion(t *testing.T) {
	t.Parallel()
	seen := make(map[int]bool)
	for shard := 0; shard < testVolumeTypeSymbolShards; shard++ {
		for _, id := range typeSymbolIDs(shard) {
			if seen[id] {
				t.Fatalf("repeated case %d", id)
			}
			seen[id] = true
		}
	}
	for i := range typeSymbolCases() {
		if !seen[i] {
			t.Fatalf("missing case %d", i)
		}
	}
	if len(seen) != len(typeSymbolCases()) {
		t.Fatal("unexpected case in union")
	}
	t.Logf("union: %d live cases, %d shards", len(seen), testVolumeTypeSymbolShards)
}

func typeSymbolSurvived(observed, truth result) error {
	if observed.err != nil || len(observed.stderr) != 0 {
		return fmt.Errorf("mutant execution: %v %s", observed.err, observed.stderr)
	}
	if bytes.Equal(observed.stdout, truth.stdout) {
		return fmt.Errorf("type-symbol checker question mutant survived")
	}
	return nil
}

func TestVolumeTypeSymbolPlantedSurvivor(t *testing.T) {
	t.Parallel()
	caught := []int{}
	// Plant identical finding bytes in the assigned shard, using the same validator.
	for shard := 0; shard < testVolumeTypeSymbolShards; shard++ {
		if len(typeSymbolIDs(shard)) != 0 && typeSymbolSurvived(result{stdout: []byte("findings")}, result{stdout: []byte("findings")}) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 0 {
		t.Fatalf("planted survivor caught by shards %v", caught)
	}
	t.Log("planted surviving type-symbol mutant caught exactly by TestVolumeTypeSymbol_000")
}

func typeSymbolArchive(h *harness) string {
	overlay := h.overlay("type-symbol", "bridge/tsgo/checker/facts.go", "name = symbol.Name", "name = symbol.Name + \"wrong\"")
	command := exec.Command("go", "build", "-buildmode=c-archive", "-o", filepath.Join(h.directory, "type-symbol-checker.a"), "-overlay", overlay, "./bridge/tsgo/archive")
	return h.sixBuildProduct("type-symbol-checker", command)
}

func typeSymbolNativeProduct(h *harness) string {
	stage0 := h.stage0()
	archive := typeSymbolArchive(h)
	inputs, build, err := typeSymbolNativeSpec(h.repository, stage0, "volume-type-symbol-native", filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"), archive, false)
	if err != nil {
		h.t.Fatal(err)
	}
	return filepath.Join(buildcache.Product(h.t, inputs, build), "native")
}

func typeSymbolFixtures(h *harness) ([]string, string) {
	cases := typeSymbolCases()
	data, err := json.Marshal(cases)
	if err != nil {
		h.t.Fatal(err)
	}
	directory := buildcache.Product(h.t, buildcache.Inputs{Name: "volume-type-symbol-fixtures", Flags: []string{string(data)}}, func(directory string) error {
		local := &harness{t: h.t, repository: h.repository, directory: directory}
		var paths []string
		for _, id := range typeSymbolIDs(0) {
			name := fmt.Sprintf("control-%03d.ts", id)
			if id == len(cases)-2 {
				name = "native-globals.d.ts"
			}
			if id == len(cases)-1 {
				name = "native-console.ts"
			}
			paths = append(paths, local.write(name, cases[id]))
		}
		var names []string
		for _, path := range paths {
			names = append(names, filepath.Base(path))
		}
		manifest := local.write("controls.manifest", strings.Join(names, "\n")+"\n")

		_ = manifest
		return nil
	})
	manifest := filepath.Join(directory, "controls.manifest")
	data, err = os.ReadFile(manifest)
	if err != nil {
		h.t.Fatal(err)
	}
	var paths []string
	for _, name := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		paths = append(paths, filepath.Join(directory, name))
	}
	return paths, h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}

func typeSymbolTruth(h *harness, oracle string, paths []string, manifest string) result {
	t, repository := h.t, h.repository
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	// Every root and the config are product inputs; Go-oracle bytes cover its
	// embedded standard-library declarations and compiled production rules.
	oracleHash, err := typeSymbolFileHash(oracle)
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{oracleHash, strings.Join(paths, "\n")}
	for _, path := range paths {
		sum, err := typeSymbolFileHash(path)
		if err != nil {
			t.Fatal(err)
		}
		flags = append(flags, sum)
	}
	outputInputs := buildcache.Inputs{Name: "volume-type-symbol-oracle-output", Files: []string{"stage1/cohere/typeaware/testdata/tsconfig.json"}, Flags: flags, Toolchain: []string{buildcache.Tool("go", "version")}}
	outputDir := buildcache.Product(t, outputInputs, func(dir string) error {
		command, cancel := typeSymbolExec(90*time.Second, oracle, config, manifest)
		defer cancel()
		command.Dir = repository
		var stdout, stderr bytes.Buffer
		command.Stdout = &stdout
		command.Stderr = &stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("oracle: %w %s", err, stderr.Bytes())
		}
		return os.WriteFile(filepath.Join(dir, "findings"), stdout.Bytes(), 0444)
	})
	truthBytes, err := os.ReadFile(filepath.Join(outputDir, "findings"))
	if err != nil {
		t.Fatal(err)
	}
	return result{stdout: truthBytes}
}

func TestVolumeTypeSymbol_000(t *testing.T) {
	t.Parallel()
	if selector := os.Getenv("ADAMIC_TEST_SHARD"); selector != "" {
		var i, n int
		if _, err := fmt.Sscanf(selector, "%d/%d", &i, &n); err != nil || n < 1 || i < 0 || i >= n {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", selector)
		}
		if i != 0 {
			t.Skip("assigned to shard selector 0")
		}
	}
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: t.TempDir()}
	oracle := typeSymbolOracle(h)
	mutant := typeSymbolNativeProduct(h)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths, manifest := typeSymbolFixtures(h)
	truth := typeSymbolTruth(h, oracle, paths, manifest)
	observed := typeSymbolMust(h, "type-symbol-run", mutant, config, manifest)
	if err := typeSymbolSurvived(observed, truth); err != nil {
		t.Fatal(err)
	}
	t.Logf("type-symbol mutant caught at byte %d over %d cases", firstDifference(observed.stdout, truth.stdout), len(paths))
}

func TestVolumeTypeSymbolCommandDeadline(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_TYPE_SYMBOL_DEADLINE_CHILD") == "1" {
		time.Sleep(30 * time.Second)
		return
	}
	command, cancel := typeSymbolExec(100*time.Millisecond, os.Args[0], "-test.run=^TestVolumeTypeSymbolCommandDeadline$")
	defer cancel()
	command.Env = append(os.Environ(), "ADAMIC_TYPE_SYMBOL_DEADLINE_CHILD=1")
	started := time.Now()
	if err := command.Run(); err == nil {
		t.Fatal("deadline child escaped cancellation")
	}
	if time.Since(started) > 3*time.Second {
		t.Fatal("deadline failed to bound child")
	}
}
