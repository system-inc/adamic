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
	for _, name := range []string{
		"scanner",
		"local",
		"field",
		"observe",
		"null_equality",
		"null_loose",
		"null_truthiness",
		"null_typeof",
		"null_coalesce",
		"null_optional",
		"null_copy",
		"null_json",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/placeholder_nonnull_" + name + ".a", true, false})
	}
	for _, name := range []string{
		"before_use",
		"saved_leak",
		"null_before_use",
		"null_saved_leak",
		"assignment_result",
		"return_assignment",
		"alias_reset",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/placeholder_nonnull_" + name + ".a", true, true})
	}
}

func TestPlaceholderUseCheckBeforeUse(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(test, "before_use", "placeholder 'value' is unset at argument via value", 0)
}

func TestPlaceholderUseCheckSavedLeak(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(test, "saved_leak", "placeholder 'value' is unset at argument via saved", 0)
}

func TestPlaceholderUseCheckNullBeforeUse(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(test, "null_before_use", "placeholder 'value' is unset at argument via value", 0)
}

func TestPlaceholderUseCheckNullSavedLeak(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(test, "null_saved_leak", "placeholder 'value' is unset at argument via saved", 0)
}

func TestPlaceholderUseCheckAssignmentResult(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(
		test,
		"assignment_result",
		"placeholder 'target' is unset at assignment result via target = value",
		0,
	)
}

func TestPlaceholderUseCheckReturnAssignment(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(
		test,
		"return_assignment",
		"placeholder 'target' is unset at return via target = value",
		0,
	)
}

func TestPlaceholderUseCheckAliasReset(test *testing.T) {
	test.Parallel()
	assertPlaceholderUseCheck(test, "alias_reset", "placeholder 'value' is unset at assignment via source.value", 0)
}

func assertPlaceholderUseCheck(test *testing.T, name, message string, nodeExit int) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_"+name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := onSliceNode(test, path)
	if source.exitCode != nodeExit {
		test.Fatalf("Node: %#v", source)
	}
	native, _ := nativelyUncached(test, program)
	backend := onSliceJavaScript(test, program)
	if difference := disagreement(backend, native); difference != "" {
		test.Fatal(difference)
	}
	if native.exitCode != 70 || string(native.stderr) != "adamic: panic: "+message+"\n" {
		test.Fatalf("checked use: %#v", native)
	}
	checks := 0
	for _, site := range program.PlaceholderChecks {
		if site.Status == "checked" {
			checks++
		}
	}
	if checks != 1 {
		test.Fatalf("checked-sites report: %d checked sites, want 1", checks)
	}
}

func TestPlaceholderNonliteralAssertionStillChecks(test *testing.T) {
	test.Parallel()
	sourcePath := filepath.Join(repository, "internal/lower/testdata/placeholders/nonliteral.a")
	source, readError := os.ReadFile(sourcePath)
	if readError != nil {
		test.Fatal(readError)
	}
	// The extension chooses the existing checked .ts policy. The committed witness
	// remains .a; no compiler adaptation changes its non-literal assertion.
	path := filepath.Join(test.TempDir(), "checked-input.ts")
	if writeError := os.WriteFile(path, source, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	node := onSliceNode(test, path)
	if node.exitCode != 0 || string(node.stdout) != "after assertion\n" {
		test.Fatalf("Node: %#v", node)
	}
	native, _ := nativelyUncached(test, program)
	backend := onSliceJavaScript(test, program)
	if difference := disagreement(backend, native); difference != "" {
		test.Fatal(difference)
	}
	if native.exitCode != 70 || string(
		native.stderr,
	) != "adamic: panic: non-null assertion failed at "+path+":4:15: missing()! is null or undefined\n" {
		test.Fatalf("assertion: %#v", native)
	}
}

// Each mutant executes safely and matches the source Node output. Only the
// missing typed-use check distinguishes it from the pinned baseline.
func TestPlaceholderFlowMutantBeforeUse(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "before_use")
}

func TestPlaceholderFlowMutantSavedLeak(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "saved_leak")
}

func TestPlaceholderFlowMutantNullBeforeUse(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "null_before_use")
}

func TestPlaceholderFlowMutantNullSavedLeak(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "null_saved_leak")
}

func TestPlaceholderFlowMutantAssignmentResult(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "assignment_result")
}

func TestPlaceholderFlowMutantReturnAssignment(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "return_assignment")
}

func TestPlaceholderFlowMutantAliasReset(test *testing.T) {
	test.Parallel()
	assertPlaceholderFlowMutant(test, "alias_reset")
}

func assertPlaceholderFlowMutant(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_"+name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	baseline, _ := nativelyUncached(test, program)
	if baseline.exitCode != 70 {
		test.Fatalf("baseline did not check the use: %#v", baseline)
	}
	changed := 0
	drop := func(node any) any {
		if check, ok := node.(ir.Coalesce); ok && check.Panic != nil {
			if message, ok := check.Panic.(ir.StringConstant); ok &&
				strings.HasPrefix(program.Strings[message.Index], "placeholder '") {
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
		test.Fatalf("mutant changed %d checks, want 1", changed)
	}
	source := onSliceNode(test, path)
	mutant, binary := nativelyUncached(test, program)
	for _, got := range []run{mutant, onSliceJavaScript(test, program)} {
		if difference := disagreement(source, got); difference != "" {
			test.Fatalf("unchecked value did not reach the receiver: %s", difference)
		}
		if disagreement(baseline, got) == "" {
			test.Fatal("mutant escaped the pinned checked-use outcome")
		}
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf(
		"drop flow check lets unset through: exit %d stdout %q; caught by baseline exit 70",
		mutant.exitCode,
		mutant.stdout,
	)
}

// Collapsing null into undefined must change ordinary presence observations.
func TestPlaceholderNullTagMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_null_equality.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := onSliceNode(test, path)
	baseline, _ := nativelyUncached(test, program)
	if difference := disagreement(source, baseline); difference != "" {
		test.Fatal(difference)
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
		test.Fatalf("mutant changed %d null initializers, want 1", changed)
	}
	mutant, binary := nativelyUncached(test, program)
	backend := onSliceJavaScript(test, program)
	if difference := disagreement(mutant, backend); difference != "" {
		test.Fatal(difference)
	}
	if mutant.exitCode != 0 || disagreement(source, mutant) == "" {
		test.Fatalf("null tag mutant escaped Node: %#v", mutant)
	}
	if report := leaksUncached(test, program, binary); report != "" {
		test.Fatal(report)
	}
	test.Logf("null collapsed to undefined caught by Node: stdout %q", mutant.stdout)
}

// A copy observes the nullish payload; readiness is checked at its later typed use.
func TestPlaceholderSpreadReadinessMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/placeholder_nonnull_null_copy.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := onSliceNode(test, path)
	baseline, _ := nativelyUncached(test, program)
	if difference := disagreement(source, baseline); difference != "" {
		test.Fatal(difference)
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
		test.Fatal("mutant changed no nullish storage markers")
	}
	mutant, _ := nativelyUncached(test, program)
	backend := onSliceJavaScript(test, program)
	if mutant.exitCode != 70 ||
		backend.exitCode != 70 ||
		disagreement(source, mutant) == "" ||
		disagreement(source, backend) == "" {
		test.Fatalf("copy readiness mutant escaped Node: native %#v, JavaScript %#v", mutant, backend)
	}
	test.Logf("copy readiness mutant caught by Node exit 0: native and JavaScript exit 70")
}

func TestSlice5WitnessAliasReset(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_alias_reset.a")
}

func TestSlice5WitnessAssignmentResult(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_assignment_result.a")
}

func TestSlice5WitnessBeforeUse(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_before_use.a")
}

func TestSlice5WitnessNullBeforeUse(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_null_before_use.a")
}

func TestSlice5WitnessNullSavedLeak(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a")
}

func TestSlice5WitnessReturnAssignment(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_return_assignment.a")
}

func TestSlice5WitnessSavedLeak(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "internal/oracle/testdata/placeholder_nonnull_saved_leak.a")
}

func TestSlice5WitnessSpeculation(test *testing.T) {
	test.Parallel()
	assertSlice5Witness(test, "stage3/parser-next/speculation/misfit.a")
}

// A witness is exact source bytes and an exact stop; its type-correct repair must agree with Node.
func assertSlice5Witness(test *testing.T, fixture string) {
	test.Helper()
	data, readError := os.ReadFile(filepath.Join(repository, "cloud/admission-corpus/negative-witnesses.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var ledger struct {
		Witnesses map[string]struct {
			Declaration  string `json:"declaration"`
			RepairFrom   string `json:"repair_from"`
			RepairTo     string `json:"repair_to"`
			RepairSha256 string `json:"type_correct_sha256"`
			Exit         int    `json:"exit"`
			Stderr       string `json:"stderr"`
		} `json:"witnesses"`
	}
	if decodeError := json.Unmarshal(data, &ledger); decodeError != nil {
		test.Fatal(decodeError)
	}
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(source))
	entry, ok := ledger.Witnesses[hash]
	if !ok || !strings.HasPrefix(entry.Declaration, fixture+":") || entry.Exit != 70 {
		test.Fatalf("unlisted or mismatched witness %s", hash)
	}
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	originalNode := onSliceNode(test, path)
	if originalNode.exitCode != 0 {
		test.Fatalf("negative source Node: %+v", originalNode)
	}
	originalNative, _ := nativelyUncached(test, program)
	originalJavaScript := onSliceJavaScript(test, program)
	for _, actual := range []run{originalNative, originalJavaScript} {
		if actual.exitCode != entry.Exit || string(actual.stderr) != strings.ReplaceAll(entry.Stderr, "{root}", root) {
			test.Fatalf("ledger stop differs: %+v", actual)
		}
	}
	if !strings.Contains(string(source), entry.RepairFrom) {
		test.Fatal("repair changes no source")
	}
	repaired := strings.ReplaceAll(string(source), entry.RepairFrom, entry.RepairTo)
	repairHash := fmt.Sprintf("%x", sha256.Sum256([]byte(repaired)))
	if repairHash != entry.RepairSha256 {
		test.Fatalf("repair hash %s, want %s", repairHash, entry.RepairSha256)
	}
	if _, listed := ledger.Witnesses[repairHash]; listed {
		test.Fatal("type-correct repair is listed as a negative witness")
	}
	repairPath := filepath.Join(test.TempDir(), filepath.Base(path))
	if writeError := os.WriteFile(repairPath, []byte(repaired), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	repairProgram, loweredError := lowered(test, repairPath)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onSliceNode(test, repairPath)
	if truth.exitCode != 0 {
		test.Fatalf("repaired Node: %+v", truth)
	}
	actual, binary := nativelyUncached(test, repairProgram)
	for _, got := range []run{actual, onSliceJavaScript(test, repairProgram)} {
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf("type-correct repair: %s", difference)
		}
	}
	if report := leaksUncached(test, repairProgram, binary); report != "" {
		test.Fatal(report)
	}
	// A retained stopping artifact cannot become an exception after its source is repaired.
	for _, retained := range []run{originalNative, originalJavaScript} {
		if disagreement(truth, retained) == "" {
			test.Fatal("retained stopping artifact mutant escaped repaired-source agreement")
		}
	}
	test.Log(
		"exact witness and repair hashes held; repaired source agrees with Node under both backends and sanitizers; retained stopping artifact mutant rejected",
	)
}
