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

// A recording piece adds its number only after declaring its per-fixture products.
// Zero registered shards is explicit: the harness does not claim a recorded corpus.
var cycleCorpusRegistered = []int{}
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
