package printer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// Only buildcache products survive a test's cleanup. Every selected shard
// fetches its own prerequisites, once per process, before its work deadline.
type printerAsGoOnce struct {
	once sync.Once
	path string
}

func (p *printerAsGoOnce) get(t *testing.T, build func() string) string {
	t.Helper()
	p.once.Do(func() { p.path = build() })
	if p.path == "" {
		t.Fatal("GraphQL product preparation failed")
	}
	return p.path
}

var asGoSource, asGoOracle, asGoLowered, asGoSanitized, asGoRelease printerAsGoOnce
var asGoCorpus [4]struct {
	once       sync.Once
	path, want string
}
var asGoModes = [...]string{"defaults", "narrow", "tight", "tabs"}

func printerAsGoSource(t *testing.T) string {
	t.Helper()
	return asGoSource.get(t, func() string {
		files := []string{"../token.ts", "../characterClasses.ts", "../blockString.ts", "../lexer.ts", "../parser.ts", "../../json/width.ts", "../../json/widthTables.ts", "doc.ts", "printer.ts", "main.ts", "printer_test.go", "grain_asgo_test.go"}
		return printerBuild(t, printerBuildInputs{Name: "GraphQL printer source", Files: printerInputFiles(t, files...), Toolchain: runtime.Version()}, func(dir string) error {
			printerDirectoryAt(t, dir, "", "", "")
			return nil
		}) + "/graphql/printer/main.ts"
	})
}

func printerAsGoOracle(t *testing.T) string {
	t.Helper()
	return asGoOracle.get(t, func() string {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		cohere := filepath.Join(root, "cohere")
		// Until GoBuild lands, Product hashes the full pinned Go dependency tree,
		// overlay sources, workspace files, flags and Go toolchain environment.
		files := printerInputFiles(t, cohere, "testdata/cohere_side_test.go", "../testdata/cohere_side_test.go", "grain_asgo_test.go", root+"/go.mod", root+"/go.work")
		for _, name := range []string{"go.sum", "go.work.sum"} {
			if _, err := os.Stat(root + "/" + name); err == nil {
				files = append(files, root+"/"+name)
			} else if !os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
		return printerBuild(t, printerBuildInputs{
			Name: "GraphQL printer Go oracle", Files: files,
			Flags:     []string{"go test -c -trimpath -ldflags=-buildid= -overlay ./internal/format/graphql"},
			Toolchain: buildcache.Tool("go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "GOEXPERIMENT", "GOFLAGS", "CGO_ENABLED", "GOTOOLCHAIN", "GOAMD64", "GOARM64", "GOARM", "GO386", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM"),
		}, func(dir string) error {
			side, _ := filepath.Abs("testdata/cohere_side_test.go")
			generator, _ := filepath.Abs("../testdata/cohere_side_test.go")
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
				cohere + "/internal/format/graphql/adamic_printer_test.go":   side,
				cohere + "/internal/format/graphql/adamic_generator_test.go": generator,
			}})
			if err != nil {
				return err
			}
			path := filepath.Join(t.TempDir(), "overlay.json")
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			// Shared setup has no test-side deadline; the gate limits its unit.
			command := exec.CommandContext(context.Background(), "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-o="+dir+"/oracle", "-overlay="+path, "./internal/format/graphql")
			command.Dir = cohere
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("Go oracle: %w\n%s", err, output)
			}
			return nil
		}) + "/oracle"
	})
}

func printerAsGoLowered(t *testing.T) string {
	t.Helper()
	return asGoLowered.get(t, func() string { return printerLoweredProduct(t, printerAsGoSource(t)) })
}
func printerAsGoNative(t *testing.T, sanitize bool) string {
	t.Helper()
	once := &asGoRelease
	if sanitize {
		once = &asGoSanitized
	}
	return once.get(t, func() string {
		return printerCompiledProduct(t, printerAsGoLowered(t), native.Options{Sanitize: sanitize})
	})
}
func printerAsGoProducts(t *testing.T) printerProducts {
	t.Helper()
	product := printerProducts{source: printerAsGoSource(t), backend: printerAsGoLowered(t) + "/program.mjs", sanitized: printerAsGoNative(t, true)}
	if runtime.GOOS == "darwin" {
		product.release = printerAsGoNative(t, false)
	}
	return product
}
func printerAsGoCases(t *testing.T, mode string) (string, string) {
	t.Helper()
	for i, candidate := range asGoModes {
		if candidate != mode {
			continue
		}
		corpus := &asGoCorpus[i]
		corpus.once.Do(func() { corpus.path, corpus.want = printerCases(t, mode, printerAsGoOracle(t)) })
		if corpus.path == "" {
			t.Fatal("GraphQL corpus preparation failed")
		}
		return corpus.path, corpus.want
	}
	t.Fatalf("unknown mode %q", mode)
	return "", ""
}

func TestProduct_GraphQLPrinterSource(t *testing.T)    { t.Parallel(); printerAsGoSource(t) }
func TestProduct_GraphQLPrinterOracle(t *testing.T)    { t.Parallel(); printerAsGoOracle(t) }
func TestProduct_GraphQLPrinterLowered(t *testing.T)   { t.Parallel(); printerAsGoLowered(t) }
func TestProduct_GraphQLPrinterSanitized(t *testing.T) { t.Parallel(); printerAsGoNative(t, true) }
func TestProduct_GraphQLPrinterRelease(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin leak check only")
	}
	printerAsGoNative(t, false)
}
func TestProduct_GraphQLPrinterDefaults(t *testing.T) { t.Parallel(); printerAsGoCases(t, "defaults") }
func TestProduct_GraphQLPrinterNarrow(t *testing.T)   { t.Parallel(); printerAsGoCases(t, "narrow") }
func TestProduct_GraphQLPrinterTight(t *testing.T)    { t.Parallel(); printerAsGoCases(t, "tight") }
func TestProduct_GraphQLPrinterTabs(t *testing.T)     { t.Parallel(); printerAsGoCases(t, "tabs") }

func TestPrinterAsGoCohere_Union(t *testing.T) { t.Parallel(); printerAsGoUnit(t, -1) }
func TestPrinterAsGoCohere_000(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 0) }
func TestPrinterAsGoCohere_001(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 1) }
func TestPrinterAsGoCohere_002(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 2) }
func TestPrinterAsGoCohere_003(t *testing.T)   { t.Parallel(); printerAsGoUnit(t, 3) }

type printerAsGoJob struct {
	name, command     string
	args, environment []string
	leakReport        bool
}
type printerAsGoResult struct {
	job printerAsGoJob
	run run
	err error
}

func printerAsGoRun(ctx context.Context, job printerAsGoJob) printerAsGoResult {
	command := exec.CommandContext(ctx, job.command, job.args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	if job.environment != nil {
		command.Env = append(os.Environ(), job.environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		err = nil
	}
	result := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: -1}
	if command.ProcessState != nil {
		result.exitCode = command.ProcessState.ExitCode()
	}
	return printerAsGoResult{job, result, err}
}

const testPrinterAsGoCohereShards = 4

// The four fixed option modes own every case in their mode. ADAMIC_TEST_SHARD=i/n
// selects shard indices modulo n equal to i; unset runs all. The gate can instead
// select TestPrinterAsGoCohere_NNN directly. Products are prepared before case timing.
func printerAsGoUnit(t *testing.T, unit int) {
	setupStart := time.Now()
	var products printerProducts
	if unit >= 0 {
		products = printerAsGoProducts(t)
	}
	var whole []printerCase
	var shards []printerShard
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		cases, want := printerAsGoCases(t, mode)
		enumeration := enumeratePrinter(t, mode, cases, want)
		whole = append(whole, enumeration...)
		shards = append(shards, printerShard{mode: mode, path: cases, cases: enumeration})
	}
	if len(shards) != testPrinterAsGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		return
	}
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			t.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			t.Cleanup(func() {
				t.Logf("own work wall %.3fs", time.Since(start).Seconds())
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mode %s, cases 0..%d", shard.mode, len(shard.cases)-1)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"--cases", shard.path, shard.mode}
			jobs := []printerAsGoJob{
				{"Node", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, products.source}, args...), nil, false},
				{"native", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=0"}, false},
				{"JS backend", "node", append([]string{"--disable-warning=ExperimentalWarning", runner, products.backend}, args...), nil, false},
			}
			switch runtime.GOOS {
			case "linux":
				jobs = append(jobs, printerAsGoJob{"leaks", products.sanitized, args, []string{"ASAN_OPTIONS=detect_leaks=1"}, false})
			case "darwin":
				jobs = append(jobs, printerAsGoJob{"leaks", "leaks", append([]string{"--atExit", "--", products.release}, args...), nil, true})
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
			// All four processes read the same immutable corpus and products.
			// Use the instance's four CPUs without dropping a backend or leak check.
			results := make(chan printerAsGoResult, len(jobs))
			for _, job := range jobs {
				go func(job printerAsGoJob) { results <- printerAsGoRun(ctx, job) }(job)
			}
			for range jobs {
				result := <-results
				if result.err != nil {
					t.Errorf("%s: %v", result.job.name, result.err)
					continue
				}
				if result.job.leakReport {
					if result.run.exitCode != 0 {
						t.Errorf("leaks: exit %d stdout %s stderr %s", result.run.exitCode, result.run.stdout, result.run.stderr)
					}
				} else if err := printerShardDisagreement(number, shard, result.run); err != nil {
					t.Errorf("%s: %v", result.job.name, err)
				}
			}
		}
	}
}

func TestPrinterShardPlantedDisagreement(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("main.ts")
	var shards []printerShard
	for _, mode := range []string{"defaults", "tabs"} {
		cases, want := printerAsGoCases(t, mode)
		enumeration := enumeratePrinter(t, mode, cases, want)
		// Use the first two enumerated oracle cases as a real Node control.
		sample := append([]printerCase(nil), enumeration[:2]...)
		filename := filepath.Join(t.TempDir(), "cases.txt")
		if err := os.WriteFile(filename, []byte(sample[0].input+"\n"+sample[1].input+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		shards = append(shards, printerShard{mode: mode, path: filename, cases: sample})
	}
	results := make([]run, len(shards))
	for i, shard := range shards {
		results[i] = onNode(t, path, "--cases", shard.path, shard.mode)
		if err := printerShardDisagreement(i, shard, results[i]); err != nil {
			t.Fatal(err)
		}
	}
	shards[0].cases[1].want += "planted disagreement"
	caught := 0
	for i, shard := range shards {
		if err := printerShardDisagreement(i, shard, results[i]); err != nil {
			caught++
			if i != 0 || !strings.Contains(err.Error(), "shard-000 defaults/case-000001") {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}
