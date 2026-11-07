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
func TestTSCAgreementKeyIsItsCorpusAndRecipe(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	write := func(name, content string) {
		test.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	corpus := "stage3/drivers/tsc/corpus/case.ts"
	files := append([]string{corpus, "stage1/cohere/tsprinter/testdata/notyet.json"}, oracles.TSCAgreementRecipe...)
	for _, name := range files {
		write(name, name)
	}
	write("internal/unrelated.go", "package unrelated")
	key := func(oracle string) string {
		test.Helper()
		key, keyError := buildcache.Key(root, oracles.TSCAgreementKeyInputs(oracle))
		if keyError != nil {
			test.Fatal(keyError)
		}
		return key
	}
	before := key("oracle")
	write("internal/unrelated.go", "package unrelated // edited")
	if key("oracle") != before {
		test.Fatal("an edit to a file the build never reads moved the key")
	}
	if key("another oracle") == before {
		test.Fatal("another oracle kept the key")
	}
	for _, name := range files {
		write(name, name+"!")
		if key("oracle") == before {
			test.Errorf("a one-byte edit to %s kept the key", name)
		}
		write(name, name)
	}
}

func tscAgreementLoad(test *testing.T) oracles.TSCAgreementData {
	test.Helper()
	tscAgreementShared.once.Do(func() { tscAgreementShared.directory = oracles.TSCAgreementPrepare(test) })
	directory := tscAgreementShared.directory
	if directory == "" {
		test.Fatal("shared corpus preparation failed")
	}

	encoded, readError := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var data oracles.TSCAgreementData
	if decodeError := json.Unmarshal(encoded, &data); decodeError != nil {
		test.Fatal(decodeError)
	}
	for i := range data.Families {
		data.Families[i].Directory = filepath.Join(directory, data.Families[i].Directory)
		for j := range data.Families[i].Cases {
			data.Families[i].Cases[j].Label = buildcache.Absolute(data.Families[i].Cases[j].Label)
		}
		source, absolutePathError := filepath.Abs(data.Families[i].Source)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
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

func tscAgreementPartition(
	test *testing.T,
	family string,
	cases []oracles.PrinterCase,
	offset int,
) ([]oracles.CorpusShard, [][]byte, string) {
	test.Helper()
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
		shards[i].Text = filepath.Join(test.TempDir(), fmt.Sprintf("%s-%03d.txt", family, i))
		if writeError := os.WriteFile(shards[i].Text, []byte(inputs[i].String()), 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return shards, answers, whole.String()
}

func tscAgreementPlan(test *testing.T, data oracles.TSCAgreementData) (*oracles.CorpusPlan, [][]byte) {
	test.Helper()
	plan := &oracles.CorpusPlan{}
	var answers [][]byte
	var whole strings.Builder
	for _, family := range data.Families {
		shards, outputs, want := tscAgreementPartition(test, family.Name, family.Cases, len(plan.Labels))
		for i, item := range family.Cases {
			plan.Labels = append(plan.Labels, fmt.Sprintf("%s/case-%06d %s", family.Name, i, item.Label))
		}
		plan.Shards = append(plan.Shards, shards...)
		answers = append(answers, outputs...)
		whole.WriteString(want)
	}
	if len(plan.Shards) != oracles.TSCCorpusAgreementShards {
		test.Fatalf("enumerated %d shards, declared %d", len(plan.Shards), oracles.TSCCorpusAgreementShards)
	}
	if expressionUnionError := oracles.ExpressionUnion(
		plan,
		answers,
		[]byte(whole.String()),
	); expressionUnionError != nil {
		test.Fatal(expressionUnionError)
	}
	test.Logf("union: %d unique family/case ids across %d shards", len(plan.Labels), len(plan.Shards))
	return plan, answers
}

func tscAgreementRun(test *testing.T, number int) {
	test.Helper()
	setupStart := time.Now()
	data := tscAgreementLoad(test)
	plan, answers := tscAgreementPlan(test, data)
	if !expressionSelection(test, oracles.TSCCorpusAgreementShards)[number] {
		test.Skip("assigned to another box")
	}
	owner := number / (oracles.TSCCorpusAgreementShards / 2)
	family, shard := data.Families[owner], plan.Shards[number]
	tscAgreementFamilyProducts[owner].once.Do(func() {
		tscAgreementFamilyProducts[owner].products = tscAgreementProducts(test, family.Source, family.Name)
	})
	products := tscAgreementFamilyProducts[owner].products
	if products.sanitized == "" {
		test.Fatal("shared printer preparation failed")
	}
	test.Logf("TestTSCCorpusAgreement (setup): %.3fs", time.Since(setupStart).Seconds())
	// The case-work deadline starts after all shared products are ready. The
	// invocation's go test -timeout 90s still bounds setup and cases together.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { test.Logf("case work: %.3fs", time.Since(started).Seconds()) }()
	node := func(path string, arguments ...string) oracles.Run {
		runner, absolutePathError := filepath.Abs(filepath.Join(oracles.Repository, "oracle/node.mjs"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		return tscAgreementExecute(
			test,
			timeoutContext,
			nil,
			"node",
			append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
		)
	}
	test.Logf("family %s: %d fragments", family.Name, len(shard.Indices))
	args := []string{"--cases", shard.Text, "80"}
	var report []map[string]string
	check := func(name string, result oracles.Run) {
		if expressionDisagreementError := expressionDisagreement(
			plan,
			number,
			answers[number],
			result,
		); expressionDisagreementError != nil {
			test.Errorf("%s: %v", name, expressionDisagreementError)
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
				report = append(report, map[string]string{
					"build":  name,
					"file":   item.Label,
					"source": item.Source,
					"port":   got,
					"go":     expected[index],
				})
			}
		}
	}
	check("Node", node(products.source, args...))
	check(
		"native",
		tscAgreementExecute(test, timeoutContext, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, args...),
	)
	check("backend", node(products.backend, args...))
	switch runtime.GOOS {
	case "linux":
		if expressionDisagreementError := expressionDisagreement(
			plan,
			number,
			answers[number],
			tscAgreementExecute(
				test,
				timeoutContext,
				[]string{"ASAN_OPTIONS=detect_leaks=1"},
				products.sanitized,
				args...,
			),
		); expressionDisagreementError != nil {
			test.Errorf("leaks: %v", expressionDisagreementError)
		}
	case "darwin":
		result := tscAgreementExecute(
			test,
			timeoutContext,
			nil,
			"leaks",
			append([]string{"--atExit", "--", products.release}, args...)...,
		)
		if result.ExitCode != 0 {
			test.Errorf("leaks: exit %d stdout %s stderr %s", result.ExitCode, result.Stdout, result.Stderr)
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
	if keep := os.Getenv("ADAMIC_TSC_PRINTER_AUDIT"); keep != "" {
		if createDirectoryError := os.MkdirAll(keep, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		cases, encodeError := json.Marshal(family.Cases)
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(filepath.Join(keep, family.Name+"-cases.json"), cases, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		coverage, readError := os.ReadFile(filepath.Join(family.Directory, "coverage.json"))
		if readError != nil {
			test.Fatal(readError)
		}
		if writeError := os.WriteFile(
			filepath.Join(keep, family.Name+"-coverage.json"),
			coverage,
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
		encoded, encodeError := json.MarshalIndent(report, "", "  ")
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile(
			filepath.Join(keep, fmt.Sprintf("%s-%03d.json", family.Name, number)),
			encoded,
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
}

func TestTSCCorpusAgreement_Union(test *testing.T) {
	test.Parallel()
	tscAgreementPlan(test, tscAgreementLoad(test))
}

func TestTSCCorpusAgreement_PlantedDisagreement(test *testing.T) {
	test.Parallel()
	data := tscAgreementLoad(test)
	plan, answers := tscAgreementPlan(test, data)
	number := 0
	for len(plan.Shards[number].Indices) == 0 {
		number++
	}
	results := make([]oracles.Run, len(answers))
	for i := range answers {
		family := data.Families[i/(oracles.TSCCorpusAgreementShards/2)]
		results[i] = onNode(test, family.Source, "--cases", plan.Shards[i].Text, "80")
		if expressionDisagreementError := expressionDisagreement(
			plan,
			i,
			answers[i],
			results[i],
		); expressionDisagreementError != nil {
			test.Fatal(expressionDisagreementError)
		}
	}
	results[number].Stdout[0] = 'X'
	caught := 0
	for i, result := range results {
		if expressionDisagreementError := expressionDisagreement(
			plan,
			i,
			answers[i],
			result,
		); expressionDisagreementError != nil {
			caught++
			if i != number {
				test.Fatalf("wrong owner: %v", expressionDisagreementError)
			}
			test.Logf("planted disagreement caught by TestTSCCorpusAgreement_%03d: %v", i, expressionDisagreementError)
		}
	}
	if caught != 1 {
		test.Fatalf("caught by %d shards, want exactly one", caught)
	}
}

// Use the existing printer helper's complete inputs for the lowered product.
func tscAgreementProducts(test *testing.T, path, family string) tsPrinterProducts {
	test.Helper()
	return printerGateFamilyProducts(test, family)
}

// Keep the same process status, stderr and byte oracles under a case-only deadline.
func tscAgreementExecute(
	test *testing.T,
	callerContext context.Context,
	environment []string,
	name string,
	arguments ...string,
) oracles.Run {
	test.Helper()
	command := exec.CommandContext(callerContext, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := childguard.Run(command, childguard.Options{})
	if callerContext.Err() != nil {
		test.Fatalf("case deadline: %v", callerContext.Err())
	}
	var exitError *exec.ExitError
	if runError != nil && !errors.As(runError, &exitError) {
		test.Fatalf("running %s: %v", name, runError)
	}
	return oracles.Run{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), ExitCode: command.ProcessState.ExitCode()}
}
