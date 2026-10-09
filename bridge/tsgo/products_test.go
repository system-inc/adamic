package tsgo_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

var bridgeProducts string
var bridgeGoProducts map[string]string
var bridgeProductNames = []string{"tsgo.a", "tsgo-asan.a", "length.a", "stale.a", "wrong.a", "leak.a", "stage0", "oracle", "api", "length-driver", "stale-driver", "leak-driver", "native-asan", "native", "wrong-native", "healthy-region", "region-stage0", "region-native", "linkage.test"}

func TestMain(m *testing.M) {
	flag.Parse()
	prepare := os.Getenv("ADAMIC_TSGO_PREPARE") == "1"
	selected, err := regexp.Compile(flag.Lookup("test.run").Value.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	needsProducts := prepare
	for index, unit := range bridgeCases {
		if selected.MatchString(unit.name) && bridgeCaseActive(index) {
			needsProducts = true
		}
	}
	if needsProducts {
		repository, err := filepath.Abs("../..")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bridgeGoProducts, err = prepareBridgeGoProducts(repository)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		key, err := bridgeProductKey(repository)
		if err == nil {
			bridgeProducts, err = bridgeProductGet(key, func(directory string) error {
				if err := buildBridgeProducts(repository, directory); err != nil {
					return err
				}
				after, err := bridgeProductKey(repository)
				if err != nil {
					return err
				}
				if after != key {
					return fmt.Errorf("bridge build inputs changed during preparation")
				}
				return nil
			})
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "bridge products: key=%s directory=%s\n", key, bridgeProducts)
	}
	if prepare {
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func bridgeProductKey(repository string) (string, error) {
	// Go validates dependency actions against content, including local replacements
	// and embeds. Build IDs avoid reimplementing the checker dependency graph.
	command := exec.Command("go", "list", "-deps", "-export", "-json", "./bridge/tsgo/archive", "./bridge/tsgo/oracle", "./cmd/adamic", "./internal/buildcache")
	command.Dir = repository
	data, err := command.Output()
	if err != nil {
		return "", err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	hash := sha256.New()
	fmt.Fprintln(hash, "bridge-test-products-v1", repository)
	for {
		var dependency struct{ ImportPath, BuildID string }
		err := decoder.Decode(&dependency)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%q %q\n", dependency.ImportPath, dependency.BuildID)
	}
	testFiles, err := filepath.Glob(filepath.Join(repository, "bridge/tsgo/*_test.go"))
	if err != nil {
		return "", err
	}
	names := []string{"bridge/tsgo/testdata/api.c", "bridge/tsgo/testdata/queries.a", "bridge/tsgo/testdata/region.a", "bridge/tsgo/tsgo.h"}
	for _, path := range testFiles {
		name, err := filepath.Rel(repository, path)
		if err != nil {
			return "", err
		}
		names = append(names, name)
	}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%q %x\n", name, sha256.Sum256(data))
	}
	for _, tool := range [][]string{{"go", "version"}, {"clang", "--version"}, {"go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOFLAGS", "GOEXPERIMENT", "CC", "CXX"}} {
		data, err := exec.Command(tool[0], tool[1:]...).CombinedOutput()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%q %q\n", tool, data)
	}
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "WASI_SYSROOT"} {
		fmt.Fprintf(hash, "%q %q\n", name, os.Getenv(name))
	}
	fmt.Fprintln(hash, "c-archive; sanitized CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all; native --sanitize; region --count; clang c11 -Wall -Wextra -Werror -pedantic -O1 -g -fsanitize=address,undefined -lpthread -ldl -lm")
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func buildBridgeProducts(repository, directory string) error {
	run := func(name string, command *exec.Cmd) error {
		command.Dir = repository
		output, err := command.CombinedOutput()
		if write := os.WriteFile(filepath.Join(directory, name+".log"), output, 0o644); write != nil {
			return write
		}
		if err != nil {
			return fmt.Errorf("%s: %w\n%s", name, err, output)
		}
		return nil
	}
	overlay := func(name, path, before, after string) (string, error) {
		original := filepath.Join(repository, path)
		text, err := os.ReadFile(original)
		if err != nil {
			return "", err
		}
		if strings.Count(string(text), before) != 1 {
			return "", fmt.Errorf("mutant %s has no unique target", name)
		}
		replacement := filepath.Join(directory, name+filepath.Ext(path))
		if err := os.WriteFile(replacement, []byte(strings.Replace(string(text), before, after, 1)), 0o644); err != nil {
			return "", err
		}
		data, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
		if err != nil {
			return "", err
		}
		output := filepath.Join(directory, name+"-overlay.json")
		return output, os.WriteFile(output, data, 0o644)
	}
	archive := func(name, replacement string, sanitize bool) error {
		arguments := []string{"build", "-buildmode=c-archive", "-o", filepath.Join(directory, name)}
		if replacement != "" {
			arguments = append(arguments, "-overlay", replacement)
		}
		command := exec.Command("go", append(arguments, "./bridge/tsgo/archive")...)
		if sanitize {
			command.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
		}
		return run(name+"-build", command)
	}
	// Ordinary products use the shared Go cache; overlays remain in this keyed bundle.
	for name, source := range bridgeGoProducts {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0o755); err != nil {
			return err
		}
	}
	for _, mutant := range []struct{ name, path, before, after string }{
		{"length", "bridge/tsgo/archive/main.go", "result._type = buffer(answer.Type)", "result._type = buffer(answer.Type)\n\tresult._type.length++"},
		{"stale", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant: retain the released program in the live registry."},
		{"wrong", "bridge/tsgo/checker/program.go", "Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(node))", "Type: typeChecker.TypeToString(typeChecker.GetTypeAtLocation(source.AsNode()))"},
		{"leak", "bridge/tsgo/archive/boundary.c", "free(buffer->data);", "/* Mutant: abandon the C output allocation. */"},
	} {
		replacement, err := overlay(mutant.name, mutant.path, mutant.before, mutant.after)
		if err != nil {
			return err
		}
		if err := archive(mutant.name+".a", replacement, false); err != nil {
			return err
		}
	}
	replacement, err := overlay("region-ownership", "internal/native/runtime/tsgo.c", "adamic_object_new_in(region, &shape)", "adamic_object_new(&shape)")
	if err != nil {
		return err
	}
	if err := run("region-stage0-build", exec.Command("go", "build", "-overlay", replacement, "-o", filepath.Join(directory, "region-stage0"), "./cmd/adamic")); err != nil {
		return err
	}
	replacement, err = overlay("linkage", "internal/lower/tsgo.go", "if !l.program.TSGoEnabled() {", "if false {")
	if err != nil {
		return err
	}
	if err := run("linkage-build", exec.Command("go", "test", "-c", "-overlay", replacement, "-o", filepath.Join(directory, "linkage.test"), "./bridge/tsgo")); err != nil {
		return err
	}
	for _, driver := range []struct{ name, archive string }{{"api", "tsgo-asan.a"}, {"length-driver", "length.a"}, {"stale-driver", "stale.a"}, {"leak-driver", "leak.a"}} {
		if err := run(driver.name+"-link", exec.Command("clang", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O1", "-g", "-fsanitize=address,undefined", "-I", filepath.Join(repository, "bridge/tsgo"), filepath.Join(repository, "bridge/tsgo/testdata/api.c"), filepath.Join(directory, driver.archive), "-lpthread", "-ldl", "-lm", "-o", filepath.Join(directory, driver.name))); err != nil {
			return err
		}
	}
	for _, binary := range []struct {
		name, compiler, fixture, archive string
		sanitized, counted               bool
	}{
		{"native-asan", "stage0", "queries.a", "tsgo-asan.a", true, false},
		{"native", "stage0", "queries.a", "tsgo.a", false, false},
		{"wrong-native", "stage0", "queries.a", "wrong.a", true, false},
		{"healthy-region", "stage0", "region.a", "tsgo-asan.a", true, true},
		{"region-native", "region-stage0", "region.a", "tsgo.a", true, false},
	} {
		arguments := []string{"build", filepath.Join(repository, "bridge/tsgo/testdata", binary.fixture), "-o", filepath.Join(directory, binary.name), "--tsgo", filepath.Join(directory, binary.archive)}
		if binary.sanitized {
			arguments = append(arguments, "--sanitize")
		}
		if binary.counted {
			arguments = append(arguments, "--count")
		}
		if err := run(binary.name+"-build", exec.Command(filepath.Join(directory, binary.compiler), arguments...)); err != nil {
			return err
		}
	}
	hashes := map[string]string{}
	for _, name := range bridgeProductNames {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return err
		}
		hashes[name] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	data, err := json.MarshalIndent(hashes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "products.json"), data, 0o644)
}

// TestMain has no testing.TB. Use GoBuild's exported key and cache entry point,
// with its reproducible flags, so these products share GoBuild's cache entries.
func prepareBridgeGoProducts(repository string) (map[string]string, error) {
	products := map[string]string{}
	for _, target := range []struct {
		name, output, pkg      string
		arguments, environment []string
	}{
		{"tsgo.a", "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, nil},
		{"tsgo-asan.a", "tsgo-asan.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, []string{"CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"}},
		{"stage0", "adamic", "./cmd/adamic", nil, nil},
		{"oracle", "oracle", "./bridge/tsgo/oracle", nil, nil},
	} {
		arguments := append([]string{"-trimpath", "-ldflags=-buildid="}, target.arguments...)
		inputs, err := buildcache.GoInputs(target.output, target.pkg, arguments, target.environment)
		if err != nil {
			return nil, err
		}
		directory, err := buildcache.Get(inputs, func(directory string) error {
			command := exec.Command("go", append(append(append([]string{"build"}, arguments...), "-o", filepath.Join(directory, target.output)), target.pkg)...)
			command.Dir = repository
			command.Env = append(os.Environ(), target.environment...)
			output, err := command.CombinedOutput()
			if write := os.WriteFile(filepath.Join(directory, "build.log"), output, 0o644); write != nil {
				return write
			}
			if err != nil {
				return fmt.Errorf("go build %s: %w\n%s", target.pkg, err, output)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		products[target.name] = filepath.Join(directory, target.output)
	}
	return products, nil
}
