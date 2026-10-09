package printer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testPrinterAsGoCohereShards = 8

type asGoCase struct{ ID, Input, Want string }
type asGoPrepared struct {
	Source, Backend, Sanitized, Release string
	Cases                               [][]asGoCase
	Whole                               []asGoCase
	Modes                               []string
}

// The pinned corpus uses two contiguous ranges for each of its four modes.
// Every leaf can prepare the shared immutable products when selected alone.
var asGoShared struct {
	once      sync.Once
	directory string
	prepared  asGoPrepared
}

func asGoProduct(t *testing.T) (string, asGoPrepared) {
	t.Helper()
	asGoShared.once.Do(func() { asGoShared.directory, asGoShared.prepared = asGoPrepare(t) })
	if asGoShared.directory == "" {
		t.Fatal("shared product preparation failed")
	}
	return asGoShared.directory, asGoShared.prepared
}

func asGoPrepare(t *testing.T) (string, asGoPrepared) {
	t.Helper()
	asGoTests := [...]func(*testing.T){TestPrinterAsGoCohere_000, TestPrinterAsGoCohere_001, TestPrinterAsGoCohere_002, TestPrinterAsGoCohere_003, TestPrinterAsGoCohere_004, TestPrinterAsGoCohere_005, TestPrinterAsGoCohere_006, TestPrinterAsGoCohere_007}
	if len(asGoTests) != testPrinterAsGoCohereShards {
		t.Fatal("top-level enumeration differs from declared shard count")
	}
	inputs := buildcache.Inputs{Name: "TestPrinterAsGoCohere setup v2", Files: []string{"internal", "stage1/cohere/graphql", "stage1/cohere/json", "oracle", "go.mod", "go.work", "cohere/internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/mutation_aliasing", "cohere/static_single_assignment", "cohere/go.mod", "cohere/go.sum"}, Flags: []string{"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		oracle := asGoOracle(t)
		path, err := filepath.Abs("main.ts")
		if err != nil {
			return err
		}
		products := preparePrinterProducts(t, path)
		prepared := asGoPrepared{Source: products.source, Backend: products.backend, Sanitized: products.sanitized, Release: products.release}
		var whole []printerCase
		var shards []printerShard
		for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
			cases, want := printerCases(t, mode, oracle)
			enumeration := enumeratePrinter(t, mode, cases, want)
			whole = append(whole, enumeration...)
			for _, item := range enumeration {
				prepared.Whole = append(prepared.Whole, asGoCase{item.id, item.input, item.want})
			}
			for part := 0; part < 2; part++ {
				sample := enumeration[len(enumeration)*part/2 : len(enumeration)*(part+1)/2]
				number := len(shards)
				name := filepath.Join(directory, fmt.Sprintf("cases-%03d.txt", number))
				var data strings.Builder
				var saved []asGoCase
				for _, item := range sample {
					data.WriteString(item.input)
					data.WriteByte('\n')
					saved = append(saved, asGoCase{item.id, item.input, item.want})
				}
				if err := os.WriteFile(name, []byte(data.String()), 0644); err != nil {
					return err
				}
				shards = append(shards, printerShard{mode: mode, cases: sample})
				prepared.Modes = append(prepared.Modes, mode)
				prepared.Cases = append(prepared.Cases, saved)
			}
		}
		if len(shards) != testPrinterAsGoCohereShards {
			return fmt.Errorf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
		}
		if err := printerShardUnion(whole, shards); err != nil {
			return err
		}
		t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
		data, err := json.Marshal(prepared)
		if err != nil {
			return err
		}
		return os.WriteFile(directory+"/prepared.json", data, 0644)
	})
	data, err := os.ReadFile(directory + "/prepared.json")
	if err != nil {
		t.Fatal(err)
	}
	var prepared asGoPrepared
	if err := json.Unmarshal(data, &prepared); err != nil {
		t.Fatal(err)
	}
	if len(prepared.Cases) != testPrinterAsGoCohereShards {
		t.Fatal("prepared shard enumeration changed")
	}
	var whole []printerCase
	for _, item := range prepared.Whole {
		whole = append(whole, printerCase{id: item.ID, input: item.Input, want: item.Want})
	}
	var shards []printerShard
	for number := range prepared.Cases {
		shards = append(shards, asGoEnumeration(directory, prepared, number))
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards (each exactly once)", len(whole), len(shards))
	return directory, prepared
}

func TestPrinterAsGoCohere_Setup_Oracle(t *testing.T) { t.Parallel(); asGoOracle(t) }

func TestPrinterAsGoCohere_Setup(t *testing.T) { t.Parallel(); asGoProduct(t) }

func asGoShard(t *testing.T, number int) {
	t.Helper()
	directory, prepared := asGoProduct(t)
	shard := asGoEnumeration(directory, prepared, number)
	start := time.Now()
	t.Cleanup(func() {
		if elapsed := time.Since(start); elapsed > 30*time.Second {
			t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
		}
	})
	t.Logf("shard-%03d mode %s: %d cases", number, shard.mode, len(shard.cases))
	check := func(name string, result run) {
		if err := printerShardDisagreement(number, shard, result); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	args := []string{"--cases", shard.path, shard.mode}
	check("Node", onNode(t, prepared.Source, args...))
	check("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, prepared.Sanitized, args...))
	check("JS backend", onNode(t, prepared.Backend, args...))
	switch runtime.GOOS {
	case "linux":
		check("leaks", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, prepared.Sanitized, args...))
	case "darwin":
		report := execute(t, nil, "leaks", append([]string{"--atExit", "--", prepared.Release}, args...)...)
		if report.exitCode != 0 {
			t.Errorf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
}
func asGoEnumeration(directory string, prepared asGoPrepared, number int) printerShard {
	shard := printerShard{mode: prepared.Modes[number], path: filepath.Join(directory, fmt.Sprintf("cases-%03d.txt", number))}
	for _, item := range prepared.Cases[number] {
		shard.cases = append(shard.cases, printerCase{id: item.ID, input: item.Input, want: item.Want})
	}
	return shard
}

func TestPrinterAsGoCohere_PlantedFailure(t *testing.T) {
	t.Parallel()
	directory, prepared := asGoProduct(t)
	caught := 0
	for number := range prepared.Cases {
		shard := asGoEnumeration(directory, prepared, number)
		result := onNode(t, prepared.Source, "--cases", shard.path, shard.mode)
		if err := printerShardDisagreement(number, shard, result); err != nil {
			t.Fatalf("unmodified control: %v", err)
		}
		if number == 0 {
			shard.cases[0].want += "planted disagreement"
		}
		if err := printerShardDisagreement(number, shard, result); err != nil {
			caught++
			if number != 0 || !strings.Contains(err.Error(), shard.cases[0].id) {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted failure caught by TestPrinterAsGoCohere_%03d: %v", number, err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught by %d shards", caught)
	}
}

func TestPrinterAsGoCohere_000(t *testing.T) {
	t.Parallel()
	asGoShard(t, 0)
}

func TestPrinterAsGoCohere_001(t *testing.T) {
	t.Parallel()
	asGoShard(t, 1)
}

func TestPrinterAsGoCohere_002(t *testing.T) {
	t.Parallel()
	asGoShard(t, 2)
}

func TestPrinterAsGoCohere_003(t *testing.T) {
	t.Parallel()
	asGoShard(t, 3)
}

func TestPrinterAsGoCohere_004(t *testing.T) {
	t.Parallel()
	asGoShard(t, 4)
}

func TestPrinterAsGoCohere_005(t *testing.T) {
	t.Parallel()
	asGoShard(t, 5)
}

func TestPrinterAsGoCohere_006(t *testing.T) {
	t.Parallel()
	asGoShard(t, 6)
}

func TestPrinterAsGoCohere_007(t *testing.T) {
	t.Parallel()
	asGoShard(t, 7)
}

// Until buildcache.GoBuild reaches main, preserve the overlay Go build command.
// The oracle is a separate Product, so its build is counted and reused too.
func asGoOracle(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	directory := buildcache.Product(t, buildcache.Inputs{
		Name:      "GraphQL agreement Go oracle",
		Files:     []string{"cohere", "go.mod", "go.work", "stage1/cohere/graphql/printer/testdata/cohere_side_test.go", "stage1/cohere/graphql/testdata/cohere_side_test.go"},
		Flags:     []string{"go test -c", "-trimpath", "-ldflags=-buildid=", "overlay=adamic_printer_test.go,adamic_generator_test.go"},
		Toolchain: []string{buildcache.Tool("go", "version"), runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}, func(directory string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
			cohere + "/internal/format/graphql/adamic_printer_test.go":   root + "/stage1/cohere/graphql/printer/testdata/cohere_side_test.go",
			cohere + "/internal/format/graphql/adamic_generator_test.go": root + "/stage1/cohere/graphql/testdata/cohere_side_test.go",
		}})
		if err != nil {
			return err
		}
		path := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-o="+directory+"/oracle", "-overlay="+path, "./internal/format/graphql")
		command.Dir = cohere
		if output, err := combinedOutput(command); err != nil {
			return fmt.Errorf("Go GraphQL printer oracle: %w\n%s", err, output)
		}
		return nil
	})
	return directory + "/oracle"
}

func TestPrinterAsGoCohere_Union(t *testing.T) {
	t.Parallel()
	asGoProduct(t)
}
