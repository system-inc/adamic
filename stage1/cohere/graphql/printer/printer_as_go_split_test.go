package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

const testPrinterAsGoCohereShards = 8

type asGoCase struct{ ID, Input, Want string }
type asGoPrepared struct {
	Source, Backend, Sanitized, Release string
	Cases                               [][]asGoCase
	Modes                               []string
}

// The pinned corpus uses two contiguous ranges for each of its four modes.
// Only Setup builds; separately invoked leaves require its immutable product.
func asGoProduct(t *testing.T, setup bool) (string, asGoPrepared) {
	t.Helper()
	asGoTests := [...]func(*testing.T){TestPrinterAsGoCohere_000, TestPrinterAsGoCohere_001, TestPrinterAsGoCohere_002, TestPrinterAsGoCohere_003, TestPrinterAsGoCohere_004, TestPrinterAsGoCohere_005, TestPrinterAsGoCohere_006, TestPrinterAsGoCohere_007}
	if len(asGoTests) != testPrinterAsGoCohereShards {
		t.Fatal("top-level enumeration differs from declared shard count")
	}
	inputs := buildcache.Inputs{Name: "TestPrinterAsGoCohere setup v1", Files: []string{"internal", "stage1/cohere/graphql", "stage1/cohere/json", "oracle", "go.mod", "go.sum", "cohere/internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/mutation_aliasing", "cohere/static_single_assignment", "cohere/go.mod", "cohere/go.sum"}, Flags: []string{"ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		if !setup {
			return fmt.Errorf("run TestPrinterAsGoCohere_Setup first; shards never build")
		}
		oracle := printerOracle(t)
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
	total := 0
	for _, cases := range prepared.Cases {
		total += len(cases)
	}
	t.Logf("union: %d mode/case ids across %d shards", total, len(prepared.Cases))
	return directory, prepared
}

// Not parallel: publishes shared build products before parallel leaves start.
func TestPrinterAsGoCohere_Setup(t *testing.T) { asGoProduct(t, true) }

func asGoShard(t *testing.T, number int) {
	t.Helper()
	directory, prepared := asGoProduct(t, false)
	shard := asGoEnumeration(directory, prepared, number)
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
	directory, prepared := asGoProduct(t, false)
	caught := 0
	for number := range prepared.Cases {
		shard := asGoEnumeration(directory, prepared, number)
		var output strings.Builder
		for _, item := range shard.cases {
			output.WriteString(item.want)
			output.WriteByte('\n')
		}
		if number == 0 {
			shard.cases[0].want += "planted disagreement"
		}
		if err := printerShardDisagreement(number, shard, run{stdout: []byte(output.String())}); err != nil {
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
