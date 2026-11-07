package fuzz

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestMovesFeatureCoverage(test *testing.T) {
	test.Parallel()
	shapes := map[string]int{}
	enabled := 0
	for seed := uint64(1); seed <= 300; seed++ {
		program := GenerateMoves(seed)
		source := program.Source()
		if source != GenerateMoves(seed).Source() {
			test.Fatalf("seed %d is not deterministic", seed)
		}
		for _, line := range strings.Split(source, "\n") {
			if strings.HasPrefix(line, "// moves-shape: ") {
				shapes[strings.TrimPrefix(line, "// moves-shape: ")]++
			}
		}
		what, fix := moveRefusal(source)
		if what != program.Refusal || fix != program.RefusalFix {
			test.Fatalf("seed %d lost its exact expectation", seed)
		}
		clone := program.Clone()
		if clone.Source() != source || clone.RefusalFix != fix || clone.Feature != "moves" {
			test.Fatalf("seed %d clone lost move metadata", seed)
		}
		if strings.Contains(Generate(seed).Source(), "// moves-shape:") {
			enabled++
		}
		for _, disabled := range []string{"moves", "parallel"} {
			if strings.Contains(GenerateWithout(seed, []string{disabled}).Source(), "// moves-shape:") {
				test.Fatalf("seed %d: -without %s still generates moves", seed, disabled)
			}
		}
	}
	for _, kind := range []string{
		"accepted",
		"after",
		"alias",
		"global",
		"closure",
		"nested-object",
		"nested-array",
		"nested-map",
	} {
		if shapes[kind] == 0 {
			test.Fatalf("no %s among 300 move seeds", kind)
		}
		test.Logf("%s: %d", kind, shapes[kind])
	}
	if enabled == 0 {
		test.Fatal("moves never enabled in the ordinary generator")
	}
}

// The default is a smoke run; ADAMIC_FUZZ_MOVES_SEEDS requests a full campaign
// or the same bounded seed search against a compiler mutant in a Go overlay.
func TestMovesCampaign_1(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 1)
}
func TestMovesCampaign_2(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 2)
}
func TestMovesCampaign_3(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 3)
}
func TestMovesCampaign_4(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 4)
}
func TestMovesCampaign_5(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 5)
}
func TestMovesCampaign_6(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 6)
}
func TestMovesCampaign_7(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 7)
}
func TestMovesCampaign_8(test *testing.T) {
	test.Parallel()
	movesCampaign(test, 8)
}

const movesCampaignTests = 8

func movesCampaignSeeds(first, count int) []uint64 {
	var seeds []uint64
	for seed := uint64(first); seed <= uint64(count); seed += movesCampaignTests {
		seeds = append(seeds, seed)
	}
	return seeds
}

func TestMovesCampaignSeedCoverage(test *testing.T) {
	test.Parallel()
	// Read the actual test declarations so removing a campaign loses its seeds.
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "moves_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	for _, count := range []int{8, 37} {
		owners := make([]int, count+1)
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(function.Name.Name, "TestMovesCampaign_") {
				continue
			}
			calls := 0
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				name, ok := call.Fun.(*ast.Ident)
				if !ok || name.Name != "movesCampaign" {
					return true
				}
				calls++
				if len(call.Args) != 2 {
					test.Fatalf("%s: invalid campaign call", function.Name.Name)
				}
				literal, ok := call.Args[1].(*ast.BasicLit)
				if !ok {
					test.Fatalf("%s: campaign seed must be literal", function.Name.Name)
				}
				first, parseError := strconv.Atoi(literal.Value)
				if parseError != nil || first < 1 || first > movesCampaignTests {
					test.Fatalf("%s: invalid first seed", function.Name.Name)
				}
				seeds := movesCampaignSeeds(first, count)
				test.Logf("count %d %s: %v", count, function.Name.Name, seeds)
				for _, seed := range seeds {
					owners[seed]++
				}
				return true
			})
			if calls != 1 {
				test.Fatalf("%s: got %d campaign calls", function.Name.Name, calls)
			}
		}
		for seed := 1; seed <= count; seed++ {
			if owners[seed] != 1 {
				test.Errorf("count %d seed %d has %d owners, want 1", count, seed, owners[seed])
			}
		}
	}
}

// A seed has exactly one owner, including an explicitly requested longer campaign.
func movesCampaign(test *testing.T, first int) {
	count := 8
	if value := os.Getenv("ADAMIC_FUZZ_MOVES_SEEDS"); value != "" {
		var parseError error
		count, parseError = strconv.Atoi(value)
		if parseError != nil || count < 1 {
			test.Fatalf("invalid ADAMIC_FUZZ_MOVES_SEEDS %q", value)
		}
	}
	directory := test.TempDir()
	checkout, prepareBuiltError := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if prepareBuiltError != nil {
		test.Fatal(prepareBuiltError)
	}
	if runtime.GOOS == "linux" && checkout.TSan != "ready" {
		test.Fatalf("TSan required for Linux move campaign: %s", checkout.TSan)
	}
	test.Logf("thread sanitizer: %s", checkout.TSan)
	started := time.Now()
	totals := map[Verdict]int{}
	for _, seed := range movesCampaignSeeds(first, count) {
		program := GenerateMoves(seed)
		outcome := checkout.Try(program.Source(), filepath.Join(directory, "program"))
		expected := Agreed
		if program.Refusal != "" {
			expected = Refused
		}
		if outcome.Verdict != expected {
			test.Fatalf(
				"seed %d after %s: %s %s\n%s",
				seed,
				time.Since(started),
				outcome.Verdict,
				outcome.Key,
				outcome.Detail,
			)
		}
		totals[outcome.Verdict]++
		test.Logf("seed %d: %s %s", seed, outcome.Verdict, outcome.Key)
	}
	test.Logf(
		"%d move seeds in %s: agreed %d refused %d",
		totals[Agreed]+totals[Refused],
		time.Since(started),
		totals[Agreed],
		totals[Refused],
	)
}

func TestMovesRunnerCanFail(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	checkout, prepareBuiltError := PrepareBuilt("../..", buildcache.Adamic(test), filepath.Join(directory, "checkout"))
	if prepareBuiltError != nil {
		test.Fatal(prepareBuiltError)
	}
	var accepted, refused *Program
	for seed := uint64(1); seed <= 300 && (accepted == nil || refused == nil); seed++ {
		program := GenerateMoves(seed)
		if program.Refusal == "" {
			accepted = program
		} else {
			refused = program
		}
	}
	if accepted == nil || refused == nil {
		test.Fatal("missing move witnesses")
	}
	for name, source := range map[string]string{
		"path-prefix": strings.Replace(refused.Source(), refused.Refusal, strings.Split(refused.Refusal, ":")[0], 1),
		"wrong-fix": strings.Replace(
			refused.Source(),
			movesFixPrefix+" "+refused.RefusalFix,
			movesFixPrefix+" "+oppositeMoveFix(refused.RefusalFix),
			1,
		),
		"not-a-refusal": accepted.Source() +
			"\n" +
			movesRefusePrefix +
			" cannot move items: whole reachable ownership is not proven\n" +
			movesFixPrefix +
			" return it through the results\n",
	} {
		outcome := checkout.Try(source, filepath.Join(directory, name))
		if outcome.Verdict != Finding {
			test.Fatalf("%s: false success: %+v", name, outcome)
		}
		if name != "not-a-refusal" && outcome.Key != "move refusal differs" {
			test.Fatalf("%s failed for the wrong reason: %+v", name, outcome)
		}
	}
	path := filepath.Join(directory, "finding.a")
	if writeFindingError := WriteFinding(path, accepted, 1, nil, nil, "example"); writeFindingError != nil {
		test.Fatal(writeFindingError)
	}
	bytes, readError := os.ReadFile(path)
	if readError != nil || !strings.Contains(string(bytes), "-only-moves") {
		test.Fatal(fmt.Sprintf("focused replay lost: %v %s", readError, bytes))
	}
}

func oppositeMoveFix(fix string) string {
	if fix == "return it through the results" {
		return "don't use it after the parallelMap"
	}
	return "return it through the results"
}

// Even a predicate fooled by the expected-refusal comment must not shrink a
// move finding down to that comment alone, with its task boundary gone.
func TestMovesShrinkKeepsBoundary(test *testing.T) {
	test.Parallel()
	program := GenerateMoves(2)
	if program.RefusalFix == "" {
		test.Fatal("seed 2 is not the alias witness")
	}
	shrunk := Shrink(program, "accepted refusal", func(candidate *Program) Outcome {
		if strings.Contains(candidate.Source(), movesRefusePrefix) &&
			strings.Contains(candidate.Source(), movesFixPrefix) {
			return Outcome{Verdict: Finding, Key: "accepted refusal"}
		}
		return Outcome{Verdict: Agreed}
	})
	if !strings.Contains(shrunk.Source(), "parallelMap(") {
		test.Fatalf("lost the boundary: %s", shrunk.Source())
	}
}
