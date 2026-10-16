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

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
)

var tscAgreementShared struct {
	once      sync.Once
	directory string
}
var tscAgreementFamilyProducts [2]struct {
	once     sync.Once
	products tsPrinterProducts
}

// The corpus product's key moves with each corpus file and the recipe, and with nothing else the tree holds: a stand-in
// tree carries every declared input and one undeclared file. Mutants: the corpus directory or notyet.json dropped from
// Files, and an edit there no longer moves the key. The recipe is the list this test reads, and its code is compiled
// in rather than read as the build runs, so no test or trace holds it: it is a judgment of which code shapes the bytes.
func TestTSCAgreementKeyIsItsCorpusAndRecipe(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	corpus := "stage3/drivers/tsc/corpus/case.ts"
	files := append([]string{corpus, "stage1/cohere/tsprinter/testdata/notyet.json"}, oracles.TSCAgreementRecipe...)
	for _, name := range files {
		write(name, name)
	}
	write("internal/unrelated.go", "package unrelated")
	key := func(oracle string) string {
		t.Helper()
		key, err := buildcache.Key(root, oracles.TSCAgreementKeyInputs(oracle))
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	before := key("oracle")
	write("internal/unrelated.go", "package unrelated // edited")
	if key("oracle") != before {
		t.Fatal("an edit to a file the build never reads moved the key")
	}
	if key("another oracle") == before {
		t.Fatal("another oracle kept the key")
	}
	for _, name := range files {
		write(name, name+"!")
		if key("oracle") == before {
			t.Errorf("a one-byte edit to %s kept the key", name)
		}
		write(name, name)
	}
}

func tscAgreementLoad(t *testing.T) oracles.TSCAgreementData {
	t.Helper()
	tscAgreementShared.once.Do(func() { tscAgreementShared.directory = oracles.TSCAgreementPrepare(t) })
	directory := tscAgreementShared.directory
	if directory == "" {
		t.Fatal("shared corpus preparation failed")
	}

	encoded, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data oracles.TSCAgreementData
	if err := json.Unmarshal(encoded, &data); err != nil {
		t.Fatal(err)
	}
	for i := range data.Families {
		data.Families[i].Directory = filepath.Join(directory, data.Families[i].Directory)
		for j := range data.Families[i].Cases {
			data.Families[i].Cases[j].Label = buildcache.Absolute(data.Families[i].Cases[j].Label)
		}
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
	return int(hash.Sum64() % uint64(oracles.TSCCorpusAgreementShards/2))
}

func tscAgreementPartition(t *testing.T, family string, cases []oracles.PrinterCase, offset int) ([]oracles.CorpusShard, [][]byte, string) {
	t.Helper()
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	count := oracles.TSCCorpusAgreementShards / 2
	shards, answers := make([]oracles.CorpusShard, count), make([][]byte, count)
	inputs := make([]strings.Builder, count)
	var whole strings.Builder
	for i, item := range cases {
		// Labels include repository paths: use the repository-relative stable key.
		label := item.Label
		if at := strings.Index(label, "stage3/drivers/tsc/corpus/"); at >= 0 {
			label = label[at:]
		}
		owner := tscAgreementOwner(label)
		shards[owner].Indices = append(shards[owner].Indices, offset+i)
		inputs[owner].WriteString(">" + escape.Replace(item.Source) + "\n")
		line := "ok\t" + escape.Replace(item.Want) + "\n"
		answers[owner] = append(answers[owner], []byte(line)...)
		whole.WriteString(line)
	}
	for i := range shards {
		shards[i].Text = filepath.Join(t.TempDir(), fmt.Sprintf("%s-%03d.txt", family, i))
		if err := os.WriteFile(shards[i].Text, []byte(inputs[i].String()), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return shards, answers, whole.String()
}

func tscAgreementPlan(t *testing.T, data oracles.TSCAgreementData) (*oracles.CorpusPlan, [][]byte) {
	t.Helper()
	plan := &oracles.CorpusPlan{}
	var answers [][]byte
	var whole strings.Builder
	for _, family := range data.Families {
		shards, outputs, want := tscAgreementPartition(t, family.Name, family.Cases, len(plan.Labels))
		for i, item := range family.Cases {
			plan.Labels = append(plan.Labels, fmt.Sprintf("%s/case-%06d %s", family.Name, i, item.Label))
		}
		plan.Shards = append(plan.Shards, shards...)
		answers = append(answers, outputs...)
		whole.WriteString(want)
	}
	if len(plan.Shards) != oracles.TSCCorpusAgreementShards {
		t.Fatalf("enumerated %d shards, declared %d", len(plan.Shards), oracles.TSCCorpusAgreementShards)
	}
	if err := oracles.ExpressionUnion(plan, answers, []byte(whole.String())); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique family/case ids across %d shards", len(plan.Labels), len(plan.Shards))
	return plan, answers
}

func tscAgreementRun(t *testing.T, number int) {
	t.Helper()
	setupStart := time.Now()
	data := tscAgreementLoad(t)
	plan, answers := tscAgreementPlan(t, data)
	if !expressionSelection(t, oracles.TSCCorpusAgreementShards)[number] {
		t.Skip("assigned to another box")
	}
	owner := number / (oracles.TSCCorpusAgreementShards / 2)
	family, shard := data.Families[owner], plan.Shards[number]
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
	node := func(path string, arguments ...string) oracles.Run {
		runner, err := filepath.Abs(filepath.Join(oracles.Repository, "oracle/node.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		return tscAgreementExecute(t, ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
	}
	t.Logf("family %s: %d fragments", family.Name, len(shard.Indices))
	args := []string{"--cases", shard.Text, "80"}
	var report []map[string]string
	check := func(name string, result oracles.Run) {
		if err := expressionDisagreement(plan, number, answers[number], result); err != nil {
			t.Errorf("%s: %v", name, err)
		}
		lines := strings.Split(strings.TrimSuffix(string(result.Stdout), "\n"), "\n")
		expected := strings.Split(strings.TrimSuffix(string(answers[number]), "\n"), "\n")
		for index, id := range shard.Indices {
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
		if result.ExitCode != 0 {
			t.Errorf("leaks: exit %d stdout %s stderr %s", result.ExitCode, result.Stdout, result.Stderr)
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
	for len(plan.Shards[number].Indices) == 0 {
		number++
	}
	results := make([]oracles.Run, len(answers))
	for i := range answers {
		family := data.Families[i/(oracles.TSCCorpusAgreementShards/2)]
		results[i] = onNode(t, family.Source, "--cases", plan.Shards[i].Text, "80")
		if err := expressionDisagreement(plan, i, answers[i], results[i]); err != nil {
			t.Fatal(err)
		}
	}
	results[number].Stdout[0] = 'X'
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

// Use the existing printer helper's complete inputs for the lowered product.
func tscAgreementProducts(t *testing.T, path, family string) tsPrinterProducts {
	t.Helper()
	return printerGateFamilyProducts(t, family)
}

// Keep the same process status, stderr and byte oracles under a case-only deadline.
func tscAgreementExecute(t *testing.T, ctx context.Context, environment []string, name string, arguments ...string) oracles.Run {
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
	return oracles.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}
