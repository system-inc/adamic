package estree

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Fixed shard counts have headroom for corpus growth, including empty shards.
// Assignment hashes a repository-relative case key, never its enumeration position.
// miscPlan validates the live union even when only some shards are selected.
func miscPlan(ids []string, count int) ([][]int, error) {
	if count < 1 || len(ids) == 0 {
		return nil, fmt.Errorf("%d cases cannot enumerate %d shards", len(ids), count)
	}
	shards := make([][]int, count)
	expected := make(map[string]bool, len(ids))
	for i, id := range ids {
		if id == "" || expected[id] {
			return nil, fmt.Errorf("empty or repeated case id %q", id)
		}
		expected[id] = true
		shard := miscShard(id, count)
		shards[shard] = append(shards[shard], i)
	}
	seen := make(map[string]bool, len(ids))
	total := 0
	for _, shard := range shards {
		for _, i := range shard {
			id := ids[i]
			if !expected[id] || seen[id] {
				return nil, fmt.Errorf("unexpected or repeated case %q", id)
			}
			seen[id] = true
			total++
		}
	}
	if len(shards) != count || total != len(ids) || len(seen) != len(expected) {
		return nil, fmt.Errorf("shard union differs: shards=%d cases=%d unique=%d", len(shards), total, len(seen))
	}
	for id := range expected {
		if !seen[id] {
			return nil, fmt.Errorf("missing case %q", id)
		}
	}
	return shards, nil
}

func miscIDs(prefix string, cases []string) []string {
	ids := make([]string, len(cases))
	for i, text := range cases {
		// A generated case keeps its identity even when another is inserted before it.
		ids[i] = fmt.Sprintf("%s:source-%x", prefix, sha256.Sum256([]byte(text)))
	}
	return ids
}

// ADAMIC_TEST_SHARD=i/n selects shard indexes congruent to i modulo n.
// Unset runs every shard. Each leaf retains every implementation for its cases.
func miscRunShards(t *testing.T, count int, ids []string, check func(*testing.T, int)) {
	t.Helper()
	shards, err := miscPlan(ids, count)
	if err != nil {
		t.Fatal(err)
	}
	selected, boxes := 0, 1
	if value := os.Getenv("ADAMIC_TEST_SHARD"); value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
		selected, err = strconv.Atoi(parts[0])
		if err != nil {
			t.Fatal(err)
		}
		boxes, err = strconv.Atoi(parts[1])
		if err != nil || boxes < 1 || boxes > count || selected < 0 || selected >= boxes {
			t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
		}
	}
	t.Logf("union: %d cases, %d unique ids, %d shards", len(ids), len(ids), len(shards))
	for shard, indexes := range shards {
		if shard%boxes != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", shard), func(t *testing.T) {
			t.Parallel()
			for _, i := range indexes {
				check(t, i)
			}
		})
	}
}

func miscCompare(want, got []byte, mutant bool) error {
	diff := firstDifference(want, got)
	if mutant {
		if diff == "" {
			return fmt.Errorf("mutant survived")
		}
		return nil
	}
	if diff != "" {
		return fmt.Errorf("%s", diff)
	}
	return nil
}

// Re-execute this tiny proof test, driving the same planner, parallel runner and
// checker as the real test. Exactly one planted case must fail exactly its leaf.
func miscPlantedProof(t *testing.T, count int, ids []string, check func(bool) error) {
	t.Helper()
	planted := len(ids) - 1
	marker := "ADAMIC_ESTREE_MISC_PROOF"
	if os.Getenv(marker) == t.Name() {
		miscRunShards(t, count, ids, func(t *testing.T, i int) {
			if err := check(i == planted); err != nil {
				t.Fatal(err)
			}
		})
		return
	}
	command := exec.Command(os.Args[0], "-test.run=^"+t.Name()+"$", "-test.v", "-test.timeout=20s")
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(value, marker+"=") {
			command.Env = append(command.Env, value)
		}
	}
	command.Env = append(command.Env, marker+"="+t.Name())
	output, err := command.CombinedOutput()
	leaf := fmt.Sprintf("%s/shard-%03d", t.Name(), miscShard(ids[planted], count))
	if err == nil || bytes.Count(output, []byte("--- FAIL: "+t.Name()+"/shard-")) != 1 || !bytes.Contains(output, []byte("--- FAIL: "+leaf+" ")) {
		t.Fatalf("planted failure must be caught only by %s: exit=%v\n%s", leaf, err, output)
	}
	t.Logf("planted case %s caught only by %s", ids[planted], leaf)
}

func miscOracle(t *testing.T) string {
	t.Helper()
	start := time.Now()
	path := goOracle(t)
	t.Logf("build Go oracle: %.6fs", time.Since(start).Seconds())
	return path
}

// Product directories are read-only after publication; all leaves share them.
func miscBuild(t *testing.T, path string) (string, string) {
	t.Helper()
	product := func(directory string) error {
		program, err := load.Load([]string{path})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := native.Build(native.C(lowered), filepath.Join(directory, "port"), native.Options{Sanitize: true}); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "port.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	}
	start := time.Now()
	flags := append([]string{"main=" + path}, native.Flags(native.Options{Sanitize: true})...)
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	directory := buildcache.Product(t, buildcache.Inputs{
		Name:      "estree-misc-sanitized-emitted",
		Files:     []string{"internal", "bridge", "go.mod", "stage1/typescript", "stage1/cohere/estree"},
		Flags:     flags,
		Toolchain: []string{runtime.Version(), miscDigest(t, os.Args[0]), buildcache.Tool("clang", "--version"), buildcache.Tool("clang", "-v"), miscArchiverTool(), buildcache.Tool("ld", "--version")},
	}, product)
	t.Logf("build lowered/sanitized/emitted products: %.6fs", time.Since(start).Seconds())
	return filepath.Join(directory, "port"), filepath.Join(directory, "port.mjs")
}

// CPU includes this test process and all completed child processes.
func miscCPU() float64 {
	var self, child syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &self); err != nil {
		panic(err)
	}
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &child); err != nil {
		panic(err)
	}
	return float64(self.Utime.Sec+self.Stime.Sec+child.Utime.Sec+child.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+child.Utime.Usec+child.Stime.Usec)/1e6
}

func miscStart(t *testing.T) func() {
	start, cpu := time.Now(), miscCPU()
	t.Cleanup(func() { t.Logf("total CPU: %.6fs", miscCPU()-cpu) })
	return func() { t.Logf("setup wall: %.6fs", time.Since(start).Seconds()) }
}

func TestMiscShardUnionRejectsInvalidEnumeration(t *testing.T) {
	for _, item := range []struct {
		ids   []string
		count int
	}{
		{[]string{"a", "a"}, 2}, {[]string{"a", ""}, 2}, {nil, 2}, {[]string{"a"}, 0},
	} {
		if _, err := miscPlan(item.ids, item.count); err == nil {
			t.Fatalf("invalid enumeration accepted: %+v", item)
		}
	}
}

func miscDigest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// These are TypeScript source mutants, not Go overlay builds.
func miscMutant(t *testing.T, file, from, to string) string {
	t.Helper()
	repo := root(t)
	inputs := buildcache.Inputs{Name: "estree-misc-source-mutant", Files: []string{"stage1/cohere/estree", "stage1/typescript"}, Flags: []string{repo, file, from, to}, Toolchain: []string{runtime.Version()}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		files, err := filepath.Glob("*.ts")
		if err != nil {
			return err
		}
		found := false
		for _, name := range files {
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			text := string(data)
			if name == file {
				found = true
				if strings.Count(text, from) != 1 {
					return fmt.Errorf("mutant anchor %q occurs %d times", from, strings.Count(text, from))
				}
				text = strings.Replace(text, from, to, 1)
			}
			text = strings.ReplaceAll(text, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(repo, "stage1/typescript"))+"/")
			if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0644); err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("missing mutant source %s", file)
		}
		return nil
	})
	return filepath.Join(directory, "main.ts")
}

type miscAnswer struct {
	Status string
	Data   []byte
}

// Cache reference outputs by the already-built oracle's bytes, not a hand-listed
// Go build key. Audit paths are discarded; only the status and canonical answer
// are inputs to the checks. All input bodies and filename extensions are keyed.
func miscAnswers(t *testing.T, oracle string, cases []boundedPortCase) []miscAnswer {
	t.Helper()
	encoded, err := json.Marshal(casesToInputs(cases))
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "estree-misc-oracle-answers", Files: []string{"stage1/cohere/estree/misc_shards_test.go"}, Flags: []string{string(encoded)}, Toolchain: []string{runtime.Version(), miscDigest(t, oracle)}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		answers := make([]miscAnswer, len(cases))
		for i, item := range cases {
			subdir := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
			if err := os.Mkdir(subdir, 0755); err != nil {
				return err
			}
			path := filepath.Join(subdir, item.filename)
			if err := os.WriteFile(path, []byte(item.text), 0644); err != nil {
				return err
			}
			if item.audit {
				list := filepath.Join(subdir, "manifest")
				if err := os.WriteFile(list, []byte(path+"\n"), 0644); err != nil {
					return err
				}
				output, err := miscReferenceCommand(oracle, "--audit", list, subdir)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(output, &answers[i]); err != nil {
					return err
				}
			} else {
				answers[i].Status = "ok"
			}
			if answers[i].Status == "ok" {
				output, err := miscReferenceCommand(oracle, path)
				if err != nil {
					return err
				}
				answers[i].Data = output
			}
			if err := os.RemoveAll(subdir); err != nil {
				return err
			}
		}
		data, err := json.Marshal(answers)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "answers.json"), data, 0644)
	})
	data, err := os.ReadFile(filepath.Join(directory, "answers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var answers []miscAnswer
	if err := json.Unmarshal(data, &answers); err != nil {
		t.Fatal(err)
	}
	if len(answers) != len(cases) {
		t.Fatal("oracle answer enumeration differs")
	}
	return answers
}

func casesToInputs(cases []boundedPortCase) [][]string {
	result := make([][]string, len(cases))
	for i, item := range cases {
		result[i] = []string{item.id, item.text, item.filename, strconv.FormatBool(item.audit)}
	}
	return result
}

func miscTextAnswers(t *testing.T, oracle string, texts []string, extension string) []miscAnswer {
	cases := make([]boundedPortCase, len(texts))
	for i, text := range texts {
		cases[i] = boundedPortCase{id: fmt.Sprint(i), text: text, filename: "0000" + extension}
	}
	return miscAnswers(t, oracle, cases)
}

func miscReferenceCommand(oracle string, args ...string) ([]byte, error) {
	command := exec.Command(oracle, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil || stderr.Len() != 0 {
		return nil, fmt.Errorf("oracle %v: %v\n%s", args, err, &stderr)
	}
	return output, nil
}

// Match native's archiver selection, including its GNU ar fallback.
func miscArchiverTool() string {
	compiler, err := exec.LookPath("clang")
	if err == nil {
		candidate := filepath.Join(filepath.Dir(compiler), "llvm-ar")
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
			return buildcache.Tool(candidate, "--version")
		}
	}
	return buildcache.Tool("ar", "--version")
}

func miscShard(id string, count int) int {
	digest := sha256.Sum256([]byte(id))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(count))
}

func TestMiscShardGrowthKeepsAssignments(t *testing.T) {
	const shards = 32
	original := []string{"stage1/cohere/estree/z.input:0:audit", "stage1/cohere/estree/m.input:0:audit", "stage1/cohere/estree/exports_test.go:export:0"}
	original = append(original, miscIDs("stage1/cohere/estree/exports_test.go:export", []string{"@d export class C {}", "@d export default class C {}"})...)
	before, err := miscPlan(original, shards)
	if err != nil {
		t.Fatal(err)
	}
	assignments := make(map[string]int)
	for shard, indexes := range before {
		for _, index := range indexes {
			assignments[original[index]] = shard
		}
	}
	// Insert a file before the existing paths, reorder them, and add a case in an
	// existing file. None of these operations may move any pre-existing case.
	grown := []string{"stage1/cohere/estree/a.input:0:audit", original[2], "stage1/cohere/estree/m.input:1:audit", original[1], original[0]}
	generated := miscIDs("stage1/cohere/estree/exports_test.go:export", []string{"new case", "@d export default class C {}", "@d export class C {}"})
	grown = append(grown, generated...)
	after, err := miscPlan(grown, shards)
	if err != nil {
		t.Fatal(err)
	}
	seen := 0
	for shard, indexes := range after {
		for _, index := range indexes {
			if previous, exists := assignments[grown[index]]; exists {
				seen++
				if shard != previous {
					t.Fatalf("case %s moved from shard-%03d to shard-%03d", grown[index], previous, shard)
				}
			}
		}
	}
	if seen != len(original) {
		t.Fatal("growth lost an original case")
	}
}
