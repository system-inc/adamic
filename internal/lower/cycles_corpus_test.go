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
// Unregistered shards remain explicitly pending; only complete recordings are registered.
var cycleCorpusRegistered = []int{0}
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
