package lower

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/testguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

const cycleCorpusReference = "ffd05d034af5003ea3b70a9080836d10b579b2e0"
const cycleCorpusShardCount = 32

// Exclusions name known compiler work that cannot fit one bounded corpus product.
var cycleCorpusExcluded = map[string]string{
	"internal/oracle/testdata/moves/accepted/objects.a": "lowering is quadratic in impliedTarget (invariance.go:628), over 50 s; filed for the compiler",
}

// A recording piece adds its number only after declaring its per-fixture products.
// Unregistered shards remain explicitly pending; only complete recordings are registered.
var cycleCorpusRegistered = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
var cycleCorpusUpdateUnits = flag.Bool("update-cycle-units", false, "declare bounded units for registered cycle shards")
var cycleCorpusAssemble = flag.Int("assemble-cycle-shard", -1, "assemble one complete shard from prepared per-case products")

type cycleCorpusAnswer struct {
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
	Snapshots []json.RawMessage `json:"snapshots,omitempty"`
}
type cycleCorpusRecording struct {
	Version int                          `json:"version"`
	Shard   int                          `json:"shard"`
	Keys    []string                     `json:"keys"`
	Answers map[string]cycleCorpusAnswer `json:"answers"`
}

func cycleCorpusRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}
func cycleCorpusShard(key string) int {
	h := fnv.New64a()
	h.Write([]byte(key))
	return int(h.Sum64() % cycleCorpusShardCount)
}
func cycleCorpusPaths(t *testing.T) []string {
	t.Helper()
	root := cycleCorpusRoot(t)
	var paths []string
	for _, scope := range []string{"internal/oracle", "stage1", "review"} {
		if _, err := os.Stat(filepath.Join(root, scope)); os.IsNotExist(err) {
			continue
		}
		names, err := tracked.PathsUnder(root, scope, "node_modules", ".generated")
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if _, excluded := cycleCorpusExcluded[name]; excluded {
				continue
			}
			if strings.HasSuffix(name, ".a") || strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".d.ts") {
				paths = append(paths, name)
			}
		}
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatal("empty cycle corpus inventory")
	}
	return paths
}
func cycleCorpusOwned(paths []string, shard int) []string {
	var result []string
	for _, path := range paths {
		if cycleCorpusShard(path) == shard {
			result = append(result, path)
		}
	}
	return result
}
func cycleCorpusInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	inputs, err := buildcache.GoInputs("probe", "./internal/lower ./stage1/cohere/lint/registry", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "cycle-corpus-" + name
	inputs.Files = append(inputs.Files, "internal/lower/cycles_corpus_test.go")
	inputs.Flags = append(inputs.Flags, "reference="+cycleCorpusReference)
	return inputs
}
func cycleCorpusBlob(t *testing.T, inputs buildcache.Inputs, path string) []byte {
	t.Helper()
	data, err := buildcache.PinnedGit(cycleCorpusRoot(t), inputs, cycleCorpusReference, nil, "show", cycleCorpusReference+":"+path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func cycleCorpusBinary(t *testing.T, mode string) string {
	t.Helper()
	inputs := cycleCorpusInputs(t, "binary-"+mode)
	root := cycleCorpusRoot(t)
	sources := map[string]string{}
	for _, path := range []string{"internal/lower/cycles.go", "internal/lower/graph_types.go"} {
		var data []byte
		var err error
		if mode == "frozen" {
			data = cycleCorpusBlob(t, inputs, path)
		} else {
			data, err = os.ReadFile(filepath.Join(root, path))
			if err != nil {
				t.Fatal(err)
			}
		}
		sources[path] = string(data)
	}
	before := "func (l *lowering) findCycles(modules []*ast.SourceFile) error {"
	replacement := "func (l *lowering) findCycles(modules []*ast.SourceFile) (cycleCorpusErr error) {"
	if strings.Count(sources["internal/lower/cycles.go"], before) != 1 {
		t.Fatal("cycle finder entry changed")
	}
	sources["internal/lower/cycles.go"] = strings.Replace(sources["internal/lower/cycles.go"], before, replacement, 1)
	anchor := "finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}"
	if strings.Count(sources["internal/lower/cycles.go"], anchor) != 1 {
		t.Fatal("cycle finder construction changed")
	}
	sources["internal/lower/cycles.go"] = strings.Replace(sources["internal/lower/cycles.go"], anchor, anchor+"\n defer cycleCorpusObserve(finder,&cycleCorpusErr)", 1)
	if mode == "miss-edge" {
		anchor := "func (f *cycleFinder) reachLinks(node cycleNode) []cycleNode {"
		if strings.Count(sources["internal/lower/cycles.go"], anchor) != 1 {
			t.Fatal("reach index anchor changed")
		}
		sources["internal/lower/cycles.go"] = strings.Replace(sources["internal/lower/cycles.go"], anchor, anchor+"\n return nil", 1)
	}
	observer := cycleCorpusObserver
	if mode == "frozen" {
		observer = strings.ReplaceAll(observer, "f.cycleUnproven(", "f.unproven(")
	}
	sources["internal/lower/cycle_corpus_observer.go"] = observer
	sources["cmd/adamic-stage1/cycle_corpus_probe.go"] = cycleCorpusDriver
	directory := buildcache.Product(t, inputs, func(directory string) error {
		replacements := map[string]string{}
		for path, source := range sources {
			target := filepath.Join(directory, strings.ReplaceAll(path, "/", "_")+".txt")
			if err := os.WriteFile(target, []byte(source), 0644); err != nil {
				return err
			}
			defer os.Remove(target)
			replacements[filepath.Join(root, path)] = target
		}
		data, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			return err
		}
		overlay := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(overlay, data, 0644); err != nil {
			return err
		}
		defer os.Remove(overlay)
		args := append([]string{"build"}, buildcache.ReproducibleGoFlags()...)
		args = append(args, "-p", "1", "-overlay="+overlay, "-o", filepath.Join(directory, "probe"), "./cmd/adamic-stage1/cycle_corpus_probe.go")
		command := toolchain.CommandIn(root, "go", args...)
		var output bytes.Buffer
		command.Stdout = &output
		command.Stderr = &output
		if err := testguard.Run(command, 50*time.Second, 55*time.Second); err != nil {
			return fmt.Errorf("cycle probe build: %w: %s", err, &output)
		}
		return nil
	})
	return filepath.Join(directory, "probe")
}
func cycleCorpusRun(t *testing.T, binary, key string, budget time.Duration) (cycleCorpusAnswer, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	command := toolchain.CommandContextIn(ctx, cycleCorpusRoot(t), binary, key)
	command.Env = append(command.Env, "GOMAXPROCS=1", "GOMEMLIMIT=2GiB")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := testguard.Run(command, budget, budget+time.Second); err != nil {
		return cycleCorpusAnswer{}, fmt.Errorf("cycle probe %s: %w: %s", key, err, stderr.String())
	}
	if stderr.Len() != 0 {
		return cycleCorpusAnswer{}, fmt.Errorf("cycle probe stderr: %s", stderr.String())
	}
	var answer cycleCorpusAnswer
	decoder := json.NewDecoder(&stdout)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&answer); err != nil {
		return answer, fmt.Errorf("cycle probe answer: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return answer, fmt.Errorf("cycle probe returned extra or malformed output")
	}
	if err := cycleCorpusValidAnswer(answer); err != nil {
		return answer, err
	}
	return answer, nil
}
func cycleCorpusValidAnswer(answer cycleCorpusAnswer) error {
	switch answer.Status {
	case "load", "lower":
		if answer.Error == "" || len(answer.Snapshots) != 0 {
			return fmt.Errorf("invalid pre-finder refusal")
		}
	case "finder":
		if len(answer.Snapshots) == 0 {
			return fmt.Errorf("finder answer has no snapshot")
		}
		for _, raw := range answer.Snapshots {
			var snapshot struct {
				Types []string
				Slots []json.RawMessage
				Graph map[string][]string
				Error string
			}
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&snapshot); err != nil || snapshot.Types == nil || snapshot.Slots == nil || snapshot.Graph == nil {
				return fmt.Errorf("invalid finder snapshot")
			}
		}
	default:
		return fmt.Errorf("invalid cycle status %q", answer.Status)
	}
	return nil
}
func cycleCorpusObservation(t *testing.T, key, mode string, prepare bool, owners ...string) cycleCorpusAnswer {
	t.Helper()
	inputs := cycleCorpusInputs(t, "answer-"+mode+"-"+fmt.Sprintf("%x", sha256.Sum256([]byte(key))))
	inputs.Flags = append(inputs.Flags, "case="+key)
	inputs.Files = append(inputs.Files, "stage1", "internal/oracle/testdata")
	if _, err := os.Stat(filepath.Join(cycleCorpusRoot(t), "review")); err == nil {
		inputs.Files = append(inputs.Files, "review")
	}
	build := func(directory string) error {
		binary := cycleCorpusBinary(t, mode)
		answer, err := cycleCorpusRun(t, binary, key, 50*time.Second)
		if err != nil {
			return err
		}
		data, err := json.Marshal(answer)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "answer.json"), data, 0644)
	}
	var directory string
	if prepare {
		directory = buildcache.Product(t, inputs, build)
	} else {
		owner := "internal/lower TestProduct_Cycle" + strings.Title(mode) + "_" + cycleCorpusUnitID(key)
		if len(owners) > 0 {
			owner = owners[0]
		}
		directory = buildcache.BuiltElsewhere(t, "prepared cycle answer", inputs, owner, func(string) error { return fmt.Errorf("prepare %s first; assembly never lowers cases", owner) })
	}
	data, err := os.ReadFile(filepath.Join(directory, "answer.json"))
	if err != nil {
		t.Fatal(err)
	}
	var answer cycleCorpusAnswer
	if err := json.Unmarshal(data, &answer); err != nil {
		t.Fatal(err)
	}
	if err := cycleCorpusValidAnswer(answer); err != nil {
		t.Fatal(err)
	}
	return answer
}
func cycleCorpusSame(one, two cycleCorpusAnswer) bool {
	a, _ := json.Marshal(one)
	b, _ := json.Marshal(two)
	return bytes.Equal(a, b)
}
func cycleCorpusValidate(record cycleCorpusRecording, keys []string, shard int) error {
	if record.Version != 1 || record.Shard != shard || !slices.Equal(record.Keys, keys) || len(record.Answers) != len(keys) {
		return fmt.Errorf("cycle shard ownership or completeness differs")
	}
	for _, key := range keys {
		answer, ok := record.Answers[key]
		if !ok {
			return fmt.Errorf("cycle shard missing %s", key)
		}
		if err := cycleCorpusValidAnswer(answer); err != nil {
			return err
		}
	}
	return nil
}
func cycleCorpusFile(shard int) string { return fmt.Sprintf("testdata/cycles/%02d.json", shard) }
func cycleCorpusRead(t *testing.T, shard int) cycleCorpusRecording {
	t.Helper()
	data, err := os.ReadFile(cycleCorpusFile(shard))
	if err != nil {
		t.Fatal(err)
	}
	var record cycleCorpusRecording
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	return record
}
func cycleCorpusVerify(t *testing.T, key string) {
	t.Helper()
	record := cycleCorpusRead(t, cycleCorpusShard(key))
	want, ok := record.Answers[key]
	if !ok {
		t.Fatal("recording does not own key", key)
	}
	frozen := cycleCorpusObservation(t, key, "frozen", false)
	current := cycleCorpusObservation(t, key, "current", false)
	if !cycleCorpusSame(frozen, current) || !cycleCorpusSame(want, current) {
		t.Fatal("full cycle answer changed", key)
	}
}
func cycleCorpusUnitID(key string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))[:16] }
func cycleCorpusUnits(t *testing.T) string {
	t.Helper()
	var body strings.Builder
	paths := cycleCorpusPaths(t)
	for _, shard := range cycleCorpusRegistered {
		for _, key := range cycleCorpusOwned(paths, shard) {
			id := cycleCorpusUnitID(key)
			fmt.Fprintf(&body, "func TestProduct_CycleFrozen_%s(t *testing.T){t.Parallel();cycleCorpusObservation(t,%q,\"frozen\",true)}\n", id, key)
			fmt.Fprintf(&body, "func TestProduct_CycleCurrent_%s(t *testing.T){t.Parallel();cycleCorpusObservation(t,%q,\"current\",true)}\n", id, key)
			fmt.Fprintf(&body, "func TestCyclesCorpus_%s(t *testing.T){t.Parallel();cycleCorpusVerify(t,%q)}\n", id, key)
		}
	}
	return body.String()
}
func TestCyclesCorpusExclusions(t *testing.T) {
	t.Parallel()
	const name = "internal/oracle/testdata/moves/accepted/objects.a"
	const reason = "lowering is quadratic in impliedTarget (invariance.go:628), over 50 s; filed for the compiler"
	if len(cycleCorpusExcluded) != 1 || cycleCorpusExcluded[name] != reason {
		t.Fatal("cycle exclusions must name only moves/accepted/objects.a with its filed compiler reason")
	}
	if _, err := os.Stat(filepath.Join(cycleCorpusRoot(t), name)); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(cycleCorpusPaths(t), name) {
		t.Fatal("excluded program still belongs to the cycle corpus")
	}
}

func TestCyclesCorpusUnitsAreDeclared(t *testing.T) {
	t.Parallel()
	path := "cycles_corpus_test.go"
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	prefix, rest, ok := strings.Cut(string(source), "// BEGIN RECORDED CYCLE UNITS\n")
	if !ok {
		t.Fatal("missing cycle unit marker")
	}
	_, suffix, ok := strings.Cut(rest, "// END RECORDED CYCLE UNITS")
	if !ok {
		t.Fatal("missing cycle unit end")
	}
	wanted := prefix + "// BEGIN RECORDED CYCLE UNITS\n" + cycleCorpusUnits(t) + "// END RECORDED CYCLE UNITS" + suffix
	data, err := format.Source([]byte(wanted))
	if err != nil {
		t.Fatal(err)
	}
	if *cycleCorpusUpdateUnits {
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if !bytes.Equal(data, source) {
		t.Fatal("cycle units differ; generate declarations before preparing answers with -update-cycle-units")
	}
}
func TestCyclesCorpusInventory(t *testing.T) {
	t.Parallel()
	paths := cycleCorpusPaths(t)
	seen := map[int]bool{}
	total := 0
	for _, shard := range cycleCorpusRegistered {
		if shard < 0 || shard >= cycleCorpusShardCount || seen[shard] {
			t.Fatal("invalid or duplicate cycle shard", shard)
		}
		seen[shard] = true
		keys := cycleCorpusOwned(paths, shard)
		record := cycleCorpusRead(t, shard)
		if err := cycleCorpusValidate(record, keys, shard); err != nil {
			t.Fatal(err)
		}
		total += len(keys)
	}
	files, err := filepath.Glob("testdata/cycles/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(seen) {
		t.Fatal("recording files and registrations differ")
	}
	t.Logf("%d selected programs; %d recorded; %d registered shards; %d explicitly pending", len(paths), total, len(seen), cycleCorpusShardCount-len(seen))
}

// Not parallel: an explicit update atomically writes one new recording file.
func TestCyclesCorpusAssembleShard(t *testing.T) {
	shard := *cycleCorpusAssemble
	if shard < 0 {
		t.Skip("prepare per-fixture products, then use -assemble-cycle-shard=N")
	}
	if shard >= cycleCorpusShardCount {
		t.Fatal("invalid cycle shard")
	}
	keys := cycleCorpusOwned(cycleCorpusPaths(t), shard)
	record := cycleCorpusRecording{Version: 1, Shard: shard, Keys: keys, Answers: map[string]cycleCorpusAnswer{}}
	for _, key := range keys {
		frozen := cycleCorpusObservation(t, key, "frozen", false)
		current := cycleCorpusObservation(t, key, "current", false)
		if !cycleCorpusSame(frozen, current) {
			t.Fatal("frozen finder differs; no recording written", key)
		}
		record.Answers[key] = current
	}
	if err := cycleCorpusValidate(record, keys, shard); err != nil {
		t.Fatal(err)
	}
	if err := cycleCorpusWrite(cycleCorpusFile(shard), record, keys, shard); err != nil {
		t.Fatal(err)
	}
}
func cycleCorpusWrite(path string, record cycleCorpusRecording, keys []string, shard int) error {
	if err := cycleCorpusValidate(record, keys, shard); err != nil {
		return err
	}
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".cycle-recording-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func TestProduct_CycleCorpusFrozen(t *testing.T)     { t.Parallel(); cycleCorpusBinary(t, "frozen") }
func TestProduct_CycleCorpusCurrent(t *testing.T)    { t.Parallel(); cycleCorpusBinary(t, "current") }
func TestProduct_CycleCorpusMissedEdge(t *testing.T) { t.Parallel(); cycleCorpusBinary(t, "miss-edge") }

const cycleCorpusControl = `class Owner {readonly tag='owner';next:Box|undefined;constructor(){this.next=undefined;}} class Box {readonly tag='box';back:Owner|undefined;constructor(){this.back=undefined;}} const owner=new Owner();const box=new Box();owner.next=box;box.back=owner;console.log(owner.tag);`

func cycleCorpusControlPath(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
func TestCyclesCorpusFrozenControl(t *testing.T) {
	t.Parallel()
	path := cycleCorpusControlPath(t, cycleCorpusControl)
	frozen, err := cycleCorpusRun(t, cycleCorpusBinary(t, "frozen"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	current, err := cycleCorpusRun(t, cycleCorpusBinary(t, "current"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if frozen.Status != "finder" || frozen.Error != "" || !cycleCorpusSame(frozen, current) {
		t.Fatal("frozen admitted graph control differs")
	}
}
func TestCyclesCorpusRefusalControl(t *testing.T) {
	t.Parallel()
	path := cycleCorpusControlPath(t, "interface Link {next?:Link} const link:Link={};link.next=link;")
	frozen, err := cycleCorpusRun(t, cycleCorpusBinary(t, "frozen"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	current, err := cycleCorpusRun(t, cycleCorpusBinary(t, "current"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if frozen.Error == "" || !cycleCorpusSame(frozen, current) {
		t.Fatal("frozen record refusal differs")
	}
}
func TestCyclesCorpusDetectsMissedEdge(t *testing.T) {
	t.Parallel()
	path := cycleCorpusControlPath(t, cycleCorpusControl)
	frozen, err := cycleCorpusRun(t, cycleCorpusBinary(t, "frozen"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	mutant, err := cycleCorpusRun(t, cycleCorpusBinary(t, "miss-edge"), path, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if cycleCorpusSame(frozen, mutant) {
		t.Fatal("missed reachability edge mutant survived full-result comparator")
	}
}
func cycleCorpusFailure(t *testing.T, key string, budget time.Duration) {
	t.Helper()
	if _, err := cycleCorpusRun(t, cycleCorpusBinary(t, "current"), key, budget); err == nil {
		t.Fatal("failed child accepted", key)
	}
}
func TestCyclesCorpusRejectsPanic(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@panic", time.Second)
}
func TestCyclesCorpusRejectsExit(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@exit", time.Second)
}
func TestCyclesCorpusRejectsMissing(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@missing", time.Second)
}
func TestCyclesCorpusRejectsMalformed(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@malformed", time.Second)
}
func TestCyclesCorpusRejectsDuplicate(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@duplicate", time.Second)
}
func TestCyclesCorpusRejectsTimeout(t *testing.T) {
	t.Parallel()
	cycleCorpusFailure(t, "@timeout", 100*time.Millisecond)
}
func TestCyclesCorpusRejectsPartialRegeneration(t *testing.T) {
	t.Parallel()
	answer := cycleCorpusAnswer{Status: "load", Error: "refused"}
	keys := []string{"a", "b"}
	complete := cycleCorpusRecording{Version: 1, Shard: 0, Keys: keys, Answers: map[string]cycleCorpusAnswer{"a": answer, "b": answer}}
	if err := cycleCorpusValidate(complete, keys, 0); err != nil {
		t.Fatal(err)
	}
	delete(complete.Answers, "b")
	if err := cycleCorpusValidate(complete, keys, 0); err == nil {
		t.Fatal("partial regeneration mutant survived")
	}
}
func TestCyclesCorpusOwnershipMutants(t *testing.T) {
	t.Parallel()
	answer := cycleCorpusAnswer{Status: "load", Error: "refused"}
	keys := []string{"a", "b"}
	for _, bad := range [][]string{{"a"}, {"a", "a"}, {"a", "unknown"}, {"b", "a"}} {
		record := cycleCorpusRecording{Version: 1, Shard: 0, Keys: bad, Answers: map[string]cycleCorpusAnswer{"a": answer, "b": answer}}
		if err := cycleCorpusValidate(record, keys, 0); err == nil {
			t.Fatal("ownership mutant survived", bad)
		}
	}
	for _, bad := range []cycleCorpusAnswer{{Status: "panic", Error: "panic"}, {Status: "finder"}, {Status: "load"}} {
		if err := cycleCorpusValidAnswer(bad); err == nil {
			t.Fatal("invalid status survived", bad)
		}
	}
}
func TestCyclesCorpusTscSelection(t *testing.T) {
	t.Parallel()
	path := "internal/oracle/testdata/program_region/census.csv"
	inputs := cycleCorpusInputs(t, "tsc-selection")
	frozen := cycleCorpusBlob(t, inputs, path)
	current, err := os.ReadFile(filepath.Join(cycleCorpusRoot(t), path))
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(bytes.NewReader(current)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1686 || !bytes.Equal(frozen, current) {
		t.Fatal("pinned 1685-record tsc selection changed")
	}
	mutant := bytes.Clone(current)
	mutant[len(mutant)/2] ^= 1
	if bytes.Equal(frozen, mutant) {
		t.Fatal("tsc selection byte mutant survived")
	}
}

const cycleCorpusAnswerControl = "internal/oracle/testdata/release_fma.a"

func TestProduct_CycleCorpusFrozenAnswer(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, cycleCorpusAnswerControl, "frozen", true)
}
func TestProduct_CycleCorpusCurrentAnswer(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, cycleCorpusAnswerControl, "current", true)
}
func TestCyclesCorpusPreparedAnswers(t *testing.T) {
	t.Parallel()
	frozen := cycleCorpusObservation(t, cycleCorpusAnswerControl, "frozen", false, "internal/lower TestProduct_CycleCorpusFrozenAnswer")
	current := cycleCorpusObservation(t, cycleCorpusAnswerControl, "current", false, "internal/lower TestProduct_CycleCorpusCurrentAnswer")
	if !cycleCorpusSame(frozen, current) {
		t.Fatal("prepared full answers differ")
	}
	keys := []string{cycleCorpusAnswerControl}
	shard := cycleCorpusShard(cycleCorpusAnswerControl)
	record := cycleCorpusRecording{Version: 1, Shard: shard, Keys: keys, Answers: map[string]cycleCorpusAnswer{cycleCorpusAnswerControl: current}}
	path := filepath.Join(t.TempDir(), "record.json")
	if err := cycleCorpusWrite(path, record, keys, shard); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	record.Answers = map[string]cycleCorpusAnswer{}
	if err := cycleCorpusWrite(path, record, keys, shard); err == nil {
		t.Fatal("partial write was accepted")
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("failed regeneration changed prior recording", err)
	}
}

// BEGIN RECORDED CYCLE UNITS
func TestProduct_CycleFrozen_4a73a04aa3deba20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/hidden_mutated.a", "frozen", true)
}
func TestProduct_CycleCurrent_4a73a04aa3deba20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/hidden_mutated.a", "current", true)
}
func TestCyclesCorpus_4a73a04aa3deba20(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_review/refused/hidden_mutated.a")
}
func TestProduct_CycleFrozen_840bbc2c3300a5d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_rest.a", "frozen", true)
}
func TestProduct_CycleCurrent_840bbc2c3300a5d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_rest.a", "current", true)
}
func TestCyclesCorpus_840bbc2c3300a5d5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_rest.a")
}
func TestProduct_CycleFrozen_5f852fb1a6610b42(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_host24.a", "frozen", true)
}
func TestProduct_CycleCurrent_5f852fb1a6610b42(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_host24.a", "current", true)
}
func TestCyclesCorpus_5f852fb1a6610b42(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/closure_convention_host24.a")
}
func TestProduct_CycleFrozen_b10d28e918db8f80(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/process.a", "frozen", true)
}
func TestProduct_CycleCurrent_b10d28e918db8f80(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/process.a", "current", true)
}
func TestCyclesCorpus_b10d28e918db8f80(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/process.a")
}
func TestProduct_CycleFrozen_7bab2768e2abf6cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_regexp_literal.a", "frozen", true)
}
func TestProduct_CycleCurrent_7bab2768e2abf6cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_regexp_literal.a", "current", true)
}
func TestCyclesCorpus_7bab2768e2abf6cf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_regexp_literal.a")
}
func TestProduct_CycleFrozen_ee532a6304452a9c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_field_present.a", "frozen", true)
}
func TestProduct_CycleCurrent_ee532a6304452a9c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_field_present.a", "current", true)
}
func TestCyclesCorpus_ee532a6304452a9c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/e4eec87_f1_field_present.a")
}
func TestProduct_CycleFrozen_58ef2b2cb7c1e428(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums.a", "frozen", true)
}
func TestProduct_CycleCurrent_58ef2b2cb7c1e428(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums.a", "current", true)
}
func TestCyclesCorpus_58ef2b2cb7c1e428(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums.a")
}
func TestProduct_CycleFrozen_8189c3ed2f791a5d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_const.a", "frozen", true)
}
func TestProduct_CycleCurrent_8189c3ed2f791a5d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_const.a", "current", true)
}
func TestCyclesCorpus_8189c3ed2f791a5d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_const.a")
}
func TestProduct_CycleFrozen_2e7c2aecde23424e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_flags_modules/main.a", "frozen", true)
}
func TestProduct_CycleCurrent_2e7c2aecde23424e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_flags_modules/main.a", "current", true)
}
func TestCyclesCorpus_2e7c2aecde23424e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_flags_modules/main.a")
}
func TestProduct_CycleFrozen_ce9f3090c42748a9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fills.a", "frozen", true)
}
func TestProduct_CycleCurrent_ce9f3090c42748a9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fills.a", "current", true)
}
func TestCyclesCorpus_ce9f3090c42748a9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fills.a")
}
func TestProduct_CycleFrozen_71cd8e2576ca0a07(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/finally_leaves.a", "frozen", true)
}
func TestProduct_CycleCurrent_71cd8e2576ca0a07(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/finally_leaves.a", "current", true)
}
func TestCyclesCorpus_71cd8e2576ca0a07(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/finally_leaves.a")
}
func TestProduct_CycleFrozen_ab2c5f390bfdac1a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_returns_argument.a", "frozen", true)
}
func TestProduct_CycleCurrent_ab2c5f390bfdac1a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_returns_argument.a", "current", true)
}
func TestCyclesCorpus_ab2c5f390bfdac1a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_returns_argument.a")
}
func TestProduct_CycleFrozen_e5f096f999d65613(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_higher_callback_pair_methods.a", "frozen", true)
}
func TestProduct_CycleCurrent_e5f096f999d65613(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_higher_callback_pair_methods.a", "current", true)
}
func TestCyclesCorpus_e5f096f999d65613(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_higher_callback_pair_methods.a")
}
func TestProduct_CycleFrozen_e2f324a5f6731d13(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_escape.a", "frozen", true)
}
func TestProduct_CycleCurrent_e2f324a5f6731d13(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_escape.a", "current", true)
}
func TestCyclesCorpus_e2f324a5f6731d13(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_escape.a")
}
func TestProduct_CycleFrozen_9586336037b17c7d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_10.a", "frozen", true)
}
func TestProduct_CycleCurrent_9586336037b17c7d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_10.a", "current", true)
}
func TestCyclesCorpus_9586336037b17c7d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_regression_10.a")
}
func TestProduct_CycleFrozen_57b7c6cd2a5d3d58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_static_parent.a", "frozen", true)
}
func TestProduct_CycleCurrent_57b7c6cd2a5d3d58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_static_parent.a", "current", true)
}
func TestCyclesCorpus_57b7c6cd2a5d3d58(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_static_parent.a")
}
func TestProduct_CycleFrozen_b4f3bca332712fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_map_iterator_union.a", "frozen", true)
}
func TestProduct_CycleCurrent_b4f3bca332712fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_map_iterator_union.a", "current", true)
}
func TestCyclesCorpus_b4f3bca332712fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/host_map_iterator_union.a")
}
func TestProduct_CycleFrozen_356690b911fc7865(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/imported_const_cases/constants.a", "frozen", true)
}
func TestProduct_CycleCurrent_356690b911fc7865(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/imported_const_cases/constants.a", "current", true)
}
func TestCyclesCorpus_356690b911fc7865(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/imported_const_cases/constants.a")
}
func TestProduct_CycleFrozen_75f1a6e2d8712cde(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_globals_typeof.a", "frozen", true)
}
func TestProduct_CycleCurrent_75f1a6e2d8712cde(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_globals_typeof.a", "current", true)
}
func TestCyclesCorpus_75f1a6e2d8712cde(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_globals_typeof.a")
}
func TestProduct_CycleFrozen_6a07885731b2dd54(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_keys_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_6a07885731b2dd54(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_keys_spread.a", "current", true)
}
func TestCyclesCorpus_6a07885731b2dd54(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_keys_spread.a")
}
func TestProduct_CycleFrozen_a88cbc28ec72e7aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_values_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_a88cbc28ec72e7aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_values_assign.a", "current", true)
}
func TestCyclesCorpus_a88cbc28ec72e7aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_values_assign.a")
}
func TestProduct_CycleFrozen_42a1ddc8d2f57972(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_42a1ddc8d2f57972(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_order.a", "current", true)
}
func TestCyclesCorpus_42a1ddc8d2f57972(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_order.a")
}
func TestProduct_CycleFrozen_a0b329fa45a0a731(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_own.a", "frozen", true)
}
func TestProduct_CycleCurrent_a0b329fa45a0a731(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_own.a", "current", true)
}
func TestCyclesCorpus_a0b329fa45a0a731(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_own.a")
}
func TestProduct_CycleFrozen_0545fb85d04c8b62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/maybe_booleans.a", "frozen", true)
}
func TestProduct_CycleCurrent_0545fb85d04c8b62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/maybe_booleans.a", "current", true)
}
func TestCyclesCorpus_0545fb85d04c8b62(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/maybe_booleans.a")
}
func TestProduct_CycleFrozen_72ba8c49a9bd2929(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_repeat.a", "frozen", true)
}
func TestProduct_CycleCurrent_72ba8c49a9bd2929(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_repeat.a", "current", true)
}
func TestCyclesCorpus_72ba8c49a9bd2929(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_repeat.a")
}
func TestProduct_CycleFrozen_ddc61976f00ff7f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespace_class_registration.a", "frozen", true)
}
func TestProduct_CycleCurrent_ddc61976f00ff7f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespace_class_registration.a", "current", true)
}
func TestCyclesCorpus_ddc61976f00ff7f4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespace_class_registration.a")
}
func TestProduct_CycleFrozen_4681da08140f8aa6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_assignment_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_4681da08140f8aa6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_assignment_return.a", "current", true)
}
func TestCyclesCorpus_4681da08140f8aa6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_assignment_return.a")
}
func TestProduct_CycleFrozen_db464c3c0ab142b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/numbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_db464c3c0ab142b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/numbers.a", "current", true)
}
func TestCyclesCorpus_db464c3c0ab142b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/numbers.a")
}
func TestProduct_CycleFrozen_6946adc38312e9a2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_exists.a", "frozen", true)
}
func TestProduct_CycleCurrent_6946adc38312e9a2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_exists.a", "current", true)
}
func TestCyclesCorpus_6946adc38312e9a2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_exists.a")
}
func TestProduct_CycleFrozen_926fee3947a88223(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_ancestor_directory.a", "frozen", true)
}
func TestProduct_CycleCurrent_926fee3947a88223(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_ancestor_directory.a", "current", true)
}
func TestCyclesCorpus_926fee3947a88223(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_ancestor_directory.a")
}
func TestProduct_CycleFrozen_c22cc51296a6beda(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/edit.a", "frozen", true)
}
func TestProduct_CycleCurrent_c22cc51296a6beda(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/edit.a", "current", true)
}
func TestCyclesCorpus_c22cc51296a6beda(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/program_region/edit.a")
}
func TestProduct_CycleFrozen_093193fab59502d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_move_before_ready.a", "frozen", true)
}
func TestProduct_CycleCurrent_093193fab59502d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_move_before_ready.a", "current", true)
}
func TestCyclesCorpus_093193fab59502d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_move_before_ready.a")
}
func TestProduct_CycleFrozen_41600e8b40673b02(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_spread_method.a", "frozen", true)
}
func TestProduct_CycleCurrent_41600e8b40673b02(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_spread_method.a", "current", true)
}
func TestCyclesCorpus_41600e8b40673b02(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_spread_method.a")
}
func TestProduct_CycleFrozen_87e3cc58cc4db068(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_pad.a", "frozen", true)
}
func TestProduct_CycleCurrent_87e3cc58cc4db068(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_pad.a", "current", true)
}
func TestCyclesCorpus_87e3cc58cc4db068(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_pad.a")
}
func TestProduct_CycleFrozen_c4c531ec1b3c89fd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e23c7ab_fold_strings.a", "frozen", true)
}
func TestProduct_CycleCurrent_c4c531ec1b3c89fd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e23c7ab_fold_strings.a", "current", true)
}
func TestCyclesCorpus_c4c531ec1b3c89fd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_e23c7ab_fold_strings.a")
}
func TestProduct_CycleFrozen_006d7fdfd2a8b5ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_numeric_error.a", "frozen", true)
}
func TestProduct_CycleCurrent_006d7fdfd2a8b5ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_numeric_error.a", "current", true)
}
func TestCyclesCorpus_006d7fdfd2a8b5ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_numeric_error.a")
}
func TestProduct_CycleFrozen_ae58b063a29f51f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_host_directory.a", "frozen", true)
}
func TestProduct_CycleCurrent_ae58b063a29f51f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_host_directory.a", "current", true)
}
func TestCyclesCorpus_ae58b063a29f51f0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_host_directory.a")
}
func TestProduct_CycleFrozen_81a1f3e9f26f0992(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p08_null_array_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_81a1f3e9f26f0992(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p08_null_array_write.a", "current", true)
}
func TestCyclesCorpus_81a1f3e9f26f0992(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p08_null_array_write.a")
}
func TestProduct_CycleFrozen_299480f834c593d8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p26_callable_or_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_299480f834c593d8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p26_callable_or_undefined.a", "current", true)
}
func TestCyclesCorpus_299480f834c593d8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p26_callable_or_undefined.a")
}
func TestProduct_CycleFrozen_09fb7fcfac147b79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p07_element_access_union_wrong.a", "frozen", true)
}
func TestProduct_CycleCurrent_09fb7fcfac147b79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p07_element_access_union_wrong.a", "current", true)
}
func TestCyclesCorpus_09fb7fcfac147b79(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p07_element_access_union_wrong.a")
}
func TestProduct_CycleFrozen_72aaebed37b01a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p32_boxed_slot_string_good.a", "frozen", true)
}
func TestProduct_CycleCurrent_72aaebed37b01a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p32_boxed_slot_string_good.a", "current", true)
}
func TestCyclesCorpus_72aaebed37b01a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p32_boxed_slot_string_good.a")
}
func TestProduct_CycleFrozen_aa6cef27a15ae523(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_closures.a", "frozen", true)
}
func TestProduct_CycleCurrent_aa6cef27a15ae523(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_closures.a", "current", true)
}
func TestCyclesCorpus_aa6cef27a15ae523(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_closures.a")
}
func TestProduct_CycleFrozen_7fd0076a9dee7402(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/spread_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_7fd0076a9dee7402(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/spread_undefined.a", "current", true)
}
func TestCyclesCorpus_7fd0076a9dee7402(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/spread_undefined.a")
}
func TestProduct_CycleFrozen_335d6fc6f4de95de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_append.a", "frozen", true)
}
func TestProduct_CycleCurrent_335d6fc6f4de95de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_append.a", "current", true)
}
func TestCyclesCorpus_335d6fc6f4de95de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_append.a")
}
func TestProduct_CycleFrozen_b59cb9d1a6e477a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_padding.a", "frozen", true)
}
func TestProduct_CycleCurrent_b59cb9d1a6e477a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_padding.a", "current", true)
}
func TestCyclesCorpus_b59cb9d1a6e477a3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_build_padding.a")
}
func TestProduct_CycleFrozen_eb66dada3114a45b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_lifetime.a", "frozen", true)
}
func TestProduct_CycleCurrent_eb66dada3114a45b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_lifetime.a", "current", true)
}
func TestCyclesCorpus_eb66dada3114a45b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_views_lifetime.a")
}
func TestProduct_CycleFrozen_d5fee70fd4c55427(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/sweeps/regexp_methods.a", "frozen", true)
}
func TestProduct_CycleCurrent_d5fee70fd4c55427(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/sweeps/regexp_methods.a", "current", true)
}
func TestCyclesCorpus_d5fee70fd4c55427(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/sweeps/regexp_methods.a")
}
func TestProduct_CycleFrozen_9bc424a8708bca72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/overload.a", "frozen", true)
}
func TestProduct_CycleCurrent_9bc424a8708bca72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/overload.a", "current", true)
}
func TestCyclesCorpus_9bc424a8708bca72(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/switch_case_declarations/overload.a")
}
func TestProduct_CycleFrozen_f61cdf5271b3778f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_sort_float64array.a", "frozen", true)
}
func TestProduct_CycleCurrent_f61cdf5271b3778f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_sort_float64array.a", "current", true)
}
func TestCyclesCorpus_f61cdf5271b3778f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_sort_float64array.a")
}
func TestProduct_CycleFrozen_c0ea048a4cdfc464(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_stdout_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_c0ea048a4cdfc464(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_stdout_order.a", "current", true)
}
func TestCyclesCorpus_c0ea048a4cdfc464(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/write_stdout_order.a")
}
func TestProduct_CycleFrozen_7c1cd96d7a77fe72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/testdata/byte_slices.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7c1cd96d7a77fe72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/testdata/byte_slices.ts", "current", true)
}
func TestCyclesCorpus_7c1cd96d7a77fe72(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/testdata/byte_slices.ts")
}
func TestProduct_CycleFrozen_62c394b3120df686(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/rawInput.ts", "frozen", true)
}
func TestProduct_CycleCurrent_62c394b3120df686(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/rawInput.ts", "current", true)
}
func TestCyclesCorpus_62c394b3120df686(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/rawInput.ts")
}
func TestProduct_CycleFrozen_ea61f1bf3cadf2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/pipeline.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ea61f1bf3cadf2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/pipeline.ts", "current", true)
}
func TestCyclesCorpus_ea61f1bf3cadf2de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/pipeline.ts")
}
func TestProduct_CycleFrozen_290de50b745da42a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceValidation.ts", "frozen", true)
}
func TestProduct_CycleCurrent_290de50b745da42a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceValidation.ts", "current", true)
}
func TestCyclesCorpus_290de50b745da42a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceValidation.ts")
}
func TestProduct_CycleFrozen_82a85ae35f75123f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_82a85ae35f75123f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/main.ts", "current", true)
}
func TestCyclesCorpus_82a85ae35f75123f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/main.ts")
}
func TestProduct_CycleFrozen_335d3b9033c8c5ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/002.a", "frozen", true)
}
func TestProduct_CycleCurrent_335d3b9033c8c5ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/002.a", "current", true)
}
func TestCyclesCorpus_335d3b9033c8c5ff(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/002.a")
}
func TestProduct_CycleFrozen_1610584a7fa53347(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_1610584a7fa53347(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018_runtime.a", "current", true)
}
func TestCyclesCorpus_1610584a7fa53347(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018_runtime.a")
}
func TestProduct_CycleFrozen_e5f47504326669a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_e5f47504326669a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025_runtime.a", "current", true)
}
func TestCyclesCorpus_e5f47504326669a8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025_runtime.a")
}
func TestProduct_CycleFrozen_06b02914ad86f82c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_06b02914ad86f82c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043_runtime.a", "current", true)
}
func TestCyclesCorpus_06b02914ad86f82c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043_runtime.a")
}
func TestProduct_CycleFrozen_2ab150cf67ae4f05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/051.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ab150cf67ae4f05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/051.a", "current", true)
}
func TestCyclesCorpus_2ab150cf67ae4f05(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/051.a")
}
func TestProduct_CycleFrozen_87784f88edd1f2d3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_87784f88edd1f2d3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069_runtime.a", "current", true)
}
func TestCyclesCorpus_87784f88edd1f2d3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069_runtime.a")
}
func TestProduct_CycleFrozen_21639cfa3cf3d55d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/077.a", "frozen", true)
}
func TestProduct_CycleCurrent_21639cfa3cf3d55d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/077.a", "current", true)
}
func TestCyclesCorpus_21639cfa3cf3d55d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/077.a")
}
func TestProduct_CycleFrozen_0b761ba82b61f240(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_0b761ba82b61f240(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087_runtime.a", "current", true)
}
func TestCyclesCorpus_0b761ba82b61f240(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087_runtime.a")
}
func TestProduct_CycleFrozen_563c7c595723013f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/s/031_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_563c7c595723013f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/s/031_runtime.a", "current", true)
}
func TestCyclesCorpus_563c7c595723013f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/s/031_runtime.a")
}
func TestProduct_CycleFrozen_6238ead901b1db01(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/comments.a", "frozen", true)
}
func TestProduct_CycleCurrent_6238ead901b1db01(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/comments.a", "current", true)
}
func TestCyclesCorpus_6238ead901b1db01(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/comments.a")
}
func TestProduct_CycleFrozen_b9fc7fb68a1d103c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-duplicate-case/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b9fc7fb68a1d103c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-duplicate-case/rule.ts", "current", true)
}
func TestCyclesCorpus_b9fc7fb68a1d103c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-duplicate-case/rule.ts")
}
func TestProduct_CycleFrozen_9dca3aec29eac081(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-pattern/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_9dca3aec29eac081(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-pattern/messages.a", "current", true)
}
func TestCyclesCorpus_9dca3aec29eac081(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-empty-pattern/messages.a")
}
func TestProduct_CycleFrozen_7964bfc950b7b5d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-static-block/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_7964bfc950b7b5d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-static-block/messages.a", "current", true)
}
func TestCyclesCorpus_7964bfc950b7b5d4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-empty-static-block/messages.a")
}
func TestProduct_CycleFrozen_1e379ae35460b106(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-self-compare/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_1e379ae35460b106(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-self-compare/rule.a", "current", true)
}
func TestCyclesCorpus_1e379ae35460b106(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-self-compare/rule.a")
}
func TestProduct_CycleFrozen_e01db50c66da8a75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-sequences/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e01db50c66da8a75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-sequences/messages.ts", "current", true)
}
func TestCyclesCorpus_e01db50c66da8a75(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-sequences/messages.ts")
}
func TestProduct_CycleFrozen_c5780aab62b140e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-template-curly-in-string/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c5780aab62b140e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-template-curly-in-string/rule.ts", "current", true)
}
func TestCyclesCorpus_c5780aab62b140e3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-template-curly-in-string/rule.ts")
}
func TestProduct_CycleFrozen_5278cc420aff3939(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-with/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5278cc420aff3939(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-with/rule.ts", "current", true)
}
func TestCyclesCorpus_5278cc420aff3939(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-with/rule.ts")
}
func TestProduct_CycleFrozen_73784df725d81979(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/children.ts", "frozen", true)
}
func TestProduct_CycleCurrent_73784df725d81979(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/children.ts", "current", true)
}
func TestCyclesCorpus_73784df725d81979(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/children.ts")
}
func TestProduct_CycleFrozen_4bd1f62635cb949d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/diagnostic.a", "frozen", true)
}
func TestProduct_CycleCurrent_4bd1f62635cb949d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/diagnostic.a", "current", true)
}
func TestCyclesCorpus_4bd1f62635cb949d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/diagnostic.a")
}
func TestProduct_CycleFrozen_b0ebc3722d4aaa21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_is_mounted.a", "frozen", true)
}
func TestProduct_CycleCurrent_b0ebc3722d4aaa21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_is_mounted.a", "current", true)
}
func TestCyclesCorpus_b0ebc3722d4aaa21(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_is_mounted.a")
}
func TestProduct_CycleFrozen_532a9feac7fd0c18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_532a9feac7fd0c18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/main.ts", "current", true)
}
func TestCyclesCorpus_532a9feac7fd0c18(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/selector/main.ts")
}
func TestProduct_CycleFrozen_7d3d2b698bb73024(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/source.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7d3d2b698bb73024(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/source.ts", "current", true)
}
func TestCyclesCorpus_7d3d2b698bb73024(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/selector/source.ts")
}
func TestProduct_CycleFrozen_e92a0696320cfb48(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/verify.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e92a0696320cfb48(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/verify.ts", "current", true)
}
func TestCyclesCorpus_e92a0696320cfb48(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/verify.ts")
}
func TestProduct_CycleFrozen_c470d0d66407ee12(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/blockLine.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c470d0d66407ee12(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/blockLine.ts", "current", true)
}
func TestCyclesCorpus_c470d0d66407ee12(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/blockLine.ts")
}
func TestProduct_CycleFrozen_e8e1b28319a63e1b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/layoutCommand.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e8e1b28319a63e1b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/layoutCommand.ts", "current", true)
}
func TestCyclesCorpus_e8e1b28319a63e1b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/layoutCommand.ts")
}
func TestProduct_CycleFrozen_30c64a39763384e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_30c64a39763384e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/main.ts", "current", true)
}
func TestCyclesCorpus_30c64a39763384e0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/scanner/main.ts")
}
func TestProduct_CycleFrozen_d48690ce9f865eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_coalesce.a", "frozen", true)
}
func TestProduct_CycleCurrent_d48690ce9f865eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_coalesce.a", "current", true)
}
func TestCyclesCorpus_d48690ce9f865eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/047cb0d_n_coalesce.a")
}
func TestProduct_CycleFrozen_87b3a5d9ce1527c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/indexing.a", "frozen", true)
}
func TestProduct_CycleCurrent_87b3a5d9ce1527c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/indexing.a", "current", true)
}
func TestCyclesCorpus_87b3a5d9ce1527c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_refused/indexing.a")
}
func TestProduct_CycleFrozen_052f779e962d591c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_arrow_thenable.a", "frozen", true)
}
func TestProduct_CycleCurrent_052f779e962d591c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_arrow_thenable.a", "current", true)
}
func TestCyclesCorpus_052f779e962d591c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_refused/async_refuse_arrow_thenable.a")
}
func TestProduct_CycleFrozen_695d721533d90311(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/booleans.a", "frozen", true)
}
func TestProduct_CycleCurrent_695d721533d90311(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/booleans.a", "current", true)
}
func TestCyclesCorpus_695d721533d90311(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/booleans.a")
}
func TestProduct_CycleFrozen_08b49d1cffbb65e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_subunion_string_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_08b49d1cffbb65e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_subunion_string_fails.a", "current", true)
}
func TestCyclesCorpus_08b49d1cffbb65e9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_subunion_string_fails.a")
}
func TestProduct_CycleFrozen_f3aa627f92107d80(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_unary_truthiness.a", "frozen", true)
}
func TestProduct_CycleCurrent_f3aa627f92107d80(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_unary_truthiness.a", "current", true)
}
func TestCyclesCorpus_f3aa627f92107d80(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_unary_truthiness.a")
}
func TestProduct_CycleFrozen_47a36b6e10dcf4e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_accessors.a", "frozen", true)
}
func TestProduct_CycleCurrent_47a36b6e10dcf4e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_accessors.a", "current", true)
}
func TestCyclesCorpus_47a36b6e10dcf4e4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_features_accessors.a")
}
func TestProduct_CycleFrozen_02dc6e6123168627(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_getter.a", "frozen", true)
}
func TestProduct_CycleCurrent_02dc6e6123168627(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_getter.a", "current", true)
}
func TestCyclesCorpus_02dc6e6123168627(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_getter.a")
}
func TestProduct_CycleFrozen_3053650e56b38045(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/large.a", "frozen", true)
}
func TestProduct_CycleCurrent_3053650e56b38045(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/large.a", "current", true)
}
func TestCyclesCorpus_3053650e56b38045(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/large.a")
}
func TestProduct_CycleFrozen_84b7e1cdbce3b521(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_receivers.a", "frozen", true)
}
func TestProduct_CycleCurrent_84b7e1cdbce3b521(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_receivers.a", "current", true)
}
func TestCyclesCorpus_84b7e1cdbce3b521(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_receivers.a")
}
func TestProduct_CycleFrozen_d6f298dcf087de30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a", "frozen", true)
}
func TestProduct_CycleCurrent_d6f298dcf087de30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a", "current", true)
}
func TestCyclesCorpus_d6f298dcf087de30(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/e4eec87_u01_undefined_field_widened.a")
}
func TestProduct_CycleFrozen_bb2c4b3f3981d119(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_captured_parameter.a", "frozen", true)
}
func TestProduct_CycleCurrent_bb2c4b3f3981d119(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_captured_parameter.a", "current", true)
}
func TestCyclesCorpus_bb2c4b3f3981d119(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_captured_parameter.a")
}
func TestProduct_CycleFrozen_73c477f4b4f67eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_sort.a", "frozen", true)
}
func TestProduct_CycleCurrent_73c477f4b4f67eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_sort.a", "current", true)
}
func TestCyclesCorpus_73c477f4b4f67eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/alias_sort.a")
}
func TestProduct_CycleFrozen_deaaa4a4f36c016a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/escaped_before.a", "frozen", true)
}
func TestProduct_CycleCurrent_deaaa4a4f36c016a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/escaped_before.a", "current", true)
}
func TestCyclesCorpus_deaaa4a4f36c016a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/escaped_before.a")
}
func TestProduct_CycleFrozen_197e46bdde9f6608(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/spread_old.a", "frozen", true)
}
func TestProduct_CycleCurrent_197e46bdde9f6608(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/spread_old.a", "current", true)
}
func TestCyclesCorpus_197e46bdde9f6608(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/spread_old.a")
}
func TestProduct_CycleFrozen_a7ad814f86297414(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/gaps.a", "frozen", true)
}
func TestProduct_CycleCurrent_a7ad814f86297414(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/gaps.a", "current", true)
}
func TestCyclesCorpus_a7ad814f86297414(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/gaps.a")
}
func TestProduct_CycleFrozen_cd8c3356f6f94501(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_counted_container.a", "frozen", true)
}
func TestProduct_CycleCurrent_cd8c3356f6f94501(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_counted_container.a", "current", true)
}
func TestCyclesCorpus_cd8c3356f6f94501(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_counted_container.a")
}
func TestProduct_CycleFrozen_b5ebb230f604119e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/b.a", "frozen", true)
}
func TestProduct_CycleCurrent_b5ebb230f604119e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/b.a", "current", true)
}
func TestCyclesCorpus_b5ebb230f604119e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/classes/b.a")
}
func TestProduct_CycleFrozen_d53059959d1cefac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_fnexpr_store.a", "frozen", true)
}
func TestProduct_CycleCurrent_d53059959d1cefac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_fnexpr_store.a", "current", true)
}
func TestCyclesCorpus_d53059959d1cefac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_fnexpr_store.a")
}
func TestProduct_CycleFrozen_b26ed44bd45e37b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_groupby_keys.a", "frozen", true)
}
func TestProduct_CycleCurrent_b26ed44bd45e37b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_groupby_keys.a", "current", true)
}
func TestCyclesCorpus_b26ed44bd45e37b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_groupby_keys.a")
}
func TestProduct_CycleFrozen_e1c05c045ca952b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/long_unicode_literals.a", "frozen", true)
}
func TestProduct_CycleCurrent_e1c05c045ca952b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/long_unicode_literals.a", "current", true)
}
func TestCyclesCorpus_e1c05c045ca952b7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/long_unicode_literals.a")
}
func TestProduct_CycleFrozen_5715fce708e5a459(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/accepted/cleanup.a", "frozen", true)
}
func TestProduct_CycleCurrent_5715fce708e5a459(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/accepted/cleanup.a", "current", true)
}
func TestCyclesCorpus_5715fce708e5a459(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/accepted/cleanup.a")
}
func TestProduct_CycleFrozen_ef824530739f75c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_returned.a", "frozen", true)
}
func TestProduct_CycleCurrent_ef824530739f75c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_returned.a", "current", true)
}
func TestCyclesCorpus_ef824530739f75c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_returned.a")
}
func TestProduct_CycleFrozen_9838856fdbdcce86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_expression.a", "frozen", true)
}
func TestProduct_CycleCurrent_9838856fdbdcce86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_expression.a", "current", true)
}
func TestCyclesCorpus_9838856fdbdcce86(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_expression.a")
}
func TestProduct_CycleFrozen_44e69f13dcd75e2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_map_entry.ts", "frozen", true)
}
func TestProduct_CycleCurrent_44e69f13dcd75e2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_map_entry.ts", "current", true)
}
func TestCyclesCorpus_44e69f13dcd75e2a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_map_entry.ts")
}
func TestProduct_CycleFrozen_993987de53239fd9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_defaults.a", "frozen", true)
}
func TestProduct_CycleCurrent_993987de53239fd9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_defaults.a", "current", true)
}
func TestCyclesCorpus_993987de53239fd9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/omitted_defaults.a")
}
func TestProduct_CycleFrozen_d5b60fa17094e1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_rm.a", "frozen", true)
}
func TestProduct_CycleCurrent_d5b60fa17094e1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_rm.a", "current", true)
}
func TestCyclesCorpus_d5b60fa17094e1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_rm.a")
}
func TestProduct_CycleFrozen_ba570183cb9c57e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p14_overload_optional_chain_container.a", "frozen", true)
}
func TestProduct_CycleCurrent_ba570183cb9c57e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p14_overload_optional_chain_container.a", "current", true)
}
func TestCyclesCorpus_ba570183cb9c57e0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p14_overload_optional_chain_container.a")
}
func TestProduct_CycleFrozen_b26d74a2499c3c8c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_coalesce.a", "frozen", true)
}
func TestProduct_CycleCurrent_b26d74a2499c3c8c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_coalesce.a", "current", true)
}
func TestCyclesCorpus_b26d74a2499c3c8c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_coalesce.a")
}
func TestProduct_CycleFrozen_8afad72a09c5f842(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_octal.a", "frozen", true)
}
func TestProduct_CycleCurrent_8afad72a09c5f842(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_octal.a", "current", true)
}
func TestCyclesCorpus_8afad72a09c5f842(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_octal.a")
}
func TestProduct_CycleFrozen_dac6b4e61fa218cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_optional.a", "frozen", true)
}
func TestProduct_CycleCurrent_dac6b4e61fa218cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_optional.a", "current", true)
}
func TestCyclesCorpus_dac6b4e61fa218cc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_structural_optional.a")
}
func TestProduct_CycleFrozen_2b98d5ed634148ba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_t5.a", "frozen", true)
}
func TestProduct_CycleCurrent_2b98d5ed634148ba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_t5.a", "current", true)
}
func TestCyclesCorpus_2b98d5ed634148ba(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_t5.a")
}
func TestProduct_CycleFrozen_774aa776b66d5bdd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_super_closure.a", "frozen", true)
}
func TestProduct_CycleCurrent_774aa776b66d5bdd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_super_closure.a", "current", true)
}
func TestCyclesCorpus_774aa776b66d5bdd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_super_closure.a")
}
func TestProduct_CycleFrozen_ed21647c8af73cf8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_normalize_long.a", "frozen", true)
}
func TestProduct_CycleCurrent_ed21647c8af73cf8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_normalize_long.a", "current", true)
}
func TestCyclesCorpus_ed21647c8af73cf8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_normalize_long.a")
}
func TestProduct_CycleFrozen_2f901e83c9978340(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_u01_undefined_field_widened.a", "frozen", true)
}
func TestProduct_CycleCurrent_2f901e83c9978340(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_u01_undefined_field_widened.a", "current", true)
}
func TestCyclesCorpus_2f901e83c9978340(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_u01_undefined_field_widened.a")
}
func TestProduct_CycleFrozen_fc81f98e16217da1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p01_maybe_boolean_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_fc81f98e16217da1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p01_maybe_boolean_spread.a", "current", true)
}
func TestCyclesCorpus_fc81f98e16217da1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p01_maybe_boolean_spread.a")
}
func TestProduct_CycleFrozen_12fd8a93ad9da466(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a", "frozen", true)
}
func TestProduct_CycleCurrent_12fd8a93ad9da466(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a", "current", true)
}
func TestCyclesCorpus_12fd8a93ad9da466(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p20_callable_member_call_good.a")
}
func TestProduct_CycleFrozen_f3e6fb2b7a0e7a60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p35_class_in_union.a", "frozen", true)
}
func TestProduct_CycleCurrent_f3e6fb2b7a0e7a60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p35_class_in_union.a", "current", true)
}
func TestCyclesCorpus_f3e6fb2b7a0e7a60(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p35_class_in_union.a")
}
func TestProduct_CycleFrozen_650703dee14f5f6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_closure.a", "frozen", true)
}
func TestProduct_CycleCurrent_650703dee14f5f6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_closure.a", "current", true)
}
func TestCyclesCorpus_650703dee14f5f6c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_closure.a")
}
func TestProduct_CycleFrozen_dad110e88303bf19(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_object_union.a", "frozen", true)
}
func TestProduct_CycleCurrent_dad110e88303bf19(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_object_union.a", "current", true)
}
func TestCyclesCorpus_dad110e88303bf19(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_object_union.a")
}
func TestProduct_CycleFrozen_f8b387cfa7423b28(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p70_length_view_over_tuple.a", "frozen", true)
}
func TestProduct_CycleCurrent_f8b387cfa7423b28(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p70_length_view_over_tuple.a", "current", true)
}
func TestCyclesCorpus_f8b387cfa7423b28(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p70_length_view_over_tuple.a")
}
func TestProduct_CycleFrozen_87a4bd855baac0d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_9984394_lib_dispatch.a", "frozen", true)
}
func TestProduct_CycleCurrent_87a4bd855baac0d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_9984394_lib_dispatch.a", "current", true)
}
func TestCyclesCorpus_87a4bd855baac0d2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_9984394_lib_dispatch.a")
}
func TestProduct_CycleFrozen_b670af44793866ca(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_override_optional.a", "frozen", true)
}
func TestProduct_CycleCurrent_b670af44793866ca(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_override_optional.a", "current", true)
}
func TestCyclesCorpus_b670af44793866ca(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_override_optional.a")
}
func TestProduct_CycleFrozen_650c96a439070a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_p06_helper_second_parameter_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_650c96a439070a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_p06_helper_second_parameter_undefined.a", "current", true)
}
func TestCyclesCorpus_650c96a439070a67(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_p06_helper_second_parameter_undefined.a")
}
func TestProduct_CycleFrozen_2fcf9959fabbf8ee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_recursive.a", "frozen", true)
}
func TestProduct_CycleCurrent_2fcf9959fabbf8ee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_recursive.a", "current", true)
}
func TestCyclesCorpus_2fcf9959fabbf8ee(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/route_targets_recursive.a")
}
func TestProduct_CycleFrozen_1e2cdf231b5fc1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_fnexpr.a", "frozen", true)
}
func TestProduct_CycleCurrent_1e2cdf231b5fc1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_fnexpr.a", "current", true)
}
func TestCyclesCorpus_1e2cdf231b5fc1f4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_fnexpr.a")
}
func TestProduct_CycleFrozen_252a31e683221fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/sort_top_level.a", "frozen", true)
}
func TestProduct_CycleCurrent_252a31e683221fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/sort_top_level.a", "current", true)
}
func TestCyclesCorpus_252a31e683221fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/sort_top_level.a")
}
func TestProduct_CycleFrozen_42afd390b0cc2057(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_stopped/template.a", "frozen", true)
}
func TestProduct_CycleCurrent_42afd390b0cc2057(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_stopped/template.a", "current", true)
}
func TestCyclesCorpus_42afd390b0cc2057(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/statements_small_stopped/template.a")
}
func TestProduct_CycleFrozen_43cc5cd2dc6bd8f9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_comma.a", "frozen", true)
}
func TestProduct_CycleCurrent_43cc5cd2dc6bd8f9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_comma.a", "current", true)
}
func TestCyclesCorpus_43cc5cd2dc6bd8f9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_comma.a")
}
func TestProduct_CycleFrozen_beb64f65d46964b5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/bridge.a", "frozen", true)
}
func TestProduct_CycleCurrent_beb64f65d46964b5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/bridge.a", "current", true)
}
func TestCyclesCorpus_beb64f65d46964b5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_exports/bridge.a")
}
func TestProduct_CycleFrozen_4ff9cf2ef0203a48(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_string_literal.a", "frozen", true)
}
func TestProduct_CycleCurrent_4ff9cf2ef0203a48(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_string_literal.a", "current", true)
}
func TestCyclesCorpus_4ff9cf2ef0203a48(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typeof_string_literal.a")
}
func TestProduct_CycleFrozen_6492b9f4eab904d0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/tree.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6492b9f4eab904d0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/tree.ts", "current", true)
}
func TestCyclesCorpus_6492b9f4eab904d0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/tree.ts")
}
func TestProduct_CycleFrozen_feb224c5d3ca7625(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/parser.ts", "frozen", true)
}
func TestProduct_CycleCurrent_feb224c5d3ca7625(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/parser.ts", "current", true)
}
func TestCyclesCorpus_feb224c5d3ca7625(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/json/parser.ts")
}
func TestProduct_CycleFrozen_dfcaebaeb321caaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/comment_ranges.a", "frozen", true)
}
func TestProduct_CycleCurrent_dfcaebaeb321caaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/comment_ranges.a", "current", true)
}
func TestCyclesCorpus_dfcaebaeb321caaa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/comment_ranges.a")
}
func TestProduct_CycleFrozen_6dc11d5c6e4acc91(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/gaps/4_last_index_position.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6dc11d5c6e4acc91(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/gaps/4_last_index_position.ts", "current", true)
}
func TestCyclesCorpus_6dc11d5c6e4acc91(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/gaps/4_last_index_position.ts")
}
func TestProduct_CycleFrozen_8582ad44ebb2df24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/comment.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8582ad44ebb2df24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/comment.ts", "current", true)
}
func TestCyclesCorpus_8582ad44ebb2df24(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/comments/comment.ts")
}
func TestProduct_CycleFrozen_a10334a26b5319f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/007_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_a10334a26b5319f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/007_runtime.a", "current", true)
}
func TestCyclesCorpus_a10334a26b5319f7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/007_runtime.a")
}
func TestProduct_CycleFrozen_e76d1b09775ffd61(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005.a", "frozen", true)
}
func TestProduct_CycleCurrent_e76d1b09775ffd61(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005.a", "current", true)
}
func TestCyclesCorpus_e76d1b09775ffd61(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005.a")
}
func TestProduct_CycleFrozen_1406759f034497fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/026_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_1406759f034497fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/026_runtime.a", "current", true)
}
func TestCyclesCorpus_1406759f034497fe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/026_runtime.a")
}
func TestProduct_CycleFrozen_aef91fd1363ecdd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/040_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_aef91fd1363ecdd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/040_runtime.a", "current", true)
}
func TestCyclesCorpus_aef91fd1363ecdd3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/040_runtime.a")
}
func TestProduct_CycleFrozen_40be9a90223072a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056.a", "frozen", true)
}
func TestProduct_CycleCurrent_40be9a90223072a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056.a", "current", true)
}
func TestCyclesCorpus_40be9a90223072a6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056.a")
}
func TestProduct_CycleFrozen_2b255cd23173d941(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070.a", "frozen", true)
}
func TestProduct_CycleCurrent_2b255cd23173d941(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070.a", "current", true)
}
func TestCyclesCorpus_2b255cd23173d941(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070.a")
}
func TestProduct_CycleFrozen_2d92f0f31146d3c0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/084_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_2d92f0f31146d3c0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/084_runtime.a", "current", true)
}
func TestCyclesCorpus_2d92f0f31146d3c0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/084_runtime.a")
}
func TestProduct_CycleFrozen_0046512b7251e92f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-nested-callbacks/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_0046512b7251e92f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-nested-callbacks/messages.a", "current", true)
}
func TestCyclesCorpus_0046512b7251e92f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/max-nested-callbacks/messages.a")
}
func TestProduct_CycleFrozen_be4d2a34eda5535d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_be4d2a34eda5535d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/rule.a", "current", true)
}
func TestCyclesCorpus_be4d2a34eda5535d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-ambiguous-identifier/rule.a")
}
func TestProduct_CycleFrozen_0686e6733e407bfe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-div-regex/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0686e6733e407bfe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-div-regex/rule.ts", "current", true)
}
func TestCyclesCorpus_0686e6733e407bfe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-div-regex/rule.ts")
}
func TestProduct_CycleFrozen_db920ac500834cfd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-concat/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_db920ac500834cfd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-concat/messages.a", "current", true)
}
func TestCyclesCorpus_db920ac500834cfd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-useless-concat/messages.a")
}
func TestProduct_CycleFrozen_3db42d59b21b17f9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/16_array_shift.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3db42d59b21b17f9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/16_array_shift.ts", "current", true)
}
func TestCyclesCorpus_3db42d59b21b17f9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/16_array_shift.ts")
}
func TestProduct_CycleFrozen_2fa7436c912699b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/1_recursive_state.ts", "frozen", true)
}
func TestProduct_CycleCurrent_2fa7436c912699b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/1_recursive_state.ts", "current", true)
}
func TestCyclesCorpus_2fa7436c912699b9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/1_recursive_state.ts")
}
func TestProduct_CycleFrozen_e446e26100290f30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/width.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e446e26100290f30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/width.ts", "current", true)
}
func TestCyclesCorpus_e446e26100290f30(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/width.ts")
}
func TestProduct_CycleFrozen_811bacc6bbe980e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/gaps/5_local_children_array.ts", "frozen", true)
}
func TestProduct_CycleCurrent_811bacc6bbe980e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/gaps/5_local_children_array.ts", "current", true)
}
func TestCyclesCorpus_811bacc6bbe980e6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/mediaquery/gaps/5_local_children_array.ts")
}
func TestProduct_CycleFrozen_ca7dd3e56d003285(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/types.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ca7dd3e56d003285(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/types.ts", "current", true)
}
func TestCyclesCorpus_ca7dd3e56d003285(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/types.ts")
}
func TestProduct_CycleFrozen_6a7338229c399862(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/unbound.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6a7338229c399862(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/unbound.ts", "current", true)
}
func TestCyclesCorpus_6a7338229c399862(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/unbound.ts")
}
func TestProduct_CycleFrozen_dcdbd249166bf770(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/compose_main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_dcdbd249166bf770(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/compose_main.ts", "current", true)
}
func TestCyclesCorpus_dcdbd249166bf770(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/compose_main.ts")
}
func TestProduct_CycleFrozen_06eeb78c1de1c37f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/printer.ts", "frozen", true)
}
func TestProduct_CycleCurrent_06eeb78c1de1c37f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/printer.ts", "current", true)
}
func TestCyclesCorpus_06eeb78c1de1c37f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/printer.ts")
}
func TestProduct_CycleFrozen_bf16ba697b3a9aba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistContext.ts", "frozen", true)
}
func TestProduct_CycleCurrent_bf16ba697b3a9aba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistContext.ts", "current", true)
}
func TestCyclesCorpus_bf16ba697b3a9aba(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/unistContext.ts")
}
func TestProduct_CycleFrozen_ebf5b7b293d3d3be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/grammar.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ebf5b7b293d3d3be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/grammar.ts", "current", true)
}
func TestCyclesCorpus_ebf5b7b293d3d3be(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/parser/grammar.ts")
}
func TestProduct_CycleFrozen_6d3726badd94a2a1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_parameters.a", "frozen", true)
}
func TestProduct_CycleCurrent_6d3726badd94a2a1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_parameters.a", "current", true)
}
func TestCyclesCorpus_6d3726badd94a2a1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_coverage_parameters.a")
}
func TestProduct_CycleFrozen_2c8b120ffb440c78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_reject_eager.a", "frozen", true)
}
func TestProduct_CycleCurrent_2c8b120ffb440c78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_reject_eager.a", "current", true)
}
func TestCyclesCorpus_2c8b120ffb440c78(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_coverage_reject_eager.a")
}
func TestProduct_CycleFrozen_4eb75c08f8d59455(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_return_thenable.a", "frozen", true)
}
func TestProduct_CycleCurrent_4eb75c08f8d59455(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_return_thenable.a", "current", true)
}
func TestCyclesCorpus_4eb75c08f8d59455(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_refused/async_refuse_return_thenable.a")
}
func TestProduct_CycleFrozen_4d6ba89e7e347b21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_map_overwrite.a", "frozen", true)
}
func TestProduct_CycleCurrent_4d6ba89e7e347b21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_map_overwrite.a", "current", true)
}
func TestCyclesCorpus_4d6ba89e7e347b21(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_map_overwrite.a")
}
func TestProduct_CycleFrozen_1f8b1168f36d5372(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_class_union_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_1f8b1168f36d5372(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_class_union_fails.a", "current", true)
}
func TestCyclesCorpus_1f8b1168f36d5372(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_class_union_fails.a")
}
func TestProduct_CycleFrozen_a7272f8cf8c76c7e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance_conditional.a", "frozen", true)
}
func TestProduct_CycleCurrent_a7272f8cf8c76c7e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance_conditional.a", "current", true)
}
func TestCyclesCorpus_a7272f8cf8c76c7e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_inheritance_conditional.a")
}
func TestProduct_CycleFrozen_f7938bec0bc65035(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_private_members.a", "frozen", true)
}
func TestProduct_CycleCurrent_f7938bec0bc65035(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_private_members.a", "current", true)
}
func TestCyclesCorpus_f7938bec0bc65035(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_private_members.a")
}
func TestProduct_CycleFrozen_5ffea53183325fe0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a", "frozen", true)
}
func TestProduct_CycleCurrent_5ffea53183325fe0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a", "current", true)
}
func TestCyclesCorpus_5ffea53183325fe0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_sym_keys_view.a")
}
func TestProduct_CycleFrozen_65db1f1ea0b480d6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_receiver_rest.a", "frozen", true)
}
func TestProduct_CycleCurrent_65db1f1ea0b480d6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_receiver_rest.a", "current", true)
}
func TestCyclesCorpus_65db1f1ea0b480d6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/closure_convention_receiver_rest.a")
}
func TestProduct_CycleFrozen_b7bcd1516fc8932c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/fresh.a", "frozen", true)
}
func TestProduct_CycleCurrent_b7bcd1516fc8932c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/fresh.a", "current", true)
}
func TestCyclesCorpus_b7bcd1516fc8932c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/fresh.a")
}
func TestProduct_CycleFrozen_4b92d7eb7906d361(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/hidden_subclass.a", "frozen", true)
}
func TestProduct_CycleCurrent_4b92d7eb7906d361(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/hidden_subclass.a", "current", true)
}
func TestCyclesCorpus_4b92d7eb7906d361(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/hidden_subclass.a")
}
func TestProduct_CycleFrozen_b90beb0ed872c25a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_array.a", "frozen", true)
}
func TestProduct_CycleCurrent_b90beb0ed872c25a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_array.a", "current", true)
}
func TestCyclesCorpus_b90beb0ed872c25a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_view_array.a")
}
func TestProduct_CycleFrozen_ad773b015147f39a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_closure.a", "frozen", true)
}
func TestProduct_CycleCurrent_ad773b015147f39a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_closure.a", "current", true)
}
func TestCyclesCorpus_ad773b015147f39a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_view_closure.a")
}
func TestProduct_CycleFrozen_99698f4598a882de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_callback_exits.a", "frozen", true)
}
func TestProduct_CycleCurrent_99698f4598a882de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_callback_exits.a", "current", true)
}
func TestCyclesCorpus_99698f4598a882de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_callback_exits.a")
}
func TestProduct_CycleFrozen_c8233e92851df221(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_keeping_call.a", "frozen", true)
}
func TestProduct_CycleCurrent_c8233e92851df221(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_keeping_call.a", "current", true)
}
func TestCyclesCorpus_c8233e92851df221(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_keeping_call.a")
}
func TestProduct_CycleFrozen_0b257bc56aebbf66(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/array_from.a", "frozen", true)
}
func TestProduct_CycleCurrent_0b257bc56aebbf66(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/array_from.a", "current", true)
}
func TestCyclesCorpus_0b257bc56aebbf66(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/array_from.a")
}
func TestProduct_CycleFrozen_758fa2e3a3dd0a58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/map_entries_pattern.a", "frozen", true)
}
func TestProduct_CycleCurrent_758fa2e3a3dd0a58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/map_entries_pattern.a", "current", true)
}
func TestCyclesCorpus_758fa2e3a3dd0a58(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/map_entries_pattern.a")
}
func TestProduct_CycleFrozen_728b480af9c97589(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_728b480af9c97589(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_return.a", "current", true)
}
func TestCyclesCorpus_728b480af9c97589(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/function_values_return.a")
}
func TestProduct_CycleFrozen_180d9f072e806c5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_binary_inline.a", "frozen", true)
}
func TestProduct_CycleCurrent_180d9f072e806c5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_binary_inline.a", "current", true)
}
func TestCyclesCorpus_180d9f072e806c5a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_binary_inline.a")
}
func TestProduct_CycleFrozen_f3d8d1e61721c441(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_update_node_function.a", "frozen", true)
}
func TestProduct_CycleCurrent_f3d8d1e61721c441(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_update_node_function.a", "current", true)
}
func TestCyclesCorpus_f3d8d1e61721c441(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_update_node_function.a")
}
func TestProduct_CycleFrozen_f4000719f50379e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_anchor.a", "frozen", true)
}
func TestProduct_CycleCurrent_f4000719f50379e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_anchor.a", "current", true)
}
func TestCyclesCorpus_f4000719f50379e8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_anchor.a")
}
func TestProduct_CycleFrozen_ac2fe1a7c6af6c62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_closure.a", "frozen", true)
}
func TestProduct_CycleCurrent_ac2fe1a7c6af6c62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_closure.a", "current", true)
}
func TestCyclesCorpus_ac2fe1a7c6af6c62(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_closure.a")
}
func TestProduct_CycleFrozen_83ee0b8df52f7a36(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_never_branches.a", "frozen", true)
}
func TestProduct_CycleCurrent_83ee0b8df52f7a36(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_never_branches.a", "current", true)
}
func TestCyclesCorpus_83ee0b8df52f7a36(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/host_never_branches.a")
}
func TestProduct_CycleFrozen_bdc481e1eb7bf341(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/c.a", "frozen", true)
}
func TestProduct_CycleCurrent_bdc481e1eb7bf341(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/c.a", "current", true)
}
func TestCyclesCorpus_bdc481e1eb7bf341(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/classes/c.a")
}
func TestProduct_CycleFrozen_b57d5592fdf9386e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/hoisted/a.a", "frozen", true)
}
func TestProduct_CycleCurrent_b57d5592fdf9386e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/hoisted/a.a", "current", true)
}
func TestCyclesCorpus_b57d5592fdf9386e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/hoisted/a.a")
}
func TestProduct_CycleFrozen_e61297b2819e1594(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/order/shared.a", "frozen", true)
}
func TestProduct_CycleCurrent_e61297b2819e1594(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/order/shared.a", "current", true)
}
func TestCyclesCorpus_e61297b2819e1594(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/order/shared.a")
}
func TestProduct_CycleFrozen_3b101b43d32df86d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/runtime/odd.a", "frozen", true)
}
func TestProduct_CycleCurrent_3b101b43d32df86d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/runtime/odd.a", "current", true)
}
func TestCyclesCorpus_3b101b43d32df86d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/runtime/odd.a")
}
func TestProduct_CycleFrozen_7bceeb6b7bd3ca5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_scalars.a", "frozen", true)
}
func TestProduct_CycleCurrent_7bceeb6b7bd3ca5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_scalars.a", "current", true)
}
func TestCyclesCorpus_7bceeb6b7bd3ca5a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_scalars.a")
}
func TestProduct_CycleFrozen_e647262fff973830(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_keys_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_e647262fff973830(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_keys_spread.a", "current", true)
}
func TestCyclesCorpus_e647262fff973830(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_keys_spread.a")
}
func TestProduct_CycleFrozen_6939cf36b4ba9db5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_values_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_6939cf36b4ba9db5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_values_assign.a", "current", true)
}
func TestCyclesCorpus_6939cf36b4ba9db5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_values_assign.a")
}
func TestProduct_CycleFrozen_7475b968bfd02970(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_closures.a", "frozen", true)
}
func TestProduct_CycleCurrent_7475b968bfd02970(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_closures.a", "current", true)
}
func TestCyclesCorpus_7475b968bfd02970(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_closures.a")
}
func TestProduct_CycleFrozen_20d712fe5e8119e1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_named_keys.a", "frozen", true)
}
func TestProduct_CycleCurrent_20d712fe5e8119e1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_named_keys.a", "current", true)
}
func TestCyclesCorpus_20d712fe5e8119e1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_named_keys.a")
}
func TestProduct_CycleFrozen_5d5b5829d3739d43(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_alias_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_5d5b5829d3739d43(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_alias_order.a", "current", true)
}
func TestCyclesCorpus_5d5b5829d3739d43(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_alias_order.a")
}
func TestProduct_CycleFrozen_b30d5aeabec41c96(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_lastindexof.a", "frozen", true)
}
func TestProduct_CycleCurrent_b30d5aeabec41c96(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_lastindexof.a", "current", true)
}
func TestCyclesCorpus_b30d5aeabec41c96(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_lastindexof.a")
}
func TestProduct_CycleFrozen_cb41025edab65e05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/direct_provider.a", "frozen", true)
}
func TestProduct_CycleCurrent_cb41025edab65e05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/direct_provider.a", "current", true)
}
func TestCyclesCorpus_cb41025edab65e05(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/module_namespace_reads/direct_provider.a")
}
func TestProduct_CycleFrozen_adfffefd9f54e90c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/helper.a", "frozen", true)
}
func TestProduct_CycleCurrent_adfffefd9f54e90c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/helper.a", "current", true)
}
func TestCyclesCorpus_adfffefd9f54e90c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/helper.a")
}
func TestProduct_CycleFrozen_690edf6463267583(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/item_destructure.a", "frozen", true)
}
func TestProduct_CycleCurrent_690edf6463267583(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/item_destructure.a", "current", true)
}
func TestCyclesCorpus_690edf6463267583(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/item_destructure.a")
}
func TestProduct_CycleFrozen_cbfcc1a6f0745909(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nbody_static_collision.a", "frozen", true)
}
func TestProduct_CycleCurrent_cbfcc1a6f0745909(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nbody_static_collision.a", "current", true)
}
func TestCyclesCorpus_cbfcc1a6f0745909(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nbody_static_collision.a")
}
func TestProduct_CycleFrozen_0e949d95088ec899(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_tdz.a", "frozen", true)
}
func TestProduct_CycleCurrent_0e949d95088ec899(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_tdz.a", "current", true)
}
func TestCyclesCorpus_0e949d95088ec899(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_tdz.a")
}
func TestProduct_CycleFrozen_792eed5e9d76e2c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_tdz_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_792eed5e9d76e2c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_tdz_write.a", "current", true)
}
func TestCyclesCorpus_792eed5e9d76e2c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_tdz_write.a")
}
func TestProduct_CycleFrozen_55978524d407d5ee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_comparison.a", "frozen", true)
}
func TestProduct_CycleCurrent_55978524d407d5ee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_comparison.a", "current", true)
}
func TestCyclesCorpus_55978524d407d5ee(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_comparison.a")
}
func TestProduct_CycleFrozen_9f502dbf8b7257ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_null.a", "frozen", true)
}
func TestProduct_CycleCurrent_9f502dbf8b7257ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_null.a", "current", true)
}
func TestCyclesCorpus_9f502dbf8b7257ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_null.a")
}
func TestProduct_CycleFrozen_c2cde38687cf2411(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_field.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c2cde38687cf2411(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_field.ts", "current", true)
}
func TestCyclesCorpus_c2cde38687cf2411(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_field.ts")
}
func TestProduct_CycleFrozen_31c007f9bc94db37(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/number_parsing.a", "frozen", true)
}
func TestProduct_CycleCurrent_31c007f9bc94db37(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/number_parsing.a", "current", true)
}
func TestCyclesCorpus_31c007f9bc94db37(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/number_parsing.a")
}
func TestProduct_CycleFrozen_c7b095a43a9c1a65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_sort_deduplicate.a", "frozen", true)
}
func TestProduct_CycleCurrent_c7b095a43a9c1a65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_sort_deduplicate.a", "current", true)
}
func TestCyclesCorpus_c7b095a43a9c1a65(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_sort_deduplicate.a")
}
func TestProduct_CycleFrozen_a6d6e49764165a6f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/pad_infinity.a", "frozen", true)
}
func TestProduct_CycleCurrent_a6d6e49764165a6f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/pad_infinity.a", "current", true)
}
func TestCyclesCorpus_a6d6e49764165a6f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/pad_infinity.a")
}
func TestProduct_CycleFrozen_b4813215ef922af5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_b4813215ef922af5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_order.a", "current", true)
}
func TestCyclesCorpus_b4813215ef922af5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/params_namespaces_order.a")
}
func TestProduct_CycleFrozen_9aad1c9ea874a5e7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p12_find_undefined_target.a", "frozen", true)
}
func TestProduct_CycleCurrent_9aad1c9ea874a5e7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p12_find_undefined_target.a", "current", true)
}
func TestCyclesCorpus_9aad1c9ea874a5e7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p12_find_undefined_target.a")
}
func TestProduct_CycleFrozen_d77763dc8ce2c201(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/constructor.a", "frozen", true)
}
func TestProduct_CycleCurrent_d77763dc8ce2c201(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/constructor.a", "current", true)
}
func TestCyclesCorpus_d77763dc8ce2c201(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/program_region/constructor.a")
}
func TestProduct_CycleFrozen_634fab82a5af4f97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reduce_undefined_initial.a", "frozen", true)
}
func TestProduct_CycleCurrent_634fab82a5af4f97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reduce_undefined_initial.a", "current", true)
}
func TestCyclesCorpus_634fab82a5af4f97(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/reduce_undefined_initial.a")
}
func TestProduct_CycleFrozen_d4712f09a209eb22(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_null_narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_d4712f09a209eb22(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_null_narrowed.a", "current", true)
}
func TestCyclesCorpus_d4712f09a209eb22(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_null_narrowed.a")
}
func TestProduct_CycleFrozen_0d857eeabd3ed292(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe_ref.a", "frozen", true)
}
func TestProduct_CycleCurrent_0d857eeabd3ed292(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe_ref.a", "current", true)
}
func TestCyclesCorpus_0d857eeabd3ed292(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe_ref.a")
}
func TestProduct_CycleFrozen_d5b58dbf21ea8d3b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_derived_symbol.a", "frozen", true)
}
func TestProduct_CycleCurrent_d5b58dbf21ea8d3b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_derived_symbol.a", "current", true)
}
func TestCyclesCorpus_d5b58dbf21ea8d3b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_derived_symbol.a")
}
func TestProduct_CycleFrozen_3c5764585091132e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_override_source.a", "frozen", true)
}
func TestProduct_CycleCurrent_3c5764585091132e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_override_source.a", "current", true)
}
func TestCyclesCorpus_3c5764585091132e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_iterators_override_source.a")
}
func TestProduct_CycleFrozen_d47a986ec963c967(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p06_element_access_boolean_wrong.a", "frozen", true)
}
func TestProduct_CycleCurrent_d47a986ec963c967(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p06_element_access_boolean_wrong.a", "current", true)
}
func TestCyclesCorpus_d47a986ec963c967(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p06_element_access_boolean_wrong.a")
}
func TestProduct_CycleFrozen_aaaccbecc0aadab7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p13_filter_wrong_parameter_helper.a", "frozen", true)
}
func TestProduct_CycleCurrent_aaaccbecc0aadab7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p13_filter_wrong_parameter_helper.a", "current", true)
}
func TestCyclesCorpus_aaaccbecc0aadab7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p13_filter_wrong_parameter_helper.a")
}
func TestProduct_CycleFrozen_83cb49b1b7197209(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_accessor.a", "frozen", true)
}
func TestProduct_CycleCurrent_83cb49b1b7197209(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_accessor.a", "current", true)
}
func TestCyclesCorpus_83cb49b1b7197209(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/route_targets_accessor.a")
}
func TestProduct_CycleFrozen_92bd4da36c4f2cd0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/shadowing.a", "frozen", true)
}
func TestProduct_CycleCurrent_92bd4da36c4f2cd0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/shadowing.a", "current", true)
}
func TestCyclesCorpus_92bd4da36c4f2cd0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/shadowing.a")
}
func TestProduct_CycleFrozen_e88d5f37f99d2f60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/erased.a", "frozen", true)
}
func TestProduct_CycleCurrent_e88d5f37f99d2f60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/erased.a", "current", true)
}
func TestCyclesCorpus_e88d5f37f99d2f60(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_exports/erased.a")
}
func TestProduct_CycleFrozen_f307ce0fd554a644(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/enum.a", "frozen", true)
}
func TestProduct_CycleCurrent_f307ce0fd554a644(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/enum.a", "current", true)
}
func TestCyclesCorpus_f307ce0fd554a644(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_star/enum.a")
}
func TestProduct_CycleFrozen_6d3ad4682e3d44aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/types.a", "frozen", true)
}
func TestProduct_CycleCurrent_6d3ad4682e3d44aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/types.a", "current", true)
}
func TestCyclesCorpus_6d3ad4682e3d44aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_star/types.a")
}
func TestProduct_CycleFrozen_1b1d0100c9888798(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/compiler/empty-else-guard/survivor-switch.a", "frozen", true)
}
func TestProduct_CycleCurrent_1b1d0100c9888798(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/compiler/empty-else-guard/survivor-switch.a", "current", true)
}
func TestCyclesCorpus_1b1d0100c9888798(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/compiler/empty-else-guard/survivor-switch.a")
}
func TestProduct_CycleFrozen_14b4c845fe2e846e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/gaps/3_optional_boolean_condition.ts", "frozen", true)
}
func TestProduct_CycleCurrent_14b4c845fe2e846e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/gaps/3_optional_boolean_condition.ts", "current", true)
}
func TestCyclesCorpus_14b4c845fe2e846e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/gaps/3_optional_boolean_condition.ts")
}
func TestProduct_CycleFrozen_a60d9a32d777727c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a60d9a32d777727c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/main.ts", "current", true)
}
func TestCyclesCorpus_a60d9a32d777727c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/main.ts")
}
func TestProduct_CycleFrozen_7eb4726b2301eda0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/typeMemberInitializer.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7eb4726b2301eda0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/typeMemberInitializer.ts", "current", true)
}
func TestCyclesCorpus_7eb4726b2301eda0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/typeMemberInitializer.ts")
}
func TestProduct_CycleFrozen_c2dcf4632bb58889(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/modifiers.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c2dcf4632bb58889(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/modifiers.ts", "current", true)
}
func TestCyclesCorpus_c2dcf4632bb58889(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/modifiers.ts")
}
func TestProduct_CycleFrozen_5178fe02f13a6b49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceModules.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5178fe02f13a6b49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceModules.ts", "current", true)
}
func TestCyclesCorpus_5178fe02f13a6b49(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceModules.ts")
}
func TestProduct_CycleFrozen_0e40ff035252d8f8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/width.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0e40ff035252d8f8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/width.ts", "current", true)
}
func TestCyclesCorpus_0e40ff035252d8f8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/json/width.ts")
}
func TestProduct_CycleFrozen_48c5add63dd9b539(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/patterns.a", "frozen", true)
}
func TestProduct_CycleCurrent_48c5add63dd9b539(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/patterns.a", "current", true)
}
func TestCyclesCorpus_48c5add63dd9b539(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/patterns.a")
}
func TestProduct_CycleFrozen_f361a4d07364a10a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004.a", "frozen", true)
}
func TestProduct_CycleCurrent_f361a4d07364a10a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004.a", "current", true)
}
func TestCyclesCorpus_f361a4d07364a10a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004.a")
}
func TestProduct_CycleFrozen_6a5a07fea2b3d822(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_6a5a07fea2b3d822(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005_runtime.a", "current", true)
}
func TestCyclesCorpus_6a5a07fea2b3d822(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/005_runtime.a")
}
func TestProduct_CycleFrozen_b63eddc8688536d3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_b63eddc8688536d3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056_runtime.a", "current", true)
}
func TestCyclesCorpus_b63eddc8688536d3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/056_runtime.a")
}
func TestProduct_CycleFrozen_1bee82c86bcb6ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057.a", "frozen", true)
}
func TestProduct_CycleCurrent_1bee82c86bcb6ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057.a", "current", true)
}
func TestCyclesCorpus_1bee82c86bcb6ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057.a")
}
func TestProduct_CycleFrozen_aedb854edb4d0be8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_aedb854edb4d0be8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070_runtime.a", "current", true)
}
func TestCyclesCorpus_aedb854edb4d0be8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/070_runtime.a")
}
func TestProduct_CycleFrozen_09c6ef602289d3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071.a", "frozen", true)
}
func TestProduct_CycleCurrent_09c6ef602289d3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071.a", "current", true)
}
func TestCyclesCorpus_09c6ef602289d3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071.a")
}
func TestProduct_CycleFrozen_44dd24e8b66ee389(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-long-line-comment/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_44dd24e8b66ee389(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-long-line-comment/rule.a", "current", true)
}
func TestCyclesCorpus_44dd24e8b66ee389(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-long-line-comment/rule.a")
}
func TestProduct_CycleFrozen_b00542987280f955(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/comment_anchors.a", "frozen", true)
}
func TestProduct_CycleCurrent_b00542987280f955(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/comment_anchors.a", "current", true)
}
func TestCyclesCorpus_b00542987280f955(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/comment_anchors.a")
}
func TestProduct_CycleFrozen_f58fc960f3157af9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-stuttering-name/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_f58fc960f3157af9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-stuttering-name/messages.a", "current", true)
}
func TestCyclesCorpus_f58fc960f3157af9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-stuttering-name/messages.a")
}
func TestProduct_CycleFrozen_0f519130fc62549c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-duplicate-case/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0f519130fc62549c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-duplicate-case/messages.ts", "current", true)
}
func TestCyclesCorpus_0f519130fc62549c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-duplicate-case/messages.ts")
}
func TestProduct_CycleFrozen_cdcf98f1cc59b757(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-template-curly-in-string/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cdcf98f1cc59b757(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-template-curly-in-string/messages.ts", "current", true)
}
func TestCyclesCorpus_cdcf98f1cc59b757(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-template-curly-in-string/messages.ts")
}
func TestProduct_CycleFrozen_fd2b85d85cc75bfc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unsafe-finally/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_fd2b85d85cc75bfc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unsafe-finally/messages.a", "current", true)
}
func TestCyclesCorpus_fd2b85d85cc75bfc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-unsafe-finally/messages.a")
}
func TestProduct_CycleFrozen_ddf97103d4bc08c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-catch/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_ddf97103d4bc08c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-catch/messages.a", "current", true)
}
func TestCyclesCorpus_ddf97103d4bc08c8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-useless-catch/messages.a")
}
func TestProduct_CycleFrozen_bbbb3d4bf468267e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-with/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_bbbb3d4bf468267e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-with/messages.ts", "current", true)
}
func TestCyclesCorpus_bbbb3d4bf468267e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-with/messages.ts")
}
func TestProduct_CycleFrozen_6bdd029799c12191(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/prefer-destructuring/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_6bdd029799c12191(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/prefer-destructuring/rule.a", "current", true)
}
func TestCyclesCorpus_6bdd029799c12191(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/prefer-destructuring/rule.a")
}
func TestProduct_CycleFrozen_875302ef835581c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/textClasses.ts", "frozen", true)
}
func TestProduct_CycleCurrent_875302ef835581c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/textClasses.ts", "current", true)
}
func TestCyclesCorpus_875302ef835581c8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/textClasses.ts")
}
func TestProduct_CycleFrozen_ac7d20cb1a2f931f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/whitespaceCodec.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ac7d20cb1a2f931f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/whitespaceCodec.ts", "current", true)
}
func TestCyclesCorpus_ac7d20cb1a2f931f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/whitespaceCodec.ts")
}
func TestProduct_CycleFrozen_e9ff03e5babac51d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_unused_private_class_members.a", "frozen", true)
}
func TestProduct_CycleCurrent_e9ff03e5babac51d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_unused_private_class_members.a", "current", true)
}
func TestCyclesCorpus_e9ff03e5babac51d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_unused_private_class_members.a")
}
func TestProduct_CycleFrozen_b73dc4c6c580725c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/phi.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b73dc4c6c580725c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/phi.ts", "current", true)
}
func TestCyclesCorpus_b73dc4c6c580725c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/phi.ts")
}
func TestProduct_CycleFrozen_72ea72981f878423(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/scan.ts", "frozen", true)
}
func TestProduct_CycleCurrent_72ea72981f878423(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/scan.ts", "current", true)
}
func TestCyclesCorpus_72ea72981f878423(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/suppression/scan.ts")
}
func TestProduct_CycleFrozen_7a4d103334843503(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/suite.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7a4d103334843503(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/suite.ts", "current", true)
}
func TestCyclesCorpus_7a4d103334843503(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/suite.ts")
}
func TestProduct_CycleFrozen_aace2849aeb7663f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/void.ts", "frozen", true)
}
func TestProduct_CycleCurrent_aace2849aeb7663f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/void.ts", "current", true)
}
func TestCyclesCorpus_aace2849aeb7663f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/void.ts")
}
func TestProduct_CycleFrozen_7d153fdf04842fef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/arenaNodeRead.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7d153fdf04842fef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/arenaNodeRead.ts", "current", true)
}
func TestCyclesCorpus_7d153fdf04842fef(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/gaps/arenaNodeRead.ts")
}
func TestProduct_CycleFrozen_92e89f8dd2c1bd40(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/returning.a", "frozen", true)
}
func TestProduct_CycleCurrent_92e89f8dd2c1bd40(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/returning.a", "current", true)
}
func TestCyclesCorpus_92e89f8dd2c1bd40(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_refused/returning.a")
}
func TestProduct_CycleFrozen_e0226fdd8f5f788b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_unrelated_type.a", "frozen", true)
}
func TestProduct_CycleCurrent_e0226fdd8f5f788b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_unrelated_type.a", "current", true)
}
func TestCyclesCorpus_e0226fdd8f5f788b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_unrelated_type.a")
}
func TestProduct_CycleFrozen_29d7a040a0a88a37(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_reassigned.a", "frozen", true)
}
func TestProduct_CycleCurrent_29d7a040a0a88a37(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_reassigned.a", "current", true)
}
func TestCyclesCorpus_29d7a040a0a88a37(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_reassigned.a")
}
func TestProduct_CycleFrozen_cc884821fb6f9ac0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_string.a", "frozen", true)
}
func TestProduct_CycleCurrent_cc884821fb6f9ac0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_string.a", "current", true)
}
func TestCyclesCorpus_cc884821fb6f9ac0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_enum_string.a")
}
func TestProduct_CycleFrozen_1542ef42f991f961(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_predicate_marker_live.a", "frozen", true)
}
func TestProduct_CycleCurrent_1542ef42f991f961(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_predicate_marker_live.a", "current", true)
}
func TestCyclesCorpus_1542ef42f991f961(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_predicate_marker_live.a")
}
func TestProduct_CycleFrozen_1db9d9d2746d9f39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_unary_numeric.a", "frozen", true)
}
func TestProduct_CycleCurrent_1db9d9d2746d9f39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_unary_numeric.a", "current", true)
}
func TestCyclesCorpus_1db9d9d2746d9f39(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_unary_numeric.a")
}
func TestProduct_CycleFrozen_0dfe3d78a0fe18c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance_memory.a", "frozen", true)
}
func TestProduct_CycleCurrent_0dfe3d78a0fe18c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance_memory.a", "current", true)
}
func TestCyclesCorpus_0dfe3d78a0fe18c8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_inheritance_memory.a")
}
func TestProduct_CycleFrozen_16a6218451dbb67e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a", "frozen", true)
}
func TestProduct_CycleCurrent_16a6218451dbb67e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a", "current", true)
}
func TestCyclesCorpus_16a6218451dbb67e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_override_this.a")
}
func TestProduct_CycleFrozen_c48578ae08db7dce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_boxed.a", "frozen", true)
}
func TestProduct_CycleCurrent_c48578ae08db7dce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_boxed.a", "current", true)
}
func TestCyclesCorpus_c48578ae08db7dce(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_checked_boxed.a")
}
func TestProduct_CycleFrozen_9d32c23b10d70a2d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open.a", "frozen", true)
}
func TestProduct_CycleCurrent_9d32c23b10d70a2d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open.a", "current", true)
}
func TestCyclesCorpus_9d32c23b10d70a2d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_open.a")
}
func TestProduct_CycleFrozen_a60fcbbb1db93e10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_a60fcbbb1db93e10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_return.a", "current", true)
}
func TestCyclesCorpus_a60fcbbb1db93e10(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_open_never_return.a")
}
func TestProduct_CycleFrozen_d83e64bea56b423f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_large.a", "frozen", true)
}
func TestProduct_CycleCurrent_d83e64bea56b423f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_large.a", "current", true)
}
func TestCyclesCorpus_d83e64bea56b423f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_large.a")
}
func TestProduct_CycleFrozen_b84e4429bc94083d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_clobbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_b84e4429bc94083d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_clobbers.a", "current", true)
}
func TestCyclesCorpus_b84e4429bc94083d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_clobbers.a")
}
func TestProduct_CycleFrozen_ba30457b50529a0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_comparator.a", "frozen", true)
}
func TestProduct_CycleCurrent_ba30457b50529a0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_comparator.a", "current", true)
}
func TestCyclesCorpus_ba30457b50529a0f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_comparator.a")
}
func TestProduct_CycleFrozen_88759ee31e6aa654(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/map_set.a", "frozen", true)
}
func TestProduct_CycleCurrent_88759ee31e6aa654(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/map_set.a", "current", true)
}
func TestCyclesCorpus_88759ee31e6aa654(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/map_set.a")
}
func TestProduct_CycleFrozen_94e63aed8501237b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_flow.a", "frozen", true)
}
func TestProduct_CycleCurrent_94e63aed8501237b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_flow.a", "current", true)
}
func TestCyclesCorpus_94e63aed8501237b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_flow.a")
}
func TestProduct_CycleFrozen_a43ea9b6e28e6b53(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_generic_capture.a", "frozen", true)
}
func TestProduct_CycleCurrent_a43ea9b6e28e6b53(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_generic_capture.a", "current", true)
}
func TestCyclesCorpus_a43ea9b6e28e6b53(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_generic_capture.a")
}
func TestProduct_CycleFrozen_836c527c307b1942(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_list.a", "frozen", true)
}
func TestProduct_CycleCurrent_836c527c307b1942(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_list.a", "current", true)
}
func TestCyclesCorpus_836c527c307b1942(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_list.a")
}
func TestProduct_CycleFrozen_5c1b0ee16501885f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_copy.a", "frozen", true)
}
func TestProduct_CycleCurrent_5c1b0ee16501885f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_copy.a", "current", true)
}
func TestCyclesCorpus_5c1b0ee16501885f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_array_copy.a")
}
func TestProduct_CycleFrozen_d8b110e5c9f2c124(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set.a", "frozen", true)
}
func TestProduct_CycleCurrent_d8b110e5c9f2c124(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set.a", "current", true)
}
func TestCyclesCorpus_d8b110e5c9f2c124(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set.a")
}
func TestProduct_CycleFrozen_5c19165879f127fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_visit.a", "frozen", true)
}
func TestProduct_CycleCurrent_5c19165879f127fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_visit.a", "current", true)
}
func TestCyclesCorpus_5c19165879f127fe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_visit.a")
}
func TestProduct_CycleFrozen_65a61f03287b5f42(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/local_console.a", "frozen", true)
}
func TestProduct_CycleCurrent_65a61f03287b5f42(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/local_console.a", "current", true)
}
func TestCyclesCorpus_65a61f03287b5f42(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/local_console.a")
}
func TestProduct_CycleFrozen_4740ec6251604a24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_fnexpr.a", "frozen", true)
}
func TestProduct_CycleCurrent_4740ec6251604a24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_fnexpr.a", "current", true)
}
func TestCyclesCorpus_4740ec6251604a24(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_fnexpr.a")
}
func TestProduct_CycleFrozen_f8e8af26f2ae4a34(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_resolved_project.a", "frozen", true)
}
func TestProduct_CycleCurrent_f8e8af26f2ae4a34(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_resolved_project.a", "current", true)
}
func TestCyclesCorpus_f8e8af26f2ae4a34(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_union_value_resolved_project.a")
}
func TestProduct_CycleFrozen_00564e90470e94de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/bare_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_00564e90470e94de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/bare_return.a", "current", true)
}
func TestCyclesCorpus_00564e90470e94de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/bare_return.a")
}
func TestProduct_CycleFrozen_91462807e03b7f3f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/result.a", "frozen", true)
}
func TestProduct_CycleCurrent_91462807e03b7f3f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/result.a", "current", true)
}
func TestCyclesCorpus_91462807e03b7f3f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/result.a")
}
func TestProduct_CycleFrozen_d423ea1775d25478(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/named_function_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_d423ea1775d25478(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/named_function_values.a", "current", true)
}
func TestCyclesCorpus_d423ea1775d25478(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/named_function_values.a")
}
func TestProduct_CycleFrozen_35742dfc667f14c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/native-arguments-length-value.a", "frozen", true)
}
func TestProduct_CycleCurrent_35742dfc667f14c8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/native-arguments-length-value.a", "current", true)
}
func TestCyclesCorpus_35742dfc667f14c8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/native-arguments-length-value.a")
}
func TestProduct_CycleFrozen_dce581df87eaa951(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_callback_levels.a", "frozen", true)
}
func TestProduct_CycleCurrent_dce581df87eaa951(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_callback_levels.a", "current", true)
}
func TestCyclesCorpus_dce581df87eaa951(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_callback_levels.a")
}
func TestProduct_CycleFrozen_cbcb44c6b2d42ce2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_minimal.a", "frozen", true)
}
func TestProduct_CycleCurrent_cbcb44c6b2d42ce2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_minimal.a", "current", true)
}
func TestCyclesCorpus_cbcb44c6b2d42ce2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_minimal.a")
}
func TestProduct_CycleFrozen_8e56ea5be136c54a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_reduced_parameter_unused.a", "frozen", true)
}
func TestProduct_CycleCurrent_8e56ea5be136c54a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_reduced_parameter_unused.a", "current", true)
}
func TestCyclesCorpus_8e56ea5be136c54a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_reduced_parameter_unused.a")
}
func TestProduct_CycleFrozen_5e9bf6cb813747aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_random.a", "frozen", true)
}
func TestProduct_CycleCurrent_5e9bf6cb813747aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_random.a", "current", true)
}
func TestCyclesCorpus_5e9bf6cb813747aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_buffer_random.a")
}
func TestProduct_CycleFrozen_8771d294bbef590d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_literal_assignment.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8771d294bbef590d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_literal_assignment.ts", "current", true)
}
func TestCyclesCorpus_8771d294bbef590d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_literal_assignment.ts")
}
func TestProduct_CycleFrozen_f2ba3c6e7e5665f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_null.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f2ba3c6e7e5665f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_null.ts", "current", true)
}
func TestCyclesCorpus_f2ba3c6e7e5665f5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_null.ts")
}
func TestProduct_CycleFrozen_f9e065cb77703cf2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_interface.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f9e065cb77703cf2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_interface.ts", "current", true)
}
func TestCyclesCorpus_f9e065cb77703cf2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_interface.ts")
}
func TestProduct_CycleFrozen_9d954cadf259dd7a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_cache.a", "frozen", true)
}
func TestProduct_CycleCurrent_9d954cadf259dd7a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_cache.a", "current", true)
}
func TestCyclesCorpus_9d954cadf259dd7a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/normalize_coverage_cache.a")
}
func TestProduct_CycleFrozen_20464c7e1e2c0787(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/jsdoc.a", "frozen", true)
}
func TestProduct_CycleCurrent_20464c7e1e2c0787(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/jsdoc.a", "current", true)
}
func TestCyclesCorpus_20464c7e1e2c0787(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_element_access/jsdoc.a")
}
func TestProduct_CycleFrozen_603131f3a40695de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/ownership-query/ring.a", "frozen", true)
}
func TestProduct_CycleCurrent_603131f3a40695de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/ownership-query/ring.a", "current", true)
}
func TestCyclesCorpus_603131f3a40695de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/ownership-query/ring.a")
}
func TestProduct_CycleFrozen_704e64415f287568(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_types.a", "frozen", true)
}
func TestProduct_CycleCurrent_704e64415f287568(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_types.a", "current", true)
}
func TestCyclesCorpus_704e64415f287568(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/params_namespaces_types.a")
}
func TestProduct_CycleFrozen_cc2b32880570a002(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/cycles_graph_parent.a", "frozen", true)
}
func TestProduct_CycleCurrent_cc2b32880570a002(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/cycles_graph_parent.a", "current", true)
}
func TestCyclesCorpus_cc2b32880570a002(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/program_region/cycles_graph_parent.a")
}
func TestProduct_CycleFrozen_4046b59b5303a9c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_spread_method_alias.a", "frozen", true)
}
func TestProduct_CycleCurrent_4046b59b5303a9c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_spread_method_alias.a", "current", true)
}
func TestCyclesCorpus_4046b59b5303a9c1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/reuse_spread_method_alias.a")
}
func TestProduct_CycleFrozen_ffcd1d0e28da3403(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_weak_single_narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_ffcd1d0e28da3403(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_weak_single_narrowed.a", "current", true)
}
func TestCyclesCorpus_ffcd1d0e28da3403(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_weak_single_narrowed.a")
}
func TestProduct_CycleFrozen_aa6e2f0c3d273854(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_v5.a", "frozen", true)
}
func TestProduct_CycleCurrent_aa6e2f0c3d273854(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_v5.a", "current", true)
}
func TestCyclesCorpus_aa6e2f0c3d273854(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_t_v5.a")
}
func TestProduct_CycleFrozen_cdcd9ed9fadfce18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_cdcd9ed9fadfce18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_throw.a", "current", true)
}
func TestCyclesCorpus_cdcd9ed9fadfce18(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_throw.a")
}
func TestProduct_CycleFrozen_dab132acb1cc39ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_crash_min4.a", "frozen", true)
}
func TestProduct_CycleCurrent_dab132acb1cc39ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_crash_min4.a", "current", true)
}
func TestCyclesCorpus_dab132acb1cc39ff(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_crash_min4.a")
}
func TestProduct_CycleFrozen_078cee1d3b593a79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p72_tuple_count_mismatch.a", "frozen", true)
}
func TestProduct_CycleCurrent_078cee1d3b593a79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p72_tuple_count_mismatch.a", "current", true)
}
func TestCyclesCorpus_078cee1d3b593a79(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p72_tuple_count_mismatch.a")
}
func TestProduct_CycleFrozen_c83ca6d7306e1270(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q18_every_mixed_literal.a", "frozen", true)
}
func TestProduct_CycleCurrent_c83ca6d7306e1270(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q18_every_mixed_literal.a", "current", true)
}
func TestCyclesCorpus_c83ca6d7306e1270(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q18_every_mixed_literal.a")
}
func TestProduct_CycleFrozen_6bcc63b0a38313db(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/splices.a", "frozen", true)
}
func TestProduct_CycleCurrent_6bcc63b0a38313db(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/splices.a", "current", true)
}
func TestCyclesCorpus_6bcc63b0a38313db(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/splices.a")
}
func TestProduct_CycleFrozen_3e3cf17bd812fd49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_boundaries.a", "frozen", true)
}
func TestProduct_CycleCurrent_3e3cf17bd812fd49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_boundaries.a", "current", true)
}
func TestCyclesCorpus_3e3cf17bd812fd49(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_build_boundaries.a")
}
func TestProduct_CycleFrozen_2fda864da88c914a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/before.a", "frozen", true)
}
func TestProduct_CycleCurrent_2fda864da88c914a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/before.a", "current", true)
}
func TestCyclesCorpus_2fda864da88c914a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_exports/before.a")
}
func TestProduct_CycleFrozen_1aa6dffb206396f3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_slots.a", "frozen", true)
}
func TestProduct_CycleCurrent_1aa6dffb206396f3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_slots.a", "current", true)
}
func TestCyclesCorpus_1aa6dffb206396f3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typeof_null_slots.a")
}
func TestProduct_CycleFrozen_8adbc46c9454fa2d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_strings.a", "frozen", true)
}
func TestProduct_CycleCurrent_8adbc46c9454fa2d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_strings.a", "current", true)
}
func TestCyclesCorpus_8adbc46c9454fa2d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/undefined_strings.a")
}
func TestProduct_CycleFrozen_533f0502e4ef2ae3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_stderr_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_533f0502e4ef2ae3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_stderr_order.a", "current", true)
}
func TestCyclesCorpus_533f0502e4ef2ae3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/write_stderr_order.a")
}
func TestProduct_CycleFrozen_eed69c17f937524c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/oct6/class_oct6_construction.a", "frozen", true)
}
func TestProduct_CycleCurrent_eed69c17f937524c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/oct6/class_oct6_construction.a", "current", true)
}
func TestCyclesCorpus_eed69c17f937524c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/oct6/class_oct6_construction.a")
}
func TestProduct_CycleFrozen_c3ed66adec861a10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/lineSeparator.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c3ed66adec861a10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/lineSeparator.ts", "current", true)
}
func TestCyclesCorpus_c3ed66adec861a10(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/lineSeparator.ts")
}
func TestProduct_CycleFrozen_0528586544f369d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/location.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0528586544f369d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/location.ts", "current", true)
}
func TestCyclesCorpus_0528586544f369d4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/location.ts")
}
func TestProduct_CycleFrozen_df29e15f6863230d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/formatfiles/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_df29e15f6863230d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/formatfiles/main.ts", "current", true)
}
func TestCyclesCorpus_df29e15f6863230d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/formatfiles/main.ts")
}
func TestProduct_CycleFrozen_52d9be5ebcb29d87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/3_later_function.ts", "frozen", true)
}
func TestProduct_CycleCurrent_52d9be5ebcb29d87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/3_later_function.ts", "current", true)
}
func TestCyclesCorpus_52d9be5ebcb29d87(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/gaps/3_later_function.ts")
}
func TestProduct_CycleFrozen_a5898a305fd77a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a5898a305fd77a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/main.ts", "current", true)
}
func TestCyclesCorpus_a5898a305fd77a67(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/main.ts")
}
func TestProduct_CycleFrozen_9c0865616e2de9de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/blockString.ts", "frozen", true)
}
func TestProduct_CycleCurrent_9c0865616e2de9de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/blockString.ts", "current", true)
}
func TestCyclesCorpus_9c0865616e2de9de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/blockString.ts")
}
func TestProduct_CycleFrozen_a588b9d4e40d14bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a588b9d4e40d14bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/main.ts", "current", true)
}
func TestCyclesCorpus_a588b9d4e40d14bc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/json/main.ts")
}
func TestProduct_CycleFrozen_40735bd7f0a609ef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018.a", "frozen", true)
}
func TestProduct_CycleCurrent_40735bd7f0a609ef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018.a", "current", true)
}
func TestCyclesCorpus_40735bd7f0a609ef(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/018.a")
}
func TestProduct_CycleFrozen_a4cd192af43adc5e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025.a", "frozen", true)
}
func TestProduct_CycleCurrent_a4cd192af43adc5e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025.a", "current", true)
}
func TestCyclesCorpus_a4cd192af43adc5e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/025.a")
}
func TestProduct_CycleFrozen_b8243705cd4357c5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_b8243705cd4357c5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033_runtime.a", "current", true)
}
func TestCyclesCorpus_b8243705cd4357c5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033_runtime.a")
}
func TestProduct_CycleFrozen_3aa059f90b2d0ba4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043.a", "frozen", true)
}
func TestProduct_CycleCurrent_3aa059f90b2d0ba4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043.a", "current", true)
}
func TestCyclesCorpus_3aa059f90b2d0ba4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/043.a")
}
func TestProduct_CycleFrozen_4e498d03a7ba183d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_4e498d03a7ba183d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059_runtime.a", "current", true)
}
func TestCyclesCorpus_4e498d03a7ba183d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059_runtime.a")
}
func TestProduct_CycleFrozen_da051452b91f6aa4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069.a", "frozen", true)
}
func TestProduct_CycleCurrent_da051452b91f6aa4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069.a", "current", true)
}
func TestCyclesCorpus_da051452b91f6aa4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/069.a")
}
func TestProduct_CycleFrozen_4f1b41fc68d2973e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087.a", "frozen", true)
}
func TestProduct_CycleCurrent_4f1b41fc68d2973e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087.a", "current", true)
}
func TestCyclesCorpus_4f1b41fc68d2973e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/087.a")
}
func TestProduct_CycleFrozen_b109ac48b309d9f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/s/031.a", "frozen", true)
}
func TestProduct_CycleCurrent_b109ac48b309d9f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/s/031.a", "current", true)
}
func TestCyclesCorpus_b109ac48b309d9f0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/s/031.a")
}
func TestProduct_CycleFrozen_2ad3c77ba2c30949(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/abbreviation_vocabulary.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ad3c77ba2c30949(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/abbreviation_vocabulary.a", "current", true)
}
func TestCyclesCorpus_2ad3c77ba2c30949(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/abbreviation_vocabulary.a")
}
func TestProduct_CycleFrozen_008b9cb7b7e5419b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-debugger/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_008b9cb7b7e5419b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-debugger/rule.ts", "current", true)
}
func TestCyclesCorpus_008b9cb7b7e5419b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-debugger/rule.ts")
}
func TestProduct_CycleFrozen_4f7bb980d7e6ba6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_4f7bb980d7e6ba6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty/messages.ts", "current", true)
}
func TestCyclesCorpus_4f7bb980d7e6ba6c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-empty/messages.ts")
}
func TestProduct_CycleFrozen_e6560a20d67301b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/require-yield/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e6560a20d67301b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/require-yield/messages.ts", "current", true)
}
func TestCyclesCorpus_e6560a20d67301b1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/require-yield/messages.ts")
}
func TestProduct_CycleFrozen_d50d4280ea498685(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/13_array_growth.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d50d4280ea498685(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/13_array_growth.ts", "current", true)
}
func TestCyclesCorpus_d50d4280ea498685(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/13_array_growth.ts")
}
func TestProduct_CycleFrozen_4c424e96d6214e20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/jsx_no_comment_textnodes.a", "frozen", true)
}
func TestProduct_CycleCurrent_4c424e96d6214e20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/jsx_no_comment_textnodes.a", "current", true)
}
func TestCyclesCorpus_4c424e96d6214e20(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/jsx_no_comment_textnodes.a")
}
func TestProduct_CycleFrozen_bcfb5fbbfa962a24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/graph.ts", "frozen", true)
}
func TestProduct_CycleCurrent_bcfb5fbbfa962a24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/graph.ts", "current", true)
}
func TestCyclesCorpus_bcfb5fbbfa962a24(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/graph.ts")
}
func TestProduct_CycleFrozen_af10900fa76b0ff9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_af10900fa76b0ff9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/main.ts", "current", true)
}
func TestCyclesCorpus_af10900fa76b0ff9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/main.ts")
}
func TestProduct_CycleFrozen_3df37a4efc11cfb0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/testdata/sharded_suite.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3df37a4efc11cfb0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/testdata/sharded_suite.ts", "current", true)
}
func TestCyclesCorpus_3df37a4efc11cfb0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/testdata/sharded_suite.ts")
}
func TestProduct_CycleFrozen_809b5abf9c5c7cfe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/structuralPosition.ts", "frozen", true)
}
func TestProduct_CycleCurrent_809b5abf9c5c7cfe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/structuralPosition.ts", "current", true)
}
func TestCyclesCorpus_809b5abf9c5c7cfe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/gaps/structuralPosition.ts")
}
func TestProduct_CycleFrozen_f9646a4ec2d91141(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_conditional.a", "frozen", true)
}
func TestProduct_CycleCurrent_f9646a4ec2d91141(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_conditional.a", "current", true)
}
func TestCyclesCorpus_f9646a4ec2d91141(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/047cb0d_n_conditional.a")
}
func TestProduct_CycleFrozen_456680cbd90f7a85(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_reduce_left.a", "frozen", true)
}
func TestProduct_CycleCurrent_456680cbd90f7a85(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_reduce_left.a", "current", true)
}
func TestCyclesCorpus_456680cbd90f7a85(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_reduce_left.a")
}
func TestProduct_CycleFrozen_ef35da7978aa8c65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_last.a", "frozen", true)
}
func TestProduct_CycleCurrent_ef35da7978aa8c65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_last.a", "current", true)
}
func TestCyclesCorpus_ef35da7978aa8c65(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/array_literal_empty_last.a")
}
func TestProduct_CycleFrozen_3c5611e7526edad6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_3c5611e7526edad6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_coverage_values.a", "current", true)
}
func TestCyclesCorpus_3c5611e7526edad6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_coverage_values.a")
}
func TestProduct_CycleFrozen_f84c9e03cba95eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/maybe_boolean_and.a", "frozen", true)
}
func TestProduct_CycleCurrent_f84c9e03cba95eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/maybe_boolean_and.a", "current", true)
}
func TestCyclesCorpus_f84c9e03cba95eaa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_review/refused/maybe_boolean_and.a")
}
func TestProduct_CycleFrozen_48cf8e252d008fa4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_element_virtual_store.a", "frozen", true)
}
func TestProduct_CycleCurrent_48cf8e252d008fa4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_element_virtual_store.a", "current", true)
}
func TestCyclesCorpus_48cf8e252d008fa4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_element_virtual_store.a")
}
func TestProduct_CycleFrozen_b2a8350147bf0b39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/call_targets_region.a", "frozen", true)
}
func TestProduct_CycleCurrent_b2a8350147bf0b39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/call_targets_region.a", "current", true)
}
func TestCyclesCorpus_b2a8350147bf0b39(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/call_targets_region.a")
}
func TestProduct_CycleFrozen_860395bfabb6602f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_boolean.a", "frozen", true)
}
func TestProduct_CycleCurrent_860395bfabb6602f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_boolean.a", "current", true)
}
func TestCyclesCorpus_860395bfabb6602f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_boolean.a")
}
func TestProduct_CycleFrozen_a9daf1836e92bc73(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_oct6_parameters.a", "frozen", true)
}
func TestProduct_CycleCurrent_a9daf1836e92bc73(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_oct6_parameters.a", "current", true)
}
func TestCyclesCorpus_a9daf1836e92bc73(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_oct6_parameters.a")
}
func TestProduct_CycleFrozen_7293fb6b957cd85e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/class_mutable.a", "frozen", true)
}
func TestProduct_CycleCurrent_7293fb6b957cd85e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/class_mutable.a", "current", true)
}
func TestCyclesCorpus_7293fb6b957cd85e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/class_mutable.a")
}
func TestProduct_CycleFrozen_62802926c86f4e60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_array.a", "frozen", true)
}
func TestProduct_CycleCurrent_62802926c86f4e60(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_array.a", "current", true)
}
func TestCyclesCorpus_62802926c86f4e60(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_array.a")
}
func TestProduct_CycleFrozen_d62434a9814c5db0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_function_field.a", "frozen", true)
}
func TestProduct_CycleCurrent_d62434a9814c5db0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_function_field.a", "current", true)
}
func TestCyclesCorpus_d62434a9814c5db0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_function_field.a")
}
func TestProduct_CycleFrozen_67260d2cfe9b85d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_modules/base.a", "frozen", true)
}
func TestProduct_CycleCurrent_67260d2cfe9b85d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_modules/base.a", "current", true)
}
func TestCyclesCorpus_67260d2cfe9b85d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_modules/base.a")
}
func TestProduct_CycleFrozen_03363ee9792228dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f2b_union_narrow_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_03363ee9792228dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f2b_union_narrow_number.a", "current", true)
}
func TestCyclesCorpus_03363ee9792228dc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/e4eec87_f2b_union_narrow_number.a")
}
func TestProduct_CycleFrozen_3063d8b213db1fde(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_fit.a", "frozen", true)
}
func TestProduct_CycleCurrent_3063d8b213db1fde(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_fit.a", "current", true)
}
func TestCyclesCorpus_3063d8b213db1fde(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_checked_fit.a")
}
func TestProduct_CycleFrozen_3c5ff7f3a0fe3b39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_loop.a", "frozen", true)
}
func TestProduct_CycleCurrent_3c5ff7f3a0fe3b39(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_loop.a", "current", true)
}
func TestCyclesCorpus_3c5ff7f3a0fe3b39(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_loop.a")
}
func TestProduct_CycleFrozen_dd1dc85dc8885813(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_owner.a", "frozen", true)
}
func TestProduct_CycleCurrent_dd1dc85dc8885813(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_owner.a", "current", true)
}
func TestCyclesCorpus_dd1dc85dc8885813(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_owner.a")
}
func TestProduct_CycleFrozen_b4292b1a3339f43a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/escapes_after_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_b4292b1a3339f43a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/escapes_after_return.a", "current", true)
}
func TestCyclesCorpus_b4292b1a3339f43a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/escapes_after_return.a")
}
func TestProduct_CycleFrozen_8a787dea26354a2c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_signature.a", "frozen", true)
}
func TestProduct_CycleCurrent_8a787dea26354a2c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_signature.a", "current", true)
}
func TestCyclesCorpus_8a787dea26354a2c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/function_values_signature.a")
}
func TestProduct_CycleFrozen_f4749095b23e70c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/generic_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_f4749095b23e70c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/generic_values.a", "current", true)
}
func TestCyclesCorpus_f4749095b23e70c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/generic_values.a")
}
func TestProduct_CycleFrozen_4f356037494e924a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions/classification_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_4f356037494e924a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions/classification_return.a", "current", true)
}
func TestCyclesCorpus_4f356037494e924a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions/classification_return.a")
}
func TestProduct_CycleFrozen_17ddde93bff1f5de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/indirect/a.a", "frozen", true)
}
func TestProduct_CycleCurrent_17ddde93bff1f5de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/indirect/a.a", "current", true)
}
func TestCyclesCorpus_17ddde93bff1f5de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/indirect/a.a")
}
func TestProduct_CycleFrozen_430a7fe0b9743fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/integer_format.a", "frozen", true)
}
func TestProduct_CycleCurrent_430a7fe0b9743fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/integer_format.a", "current", true)
}
func TestCyclesCorpus_430a7fe0b9743fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/integer_format.a")
}
func TestProduct_CycleFrozen_f848c9e77e05b30b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_normalize.a", "frozen", true)
}
func TestProduct_CycleCurrent_f848c9e77e05b30b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_normalize.a", "current", true)
}
func TestCyclesCorpus_f848c9e77e05b30b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_normalize.a")
}
func TestProduct_CycleFrozen_9acbd74c41b09969(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_substring.a", "frozen", true)
}
func TestProduct_CycleCurrent_9acbd74c41b09969(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_substring.a", "current", true)
}
func TestCyclesCorpus_9acbd74c41b09969(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_substring.a")
}
func TestProduct_CycleFrozen_55e37090a3bf0973(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/alias.a", "frozen", true)
}
func TestProduct_CycleCurrent_55e37090a3bf0973(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/alias.a", "current", true)
}
func TestCyclesCorpus_55e37090a3bf0973(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/alias.a")
}
func TestProduct_CycleFrozen_0c9e57b9f06496b3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_modules/main.a", "frozen", true)
}
func TestProduct_CycleCurrent_0c9e57b9f06496b3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_modules/main.a", "current", true)
}
func TestCyclesCorpus_0c9e57b9f06496b3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_modules/main.a")
}
func TestProduct_CycleFrozen_f99ac29b257eace9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_capture.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f99ac29b257eace9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_capture.ts", "current", true)
}
func TestCyclesCorpus_f99ac29b257eace9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_capture.ts")
}
func TestProduct_CycleFrozen_8887d4374b6e9659(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_definite.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8887d4374b6e9659(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_definite.ts", "current", true)
}
func TestCyclesCorpus_8887d4374b6e9659(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_definite.ts")
}
func TestProduct_CycleFrozen_b5ac57dc9f739f5f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/number_edges.a", "frozen", true)
}
func TestProduct_CycleCurrent_b5ac57dc9f739f5f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/number_edges.a", "current", true)
}
func TestCyclesCorpus_b5ac57dc9f739f5f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/number_edges.a")
}
func TestProduct_CycleFrozen_95c58c52fc04a09b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_array_to_multimap.a", "frozen", true)
}
func TestProduct_CycleCurrent_95c58c52fc04a09b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_array_to_multimap.a", "current", true)
}
func TestCyclesCorpus_95c58c52fc04a09b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_array_to_multimap.a")
}
func TestProduct_CycleFrozen_abacdbe5939d6194(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/precision_range.a", "frozen", true)
}
func TestProduct_CycleCurrent_abacdbe5939d6194(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/precision_range.a", "current", true)
}
func TestCyclesCorpus_abacdbe5939d6194(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/precision_range.a")
}
func TestProduct_CycleFrozen_17a4fbcda7ac04ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_class_guards.a", "frozen", true)
}
func TestProduct_CycleCurrent_17a4fbcda7ac04ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_class_guards.a", "current", true)
}
func TestCyclesCorpus_17a4fbcda7ac04ff(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/proven_class_guards.a")
}
func TestProduct_CycleFrozen_cf8cd655af2d6689(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe.a", "frozen", true)
}
func TestProduct_CycleCurrent_cf8cd655af2d6689(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe.a", "current", true)
}
func TestCyclesCorpus_cf8cd655af2d6689(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_f_maybe.a")
}
func TestProduct_CycleFrozen_f2bd6d487dc335db(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_aside_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_f2bd6d487dc335db(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_aside_throw.a", "current", true)
}
func TestCyclesCorpus_f2bd6d487dc335db(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_aside_throw.a")
}
func TestProduct_CycleFrozen_ee891ae8f9d939a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method_parameter.a", "frozen", true)
}
func TestProduct_CycleCurrent_ee891ae8f9d939a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method_parameter.a", "current", true)
}
func TestCyclesCorpus_ee891ae8f9d939a6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method_parameter.a")
}
func TestProduct_CycleFrozen_280d915265bda437(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_hash_small.a", "frozen", true)
}
func TestProduct_CycleCurrent_280d915265bda437(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_hash_small.a", "current", true)
}
func TestCyclesCorpus_280d915265bda437(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_hash_small.a")
}
func TestProduct_CycleFrozen_d58b44ad687c8d9d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_clobber_link_loop.a", "frozen", true)
}
func TestProduct_CycleCurrent_d58b44ad687c8d9d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_clobber_link_loop.a", "current", true)
}
func TestCyclesCorpus_d58b44ad687c8d9d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_clobber_link_loop.a")
}
func TestProduct_CycleFrozen_974362d4007c8774(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f2b_union_narrow_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_974362d4007c8774(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f2b_union_narrow_number.a", "current", true)
}
func TestCyclesCorpus_974362d4007c8774(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f2b_union_narrow_number.a")
}
func TestProduct_CycleFrozen_6c1e8c4ae2731a6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_predicates_p27_overload_parameter_rebound.a", "frozen", true)
}
func TestProduct_CycleCurrent_6c1e8c4ae2731a6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_predicates_p27_overload_parameter_rebound.a", "current", true)
}
func TestCyclesCorpus_6c1e8c4ae2731a6a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_predicates_p27_overload_parameter_rebound.a")
}
func TestProduct_CycleFrozen_a98b407cb4ccaa3f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p60_named_function_producer.a", "frozen", true)
}
func TestProduct_CycleCurrent_a98b407cb4ccaa3f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p60_named_function_producer.a", "current", true)
}
func TestCyclesCorpus_a98b407cb4ccaa3f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p60_named_function_producer.a")
}
func TestProduct_CycleFrozen_3db54225c45aa58a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p09_tuple_destructure_wrong.a", "frozen", true)
}
func TestProduct_CycleCurrent_3db54225c45aa58a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p09_tuple_destructure_wrong.a", "current", true)
}
func TestCyclesCorpus_3db54225c45aa58a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p09_tuple_destructure_wrong.a")
}
func TestProduct_CycleFrozen_69c680e82adb3841(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p59_union_member_method_shorthand.a", "frozen", true)
}
func TestProduct_CycleCurrent_69c680e82adb3841(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p59_union_member_method_shorthand.a", "current", true)
}
func TestCyclesCorpus_69c680e82adb3841(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p59_union_member_method_shorthand.a")
}
func TestProduct_CycleFrozen_eaf35edbdcb567b2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_static.a", "frozen", true)
}
func TestProduct_CycleCurrent_eaf35edbdcb567b2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_static.a", "current", true)
}
func TestCyclesCorpus_eaf35edbdcb567b2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/structural_statics_static.a")
}
func TestProduct_CycleFrozen_75249f9c82731351(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/trig_reduction.a", "frozen", true)
}
func TestProduct_CycleCurrent_75249f9c82731351(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/trig_reduction.a", "current", true)
}
func TestCyclesCorpus_75249f9c82731351(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/trig_reduction.a")
}
func TestProduct_CycleFrozen_624aa5284ec8383f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_624aa5284ec8383f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_order.a", "current", true)
}
func TestCyclesCorpus_624aa5284ec8383f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_order.a")
}
func TestProduct_CycleFrozen_0ec6c0eebd55d247(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_elements.a", "frozen", true)
}
func TestProduct_CycleCurrent_0ec6c0eebd55d247(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_elements.a", "current", true)
}
func TestCyclesCorpus_0ec6c0eebd55d247(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/undefined_elements.a")
}
func TestProduct_CycleFrozen_14d0d0563f5d1a2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/updates.a", "frozen", true)
}
func TestProduct_CycleCurrent_14d0d0563f5d1a2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/updates.a", "current", true)
}
func TestCyclesCorpus_14d0d0563f5d1a2a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/updates.a")
}
func TestProduct_CycleFrozen_2eb58f5ef0170305(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/visits.a", "frozen", true)
}
func TestProduct_CycleCurrent_2eb58f5ef0170305(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/visits.a", "current", true)
}
func TestCyclesCorpus_2eb58f5ef0170305(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/visits.a")
}
func TestProduct_CycleFrozen_f0fb1489fbfff6f2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/convert.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f0fb1489fbfff6f2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/convert.ts", "current", true)
}
func TestCyclesCorpus_f0fb1489fbfff6f2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/convert.ts")
}
func TestProduct_CycleFrozen_8cce9eea321139b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/values.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8cce9eea321139b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/values.ts", "current", true)
}
func TestCyclesCorpus_8cce9eea321139b9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/values.ts")
}
func TestProduct_CycleFrozen_a864a97b736983eb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/7_maybe_number_compared.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a864a97b736983eb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/7_maybe_number_compared.ts", "current", true)
}
func TestCyclesCorpus_a864a97b736983eb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/gaps/7_maybe_number_compared.ts")
}
func TestProduct_CycleFrozen_13a1f999c107a66c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/doc.ts", "frozen", true)
}
func TestProduct_CycleCurrent_13a1f999c107a66c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/doc.ts", "current", true)
}
func TestCyclesCorpus_13a1f999c107a66c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/json/doc.ts")
}
func TestProduct_CycleFrozen_0e0d9e81e9b140ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/finding.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0e0d9e81e9b140ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/finding.ts", "current", true)
}
func TestCyclesCorpus_0e0d9e81e9b140ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/finding.ts")
}
func TestProduct_CycleFrozen_d1940b3ea57cfab4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/lint.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d1940b3ea57cfab4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/lint.ts", "current", true)
}
func TestCyclesCorpus_d1940b3ea57cfab4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/lint.ts")
}
func TestProduct_CycleFrozen_4f12daaf45db746c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_4f12daaf45db746c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032_runtime.a", "current", true)
}
func TestCyclesCorpus_4f12daaf45db746c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032_runtime.a")
}
func TestProduct_CycleFrozen_8bea626bb4054830(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/042.a", "frozen", true)
}
func TestProduct_CycleCurrent_8bea626bb4054830(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/042.a", "current", true)
}
func TestCyclesCorpus_8bea626bb4054830(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/042.a")
}
func TestProduct_CycleFrozen_11b1da7f88c674ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_11b1da7f88c674ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058_runtime.a", "current", true)
}
func TestCyclesCorpus_11b1da7f88c674ea(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058_runtime.a")
}
func TestProduct_CycleFrozen_aab79b33c93bfdb0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/068.a", "frozen", true)
}
func TestProduct_CycleCurrent_aab79b33c93bfdb0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/068.a", "current", true)
}
func TestCyclesCorpus_aab79b33c93bfdb0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/068.a")
}
func TestProduct_CycleFrozen_e471628d25daf8d8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/086.a", "frozen", true)
}
func TestProduct_CycleCurrent_e471628d25daf8d8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/086.a", "current", true)
}
func TestCyclesCorpus_e471628d25daf8d8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/086.a")
}
func TestProduct_CycleFrozen_6bad334c16ffab56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-classes-per-file/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_6bad334c16ffab56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-classes-per-file/rule.a", "current", true)
}
func TestCyclesCorpus_6bad334c16ffab56(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/max-classes-per-file/rule.a")
}
func TestProduct_CycleFrozen_59ad65002c0328f6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_59ad65002c0328f6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/messages.a", "current", true)
}
func TestCyclesCorpus_59ad65002c0328f6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/messages.a")
}
func TestProduct_CycleFrozen_14c2efc50f1307ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-underscore-dangle/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_14c2efc50f1307ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-underscore-dangle/rule.a", "current", true)
}
func TestCyclesCorpus_14c2efc50f1307ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-underscore-dangle/rule.a")
}
func TestProduct_CycleFrozen_4b2615332a201da8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-init-declarations/initializer.a", "frozen", true)
}
func TestProduct_CycleCurrent_4b2615332a201da8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-init-declarations/initializer.a", "current", true)
}
func TestCyclesCorpus_4b2615332a201da8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-init-declarations/initializer.a")
}
func TestProduct_CycleFrozen_3ddb347f678f24b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/unicode-bom/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3ddb347f678f24b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/unicode-bom/messages.ts", "current", true)
}
func TestCyclesCorpus_3ddb347f678f24b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/unicode-bom/messages.ts")
}
func TestProduct_CycleFrozen_815e8d194590567b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/settings.ts", "frozen", true)
}
func TestProduct_CycleCurrent_815e8d194590567b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/settings.ts", "current", true)
}
func TestCyclesCorpus_815e8d194590567b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/settings.ts")
}
func TestProduct_CycleFrozen_e24bc92722a95675(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/4_structural_ranges.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e24bc92722a95675(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/4_structural_ranges.ts", "current", true)
}
func TestCyclesCorpus_e24bc92722a95675(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/4_structural_ranges.ts")
}
func TestProduct_CycleFrozen_7c75d3404f15f31f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/classes.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7c75d3404f15f31f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/classes.ts", "current", true)
}
func TestCyclesCorpus_7c75d3404f15f31f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdowninline/classes.ts")
}
func TestProduct_CycleFrozen_16452c8c74746c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_octal_escape.a", "frozen", true)
}
func TestProduct_CycleCurrent_16452c8c74746c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_octal_escape.a", "current", true)
}
func TestCyclesCorpus_16452c8c74746c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_octal_escape.a")
}
func TestProduct_CycleFrozen_34005fbbc4d10f6d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/registry.a", "frozen", true)
}
func TestProduct_CycleCurrent_34005fbbc4d10f6d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/registry.a", "current", true)
}
func TestCyclesCorpus_34005fbbc4d10f6d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/registry.a")
}
func TestProduct_CycleFrozen_fb9b23ad19f783d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/files.ts", "frozen", true)
}
func TestProduct_CycleCurrent_fb9b23ad19f783d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/files.ts", "current", true)
}
func TestCyclesCorpus_fb9b23ad19f783d4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/files.ts")
}
func TestProduct_CycleFrozen_3f5d975ad4c5a084(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/1_index_of_position.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3f5d975ad4c5a084(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/1_index_of_position.ts", "current", true)
}
func TestCyclesCorpus_3f5d975ad4c5a084(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/values/gaps/1_index_of_position.ts")
}
func TestProduct_CycleFrozen_dcb8ba47c3e125a0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/format.ts", "frozen", true)
}
func TestProduct_CycleCurrent_dcb8ba47c3e125a0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/format.ts", "current", true)
}
func TestCyclesCorpus_dcb8ba47c3e125a0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/format.ts")
}
func TestProduct_CycleFrozen_a34ff0eabb652631(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/writing.a", "frozen", true)
}
func TestProduct_CycleCurrent_a34ff0eabb652631(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_refused/writing.a", "current", true)
}
func TestCyclesCorpus_a34ff0eabb652631(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_refused/writing.a")
}
func TestProduct_CycleFrozen_4a4e38f78b37dbee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_union.a", "frozen", true)
}
func TestProduct_CycleCurrent_4a4e38f78b37dbee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_union.a", "current", true)
}
func TestCyclesCorpus_4a4e38f78b37dbee(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/array_literal_empty_union.a")
}
func TestProduct_CycleFrozen_e1f41f6e96c1be74(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_three.a", "frozen", true)
}
func TestProduct_CycleCurrent_e1f41f6e96c1be74(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_three.a", "current", true)
}
func TestCyclesCorpus_e1f41f6e96c1be74(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_three.a")
}
func TestProduct_CycleFrozen_1f0156b51b6305f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_numeric_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_1f0156b51b6305f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_numeric_fails.a", "current", true)
}
func TestCyclesCorpus_1f0156b51b6305f5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_enum_numeric_fails.a")
}
func TestProduct_CycleFrozen_0b06e9f4d2dae0da(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_operand_once.a", "frozen", true)
}
func TestProduct_CycleCurrent_0b06e9f4d2dae0da(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_operand_once.a", "current", true)
}
func TestCyclesCorpus_0b06e9f4d2dae0da(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_operand_once.a")
}
func TestProduct_CycleFrozen_88c9a3422a449c83(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_static_private_instance.a", "frozen", true)
}
func TestProduct_CycleCurrent_88c9a3422a449c83(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_static_private_instance.a", "current", true)
}
func TestCyclesCorpus_88c9a3422a449c83(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_static_private_instance.a")
}
func TestProduct_CycleFrozen_fc7b8b873bffe667(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_fc7b8b873bffe667(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a", "current", true)
}
func TestCyclesCorpus_fc7b8b873bffe667(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/iterators_hidden_return.a")
}
func TestProduct_CycleFrozen_7d4c93b71d5cd69c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/recursive_tree.a", "frozen", true)
}
func TestProduct_CycleCurrent_7d4c93b71d5cd69c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/recursive_tree.a", "current", true)
}
func TestCyclesCorpus_7d4c93b71d5cd69c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/recursive_tree.a")
}
func TestProduct_CycleFrozen_680a0cd27ea4aa71(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/async.a", "frozen", true)
}
func TestProduct_CycleCurrent_680a0cd27ea4aa71(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/async.a", "current", true)
}
func TestCyclesCorpus_680a0cd27ea4aa71(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/async.a")
}
func TestProduct_CycleFrozen_9767fadc67ec304b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/date.a", "frozen", true)
}
func TestProduct_CycleCurrent_9767fadc67ec304b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/date.a", "current", true)
}
func TestCyclesCorpus_9767fadc67ec304b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/date.a")
}
func TestProduct_CycleFrozen_c4675fea3755e117(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_abstract.a", "frozen", true)
}
func TestProduct_CycleCurrent_c4675fea3755e117(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_abstract.a", "current", true)
}
func TestCyclesCorpus_c4675fea3755e117(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_abstract.a")
}
func TestProduct_CycleFrozen_c2b8fb33a14a0920(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_twice_deferred.a", "frozen", true)
}
func TestProduct_CycleCurrent_c2b8fb33a14a0920(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_twice_deferred.a", "current", true)
}
func TestCyclesCorpus_c2b8fb33a14a0920(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_twice_deferred.a")
}
func TestProduct_CycleFrozen_0199735f9b17fcf0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_options.a", "frozen", true)
}
func TestProduct_CycleCurrent_0199735f9b17fcf0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_options.a", "current", true)
}
func TestCyclesCorpus_0199735f9b17fcf0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_options.a")
}
func TestProduct_CycleFrozen_d399a7b38af1e306(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_math_number_math.a", "frozen", true)
}
func TestProduct_CycleCurrent_d399a7b38af1e306(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_math_number_math.a", "current", true)
}
func TestCyclesCorpus_d399a7b38af1e306(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_math_number_math.a")
}
func TestProduct_CycleFrozen_33915c549919ee56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/long_literals.a", "frozen", true)
}
func TestProduct_CycleCurrent_33915c549919ee56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/long_literals.a", "current", true)
}
func TestCyclesCorpus_33915c549919ee56(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/long_literals.a")
}
func TestProduct_CycleFrozen_017c38d67e8ca316(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/loops.a", "frozen", true)
}
func TestProduct_CycleCurrent_017c38d67e8ca316(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/loops.a", "current", true)
}
func TestCyclesCorpus_017c38d67e8ca316(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/loops.a")
}
func TestProduct_CycleFrozen_ceb767542e59567d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_shrinks.a", "frozen", true)
}
func TestProduct_CycleCurrent_ceb767542e59567d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_shrinks.a", "current", true)
}
func TestCyclesCorpus_ceb767542e59567d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_shrinks.a")
}
func TestProduct_CycleFrozen_96b444783586e63b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested_race.a", "frozen", true)
}
func TestProduct_CycleCurrent_96b444783586e63b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested_race.a", "current", true)
}
func TestCyclesCorpus_96b444783586e63b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/nested_race.a")
}
func TestProduct_CycleFrozen_382907e47b3b5e4a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_parser_factory.a", "frozen", true)
}
func TestProduct_CycleCurrent_382907e47b3b5e4a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_parser_factory.a", "current", true)
}
func TestCyclesCorpus_382907e47b3b5e4a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_parser_factory.a")
}
func TestProduct_CycleFrozen_b1e1cf56c468207f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_crypto.a", "frozen", true)
}
func TestProduct_CycleCurrent_b1e1cf56c468207f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_crypto.a", "current", true)
}
func TestCyclesCorpus_b1e1cf56c468207f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_buffer_crypto.a")
}
func TestProduct_CycleFrozen_fa22f5de15f5ff6f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_fa22f5de15f5ff6f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_return.a", "current", true)
}
func TestCyclesCorpus_fa22f5de15f5ff6f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_return.a")
}
func TestProduct_CycleFrozen_e47441c501d8f77b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_undefined.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e47441c501d8f77b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_undefined.ts", "current", true)
}
func TestCyclesCorpus_e47441c501d8f77b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_undefined.ts")
}
func TestProduct_CycleFrozen_865858a432246e8a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refused/parser_index.a", "frozen", true)
}
func TestProduct_CycleCurrent_865858a432246e8a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refused/parser_index.a", "current", true)
}
func TestCyclesCorpus_865858a432246e8a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refused/parser_index.a")
}
func TestProduct_CycleFrozen_acffa1cd096c7abd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_static_initialized.ts", "frozen", true)
}
func TestProduct_CycleCurrent_acffa1cd096c7abd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_static_initialized.ts", "current", true)
}
func TestCyclesCorpus_acffa1cd096c7abd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_static_initialized.ts")
}
func TestProduct_CycleFrozen_06f0b41a4287c7fc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_default.ts", "frozen", true)
}
func TestProduct_CycleCurrent_06f0b41a4287c7fc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_default.ts", "current", true)
}
func TestCyclesCorpus_06f0b41a4287c7fc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_default.ts")
}
func TestProduct_CycleFrozen_7350a17245e54c05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_library_string_bounds.a", "frozen", true)
}
func TestProduct_CycleCurrent_7350a17245e54c05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_library_string_bounds.a", "current", true)
}
func TestCyclesCorpus_7350a17245e54c05(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_library_string_bounds.a")
}
func TestProduct_CycleFrozen_3447cebeda2921fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_scanner_required.a", "frozen", true)
}
func TestProduct_CycleCurrent_3447cebeda2921fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_scanner_required.a", "current", true)
}
func TestCyclesCorpus_3447cebeda2921fe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/omitted_scanner_required.a")
}
func TestProduct_CycleFrozen_bd170862f90db771(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/prompt_then_read.a", "frozen", true)
}
func TestProduct_CycleCurrent_bd170862f90db771(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/prompt_then_read.a", "current", true)
}
func TestCyclesCorpus_bd170862f90db771(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/prompt_then_read.a")
}
func TestProduct_CycleFrozen_9aaed99f5da75d93(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_unicode.a", "frozen", true)
}
func TestProduct_CycleCurrent_9aaed99f5da75d93(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_unicode.a", "current", true)
}
func TestCyclesCorpus_9aaed99f5da75d93(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_unicode.a")
}
func TestProduct_CycleFrozen_6a6a3c1cd033c9a5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_raw_crlf.a", "frozen", true)
}
func TestProduct_CycleCurrent_6a6a3c1cd033c9a5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_raw_crlf.a", "current", true)
}
func TestCyclesCorpus_6a6a3c1cd033c9a5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_raw_crlf.a")
}
func TestProduct_CycleFrozen_5c36e6139c3d4cf6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent_field.a", "frozen", true)
}
func TestProduct_CycleCurrent_5c36e6139c3d4cf6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent_field.a", "current", true)
}
func TestCyclesCorpus_5c36e6139c3d4cf6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent_field.a")
}
func TestProduct_CycleFrozen_fccf7d9849bf7e56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a", "frozen", true)
}
func TestProduct_CycleCurrent_fccf7d9849bf7e56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a", "current", true)
}
func TestCyclesCorpus_fccf7d9849bf7e56(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p68_object_view_over_tuple.a")
}
func TestProduct_CycleFrozen_7f934af3dfbd1216(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_903f25b_try_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_7f934af3dfbd1216(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_903f25b_try_assign.a", "current", true)
}
func TestCyclesCorpus_7f934af3dfbd1216(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_903f25b_try_assign.a")
}
func TestProduct_CycleFrozen_bd1c5d97f3421e85(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_try_loop.a", "frozen", true)
}
func TestProduct_CycleCurrent_bd1c5d97f3421e85(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_try_loop.a", "current", true)
}
func TestCyclesCorpus_bd1c5d97f3421e85(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/route_targets_try_loop.a")
}
func TestProduct_CycleFrozen_454f256b10f71deb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_virtual_fresh.a", "frozen", true)
}
func TestProduct_CycleCurrent_454f256b10f71deb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_virtual_fresh.a", "current", true)
}
func TestCyclesCorpus_454f256b10f71deb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/route_targets_virtual_fresh.a")
}
func TestProduct_CycleFrozen_31045a5723cfb12b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_named_more.a", "frozen", true)
}
func TestProduct_CycleCurrent_31045a5723cfb12b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_named_more.a", "current", true)
}
func TestCyclesCorpus_31045a5723cfb12b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_named_more.a")
}
func TestProduct_CycleFrozen_3ac77bc1d131a7e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_3ac77bc1d131a7e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_throw.a", "current", true)
}
func TestCyclesCorpus_3ac77bc1d131a7e6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/statements_small_throw.a")
}
func TestProduct_CycleFrozen_c260de0ad946e923(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a", "frozen", true)
}
func TestProduct_CycleCurrent_c260de0ad946e923(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a", "current", true)
}
func TestCyclesCorpus_c260de0ad946e923(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_direct.a")
}
func TestProduct_CycleFrozen_59bb430595eaa69c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_59bb430595eaa69c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_write.a", "current", true)
}
func TestCyclesCorpus_59bb430595eaa69c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/switch_case_declarations/dead_zone_write.a")
}
func TestProduct_CycleFrozen_ee233b1b2faa65ce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_compare.a", "frozen", true)
}
func TestProduct_CycleCurrent_ee233b1b2faa65ce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_compare.a", "current", true)
}
func TestCyclesCorpus_ee233b1b2faa65ce(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typeof_null_compare.a")
}
func TestProduct_CycleFrozen_dc1cab4741c161ca(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/unused_parameters.a", "frozen", true)
}
func TestProduct_CycleCurrent_dc1cab4741c161ca(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/unused_parameters.a", "current", true)
}
func TestCyclesCorpus_dc1cab4741c161ca(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/unused_parameters.a")
}
func TestProduct_CycleFrozen_8cdf0f7f97f84dea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/utf8_view_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_8cdf0f7f97f84dea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/utf8_view_fails.a", "current", true)
}
func TestCyclesCorpus_8cdf0f7f97f84dea(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/utf8_view_fails.a")
}
func TestProduct_CycleFrozen_6dc7d32cef6c2169(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/gaps/6_printer_regex_tree.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6dc7d32cef6c2169(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/gaps/6_printer_regex_tree.ts", "current", true)
}
func TestCyclesCorpus_6dc7d32cef6c2169(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/gaps/6_printer_regex_tree.ts")
}
func TestProduct_CycleFrozen_f6a04c3273f36224(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/methodReplacement.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f6a04c3273f36224(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/methodReplacement.ts", "current", true)
}
func TestCyclesCorpus_f6a04c3273f36224(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/methodReplacement.ts")
}
func TestProduct_CycleFrozen_465c9db4bb3a435b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/literal.ts", "frozen", true)
}
func TestProduct_CycleCurrent_465c9db4bb3a435b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/literal.ts", "current", true)
}
func TestCyclesCorpus_465c9db4bb3a435b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/literal.ts")
}
func TestProduct_CycleFrozen_605e04e91951ebda(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceLexical.ts", "frozen", true)
}
func TestProduct_CycleCurrent_605e04e91951ebda(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceLexical.ts", "current", true)
}
func TestCyclesCorpus_605e04e91951ebda(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceLexical.ts")
}
func TestProduct_CycleFrozen_16071be3bc835e73(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/lexer.ts", "frozen", true)
}
func TestProduct_CycleCurrent_16071be3bc835e73(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/lexer.ts", "current", true)
}
func TestCyclesCorpus_16071be3bc835e73(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/lexer.ts")
}
func TestProduct_CycleFrozen_c9c077522608b1ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/widthTables.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c9c077522608b1ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/json/widthTables.ts", "current", true)
}
func TestCyclesCorpus_c9c077522608b1ab(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/json/widthTables.ts")
}
func TestProduct_CycleFrozen_bdbd5c31dfe41807(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/015.a", "frozen", true)
}
func TestProduct_CycleCurrent_bdbd5c31dfe41807(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/015.a", "current", true)
}
func TestCyclesCorpus_bdbd5c31dfe41807(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/015.a")
}
func TestProduct_CycleFrozen_ba72b7f08bc8db4e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/m/044.a", "frozen", true)
}
func TestProduct_CycleCurrent_ba72b7f08bc8db4e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/m/044.a", "current", true)
}
func TestCyclesCorpus_ba72b7f08bc8db4e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/m/044.a")
}
func TestProduct_CycleFrozen_877b018ce7ac2bd6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/022_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_877b018ce7ac2bd6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/022_runtime.a", "current", true)
}
func TestCyclesCorpus_877b018ce7ac2bd6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/022_runtime.a")
}
func TestProduct_CycleFrozen_e718abcad4aa0df2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/067.a", "frozen", true)
}
func TestProduct_CycleCurrent_e718abcad4aa0df2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/067.a", "current", true)
}
func TestCyclesCorpus_e718abcad4aa0df2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/067.a")
}
func TestProduct_CycleFrozen_66a16ae0623fbf00(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/080_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_66a16ae0623fbf00(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/080_runtime.a", "current", true)
}
func TestCyclesCorpus_66a16ae0623fbf00(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/080_runtime.a")
}
func TestProduct_CycleFrozen_b9422a7db8820ad4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-sparse-arrays/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b9422a7db8820ad4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-sparse-arrays/rule.ts", "current", true)
}
func TestCyclesCorpus_b9422a7db8820ad4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-sparse-arrays/rule.ts")
}
func TestProduct_CycleFrozen_e9aea05f55fc98f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-warning-comments/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e9aea05f55fc98f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-warning-comments/rule.ts", "current", true)
}
func TestCyclesCorpus_e9aea05f55fc98f7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-warning-comments/rule.ts")
}
func TestProduct_CycleFrozen_a99029883858540e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-no-this-alias/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_a99029883858540e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-no-this-alias/rule.a", "current", true)
}
func TestCyclesCorpus_a99029883858540e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-no-this-alias/rule.a")
}
func TestProduct_CycleFrozen_259eb2dbeb5738e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-no-wrapper-object-types/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_259eb2dbeb5738e0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-no-wrapper-object-types/messages.ts", "current", true)
}
func TestCyclesCorpus_259eb2dbeb5738e0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-no-wrapper-object-types/messages.ts")
}
func TestProduct_CycleFrozen_5df73cab74326cbf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/quotePrefix.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5df73cab74326cbf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/quotePrefix.ts", "current", true)
}
func TestCyclesCorpus_5df73cab74326cbf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/quotePrefix.ts")
}
func TestProduct_CycleFrozen_cb80718694153ffd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/testdata/list_probe.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cb80718694153ffd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/testdata/list_probe.ts", "current", true)
}
func TestCyclesCorpus_cb80718694153ffd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/testdata/list_probe.ts")
}
func TestProduct_CycleFrozen_5ea3c02d5b256993(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/tokenizerEvents.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5ea3c02d5b256993(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/tokenizerEvents.ts", "current", true)
}
func TestCyclesCorpus_5ea3c02d5b256993(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/tokenizerEvents.ts")
}
func TestProduct_CycleFrozen_5fe188d9de0584e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5fe188d9de0584e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/main.ts", "current", true)
}
func TestCyclesCorpus_5fe188d9de0584e3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdowninline/main.ts")
}
func TestProduct_CycleFrozen_5ebb3bbdf0cb4b11(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/whitespace.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5ebb3bbdf0cb4b11(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/whitespace.ts", "current", true)
}
func TestCyclesCorpus_5ebb3bbdf0cb4b11(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/mediaquery/whitespace.ts")
}
func TestProduct_CycleFrozen_eba9c3d919633915(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/parentheses.ts", "frozen", true)
}
func TestProduct_CycleCurrent_eba9c3d919633915(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/parentheses.ts", "current", true)
}
func TestCyclesCorpus_eba9c3d919633915(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/parentheses.ts")
}
func TestProduct_CycleFrozen_e113a2928fafe0ed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/testdata/numberConstructor.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e113a2928fafe0ed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/testdata/numberConstructor.ts", "current", true)
}
func TestCyclesCorpus_e113a2928fafe0ed(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/testdata/numberConstructor.ts")
}
func TestProduct_CycleFrozen_2b83659b9a103d3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/5_empty_array_default.ts", "frozen", true)
}
func TestProduct_CycleCurrent_2b83659b9a103d3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/5_empty_array_default.ts", "current", true)
}
func TestCyclesCorpus_2b83659b9a103d3a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/values/gaps/5_empty_array_default.ts")
}
func TestProduct_CycleFrozen_310af666f86089e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistDocumentData.ts", "frozen", true)
}
func TestProduct_CycleCurrent_310af666f86089e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistDocumentData.ts", "current", true)
}
func TestCyclesCorpus_310af666f86089e8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/unistDocumentData.ts")
}
func TestProduct_CycleFrozen_33d9113e35b0e540(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/93122fa_s2.a", "frozen", true)
}
func TestProduct_CycleCurrent_33d9113e35b0e540(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/93122fa_s2.a", "current", true)
}
func TestCyclesCorpus_33d9113e35b0e540(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/93122fa_s2.a")
}
func TestProduct_CycleFrozen_39fffacbdaefa859(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/accepted/recursion.a", "frozen", true)
}
func TestProduct_CycleCurrent_39fffacbdaefa859(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/accepted/recursion.a", "current", true)
}
func TestCyclesCorpus_39fffacbdaefa859(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_review/accepted/recursion.a")
}
func TestProduct_CycleFrozen_12997f4f792e9eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_debug_fail_exact.a", "frozen", true)
}
func TestProduct_CycleCurrent_12997f4f792e9eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_debug_fail_exact.a", "current", true)
}
func TestCyclesCorpus_12997f4f792e9eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_stopped/census_debug_fail_exact.a")
}
func TestProduct_CycleFrozen_8e14bd7a10267a4a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_private.a", "frozen", true)
}
func TestProduct_CycleCurrent_8e14bd7a10267a4a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_private.a", "current", true)
}
func TestCyclesCorpus_8e14bd7a10267a4a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_features_private.a")
}
func TestProduct_CycleFrozen_02331d7d3c49deb6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a", "frozen", true)
}
func TestProduct_CycleCurrent_02331d7d3c49deb6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a", "current", true)
}
func TestCyclesCorpus_02331d7d3c49deb6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a")
}
func TestProduct_CycleFrozen_779e62a0ad68e7f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/global_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_779e62a0ad68e7f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/global_write.a", "current", true)
}
func TestCyclesCorpus_779e62a0ad68e7f5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/global_write.a")
}
func TestProduct_CycleFrozen_0770e80d3b235dce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_custom_message.a", "frozen", true)
}
func TestProduct_CycleCurrent_0770e80d3b235dce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_custom_message.a", "current", true)
}
func TestCyclesCorpus_0770e80d3b235dce(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_uncaught_custom_message.a")
}
func TestProduct_CycleFrozen_2d90bbd05b55e7f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_alias_join.a", "frozen", true)
}
func TestProduct_CycleCurrent_2d90bbd05b55e7f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_alias_join.a", "current", true)
}
func TestCyclesCorpus_2d90bbd05b55e7f7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_alias_join.a")
}
func TestProduct_CycleFrozen_c80c679eebdca47a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_unknown_callback.a", "frozen", true)
}
func TestProduct_CycleCurrent_c80c679eebdca47a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_unknown_callback.a", "current", true)
}
func TestCyclesCorpus_c80c679eebdca47a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_unknown_callback.a")
}
func TestProduct_CycleFrozen_732533274656ffc6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_coverage_merge.a", "frozen", true)
}
func TestProduct_CycleCurrent_732533274656ffc6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_coverage_merge.a", "current", true)
}
func TestCyclesCorpus_732533274656ffc6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_coverage_merge.a")
}
func TestProduct_CycleFrozen_274ea7078d7701b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_01.a", "frozen", true)
}
func TestProduct_CycleCurrent_274ea7078d7701b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_01.a", "current", true)
}
func TestCyclesCorpus_274ea7078d7701b0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_regression_01.a")
}
func TestProduct_CycleFrozen_b39d24eee142bb75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_array_predicate.a", "frozen", true)
}
func TestProduct_CycleCurrent_b39d24eee142bb75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_array_predicate.a", "current", true)
}
func TestCyclesCorpus_b39d24eee142bb75(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/host_array_predicate.a")
}
func TestProduct_CycleFrozen_bf63ccd93dbd5224(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_search.a", "frozen", true)
}
func TestProduct_CycleCurrent_bf63ccd93dbd5224(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_search.a", "current", true)
}
func TestCyclesCorpus_bf63ccd93dbd5224(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_array_search.a")
}
func TestProduct_CycleFrozen_e5e69f4aa9d35953(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lint_runtime_search_boundaries.a", "frozen", true)
}
func TestProduct_CycleCurrent_e5e69f4aa9d35953(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lint_runtime_search_boundaries.a", "current", true)
}
func TestCyclesCorpus_e5e69f4aa9d35953(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/lint_runtime_search_boundaries.a")
}
func TestProduct_CycleFrozen_026225b9ecb6c0f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_object_statics.a", "frozen", true)
}
func TestProduct_CycleCurrent_026225b9ecb6c0f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_object_statics.a", "current", true)
}
func TestCyclesCorpus_026225b9ecb6c0f0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_object_statics.a")
}
func TestProduct_CycleFrozen_0eb2ddbd5bb8be7a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/cycle_barrel.a", "frozen", true)
}
func TestProduct_CycleCurrent_0eb2ddbd5bb8be7a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/cycle_barrel.a", "current", true)
}
func TestCyclesCorpus_0eb2ddbd5bb8be7a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/module_namespace_reads/cycle_barrel.a")
}
func TestProduct_CycleFrozen_38f1e49d2703948e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_call_cycle.a", "frozen", true)
}
func TestProduct_CycleCurrent_38f1e49d2703948e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_call_cycle.a", "current", true)
}
func TestCyclesCorpus_38f1e49d2703948e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_call_cycle.a")
}
func TestProduct_CycleFrozen_d47ac29c1de87ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_parser_state.a", "frozen", true)
}
func TestProduct_CycleCurrent_d47ac29c1de87ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_parser_state.a", "current", true)
}
func TestCyclesCorpus_d47ac29c1de87ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_parser_state.a")
}
func TestProduct_CycleFrozen_344129f0c939de2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_write_buffer.a", "frozen", true)
}
func TestProduct_CycleCurrent_344129f0c939de2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_write_buffer.a", "current", true)
}
func TestCyclesCorpus_344129f0c939de2a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_fs_file_write_buffer.a")
}
func TestProduct_CycleFrozen_136bcb5681ea6ad9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/object_prototype.a", "frozen", true)
}
func TestProduct_CycleCurrent_136bcb5681ea6ad9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/object_prototype.a", "current", true)
}
func TestCyclesCorpus_136bcb5681ea6ad9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/object_prototype.a")
}
func TestProduct_CycleFrozen_d25d312c5db6c379(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/objects.a", "frozen", true)
}
func TestProduct_CycleCurrent_d25d312c5db6c379(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/objects.a", "current", true)
}
func TestCyclesCorpus_d25d312c5db6c379(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/objects.a")
}
func TestProduct_CycleFrozen_f597adc55e545d15(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_typed_array.a", "frozen", true)
}
func TestProduct_CycleCurrent_f597adc55e545d15(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_typed_array.a", "current", true)
}
func TestCyclesCorpus_f597adc55e545d15(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_indexing_typed_array.a")
}
func TestProduct_CycleFrozen_292ff69358105768(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p26_assert_nested_if_without_else.a", "frozen", true)
}
func TestProduct_CycleCurrent_292ff69358105768(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p26_assert_nested_if_without_else.a", "current", true)
}
func TestCyclesCorpus_292ff69358105768(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p26_assert_nested_if_without_else.a")
}
func TestProduct_CycleFrozen_7e1920f92ef7c884(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_cycle_collections.a", "frozen", true)
}
func TestProduct_CycleCurrent_7e1920f92ef7c884(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_cycle_collections.a", "current", true)
}
func TestCyclesCorpus_7e1920f92ef7c884(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_cycle_collections.a")
}
func TestProduct_CycleFrozen_4724d8cab9c45650(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace.a", "frozen", true)
}
func TestProduct_CycleCurrent_4724d8cab9c45650(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace.a", "current", true)
}
func TestCyclesCorpus_4724d8cab9c45650(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_replace.a")
}
func TestProduct_CycleFrozen_23dd1589f1f583b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_optional_field_min.a", "frozen", true)
}
func TestProduct_CycleCurrent_23dd1589f1f583b9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_optional_field_min.a", "current", true)
}
func TestCyclesCorpus_23dd1589f1f583b9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_optional_field_min.a")
}
func TestProduct_CycleFrozen_4b087be9774d93ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_s2.a", "frozen", true)
}
func TestProduct_CycleCurrent_4b087be9774d93ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_s2.a", "current", true)
}
func TestCyclesCorpus_4b087be9774d93ab(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_93122fa_s2.a")
}
func TestProduct_CycleFrozen_757652df256dde2e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent.a", "frozen", true)
}
func TestProduct_CycleCurrent_757652df256dde2e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent.a", "current", true)
}
func TestCyclesCorpus_757652df256dde2e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_defined_lent.a")
}
func TestProduct_CycleFrozen_fe8f9ec56b208891(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_escaped_before.a", "frozen", true)
}
func TestProduct_CycleCurrent_fe8f9ec56b208891(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_escaped_before.a", "current", true)
}
func TestCyclesCorpus_fe8f9ec56b208891(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_escaped_before.a")
}
func TestProduct_CycleFrozen_ceaafd9d158c92f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_method_view.a", "frozen", true)
}
func TestProduct_CycleCurrent_ceaafd9d158c92f5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_method_view.a", "current", true)
}
func TestCyclesCorpus_ceaafd9d158c92f5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_method_view.a")
}
func TestProduct_CycleFrozen_4f8a8c848e574b81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_repeat.a", "frozen", true)
}
func TestProduct_CycleCurrent_4f8a8c848e574b81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_repeat.a", "current", true)
}
func TestCyclesCorpus_4f8a8c848e574b81(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_repeat.a")
}
func TestProduct_CycleFrozen_a99fec6923aab0b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p14_callable_union_no_view.a", "frozen", true)
}
func TestProduct_CycleCurrent_a99fec6923aab0b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p14_callable_union_no_view.a", "current", true)
}
func TestCyclesCorpus_a99fec6923aab0b7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p14_callable_union_no_view.a")
}
func TestProduct_CycleFrozen_e48dc4803c5c2ed4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p21_callable_view_raw_union_slot.a", "frozen", true)
}
func TestProduct_CycleCurrent_e48dc4803c5c2ed4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p21_callable_view_raw_union_slot.a", "current", true)
}
func TestCyclesCorpus_e48dc4803c5c2ed4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p21_callable_view_raw_union_slot.a")
}
func TestProduct_CycleFrozen_c46314e585a92294(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_70.a", "frozen", true)
}
func TestProduct_CycleCurrent_c46314e585a92294(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_70.a", "current", true)
}
func TestCyclesCorpus_c46314e585a92294(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_70.a")
}
func TestProduct_CycleFrozen_eed8cd4f17e70c0d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_this.a", "frozen", true)
}
func TestProduct_CycleCurrent_eed8cd4f17e70c0d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_this.a", "current", true)
}
func TestCyclesCorpus_eed8cd4f17e70c0d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_this.a")
}
func TestProduct_CycleFrozen_9b7a8e2725b8f474(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_sym_keys_view.a", "frozen", true)
}
func TestProduct_CycleCurrent_9b7a8e2725b8f474(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_sym_keys_view.a", "current", true)
}
func TestCyclesCorpus_9b7a8e2725b8f474(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_sym_keys_view.a")
}
func TestProduct_CycleFrozen_99c924caf9986e50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct9_views_p75_callable_from_map_value.a", "frozen", true)
}
func TestProduct_CycleCurrent_99c924caf9986e50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct9_views_p75_callable_from_map_value.a", "current", true)
}
func TestCyclesCorpus_99c924caf9986e50(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_oct9_views_p75_callable_from_map_value.a")
}
func TestProduct_CycleFrozen_29ecc823fc9c2055(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_named_objects.a", "frozen", true)
}
func TestProduct_CycleCurrent_29ecc823fc9c2055(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_named_objects.a", "current", true)
}
func TestCyclesCorpus_29ecc823fc9c2055(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_named_objects.a")
}
func TestProduct_CycleFrozen_455a5826edf9e854(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_repeat.a", "frozen", true)
}
func TestProduct_CycleCurrent_455a5826edf9e854(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_repeat.a", "current", true)
}
func TestCyclesCorpus_455a5826edf9e854(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_build_repeat.a")
}
func TestProduct_CycleFrozen_48507776c8575b47(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_positions.a", "frozen", true)
}
func TestProduct_CycleCurrent_48507776c8575b47(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_positions.a", "current", true)
}
func TestCyclesCorpus_48507776c8575b47(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_positions.a")
}
func TestProduct_CycleFrozen_e8e492d5484386fb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/template_nullish.a", "frozen", true)
}
func TestProduct_CycleCurrent_e8e492d5484386fb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/template_nullish.a", "current", true)
}
func TestCyclesCorpus_e8e492d5484386fb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/template_nullish.a")
}
func TestProduct_CycleFrozen_2380efdca712b519(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes_in_try.a", "frozen", true)
}
func TestProduct_CycleCurrent_2380efdca712b519(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes_in_try.a", "current", true)
}
func TestCyclesCorpus_2380efdca712b519(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/writes_in_try.a")
}
func TestProduct_CycleFrozen_1800dadf135d550b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/host_array_typed_predicate.a", "frozen", true)
}
func TestProduct_CycleCurrent_1800dadf135d550b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/host_array_typed_predicate.a", "current", true)
}
func TestCyclesCorpus_1800dadf135d550b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/host_array_typed_predicate.a")
}
func TestProduct_CycleFrozen_47323060677986a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceGrammar.ts", "frozen", true)
}
func TestProduct_CycleCurrent_47323060677986a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceGrammar.ts", "current", true)
}
func TestCyclesCorpus_47323060677986a8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceGrammar.ts")
}
func TestProduct_CycleFrozen_45a89c9b7b455057(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/strippedText.ts", "frozen", true)
}
func TestProduct_CycleCurrent_45a89c9b7b455057(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/strippedText.ts", "current", true)
}
func TestCyclesCorpus_45a89c9b7b455057(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/strippedText.ts")
}
func TestProduct_CycleFrozen_167ce8dd1b2bb2d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/xhtmlEntities.ts", "frozen", true)
}
func TestProduct_CycleCurrent_167ce8dd1b2bb2d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/xhtmlEntities.ts", "current", true)
}
func TestCyclesCorpus_167ce8dd1b2bb2d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/xhtmlEntities.ts")
}
func TestProduct_CycleFrozen_cc8b3d0a587bba0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/5_array_from.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cc8b3d0a587bba0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/5_array_from.ts", "current", true)
}
func TestCyclesCorpus_cc8b3d0a587bba0a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/gaps/5_array_from.ts")
}
func TestProduct_CycleFrozen_28318c2b9773662e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/parser.ts", "frozen", true)
}
func TestProduct_CycleCurrent_28318c2b9773662e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/parser.ts", "current", true)
}
func TestCyclesCorpus_28318c2b9773662e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/parser.ts")
}
func TestProduct_CycleFrozen_90103f8a7cd5f95a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/context.ts", "frozen", true)
}
func TestProduct_CycleCurrent_90103f8a7cd5f95a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/context.ts", "current", true)
}
func TestCyclesCorpus_90103f8a7cd5f95a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/context.ts")
}
func TestProduct_CycleFrozen_464621007371c4b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/sort_by_position.ts", "frozen", true)
}
func TestProduct_CycleCurrent_464621007371c4b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/sort_by_position.ts", "current", true)
}
func TestCyclesCorpus_464621007371c4b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/comments/sort_by_position.ts")
}
func TestProduct_CycleFrozen_405f0d6da14efcc8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/014.a", "frozen", true)
}
func TestProduct_CycleCurrent_405f0d6da14efcc8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/014.a", "current", true)
}
func TestCyclesCorpus_405f0d6da14efcc8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/014.a")
}
func TestProduct_CycleFrozen_87d4b89c8c032cae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/001_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_87d4b89c8c032cae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/001_runtime.a", "current", true)
}
func TestCyclesCorpus_87d4b89c8c032cae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/001_runtime.a")
}
func TestProduct_CycleFrozen_8ee2e14221e2e63d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/038_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_8ee2e14221e2e63d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/038_runtime.a", "current", true)
}
func TestCyclesCorpus_8ee2e14221e2e63d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/038_runtime.a")
}
func TestProduct_CycleFrozen_1108bc76c5525a79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/052_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_1108bc76c5525a79(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/052_runtime.a", "current", true)
}
func TestCyclesCorpus_1108bc76c5525a79(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/052_runtime.a")
}
func TestProduct_CycleFrozen_4044139903133385(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/066.a", "frozen", true)
}
func TestProduct_CycleCurrent_4044139903133385(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/066.a", "current", true)
}
func TestCyclesCorpus_4044139903133385(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/066.a")
}
func TestProduct_CycleFrozen_d400185d37686dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/074_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_d400185d37686dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/074_runtime.a", "current", true)
}
func TestCyclesCorpus_d400185d37686dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/074_runtime.a")
}
func TestProduct_CycleFrozen_d1bd511c02cd1eee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/088.a", "frozen", true)
}
func TestProduct_CycleCurrent_d1bd511c02cd1eee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/088.a", "current", true)
}
func TestCyclesCorpus_d1bd511c02cd1eee(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/088.a")
}
func TestProduct_CycleFrozen_2f27d9d972199ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-for-in/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_2f27d9d972199ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-for-in/rule.a", "current", true)
}
func TestCyclesCorpus_2f27d9d972199ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-for-in/rule.a")
}
func TestProduct_CycleFrozen_fa1fefee2f3294cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-iterator/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_fa1fefee2f3294cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-iterator/rule.a", "current", true)
}
func TestCyclesCorpus_fa1fefee2f3294cd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-iterator/rule.a")
}
func TestProduct_CycleFrozen_d46b7368547d087c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-plusplus/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d46b7368547d087c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-plusplus/messages.ts", "current", true)
}
func TestCyclesCorpus_d46b7368547d087c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-plusplus/messages.ts")
}
func TestProduct_CycleFrozen_509b78f1fca264c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-proto/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_509b78f1fca264c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-proto/rule.a", "current", true)
}
func TestCyclesCorpus_509b78f1fca264c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-proto/rule.a")
}
func TestProduct_CycleFrozen_b61d5f719bff6aa9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-return-assign/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b61d5f719bff6aa9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-return-assign/messages.ts", "current", true)
}
func TestCyclesCorpus_b61d5f719bff6aa9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-return-assign/messages.ts")
}
func TestProduct_CycleFrozen_5134828cff202f1f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unneeded-ternary/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5134828cff202f1f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unneeded-ternary/messages.ts", "current", true)
}
func TestCyclesCorpus_5134828cff202f1f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-unneeded-ternary/messages.ts")
}
func TestProduct_CycleFrozen_40b7c7460b09bad1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/unicode.ts", "frozen", true)
}
func TestProduct_CycleCurrent_40b7c7460b09bad1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/unicode.ts", "current", true)
}
func TestCyclesCorpus_40b7c7460b09bad1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/unicode.ts")
}
func TestProduct_CycleFrozen_91db6175b6528532(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/10_multiple_push.ts", "frozen", true)
}
func TestProduct_CycleCurrent_91db6175b6528532(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/10_multiple_push.ts", "current", true)
}
func TestCyclesCorpus_91db6175b6528532(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/10_multiple_push.ts")
}
func TestProduct_CycleFrozen_cd82c1b83e843970(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/2_state_arrow_cycle.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cd82c1b83e843970(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/2_state_arrow_cycle.ts", "current", true)
}
func TestCyclesCorpus_cd82c1b83e843970(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/2_state_arrow_cycle.ts")
}
func TestProduct_CycleFrozen_cac248fa52dba8bf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/lists.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cac248fa52dba8bf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/lists.ts", "current", true)
}
func TestCyclesCorpus_cac248fa52dba8bf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/lists.ts")
}
func TestProduct_CycleFrozen_186fef57185ee011(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/testdata/whitespace_probe.ts", "frozen", true)
}
func TestProduct_CycleCurrent_186fef57185ee011(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/testdata/whitespace_probe.ts", "current", true)
}
func TestCyclesCorpus_186fef57185ee011(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/testdata/whitespace_probe.ts")
}
func TestProduct_CycleFrozen_388bcbd4636270f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/textTokens.ts", "frozen", true)
}
func TestProduct_CycleCurrent_388bcbd4636270f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/textTokens.ts", "current", true)
}
func TestCyclesCorpus_388bcbd4636270f7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/textTokens.ts")
}
func TestProduct_CycleFrozen_d475a30ee82e07a9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/upperEntityNames.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d475a30ee82e07a9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/upperEntityNames.ts", "current", true)
}
func TestCyclesCorpus_d475a30ee82e07a9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/upperEntityNames.ts")
}
func TestProduct_CycleFrozen_211f55646be50c3b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_useless_constructor.a", "frozen", true)
}
func TestProduct_CycleCurrent_211f55646be50c3b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_useless_constructor.a", "current", true)
}
func TestCyclesCorpus_211f55646be50c3b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_useless_constructor.a")
}
func TestProduct_CycleFrozen_b598c1f86e1627ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/parser.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b598c1f86e1627ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/parser.ts", "current", true)
}
func TestCyclesCorpus_b598c1f86e1627ea(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/selector/parser.ts")
}
func TestProduct_CycleFrozen_b9fc984b88c84812(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/testdata/fact_cost_volume.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b9fc984b88c84812(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/testdata/fact_cost_volume.ts", "current", true)
}
func TestCyclesCorpus_b9fc984b88c84812(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/testdata/fact_cost_volume.ts")
}
func TestProduct_CycleFrozen_cf81ada3a4ddb066(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/composedDocument.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cf81ada3a4ddb066(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/composedDocument.ts", "current", true)
}
func TestCyclesCorpus_cf81ada3a4ddb066(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/composedDocument.ts")
}
func TestProduct_CycleFrozen_9d940080d5b1f385(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistPoint.ts", "frozen", true)
}
func TestProduct_CycleCurrent_9d940080d5b1f385(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/unistPoint.ts", "current", true)
}
func TestCyclesCorpus_9d940080d5b1f385(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/unistPoint.ts")
}
func TestProduct_CycleFrozen_5a878c41f56ef856(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/adversarial_order.a", "frozen", true)
}
func TestProduct_CycleCurrent_5a878c41f56ef856(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/adversarial_order.a", "current", true)
}
func TestCyclesCorpus_5a878c41f56ef856(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/adversarial_order.a")
}
func TestProduct_CycleFrozen_d98951fda2caa3a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_walk.a", "frozen", true)
}
func TestProduct_CycleCurrent_d98951fda2caa3a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_walk.a", "current", true)
}
func TestCyclesCorpus_d98951fda2caa3a3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_walk.a")
}
func TestProduct_CycleFrozen_419f6a74cd083396(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_append_overload.a", "frozen", true)
}
func TestProduct_CycleCurrent_419f6a74cd083396(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_append_overload.a", "current", true)
}
func TestCyclesCorpus_419f6a74cd083396(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_append_overload.a")
}
func TestProduct_CycleFrozen_0e01267f8c26ce87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_static.a", "frozen", true)
}
func TestProduct_CycleCurrent_0e01267f8c26ce87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_static.a", "current", true)
}
func TestCyclesCorpus_0e01267f8c26ce87(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_features_static.a")
}
func TestProduct_CycleFrozen_2ee025c895dc0737(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_static_private.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ee025c895dc0737(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_static_private.a", "current", true)
}
func TestCyclesCorpus_2ee025c895dc0737(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_features_static_private.a")
}
func TestProduct_CycleFrozen_d6de2f9c76ab129d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_d6de2f9c76ab129d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_number.a", "current", true)
}
func TestCyclesCorpus_d6de2f9c76ab129d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_wrong_output_refused/classfeat_init_super_number.a")
}
func TestProduct_CycleFrozen_5839562a503be341(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/collections.a", "frozen", true)
}
func TestProduct_CycleCurrent_5839562a503be341(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/collections.a", "current", true)
}
func TestCyclesCorpus_5839562a503be341(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/collections.a")
}
func TestProduct_CycleFrozen_a6f026a8d372798c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/dispatch.a", "frozen", true)
}
func TestProduct_CycleCurrent_a6f026a8d372798c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/dispatch.a", "current", true)
}
func TestCyclesCorpus_a6f026a8d372798c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/dispatch.a")
}
func TestProduct_CycleFrozen_af664abfbdce6bb9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_ledger78.a", "frozen", true)
}
func TestProduct_CycleCurrent_af664abfbdce6bb9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_ledger78.a", "current", true)
}
func TestCyclesCorpus_af664abfbdce6bb9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_ledger78.a")
}
func TestProduct_CycleFrozen_523e3b27e19bc9ec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_boolean.a", "frozen", true)
}
func TestProduct_CycleCurrent_523e3b27e19bc9ec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_boolean.a", "current", true)
}
func TestCyclesCorpus_523e3b27e19bc9ec(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_view_boolean.a")
}
func TestProduct_CycleFrozen_efdc160864804a7d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_accessors.a", "frozen", true)
}
func TestProduct_CycleCurrent_efdc160864804a7d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_accessors.a", "current", true)
}
func TestCyclesCorpus_efdc160864804a7d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_accessors.a")
}
func TestProduct_CycleFrozen_1d5eaead5980e6c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_boolean.a", "frozen", true)
}
func TestProduct_CycleCurrent_1d5eaead5980e6c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_boolean.a", "current", true)
}
func TestCyclesCorpus_1d5eaead5980e6c1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_checked_boolean.a")
}
func TestProduct_CycleFrozen_d2a4d22f2c4d2deb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_callback_escape.a", "frozen", true)
}
func TestProduct_CycleCurrent_d2a4d22f2c4d2deb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_callback_escape.a", "current", true)
}
func TestCyclesCorpus_d2a4d22f2c4d2deb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_callback_escape.a")
}
func TestProduct_CycleFrozen_0132a775ae07f104(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_local_containers.a", "frozen", true)
}
func TestProduct_CycleCurrent_0132a775ae07f104(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_local_containers.a", "current", true)
}
func TestCyclesCorpus_0132a775ae07f104(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_local_containers.a")
}
func TestProduct_CycleFrozen_4ae6f642b92147bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_mapset.a", "frozen", true)
}
func TestProduct_CycleCurrent_4ae6f642b92147bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_mapset.a", "current", true)
}
func TestCyclesCorpus_4ae6f642b92147bc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/alias_mapset.a")
}
func TestProduct_CycleFrozen_c6b6f01be9c5dd35(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_grandchild.a", "frozen", true)
}
func TestProduct_CycleCurrent_c6b6f01be9c5dd35(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_grandchild.a", "current", true)
}
func TestCyclesCorpus_c6b6f01be9c5dd35(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_grandchild.a")
}
func TestProduct_CycleFrozen_33d45f5c43ac0df7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/spread_outside.a", "frozen", true)
}
func TestProduct_CycleCurrent_33d45f5c43ac0df7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/spread_outside.a", "current", true)
}
func TestCyclesCorpus_33d45f5c43ac0df7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/spread_outside.a")
}
func TestProduct_CycleFrozen_8baedf14a050a3ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_closure_number_methods.a", "frozen", true)
}
func TestProduct_CycleCurrent_8baedf14a050a3ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_closure_number_methods.a", "current", true)
}
func TestCyclesCorpus_8baedf14a050a3ff(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_closure_number_methods.a")
}
func TestProduct_CycleFrozen_911dd8397bf25c81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_entries.a", "frozen", true)
}
func TestProduct_CycleCurrent_911dd8397bf25c81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_entries.a", "current", true)
}
func TestCyclesCorpus_911dd8397bf25c81(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_entries.a")
}
func TestProduct_CycleFrozen_dd5e4df4b49361d6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_dd5e4df4b49361d6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_assign.a", "current", true)
}
func TestCyclesCorpus_dd5e4df4b49361d6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_assign.a")
}
func TestProduct_CycleFrozen_1c4e756b9c3556b4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_host_file_info.a", "frozen", true)
}
func TestProduct_CycleCurrent_1c4e756b9c3556b4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_host_file_info.a", "current", true)
}
func TestCyclesCorpus_1c4e756b9c3556b4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_union_value_host_file_info.a")
}
func TestProduct_CycleFrozen_abe427184e7d2d65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_array_mutations.a", "frozen", true)
}
func TestProduct_CycleCurrent_abe427184e7d2d65(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_array_mutations.a", "current", true)
}
func TestCyclesCorpus_abe427184e7d2d65(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_array_mutations.a")
}
func TestProduct_CycleFrozen_4c6a234482c30a27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_codepointat.a", "frozen", true)
}
func TestProduct_CycleCurrent_4c6a234482c30a27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_codepointat.a", "current", true)
}
func TestCyclesCorpus_4c6a234482c30a27(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_codepointat.a")
}
func TestProduct_CycleFrozen_396c692235efb2e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_concat.a", "frozen", true)
}
func TestProduct_CycleCurrent_396c692235efb2e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_concat.a", "current", true)
}
func TestCyclesCorpus_396c692235efb2e8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_concat.a")
}
func TestProduct_CycleFrozen_51c5386a5956ed86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_trim.a", "frozen", true)
}
func TestProduct_CycleCurrent_51c5386a5956ed86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_trim.a", "current", true)
}
func TestCyclesCorpus_51c5386a5956ed86(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_trim.a")
}
func TestProduct_CycleFrozen_f8b33a90ca406122(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/modules/main.a", "frozen", true)
}
func TestProduct_CycleCurrent_f8b33a90ca406122(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/modules/main.a", "current", true)
}
func TestCyclesCorpus_f8b33a90ca406122(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/modules/main.a")
}
func TestProduct_CycleFrozen_6c5795bbf9741289(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/callback_closure.a", "frozen", true)
}
func TestProduct_CycleCurrent_6c5795bbf9741289(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/callback_closure.a", "current", true)
}
func TestCyclesCorpus_6c5795bbf9741289(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/callback_closure.a")
}
func TestProduct_CycleFrozen_41898bd0d3793f35(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_modules/values.a", "frozen", true)
}
func TestProduct_CycleCurrent_41898bd0d3793f35(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_modules/values.a", "current", true)
}
func TestCyclesCorpus_41898bd0d3793f35(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_modules/values.a")
}
func TestProduct_CycleFrozen_10b150c9d8b82aec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_unknown_enum_before.a", "frozen", true)
}
func TestProduct_CycleCurrent_10b150c9d8b82aec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_unknown_enum_before.a", "current", true)
}
func TestCyclesCorpus_10b150c9d8b82aec(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_unknown_enum_before.a")
}
func TestProduct_CycleFrozen_fa579c8a5f17b23b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_destructured_tdz.a", "frozen", true)
}
func TestProduct_CycleCurrent_fa579c8a5f17b23b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_destructured_tdz.a", "current", true)
}
func TestCyclesCorpus_fa579c8a5f17b23b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_destructured_tdz.a")
}
func TestProduct_CycleFrozen_83a6832b45d87f1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_hoisting.a", "frozen", true)
}
func TestProduct_CycleCurrent_83a6832b45d87f1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_hoisting.a", "current", true)
}
func TestCyclesCorpus_83a6832b45d87f1c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_hoisting.a")
}
func TestProduct_CycleFrozen_37942eb54a4b2e4f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_mixed.a", "frozen", true)
}
func TestProduct_CycleCurrent_37942eb54a4b2e4f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_mixed.a", "current", true)
}
func TestCyclesCorpus_37942eb54a4b2e4f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_mixed.a")
}
func TestProduct_CycleFrozen_6093a4924826c8c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_conditional.a", "frozen", true)
}
func TestProduct_CycleCurrent_6093a4924826c8c4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_conditional.a", "current", true)
}
func TestCyclesCorpus_6093a4924826c8c4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_conditional.a")
}
func TestProduct_CycleFrozen_c086f8f9f5da4a90(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_element.a", "frozen", true)
}
func TestProduct_CycleCurrent_c086f8f9f5da4a90(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_element.a", "current", true)
}
func TestCyclesCorpus_c086f8f9f5da4a90(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_element.a")
}
func TestProduct_CycleFrozen_45406343718cee8d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_iteration.ts", "frozen", true)
}
func TestProduct_CycleCurrent_45406343718cee8d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_iteration.ts", "current", true)
}
func TestCyclesCorpus_45406343718cee8d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_iteration.ts")
}
func TestProduct_CycleFrozen_544a2e439611b2e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/numeric_record.a", "frozen", true)
}
func TestProduct_CycleCurrent_544a2e439611b2e3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/numeric_record.a", "current", true)
}
func TestCyclesCorpus_544a2e439611b2e3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_element_access/numeric_record.a")
}
func TestProduct_CycleFrozen_5497e600ac19da18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_unlink.a", "frozen", true)
}
func TestProduct_CycleCurrent_5497e600ac19da18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_unlink.a", "current", true)
}
func TestCyclesCorpus_5497e600ac19da18(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_unlink.a")
}
func TestProduct_CycleFrozen_81beebaa29d9e102(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p05_helper_second_parameter.a", "frozen", true)
}
func TestProduct_CycleCurrent_81beebaa29d9e102(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p05_helper_second_parameter.a", "current", true)
}
func TestCyclesCorpus_81beebaa29d9e102(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p05_helper_second_parameter.a")
}
func TestProduct_CycleFrozen_75771077035fd52e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/read_arguments.a", "frozen", true)
}
func TestProduct_CycleCurrent_75771077035fd52e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/read_arguments.a", "current", true)
}
func TestCyclesCorpus_75771077035fd52e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/read_arguments.a")
}
func TestProduct_CycleFrozen_738fb5e4dd2d3873(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_match.a", "frozen", true)
}
func TestProduct_CycleCurrent_738fb5e4dd2d3873(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_match.a", "current", true)
}
func TestCyclesCorpus_738fb5e4dd2d3873(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_match.a")
}
func TestProduct_CycleFrozen_d2f442c88db9a24e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/release_fast_graph.a", "frozen", true)
}
func TestProduct_CycleCurrent_d2f442c88db9a24e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/release_fast_graph.a", "current", true)
}
func TestCyclesCorpus_d2f442c88db9a24e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/release_fast_graph.a")
}
func TestProduct_CycleFrozen_3e64ce5bec402b1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_long_name.a", "frozen", true)
}
func TestProduct_CycleCurrent_3e64ce5bec402b1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_long_name.a", "current", true)
}
func TestCyclesCorpus_3e64ce5bec402b1c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_4ddd17f_long_name.a")
}
func TestProduct_CycleFrozen_2c28d1a1e982d1a0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_generic_conversion.a", "frozen", true)
}
func TestProduct_CycleCurrent_2c28d1a1e982d1a0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_generic_conversion.a", "current", true)
}
func TestCyclesCorpus_2c28d1a1e982d1a0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_4e1649c_generic_conversion.a")
}
func TestProduct_CycleFrozen_bb3311639fa5766b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_null_undefined_inline.a", "frozen", true)
}
func TestProduct_CycleCurrent_bb3311639fa5766b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_null_undefined_inline.a", "current", true)
}
func TestCyclesCorpus_bb3311639fa5766b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_null_undefined_inline.a")
}
func TestProduct_CycleFrozen_53e37eee058aa619(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_903f25b_iterator_tag.a", "frozen", true)
}
func TestProduct_CycleCurrent_53e37eee058aa619(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_903f25b_iterator_tag.a", "current", true)
}
func TestCyclesCorpus_53e37eee058aa619(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_903f25b_iterator_tag.a")
}
func TestProduct_CycleFrozen_cf7bdd72bccb305d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_pin_held.a", "frozen", true)
}
func TestProduct_CycleCurrent_cf7bdd72bccb305d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_pin_held.a", "current", true)
}
func TestCyclesCorpus_cf7bdd72bccb305d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_a34388d_pin_held.a")
}
func TestProduct_CycleFrozen_9994b2f985b2c724(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_region.a", "frozen", true)
}
func TestProduct_CycleCurrent_9994b2f985b2c724(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_region.a", "current", true)
}
func TestCyclesCorpus_9994b2f985b2c724(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_abfe962_split_region.a")
}
func TestProduct_CycleFrozen_e835a38d7f0bdc86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_error_preview.a", "frozen", true)
}
func TestProduct_CycleCurrent_e835a38d7f0bdc86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_error_preview.a", "current", true)
}
func TestCyclesCorpus_e835a38d7f0bdc86(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct8_nodehost_fs_error_preview.a")
}
func TestProduct_CycleFrozen_d13c2bb05bf8d3f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p17_callable_call_first_half.a", "frozen", true)
}
func TestProduct_CycleCurrent_d13c2bb05bf8d3f7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p17_callable_call_first_half.a", "current", true)
}
func TestCyclesCorpus_d13c2bb05bf8d3f7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p17_callable_call_first_half.a")
}
func TestProduct_CycleFrozen_2bb67440810b6313(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_b.a", "frozen", true)
}
func TestProduct_CycleCurrent_2bb67440810b6313(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_b.a", "current", true)
}
func TestCyclesCorpus_2bb67440810b6313(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_b.a")
}
func TestProduct_CycleFrozen_5294871ff2c3d114(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_a34388d_it_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_5294871ff2c3d114(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_a34388d_it_spread.a", "current", true)
}
func TestCyclesCorpus_5294871ff2c3d114(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_a34388d_it_spread.a")
}
func TestProduct_CycleFrozen_b80c98f4f4f4bd51(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early.a", "frozen", true)
}
func TestProduct_CycleCurrent_b80c98f4f4f4bd51(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early.a", "current", true)
}
func TestCyclesCorpus_b80c98f4f4f4bd51(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early.a")
}
func TestProduct_CycleFrozen_6c97ecf8b9e52c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_keep.a", "frozen", true)
}
func TestProduct_CycleCurrent_6c97ecf8b9e52c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_keep.a", "current", true)
}
func TestCyclesCorpus_6c97ecf8b9e52c1c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_keep.a")
}
func TestProduct_CycleFrozen_1a09c9cd594ee2fc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/barrel.a", "frozen", true)
}
func TestProduct_CycleCurrent_1a09c9cd594ee2fc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_star/barrel.a", "current", true)
}
func TestCyclesCorpus_1a09c9cd594ee2fc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_star/barrel.a")
}
func TestProduct_CycleFrozen_7892343ac25d3461(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_int32array.a", "frozen", true)
}
func TestProduct_CycleCurrent_7892343ac25d3461(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_int32array.a", "current", true)
}
func TestCyclesCorpus_7892343ac25d3461(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_int32array.a")
}
func TestProduct_CycleFrozen_1035e923e7fb48af(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_references.a", "frozen", true)
}
func TestProduct_CycleCurrent_1035e923e7fb48af(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/undefined_references.a", "current", true)
}
func TestCyclesCorpus_1035e923e7fb48af(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/undefined_references.a")
}
func TestProduct_CycleFrozen_e1a139150282d43e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak/reuse.a", "frozen", true)
}
func TestProduct_CycleCurrent_e1a139150282d43e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak/reuse.a", "current", true)
}
func TestCyclesCorpus_e1a139150282d43e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/weak/reuse.a")
}
func TestProduct_CycleFrozen_dd216cc07c46340a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/byte_view.ts", "frozen", true)
}
func TestProduct_CycleCurrent_dd216cc07c46340a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/byte_view.ts", "current", true)
}
func TestCyclesCorpus_dd216cc07c46340a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/byte_view.ts")
}
func TestProduct_CycleFrozen_003d5aefab4c16d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/nbsp.ts", "frozen", true)
}
func TestProduct_CycleCurrent_003d5aefab4c16d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/nbsp.ts", "current", true)
}
func TestCyclesCorpus_003d5aefab4c16d5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/nbsp.ts")
}
func TestProduct_CycleFrozen_033f43f9f070bceb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceReferences.ts", "frozen", true)
}
func TestProduct_CycleCurrent_033f43f9f070bceb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceReferences.ts", "current", true)
}
func TestCyclesCorpus_033f43f9f070bceb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceReferences.ts")
}
func TestProduct_CycleFrozen_9481dd412f36f6b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceSpelling.ts", "frozen", true)
}
func TestProduct_CycleCurrent_9481dd412f36f6b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceSpelling.ts", "current", true)
}
func TestCyclesCorpus_9481dd412f36f6b7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceSpelling.ts")
}
func TestProduct_CycleFrozen_4c7acd896390cc98(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/typeLists.ts", "frozen", true)
}
func TestProduct_CycleCurrent_4c7acd896390cc98(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/typeLists.ts", "current", true)
}
func TestCyclesCorpus_4c7acd896390cc98(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/typeLists.ts")
}
func TestProduct_CycleFrozen_7733ba6916c2dd59(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/formatfiles/disk.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7733ba6916c2dd59(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/formatfiles/disk.ts", "current", true)
}
func TestCyclesCorpus_7733ba6916c2dd59(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/formatfiles/disk.ts")
}
func TestProduct_CycleFrozen_ee3241c953c5c8b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/gaps/2_error_made_elsewhere.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ee3241c953c5c8b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/gaps/2_error_made_elsewhere.ts", "current", true)
}
func TestCyclesCorpus_ee3241c953c5c8b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/gaps/2_error_made_elsewhere.ts")
}
func TestProduct_CycleFrozen_bce251f8e5946331(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_bce251f8e5946331(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006_runtime.a", "current", true)
}
func TestCyclesCorpus_bce251f8e5946331(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006_runtime.a")
}
func TestProduct_CycleFrozen_a5db2802cd363a9e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032.a", "frozen", true)
}
func TestProduct_CycleCurrent_a5db2802cd363a9e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032.a", "current", true)
}
func TestCyclesCorpus_a5db2802cd363a9e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/032.a")
}
func TestProduct_CycleFrozen_b1218d5d5caa878b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_b1218d5d5caa878b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055_runtime.a", "current", true)
}
func TestCyclesCorpus_b1218d5d5caa878b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055_runtime.a")
}
func TestProduct_CycleFrozen_834e4bd71a5e7e87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058.a", "frozen", true)
}
func TestProduct_CycleCurrent_834e4bd71a5e7e87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058.a", "current", true)
}
func TestCyclesCorpus_834e4bd71a5e7e87(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/058.a")
}
func TestProduct_CycleFrozen_fe29bb6c2ab02a1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_fe29bb6c2ab02a1c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073_runtime.a", "current", true)
}
func TestCyclesCorpus_fe29bb6c2ab02a1c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073_runtime.a")
}
func TestProduct_CycleFrozen_b86fa905e0910177(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-nested-callbacks/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_b86fa905e0910177(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-nested-callbacks/rule.a", "current", true)
}
func TestCyclesCorpus_b86fa905e0910177(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/max-nested-callbacks/rule.a")
}
func TestProduct_CycleFrozen_bc319b40b91c13de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/patterns.a", "frozen", true)
}
func TestProduct_CycleCurrent_bc319b40b91c13de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/patterns.a", "current", true)
}
func TestCyclesCorpus_bc319b40b91c13de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/patterns.a")
}
func TestProduct_CycleFrozen_6fee97d9085be44d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-concat/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_6fee97d9085be44d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-useless-concat/rule.a", "current", true)
}
func TestCyclesCorpus_6fee97d9085be44d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-useless-concat/rule.a")
}
func TestProduct_CycleFrozen_8dab5b20c36cecd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-no-find-dom-node/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8dab5b20c36cecd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-no-find-dom-node/rule.ts", "current", true)
}
func TestCyclesCorpus_8dab5b20c36cecd3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/react-no-find-dom-node/rule.ts")
}
func TestProduct_CycleFrozen_623503aaec36f9b5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/preprocess.ts", "frozen", true)
}
func TestProduct_CycleCurrent_623503aaec36f9b5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/preprocess.ts", "current", true)
}
func TestCyclesCorpus_623503aaec36f9b5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/preprocess.ts")
}
func TestProduct_CycleFrozen_782901e3e30434bb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/directives.ts", "frozen", true)
}
func TestProduct_CycleCurrent_782901e3e30434bb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/directives.ts", "current", true)
}
func TestCyclesCorpus_782901e3e30434bb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/suppression/directives.ts")
}
func TestProduct_CycleFrozen_d0c5cf294a088032(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/facts.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d0c5cf294a088032(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/facts.ts", "current", true)
}
func TestCyclesCorpus_d0c5cf294a088032(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/facts.ts")
}
func TestProduct_CycleFrozen_a5632e8ad2e85eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/returns.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a5632e8ad2e85eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/returns.ts", "current", true)
}
func TestCyclesCorpus_a5632e8ad2e85eb8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/returns.ts")
}
func TestProduct_CycleFrozen_0c237f26503e96e7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_finally.a", "frozen", true)
}
func TestProduct_CycleCurrent_0c237f26503e96e7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_finally.a", "current", true)
}
func TestCyclesCorpus_0c237f26503e96e7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_coverage_finally.a")
}
func TestProduct_CycleFrozen_c715244695051ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_overload.a", "frozen", true)
}
func TestProduct_CycleCurrent_c715244695051ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_overload.a", "current", true)
}
func TestCyclesCorpus_c715244695051ea5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_overload.a")
}
func TestProduct_CycleFrozen_39a9e1612ef60e72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_boolean_absent_write_stopped.a", "frozen", true)
}
func TestProduct_CycleCurrent_39a9e1612ef60e72(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_boolean_absent_write_stopped.a", "current", true)
}
func TestCyclesCorpus_39a9e1612ef60e72(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_stopped/census_boolean_absent_write_stopped.a")
}
func TestProduct_CycleFrozen_72880a0e50d0d0a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/numbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_72880a0e50d0d0a6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/numbers.a", "current", true)
}
func TestCyclesCorpus_72880a0e50d0d0a6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/numbers.a")
}
func TestProduct_CycleFrozen_2337f13d6f59d1a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_2337f13d6f59d1a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_number.a", "current", true)
}
func TestCyclesCorpus_2337f13d6f59d1a3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_number.a")
}
func TestProduct_CycleFrozen_3f3536e7d62ff499(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize.a", "frozen", true)
}
func TestProduct_CycleCurrent_3f3536e7d62ff499(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize.a", "current", true)
}
func TestCyclesCorpus_3f3536e7d62ff499(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize.a")
}
func TestProduct_CycleFrozen_2328616660da27dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/exceptions_empty.a", "frozen", true)
}
func TestProduct_CycleCurrent_2328616660da27dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/exceptions_empty.a", "current", true)
}
func TestCyclesCorpus_2328616660da27dc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/exceptions_empty.a")
}
func TestProduct_CycleFrozen_2af6c90496971e20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fallthrough_labels.a", "frozen", true)
}
func TestProduct_CycleCurrent_2af6c90496971e20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fallthrough_labels.a", "current", true)
}
func TestCyclesCorpus_2af6c90496971e20(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fallthrough_labels.a")
}
func TestProduct_CycleFrozen_aee202b549d2aa10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_calls.a", "frozen", true)
}
func TestProduct_CycleCurrent_aee202b549d2aa10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_calls.a", "current", true)
}
func TestCyclesCorpus_aee202b549d2aa10(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_calls.a")
}
func TestProduct_CycleFrozen_dc9e7e0632cea206(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_fill.a", "frozen", true)
}
func TestProduct_CycleCurrent_dc9e7e0632cea206(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/alias_fill.a", "current", true)
}
func TestCyclesCorpus_dc9e7e0632cea206(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/alias_fill.a")
}
func TestProduct_CycleFrozen_cc30687c9b0d6a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/global_then_function.a", "frozen", true)
}
func TestProduct_CycleCurrent_cc30687c9b0d6a67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/global_then_function.a", "current", true)
}
func TestCyclesCorpus_cc30687c9b0d6a67(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/global_then_function.a")
}
func TestProduct_CycleFrozen_edff34ed87c48c8b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_for_in_live.a", "frozen", true)
}
func TestProduct_CycleCurrent_edff34ed87c48c8b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_for_in_live.a", "current", true)
}
func TestCyclesCorpus_edff34ed87c48c8b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_for_in_live.a")
}
func TestProduct_CycleFrozen_d577e18ff9cf6b88(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_function_expressions.a", "frozen", true)
}
func TestProduct_CycleCurrent_d577e18ff9cf6b88(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_function_expressions.a", "current", true)
}
func TestCyclesCorpus_d577e18ff9cf6b88(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_function_expressions.a")
}
func TestProduct_CycleFrozen_375acf73758fdd38(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_exhausted.a", "frozen", true)
}
func TestProduct_CycleCurrent_375acf73758fdd38(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_exhausted.a", "current", true)
}
func TestCyclesCorpus_375acf73758fdd38(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_exhausted.a")
}
func TestProduct_CycleFrozen_d82c7ae564b73a16(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_entries_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_d82c7ae564b73a16(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_entries_assign.a", "current", true)
}
func TestCyclesCorpus_d82c7ae564b73a16(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_set_entries_assign.a")
}
func TestProduct_CycleFrozen_fea2226d5405f88d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_same.a", "frozen", true)
}
func TestProduct_CycleCurrent_fea2226d5405f88d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_same.a", "current", true)
}
func TestCyclesCorpus_fea2226d5405f88d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_same.a")
}
func TestProduct_CycleFrozen_2ec3e0b2b6219edc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lint_runtime_equal_headers.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ec3e0b2b6219edc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lint_runtime_equal_headers.a", "current", true)
}
func TestCyclesCorpus_2ec3e0b2b6219edc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/lint_runtime_equal_headers.a")
}
func TestProduct_CycleFrozen_bc02e9f074e98483(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/logical_and_reference_refused/logical_and_reference_null.a", "frozen", true)
}
func TestProduct_CycleCurrent_bc02e9f074e98483(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/logical_and_reference_refused/logical_and_reference_null.a", "current", true)
}
func TestCyclesCorpus_bc02e9f074e98483(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/logical_and_reference_refused/logical_and_reference_null.a")
}
func TestProduct_CycleFrozen_549db8ae147d436b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_keys.a", "frozen", true)
}
func TestProduct_CycleCurrent_549db8ae147d436b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_keys.a", "current", true)
}
func TestCyclesCorpus_549db8ae147d436b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_keys.a")
}
func TestProduct_CycleFrozen_33564434b22a5436(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/memory_examples/strings.a", "frozen", true)
}
func TestProduct_CycleCurrent_33564434b22a5436(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/memory_examples/strings.a", "current", true)
}
func TestCyclesCorpus_33564434b22a5436(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/memory_examples/strings.a")
}
func TestProduct_CycleFrozen_96b00288bae3ed94(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_map_parsers.a", "frozen", true)
}
func TestProduct_CycleCurrent_96b00288bae3ed94(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_map_parsers.a", "current", true)
}
func TestCyclesCorpus_96b00288bae3ed94(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_map_parsers.a")
}
func TestProduct_CycleFrozen_df714e2eedb9b74e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/direct_early.a", "frozen", true)
}
func TestProduct_CycleCurrent_df714e2eedb9b74e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/direct_early.a", "current", true)
}
func TestCyclesCorpus_df714e2eedb9b74e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/module_namespace_reads/direct_early.a")
}
func TestProduct_CycleFrozen_1e9722469ca2a0c2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_debug_modules/main.a", "frozen", true)
}
func TestProduct_CycleCurrent_1e9722469ca2a0c2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_debug_modules/main.a", "current", true)
}
func TestCyclesCorpus_1e9722469ca2a0c2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_debug_modules/main.a")
}
func TestProduct_CycleFrozen_60c07030991eaace(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/narrowed_reads.a", "frozen", true)
}
func TestProduct_CycleCurrent_60c07030991eaace(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/narrowed_reads.a", "current", true)
}
func TestCyclesCorpus_60c07030991eaace(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/narrowed_reads.a")
}
func TestProduct_CycleFrozen_bd4180055d9d7d67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/new_expression_class_cache_capture_notyet.a", "frozen", true)
}
func TestProduct_CycleCurrent_bd4180055d9d7d67(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/new_expression_class_cache_capture_notyet.a", "current", true)
}
func TestCyclesCorpus_bd4180055d9d7d67(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/new_expression_class_cache_capture_notyet.a")
}
func TestProduct_CycleFrozen_a642c1588b5a2125(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_finalized.a", "frozen", true)
}
func TestProduct_CycleCurrent_a642c1588b5a2125(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_buffer_finalized.a", "current", true)
}
func TestCyclesCorpus_a642c1588b5a2125(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_buffer_finalized.a")
}
func TestProduct_CycleFrozen_b4b50204bcf5f2c3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_optional.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b4b50204bcf5f2c3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_optional.ts", "current", true)
}
func TestCyclesCorpus_b4b50204bcf5f2c3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_optional.ts")
}
func TestProduct_CycleFrozen_f668a79ae4312d46(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/paths.a", "frozen", true)
}
func TestProduct_CycleCurrent_f668a79ae4312d46(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/paths.a", "current", true)
}
func TestCyclesCorpus_f668a79ae4312d46(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_element_access/paths.a")
}
func TestProduct_CycleFrozen_41bf5f178f087595(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_string_gap.a", "frozen", true)
}
func TestProduct_CycleCurrent_41bf5f178f087595(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_string_gap.a", "current", true)
}
func TestCyclesCorpus_41bf5f178f087595(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_indexing_string_gap.a")
}
func TestProduct_CycleFrozen_d169942f1d842f81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_contracts/token.a", "frozen", true)
}
func TestProduct_CycleCurrent_d169942f1d842f81(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_contracts/token.a", "current", true)
}
func TestCyclesCorpus_d169942f1d842f81(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_contracts/token.a")
}
func TestProduct_CycleFrozen_77300f2c88780dbc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/pow_fractional.a", "frozen", true)
}
func TestProduct_CycleCurrent_77300f2c88780dbc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/pow_fractional.a", "current", true)
}
func TestCyclesCorpus_77300f2c88780dbc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/pow_fractional.a")
}
func TestProduct_CycleFrozen_eeea76a2b61c7944(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p13_filter_wrong_parameter_helper.a", "frozen", true)
}
func TestProduct_CycleCurrent_eeea76a2b61c7944(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p13_filter_wrong_parameter_helper.a", "current", true)
}
func TestCyclesCorpus_eeea76a2b61c7944(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/oct8_predicates_p13_filter_wrong_parameter_helper.a")
}
func TestProduct_CycleFrozen_0a3e870c00afde2b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/representation_controls.a", "frozen", true)
}
func TestProduct_CycleCurrent_0a3e870c00afde2b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/predicate_refusals/representation_controls.a", "current", true)
}
func TestCyclesCorpus_0a3e870c00afde2b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/predicate_refusals/representation_controls.a")
}
func TestProduct_CycleFrozen_d3765fbc240727f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/presence_refused/in_union_narrowing.a", "frozen", true)
}
func TestProduct_CycleCurrent_d3765fbc240727f0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/presence_refused/in_union_narrowing.a", "current", true)
}
func TestCyclesCorpus_d3765fbc240727f0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/presence_refused/in_union_narrowing.a")
}
func TestProduct_CycleFrozen_b5b53288a185e724(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_escapes.a", "frozen", true)
}
func TestProduct_CycleCurrent_b5b53288a185e724(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_escapes.a", "current", true)
}
func TestCyclesCorpus_b5b53288a185e724(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regions_escapes.a")
}
func TestProduct_CycleFrozen_c84fec2c76559d7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_c84fec2c76559d7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_throw.a", "current", true)
}
func TestCyclesCorpus_c84fec2c76559d7f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regions_throw.a")
}
func TestProduct_CycleFrozen_e8e9e25c768dfb08(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks.a", "frozen", true)
}
func TestProduct_CycleCurrent_e8e9e25c768dfb08(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks.a", "current", true)
}
func TestCyclesCorpus_e8e9e25c768dfb08(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks.a")
}
func TestProduct_CycleFrozen_2173c9defce1564d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks2.a", "frozen", true)
}
func TestProduct_CycleCurrent_2173c9defce1564d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks2.a", "current", true)
}
func TestCyclesCorpus_2173c9defce1564d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_57f2d04_find_last_shrinks2.a")
}
func TestProduct_CycleFrozen_b61522db7c12e063(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_narrowed_sort.a", "frozen", true)
}
func TestProduct_CycleCurrent_b61522db7c12e063(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_narrowed_sort.a", "current", true)
}
func TestCyclesCorpus_b61522db7c12e063(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_narrowed_sort.a")
}
func TestProduct_CycleFrozen_97337f6dc9f249d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method.a", "frozen", true)
}
func TestProduct_CycleCurrent_97337f6dc9f249d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method.a", "current", true)
}
func TestCyclesCorpus_97337f6dc9f249d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_9984394_view_method.a")
}
func TestProduct_CycleFrozen_d0e76fa961750c6b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_fa49e43_spreadfresh.a", "frozen", true)
}
func TestProduct_CycleCurrent_d0e76fa961750c6b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_fa49e43_spreadfresh.a", "current", true)
}
func TestCyclesCorpus_d0e76fa961750c6b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_fa49e43_spreadfresh.a")
}
func TestProduct_CycleFrozen_7703f68e704c9f10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q02_findlast_inferred.a", "frozen", true)
}
func TestProduct_CycleCurrent_7703f68e704c9f10(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q02_findlast_inferred.a", "current", true)
}
func TestCyclesCorpus_7703f68e704c9f10(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q02_findlast_inferred.a")
}
func TestProduct_CycleFrozen_63c710837f59e7ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_maybe_numbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_63c710837f59e7ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_maybe_numbers.a", "current", true)
}
func TestCyclesCorpus_63c710837f59e7ae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_maybe_numbers.a")
}
func TestProduct_CycleFrozen_89cb110a053e77e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_loops.a", "frozen", true)
}
func TestProduct_CycleCurrent_89cb110a053e77e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_loops.a", "current", true)
}
func TestCyclesCorpus_89cb110a053e77e9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_views_loops.a")
}
func TestProduct_CycleFrozen_c6a07132f9544e0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/print.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c6a07132f9544e0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/print.ts", "current", true)
}
func TestCyclesCorpus_c6a07132f9544e0a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/print.ts")
}
func TestProduct_CycleFrozen_739a27e3f07c0ee4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/binaryConvert.ts", "frozen", true)
}
func TestProduct_CycleCurrent_739a27e3f07c0ee4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/binaryConvert.ts", "current", true)
}
func TestCyclesCorpus_739a27e3f07c0ee4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/binaryConvert.ts")
}
func TestProduct_CycleFrozen_3a672f33373c7781(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/comments.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3a672f33373c7781(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/comments.ts", "current", true)
}
func TestCyclesCorpus_3a672f33373c7781(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/comments.ts")
}
func TestProduct_CycleFrozen_3745cdc142b87c27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/postfixValue.ts", "frozen", true)
}
func TestProduct_CycleCurrent_3745cdc142b87c27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/postfixValue.ts", "current", true)
}
func TestCyclesCorpus_3745cdc142b87c27(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/postfixValue.ts")
}
func TestProduct_CycleFrozen_6db6cb7a2e41556d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/case.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6db6cb7a2e41556d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/case.ts", "current", true)
}
func TestCyclesCorpus_6db6cb7a2e41556d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/case.ts")
}
func TestProduct_CycleFrozen_f21923f82b69341f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/corpus.a", "frozen", true)
}
func TestProduct_CycleCurrent_f21923f82b69341f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/corpus.a", "current", true)
}
func TestCyclesCorpus_f21923f82b69341f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/corpus.a")
}
func TestProduct_CycleFrozen_c074c9ad0ac0d9e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_c074c9ad0ac0d9e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048_runtime.a", "current", true)
}
func TestCyclesCorpus_c074c9ad0ac0d9e4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048_runtime.a")
}
func TestProduct_CycleFrozen_14f1f6a8ad07273d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/019_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_14f1f6a8ad07273d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/019_runtime.a", "current", true)
}
func TestCyclesCorpus_14f1f6a8ad07273d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/i/019_runtime.a")
}
func TestProduct_CycleFrozen_1a7f4b9557679945(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033.a", "frozen", true)
}
func TestProduct_CycleCurrent_1a7f4b9557679945(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033.a", "current", true)
}
func TestCyclesCorpus_1a7f4b9557679945(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/033.a")
}
func TestProduct_CycleFrozen_52b81dbfa2e1aa1f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_52b81dbfa2e1aa1f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054_runtime.a", "current", true)
}
func TestCyclesCorpus_52b81dbfa2e1aa1f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054_runtime.a")
}
func TestProduct_CycleFrozen_ec21a3be402c01aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059.a", "frozen", true)
}
func TestProduct_CycleCurrent_ec21a3be402c01aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059.a", "current", true)
}
func TestCyclesCorpus_ec21a3be402c01aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/059.a")
}
func TestProduct_CycleFrozen_f03597742b2c6b2c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_f03597742b2c6b2c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072_runtime.a", "current", true)
}
func TestCyclesCorpus_f03597742b2c6b2c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072_runtime.a")
}
func TestProduct_CycleFrozen_65f6316ebd5afbad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/adamic-no-type-predicate/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_65f6316ebd5afbad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/adamic-no-type-predicate/messages.ts", "current", true)
}
func TestCyclesCorpus_65f6316ebd5afbad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/adamic-no-type-predicate/messages.ts")
}
func TestProduct_CycleFrozen_059f6c46bdad265e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-boundary-no-global-container/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_059f6c46bdad265e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-boundary-no-global-container/messages.a", "current", true)
}
func TestCyclesCorpus_059f6c46bdad265e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/base-boundary-no-global-container/messages.a")
}
func TestProduct_CycleFrozen_97270c35a95607d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-consistency-no-console/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_97270c35a95607d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-consistency-no-console/messages.ts", "current", true)
}
func TestCyclesCorpus_97270c35a95607d1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/base-consistency-no-console/messages.ts")
}
func TestProduct_CycleFrozen_1a1bebba2d1a5f12(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-negated-condition/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_1a1bebba2d1a5f12(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-negated-condition/messages.ts", "current", true)
}
func TestCyclesCorpus_1a1bebba2d1a5f12(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-negated-condition/messages.ts")
}
func TestProduct_CycleFrozen_a9484b3d0c93820b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-init-declarations/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_a9484b3d0c93820b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-init-declarations/rule.a", "current", true)
}
func TestCyclesCorpus_a9484b3d0c93820b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-init-declarations/rule.a")
}
func TestProduct_CycleFrozen_fc5ff5b041ee1b8f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/parseFrontMatter.ts", "frozen", true)
}
func TestProduct_CycleCurrent_fc5ff5b041ee1b8f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/parseFrontMatter.ts", "current", true)
}
func TestCyclesCorpus_fc5ff5b041ee1b8f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/parseFrontMatter.ts")
}
func TestProduct_CycleFrozen_a63529142e8c463a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_redundant_should_component_update.a", "frozen", true)
}
func TestProduct_CycleCurrent_a63529142e8c463a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_redundant_should_component_update.a", "current", true)
}
func TestCyclesCorpus_a63529142e8c463a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_redundant_should_component_update.a")
}
func TestProduct_CycleFrozen_7a957f0a840e7dfb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_unexpected_multiline.a", "frozen", true)
}
func TestProduct_CycleCurrent_7a957f0a840e7dfb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/no_unexpected_multiline.a", "current", true)
}
func TestCyclesCorpus_7a957f0a840e7dfb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/no_unexpected_multiline.a")
}
func TestProduct_CycleFrozen_bc34f374100f7a7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/parse.a", "frozen", true)
}
func TestProduct_CycleCurrent_bc34f374100f7a7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/parse.a", "current", true)
}
func TestCyclesCorpus_bc34f374100f7a7f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/parse.a")
}
func TestProduct_CycleFrozen_a43b467607e69960(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/nodes.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a43b467607e69960(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/nodes.ts", "current", true)
}
func TestCyclesCorpus_a43b467607e69960(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/selector/nodes.ts")
}
func TestProduct_CycleFrozen_f12d126361316099(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f12d126361316099(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/main.ts", "current", true)
}
func TestCyclesCorpus_f12d126361316099(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/main.ts")
}
func TestProduct_CycleFrozen_aea045997f89492e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/gaps/defaultSort.ts", "frozen", true)
}
func TestProduct_CycleCurrent_aea045997f89492e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/gaps/defaultSort.ts", "current", true)
}
func TestCyclesCorpus_aea045997f89492e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/gaps/defaultSort.ts")
}
func TestProduct_CycleFrozen_8a5914b4d9b079ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/4_private_method.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8a5914b4d9b079ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/4_private_method.ts", "current", true)
}
func TestCyclesCorpus_8a5914b4d9b079ea(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/values/gaps/4_private_method.ts")
}
func TestProduct_CycleFrozen_7de799fcbe751c49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/nested_rebinding.a", "frozen", true)
}
func TestProduct_CycleCurrent_7de799fcbe751c49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/nested_rebinding.a", "current", true)
}
func TestCyclesCorpus_7de799fcbe751c49(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/refusals/nested_rebinding.a")
}
func TestProduct_CycleFrozen_aa5dc81c5a6466b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_param.a", "frozen", true)
}
func TestProduct_CycleCurrent_aa5dc81c5a6466b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_param.a", "current", true)
}
func TestCyclesCorpus_aa5dc81c5a6466b6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/047cb0d_n_param.a")
}
func TestProduct_CycleFrozen_418266c14f83bdf5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_extended.a", "frozen", true)
}
func TestProduct_CycleCurrent_418266c14f83bdf5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_extended.a", "current", true)
}
func TestCyclesCorpus_418266c14f83bdf5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_extended.a")
}
func TestProduct_CycleFrozen_e20b71538a2aa7ce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_method.a", "frozen", true)
}
func TestProduct_CycleCurrent_e20b71538a2aa7ce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_method.a", "current", true)
}
func TestCyclesCorpus_e20b71538a2aa7ce(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_method.a")
}
func TestProduct_CycleFrozen_c4f08d83d4d33cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_no_reader.a", "frozen", true)
}
func TestProduct_CycleCurrent_c4f08d83d4d33cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_no_reader.a", "current", true)
}
func TestCyclesCorpus_c4f08d83d4d33cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_no_reader.a")
}
func TestProduct_CycleFrozen_0c0cf6c7a95e675f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_from_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_0c0cf6c7a95e675f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_from_undefined.a", "current", true)
}
func TestCyclesCorpus_0c0cf6c7a95e675f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/array_from_undefined.a")
}
func TestProduct_CycleFrozen_1d1b4e5bff08ecee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_frame_capture_safe.a", "frozen", true)
}
func TestProduct_CycleCurrent_1d1b4e5bff08ecee(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_frame_capture_safe.a", "current", true)
}
func TestCyclesCorpus_1d1b4e5bff08ecee(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_frame_capture_safe.a")
}
func TestProduct_CycleFrozen_4009f050b6840b9a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_depths.a", "frozen", true)
}
func TestProduct_CycleCurrent_4009f050b6840b9a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_depths.a", "current", true)
}
func TestCyclesCorpus_4009f050b6840b9a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_coverage_depths.a")
}
func TestProduct_CycleFrozen_82ba489a88e716e5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_82ba489a88e716e5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_fails.a", "current", true)
}
func TestCyclesCorpus_82ba489a88e716e5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_fails.a")
}
func TestProduct_CycleFrozen_d38fb3f26aacb150(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance.a", "frozen", true)
}
func TestProduct_CycleCurrent_d38fb3f26aacb150(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_inheritance.a", "current", true)
}
func TestCyclesCorpus_d38fb3f26aacb150(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_inheritance.a")
}
func TestProduct_CycleFrozen_7158231082db182b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_instance_key_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_7158231082db182b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_instance_key_throw.a", "current", true)
}
func TestCyclesCorpus_7158231082db182b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_instance_key_throw.a")
}
func TestProduct_CycleFrozen_f36028b8805607ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_regexp_count.a", "frozen", true)
}
func TestProduct_CycleCurrent_f36028b8805607ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closure_convention_regexp_count.a", "current", true)
}
func TestCyclesCorpus_f36028b8805607ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/closure_convention_regexp_count.a")
}
func TestProduct_CycleFrozen_4ddf163ab8fb1564(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_optional_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_4ddf163ab8fb1564(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_optional_number.a", "current", true)
}
func TestCyclesCorpus_4ddf163ab8fb1564(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_optional_number.a")
}
func TestProduct_CycleFrozen_20ef06a9470b8560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_regexp_new.a", "frozen", true)
}
func TestProduct_CycleCurrent_20ef06a9470b8560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_regexp_new.a", "current", true)
}
func TestCyclesCorpus_20ef06a9470b8560(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_regexp_new.a")
}
func TestProduct_CycleFrozen_856a1609fb2575ec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_factory.a", "frozen", true)
}
func TestProduct_CycleCurrent_856a1609fb2575ec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/devirtualize_factory.a", "current", true)
}
func TestCyclesCorpus_856a1609fb2575ec(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/devirtualize_factory.a")
}
func TestProduct_CycleFrozen_a4a36291f1a05cba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/doubly_linked.a", "frozen", true)
}
func TestProduct_CycleCurrent_a4a36291f1a05cba(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/doubly_linked.a", "current", true)
}
func TestCyclesCorpus_a4a36291f1a05cba(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/doubly_linked.a")
}
func TestProduct_CycleFrozen_ae1e1eaed79380e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_proven.a", "frozen", true)
}
func TestProduct_CycleCurrent_ae1e1eaed79380e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_proven.a", "current", true)
}
func TestCyclesCorpus_ae1e1eaed79380e4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_proven.a")
}
func TestProduct_CycleFrozen_077eded15d363f32(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_bounded_call.a", "frozen", true)
}
func TestProduct_CycleCurrent_077eded15d363f32(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_bounded_call.a", "current", true)
}
func TestCyclesCorpus_077eded15d363f32(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_bounded_call.a")
}
func TestProduct_CycleFrozen_1e487637475fb6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_returned_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_1e487637475fb6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_coverage_returned_values.a", "current", true)
}
func TestCyclesCorpus_1e487637475fb6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_coverage_returned_values.a")
}
func TestProduct_CycleFrozen_eb13bfd3d674151f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_siblings.a", "frozen", true)
}
func TestProduct_CycleCurrent_eb13bfd3d674151f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_siblings.a", "current", true)
}
func TestCyclesCorpus_eb13bfd3d674151f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_siblings.a")
}
func TestProduct_CycleFrozen_b3cb394c08146ae1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fallthrough_probe.a", "frozen", true)
}
func TestProduct_CycleCurrent_b3cb394c08146ae1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fallthrough_probe.a", "current", true)
}
func TestCyclesCorpus_b3cb394c08146ae1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fallthrough_probe.a")
}
func TestProduct_CycleFrozen_f2ef97136396e512(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/find_index_shrinks.a", "frozen", true)
}
func TestProduct_CycleCurrent_f2ef97136396e512(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/find_index_shrinks.a", "current", true)
}
func TestCyclesCorpus_f2ef97136396e512(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/find_index_shrinks.a")
}
func TestProduct_CycleFrozen_8a37df132672235f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_static_interface.a", "frozen", true)
}
func TestProduct_CycleCurrent_8a37df132672235f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_static_interface.a", "current", true)
}
func TestCyclesCorpus_8a37df132672235f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_static_interface.a")
}
func TestProduct_CycleFrozen_c99c14fbd82ffc9e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/inherited_static_field_read.a", "frozen", true)
}
func TestProduct_CycleCurrent_c99c14fbd82ffc9e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/inherited_static_field_read.a", "current", true)
}
func TestCyclesCorpus_c99c14fbd82ffc9e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/inherited_static_field_read.a")
}
func TestProduct_CycleFrozen_2d79da3fc7fe49b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_indent.a", "frozen", true)
}
func TestProduct_CycleCurrent_2d79da3fc7fe49b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_indent.a", "current", true)
}
func TestCyclesCorpus_2d79da3fc7fe49b0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_indent.a")
}
func TestProduct_CycleFrozen_c7471455e95eefeb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_c7471455e95eefeb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_values.a", "current", true)
}
func TestCyclesCorpus_c7471455e95eefeb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_values.a")
}
func TestProduct_CycleFrozen_6f2438b5572fc1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_copy_within.a", "frozen", true)
}
func TestProduct_CycleCurrent_6f2438b5572fc1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_copy_within.a", "current", true)
}
func TestCyclesCorpus_6f2438b5572fc1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_array_copy_within.a")
}
func TestProduct_CycleFrozen_7da0966b99045624(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_flat.a", "frozen", true)
}
func TestProduct_CycleCurrent_7da0966b99045624(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_array_flat.a", "current", true)
}
func TestCyclesCorpus_7da0966b99045624(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_array_flat.a")
}
func TestProduct_CycleFrozen_046081a244b0dea1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_fnexpr_recurse.a", "frozen", true)
}
func TestProduct_CycleCurrent_046081a244b0dea1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_fnexpr_recurse.a", "current", true)
}
func TestCyclesCorpus_046081a244b0dea1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_fnexpr_recurse.a")
}
func TestProduct_CycleFrozen_93aa79259b4267d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_freeze_assign.a", "frozen", true)
}
func TestProduct_CycleCurrent_93aa79259b4267d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_freeze_assign.a", "current", true)
}
func TestCyclesCorpus_93aa79259b4267d2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_freeze_assign.a")
}
func TestProduct_CycleFrozen_f82a1f1e4c4fbb70(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_keep.a", "frozen", true)
}
func TestProduct_CycleCurrent_f82a1f1e4c4fbb70(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_keep.a", "current", true)
}
func TestCyclesCorpus_f82a1f1e4c4fbb70(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_keep.a")
}
func TestProduct_CycleFrozen_2aabbc2c79165fb3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_filesystem_entries.a", "frozen", true)
}
func TestProduct_CycleCurrent_2aabbc2c79165fb3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_filesystem_entries.a", "current", true)
}
func TestCyclesCorpus_2aabbc2c79165fb3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_union_value_filesystem_entries.a")
}
func TestProduct_CycleFrozen_c5ac1abd480c636c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_source_file.a", "frozen", true)
}
func TestProduct_CycleCurrent_c5ac1abd480c636c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_source_file.a", "current", true)
}
func TestCyclesCorpus_c5ac1abd480c636c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_union_value_source_file.a")
}
func TestProduct_CycleFrozen_c24749e0a8014dfa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_tolowercase.a", "frozen", true)
}
func TestProduct_CycleCurrent_c24749e0a8014dfa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_tolowercase.a", "current", true)
}
func TestCyclesCorpus_c24749e0a8014dfa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_tolowercase.a")
}
func TestProduct_CycleFrozen_5b031aada0e34b77(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/array_owner.a", "frozen", true)
}
func TestProduct_CycleCurrent_5b031aada0e34b77(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/array_owner.a", "current", true)
}
func TestCyclesCorpus_5b031aada0e34b77(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/array_owner.a")
}
func TestProduct_CycleFrozen_24e6878c3e1b15d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/default_parameter.a", "frozen", true)
}
func TestProduct_CycleCurrent_24e6878c3e1b15d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/default_parameter.a", "current", true)
}
func TestCyclesCorpus_24e6878c3e1b15d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/default_parameter.a")
}
func TestProduct_CycleFrozen_968f5987de9fe6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/escape.a", "frozen", true)
}
func TestProduct_CycleCurrent_968f5987de9fe6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/escape.a", "current", true)
}
func TestCyclesCorpus_968f5987de9fe6d7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/escape.a")
}
func TestProduct_CycleFrozen_4003c1fd6775a2fa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_mkdir.a", "frozen", true)
}
func TestProduct_CycleCurrent_4003c1fd6775a2fa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_mkdir.a", "current", true)
}
func TestCyclesCorpus_4003c1fd6775a2fa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_fs_file_mkdir.a")
}
func TestProduct_CycleFrozen_da9554acb234dc0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_options_widening.a", "frozen", true)
}
func TestProduct_CycleCurrent_da9554acb234dc0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_options_widening.a", "current", true)
}
func TestCyclesCorpus_da9554acb234dc0f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_options_widening.a")
}
func TestProduct_CycleFrozen_568032252380e12f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_chain.a", "frozen", true)
}
func TestProduct_CycleCurrent_568032252380e12f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_indexing_chain.a", "current", true)
}
func TestCyclesCorpus_568032252380e12f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_indexing_chain.a")
}
func TestProduct_CycleFrozen_5a08baf46107777d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/panic.a", "frozen", true)
}
func TestProduct_CycleCurrent_5a08baf46107777d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/panic.a", "current", true)
}
func TestCyclesCorpus_5a08baf46107777d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/panic.a")
}
func TestProduct_CycleFrozen_6480e9ea8a9f0e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_shared/left.a", "frozen", true)
}
func TestProduct_CycleCurrent_6480e9ea8a9f0e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_shared/left.a", "current", true)
}
func TestCyclesCorpus_6480e9ea8a9f0e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/params_namespaces_shared/left.a")
}
func TestProduct_CycleFrozen_9ba55f10e2361f26(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/context.a", "frozen", true)
}
func TestProduct_CycleCurrent_9ba55f10e2361f26(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/context.a", "current", true)
}
func TestCyclesCorpus_9ba55f10e2361f26(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/program_region/context.a")
}
func TestProduct_CycleFrozen_fddd56cf28a66253(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/parse.a", "frozen", true)
}
func TestProduct_CycleCurrent_fddd56cf28a66253(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/program_region/parse.a", "current", true)
}
func TestCyclesCorpus_fddd56cf28a66253(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/program_region/parse.a")
}
func TestProduct_CycleFrozen_ae6d74bccb133909(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_exec.a", "frozen", true)
}
func TestProduct_CycleCurrent_ae6d74bccb133909(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_exec.a", "current", true)
}
func TestCyclesCorpus_ae6d74bccb133909(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_exec.a")
}
func TestProduct_CycleFrozen_b580f00a20d9b140(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_b580f00a20d9b140(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/throw.a", "current", true)
}
func TestCyclesCorpus_b580f00a20d9b140(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_replace/throw.a")
}
func TestProduct_CycleFrozen_dcdf022556356969(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_param.a", "frozen", true)
}
func TestProduct_CycleCurrent_dcdf022556356969(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_param.a", "current", true)
}
func TestCyclesCorpus_dcdf022556356969(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_param.a")
}
func TestProduct_CycleFrozen_125d82e76a2ba9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_setadd.a", "frozen", true)
}
func TestProduct_CycleCurrent_125d82e76a2ba9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_setadd.a", "current", true)
}
func TestCyclesCorpus_125d82e76a2ba9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aee89b2_alias_setadd.a")
}
func TestProduct_CycleFrozen_3e57cdf97201ecd7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d7054e9_share_paths.a", "frozen", true)
}
func TestProduct_CycleCurrent_3e57cdf97201ecd7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d7054e9_share_paths.a", "current", true)
}
func TestCyclesCorpus_3e57cdf97201ecd7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d7054e9_share_paths.a")
}
func TestProduct_CycleFrozen_7b6102ddd5b21630(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_finally_concat.a", "frozen", true)
}
func TestProduct_CycleCurrent_7b6102ddd5b21630(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_finally_concat.a", "current", true)
}
func TestCyclesCorpus_7b6102ddd5b21630(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_finally_concat.a")
}
func TestProduct_CycleFrozen_c07458e74a4ba6cb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_stack.a", "frozen", true)
}
func TestProduct_CycleCurrent_c07458e74a4ba6cb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_stack.a", "current", true)
}
func TestCyclesCorpus_c07458e74a4ba6cb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_d96d304_try_stack.a")
}
func TestProduct_CycleFrozen_81ee9178cced4a14(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a", "frozen", true)
}
func TestProduct_CycleCurrent_81ee9178cced4a14(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a", "current", true)
}
func TestCyclesCorpus_81ee9178cced4a14(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p18_cast_kind_no_field_types.a")
}
func TestProduct_CycleFrozen_142cbaa3df31af50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p37_union_field_same_name.a", "frozen", true)
}
func TestProduct_CycleCurrent_142cbaa3df31af50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p37_union_field_same_name.a", "current", true)
}
func TestCyclesCorpus_142cbaa3df31af50(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p37_union_field_same_name.a")
}
func TestProduct_CycleFrozen_d42756e0c9d4e00e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_hidden_return.a", "frozen", true)
}
func TestProduct_CycleCurrent_d42756e0c9d4e00e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_hidden_return.a", "current", true)
}
func TestCyclesCorpus_d42756e0c9d4e00e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_iterators_hidden_return.a")
}
func TestProduct_CycleFrozen_e040cd65e8756138(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_parameters.a", "frozen", true)
}
func TestProduct_CycleCurrent_e040cd65e8756138(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_parameters.a", "current", true)
}
func TestCyclesCorpus_e040cd65e8756138(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/statements_small_parameters.a")
}
func TestProduct_CycleFrozen_54710eef7574a6d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_stopped/structural_error.a", "frozen", true)
}
func TestProduct_CycleCurrent_54710eef7574a6d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/statements_small_stopped/structural_error.a", "current", true)
}
func TestCyclesCorpus_54710eef7574a6d1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/statements_small_stopped/structural_error.a")
}
func TestProduct_CycleFrozen_450772d2d43bcfcc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/static_constructor_spread_owned.a", "frozen", true)
}
func TestProduct_CycleCurrent_450772d2d43bcfcc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/static_constructor_spread_owned.a", "current", true)
}
func TestCyclesCorpus_450772d2d43bcfcc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/static_constructor_spread_owned.a")
}
func TestProduct_CycleFrozen_f030d30e804395ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_instance.a", "frozen", true)
}
func TestProduct_CycleCurrent_f030d30e804395ea(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_instance.a", "current", true)
}
func TestCyclesCorpus_f030d30e804395ea(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/structural_statics_instance.a")
}
func TestProduct_CycleFrozen_1d4276be4d22a8b4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone.a", "frozen", true)
}
func TestProduct_CycleCurrent_1d4276be4d22a8b4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/switch_case_declarations/dead_zone.a", "current", true)
}
func TestCyclesCorpus_1d4276be4d22a8b4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/switch_case_declarations/dead_zone.a")
}
func TestProduct_CycleFrozen_242054e33a7a7f2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/main.a", "frozen", true)
}
func TestProduct_CycleCurrent_242054e33a7a7f2a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_exports/main.a", "current", true)
}
func TestCyclesCorpus_242054e33a7a7f2a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_exports/main.a")
}
func TestProduct_CycleFrozen_ca503f111919e3b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/throw_in_writes.a", "frozen", true)
}
func TestProduct_CycleCurrent_ca503f111919e3b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/throw_in_writes.a", "current", true)
}
func TestCyclesCorpus_ca503f111919e3b0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/throw_in_writes.a")
}
func TestProduct_CycleFrozen_19e54537b4e95ab6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_integer_stop.a", "frozen", true)
}
func TestProduct_CycleCurrent_19e54537b4e95ab6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_integer_stop.a", "current", true)
}
func TestCyclesCorpus_19e54537b4e95ab6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_integer_stop.a")
}
func TestProduct_CycleFrozen_c2e81f86c526ec7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak/narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_c2e81f86c526ec7f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak/narrowed.a", "current", true)
}
func TestCyclesCorpus_c2e81f86c526ec7f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/weak/narrowed.a")
}
func TestProduct_CycleFrozen_07e185fc1a656b6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_files.a", "frozen", true)
}
func TestProduct_CycleCurrent_07e185fc1a656b6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_files.a", "current", true)
}
func TestCyclesCorpus_07e185fc1a656b6e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/write_files.a")
}
func TestProduct_CycleFrozen_b2b4d6a320cc94b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/optional-indexing/map-shape.a", "frozen", true)
}
func TestProduct_CycleCurrent_b2b4d6a320cc94b8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/optional-indexing/map-shape.a", "current", true)
}
func TestCyclesCorpus_b2b4d6a320cc94b8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/optional-indexing/map-shape.a")
}
func TestProduct_CycleFrozen_8cafc445bfe7b397(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/emptyGenerics.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8cafc445bfe7b397(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/emptyGenerics.ts", "current", true)
}
func TestCyclesCorpus_8cafc445bfe7b397(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/emptyGenerics.ts")
}
func TestProduct_CycleFrozen_ed2db6a378ae84b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/pathsweep.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ed2db6a378ae84b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/pathsweep.ts", "current", true)
}
func TestCyclesCorpus_ed2db6a378ae84b1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/pathsweep.ts")
}
func TestProduct_CycleFrozen_8b90b0123f140f18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/gaps/1_nested_constructor.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8b90b0123f140f18(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/gaps/1_nested_constructor.ts", "current", true)
}
func TestCyclesCorpus_8b90b0123f140f18(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/gaps/1_nested_constructor.ts")
}
func TestProduct_CycleFrozen_0cb69d1e01fb82cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/strict_options.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0cb69d1e01fb82cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/strict_options.ts", "current", true)
}
func TestCyclesCorpus_0cb69d1e01fb82cc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/strict_options.ts")
}
func TestProduct_CycleFrozen_ab3f8ab153ee8b32(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/013.a", "frozen", true)
}
func TestProduct_CycleCurrent_ab3f8ab153ee8b32(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/013.a", "current", true)
}
func TestCyclesCorpus_ab3f8ab153ee8b32(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/i/013.a")
}
func TestProduct_CycleFrozen_4eacb554a7a62e0c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_4eacb554a7a62e0c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004_runtime.a", "current", true)
}
func TestCyclesCorpus_4eacb554a7a62e0c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/004_runtime.a")
}
func TestProduct_CycleFrozen_fd3e649f19575fd2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034.a", "frozen", true)
}
func TestProduct_CycleCurrent_fd3e649f19575fd2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034.a", "current", true)
}
func TestCyclesCorpus_fd3e649f19575fd2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034.a")
}
func TestProduct_CycleFrozen_e9917cf26b92d254(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_e9917cf26b92d254(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057_runtime.a", "current", true)
}
func TestCyclesCorpus_e9917cf26b92d254(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/057_runtime.a")
}
func TestProduct_CycleFrozen_654f62cde6f97c20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_654f62cde6f97c20(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071_runtime.a", "current", true)
}
func TestCyclesCorpus_654f62cde6f97c20(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/071_runtime.a")
}
func TestProduct_CycleFrozen_1c2f47ca8598d228(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078.a", "frozen", true)
}
func TestProduct_CycleCurrent_1c2f47ca8598d228(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078.a", "current", true)
}
func TestCyclesCorpus_1c2f47ca8598d228(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078.a")
}
func TestProduct_CycleFrozen_f45b8aa344cc3528(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-consistency-no-hand-built-declared-error/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_f45b8aa344cc3528(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/base-consistency-no-hand-built-declared-error/messages.a", "current", true)
}
func TestCyclesCorpus_f45b8aa344cc3528(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/base-consistency-no-hand-built-declared-error/messages.a")
}
func TestProduct_CycleFrozen_ec74131b5f3c93c0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/patterns.a", "frozen", true)
}
func TestProduct_CycleCurrent_ec74131b5f3c93c0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/patterns.a", "current", true)
}
func TestCyclesCorpus_ec74131b5f3c93c0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/patterns.a")
}
func TestProduct_CycleFrozen_b9c7052ab282abed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-labels/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b9c7052ab282abed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-labels/messages.ts", "current", true)
}
func TestCyclesCorpus_b9c7052ab282abed(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-labels/messages.ts")
}
func TestProduct_CycleFrozen_ba2a4170689f67d9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/emojiMatcher.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ba2a4170689f67d9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/emojiMatcher.ts", "current", true)
}
func TestCyclesCorpus_ba2a4170689f67d9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/emojiMatcher.ts")
}
func TestProduct_CycleFrozen_0a714e86c37ee38f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/15_string_or.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0a714e86c37ee38f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/15_string_or.ts", "current", true)
}
func TestCyclesCorpus_0a714e86c37ee38f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/15_string_or.ts")
}
func TestProduct_CycleFrozen_f918ad52b0608040(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/lowerEntityValues.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f918ad52b0608040(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/lowerEntityValues.ts", "current", true)
}
func TestCyclesCorpus_f918ad52b0608040(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/lowerEntityValues.ts")
}
func TestProduct_CycleFrozen_d98cd2c031ab2fce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/binding.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d98cd2c031ab2fce(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/binding.ts", "current", true)
}
func TestCyclesCorpus_d98cd2c031ab2fce(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/binding.ts")
}
func TestProduct_CycleFrozen_753787888e76c29d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/2_boolean_or_undefined_field.ts", "frozen", true)
}
func TestProduct_CycleCurrent_753787888e76c29d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/values/gaps/2_boolean_or_undefined_field.ts", "current", true)
}
func TestCyclesCorpus_753787888e76c29d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/values/gaps/2_boolean_or_undefined_field.ts")
}
func TestProduct_CycleFrozen_1ce85ec9417ffddc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/assignmentValue.ts", "frozen", true)
}
func TestProduct_CycleCurrent_1ce85ec9417ffddc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/assignmentValue.ts", "current", true)
}
func TestCyclesCorpus_1ce85ec9417ffddc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/gaps/assignmentValue.ts")
}
func TestProduct_CycleFrozen_e185620882a72209(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/scalar_main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e185620882a72209(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/scalar_main.ts", "current", true)
}
func TestCyclesCorpus_e185620882a72209(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/scalar_main.ts")
}
func TestProduct_CycleFrozen_2a5281d9b0789647(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/gaps/6_conditional_empty_array.ts", "frozen", true)
}
func TestProduct_CycleCurrent_2a5281d9b0789647(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/gaps/6_conditional_empty_array.ts", "current", true)
}
func TestCyclesCorpus_2a5281d9b0789647(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/parser/gaps/6_conditional_empty_array.ts")
}
func TestProduct_CycleFrozen_5821ad62743003ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/testdata/type_import_cycle/peer.a", "frozen", true)
}
func TestProduct_CycleCurrent_5821ad62743003ff(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/parser/testdata/type_import_cycle/peer.a", "current", true)
}
func TestCyclesCorpus_5821ad62743003ff(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/parser/testdata/type_import_cycle/peer.a")
}
func TestProduct_CycleFrozen_b7a3511e53300ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/gaps/1_push.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b7a3511e53300ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/gaps/1_push.ts", "current", true)
}
func TestCyclesCorpus_b7a3511e53300ec0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/scanner/gaps/1_push.ts")
}
func TestProduct_CycleFrozen_c0410e7e9961b2bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/scanner.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c0410e7e9961b2bc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/scanner.ts", "current", true)
}
func TestCyclesCorpus_c0410e7e9961b2bc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/scanner/scanner.ts")
}
func TestProduct_CycleFrozen_c32cf9fa1ad07d87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments.a", "frozen", true)
}
func TestProduct_CycleCurrent_c32cf9fa1ad07d87(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments.a", "current", true)
}
func TestCyclesCorpus_c32cf9fa1ad07d87(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments.a")
}
func TestProduct_CycleFrozen_5fb3a7dc7a12913c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_from.a", "frozen", true)
}
func TestProduct_CycleCurrent_5fb3a7dc7a12913c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_from.a", "current", true)
}
func TestCyclesCorpus_5fb3a7dc7a12913c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/array_from.a")
}
func TestProduct_CycleFrozen_60b28c7a82b8edb7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_catches.a", "frozen", true)
}
func TestProduct_CycleCurrent_60b28c7a82b8edb7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_catches.a", "current", true)
}
func TestCyclesCorpus_60b28c7a82b8edb7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_catches.a")
}
func TestProduct_CycleFrozen_6dd6a1eff99ef419(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_parents.a", "frozen", true)
}
func TestProduct_CycleCurrent_6dd6a1eff99ef419(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_coverage_parents.a", "current", true)
}
func TestCyclesCorpus_6dd6a1eff99ef419(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_coverage_parents.a")
}
func TestProduct_CycleFrozen_8157dff1794a2dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_binders.a", "frozen", true)
}
func TestProduct_CycleCurrent_8157dff1794a2dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_binders.a", "current", true)
}
func TestCyclesCorpus_8157dff1794a2dbd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_overload_binders.a")
}
func TestProduct_CycleFrozen_9f1972bf677aee8e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_overload_correlated_result.a", "frozen", true)
}
func TestProduct_CycleCurrent_9f1972bf677aee8e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_overload_correlated_result.a", "current", true)
}
func TestCyclesCorpus_9f1972bf677aee8e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_stopped/census_overload_correlated_result.a")
}
func TestProduct_CycleFrozen_39b23fee3057e33e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/readonly_forms.a", "frozen", true)
}
func TestProduct_CycleCurrent_39b23fee3057e33e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/readonly_forms.a", "current", true)
}
func TestCyclesCorpus_39b23fee3057e33e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/readonly_forms.a")
}
func TestProduct_CycleFrozen_e484eb1a055dba97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/directory.a", "frozen", true)
}
func TestProduct_CycleCurrent_e484eb1a055dba97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/directory.a", "current", true)
}
func TestCyclesCorpus_e484eb1a055dba97(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/directory.a")
}
func TestProduct_CycleFrozen_99cce5c182093c06(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_union.a", "frozen", true)
}
func TestProduct_CycleCurrent_99cce5c182093c06(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_union.a", "current", true)
}
func TestCyclesCorpus_99cce5c182093c06(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_union.a")
}
func TestProduct_CycleFrozen_c99d757a23e8c287(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_no_argument.a", "frozen", true)
}
func TestProduct_CycleCurrent_c99d757a23e8c287(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_no_argument.a", "current", true)
}
func TestCyclesCorpus_c99d757a23e8c287(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_uncaught_no_argument.a")
}
func TestProduct_CycleFrozen_c1c227ef67672124(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_string.a", "frozen", true)
}
func TestProduct_CycleCurrent_c1c227ef67672124(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_view_string.a", "current", true)
}
func TestCyclesCorpus_c1c227ef67672124(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_view_string.a")
}
func TestProduct_CycleFrozen_70d03c8bb0641ce6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/defaults.a", "frozen", true)
}
func TestProduct_CycleCurrent_70d03c8bb0641ce6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/defaults.a", "current", true)
}
func TestCyclesCorpus_70d03c8bb0641ce6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/defaults.a")
}
func TestProduct_CycleFrozen_237a84b45137b65a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_optional.a", "frozen", true)
}
func TestProduct_CycleCurrent_237a84b45137b65a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_optional.a", "current", true)
}
func TestCyclesCorpus_237a84b45137b65a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/element_access_optional.a")
}
func TestProduct_CycleFrozen_92988454d73e9066(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_tuple.a", "frozen", true)
}
func TestProduct_CycleCurrent_92988454d73e9066(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_tuple.a", "current", true)
}
func TestCyclesCorpus_92988454d73e9066(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/element_access_tuple.a")
}
func TestProduct_CycleFrozen_221706dd6484db92(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_capture.a", "frozen", true)
}
func TestProduct_CycleCurrent_221706dd6484db92(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_capture.a", "current", true)
}
func TestCyclesCorpus_221706dd6484db92(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_capture.a")
}
func TestProduct_CycleFrozen_6410134685becf95(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_virtual_call.a", "frozen", true)
}
func TestProduct_CycleCurrent_6410134685becf95(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/environment_virtual_call.a", "current", true)
}
func TestCyclesCorpus_6410134685becf95(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/environment_virtual_call.a")
}
func TestProduct_CycleFrozen_136381cefd2fb8e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_leaks_parent.a", "frozen", true)
}
func TestProduct_CycleCurrent_136381cefd2fb8e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/call_leaks_parent.a", "current", true)
}
func TestCyclesCorpus_136381cefd2fb8e8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/call_leaks_parent.a")
}
func TestProduct_CycleFrozen_8bbef450a8e27cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a", "frozen", true)
}
func TestProduct_CycleCurrent_8bbef450a8e27cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a", "current", true)
}
func TestCyclesCorpus_8bbef450a8e27cb1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/devirt_fresh_method_keeps_argument.a")
}
func TestProduct_CycleFrozen_80a9d519febbeaeb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/set_add.a", "frozen", true)
}
func TestProduct_CycleCurrent_80a9d519febbeaeb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/set_add.a", "current", true)
}
func TestCyclesCorpus_80a9d519febbeaeb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/set_add.a")
}
func TestProduct_CycleFrozen_8adb2158d4fc221f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_boundary.a", "frozen", true)
}
func TestProduct_CycleCurrent_8adb2158d4fc221f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/function_values_boundary.a", "current", true)
}
func TestCyclesCorpus_8adb2158d4fc221f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/function_values_boundary.a")
}
func TestProduct_CycleFrozen_dc7adb8693de6a7c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_primary_function.a", "frozen", true)
}
func TestProduct_CycleCurrent_dc7adb8693de6a7c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_primary_function.a", "current", true)
}
func TestCyclesCorpus_dc7adb8693de6a7c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_primary_function.a")
}
func TestProduct_CycleFrozen_2cd72cd2cbc9dba7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions/classification_override.a", "frozen", true)
}
func TestProduct_CycleCurrent_2cd72cd2cbc9dba7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions/classification_override.a", "current", true)
}
func TestCyclesCorpus_2cd72cd2cbc9dba7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions/classification_override.a")
}
func TestProduct_CycleFrozen_7dde0023d85d70b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_scanner_never.a", "frozen", true)
}
func TestProduct_CycleCurrent_7dde0023d85d70b1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_scanner_never.a", "current", true)
}
func TestCyclesCorpus_7dde0023d85d70b1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/host_scanner_never.a")
}
func TestProduct_CycleFrozen_ca0942863351133c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/read/a.a", "frozen", true)
}
func TestProduct_CycleCurrent_ca0942863351133c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/read/a.a", "current", true)
}
func TestCyclesCorpus_ca0942863351133c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/read/a.a")
}
func TestProduct_CycleFrozen_bd9cd6050a8169f1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_globals.a", "frozen", true)
}
func TestProduct_CycleCurrent_bd9cd6050a8169f1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_globals.a", "current", true)
}
func TestCyclesCorpus_bd9cd6050a8169f1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_globals.a")
}
func TestProduct_CycleFrozen_b1645360ab94b304(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_keys.a", "frozen", true)
}
func TestProduct_CycleCurrent_b1645360ab94b304(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_keys.a", "current", true)
}
func TestCyclesCorpus_b1645360ab94b304(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_keys.a")
}
func TestProduct_CycleFrozen_3f073be6153d46e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_keys.a", "frozen", true)
}
func TestProduct_CycleCurrent_3f073be6153d46e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_keys.a", "current", true)
}
func TestCyclesCorpus_3f073be6153d46e6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_keys.a")
}
func TestProduct_CycleFrozen_ad96ac8a902142be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/logical_and_reference_maybe.a", "frozen", true)
}
func TestProduct_CycleCurrent_ad96ac8a902142be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/logical_and_reference_maybe.a", "current", true)
}
func TestCyclesCorpus_ad96ac8a902142be(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/logical_and_reference_maybe.a")
}
func TestProduct_CycleFrozen_37749d6778d05e99(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_charat.a", "frozen", true)
}
func TestProduct_CycleCurrent_37749d6778d05e99(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_charat.a", "current", true)
}
func TestCyclesCorpus_37749d6778d05e99(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_charat.a")
}
func TestProduct_CycleFrozen_c0b84dd19e3d82ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/early.a", "frozen", true)
}
func TestProduct_CycleCurrent_c0b84dd19e3d82ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/module_namespace_reads/early.a", "current", true)
}
func TestCyclesCorpus_c0b84dd19e3d82ae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/module_namespace_reads/early.a")
}
func TestProduct_CycleFrozen_7eb5a9d4f2de0a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested_arrays.a", "frozen", true)
}
func TestProduct_CycleCurrent_7eb5a9d4f2de0a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested_arrays.a", "current", true)
}
func TestCyclesCorpus_7eb5a9d4f2de0a5c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/nested_arrays.a")
}
func TestProduct_CycleFrozen_a065c9111725a36a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nbody_spread_fields.a", "frozen", true)
}
func TestProduct_CycleCurrent_a065c9111725a36a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nbody_spread_fields.a", "current", true)
}
func TestCyclesCorpus_a065c9111725a36a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nbody_spread_fields.a")
}
func TestProduct_CycleFrozen_bf63b46a87c7e3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_callback_recursive.a", "frozen", true)
}
func TestProduct_CycleCurrent_bf63b46a87c7e3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_callback_recursive.a", "current", true)
}
func TestCyclesCorpus_bf63b46a87c7e3cf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_callback_recursive.a")
}
func TestProduct_CycleFrozen_f3ed5050f30c6dd2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_destructured.a", "frozen", true)
}
func TestProduct_CycleCurrent_f3ed5050f30c6dd2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_destructured.a", "current", true)
}
func TestCyclesCorpus_f3ed5050f30c6dd2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_destructured.a")
}
func TestProduct_CycleFrozen_6e110e3dbbd2ef68(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_literal_return.ts", "frozen", true)
}
func TestProduct_CycleCurrent_6e110e3dbbd2ef68(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_literal_return.ts", "current", true)
}
func TestCyclesCorpus_6e110e3dbbd2ef68(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_literal_return.ts")
}
func TestProduct_CycleFrozen_f6e6078b272ad208(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_static_uninitialized.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f6e6078b272ad208(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_static_uninitialized.ts", "current", true)
}
func TestCyclesCorpus_f6e6078b272ad208(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_static_uninitialized.ts")
}
func TestProduct_CycleFrozen_74230c0e94167dac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_quick.a", "frozen", true)
}
func TestProduct_CycleCurrent_74230c0e94167dac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_quick.a", "current", true)
}
func TestCyclesCorpus_74230c0e94167dac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/normalize_coverage_quick.a")
}
func TestProduct_CycleFrozen_2ccc03d175000f05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_repeat.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ccc03d175000f05(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/normalize_coverage_repeat.a", "current", true)
}
func TestCyclesCorpus_2ccc03d175000f05(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/normalize_coverage_repeat.a")
}
func TestProduct_CycleFrozen_8ef14c3b097272ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/string_dictionary.a", "frozen", true)
}
func TestProduct_CycleCurrent_8ef14c3b097272ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/string_dictionary.a", "current", true)
}
func TestCyclesCorpus_8ef14c3b097272ae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_element_access/string_dictionary.a")
}
func TestProduct_CycleFrozen_15beac9fd4382afb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_mutate_map.a", "frozen", true)
}
func TestProduct_CycleCurrent_15beac9fd4382afb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_mutate_map.a", "current", true)
}
func TestCyclesCorpus_15beac9fd4382afb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_mutate_map.a")
}
func TestProduct_CycleFrozen_3daafdb1e77104eb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/ownership-query/flat.a", "frozen", true)
}
func TestProduct_CycleCurrent_3daafdb1e77104eb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/ownership-query/flat.a", "current", true)
}
func TestCyclesCorpus_3daafdb1e77104eb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/ownership-query/flat.a")
}
func TestProduct_CycleFrozen_26177b6d18b001e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_assertions.a", "frozen", true)
}
func TestProduct_CycleCurrent_26177b6d18b001e4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_assertions.a", "current", true)
}
func TestCyclesCorpus_26177b6d18b001e4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/proven_assertions.a")
}
func TestProduct_CycleFrozen_939725234afa18c9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_cycle_fields.a", "frozen", true)
}
func TestProduct_CycleCurrent_939725234afa18c9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_cycle_fields.a", "current", true)
}
func TestCyclesCorpus_939725234afa18c9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_cycle_fields.a")
}
func TestProduct_CycleFrozen_6ee601df0fd8deb3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_spread_fresh.a", "frozen", true)
}
func TestProduct_CycleCurrent_6ee601df0fd8deb3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions_spread_fresh.a", "current", true)
}
func TestCyclesCorpus_6ee601df0fd8deb3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regions_spread_fresh.a")
}
func TestProduct_CycleFrozen_42bd88281ba8cb86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/replace_all_large.a", "frozen", true)
}
func TestProduct_CycleCurrent_42bd88281ba8cb86(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/replace_all_large.a", "current", true)
}
func TestCyclesCorpus_42bd88281ba8cb86(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/replace_all_large.a")
}
func TestProduct_CycleFrozen_427471dc86de5be1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_escapes.a", "frozen", true)
}
func TestProduct_CycleCurrent_427471dc86de5be1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_escapes.a", "current", true)
}
func TestCyclesCorpus_427471dc86de5be1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_8e67677_class_escapes.a")
}
func TestProduct_CycleFrozen_f84aee835f3524c2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_maybe_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_f84aee835f3524c2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_maybe_number.a", "current", true)
}
func TestCyclesCorpus_f84aee835f3524c2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p38_same_name_maybe_number.a")
}
func TestProduct_CycleFrozen_8f91198090ec3618(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/search_halves.a", "frozen", true)
}
func TestProduct_CycleCurrent_8f91198090ec3618(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/search_halves.a", "current", true)
}
func TestCyclesCorpus_8f91198090ec3618(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/search_halves.a")
}
func TestProduct_CycleFrozen_af35cc5095917dd5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_numbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_af35cc5095917dd5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_numbers.a", "current", true)
}
func TestCyclesCorpus_af35cc5095917dd5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_numbers.a")
}
func TestProduct_CycleFrozen_08396432edabce0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/stack_tail_call.a", "frozen", true)
}
func TestProduct_CycleCurrent_08396432edabce0f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/stack_tail_call.a", "current", true)
}
func TestCyclesCorpus_08396432edabce0f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/stack_tail_call.a")
}
func TestProduct_CycleFrozen_ccf61fa87c4f6c74(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_caches.a", "frozen", true)
}
func TestProduct_CycleCurrent_ccf61fa87c4f6c74(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_build_caches.a", "current", true)
}
func TestCyclesCorpus_ccf61fa87c4f6c74(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_build_caches.a")
}
func TestProduct_CycleFrozen_e3f8a4886ffc967d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_inherited.a", "frozen", true)
}
func TestProduct_CycleCurrent_e3f8a4886ffc967d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_inherited.a", "current", true)
}
func TestCyclesCorpus_e3f8a4886ffc967d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/structural_statics_inherited.a")
}
func TestProduct_CycleFrozen_ac92dc4c351cf1cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_optional.a", "frozen", true)
}
func TestProduct_CycleCurrent_ac92dc4c351cf1cc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/structural_statics_optional.a", "current", true)
}
func TestCyclesCorpus_ac92dc4c351cf1cc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/structural_statics_optional.a")
}
func TestProduct_CycleFrozen_02e8ef19e90068d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_stage3_representations.a", "frozen", true)
}
func TestProduct_CycleCurrent_02e8ef19e90068d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_stage3_representations.a", "current", true)
}
func TestCyclesCorpus_02e8ef19e90068d1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_stage3_representations.a")
}
func TestProduct_CycleFrozen_7cbb32b627781855(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_views.a", "frozen", true)
}
func TestProduct_CycleCurrent_7cbb32b627781855(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_views.a", "current", true)
}
func TestCyclesCorpus_7cbb32b627781855(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_views.a")
}
func TestProduct_CycleFrozen_4b53c06fa3229c56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_after_shrink.a", "frozen", true)
}
func TestProduct_CycleCurrent_4b53c06fa3229c56(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/write_after_shrink.a", "current", true)
}
func TestCyclesCorpus_4b53c06fa3229c56(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/write_after_shrink.a")
}
func TestProduct_CycleFrozen_25b6b4addaf0956b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/jsxEntities.ts", "frozen", true)
}
func TestProduct_CycleCurrent_25b6b4addaf0956b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/jsxEntities.ts", "current", true)
}
func TestCyclesCorpus_25b6b4addaf0956b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/jsxEntities.ts")
}
func TestProduct_CycleFrozen_c020c0b62e1823c9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/013_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_c020c0b62e1823c9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/013_runtime.a", "current", true)
}
func TestCyclesCorpus_c020c0b62e1823c9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/i/013_runtime.a")
}
func TestProduct_CycleFrozen_970ab967ecf36fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_970ab967ecf36fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034_runtime.a", "current", true)
}
func TestCyclesCorpus_970ab967ecf36fe6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/034_runtime.a")
}
func TestProduct_CycleFrozen_9b7a2e0dd7cdd59d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/035.a", "frozen", true)
}
func TestProduct_CycleCurrent_9b7a2e0dd7cdd59d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/035.a", "current", true)
}
func TestCyclesCorpus_9b7a2e0dd7cdd59d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/035.a")
}
func TestProduct_CycleFrozen_02dab9a51d41e284(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_02dab9a51d41e284(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078_runtime.a", "current", true)
}
func TestCyclesCorpus_02dab9a51d41e284(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/078_runtime.a")
}
func TestProduct_CycleFrozen_83cf4d2c957d8bfd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/079.a", "frozen", true)
}
func TestProduct_CycleCurrent_83cf4d2c957d8bfd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/079.a", "current", true)
}
func TestCyclesCorpus_83cf4d2c957d8bfd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/079.a")
}
func TestProduct_CycleFrozen_9d0de8e9c24c2847(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-depth/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_9d0de8e9c24c2847(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/max-depth/messages.a", "current", true)
}
func TestCyclesCorpus_9d0de8e9c24c2847(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/max-depth/messages.a")
}
func TestProduct_CycleFrozen_e1b28d0220745d45(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_e1b28d0220745d45(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/messages.a", "current", true)
}
func TestCyclesCorpus_e1b28d0220745d45(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-shouting/messages.a")
}
func TestProduct_CycleFrozen_01ad9f5f62466e1a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_01ad9f5f62466e1a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/rule.a", "current", true)
}
func TestCyclesCorpus_01ad9f5f62466e1a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/nexus-consistency-no-single-line-jsdoc/rule.a")
}
func TestProduct_CycleFrozen_4d3552d5f532b171(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-character-class/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_4d3552d5f532b171(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-empty-character-class/rule.a", "current", true)
}
func TestCyclesCorpus_4d3552d5f532b171(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-empty-character-class/rule.a")
}
func TestProduct_CycleFrozen_d312f3b27cf8297a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d312f3b27cf8297a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts", "current", true)
}
func TestCyclesCorpus_d312f3b27cf8297a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/react-jsx-no-comment-textnodes/rule.ts")
}
func TestProduct_CycleFrozen_66812fb0592be620(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-method-signature-style/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_66812fb0592be620(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-method-signature-style/rule.ts", "current", true)
}
func TestCyclesCorpus_66812fb0592be620(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-method-signature-style/rule.ts")
}
func TestProduct_CycleFrozen_7bdf84879daf8500(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/structure.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7bdf84879daf8500(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/structure.ts", "current", true)
}
func TestCyclesCorpus_7bdf84879daf8500(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/structure.ts")
}
func TestProduct_CycleFrozen_af8dfc794969cf6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/whitespace.ts", "frozen", true)
}
func TestProduct_CycleCurrent_af8dfc794969cf6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/whitespace.ts", "current", true)
}
func TestCyclesCorpus_af8dfc794969cf6c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/whitespace.ts")
}
func TestProduct_CycleFrozen_8352e67887f9c09d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/widthTables.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8352e67887f9c09d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/widthTables.ts", "current", true)
}
func TestCyclesCorpus_8352e67887f9c09d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/widthTables.ts")
}
func TestProduct_CycleFrozen_ab5cdf77db645cb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/gaps/1_mutual_recursion.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ab5cdf77db645cb8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/suppression/gaps/1_mutual_recursion.ts", "current", true)
}
func TestCyclesCorpus_ab5cdf77db645cb8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/suppression/gaps/1_mutual_recursion.ts")
}
func TestProduct_CycleFrozen_cc6f7e1c95eb5ca7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/transport.ts", "frozen", true)
}
func TestProduct_CycleCurrent_cc6f7e1c95eb5ca7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/transport.ts", "current", true)
}
func TestCyclesCorpus_cc6f7e1c95eb5ca7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/transport.ts")
}
func TestProduct_CycleFrozen_17b8fce6cfd97e7e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/completion.ts", "frozen", true)
}
func TestProduct_CycleCurrent_17b8fce6cfd97e7e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/completion.ts", "current", true)
}
func TestCyclesCorpus_17b8fce6cfd97e7e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/completion.ts")
}
func TestProduct_CycleFrozen_fb883aecbdcdda09(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/unary_minus.ts", "frozen", true)
}
func TestProduct_CycleCurrent_fb883aecbdcdda09(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/unary_minus.ts", "current", true)
}
func TestCyclesCorpus_fb883aecbdcdda09(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/unary_minus.ts")
}
func TestProduct_CycleFrozen_4e0b8fe054d079cb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_typeof.a", "frozen", true)
}
func TestProduct_CycleCurrent_4e0b8fe054d079cb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/047cb0d_n_typeof.a", "current", true)
}
func TestCyclesCorpus_4e0b8fe054d079cb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/047cb0d_n_typeof.a")
}
func TestProduct_CycleFrozen_d21581cd77b6ff5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_d21581cd77b6ff5a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/arguments_length_spread.a", "current", true)
}
func TestCyclesCorpus_d21581cd77b6ff5a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/arguments_length_spread.a")
}
func TestProduct_CycleFrozen_e18f625e252be8a4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_closures.a", "frozen", true)
}
func TestProduct_CycleCurrent_e18f625e252be8a4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_closures.a", "current", true)
}
func TestCyclesCorpus_e18f625e252be8a4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_closures.a")
}
func TestProduct_CycleFrozen_75687ed22eced856(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/unawaited_rejection.a", "frozen", true)
}
func TestProduct_CycleCurrent_75687ed22eced856(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/unawaited_rejection.a", "current", true)
}
func TestCyclesCorpus_75687ed22eced856(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_review/refused/unawaited_rejection.a")
}
func TestProduct_CycleFrozen_53ff20ba195e3942(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_captured_root.a", "frozen", true)
}
func TestProduct_CycleCurrent_53ff20ba195e3942(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_chain_captured_root.a", "current", true)
}
func TestCyclesCorpus_53ff20ba195e3942(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_chain_captured_root.a")
}
func TestProduct_CycleFrozen_22c73a7e23c3137d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_const.a", "frozen", true)
}
func TestProduct_CycleCurrent_22c73a7e23c3137d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_const.a", "current", true)
}
func TestCyclesCorpus_22c73a7e23c3137d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_enum_const.a")
}
func TestProduct_CycleFrozen_de12862b5b408a93(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_small_default_stopped.a", "frozen", true)
}
func TestProduct_CycleCurrent_de12862b5b408a93(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_small_default_stopped.a", "current", true)
}
func TestCyclesCorpus_de12862b5b408a93(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_stopped/census_small_default_stopped.a")
}
func TestProduct_CycleFrozen_763f6bae208d2074(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_oct6_subclass_holder.a", "frozen", true)
}
func TestProduct_CycleCurrent_763f6bae208d2074(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_oct6_subclass_holder.a", "current", true)
}
func TestCyclesCorpus_763f6bae208d2074(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_oct6_subclass_holder.a")
}
func TestProduct_CycleFrozen_9691ef79af0c1729(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closures_throw_uncaught.a", "frozen", true)
}
func TestProduct_CycleCurrent_9691ef79af0c1729(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/closures_throw_uncaught.a", "current", true)
}
func TestCyclesCorpus_9691ef79af0c1729(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/closures_throw_uncaught.a")
}
func TestProduct_CycleFrozen_6efaae8aabb13139(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/file_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_6efaae8aabb13139(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/file_write.a", "current", true)
}
func TestCyclesCorpus_6efaae8aabb13139(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/file_write.a")
}
func TestProduct_CycleFrozen_4ce8fcc4abae0753(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_evaluation.a", "frozen", true)
}
func TestProduct_CycleCurrent_4ce8fcc4abae0753(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_evaluation.a", "current", true)
}
func TestCyclesCorpus_4ce8fcc4abae0753(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_evaluation.a")
}
func TestProduct_CycleFrozen_ceae4db403d8bc6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_scanner.a", "frozen", true)
}
func TestProduct_CycleCurrent_ceae4db403d8bc6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_scanner.a", "current", true)
}
func TestCyclesCorpus_ceae4db403d8bc6a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_scanner.a")
}
func TestProduct_CycleFrozen_cd0963c5359996ef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/returns_argument.a", "frozen", true)
}
func TestProduct_CycleCurrent_cd0963c5359996ef(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/returns_argument.a", "current", true)
}
func TestCyclesCorpus_cd0963c5359996ef(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/returns_argument.a")
}
func TestProduct_CycleFrozen_ae3a23c38d0a3760(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_literals.a", "frozen", true)
}
func TestProduct_CycleCurrent_ae3a23c38d0a3760(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_literals.a", "current", true)
}
func TestCyclesCorpus_ae3a23c38d0a3760(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_literals.a")
}
func TestProduct_CycleFrozen_45111f8101125779(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_nested_disjoint.a", "frozen", true)
}
func TestProduct_CycleCurrent_45111f8101125779(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_nested_disjoint.a", "current", true)
}
func TestCyclesCorpus_45111f8101125779(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_nested_disjoint.a")
}
func TestProduct_CycleFrozen_e820b60b95fc62cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_15.a", "frozen", true)
}
func TestProduct_CycleCurrent_e820b60b95fc62cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_15.a", "current", true)
}
func TestCyclesCorpus_e820b60b95fc62cf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_regression_15.a")
}
func TestProduct_CycleFrozen_69000afc42d57887(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_array_unknown_predicate.a", "frozen", true)
}
func TestProduct_CycleCurrent_69000afc42d57887(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/host_array_unknown_predicate.a", "current", true)
}
func TestCyclesCorpus_69000afc42d57887(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/host_array_unknown_predicate.a")
}
func TestProduct_CycleFrozen_1177b15f5f5667d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/hoisted/b.a", "frozen", true)
}
func TestProduct_CycleCurrent_1177b15f5f5667d4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/hoisted/b.a", "current", true)
}
func TestCyclesCorpus_1177b15f5f5667d4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/hoisted/b.a")
}
func TestProduct_CycleFrozen_4fc9554365c8884f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/input_spread_ordinary.a", "frozen", true)
}
func TestProduct_CycleCurrent_4fc9554365c8884f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/input_spread_ordinary.a", "current", true)
}
func TestCyclesCorpus_4fc9554365c8884f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/input_spread_ordinary.a")
}
func TestProduct_CycleFrozen_2d14517815b8a77c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_2d14517815b8a77c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_undefined.a", "current", true)
}
func TestCyclesCorpus_2d14517815b8a77c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_undefined.a")
}
func TestProduct_CycleFrozen_77125a9de1027fc7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_assign_fields.a", "frozen", true)
}
func TestProduct_CycleCurrent_77125a9de1027fc7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_object_assign_fields.a", "current", true)
}
func TestCyclesCorpus_77125a9de1027fc7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_object_assign_fields.a")
}
func TestProduct_CycleFrozen_74437dc3a5a5b993(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_string_indices.a", "frozen", true)
}
func TestProduct_CycleCurrent_74437dc3a5a5b993(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_string_indices.a", "current", true)
}
func TestCyclesCorpus_74437dc3a5a5b993(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_string_indices.a")
}
func TestProduct_CycleFrozen_362bf31e7adafdcb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/maps_and_text.a", "frozen", true)
}
func TestProduct_CycleCurrent_362bf31e7adafdcb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/maps_and_text.a", "current", true)
}
func TestCyclesCorpus_362bf31e7adafdcb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/maps_and_text.a")
}
func TestProduct_CycleFrozen_ed8326a4ee69da49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_math.a", "frozen", true)
}
func TestProduct_CycleCurrent_ed8326a4ee69da49(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_math.a", "current", true)
}
func TestCyclesCorpus_ed8326a4ee69da49(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_math.a")
}
func TestProduct_CycleFrozen_595d6a7ae826820a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_number_formats.a", "frozen", true)
}
func TestProduct_CycleCurrent_595d6a7ae826820a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_number_formats.a", "current", true)
}
func TestCyclesCorpus_595d6a7ae826820a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_number_formats.a")
}
func TestProduct_CycleFrozen_b1faf0270be3c292(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_charcodeat.a", "frozen", true)
}
func TestProduct_CycleCurrent_b1faf0270be3c292(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_charcodeat.a", "current", true)
}
func TestCyclesCorpus_b1faf0270be3c292(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_charcodeat.a")
}
func TestProduct_CycleFrozen_942caefd15287d21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_map_before.a", "frozen", true)
}
func TestProduct_CycleCurrent_942caefd15287d21(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_map_before.a", "current", true)
}
func TestCyclesCorpus_942caefd15287d21(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_map_before.a")
}
func TestProduct_CycleFrozen_be33bee65fb549c7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_repeated_var.a", "frozen", true)
}
func TestProduct_CycleCurrent_be33bee65fb549c7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_repeated_var.a", "current", true)
}
func TestCyclesCorpus_be33bee65fb549c7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_repeated_var.a")
}
func TestProduct_CycleFrozen_f92f2cb1483308dd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_lazy_field.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f92f2cb1483308dd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_lazy_field.ts", "current", true)
}
func TestCyclesCorpus_f92f2cb1483308dd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_lazy_field.ts")
}
func TestProduct_CycleFrozen_bee3807f8782a06e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_bee3807f8782a06e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_undefined.a", "current", true)
}
func TestCyclesCorpus_bee3807f8782a06e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_undefined.a")
}
func TestProduct_CycleFrozen_1af062e784796e13(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_union.ts", "frozen", true)
}
func TestProduct_CycleCurrent_1af062e784796e13(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_union.ts", "current", true)
}
func TestCyclesCorpus_1af062e784796e13(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_union.ts")
}
func TestProduct_CycleFrozen_985d866b3cd8b886(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_boolean.a", "frozen", true)
}
func TestProduct_CycleCurrent_985d866b3cd8b886(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/omitted_boolean.a", "current", true)
}
func TestCyclesCorpus_985d866b3cd8b886(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/omitted_boolean.a")
}
func TestProduct_CycleFrozen_e167ed46db5b527e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/output_then_panic.a", "frozen", true)
}
func TestProduct_CycleCurrent_e167ed46db5b527e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/output_then_panic.a", "current", true)
}
func TestCyclesCorpus_e167ed46db5b527e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/output_then_panic.a")
}
func TestProduct_CycleFrozen_c193d9ee3f5f9cd8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_mutate_map_skipping_new.a", "frozen", true)
}
func TestProduct_CycleCurrent_c193d9ee3f5f9cd8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/overload_mutate_map_skipping_new.a", "current", true)
}
func TestCyclesCorpus_c193d9ee3f5f9cd8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/overload_mutate_map_skipping_new.a")
}
func TestProduct_CycleFrozen_be5716dc452c8208(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/parameter_properties_ownership.a", "frozen", true)
}
func TestProduct_CycleCurrent_be5716dc452c8208(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/parameter_properties_ownership.a", "current", true)
}
func TestCyclesCorpus_be5716dc452c8208(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/parameter_properties_ownership.a")
}
func TestProduct_CycleFrozen_cd5300ac4bb3fd5b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_overrides.a", "frozen", true)
}
func TestProduct_CycleCurrent_cd5300ac4bb3fd5b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/params_namespaces_overrides.a", "current", true)
}
func TestCyclesCorpus_cd5300ac4bb3fd5b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/params_namespaces_overrides.a")
}
func TestProduct_CycleFrozen_a6d59a5966a44ab0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions.a", "frozen", true)
}
func TestProduct_CycleCurrent_a6d59a5966a44ab0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regions.a", "current", true)
}
func TestCyclesCorpus_a6d59a5966a44ab0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regions.a")
}
func TestProduct_CycleFrozen_3368f670dcd8ad58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_typeof.a", "frozen", true)
}
func TestProduct_CycleCurrent_3368f670dcd8ad58(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_typeof.a", "current", true)
}
func TestCyclesCorpus_3368f670dcd8ad58(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_047cb0d_n_typeof.a")
}
func TestProduct_CycleFrozen_aeb5eb7e18856eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_comparator_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_aeb5eb7e18856eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_comparator_spread.a", "current", true)
}
func TestCyclesCorpus_aeb5eb7e18856eb1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_aa17d3c_comparator_spread.a")
}
func TestProduct_CycleFrozen_b7324e622d0949dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_maybe_setter.a", "frozen", true)
}
func TestProduct_CycleCurrent_b7324e622d0949dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_maybe_setter.a", "current", true)
}
func TestCyclesCorpus_b7324e622d0949dc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_classfeat_maybe_setter.a")
}
func TestProduct_CycleFrozen_bfc9d3068860f63d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_n.a", "frozen", true)
}
func TestProduct_CycleCurrent_bfc9d3068860f63d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_n.a", "current", true)
}
func TestCyclesCorpus_bfc9d3068860f63d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p54_maybe_n.a")
}
func TestProduct_CycleFrozen_202a4e8af8fde258(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p59_map_null_write.a", "frozen", true)
}
func TestProduct_CycleCurrent_202a4e8af8fde258(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p59_map_null_write.a", "current", true)
}
func TestCyclesCorpus_202a4e8af8fde258(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p59_map_null_write.a")
}
func TestProduct_CycleFrozen_04f85a889a6f9908(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q15_every_optional_string.a", "frozen", true)
}
func TestProduct_CycleCurrent_04f85a889a6f9908(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q15_every_optional_string.a", "current", true)
}
func TestCyclesCorpus_04f85a889a6f9908(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q15_every_optional_string.a")
}
func TestProduct_CycleFrozen_56efbe22b5e4cc71(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q19_union_literal_index.a", "frozen", true)
}
func TestProduct_CycleCurrent_56efbe22b5e4cc71(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q19_union_literal_index.a", "current", true)
}
func TestCyclesCorpus_56efbe22b5e4cc71(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q19_union_literal_index.a")
}
func TestProduct_CycleFrozen_935f1c5ae55cd6cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p23_callable_union_literal_result.a", "frozen", true)
}
func TestProduct_CycleCurrent_935f1c5ae55cd6cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p23_callable_union_literal_result.a", "current", true)
}
func TestCyclesCorpus_935f1c5ae55cd6cd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p23_callable_union_literal_result.a")
}
func TestProduct_CycleFrozen_30077afa8fd8c0aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_4e1649c_raw_named_writer.a", "frozen", true)
}
func TestProduct_CycleCurrent_30077afa8fd8c0aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_4e1649c_raw_named_writer.a", "current", true)
}
func TestCyclesCorpus_30077afa8fd8c0aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_4e1649c_raw_named_writer.a")
}
func TestProduct_CycleFrozen_3f0d35ac300f3e97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_silent.a", "frozen", true)
}
func TestProduct_CycleCurrent_3f0d35ac300f3e97(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_silent.a", "current", true)
}
func TestCyclesCorpus_3f0d35ac300f3e97(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_abfe962_super_early_silent.a")
}
func TestProduct_CycleFrozen_43afc59ef5266d30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/shared_slices.a", "frozen", true)
}
func TestProduct_CycleCurrent_43afc59ef5266d30(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/shared_slices.a", "current", true)
}
func TestCyclesCorpus_43afc59ef5266d30(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/shared_slices.a")
}
func TestProduct_CycleFrozen_0adebde91bed7c33(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/stack_over.a", "frozen", true)
}
func TestProduct_CycleCurrent_0adebde91bed7c33(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/stack_over.a", "current", true)
}
func TestCyclesCorpus_0adebde91bed7c33(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/stack_over.a")
}
func TestProduct_CycleFrozen_2ecc03b1321a5d2b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_characters.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ecc03b1321a5d2b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_characters.a", "current", true)
}
func TestCyclesCorpus_2ecc03b1321a5d2b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_views_characters.a")
}
func TestProduct_CycleFrozen_fb9fb8c7eac415c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_fb9fb8c7eac415c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/string_views_throw.a", "current", true)
}
func TestCyclesCorpus_fb9fb8c7eac415c1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/string_views_throw.a")
}
func TestProduct_CycleFrozen_816a38c640682c92(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/syntax_substr.a", "frozen", true)
}
func TestProduct_CycleCurrent_816a38c640682c92(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/syntax_substr.a", "current", true)
}
func TestCyclesCorpus_816a38c640682c92(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/syntax_substr.a")
}
func TestProduct_CycleFrozen_f990a266c5db1d31(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_uint8array.a", "frozen", true)
}
func TestProduct_CycleCurrent_f990a266c5db1d31(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_uint8array.a", "current", true)
}
func TestCyclesCorpus_f990a266c5db1d31(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_uint8array.a")
}
func TestProduct_CycleFrozen_2aa77a3ad03287ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_roll.a", "frozen", true)
}
func TestProduct_CycleCurrent_2aa77a3ad03287ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typeof_null_roll.a", "current", true)
}
func TestCyclesCorpus_2aa77a3ad03287ad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typeof_null_roll.a")
}
func TestProduct_CycleFrozen_9b20473192ab4236(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak_parent.a", "frozen", true)
}
func TestProduct_CycleCurrent_9b20473192ab4236(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/weak_parent.a", "current", true)
}
func TestCyclesCorpus_9b20473192ab4236(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/weak_parent.a")
}
func TestProduct_CycleFrozen_0068d082794ec58c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/optional-indexing/sparse-array.a", "frozen", true)
}
func TestProduct_CycleCurrent_0068d082794ec58c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/optional-indexing/sparse-array.a", "current", true)
}
func TestCyclesCorpus_0068d082794ec58c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/optional-indexing/sparse-array.a")
}
func TestProduct_CycleFrozen_d8e2d0291e702dcf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/crlf.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d8e2d0291e702dcf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/gaps/crlf.ts", "current", true)
}
func TestCyclesCorpus_d8e2d0291e702dcf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/gaps/crlf.ts")
}
func TestProduct_CycleFrozen_0094f2d7e276e2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceStatements.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0094f2d7e276e2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceStatements.ts", "current", true)
}
func TestCyclesCorpus_0094f2d7e276e2de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceStatements.ts")
}
func TestProduct_CycleFrozen_c4e59a6b0e7fe090(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/all.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c4e59a6b0e7fe090(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/all.ts", "current", true)
}
func TestCyclesCorpus_c4e59a6b0e7fe090(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/comments/all.ts")
}
func TestProduct_CycleFrozen_974e504ae54c65a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/policy_message.ts", "frozen", true)
}
func TestProduct_CycleCurrent_974e504ae54c65a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/policy_message.ts", "current", true)
}
func TestCyclesCorpus_974e504ae54c65a3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/policy_message.ts")
}
func TestProduct_CycleFrozen_7b109f0040197d62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/options.a", "frozen", true)
}
func TestProduct_CycleCurrent_7b109f0040197d62(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/options.a", "current", true)
}
func TestCyclesCorpus_7b109f0040197d62(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/options.a")
}
func TestProduct_CycleFrozen_061298f8a21503dd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/016_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_061298f8a21503dd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/016_runtime.a", "current", true)
}
func TestCyclesCorpus_061298f8a21503dd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/016_runtime.a")
}
func TestProduct_CycleFrozen_d44b0e114c3c315c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048.a", "frozen", true)
}
func TestProduct_CycleCurrent_d44b0e114c3c315c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048.a", "current", true)
}
func TestCyclesCorpus_d44b0e114c3c315c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/048.a")
}
func TestProduct_CycleFrozen_71e353ba47bcef26(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/019.a", "frozen", true)
}
func TestProduct_CycleCurrent_71e353ba47bcef26(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/i/019.a", "current", true)
}
func TestCyclesCorpus_71e353ba47bcef26(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/i/019.a")
}
func TestProduct_CycleFrozen_873102ba198d6a8e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/028_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_873102ba198d6a8e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/028_runtime.a", "current", true)
}
func TestCyclesCorpus_873102ba198d6a8e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/028_runtime.a")
}
func TestProduct_CycleFrozen_c15ab2cdf7c0ca3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054.a", "frozen", true)
}
func TestProduct_CycleCurrent_c15ab2cdf7c0ca3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054.a", "current", true)
}
func TestCyclesCorpus_c15ab2cdf7c0ca3a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/054.a")
}
func TestProduct_CycleFrozen_0c523407f2743ddd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/064_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_0c523407f2743ddd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/064_runtime.a", "current", true)
}
func TestCyclesCorpus_0c523407f2743ddd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/064_runtime.a")
}
func TestProduct_CycleFrozen_a8ebd06798fdc2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072.a", "frozen", true)
}
func TestProduct_CycleCurrent_a8ebd06798fdc2de(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072.a", "current", true)
}
func TestCyclesCorpus_a8ebd06798fdc2de(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/072.a")
}
func TestProduct_CycleFrozen_02fa7c3de47468ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-div-regex/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_02fa7c3de47468ac(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-div-regex/messages.ts", "current", true)
}
func TestCyclesCorpus_02fa7c3de47468ac(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-div-regex/messages.ts")
}
func TestProduct_CycleFrozen_8d268633bdbd716e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-var/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8d268633bdbd716e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-var/messages.ts", "current", true)
}
func TestCyclesCorpus_8d268633bdbd716e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-var/messages.ts")
}
func TestProduct_CycleFrozen_43c0dd9fb4109c11(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_43c0dd9fb4109c11(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/rule.a", "current", true)
}
func TestCyclesCorpus_43c0dd9fb4109c11(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/rule.a")
}
func TestProduct_CycleFrozen_b17fbe1a6b8cb8a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-prefer-literal-enum-member/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_b17fbe1a6b8cb8a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-prefer-literal-enum-member/rule.ts", "current", true)
}
func TestCyclesCorpus_b17fbe1a6b8cb8a8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-prefer-literal-enum-member/rule.ts")
}
func TestProduct_CycleFrozen_5f7f5609c3d7fbe8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/12_conditional_empty_array.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5f7f5609c3d7fbe8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/12_conditional_empty_array.ts", "current", true)
}
func TestCyclesCorpus_5f7f5609c3d7fbe8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/12_conditional_empty_array.ts")
}
func TestProduct_CycleFrozen_29bf15cd60d13219(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/14_missing_path_key.ts", "frozen", true)
}
func TestProduct_CycleCurrent_29bf15cd60d13219(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/14_missing_path_key.ts", "current", true)
}
func TestCyclesCorpus_29bf15cd60d13219(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/14_missing_path_key.ts")
}
func TestProduct_CycleFrozen_16f3b1ccba2602a7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/8_generic_callback_result.ts", "frozen", true)
}
func TestProduct_CycleCurrent_16f3b1ccba2602a7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/8_generic_callback_result.ts", "current", true)
}
func TestCyclesCorpus_16f3b1ccba2602a7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/8_generic_callback_result.ts")
}
func TestProduct_CycleFrozen_df3500aab8adf2fd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/edit.a", "frozen", true)
}
func TestProduct_CycleCurrent_df3500aab8adf2fd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/edit.a", "current", true)
}
func TestCyclesCorpus_df3500aab8adf2fd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/edit.a")
}
func TestProduct_CycleFrozen_884d3dc76694d2d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_884d3dc76694d2d5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/messages.a", "current", true)
}
func TestCyclesCorpus_884d3dc76694d2d5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/messages.a")
}
func TestProduct_CycleFrozen_dab04e7bb17a36e1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/suggestion.a", "frozen", true)
}
func TestProduct_CycleCurrent_dab04e7bb17a36e1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/parse/suggestion.a", "current", true)
}
func TestCyclesCorpus_dab04e7bb17a36e1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/parse/suggestion.a")
}
func TestProduct_CycleFrozen_c93e06e42b513049(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/eliminate.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c93e06e42b513049(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/eliminate.ts", "current", true)
}
func TestCyclesCorpus_c93e06e42b513049(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/eliminate.ts")
}
func TestProduct_CycleFrozen_77028323c3c8067e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/assignment.ts", "frozen", true)
}
func TestProduct_CycleCurrent_77028323c3c8067e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/typeaware/assignment.ts", "current", true)
}
func TestCyclesCorpus_77028323c3c8067e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/typeaware/assignment.ts")
}
func TestProduct_CycleFrozen_709a682bbd720d75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/composeError.ts", "frozen", true)
}
func TestProduct_CycleCurrent_709a682bbd720d75(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/composeError.ts", "current", true)
}
func TestCyclesCorpus_709a682bbd720d75(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/composeError.ts")
}
func TestProduct_CycleFrozen_5961ec7d5bd42acf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/cst.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5961ec7d5bd42acf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/cst.ts", "current", true)
}
func TestCyclesCorpus_5961ec7d5bd42acf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/cst.ts")
}
func TestProduct_CycleFrozen_824b7ea37132f76f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/cst_main.ts", "frozen", true)
}
func TestProduct_CycleCurrent_824b7ea37132f76f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/cst_main.ts", "current", true)
}
func TestCyclesCorpus_824b7ea37132f76f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/cst_main.ts")
}
func TestProduct_CycleFrozen_ac2f2dfde05e8869(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/patternNode.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ac2f2dfde05e8869(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/patternNode.ts", "current", true)
}
func TestCyclesCorpus_ac2f2dfde05e8869(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/patternNode.ts")
}
func TestProduct_CycleFrozen_aed494c4f2b322b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/characters.ts", "frozen", true)
}
func TestProduct_CycleCurrent_aed494c4f2b322b0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/typescript/scanner/characters.ts", "current", true)
}
func TestCyclesCorpus_aed494c4f2b322b0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/typescript/scanner/characters.ts")
}
func TestProduct_CycleFrozen_a84488d414e665c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/omitted_scanner_original.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a84488d414e665c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/omitted_scanner_original.ts", "current", true)
}
func TestCyclesCorpus_a84488d414e665c1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/refusals/omitted_scanner_original.ts")
}
func TestProduct_CycleFrozen_93c8ec1b9edbc29c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_first.a", "frozen", true)
}
func TestProduct_CycleCurrent_93c8ec1b9edbc29c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/array_literal_empty_first.a", "current", true)
}
func TestCyclesCorpus_93c8ec1b9edbc29c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/array_literal_empty_first.a")
}
func TestProduct_CycleFrozen_5413d998c6e86539(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_control.a", "frozen", true)
}
func TestProduct_CycleCurrent_5413d998c6e86539(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_control.a", "current", true)
}
func TestCyclesCorpus_5413d998c6e86539(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_control.a")
}
func TestProduct_CycleFrozen_54a5ea9d952abeae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_live_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_54a5ea9d952abeae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_live_values.a", "current", true)
}
func TestCyclesCorpus_54a5ea9d952abeae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_live_values.a")
}
func TestProduct_CycleFrozen_19542d4bd11783d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_loop_body_capture.a", "frozen", true)
}
func TestProduct_CycleCurrent_19542d4bd11783d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_refused/async_refuse_loop_body_capture.a", "current", true)
}
func TestCyclesCorpus_19542d4bd11783d1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_refused/async_refuse_loop_body_capture.a")
}
func TestProduct_CycleFrozen_84b16d95983e6bf4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_snapshots.a", "frozen", true)
}
func TestProduct_CycleCurrent_84b16d95983e6bf4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_snapshots.a", "current", true)
}
func TestCyclesCorpus_84b16d95983e6bf4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_snapshots.a")
}
func TestProduct_CycleFrozen_0ee4074ed1d1fef0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_defined_lent_field.a", "frozen", true)
}
func TestProduct_CycleCurrent_0ee4074ed1d1fef0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/borrow_defined_lent_field.a", "current", true)
}
func TestCyclesCorpus_0ee4074ed1d1fef0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/borrow_defined_lent_field.a")
}
func TestProduct_CycleFrozen_1297f494f22b5cec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_contracts.a", "frozen", true)
}
func TestProduct_CycleCurrent_1297f494f22b5cec(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_contracts.a", "current", true)
}
func TestCyclesCorpus_1297f494f22b5cec(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_overload_contracts.a")
}
func TestProduct_CycleFrozen_dbedd44d6cd22098(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a", "frozen", true)
}
func TestProduct_CycleCurrent_dbedd44d6cd22098(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a", "current", true)
}
func TestCyclesCorpus_dbedd44d6cd22098(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_small_stopped/census_small_optional_stopped.a")
}
func TestProduct_CycleFrozen_10c11ed6c871b074(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/records.a", "frozen", true)
}
func TestProduct_CycleCurrent_10c11ed6c871b074(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/accepted/records.a", "current", true)
}
func TestCyclesCorpus_10c11ed6c871b074(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/accepted/records.a")
}
func TestProduct_CycleFrozen_0ddbc8d86c9c0bad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_custom_empty.a", "frozen", true)
}
func TestProduct_CycleCurrent_0ddbc8d86c9c0bad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_uncaught_custom_empty.a", "current", true)
}
func TestCyclesCorpus_0ddbc8d86c9c0bad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_uncaught_custom_empty.a")
}
func TestProduct_CycleFrozen_b2d66897e6f3dccc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_boolean.a", "frozen", true)
}
func TestProduct_CycleCurrent_b2d66897e6f3dccc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/element_access_boolean.a", "current", true)
}
func TestCyclesCorpus_b2d66897e6f3dccc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/element_access_boolean.a")
}
func TestProduct_CycleFrozen_0563ffad1815ec98(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_update.a", "frozen", true)
}
func TestProduct_CycleCurrent_0563ffad1815ec98(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_update.a", "current", true)
}
func TestCyclesCorpus_0563ffad1815ec98(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_open_never_update.a")
}
func TestProduct_CycleFrozen_da17f6a018bd8fd8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/caught.a", "frozen", true)
}
func TestProduct_CycleCurrent_da17f6a018bd8fd8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/caught.a", "current", true)
}
func TestCyclesCorpus_da17f6a018bd8fd8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/caught.a")
}
func TestProduct_CycleFrozen_f33771aa66678674(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/ctor_wrapped.a", "frozen", true)
}
func TestProduct_CycleCurrent_f33771aa66678674(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/ctor_wrapped.a", "current", true)
}
func TestCyclesCorpus_f33771aa66678674(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/ctor_wrapped.a")
}
func TestProduct_CycleFrozen_d4f6ce2c44ccb4c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/older_instance.a", "frozen", true)
}
func TestProduct_CycleCurrent_d4f6ce2c44ccb4c1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/fresh_refused/older_instance.a", "current", true)
}
func TestCyclesCorpus_d4f6ce2c44ccb4c1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/fresh_refused/older_instance.a")
}
func TestProduct_CycleFrozen_1752a2cfb4d46627(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_class_map.a", "frozen", true)
}
func TestProduct_CycleCurrent_1752a2cfb4d46627(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_class_map.a", "current", true)
}
func TestCyclesCorpus_1752a2cfb4d46627(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_class_map.a")
}
func TestProduct_CycleFrozen_24ad5e5463bc3116(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_14.a", "frozen", true)
}
func TestProduct_CycleCurrent_24ad5e5463bc3116(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_14.a", "current", true)
}
func TestCyclesCorpus_24ad5e5463bc3116(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_regression_14.a")
}
func TestProduct_CycleFrozen_6673f0869b88521f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/a.a", "frozen", true)
}
func TestProduct_CycleCurrent_6673f0869b88521f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/classes/a.a", "current", true)
}
func TestCyclesCorpus_6673f0869b88521f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/classes/a.a")
}
func TestProduct_CycleFrozen_2d969dea49c10286(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/indexing.a", "frozen", true)
}
func TestProduct_CycleCurrent_2d969dea49c10286(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/indexing.a", "current", true)
}
func TestCyclesCorpus_2d969dea49c10286(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/indexing.a")
}
func TestProduct_CycleFrozen_6ed804d1e0eee64e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_entries_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_6ed804d1e0eee64e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_entries_spread.a", "current", true)
}
func TestCyclesCorpus_6ed804d1e0eee64e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_map_entries_spread.a")
}
func TestProduct_CycleFrozen_265d6fe4d78a0ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_string_conversion.a", "frozen", true)
}
func TestProduct_CycleCurrent_265d6fe4d78a0ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/library_string_conversion.a", "current", true)
}
func TestCyclesCorpus_265d6fe4d78a0ad5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/library_string_conversion.a")
}
func TestProduct_CycleFrozen_0ae830c562704687(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lone_surrogates.a", "frozen", true)
}
func TestProduct_CycleCurrent_0ae830c562704687(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/lone_surrogates.a", "current", true)
}
func TestCyclesCorpus_0ae830c562704687(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/lone_surrogates.a")
}
func TestProduct_CycleFrozen_0aa7e17160178f16(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_numbers.a", "frozen", true)
}
func TestProduct_CycleCurrent_0aa7e17160178f16(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_foreach_numbers.a", "current", true)
}
func TestCyclesCorpus_0aa7e17160178f16(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_foreach_numbers.a")
}
func TestProduct_CycleFrozen_29da78d17beef06c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_string_false.a", "frozen", true)
}
func TestProduct_CycleCurrent_29da78d17beef06c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/map_union_value_string_false.a", "current", true)
}
func TestCyclesCorpus_29da78d17beef06c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/map_union_value_string_false.a")
}
func TestProduct_CycleFrozen_ba847819c4a962be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/memory_examples/tree.a", "frozen", true)
}
func TestProduct_CycleCurrent_ba847819c4a962be(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/memory_examples/tree.a", "current", true)
}
func TestCyclesCorpus_ba847819c4a962be(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/memory_examples/tree.a")
}
func TestProduct_CycleFrozen_774e49d752fa648d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_conversions.a", "frozen", true)
}
func TestProduct_CycleCurrent_774e49d752fa648d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_conversions.a", "current", true)
}
func TestCyclesCorpus_774e49d752fa648d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_conversions.a")
}
func TestProduct_CycleFrozen_978a312b29f459e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_number_statics.a", "frozen", true)
}
func TestProduct_CycleCurrent_978a312b29f459e8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_number_statics.a", "current", true)
}
func TestCyclesCorpus_978a312b29f459e8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_number_statics.a")
}
func TestProduct_CycleFrozen_bd58c950ffb1faa5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_padstart.a", "frozen", true)
}
func TestProduct_CycleCurrent_bd58c950ffb1faa5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_string_padstart.a", "current", true)
}
func TestCyclesCorpus_bd58c950ffb1faa5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_string_padstart.a")
}
func TestProduct_CycleFrozen_22a15c7685a6063c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/async.a", "frozen", true)
}
func TestProduct_CycleCurrent_22a15c7685a6063c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/async.a", "current", true)
}
func TestCyclesCorpus_22a15c7685a6063c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/async.a")
}
func TestProduct_CycleFrozen_7717e53873aa4dd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/race_alias.a", "frozen", true)
}
func TestProduct_CycleCurrent_7717e53873aa4dd3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/race_alias.a", "current", true)
}
func TestCyclesCorpus_7717e53873aa4dd3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/race_alias.a")
}
func TestProduct_CycleFrozen_2b1ae97f14453e31(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/narrowed_compared.a", "frozen", true)
}
func TestProduct_CycleCurrent_2b1ae97f14453e31(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/narrowed_compared.a", "current", true)
}
func TestCyclesCorpus_2b1ae97f14453e31(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/narrowed_compared.a")
}
func TestProduct_CycleFrozen_fb2d258c784e545d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_date.a", "frozen", true)
}
func TestProduct_CycleCurrent_fb2d258c784e545d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/node_fs_file_date.a", "current", true)
}
func TestCyclesCorpus_fb2d258c784e545d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/node_fs_file_date.a")
}
func TestProduct_CycleFrozen_34aeb073472230ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_checked_undefined.ts", "frozen", true)
}
func TestProduct_CycleCurrent_34aeb073472230ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_checked_undefined.ts", "current", true)
}
func TestCyclesCorpus_34aeb073472230ad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_checked_undefined.ts")
}
func TestProduct_CycleFrozen_5525afb34c556d9a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_spread.a", "frozen", true)
}
func TestProduct_CycleCurrent_5525afb34c556d9a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_refuse_spread.a", "current", true)
}
func TestCyclesCorpus_5525afb34c556d9a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_refuse_spread.a")
}
func TestProduct_CycleFrozen_1a0a2fdda6bddcb6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_weak.ts", "frozen", true)
}
func TestProduct_CycleCurrent_1a0a2fdda6bddcb6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_weak.ts", "current", true)
}
func TestCyclesCorpus_1a0a2fdda6bddcb6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_weak.ts")
}
func TestProduct_CycleFrozen_a7026b3a015bd1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/template.a", "frozen", true)
}
func TestProduct_CycleCurrent_a7026b3a015bd1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/notyet_element_access/template.a", "current", true)
}
func TestCyclesCorpus_a7026b3a015bd1b7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/notyet_element_access/template.a")
}
func TestProduct_CycleFrozen_56110fe0a7fe3a43(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_mkdtemp.a", "frozen", true)
}
func TestProduct_CycleCurrent_56110fe0a7fe3a43(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_mkdtemp.a", "current", true)
}
func TestCyclesCorpus_56110fe0a7fe3a43(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_widening_refused/node_fs_file_mkdtemp.a")
}
func TestProduct_CycleFrozen_1f42b38aba1fbb3e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/group_guard.a", "frozen", true)
}
func TestProduct_CycleCurrent_1f42b38aba1fbb3e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/group_guard.a", "current", true)
}
func TestCyclesCorpus_1f42b38aba1fbb3e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_replace/group_guard.a")
}
func TestProduct_CycleFrozen_f4ac40f5bdd6c108(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reland_refused/generic_instance_key_json.a", "frozen", true)
}
func TestProduct_CycleCurrent_f4ac40f5bdd6c108(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reland_refused/generic_instance_key_json.a", "current", true)
}
func TestCyclesCorpus_f4ac40f5bdd6c108(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/reland_refused/generic_instance_key_json.a")
}
func TestProduct_CycleFrozen_ab564488305c98e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_move_before_ready.a", "frozen", true)
}
func TestProduct_CycleCurrent_ab564488305c98e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_move_before_ready.a", "current", true)
}
func TestCyclesCorpus_ab564488305c98e9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/reuse_move_before_ready.a")
}
func TestProduct_CycleFrozen_d885f7694cdf8811(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_3ab8f94_def_variants.a", "frozen", true)
}
func TestProduct_CycleCurrent_d885f7694cdf8811(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_3ab8f94_def_variants.a", "current", true)
}
func TestCyclesCorpus_d885f7694cdf8811(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_3ab8f94_def_variants.a")
}
func TestProduct_CycleFrozen_d11342c7839d2d9c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p67_static_class_union_member.a", "frozen", true)
}
func TestProduct_CycleCurrent_d11342c7839d2d9c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p67_static_class_union_member.a", "current", true)
}
func TestCyclesCorpus_d11342c7839d2d9c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p67_static_class_union_member.a")
}
func TestProduct_CycleFrozen_517252442280a08a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q04_optional_call_container.a", "frozen", true)
}
func TestProduct_CycleCurrent_517252442280a08a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q04_optional_call_container.a", "current", true)
}
func TestCyclesCorpus_517252442280a08a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_predicates_q04_optional_call_container.a")
}
func TestProduct_CycleFrozen_da00535f0814696f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p72_mixed_element_access_wrong.a", "frozen", true)
}
func TestProduct_CycleCurrent_da00535f0814696f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p72_mixed_element_access_wrong.a", "current", true)
}
func TestCyclesCorpus_da00535f0814696f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p72_mixed_element_access_wrong.a")
}
func TestProduct_CycleFrozen_8f4a5368c577d16c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_bdb950b_generic.a", "frozen", true)
}
func TestProduct_CycleCurrent_8f4a5368c577d16c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_bdb950b_generic.a", "current", true)
}
func TestCyclesCorpus_8f4a5368c577d16c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_bdb950b_generic.a")
}
func TestProduct_CycleFrozen_9c2e361f9a916fbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/search_from_sweep.a", "frozen", true)
}
func TestProduct_CycleCurrent_9c2e361f9a916fbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/search_from_sweep.a", "current", true)
}
func TestCyclesCorpus_9c2e361f9a916fbd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/search_from_sweep.a")
}
func TestProduct_CycleFrozen_42750c929decaad9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/searches.a", "frozen", true)
}
func TestProduct_CycleCurrent_42750c929decaad9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/searches.a", "current", true)
}
func TestCyclesCorpus_42750c929decaad9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/searches.a")
}
func TestProduct_CycleFrozen_f5f44dd87bd17e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_arrays.a", "frozen", true)
}
func TestProduct_CycleCurrent_f5f44dd87bd17e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_arrays.a", "current", true)
}
func TestCyclesCorpus_f5f44dd87bd17e6e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_arrays.a")
}
func TestProduct_CycleFrozen_1ea9605a3319d270(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_objects.a", "frozen", true)
}
func TestProduct_CycleCurrent_1ea9605a3319d270(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_foreach_objects.a", "current", true)
}
func TestCyclesCorpus_1ea9605a3319d270(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_foreach_objects.a")
}
func TestProduct_CycleFrozen_d169f59bf864be54(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/throw_keeps_old_value_variants.a", "frozen", true)
}
func TestProduct_CycleCurrent_d169f59bf864be54(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/throw_keeps_old_value_variants.a", "current", true)
}
func TestCyclesCorpus_d169f59bf864be54(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/throw_keeps_old_value_variants.a")
}
func TestProduct_CycleFrozen_666149d4776adc50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/unions.a", "frozen", true)
}
func TestProduct_CycleCurrent_666149d4776adc50(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/unions.a", "current", true)
}
func TestCyclesCorpus_666149d4776adc50(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/unions.a")
}
func TestProduct_CycleFrozen_309109ab1449d6e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes.a", "frozen", true)
}
func TestProduct_CycleCurrent_309109ab1449d6e6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes.a", "current", true)
}
func TestCyclesCorpus_309109ab1449d6e6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/writes.a")
}
func TestProduct_CycleFrozen_b928c82fb62f64b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/host_nullable_date.a", "frozen", true)
}
func TestProduct_CycleCurrent_b928c82fb62f64b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "review/host_nullable_date.a", "current", true)
}
func TestCyclesCorpus_b928c82fb62f64b6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "review/host_nullable_date.a")
}
func TestProduct_CycleFrozen_a36e6e79b023d6e2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/location.ts", "frozen", true)
}
func TestProduct_CycleCurrent_a36e6e79b023d6e2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/location.ts", "current", true)
}
func TestCyclesCorpus_a36e6e79b023d6e2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/location.ts")
}
func TestProduct_CycleFrozen_c48d929cc1ed89cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceParser.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c48d929cc1ed89cd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/sourceParser.ts", "current", true)
}
func TestCyclesCorpus_c48d929cc1ed89cd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/sourceParser.ts")
}
func TestProduct_CycleFrozen_f3f9c1ee81e05697(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/6_boolean_element.ts", "frozen", true)
}
func TestProduct_CycleCurrent_f3f9c1ee81e05697(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/6_boolean_element.ts", "current", true)
}
func TestCyclesCorpus_f3f9c1ee81e05697(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/gaps/6_boolean_element.ts")
}
func TestProduct_CycleFrozen_64dc815183eaa8b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/can_begin_at.ts", "frozen", true)
}
func TestProduct_CycleCurrent_64dc815183eaa8b6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/helpers/comments/can_begin_at.ts", "current", true)
}
func TestCyclesCorpus_64dc815183eaa8b6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/helpers/comments/can_begin_at.ts")
}
func TestProduct_CycleFrozen_c08d0dcbe7386818(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/017_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_c08d0dcbe7386818(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/017_runtime.a", "current", true)
}
func TestCyclesCorpus_c08d0dcbe7386818(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/017_runtime.a")
}
func TestProduct_CycleFrozen_f8c1e780c7691c06(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006.a", "frozen", true)
}
func TestProduct_CycleCurrent_f8c1e780c7691c06(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006.a", "current", true)
}
func TestCyclesCorpus_f8c1e780c7691c06(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/006.a")
}
func TestProduct_CycleFrozen_a3c89699dd552b6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/029_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_a3c89699dd552b6a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/029_runtime.a", "current", true)
}
func TestCyclesCorpus_a3c89699dd552b6a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/029_runtime.a")
}
func TestProduct_CycleFrozen_e8fd90e103491603(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055.a", "frozen", true)
}
func TestProduct_CycleCurrent_e8fd90e103491603(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055.a", "current", true)
}
func TestCyclesCorpus_e8fd90e103491603(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/055.a")
}
func TestProduct_CycleFrozen_2c65aa743ea11699(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/065_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_2c65aa743ea11699(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/065_runtime.a", "current", true)
}
func TestCyclesCorpus_2c65aa743ea11699(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/065_runtime.a")
}
func TestProduct_CycleFrozen_dff841975ba5d06d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073.a", "frozen", true)
}
func TestProduct_CycleCurrent_dff841975ba5d06d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073.a", "current", true)
}
func TestCyclesCorpus_dff841975ba5d06d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/073.a")
}
func TestProduct_CycleFrozen_ae8c5555ab96ccbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-plusplus/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ae8c5555ab96ccbd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-plusplus/rule.ts", "current", true)
}
func TestCyclesCorpus_ae8c5555ab96ccbd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-plusplus/rule.ts")
}
func TestProduct_CycleFrozen_488e25e8f091a181(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-return-assign/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_488e25e8f091a181(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-return-assign/rule.ts", "current", true)
}
func TestCyclesCorpus_488e25e8f091a181(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-return-assign/rule.ts")
}
func TestProduct_CycleFrozen_d3691325c7fe0be0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unneeded-ternary/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d3691325c7fe0be0(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-unneeded-ternary/rule.ts", "current", true)
}
func TestCyclesCorpus_d3691325c7fe0be0(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-unneeded-ternary/rule.ts")
}
func TestProduct_CycleFrozen_c8695de4ddf7f13e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/astArena.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c8695de4ddf7f13e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/astArena.ts", "current", true)
}
func TestCyclesCorpus_c8695de4ddf7f13e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/astArena.ts")
}
func TestProduct_CycleFrozen_5bbaf7d47c6cfec1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/codeblocks.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5bbaf7d47c6cfec1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/codeblocks.ts", "current", true)
}
func TestCyclesCorpus_5bbaf7d47c6cfec1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/codeblocks.ts")
}
func TestProduct_CycleFrozen_793ee1b1ecfea7b2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/htmlblocks.ts", "frozen", true)
}
func TestProduct_CycleCurrent_793ee1b1ecfea7b2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/htmlblocks.ts", "current", true)
}
func TestCyclesCorpus_793ee1b1ecfea7b2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/htmlblocks.ts")
}
func TestProduct_CycleFrozen_d046297441a06ad3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/mdastNode.ts", "frozen", true)
}
func TestProduct_CycleCurrent_d046297441a06ad3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/mdastNode.ts", "current", true)
}
func TestCyclesCorpus_d046297441a06ad3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/mdastNode.ts")
}
func TestProduct_CycleFrozen_bbc490f3de5c0b3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/gaps/2_return_undefined.ts", "frozen", true)
}
func TestProduct_CycleCurrent_bbc490f3de5c0b3a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/gaps/2_return_undefined.ts", "current", true)
}
func TestCyclesCorpus_bbc490f3de5c0b3a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/mediaquery/gaps/2_return_undefined.ts")
}
func TestProduct_CycleFrozen_9fc381de99b8681a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/index.ts", "frozen", true)
}
func TestProduct_CycleCurrent_9fc381de99b8681a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/mediaquery/index.ts", "current", true)
}
func TestCyclesCorpus_9fc381de99b8681a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/mediaquery/index.ts")
}
func TestProduct_CycleFrozen_32a97c36091657a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/collectionItem.ts", "frozen", true)
}
func TestProduct_CycleCurrent_32a97c36091657a3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/collectionItem.ts", "current", true)
}
func TestCyclesCorpus_32a97c36091657a3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/collectionItem.ts")
}
func TestProduct_CycleFrozen_ab87e4c319cf6979(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/scalarText.ts", "frozen", true)
}
func TestProduct_CycleCurrent_ab87e4c319cf6979(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/scalarText.ts", "current", true)
}
func TestCyclesCorpus_ab87e4c319cf6979(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/scalarText.ts")
}
func TestProduct_CycleFrozen_a631c406351a0816(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/nested_cycle.a", "frozen", true)
}
func TestProduct_CycleCurrent_a631c406351a0816(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/refusals/nested_cycle.a", "current", true)
}
func TestCyclesCorpus_a631c406351a0816(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/refusals/nested_cycle.a")
}
func TestProduct_CycleFrozen_fa7e0a56e0622f78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_promise_expression.a", "frozen", true)
}
func TestProduct_CycleCurrent_fa7e0a56e0622f78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_promise_expression.a", "current", true)
}
func TestCyclesCorpus_fa7e0a56e0622f78(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_promise_expression.a")
}
func TestProduct_CycleFrozen_bf2022cad768e7a1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/modules/sibling.a", "frozen", true)
}
func TestProduct_CycleCurrent_bf2022cad768e7a1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/async_review/refused/modules/sibling.a", "current", true)
}
func TestCyclesCorpus_bf2022cad768e7a1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/async_review/refused/modules/sibling.a")
}
func TestProduct_CycleFrozen_55686eec73caab88(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/bitwise.a", "frozen", true)
}
func TestProduct_CycleCurrent_55686eec73caab88(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/bitwise.a", "current", true)
}
func TestCyclesCorpus_55686eec73caab88(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/bitwise.a")
}
func TestProduct_CycleFrozen_b1cdd58ea40621ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_const_string_fails.a", "frozen", true)
}
func TestProduct_CycleCurrent_b1cdd58ea40621ab(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/cast_enum_const_string_fails.a", "current", true)
}
func TestCyclesCorpus_b1cdd58ea40621ab(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/cast_enum_const_string_fails.a")
}
func TestProduct_CycleFrozen_907a0682824e1ee7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_boolean_sieve.a", "frozen", true)
}
func TestProduct_CycleCurrent_907a0682824e1ee7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_boolean_sieve.a", "current", true)
}
func TestCyclesCorpus_907a0682824e1ee7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_boolean_sieve.a")
}
func TestProduct_CycleFrozen_9724578a89caded3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_lie.a", "frozen", true)
}
func TestProduct_CycleCurrent_9724578a89caded3(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/census_overload_lie.a", "current", true)
}
func TestCyclesCorpus_9724578a89caded3(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/census_overload_lie.a")
}
func TestProduct_CycleFrozen_dd5cc62b1ad1f097(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_distinct.a", "frozen", true)
}
func TestProduct_CycleCurrent_dd5cc62b1ad1f097(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/class_features_distinct.a", "current", true)
}
func TestCyclesCorpus_dd5cc62b1ad1f097(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/class_features_distinct.a")
}
func TestProduct_CycleFrozen_64c411adccd3cddd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/mutable_array.a", "frozen", true)
}
func TestProduct_CycleCurrent_64c411adccd3cddd(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/concurrency/refused/mutable_array.a", "current", true)
}
func TestCyclesCorpus_64c411adccd3cddd(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/concurrency/refused/mutable_array.a")
}
func TestProduct_CycleFrozen_a0bd750d6f7d8ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_function.a", "frozen", true)
}
func TestProduct_CycleCurrent_a0bd750d6f7d8ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/conditions_function.a", "current", true)
}
func TestCyclesCorpus_a0bd750d6f7d8ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/conditions_function.a")
}
func TestProduct_CycleFrozen_ab2a3a44b8376718(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_plain_view.a", "frozen", true)
}
func TestProduct_CycleCurrent_ab2a3a44b8376718(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/coverage_plain_view.a", "current", true)
}
func TestCyclesCorpus_ab2a3a44b8376718(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/coverage_plain_view.a")
}
func TestProduct_CycleFrozen_e93603d00140c0d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_class_narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_e93603d00140c0d1(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_class_narrowed.a", "current", true)
}
func TestCyclesCorpus_e93603d00140c0d1(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/e4eec87_f1_class_narrowed.a")
}
func TestProduct_CycleFrozen_4b3db8058c3758cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_field_narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_4b3db8058c3758cf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/e4eec87_f1_field_narrowed.a", "current", true)
}
func TestCyclesCorpus_4b3db8058c3758cf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/e4eec87_f1_field_narrowed.a")
}
func TestProduct_CycleFrozen_fc632e3ba5431dc7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_misfit.a", "frozen", true)
}
func TestProduct_CycleCurrent_fc632e3ba5431dc7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_checked_misfit.a", "current", true)
}
func TestCyclesCorpus_fc632e3ba5431dc7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_checked_misfit.a")
}
func TestProduct_CycleFrozen_727ab2c9316df045(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_proven_import.a", "frozen", true)
}
func TestProduct_CycleCurrent_727ab2c9316df045(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/entries_proven_import.a", "current", true)
}
func TestCyclesCorpus_727ab2c9316df045(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/entries_proven_import.a")
}
func TestProduct_CycleFrozen_4cdc4a6229885edb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_field.a", "frozen", true)
}
func TestProduct_CycleCurrent_4cdc4a6229885edb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_field.a", "current", true)
}
func TestCyclesCorpus_4cdc4a6229885edb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_open_never_field.a")
}
func TestProduct_CycleFrozen_6cba6d21104fc8ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_index.a", "frozen", true)
}
func TestProduct_CycleCurrent_6cba6d21104fc8ae(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/enums_open_never_index.a", "current", true)
}
func TestCyclesCorpus_6cba6d21104fc8ae(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/enums_open_never_index.a")
}
func TestProduct_CycleFrozen_b7532aa4f0a08ad6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_optional_boolean_function.a", "frozen", true)
}
func TestProduct_CycleCurrent_b7532aa4f0a08ad6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_optional_boolean_function.a", "current", true)
}
func TestCyclesCorpus_b7532aa4f0a08ad6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_optional_boolean_function.a")
}
func TestProduct_CycleFrozen_da785d8985d4de27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_snapshot_next.a", "frozen", true)
}
func TestProduct_CycleCurrent_da785d8985d4de27(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/getters_census_snapshot_next.a", "current", true)
}
func TestCyclesCorpus_da785d8985d4de27(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/getters_census_snapshot_next.a")
}
func TestProduct_CycleFrozen_2ffbda241a58f03d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_08.a", "frozen", true)
}
func TestProduct_CycleCurrent_2ffbda241a58f03d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/graph_regions_regression_08.a", "current", true)
}
func TestCyclesCorpus_2ffbda241a58f03d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/graph_regions_regression_08.a")
}
func TestProduct_CycleFrozen_9de03891b249e316(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/order/b.a", "frozen", true)
}
func TestProduct_CycleCurrent_9de03891b249e316(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/import_cycles/order/b.a", "current", true)
}
func TestCyclesCorpus_9de03891b249e316(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/import_cycles/order/b.a")
}
func TestProduct_CycleFrozen_7d20384062f90580(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_replacer.a", "frozen", true)
}
func TestProduct_CycleCurrent_7d20384062f90580(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/json_stringify_replacer.a", "current", true)
}
func TestCyclesCorpus_7d20384062f90580(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/json_stringify_replacer.a")
}
func TestProduct_CycleFrozen_51750d7cfc86968d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_array_values.a", "frozen", true)
}
func TestProduct_CycleCurrent_51750d7cfc86968d(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/method_coverage_array_values.a", "current", true)
}
func TestCyclesCorpus_51750d7cfc86968d(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/method_coverage_array_values.a")
}
func TestProduct_CycleFrozen_937b3453ba98dbf6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/modules/format.a", "frozen", true)
}
func TestProduct_CycleCurrent_937b3453ba98dbf6(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/modules/format.a", "current", true)
}
func TestCyclesCorpus_937b3453ba98dbf6(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/modules/format.a")
}
func TestProduct_CycleFrozen_e921bcf77847af47(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested.a", "frozen", true)
}
func TestProduct_CycleCurrent_e921bcf77847af47(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/moves/refused/nested.a", "current", true)
}
func TestCyclesCorpus_e921bcf77847af47(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/moves/refused/nested.a")
}
func TestProduct_CycleFrozen_36c63aef9ce76186(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_debug_probe.a", "frozen", true)
}
func TestProduct_CycleCurrent_36c63aef9ce76186(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/namespaces_debug_probe.a", "current", true)
}
func TestCyclesCorpus_36c63aef9ce76186(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/namespaces_debug_probe.a")
}
func TestProduct_CycleFrozen_b5d2bde2c834657f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_owned_parameter.a", "frozen", true)
}
func TestProduct_CycleCurrent_b5d2bde2c834657f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/nested_owned_parameter.a", "current", true)
}
func TestCyclesCorpus_b5d2bde2c834657f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/nested_owned_parameter.a")
}
func TestProduct_CycleFrozen_e7548c511c811135(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_exception.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e7548c511c811135(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/non_null_uninitialized_exception.ts", "current", true)
}
func TestCyclesCorpus_e7548c511c811135(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/non_null_uninitialized_exception.ts")
}
func TestProduct_CycleFrozen_b28b5f54d3c9648c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_strings.a", "frozen", true)
}
func TestProduct_CycleCurrent_b28b5f54d3c9648c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/optional_strings.a", "current", true)
}
func TestCyclesCorpus_b28b5f54d3c9648c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/optional_strings.a")
}
func TestProduct_CycleFrozen_fd1686c314a4bcb5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/override_same_representation.a", "frozen", true)
}
func TestProduct_CycleCurrent_fd1686c314a4bcb5(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/override_same_representation.a", "current", true)
}
func TestCyclesCorpus_fd1686c314a4bcb5(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/override_same_representation.a")
}
func TestProduct_CycleFrozen_0c7e0c0a36639b1b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_guards.a", "frozen", true)
}
func TestProduct_CycleCurrent_0c7e0c0a36639b1b(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/proven_guards.a", "current", true)
}
func TestCyclesCorpus_0c7e0c0a36639b1b(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/proven_guards.a")
}
func TestProduct_CycleFrozen_679e752d586efc9f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp.a", "frozen", true)
}
func TestProduct_CycleCurrent_679e752d586efc9f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp.a", "current", true)
}
func TestCyclesCorpus_679e752d586efc9f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp.a")
}
func TestProduct_CycleFrozen_f41dd9f25bc3d9fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/arguments.a", "frozen", true)
}
func TestProduct_CycleCurrent_f41dd9f25bc3d9fe(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/regexp_replace/arguments.a", "current", true)
}
func TestCyclesCorpus_f41dd9f25bc3d9fe(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/regexp_replace/arguments.a")
}
func TestProduct_CycleFrozen_0a5bbd8ee62679af(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/repeat_huge.a", "frozen", true)
}
func TestProduct_CycleCurrent_0a5bbd8ee62679af(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/repeat_huge.a", "current", true)
}
func TestCyclesCorpus_0a5bbd8ee62679af(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/repeat_huge.a")
}
func TestProduct_CycleFrozen_4a1833e247e2924a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_handover_throw.a", "frozen", true)
}
func TestProduct_CycleCurrent_4a1833e247e2924a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/reuse_handover_throw.a", "current", true)
}
func TestCyclesCorpus_4a1833e247e2924a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/reuse_handover_throw.a")
}
func TestProduct_CycleFrozen_17fd2b4923fe6826(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_math_edges.a", "frozen", true)
}
func TestProduct_CycleCurrent_17fd2b4923fe6826(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_math_edges.a", "current", true)
}
func TestCyclesCorpus_17fd2b4923fe6826(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_1d3929c_math_edges.a")
}
func TestProduct_CycleFrozen_e172f29285b2a3a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f1_field_narrowed.a", "frozen", true)
}
func TestProduct_CycleCurrent_e172f29285b2a3a8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f1_field_narrowed.a", "current", true)
}
func TestCyclesCorpus_e172f29285b2a3a8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_e4eec87_f1_field_narrowed.a")
}
func TestProduct_CycleFrozen_d3f519aa0b9102d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_40.a", "frozen", true)
}
func TestProduct_CycleCurrent_d3f519aa0b9102d2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_40.a", "current", true)
}
func TestCyclesCorpus_d3f519aa0b9102d2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_native_p70_recursive_depth_40.a")
}
func TestProduct_CycleFrozen_82f23ad42dd46ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p53_callable_union_destructure_good.a", "frozen", true)
}
func TestProduct_CycleCurrent_82f23ad42dd46ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p53_callable_union_destructure_good.a", "current", true)
}
func TestCyclesCorpus_82f23ad42dd46ba8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/agree/fxspptb_oct9_views_p53_callable_union_destructure_good.a")
}
func TestProduct_CycleFrozen_25aee0287b3ea838(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_e23c7ab_v8_divergence.a", "frozen", true)
}
func TestProduct_CycleCurrent_25aee0287b3ea838(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_e23c7ab_v8_divergence.a", "current", true)
}
func TestCyclesCorpus_25aee0287b3ea838(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_e23c7ab_v8_divergence.a")
}
func TestProduct_CycleFrozen_b3936ff382e9d3e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p21_find_inferred_number.a", "frozen", true)
}
func TestProduct_CycleCurrent_b3936ff382e9d3e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p21_find_inferred_number.a", "current", true)
}
func TestCyclesCorpus_b3936ff382e9d3e9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p21_find_inferred_number.a")
}
func TestProduct_CycleFrozen_b787fbb9a8d0df6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p22_filter_inferred_narrowing.a", "frozen", true)
}
func TestProduct_CycleCurrent_b787fbb9a8d0df6c(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p22_filter_inferred_narrowing.a", "current", true)
}
func TestCyclesCorpus_b787fbb9a8d0df6c(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/review/refused/fxspptb_oct8_predicates_p22_filter_inferred_narrowing.a")
}
func TestProduct_CycleFrozen_cd0bd0fbd9f2327e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_bound_method.a", "frozen", true)
}
func TestProduct_CycleCurrent_cd0bd0fbd9f2327e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/route_targets_bound_method.a", "current", true)
}
func TestCyclesCorpus_cd0bd0fbd9f2327e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/route_targets_bound_method.a")
}
func TestProduct_CycleFrozen_8b48aa216f866b5f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_undefined.a", "frozen", true)
}
func TestProduct_CycleCurrent_8b48aa216f866b5f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/set_undefined.a", "current", true)
}
func TestCyclesCorpus_8b48aa216f866b5f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/set_undefined.a")
}
func TestProduct_CycleFrozen_ab78bd85933da489(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_logical_assignment.a", "frozen", true)
}
func TestProduct_CycleCurrent_ab78bd85933da489(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/taste_logical_assignment.a", "current", true)
}
func TestCyclesCorpus_ab78bd85933da489(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/taste_logical_assignment.a")
}
func TestProduct_CycleFrozen_b20241e4bab9ead4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_sort_uint16array.a", "frozen", true)
}
func TestProduct_CycleCurrent_b20241e4bab9ead4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/typed_arrays_sort_uint16array.a", "current", true)
}
func TestCyclesCorpus_b20241e4bab9ead4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/typed_arrays_sort_uint16array.a")
}
func TestProduct_CycleFrozen_460a0ba4144db50f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/user_iterators_rest_tdz.a", "frozen", true)
}
func TestProduct_CycleCurrent_460a0ba4144db50f(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/user_iterators_rest_tdz.a", "current", true)
}
func TestCyclesCorpus_460a0ba4144db50f(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/user_iterators_rest_tdz.a")
}
func TestProduct_CycleFrozen_880689c907b3face(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/walk.a", "frozen", true)
}
func TestProduct_CycleCurrent_880689c907b3face(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/walk.a", "current", true)
}
func TestCyclesCorpus_880689c907b3face(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/walk.a")
}
func TestProduct_CycleFrozen_eeadcdf01fe70fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes_past_end.a", "frozen", true)
}
func TestProduct_CycleCurrent_eeadcdf01fe70fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "internal/oracle/testdata/writes_past_end.a", "current", true)
}
func TestCyclesCorpus_eeadcdf01fe70fdf(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "internal/oracle/testdata/writes_past_end.a")
}
func TestProduct_CycleFrozen_99358127b63505e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/testdata/range_guard.ts", "frozen", true)
}
func TestProduct_CycleCurrent_99358127b63505e9(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/css/testdata/range_guard.ts", "current", true)
}
func TestCyclesCorpus_99358127b63505e9(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/css/testdata/range_guard.ts")
}
func TestProduct_CycleFrozen_71a3fb6fca94ffa7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/arena.ts", "frozen", true)
}
func TestProduct_CycleCurrent_71a3fb6fca94ffa7(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/arena.ts", "current", true)
}
func TestCyclesCorpus_71a3fb6fca94ffa7(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/arena.ts")
}
func TestProduct_CycleFrozen_e2abb6f072c083aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/postprocess.ts", "frozen", true)
}
func TestProduct_CycleCurrent_e2abb6f072c083aa(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/estree/postprocess.ts", "current", true)
}
func TestCyclesCorpus_e2abb6f072c083aa(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/estree/postprocess.ts")
}
func TestProduct_CycleFrozen_4a9cfce1ec7fc8fb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/1_from_char_code.ts", "frozen", true)
}
func TestProduct_CycleCurrent_4a9cfce1ec7fc8fb(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/gitignore/gaps/1_from_char_code.ts", "current", true)
}
func TestCyclesCorpus_4a9cfce1ec7fc8fb(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/gitignore/gaps/1_from_char_code.ts")
}
func TestProduct_CycleFrozen_50fb4800a964829e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/printer/gaps/nestedInner.ts", "frozen", true)
}
func TestProduct_CycleCurrent_50fb4800a964829e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/graphql/printer/gaps/nestedInner.ts", "current", true)
}
func TestCyclesCorpus_50fb4800a964829e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/graphql/printer/gaps/nestedInner.ts")
}
func TestProduct_CycleFrozen_e230441f2e2d1890(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/024.a", "frozen", true)
}
func TestProduct_CycleCurrent_e230441f2e2d1890(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/024.a", "current", true)
}
func TestCyclesCorpus_e230441f2e2d1890(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/dynamic/024.a")
}
func TestProduct_CycleFrozen_db01f7f5dbe26f24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/000_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_db01f7f5dbe26f24(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/000_runtime.a", "current", true)
}
func TestCyclesCorpus_db01f7f5dbe26f24(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/000_runtime.a")
}
func TestProduct_CycleFrozen_cdc749856c6b499a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/023.a", "frozen", true)
}
func TestProduct_CycleCurrent_cdc749856c6b499a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/023.a", "current", true)
}
func TestCyclesCorpus_cdc749856c6b499a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/023.a")
}
func TestProduct_CycleFrozen_477114bc4ebd5166(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/039_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_477114bc4ebd5166(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/039_runtime.a", "current", true)
}
func TestCyclesCorpus_477114bc4ebd5166(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/039_runtime.a")
}
func TestProduct_CycleFrozen_4d522e66a45ec9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/045.a", "frozen", true)
}
func TestProduct_CycleCurrent_4d522e66a45ec9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/045.a", "current", true)
}
func TestCyclesCorpus_4d522e66a45ec9ad(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/045.a")
}
func TestProduct_CycleFrozen_9a36adacc8b36a22(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/053_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_9a36adacc8b36a22(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/053_runtime.a", "current", true)
}
func TestCyclesCorpus_9a36adacc8b36a22(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/053_runtime.a")
}
func TestProduct_CycleFrozen_566f7e4f9a1eb848(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/075_runtime.a", "frozen", true)
}
func TestProduct_CycleCurrent_566f7e4f9a1eb848(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/075_runtime.a", "current", true)
}
func TestCyclesCorpus_566f7e4f9a1eb848(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/075_runtime.a")
}
func TestProduct_CycleFrozen_bee32c3cf97622dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/081.a", "frozen", true)
}
func TestProduct_CycleCurrent_bee32c3cf97622dc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/regex/testdata/shapes/plain/081.a", "current", true)
}
func TestCyclesCorpus_bee32c3cf97622dc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/regex/testdata/shapes/plain/081.a")
}
func TestProduct_CycleFrozen_5a8014bf5524aabc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-continue/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_5a8014bf5524aabc(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-continue/rule.ts", "current", true)
}
func TestCyclesCorpus_5a8014bf5524aabc(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-continue/rule.ts")
}
func TestProduct_CycleFrozen_8847293acdb2f2ed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-octal-escape/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_8847293acdb2f2ed(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-octal-escape/rule.ts", "current", true)
}
func TestCyclesCorpus_8847293acdb2f2ed(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-octal-escape/rule.ts")
}
func TestProduct_CycleFrozen_6700507d266a5372(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-script-url/rule.a", "frozen", true)
}
func TestProduct_CycleCurrent_6700507d266a5372(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/no-script-url/rule.a", "current", true)
}
func TestCyclesCorpus_6700507d266a5372(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/no-script-url/rule.a")
}
func TestProduct_CycleFrozen_95340aa18922f560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/prefer-template/rule.ts", "frozen", true)
}
func TestProduct_CycleCurrent_95340aa18922f560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/prefer-template/rule.ts", "current", true)
}
func TestCyclesCorpus_95340aa18922f560(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/prefer-template/rule.ts")
}
func TestProduct_CycleFrozen_0baf928bdff260f2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-no-find-dom-node/messages.ts", "frozen", true)
}
func TestProduct_CycleCurrent_0baf928bdff260f2(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/react-no-find-dom-node/messages.ts", "current", true)
}
func TestCyclesCorpus_0baf928bdff260f2(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/react-no-find-dom-node/messages.ts")
}
func TestProduct_CycleFrozen_1e1cb948aeb86cf4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/messages.a", "frozen", true)
}
func TestProduct_CycleCurrent_1e1cb948aeb86cf4(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/messages.a", "current", true)
}
func TestCyclesCorpus_1e1cb948aeb86cf4(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/lint/rules/typescript-adjacent-overload-signatures/messages.a")
}
func TestProduct_CycleFrozen_71abc6db168e1301(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/17_long_optional_chain.ts", "frozen", true)
}
func TestProduct_CycleCurrent_71abc6db168e1301(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/gaps/17_long_optional_chain.ts", "current", true)
}
func TestCyclesCorpus_71abc6db168e1301(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/gaps/17_long_optional_chain.ts")
}
func TestProduct_CycleFrozen_1be2ae564d8bbf78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/mdastCompile.ts", "frozen", true)
}
func TestProduct_CycleCurrent_1be2ae564d8bbf78(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/mdastCompile.ts", "current", true)
}
func TestCyclesCorpus_1be2ae564d8bbf78(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/mdastCompile.ts")
}
func TestProduct_CycleFrozen_7fce49a2469cb118(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/tokenSource.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7fce49a2469cb118(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdownblocks/tokenSource.ts", "current", true)
}
func TestCyclesCorpus_7fce49a2469cb118(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdownblocks/tokenSource.ts")
}
func TestProduct_CycleFrozen_7e0bbf17d1b69576(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/inline.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7e0bbf17d1b69576(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/markdowninline/inline.ts", "current", true)
}
func TestCyclesCorpus_7e0bbf17d1b69576(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/markdowninline/inline.ts")
}
func TestProduct_CycleFrozen_966fde45e48e9560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/gaps/5_conditional_panic.ts", "frozen", true)
}
func TestProduct_CycleCurrent_966fde45e48e9560(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/selector/gaps/5_conditional_panic.ts", "current", true)
}
func TestCyclesCorpus_966fde45e48e9560(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/selector/gaps/5_conditional_panic.ts")
}
func TestProduct_CycleFrozen_7c3b0cd953991b0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/gaps/2_call_through_optional_chain.ts", "frozen", true)
}
func TestProduct_CycleCurrent_7c3b0cd953991b0a(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/static_single_assignment/gaps/2_call_through_optional_chain.ts", "current", true)
}
func TestCyclesCorpus_7c3b0cd953991b0a(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/static_single_assignment/gaps/2_call_through_optional_chain.ts")
}
func TestProduct_CycleFrozen_dca4ba2f5fcd621e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/keys.ts", "frozen", true)
}
func TestProduct_CycleCurrent_dca4ba2f5fcd621e(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/tsprinter/keys.ts", "current", true)
}
func TestCyclesCorpus_dca4ba2f5fcd621e(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/tsprinter/keys.ts")
}
func TestProduct_CycleFrozen_c1df1385618a12f8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/stringUnitScan.ts", "frozen", true)
}
func TestProduct_CycleCurrent_c1df1385618a12f8(t *testing.T) {
	t.Parallel()
	cycleCorpusObservation(t, "stage1/cohere/yaml/gaps/stringUnitScan.ts", "current", true)
}
func TestCyclesCorpus_c1df1385618a12f8(t *testing.T) {
	t.Parallel()
	cycleCorpusVerify(t, "stage1/cohere/yaml/gaps/stringUnitScan.ts")
}

// END RECORDED CYCLE UNITS

const cycleCorpusDriver = `package main
import("context";"encoding/json";"fmt";"os";"path/filepath";"strings";"time";"github.com/system-inc/adamic/internal/load";"github.com/system-inc/adamic/internal/lower";"github.com/system-inc/adamic/stage1/cohere/lint/registry")
func main(){key:=os.Args[1];switch key{case "@panic":panic("cycle child mutant");case "@exit":fmt.Print("{}");os.Exit(7);case "@missing":return;case "@malformed":fmt.Print("{");return;case "@duplicate":fmt.Print("{}{}");return;case "@timeout":for{time.Sleep(time.Hour)}}
if _,err:=registry.Generate("stage1/cohere/lint");err!=nil{panic(err)}
status:="load";program,err:=load.Load([]string{key});if err==nil{status="lower";ctx,cancel:=context.WithTimeout(context.Background(),45*time.Second);defer cancel();_,err=lower.Lower(ctx,program)}
snapshots:=lower.CycleCorpusSnapshots();if len(snapshots)>0{status="finder"};message:="";if err!=nil{message=fmt.Sprintf("%T: %v",err,err)};root,_:=os.Getwd();message=strings.ReplaceAll(filepath.ToSlash(message),filepath.ToSlash(root)+"/","")
if err:=json.NewEncoder(os.Stdout).Encode(struct{Status string ` + "`json:\"status\"`" + `;Error string ` + "`json:\"error,omitempty\"`" + `;Snapshots []json.RawMessage ` + "`json:\"snapshots,omitempty\"`" + `}{status,message,snapshots});err!=nil{panic(err)}}
`

const cycleCorpusObserver = `package lower
import("encoding/json";"fmt";"os";"path/filepath";"reflect";"sort";"strings";"github.com/microsoft/TypeScript/tsc/shim/ast";"github.com/microsoft/TypeScript/tsc/shim/checker";"github.com/system-inc/adamic/internal/fresh")
var cycleCorpusSnapshots []json.RawMessage
func CycleCorpusSnapshots()[]json.RawMessage{return cycleCorpusSnapshots}
func cycleCorpusObserve(f *cycleFinder,result *error){
root,_:=os.Getwd();clean:=func(s string)string{return strings.ReplaceAll(filepath.ToSlash(s),filepath.ToSlash(root)+"/","")}
describe:=func(t *checker.Type)string{if t==nil{return ""};name:=fmt.Sprintf("%s [flags=%d object=%d]",f.l.checker.TypeToStringEx(t,nil,checker.TypeFormatFlagsNoTruncation,nil),t.Flags(),t.ObjectFlags()&0xffff);for _,p:=range f.l.checker.GetPropertiesOfType(t){name+=fmt.Sprintf(";%s:optional=%t,readonly=%t",p.Name,p.Flags&ast.SymbolFlagsOptional!=0,f.l.checker.IsReadonlySymbol(p))};return clean(name)}
type slot struct{Holder,Name,Target string;Weak,Reaches bool;Unproven *fresh.Write;Refusal string}
slots:=[]slot{};add:=func(holder,target *checker.Type,kind fresh.WriteKind,name,refusal string){s:=slot{Holder:describe(holder),Name:name,Target:describe(target),Weak:f.weak(target),Reaches:f.reaches(target,cycleNode{proven:holder}),Refusal:refusal};if !s.Weak&&s.Reaches{write:=name;if kind!=fresh.WriteField{write=""};s.Unproven=f.cycleUnproven(kind,holder,write)};slots=append(slots,s)}
for _,holder:=range f.seen{if holder.Flags()&checker.TypeFlagsObject==0||f.isFunction(holder){continue};refusal:="";if err:=f.slotsOf(holder);err!=nil{refusal=clean(fmt.Sprintf("%T: %v",err,err))};switch{case f.l.checker.IsArrayType(holder)||checker.IsTupleType(holder):for _,target:=range f.graphArguments(holder){add(holder,target,fresh.WriteElement,"[]",refusal)};case f.l.isLibraryType(holder,"Map","ReadonlyMap"):for i,target:=range f.graphArguments(holder){add(holder,target,fresh.WriteMapEntry,fmt.Sprintf("map[%d]",i),refusal)};case f.l.isLibraryType(holder,"Set","ReadonlySet"):for _,target:=range f.graphArguments(holder){add(holder,target,fresh.WriteSetElement,"set",refusal)};default:for _,field:=range f.fields(holder){add(holder,f.l.checker.GetTypeOfSymbol(field),fresh.WriteField,field.Name,refusal)}}}
for local,declared:=range f.l.result.Locals{if declared.Captured&&!declared.Global{if t:=f.l.localTypes[local];t!=nil{slots=append(slots,slot{Holder:fmt.Sprintf("cell:%d",local+1),Target:describe(t),Weak:f.weak(t),Reaches:f.reaches(t,cycleNode{cell:local+1})})}}}
sort.Slice(slots,func(i,j int)bool{a,_:=json.Marshal(slots[i]);b,_:=json.Marshal(slots[j]);return string(a)<string(b)})
identities:=map[int]string{};seen:=map[cycleNode]bool{};queue:=[]cycleNode{};for t:=range f.where{queue=append(queue,cycleNode{proven:t})};for _,t:=range f.shapes{queue=append(queue,cycleNode{proven:t})};for len(queue)>0{n:=queue[len(queue)-1];queue=queue[:len(queue)-1];if seen[n]{continue};seen[n]=true;if n.proven!=nil{identities[int(n.proven.Id())]=describe(n.proven)};queue=append(queue,f.graphLinks(n)...)}
identity:=func(id int)string{if id<=0{return fmt.Sprintf("cell:%d",-id)};if name,ok:=identities[id];ok{return name};panic(fmt.Sprintf("unresolved graph identity %d",id))}
graph:=map[string][]string{};var visit func(reflect.Value,string)
visit=func(v reflect.Value,path string){switch v.Kind(){case reflect.Interface,reflect.Pointer:if !v.IsNil(){visit(v.Elem(),path)};case reflect.Struct:for i:=0;i<v.NumField();i++{field:=v.Field(i);name:=v.Type().Field(i).Name;next:=path+"."+name;if name=="GraphTypes"{selected:=[]string{};if field.Kind()==reflect.Map{it:=field.MapRange();for it.Next(){selected=append(selected,fmt.Sprintf("%s=%t",identity(int(it.Key().Int())),it.Value().Bool()))}}else{for j:=0;j<field.Len();j++{id:=int(field.Index(j).Int());if f.l.result.GraphTypes[id]{selected=append(selected,identity(id))}}};sort.Strings(selected);if len(selected)>0{graph[next]=selected}}else if (name=="Graph"||name=="ProgramRegion")&&field.Kind()==reflect.Bool{if field.Bool(){graph[next]=[]string{"true"}}}else{visit(field,next)}};case reflect.Slice,reflect.Array:for i:=0;i<v.Len();i++{visit(v.Index(i),fmt.Sprintf("%s[%d]",path,i))}}}
visit(reflect.ValueOf(f.l.result),"IR");message:="";if *result!=nil{message=clean(fmt.Sprintf("%T: %v",*result,*result))};names:=map[string]bool{};for _,slot:=range slots{names[slot.Holder]=true;names[slot.Target]=true};for _,values:=range graph{for _,value:=range values{names[value]=true}};types:=[]string{};for name:=range names{types=append(types,name)};sort.Strings(types);refs:=map[string]string{};for i,name:=range types{refs[name]=fmt.Sprintf("type:%d",i)};for i:=range slots{slots[i].Holder=refs[slots[i].Holder];slots[i].Target=refs[slots[i].Target]};for path,values:=range graph{for i,value:=range values{values[i]=refs[value]};graph[path]=values};data,err:=json.Marshal(struct{Types []string;Slots []slot;Graph map[string][]string;Error string}{types,slots,graph,message});if err!=nil{panic(err)};cycleCorpusSnapshots=append(cycleCorpusSnapshots,data)
}
`
