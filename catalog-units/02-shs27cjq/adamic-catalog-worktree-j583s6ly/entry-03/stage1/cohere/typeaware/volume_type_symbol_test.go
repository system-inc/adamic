package typeaware

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func typeSymbolCommand(directory string, command *exec.Cmd) error {
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", command, err, output)
	}
	return nil
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
		err := typeSymbolCommand(repository, exec.Command(stage0, args...))
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
	// GoBuild is not on this main base yet. Go products build once per invocation;
	// overlay mutant archives remain private, as required by buildcache.
	stage0 := filepath.Join(h.directory, "adamic")
	h.must("type-symbol-stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	oracle := volumeOracle(h, "type-symbol-oracle", "oracle_volume.go")
	overlay := h.overlay("type-symbol", "bridge/tsgo/checker/facts.go", "name = symbol.Name", "name = symbol.Name + \"wrong\"")
	archive := h.archive("type-symbol-checker", overlay, false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	inputs, build, err := typeSymbolNativeSpec(repository, stage0, "volume-type-symbol-native", entry, archive, false)
	if err != nil {
		t.Fatal(err)
	}
	mutant := filepath.Join(buildcache.Product(t, inputs, build), "native")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	cases := typeSymbolCases()
	var paths []string
	for _, id := range typeSymbolIDs(0) {
		name := fmt.Sprintf("control-%03d.ts", id)
		if id == len(cases)-2 {
			name = "native-globals.d.ts"
		}
		if id == len(cases)-1 {
			name = "native-console.ts"
		}
		paths = append(paths, h.write(name, cases[id]))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
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
		command := exec.Command(oracle, config, manifest)
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
	truth := result{stdout: truthBytes}
	observed := h.must("type-symbol-run", exec.Command(mutant, config, manifest))
	if err := typeSymbolSurvived(observed, truth); err != nil {
		t.Fatal(err)
	}
	t.Logf("type-symbol mutant caught at byte %d over %d cases", firstDifference(observed.stdout, truth.stdout), len(paths))
}
