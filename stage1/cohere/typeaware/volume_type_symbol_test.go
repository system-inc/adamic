package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
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
	timeoutContext, cancel := context.WithTimeout(context.Background(), limit)
	return typeSymbolExecContext(timeoutContext, name, args...), cancel
}

// Both setup and check commands retain cancellation of compiler descendants. No WaitDelay: the runs that read their
// output go through childguard.Capture, and one would only let TestVolumeTypeSymbolCommandDeadline pass on a Cancel
// that never killed its child.
func typeSymbolExecContext(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func typeSymbolCommand(directory string, name string, args ...string) error {
	// Building shared products has no test deadline; Loom bounds the unit.
	cancelContext, cancel := context.WithCancel(context.Background())
	command := typeSymbolExecContext(cancelContext, name, args...)
	defer cancel()
	command.Dir = directory
	// Read through childguard.Capture, to its end with no WaitDelay deciding when: exec's own copying under a
	// one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the build
	// exited, so a finished build could fail its product under load (#n3gk22h's shape).
	stdout, stderr, captureError := childguard.Capture(cancelContext, command)
	if captureError != nil {
		return fmt.Errorf("%s: %w\n%s%s", command, captureError, stdout, stderr)
	}
	return nil
}

func typeSymbolMust(h *harness, label, name string, args ...string) result {
	command, cancel := typeSymbolExec(90*time.Second, name, args...)
	defer cancel()
	return h.must(label, command)
}

func typeSymbolOracle(h *harness) string {
	return typeAwareOverlayOracle(h, "type-symbol-oracle", "adamic_type_symbol_oracle.go", "oracle_volume.go")
}

func typeSymbolFileHash(path string) (string, error) {
	data, readError := os.ReadFile(path)
	if readError != nil {
		return "", readError
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

// The callback writes only its product directory. Flags key fetched build inputs
// by bytes, not by cache location; Files cover transitive TypeScript imports.
func typeSymbolNativeSpec(
	repository, stage0, name, entry, archive string,
	sanitize bool,
) (buildcache.Inputs, func(string) error, error) {
	files, typeAwareTrackedTypeScriptError := typeAwareTrackedTypeScript(
		repository,
		"stage1/typescript",
		"stage1/cohere/lint",
		"stage1/cohere/typeaware",
	)
	if typeAwareTrackedTypeScriptError != nil {
		return buildcache.Inputs{}, nil, typeAwareTrackedTypeScriptError
	}
	var flags []string
	for _, input := range []struct{ name, path string }{{"stage0", stage0}, {"archive", archive}, {"entry", entry}} {
		sum, typeSymbolFileHashError := typeSymbolFileHash(input.path)
		if typeSymbolFileHashError != nil {
			return buildcache.Inputs{}, nil, typeSymbolFileHashError
		}
		flags = append(flags, input.name+"="+sum)
	}
	flags = append(flags, fmt.Sprintf("sanitize=%t", sanitize))
	for _, key := range []string{
		"ADAMIC_NATIVE_SPLIT",
		"CPATH",
		"C_INCLUDE_PATH",
		"LIBRARY_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	} {
		flags = append(flags, key+"="+os.Getenv(key))
	}
	inputs := buildcache.Inputs{Name: "typeaware-" + name, Files: files, Flags: flags, Toolchain: []string{
		buildcache.Tool("clang", "--version"),
		buildcache.Tool("clang", "-print-search-dirs"),
		buildcache.Tool("clang", "-print-resource-dir"),
	}}
	build := func(directory string) error {
		args := []string{"build", entry, "-o", filepath.Join(directory, "native"), "--tsgo", archive}
		if sanitize {
			args = append(args, "--sanitize")
		}
		started := time.Now()
		typeSymbolCommandError := typeSymbolCommand(repository, stage0, args...)
		fmt.Printf("typeSymbol-build %s cold=%.6fs\n", name, time.Since(started).Seconds())
		return typeSymbolCommandError
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

func TestVolumeTypeSymbolUnion(test *testing.T) {
	test.Parallel()
	seen := make(map[int]bool)
	for shard := 0; shard < testVolumeTypeSymbolShards; shard++ {
		for _, id := range typeSymbolIDs(shard) {
			if seen[id] {
				test.Fatalf("repeated case %d", id)
			}
			seen[id] = true
		}
	}
	for i := range typeSymbolCases() {
		if !seen[i] {
			test.Fatalf("missing case %d", i)
		}
	}
	if len(seen) != len(typeSymbolCases()) {
		test.Fatal("unexpected case in union")
	}
	test.Logf("union: %d live cases, %d shards", len(seen), testVolumeTypeSymbolShards)
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

func TestVolumeTypeSymbolPlantedSurvivor(test *testing.T) {
	test.Parallel()
	caught := []int{}
	// Plant identical finding bytes in the assigned shard, using the same validator.
	for shard := 0; shard < testVolumeTypeSymbolShards; shard++ {
		if len(typeSymbolIDs(shard)) != 0 &&
			typeSymbolSurvived(result{stdout: []byte("findings")}, result{stdout: []byte("findings")}) != nil {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != 0 {
		test.Fatalf("planted survivor caught by shards %v", caught)
	}
	test.Log("planted surviving type-symbol mutant caught exactly by TestVolumeTypeSymbol_000")
}

func typeSymbolArchive(h *harness) string {
	return plantedCheckerArchive(
		h.t,
		"type-symbol-checker.a",
		buildcache.Mutation{File: checkerFacts, Before: "name = symbol.Name", After: "name = symbol.Name + \"wrong\""},
	)
}

func typeSymbolNativeProduct(h *harness) string {
	stage0 := h.stage0()
	archive := typeSymbolArchive(h)
	inputs, build, typeSymbolNativeSpecError := typeSymbolNativeSpec(
		h.repository,
		stage0,
		"volume-type-symbol-native",
		filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts"),
		archive,
		false,
	)
	if typeSymbolNativeSpecError != nil {
		h.t.Fatal(typeSymbolNativeSpecError)
	}
	return filepath.Join(buildcache.Product(h.t, inputs, build), "native")
}

func typeSymbolFixtures(h *harness) ([]string, string) {
	cases := typeSymbolCases()
	data, encodeError := json.Marshal(cases)
	if encodeError != nil {
		h.t.Fatal(encodeError)
	}
	directory := buildcache.Product(h.t, buildcache.Inputs{
		Name:  "volume-type-symbol-fixtures",
		Flags: []string{string(data)},
	}, func(directory string) error {
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
	data, readError := os.ReadFile(manifest)
	if readError != nil {
		h.t.Fatal(readError)
	}
	var paths []string
	for _, name := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		paths = append(paths, filepath.Join(directory, name))
	}
	return paths, h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}

func typeSymbolTruth(h *harness, oracle string, paths []string, manifest string) result {
	test, repository := h.t, h.repository
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	// Every root and the config are product inputs; Go-oracle bytes cover its
	// embedded standard-library declarations and compiled production rules.
	oracleHash, typeSymbolFileHashError := typeSymbolFileHash(oracle)
	if typeSymbolFileHashError != nil {
		test.Fatal(typeSymbolFileHashError)
	}
	flags := []string{oracleHash, strings.Join(paths, "\n")}
	for _, path := range paths {
		sum, typeSymbolFileHashError := typeSymbolFileHash(path)
		if typeSymbolFileHashError != nil {
			test.Fatal(typeSymbolFileHashError)
		}
		flags = append(flags, sum)
	}
	outputInputs := buildcache.Inputs{
		Name:      "volume-type-symbol-oracle-output",
		Files:     []string{"stage1/cohere/typeaware/testdata/tsconfig.json"},
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("go", "version")},
	}
	outputDir := buildcache.Product(test, outputInputs, func(dir string) error {
		// The oracle's run is the product's fill, as typeSymbolCommand's builds are: no clock of its own, the unit
		// bounds it, so a builder's load never decides the product (#h4hrxpe).
		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()
		command := typeSymbolExecContext(cancelContext, oracle, config, manifest)
		command.Dir = repository
		// Read through childguard.Capture for typeSymbolCommand's reason: a WaitDelay could fail a run that had
		// finished but was slow to drain.
		stdout, stderr, captureError := childguard.Capture(cancelContext, command)
		if captureError != nil {
			return fmt.Errorf("oracle: %w %s", captureError, stderr)
		}
		// Each finding names its fixture in the fixtures product, by that product's key (#tqrqx60).
		return os.WriteFile(filepath.Join(dir, "findings"), []byte(buildcache.Relative(string(stdout))), 0444)
	})
	truthBytes, readError := os.ReadFile(filepath.Join(outputDir, "findings"))
	if readError != nil {
		test.Fatal(readError)
	}
	return result{stdout: []byte(buildcache.Absolute(string(truthBytes)))}
}

func TestVolumeTypeSymbol_000(test *testing.T) {
	test.Parallel()
	if selector := os.Getenv("ADAMIC_TEST_SHARD"); selector != "" {
		var i, n int
		if _, scanError := fmt.Sscanf(selector, "%d/%d", &i, &n); scanError != nil || n < 1 || i < 0 || i >= n {
			test.Fatalf("invalid ADAMIC_TEST_SHARD %q", selector)
		}
		if i != 0 {
			test.Skip("assigned to shard selector 0")
		}
	}
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	h := &harness{t: test, repository: repository, directory: test.TempDir()}
	oracle := typeSymbolOracle(h)
	mutant := typeSymbolNativeProduct(h)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	paths, manifest := typeSymbolFixtures(h)
	truth := typeSymbolTruth(h, oracle, paths, manifest)
	observed := typeSymbolMust(h, "type-symbol-run", mutant, config, manifest)
	if typeSymbolSurvivedError := typeSymbolSurvived(observed, truth); typeSymbolSurvivedError != nil {
		test.Fatal(typeSymbolSurvivedError)
	}
	test.Logf(
		"type-symbol mutant caught at byte %d over %d cases",
		firstDifference(observed.stdout, truth.stdout),
		len(paths),
	)
}

func TestVolumeTypeSymbolCommandDeadline(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TYPE_SYMBOL_DEADLINE_CHILD") == "1" {
		time.Sleep(30 * time.Second)
		return
	}
	command, cancel := typeSymbolExec(
		100*time.Millisecond,
		os.Args[0],
		"-test.run=^TestVolumeTypeSymbolCommandDeadline$",
	)
	defer cancel()
	command.Env = append(os.Environ(), "ADAMIC_TYPE_SYMBOL_DEADLINE_CHILD=1")
	started := time.Now()
	if runError := command.Run(); runError == nil {
		test.Fatal("deadline child escaped cancellation")
	}
	if time.Since(started) > 3*time.Second {
		test.Fatal("deadline failed to bound child")
	}
}
