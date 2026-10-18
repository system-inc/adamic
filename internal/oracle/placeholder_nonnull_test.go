package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	for _, name := range []string{"scanner", "local", "field", "observe", "null_equality", "null_loose", "null_truthiness", "null_typeof", "null_coalesce", "null_optional", "null_copy", "null_json"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/placeholder_nonnull_" + name + ".a", true, false})
	}
	for _, name := range []string{"before_use", "saved_leak", "null_before_use", "null_saved_leak", "assignment_result", "return_assignment", "alias_reset"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/placeholder_nonnull_" + name + ".a", true, true})
	}
}

func TestPlaceholderUseCheckBeforeUse(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "before_use", "placeholder 'value' is unset at argument via value", 0)
}

func TestPlaceholderUseCheckSavedLeak(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "saved_leak", "placeholder 'value' is unset at argument via saved", 0)
}

func TestPlaceholderUseCheckNullBeforeUse(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "null_before_use", "placeholder 'value' is unset at argument via value", 0)
}

func TestPlaceholderUseCheckNullSavedLeak(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "null_saved_leak", "placeholder 'value' is unset at argument via saved", 0)
}

func TestPlaceholderUseCheckAssignmentResult(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "assignment_result", "placeholder 'target' is unset at assignment result via target = value", 0)
}

func TestPlaceholderUseCheckReturnAssignment(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "return_assignment", "placeholder 'target' is unset at return via target = value", 0)
}

func TestPlaceholderUseCheckAliasReset(t *testing.T) {
	t.Parallel()
	assertPlaceholderUseCheck(t, "alias_reset", "placeholder 'value' is unset at assignment via source.value", 0)
}

func assertPlaceholderUseCheck(t *testing.T, name, message string, nodeExit int) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onSliceNode(t, path)
	if source.exitCode != nodeExit {
		t.Fatalf("Node: %#v", source)
	}
	native, _ := nativelyUncached(t, program)
	backend := onSliceJavaScript(t, program)
	if difference := disagreement(backend, native); difference != "" {
		t.Fatal(difference)
	}
	if native.exitCode != 70 || string(native.stderr) != "adamic: panic: "+message+"\n" {
		t.Fatalf("checked use: %#v", native)
	}
	checks := 0
	for _, site := range program.PlaceholderChecks {
		if site.Status == "checked" {
			checks++
		}
	}
	if checks != 1 {
		t.Fatalf("checked-sites report: %d checked sites, want 1", checks)
	}
}

func TestPlaceholderNonliteralAssertionStillChecks(t *testing.T) {
	t.Parallel()
	sourcePath := filepath.Join(repository, "internal/lower/testdata/placeholders/nonliteral.a")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	// The extension chooses the existing checked .ts policy. The committed witness
	// remains .a; no compiler adaptation changes its non-literal assertion.
	path := filepath.Join(t.TempDir(), "checked-input.ts")
	if err := os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	node := onSliceNode(t, path)
	if node.exitCode != 0 || string(node.stdout) != "after assertion\n" {
		t.Fatalf("Node: %#v", node)
	}
	native, _ := nativelyUncached(t, program)
	backend := onSliceJavaScript(t, program)
	if difference := disagreement(backend, native); difference != "" {
		t.Fatal(difference)
	}
	if native.exitCode != 70 || string(native.stderr) != "adamic: panic: non-null assertion failed at "+path+":4:15: missing()! is null or undefined\n" {
		t.Fatalf("assertion: %#v", native)
	}
}

// Each mutant executes safely and matches the source Node output. Only the
// missing typed-use check distinguishes it from the pinned baseline.
func TestPlaceholderFlowMutantBeforeUse(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "before_use")
}

func TestPlaceholderFlowMutantSavedLeak(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "saved_leak")
}

func TestPlaceholderFlowMutantNullBeforeUse(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "null_before_use")
}

func TestPlaceholderFlowMutantNullSavedLeak(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "null_saved_leak")
}

func TestPlaceholderFlowMutantAssignmentResult(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "assignment_result")
}

func TestPlaceholderFlowMutantReturnAssignment(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "return_assignment")
}

func TestPlaceholderFlowMutantAliasReset(t *testing.T) {
	t.Parallel()
	assertPlaceholderFlowMutant(t, "alias_reset")
}

func assertPlaceholderFlowMutant(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	baseline, _ := nativelyUncached(t, program)
	if baseline.exitCode != 70 {
		t.Fatalf("baseline did not check the use: %#v", baseline)
	}
	changed := 0
	drop := func(node any) any {
		if check, ok := node.(ir.Coalesce); ok && check.Panic != nil {
			if message, ok := check.Panic.(ir.StringConstant); ok && strings.HasPrefix(program.Strings[message.Index], "placeholder '") {
				changed++
				return check.Value
			}
		}
		return node
	}
	program.Main = mutateReadiness(program.Main, drop)
	for index := range program.Functions {
		program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, drop)
	}
	if changed != 1 {
		t.Fatalf("mutant changed %d checks, want 1", changed)
	}
	source := onSliceNode(t, path)
	mutant, binary := nativelyUncached(t, program)
	for _, got := range []run{mutant, onSliceJavaScript(t, program)} {
		if difference := disagreement(source, got); difference != "" {
			t.Fatalf("unchecked value did not reach the receiver: %s", difference)
		}
		if disagreement(baseline, got) == "" {
			t.Fatal("mutant escaped the pinned checked-use outcome")
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("drop flow check lets unset through: exit %d stdout %q; caught by baseline exit 70", mutant.exitCode, mutant.stdout)
}

// Collapsing null into undefined must change ordinary presence observations.
func TestPlaceholderNullTagMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_null_equality.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onSliceNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(source, baseline); difference != "" {
		t.Fatal(difference)
	}
	changed := 0
	program.Main = mutateReadiness(program.Main, func(node any) any {
		if boxed, ok := node.(ir.Box); ok {
			if _, null := boxed.Value.(ir.Null); null {
				changed++
				return ir.Undefined{Of: ir.Union}
			}
		}
		return node
	})
	if changed != 1 {
		t.Fatalf("mutant changed %d null initializers, want 1", changed)
	}
	mutant, binary := nativelyUncached(t, program)
	backend := onSliceJavaScript(t, program)
	if difference := disagreement(mutant, backend); difference != "" {
		t.Fatal(difference)
	}
	if mutant.exitCode != 0 || disagreement(source, mutant) == "" {
		t.Fatalf("null tag mutant escaped Node: %#v", mutant)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	t.Logf("null collapsed to undefined caught by Node: stdout %q", mutant.stdout)
}

// A copy observes the nullish payload; readiness is checked at its later typed use.
func TestPlaceholderSpreadReadinessMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_null_copy.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onSliceNode(t, path)
	baseline, _ := nativelyUncached(t, program)
	if difference := disagreement(source, baseline); difference != "" {
		t.Fatal(difference)
	}
	changed := 0
	drop := func(node any) any {
		switch node := node.(type) {
		case ir.Field:
			if node.Unset {
				node.Unset = false
				changed++
				return node
			}
		case ir.SetProperty:
			if node.Unset {
				node.Unset = false
				changed++
				return node
			}
		}
		return node
	}
	program.Main = mutateReadiness(program.Main, drop)
	for index := range program.Functions {
		program.Functions[index].Body = mutateReadiness(program.Functions[index].Body, drop)
	}
	if changed == 0 {
		t.Fatal("mutant changed no nullish storage markers")
	}
	mutant, _ := nativelyUncached(t, program)
	backend := onSliceJavaScript(t, program)
	if mutant.exitCode != 70 || backend.exitCode != 70 || disagreement(source, mutant) == "" || disagreement(source, backend) == "" {
		t.Fatalf("copy readiness mutant escaped Node: native %#v, JavaScript %#v", mutant, backend)
	}
	t.Logf("copy readiness mutant caught by Node exit 0: native and JavaScript exit 70")
}

func TestSlice5WitnessAliasReset(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_alias_reset.a")
}

func TestSlice5WitnessAssignmentResult(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_assignment_result.a")
}

func TestSlice5WitnessBeforeUse(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_before_use.a")
}

func TestSlice5WitnessNullBeforeUse(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_null_before_use.a")
}

func TestSlice5WitnessNullSavedLeak(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a")
}

func TestSlice5WitnessReturnAssignment(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_return_assignment.a")
}

func TestSlice5WitnessSavedLeak(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a")
}

func TestSlice5WitnessSpeculation(t *testing.T) {
	t.Parallel()
	assertSlice5Witness(t, "stage3/parser-next/speculation/misfit.a")
}

// A witness is exact source bytes and an exact stop; its type-correct repair must agree with Node.
func assertSlice5Witness(t *testing.T, fixture string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repository, "cloud/admission-corpus/negative-witnesses.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Witnesses map[string]struct {
			Declaration  string `json:"declaration"`
			RepairFrom   string `json:"repair_from"`
			RepairTo     string `json:"repair_to"`
			RepairSHA256 string `json:"type_correct_sha256"`
			Exit         int    `json:"exit"`
			Stderr       string `json:"stderr"`
		} `json:"witnesses"`
	}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(source))
	entry, ok := ledger.Witnesses[hash]
	if !ok || !strings.HasPrefix(entry.Declaration, fixture+":") || entry.Exit != 70 {
		t.Fatalf("unlisted or mismatched witness %s", hash)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	originalNode := onSliceNode(t, path)
	if originalNode.exitCode != 0 {
		t.Fatalf("negative source Node: %+v", originalNode)
	}
	originalNative, _ := nativelyUncached(t, program)
	originalJavaScript := onSliceJavaScript(t, program)
	for _, actual := range []run{originalNative, originalJavaScript} {
		if actual.exitCode != entry.Exit || string(actual.stderr) != strings.ReplaceAll(entry.Stderr, "{root}", root) {
			t.Fatalf("ledger stop differs: %+v", actual)
		}
	}
	if !strings.Contains(string(source), entry.RepairFrom) {
		t.Fatal("repair changes no source")
	}
	repaired := strings.ReplaceAll(string(source), entry.RepairFrom, entry.RepairTo)
	repairHash := fmt.Sprintf("%x", sha256.Sum256([]byte(repaired)))
	if repairHash != entry.RepairSHA256 {
		t.Fatalf("repair hash %s, want %s", repairHash, entry.RepairSHA256)
	}
	if _, listed := ledger.Witnesses[repairHash]; listed {
		t.Fatal("type-correct repair is listed as a negative witness")
	}
	repairPath := filepath.Join(t.TempDir(), filepath.Base(path))
	if err := os.WriteFile(repairPath, []byte(repaired), 0600); err != nil {
		t.Fatal(err)
	}
	repairProgram, err := lowered(t, repairPath)
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, repairPath)
	if truth.exitCode != 0 {
		t.Fatalf("repaired Node: %+v", truth)
	}
	actual, binary := nativelyUncached(t, repairProgram)
	for _, got := range []run{actual, onSliceJavaScript(t, repairProgram)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("type-correct repair: %s", difference)
		}
	}
	if report := leaksUncached(t, repairProgram, binary); report != "" {
		t.Fatal(report)
	}
	// A retained stopping artifact cannot become an exception after its source is repaired.
	for _, retained := range []run{originalNative, originalJavaScript} {
		if disagreement(truth, retained) == "" {
			t.Fatal("retained stopping artifact mutant escaped repaired-source agreement")
		}
	}
	t.Log("exact witness and repair hashes held; repaired source agrees with Node under both backends and sanitizers; retained stopping artifact mutant rejected")
}
