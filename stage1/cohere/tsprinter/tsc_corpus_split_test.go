package tsprinter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

const testTSCCorpusAgreementShards = 8

type tscAgreementFamily struct {
	Name, Directory                     string
	Cases                               []printerCase
	Source, Backend, Sanitized, Release string
}
type tscAgreementData struct{ Families []tscAgreementFamily }

var tscAgreementShared struct {
	once      sync.Once
	directory string
}
var tscAgreementFamilyProducts [2]struct {
	once     sync.Once
	products tsPrinterProducts
}

func tscAgreementInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "TSC agreement corpus v3", Flags: []string{fmt.Sprint(testTSCCorpusAgreementShards), "GOFLAGS=" + os.Getenv("GOFLAGS"), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT")}, Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	// Match the existing printer product helper's complete repository inputs,
	// including cohere's rule runner, shims and all embedded compiler resources.
	for _, file := range expressionInputFiles(t, root) {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
	}
	return inputs
}

// Every independently selected leaf can prepare the corpus; no test ordering is required.
func tscAgreementPrepare(t *testing.T) string {
	t.Helper()
	return buildcache.Product(t, tscAgreementInputs(t), func(directory string) error {
		root, err := filepath.Abs(repository)
		if err != nil {
			return err
		}
		files, err := trackedRootFiles(t, root, "stage3/drivers/tsc/corpus")
		if err != nil {
			return err
		}
		oracle := tscOracle(t, root)
		data := tscAgreementData{}
		for _, statements := range []bool{false, true} {
			family, entry := "expressions", "main.ts"
			if statements {
				family, entry = "statements", "statementsMain.ts"
			}
			cases, original := tscCorpus(t, root, files, family, statements, oracle)
			output := filepath.Join(directory, family)
			if err := os.MkdirAll(output, 0755); err != nil {
				return err
			}
			for _, name := range []string{"coverage.json"} {
				content, err := os.ReadFile(filepath.Join(original, name))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(output, name), content, 0644); err != nil {
					return err
				}
			}
			data.Families = append(data.Families, tscAgreementFamily{Name: family, Directory: family, Cases: cases, Source: entry})
		}
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "manifest.json"), encoded, 0644)
	})
}

func tscAgreementLoad(t *testing.T) tscAgreementData {
	t.Helper()
	tscAgreementShared.once.Do(func() { tscAgreementShared.directory = tscAgreementPrepare(t) })
	directory := tscAgreementShared.directory
	if directory == "" {
		t.Fatal("shared corpus preparation failed")
	}

	encoded, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data tscAgreementData
	if err := json.Unmarshal(encoded, &data); err != nil {
		t.Fatal(err)
	}
	for i := range data.Families {
		data.Families[i].Directory = filepath.Join(directory, data.Families[i].Directory)
		source, err := filepath.Abs(data.Families[i].Source)
		if err != nil {
			t.Fatal(err)
		}
		data.Families[i].Source = source
	}
	return data
}

func tscAgreementOwner(label string) int {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(label))
	return int(hash.Sum64() % uint64(testTSCCorpusAgreementShards/2))
}

func tscAgreementPartition(t *testing.T, family string, cases []printerCase, offset int) ([]corpusShard, [][]byte, string) {
	t.Helper()
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	count := testTSCCorpusAgreementShards / 2
	shards, answers := make([]corpusShard, count), make([][]byte, count)
	inputs := make([]strings.Builder, count)
	var whole strings.Builder
	for i, item := range cases {
		// Labels include repository paths: use the repository-relative stable key.
		label := item.Label
		if at := strings.Index(label, "stage3/drivers/tsc/corpus/"); at >= 0 {
			label = label[at:]
		}
		owner := tscAgreementOwner(label)
		shards[owner].indices = append(shards[owner].indices, offset+i)
		inputs[owner].WriteString(">" + escape.Replace(item.Source) + "\n")
		line := "ok\t" + escape.Replace(item.Want) + "\n"
		answers[owner] = append(answers[owner], []byte(line)...)
		whole.WriteString(line)
	}
	for i := range shards {
		shards[i].text = filepath.Join(t.TempDir(), fmt.Sprintf("%s-%03d.txt", family, i))
		if err := os.WriteFile(shards[i].text, []byte(inputs[i].String()), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return shards, answers, whole.String()
}

func tscAgreementPlan(t *testing.T, data tscAgreementData) (*corpusPlan, [][]byte) {
	t.Helper()
	plan := &corpusPlan{}
	var answers [][]byte
	var whole strings.Builder
	for _, family := range data.Families {
		shards, outputs, want := tscAgreementPartition(t, family.Name, family.Cases, len(plan.labels))
		for i, item := range family.Cases {
			plan.labels = append(plan.labels, fmt.Sprintf("%s/case-%06d %s", family.Name, i, item.Label))
		}
		plan.shards = append(plan.shards, shards...)
		answers = append(answers, outputs...)
		whole.WriteString(want)
	}
	if len(plan.shards) != testTSCCorpusAgreementShards {
		t.Fatalf("enumerated %d shards, declared %d", len(plan.shards), testTSCCorpusAgreementShards)
	}
	if err := expressionUnion(plan, answers, []byte(whole.String())); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique family/case ids across %d shards", len(plan.labels), len(plan.shards))
	return plan, answers
}

func tscAgreementRun(t *testing.T, number int) {
	t.Helper()
	setupStart := time.Now()
	data := tscAgreementLoad(t)
	plan, answers := tscAgreementPlan(t, data)
	if !expressionSelection(t, testTSCCorpusAgreementShards)[number] {
		t.Skip("assigned to another box")
	}
	owner := number / (testTSCCorpusAgreementShards / 2)
	family, shard := data.Families[owner], plan.shards[number]
	tscAgreementFamilyProducts[owner].once.Do(func() {
		tscAgreementFamilyProducts[owner].products = tscAgreementProducts(t, family.Source, family.Name)
	})
	products := tscAgreementFamilyProducts[owner].products
	if products.sanitized == "" {
		t.Fatal("shared printer preparation failed")
	}
	t.Logf("TestTSCCorpusAgreement (setup): %.3fs", time.Since(setupStart).Seconds())
	// The case-work deadline starts after all shared products are ready. The
	// invocation's go test -timeout 90s still bounds setup and cases together.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("case work: %.3fs", time.Since(started).Seconds()) }()
	node := func(path string, arguments ...string) run {
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		return tscAgreementExecute(t, ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
	}
	t.Logf("family %s: %d fragments", family.Name, len(shard.indices))
	args := []string{"--cases", shard.text, "80"}
	var report []map[string]string
	check := func(name string, result run) {
		if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		lines := strings.Split(strings.TrimSuffix(string(result.stdout), "\n"), "\n")
		expected := strings.Split(strings.TrimSuffix(string(answers[number]), "\n"), "\n")
		for index, id := range shard.indices {
			got := ""
			if index < len(lines) {
				got = lines[index]
			}
			if got != expected[index] {
				local := id
				if owner == 1 {
					local -= len(data.Families[0].Cases)
				}
				item := family.Cases[local]
				report = append(report, map[string]string{"build": name, "file": item.Label, "source": item.Source, "port": got, "go": expected[index]})
			}
		}
	}
	check("Node", node(products.source, args...))
	check("native", tscAgreementExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, args...))
	check("backend", node(products.backend, args...))
	switch runtime.GOOS {
	case "linux":
		if err := expressionDisagreement(plan, number, answers[number], tscAgreementExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, args...)); err != nil {
			t.Errorf("leaks: %v", err)
		}
	case "darwin":
		result := tscAgreementExecute(t, ctx, nil, "leaks", append([]string{"--atExit", "--", products.release}, args...)...)
		if result.exitCode != 0 {
			t.Errorf("leaks: exit %d stdout %s stderr %s", result.exitCode, result.stdout, result.stderr)
		}
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
	if keep := os.Getenv("ADAMIC_TSC_PRINTER_AUDIT"); keep != "" {
		if err := os.MkdirAll(keep, 0755); err != nil {
			t.Fatal(err)
		}
		cases, err := json.Marshal(family.Cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(keep, family.Name+"-cases.json"), cases, 0644); err != nil {
			t.Fatal(err)
		}
		coverage, err := os.ReadFile(filepath.Join(family.Directory, "coverage.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(keep, family.Name+"-coverage.json"), coverage, 0644); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(keep, fmt.Sprintf("%s-%03d.json", family.Name, number)), encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTSCCorpusAgreement_Union(t *testing.T) {
	t.Parallel()
	tscAgreementPlan(t, tscAgreementLoad(t))
}

func TestTSCCorpusAgreement_PlantedDisagreement(t *testing.T) {
	t.Parallel()
	data := tscAgreementLoad(t)
	plan, answers := tscAgreementPlan(t, data)
	number := 0
	for len(plan.shards[number].indices) == 0 {
		number++
	}
	results := make([]run, len(answers))
	for i := range answers {
		family := data.Families[i/(testTSCCorpusAgreementShards/2)]
		results[i] = onNode(t, family.Source, "--cases", plan.shards[i].text, "80")
		if err := expressionDisagreement(plan, i, answers[i], results[i]); err != nil {
			t.Fatal(err)
		}
	}
	results[number].stdout[0] = 'X'
	caught := 0
	for i, result := range results {
		if err := expressionDisagreement(plan, i, answers[i], result); err != nil {
			caught++
			if i != number {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught by TestTSCCorpusAgreement_%03d: %v", i, err)
		}
	}
	if caught != 1 {
		t.Fatalf("caught by %d shards, want exactly one", caught)
	}
}

func TestTSCCorpusAgreement_000(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 0)
}

func TestTSCCorpusAgreement_001(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 1)
}

func TestTSCCorpusAgreement_002(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 2)
}

func TestTSCCorpusAgreement_003(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 3)
}

func TestTSCCorpusAgreement_004(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 4)
}

func TestTSCCorpusAgreement_005(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 5)
}

func TestTSCCorpusAgreement_006(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 6)
}

func TestTSCCorpusAgreement_007(t *testing.T) {
	t.Parallel()
	tscAgreementRun(t, 7)
}

// Use the existing printer helper's complete inputs for the lowered product.
func tscAgreementProducts(t *testing.T, path, family string) tsPrinterProducts {
	t.Helper()
	loweredProduct := expressionBuild(t, printerBuildInputs{Name: "lowered TSC " + family, Files: expressionInputFiles(t, repository), Toolchain: runtime.Version()}, func(dir string) error {
		program := lowered(t, path)
		if err := os.WriteFile(dir+"/port.c", []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
	data, err := os.ReadFile(loweredProduct + "/port.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	clang := executeOne(t, nil, "clang", "--version")
	if clang.exitCode != 0 {
		t.Fatalf("clang version: %s", clang.stderr)
	}
	inputs := append([]string{loweredProduct + "/port.c"}, expressionInputFiles(t, filepath.Join(repository, "internal/native"))...)
	sanitized := expressionBuild(t, printerBuildInputs{Name: "sanitized split TSC " + family, Files: inputs, Flags: append(native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}), "split=true", "jobs=4"), Toolchain: string(clang.stdout)}, func(dir string) error {
		return native.Build(source, dir+"/port", native.Options{Sanitize: true, Split: true, Jobs: 4})
	}) + "/port"
	release := ""
	if runtime.GOOS == "darwin" {
		release = expressionBuild(t, printerBuildInputs{Name: "release TSC " + family, Files: inputs, Flags: native.Flags(native.Options{}), Toolchain: string(clang.stdout)}, func(dir string) error {
			return native.Build(source, dir+"/port", native.Options{})
		}) + "/port"
	}
	return tsPrinterProducts{source: path, backend: loweredProduct + "/program.mjs", sanitized: sanitized, release: release}
}

// Keep the same process status, stderr and byte oracles under a case-only deadline.
func tscAgreementExecute(t *testing.T, ctx context.Context, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := exec.CommandContext(ctx, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := childguard.Run(command, childguard.Options{})
	if ctx.Err() != nil {
		t.Fatalf("case deadline: %v", ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}
