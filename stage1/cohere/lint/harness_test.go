package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const testEmittedJavaScriptMismatchShards = 1

// ADAMIC_TEST_SHARD=i/n selects local case shards; unset runs all. The gate uses
// -run '^TestEmittedJavaScriptMismatch$/^shard-NNN$'. Setup shares builds once
// until internal/buildcache supplies read-only products by hash. The planted
// mismatch runs in a subprocess with those same products and must fail its shard.
func TestEmittedJavaScriptMismatch(t *testing.T) {
	t.Parallel()
	products, supplied := harnessSuppliedProducts(t)
	if !supplied {
		directory, err := filepath.Abs(".")
		if err != nil {
			t.Fatal(err)
		}
		products.Rows = []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"}
		oracle := harnessOracle(t, directory, "emitted-mismatch")
		archive := ""
		products.Original = harnessBuild(t, directory, "emitted-mismatch", &archive, true)
		products.Original.Oracle = oracle
	}
	harnessShards(t, []string{"no-var-emitted-javascript-mismatch"}, testEmittedJavaScriptMismatchShards, func(t *testing.T) {
		harnessCompare(t, products.Original, manifest(t, products.Rows))
		if !supplied {
			harnessPlantedDisagreement(t, products, false, "planted emitted JavaScript mismatch")
			t.Log("ordinary comparison rejected clean-running emitted JavaScript mutant")
		}
	})
}

const testDotARenameShards = 1

// ADAMIC_TEST_SHARD=i/n selects local case shards; unset runs all. The gate uses
// -run '^TestDotARename$/^shard-NNN$'. Until internal/buildcache is available,
// setup builds each product once and shares it with the leaf units.
func TestDotARename(t *testing.T) {
	t.Parallel()
	products, supplied := harnessSuppliedProducts(t)
	if !supplied {
		directory, err := filepath.Abs(".")
		if err != nil {
			t.Fatal(err)
		}
		products.Rows = []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"}
		oracle := harnessOracle(t, directory, "dot-a")
		archive := ""
		products.Original = harnessBuild(t, directory, "dot-a-original", &archive, true)
		products.Original.Oracle = oracle
		copied := mutant(t, "", "")
		entry := filepath.Join(copied, "rules/no-var/rule.a")
		products.Before, err = os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		renamed := filepath.Join(copied, "rules/no-var/rule.ts")
		if err := os.Rename(entry, renamed); err != nil {
			t.Fatal(err)
		}
		products.After, err = os.ReadFile(renamed)
		if err != nil {
			t.Fatal(err)
		}
		products.Changed = harnessBuild(t, copied, "dot-a-renamed", &archive, true)
		products.Changed.Oracle = oracle
	}
	harnessShards(t, []string{"no-var-a-to-ts"}, testDotARenameShards, func(t *testing.T) {
		if !bytes.Equal(products.Before, products.After) {
			t.Fatal("rename changed module bytes")
		}
		path := manifest(t, products.Rows)
		want := harnessCompare(t, products.Original, path)
		got := harnessCompare(t, products.Changed, path)
		if !bytes.Equal(got, want) {
			t.Fatal("rename changed results")
		}
		t.Logf("rename only: .ts and .a identical on all three runtimes against Go (%d bytes)", len(want))
		if !supplied {
			harnessPlantedDisagreement(t, products, true)
		}
	})
}

func serializationPort(t *testing.T) string {
	directory := mutant(t, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	return directory
}

const testCompleteSuggestionSerializationShards = 1

// ADAMIC_TEST_SHARD=i/n selects local case shards; unset runs all. The gate uses
// -run '^TestCompleteSuggestionSerialization$/^shard-NNN$'. Builds are shared
// once in setup until internal/buildcache can supply their products by hash.
func TestCompleteSuggestionSerialization(t *testing.T) {
	t.Parallel()
	products, supplied := harnessSuppliedProducts(t)
	if !supplied {
		directory := serializationPort(t)
		source := filepath.Join(t.TempDir(), "suggestions.ts")
		if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
			t.Fatal(err)
		}
		products.Rows = []string{source + "\tno-debugger"}
		oracle := harnessOracle(t, directory, "serialization")
		archive := ""
		products.Original = harnessBuild(t, directory, "serialization-original", &archive, true)
		products.Original.Oracle = oracle
		changedDirectory := serializationPort(t)
		changed := filepath.Join(changedDirectory, "rules/no-debugger/rule.a")
		data, err := os.ReadFile(changed)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Count(data, []byte("start + 1, start + 2, ''")) != 1 {
			t.Fatal("second suggestion edit anchor changed")
		}
		data = bytes.Replace(data, []byte("start + 1, start + 2, ''"), []byte("start + 1, start + 3, ''"), 1)
		if err := os.WriteFile(changed, data, 0644); err != nil {
			t.Fatal(err)
		}
		products.Changed = harnessBuild(t, changedDirectory, "serialization-edit-mutant", &archive, false)
	}
	harnessShards(t, []string{"unicode-complete-suggestions"}, testCompleteSuggestionSerializationShards, func(t *testing.T) {
		path := manifest(t, products.Rows)
		want := harnessCompare(t, products.Original, path)
		for _, field := range []string{"suggestion\tfirst", "suggestion\tsecond", "suggestion\tempty", "suggestion-edit\t8 9", "fixed\t/*"} {
			if !bytes.Contains(want, []byte(field)) {
				t.Fatalf("missing field %q: %s", field, want)
			}
		}
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"Node", node(t, products.Changed.Directory, path, false)},
			{"emitted JavaScript", runJavaScript(t, products.Changed.JavaScript, path, false)},
		} {
			if bytes.Equal(side.run.output, want) {
				t.Fatalf("second suggestion edit mutant survived on %s", side.name)
			}
			t.Logf("second suggestion edit mutant caught on %s: %s", side.name, difference(side.run.output, want))
		}
		if !supplied {
			harnessPlantedDisagreement(t, products, false)
		}
	})
}

const testSuggestionAlongsideAutomaticFixShards = 1

// ADAMIC_TEST_SHARD=i/n selects local case shards; unset runs all. The gate uses
// -run '^TestSuggestionAlongsideAutomaticFix$/^shard-NNN$'. Setup shares builds
// once until internal/buildcache supplies read-only products by hash.
func TestSuggestionAlongsideAutomaticFix(t *testing.T) {
	t.Parallel()
	products, supplied := harnessSuppliedProducts(t)
	if !supplied {
		directory := serializationPort(t)
		for _, change := range []struct{ name, from, to string }{
			{"rule.a", "debuggerMessage, '', '', ''", "debuggerMessage, 'fix', ';', ''"},
			{"oracle.go", "d.Fixes = nil", "d.Fixes[0].Text = \";\""},
		} {
			path := filepath.Join(directory, "rules/no-debugger", change.name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Count(data, []byte(change.from)) != 1 {
				t.Fatal("automatic fix anchor changed")
			}
			if err := os.WriteFile(path, bytes.Replace(data, []byte(change.from), []byte(change.to), 1), 0644); err != nil {
				t.Fatal(err)
			}
		}
		source := filepath.Join(t.TempDir(), "mixed.ts")
		if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
			t.Fatal(err)
		}

		products.Rows = []string{source + "\tno-debugger"}
		oracle := harnessOracle(t, directory, "suggestions-with-fix")
		archive := ""
		products.Original = harnessBuild(t, directory, "suggestions-with-fix", &archive, true)
		products.Original.Oracle = oracle
	}
	harnessShards(t, []string{"unicode-suggestions-with-automatic-fix"}, testSuggestionAlongsideAutomaticFixShards, func(t *testing.T) {
		got := harnessCompare(t, products.Original, manifest(t, products.Rows))
		if !bytes.Contains(got, []byte("fixed\t/*\\ud83d\\ude00*/;\\u000a")) {
			t.Fatalf("automatic fix lost: %s", got)
		}
		t.Log("automatic fix remains applied while all three suggestions remain unapplied and serialized")
		if !supplied {
			harnessPlantedDisagreement(t, products, false)
		}
	})
}

const testWitnessScriptKindShards = 1

// ADAMIC_TEST_SHARD=i/n selects deterministic case shards (zero-based); unset runs all.
// The gate invokes each unit as -run '^TestWitnessScriptKind$/^shard-NNN$'.
// Until internal/buildcache lands, setup builds its products once and shares them;
// its measured time includes those builds.
// Build products are prepared once before parallel units. The original enumeration has one case.
func TestWitnessScriptKind(t *testing.T) {
	t.Parallel()
	products := witnessProducts(t)
	ids := []string{"no-debugger-tsx"}
	shards := checkedCaseShards(t, ids, len(ids))
	if len(shards) != testWitnessScriptKindShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testWitnessScriptKindShards)
	}
	for i, cases := range shards {
		if !selectedCaseShard(t, i) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			t.Logf("case ids: %v", cases)
			for range cases {
				if filepath.Ext(products.Source) != ".tsx" {
					t.Fatal("witness script kind lost")
				}
				compareWithJavaScript(t, products.Oracle, products.Native, products.Directory,
					manifest(t, []string{products.Source + "\tno-debugger"}), products.JavaScript)
			}
			if os.Getenv("ADAMIC_LINT_WITNESS_PRODUCTS") == "" {
				checkWitnessPlantedDisagreement(t, products)
			}
		})
	}
}

// Each product maker writes only into its supplied directory; its complete input
// description is adjacent. These are shared once per run, not a package cache.
// Replace unitProduct with internal/buildcache.Product when it reaches this base.
type harnessRunProducts struct {
	Original, Changed witnessUnitProducts
	Rows              []string
	Before, After     []byte
}

func harnessSuppliedProducts(t *testing.T) (harnessRunProducts, bool) {
	t.Helper()
	var p harnessRunProducts
	path := os.Getenv("ADAMIC_LINT_HARNESS_PRODUCTS")
	if path == "" {
		return p, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	return p, true
}

func harnessInputs(t *testing.T, directory, name string, flags []string) unitBuildInputs {
	t.Helper()
	version, err := exec.Command("go", "version").Output()
	if err != nil {
		t.Fatal(err)
	}
	clang, err := exec.Command("clang", "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	return unitBuildInputs{Name: name, Files: unitSourceFiles(t, directory,
		filepath.Join(repository, "internal"), filepath.Join(repository, "bridge"),
		filepath.Join(repository, "cohere"), filepath.Join(repository, "go.mod"),
		filepath.Join(repository, "go.work"), filepath.Join(packageDirectory, "testdata/oracle.go")),
		Flags: flags, Toolchain: strings.TrimSpace(string(version)) + "; " + strings.Split(string(clang), "\n")[0]}
}

func harnessOracle(t *testing.T, directory, name string) string {
	t.Helper()
	prepareRegistry(t, directory)
	inputs := harnessInputs(t, directory, name+"-go-oracle", []string{"build", "-overlay"})
	product := unitProduct(t, inputs, func(dir string) error {
		_, err := goOracleIn(directory, dir, unitGoEnvironment(dir)...)
		return err
	})
	return filepath.Join(product, "oracle")
}

func harnessBuild(t *testing.T, directory, name string, archive *string, sanitizedNative bool) witnessUnitProducts {
	t.Helper()
	prepareRegistry(t, directory)
	inputs := harnessInputs(t, directory, name+"-lowered-C-and-JavaScript", []string{"EnableTSGo"})
	bridge := false
	lowered := unitProduct(t, inputs, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		program.EnableTSGo()
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		bridge = native.UsesTSGo(ir)
		c := ""
		if bridge {
			c, err = native.TSGoC(ir)
		} else {
			c = native.C(ir)
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "lint.c"), []byte(c), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "lint.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	p := witnessUnitProducts{Directory: directory, JavaScript: filepath.Join(lowered, "lint.mjs")}
	if !sanitizedNative {
		return p
	}
	if bridge && *archive == "" {
		archiveInputs := inputs
		archiveInputs.Name = name + "-sanitized-checker-archive"
		archiveInputs.Flags = []string{"-buildmode=c-archive", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"}
		product := unitProduct(t, archiveInputs, func(dir string) error {
			root, err := filepath.Abs(repository)
			if err != nil {
				return err
			}
			_, err = run(root, append(unitGoEnvironment(dir), "GOMAXPROCS=4", "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"), "go", "build", "-buildmode=c-archive", "-o", filepath.Join(dir, "checker.a"), "./bridge/tsgo/archive")
			return err
		})
		*archive = filepath.Join(product, "checker.a")
	}
	nativeInputs := inputs
	nativeInputs.Name = name + "-sanitized-native"
	nativeInputs.Files = append(unitSourceFiles(t, filepath.Join(repository, "internal/native")), filepath.Join(lowered, "lint.c"))
	if bridge {
		nativeInputs.Files = append(nativeInputs.Files, *archive)
	}
	nativeInputs.Flags = native.Flags(native.Options{Sanitize: true})
	product := unitProduct(t, nativeInputs, func(dir string) error {
		data, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		return nativeBuild(func() error {
			if bridge {
				return buildCheckerWithRuntime(string(data), filepath.Join(dir, "scanner"), *archive, native.Options{Sanitize: true})
			}
			return native.Build(string(data), filepath.Join(dir, "scanner"), native.Options{Sanitize: true})
		})
	})
	p.Native = filepath.Join(product, "scanner")
	return p
}

func harnessCompare(t *testing.T, p witnessUnitProducts, path string) []byte {
	t.Helper()
	return compareWithJavaScript(t, p.Oracle, p.Native, p.Directory, path, p.JavaScript)
}

func harnessShards(t *testing.T, ids []string, count int, check func(*testing.T)) {
	t.Helper()
	// One original case per shard. Validate enumeration independently of the literal.
	shards := checkedCaseShards(t, ids, len(ids))
	if len(shards) != count {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), count)
	}
	for i, cases := range shards {
		if !selectedCaseShard(t, i) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) {
			t.Parallel()
			t.Logf("case ids: %v", cases)
			check(t)
		})
	}
}

func harnessPlantedDisagreement(t *testing.T, products harnessRunProducts, changed bool, marker ...string) {
	t.Helper()
	message := "planted harness disagreement"
	if len(marker) != 0 {
		message = marker[0]
	}
	p := &products.Original
	if changed {
		p = &products.Changed
	}
	data, err := os.ReadFile(p.JavaScript)
	if err != nil {
		t.Fatal(err)
	}
	p.JavaScript = filepath.Join(t.TempDir(), "planted.mjs")
	if err := os.WriteFile(p.JavaScript, append(data, []byte(fmt.Sprintf("\nconsole.log(%q);\n", message))...), 0644); err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(products)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "products.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	name := strings.Split(t.Name(), "/")[0]
	for seat := 0; seat < 2; seat++ {
		command := exec.Command(os.Args[0], "-test.run=^"+name+"$/^shard-000$", "-test.v", "-test.timeout=30s")
		command.Env = append(os.Environ(), "ADAMIC_LINT_HARNESS_PRODUCTS="+path, fmt.Sprintf("ADAMIC_TEST_SHARD=%d/2", seat))
		output, err := command.CombinedOutput()
		text := string(output)
		if seat == 0 {
			if err == nil || strings.Count(text, "--- FAIL: "+name+"/shard-000") != 1 || !strings.Contains(text, "emitted JavaScript:") || !strings.Contains(text, message) {
				t.Fatalf("wrong planted failure: %v\n%s", err, text)
			}
			t.Log("planted disagreement caught exactly by shard-000 on seat 0/2")
		} else if err != nil || strings.Contains(text, "--- FAIL:") || strings.Contains(text, "=== RUN   "+name+"/shard-") {
			t.Fatalf("non-owning seat ran or caught planted case: %v\n%s", err, text)
		}
	}
}
