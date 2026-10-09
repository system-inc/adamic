package tsprinter

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func expressionCorpus(t *testing.T, processPlan ...bool) (string, string, string) {
	t.Helper()
	root, _ := filepath.Abs(repository)
	files := printerCorpusFiles(t)
	gaps, _ := filepath.Abs("testdata/notyet.json")
	private := t.TempDir()
	side, _ := filepath.Abs("testdata/expressions_side_test.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{root + "/cohere/internal/format/javascript/adamic_expressions_test.go": side}})
	path := private + "/overlay.json"
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	// Go builds stay private until the GoBuild API covers this test overlay.
	command := bounded(t, "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-o="+private+"/oracle", "-overlay="+path, "./internal/format/javascript")
	command.Dir = root + "/cohere"
	if output, err := combinedOutput(command); err != nil {
		t.Fatalf("Go expression oracle: %v\n%s", err, output)
	}
	t.Logf("build Go expression oracle cold wall %.3fs (overlay, uncached)", time.Since(start).Seconds())
	directory := tsPrinterOracleOutputs(t, root, files, gaps, "expressions", private+"/oracle")
	answers, err := os.ReadFile(directory + "/answers.txt")
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_TS_PRINTER_KEEP"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"coverage.json", "cases.json", "cases.txt", "answers.txt", "gaps.txt", "gap-answers.txt", "gaps.json"} {
			data, err := os.ReadFile(directory + "/" + name)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(keep+"/"+name, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if len(processPlan) == 0 || processPlan[0] {
		planCorpus(t, directory+"/cases.txt")
	}
	return directory + "/cases.txt", string(answers), directory + "/cases.json"
}

const testExpressionsAgainstGoAndPrettierShards = 65

// ADAMIC_TEST_SHARD=i/n selects units whose zero-based index modulo n is i;
// unset runs all units. Builds are shared inputs, prepared before t.Parallel.
func TestExpressionsAgainstGoAndPrettier(t *testing.T) {
	cases, want, _ := expressionCorpus(t, false)
	plan, answers := expressionUnits(t, cases, want, testExpressionsAgainstGoAndPrettierShards-1)
	if got := len(plan.shards) + 1; got != testExpressionsAgainstGoAndPrettierShards {
		t.Fatalf("enumerated %d shards, declared %d", got, testExpressionsAgainstGoAndPrettierShards)
	}
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_TS_PRETTIER to prettier@3.9.6")
	}
	port, _ := filepath.Abs("main.ts")
	clang := executeOne(t, nil, "clang", "--version")
	if clang.exitCode != 0 {
		t.Fatalf("clang version: %s", clang.stderr)
	}
	loweredProduct := expressionBuild(t, printerBuildInputs{Name: "lowered expression program", Files: expressionInputFiles(t, repository), Toolchain: "Adamic " + runtime.Version()}, func(dir string) error {
		program := lowered(t, port)
		source := native.C(program)
		if err := os.WriteFile(dir+"/port.c", []byte(source), 0644); err != nil {
			return err
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
	data, err := os.ReadFile(loweredProduct + "/port.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)

	sanitized := expressionBuild(t, printerBuildInputs{Name: "sanitized expression binary", Files: append([]string{loweredProduct + "/port.c"}, expressionInputFiles(t, filepath.Join(repository, "internal/native"))...), Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: string(clang.stdout)}, func(dir string) error {
		return native.Build(source, dir+"/port", native.Options{Sanitize: true})
	}) + "/port"
	release := expressionBuild(t, printerBuildInputs{Name: "release expression binary", Files: append([]string{loweredProduct + "/port.c"}, expressionInputFiles(t, filepath.Join(repository, "internal/native"))...), Flags: native.Flags(native.Options{}), Toolchain: string(clang.stdout)}, func(dir string) error {
		return native.Build(source, dir+"/port", native.Options{})
	}) + "/port"
	if artifacts := os.Getenv("ADAMIC_TS_PRINTER_ARTIFACTS"); artifacts != "" {
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		for _, item := range []struct {
			from, to string
			mode     os.FileMode
		}{{loweredProduct + "/port.c", artifacts + "/port.c", 0644}, {release, artifacts + "/port", 0755}} {
			data, err := os.ReadFile(item.from)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(item.to, data, item.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	script, _ := filepath.Abs("testdata/expressions.mjs")
	embeddedScript, _ := filepath.Abs("testdata/embedded.mjs")
	bundles, _ := filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	selected := expressionSelection(t, len(plan.shards)+1)
	for number, shard := range plan.shards {
		if !selected[number] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			check := func(name string, result run) {
				if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
					t.Fatalf("%s: %v", name, err)
				}
			}
			check("Node", onNode(t, port, "--cases", shard.text, "80"))
			check("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, sanitized, "--cases", shard.text, "80"))
			check("backend", onNode(t, loweredProduct+"/program.mjs", "--cases", shard.text, "80"))
			if runtime.GOOS == "darwin" {
				report := execute(t, nil, "leaks", "--atExit", "--", release, "--cases", shard.text, "80")
				if report.exitCode != 0 {
					t.Fatalf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
				}
			} else if runtime.GOOS == "linux" {
				check("leaks", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, "--cases", shard.text, "80"))
			} else {
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
			check("native release", execute(t, nil, release, "--cases", shard.text, "80"))
			comparePrinterLibraryCases(t, "Prettier", execute(t, nil, "node", script, library, shard.specs), shard.specs, "expressions", false, false)
			comparePrinterLibraryCases(t, "embedded Prettier", execute(t, nil, "node", embeddedScript, bundles, shard.specs), shard.specs, "expressions", true, false)
		})
	}
	if selected[len(plan.shards)] {
		t.Run(fmt.Sprintf("shard-%03d", len(plan.shards)), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			directory := filepath.Dir(cases)
			gapWant, err := os.ReadFile(directory + "/gap-answers.txt")
			if err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name   string
				result run
			}{
				{"Node", onNode(t, port, "--cases", directory+"/gaps.txt", "80")},
				{"native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, "--cases", directory+"/gaps.txt", "80")},
				{"backend", onNode(t, loweredProduct+"/program.mjs", "--cases", directory+"/gaps.txt", "80")},
			} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 || string(side.result.stdout) != string(gapWant) {
					t.Fatalf("%s loud gaps: exit %d stderr %s diff %s", side.name, side.result.exitCode, side.result.stderr, firstDifference(string(side.result.stdout), string(gapWant)))
				}
			}
			result := execute(t, nil, "node", script, library, directory+"/gaps.json")
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("Go/Prettier gap proof: %s", result.stderr)
			}
		})
	}
}

// Inputs describe non-Go read-only shared products; Go builds stay separate.
type printerBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func expressionBuild(t *testing.T, inputs printerBuildInputs, build func(dir string) error) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.Inputs{Name: inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	for _, file := range inputs.Files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			key.Flags = append(key.Flags, fmt.Sprintf("generated:%s:%x", filepath.Base(file), sha256.Sum256(data)))
		} else {
			key.Files = append(key.Files, filepath.ToSlash(relative))
		}
	}
	key.Flags = append(key.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	return buildcache.Product(t, key, build)
}
