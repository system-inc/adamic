package lint

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const compilerAgreementCheckerGroupBudgetMillis = 10_000
const compilerAgreementCheckerGroupCount = 27

// Standalone checker.ts native CPU prices, rounded up to milliseconds: 6ef6b9dc,
// Clang 20.1.8, -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all,
// native.Options{Sanitize:true, Jobs:1}, Split=false, CPU 0, one core. The cut
// consumes these prices in production dispatch order; they are not group IDs.
var compilerAgreementCheckerPrices = map[string]int{
	// no-cond-assign: measured on this port over pinned checker.ts, CPU 0, Clang 20.1.8;
	// -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all, Split=false, Jobs=1.
	"no-cond-assign":                                    2710,
	"no-debugger":                                       2578,
	"no-empty":                                          2481,
	"eqeqeq":                                            2576,
	"no-var":                                            2409,
	"no-duplicate-case":                                 2460,
	"no-octal-escape":                                   2423,
	"no-unexpected-multiline":                           2469,
	"no-unused-private-class-members":                   2394,
	"no-useless-constructor":                            2462,
	"prefer-template":                                   3745,
	"react-forward-ref-uses-ref":                        2555,
	"react-jsx-no-comment-textnodes":                    2514,
	"react-no-find-dom-node":                            2592,
	"react-no-is-mounted":                               2552,
	"react-no-redundant-should-component-update":        2493,
	"typescript-adjacent-overload-signatures":           2568,
	"typescript-init-declarations":                      2410,
	"typescript-method-signature-style":                 3782,
	"typescript-no-this-alias":                          2441,
	"typescript-no-wrapper-object-types":                4019,
	"typescript-prefer-literal-enum-member":             2446,
	"adamic-no-type-predicate":                          2409,
	"base-boundary-no-global-container":                 2580,
	"base-consistency-no-console":                       2423,
	"base-consistency-no-hand-built-declared-error":     2514,
	"base-consistency-require-pagination-argument-name": 2514,
	"max-classes-per-file":                              2455,
	"max-depth":                                         2574,
	"max-nested-callbacks":                              2476,
	"nexus-consistency-no-abbreviated-identifier":       3926,
	"nexus-consistency-no-ambiguous-identifier":         2649,
	"nexus-consistency-no-enum":                         2416,
	"nexus-consistency-no-for-in":                       2477,
	"nexus-consistency-no-long-line-comment":            5246,
	"nexus-consistency-no-multiline-arrow-function":     6720,
	"nexus-consistency-no-shouting":                     3511,
	"nexus-consistency-no-single-line-jsdoc":            5137,
	"nexus-consistency-no-stuttering-name":              2702,
	"nexus-consistency-require-type-suffix":             2640,
	"no-await-in-loop":                                  2551,
	"no-bitwise":                                        2521,
	"no-caller":                                         2375,
	"no-continue":                                       2629,
	"no-div-regex":                                      2631,
	"no-empty-character-class":                          2533,
	"no-empty-pattern":                                  2448,
	"no-empty-static-block":                             2508,
	"no-ex-assign":                                      2580,
	"no-iterator":                                       2591,
	"no-labels":                                         2585,
	"no-multi-assign":                                   2515,
	"no-negated-condition":                              2520,
	"no-new":                                            2401,
	"no-plusplus":                                       2646,
	"no-proto":                                          2506,
	"no-return-assign":                                  2446,
	"no-script-url":                                     2504,
	"no-self-compare":                                   2832,
	"no-sequences":                                      2526,
	"no-sparse-arrays":                                  2608,
	"no-template-curly-in-string":                       2483,
	"no-underscore-dangle":                              2491,
	"no-unneeded-ternary":                               3773,
	"no-unsafe-finally":                                 2627,
	"no-unsafe-negation":                                2634,
	"no-unsafe-optional-chaining":                       2793,
	"no-useless-catch":                                  2539,
	"no-useless-concat":                                 2502,
	"no-warning-comments":                               3200,
	"no-with":                                           2519,
	"prefer-destructuring":                              3952,
	"require-yield":                                     2456,
	"unicode-bom":                                       2558,
	"vars-on-top":                                       2685,
}

type compilerAgreementCheckerGroup struct {
	rules  []registry.Descriptor
	millis int
}

// A comma at each end distinguishes a set even when it has only one rule.
// The port and Go manifest driver filter their existing registry in place, so
// selecting a group cannot reorder listeners or give each rule a separate tree.
func (g compilerAgreementCheckerGroup) selection() string {
	names := make([]string, len(g.rules))
	for i, d := range g.rules {
		names[i] = d.Name
	}
	return "," + strings.Join(names, ",") + ","
}

func compilerAgreementCutCheckerGroups(descriptors []registry.Descriptor) ([]compilerAgreementCheckerGroup, error) {
	var groups []compilerAgreementCheckerGroup
	var current compilerAgreementCheckerGroup
	for _, d := range descriptors {
		price := compilerAgreementCheckerPrices[d.Slug]
		if price <= 0 || price > compilerAgreementCheckerGroupBudgetMillis {
			return nil, fmt.Errorf("rule %s has standalone price %d ms; measure it before cutting groups", d.Slug, price)
		}
		if len(current.rules) > 0 && current.millis+price > compilerAgreementCheckerGroupBudgetMillis {
			groups = append(groups, current)
			current = compilerAgreementCheckerGroup{}
		}
		current.rules = append(current.rules, d)
		current.millis += price
	}
	if len(current.rules) > 0 {
		groups = append(groups, current)
	}
	if err := compilerAgreementCheckerGroupUnion(groups, descriptors); err != nil {
		return nil, err
	}
	return groups, nil
}

func compilerAgreementCheckerGroupUnion(groups []compilerAgreementCheckerGroup, descriptors []registry.Descriptor) error {
	known := map[string]bool{}
	for _, d := range descriptors {
		known[d.Name] = true
	}
	seen := map[string]int{}
	var order []string
	for part, g := range groups {
		if len(g.rules) == 0 {
			return fmt.Errorf("checker group %d is empty", part)
		}
		if g.millis <= 0 || g.millis > compilerAgreementCheckerGroupBudgetMillis {
			return fmt.Errorf("checker group %d costs %d ms", part, g.millis)
		}
		for _, d := range g.rules {
			if !known[d.Name] {
				return fmt.Errorf("checker group %d contains unregistered rule %s", part, d.Name)
			}
			seen[d.Name]++
			order = append(order, d.Name)
		}
	}
	for _, d := range descriptors {
		if seen[d.Name] != 1 {
			return fmt.Errorf("rule %s is in %d checker groups, want exactly one", d.Name, seen[d.Name])
		}
	}
	for i, d := range descriptors {
		if order[i] != d.Name {
			return fmt.Errorf("checker groups reorder production rule %s at %d", d.Name, i)
		}
	}
	return nil
}

func compilerAgreementCombinedGroups(t *testing.T) []compilerAgreementCheckerGroup {
	t.Helper()
	groups, err := compilerAgreementCutCheckerGroups(prepareRegistry(t, "."))
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != compilerAgreementCheckerGroupCount {
		t.Fatalf("greedy cut made %d checker groups, want %d; update the part declarations with the measured cut", len(groups), compilerAgreementCheckerGroupCount)
	}
	return groups
}

// Only side 4 expands the single checker/all case. Its other sides keep their
// whole-rule interaction case and their existing shared Go answer unchanged.
func compilerAgreementSideCases(t *testing.T, cases []compilerAgreementCase, side int) []compilerAgreementCase {
	if side != 4 {
		return cases
	}
	groups := compilerAgreementCombinedGroups(t)
	expanded := make([]compilerAgreementCase, 0, len(cases)+len(groups)-1)
	for _, c := range cases {
		if c.key == compilerAgreementCheckerKey && c.rule == "all" {
			for part, g := range groups {
				grouped := c
				grouped.rule = g.selection()
				grouped.group = part + 1
				expanded = append(expanded, grouped)
			}
		} else {
			expanded = append(expanded, c)
		}
	}
	return expanded
}

// compilerAgreementCheckerGroupRun is also the entry for a sanitizer witness:
// pass its native mutant binary here to run the exact same group and Go row.
// Executable preparation belongs to a keyed product, outside this helper's work.
func compilerAgreementCheckerGroupRun(t *testing.T, part int, binary string) (got, want execution) {
	t.Helper()
	groups := compilerAgreementCombinedGroups(t)
	if part < 0 || part >= len(groups) {
		t.Fatalf("invalid checker group %d", part)
	}
	owned := compilerAgreementCheckerAllOwned(t)
	if len(owned) != 1 {
		t.Fatalf("checker/all has %d cases, want one", len(owned))
	}
	group := groups[part]
	path := manifest(t, []string{owned[0].path + "\t" + group.selection()})
	oracle := lintGoOracleProduct(t)
	started := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(started).Seconds()) }()
	want = execute(t, "", oracle, "--manifest", path)
	got = execute(t, "", binary, "--manifest", path)
	t.Logf("side 4: %s, oracle %s, checker group %d, %d rules, standalone price %d ms", got.duration, want.duration, part, len(group.rules), group.millis)
	return got, want
}

const compilerAgreementCombinedOutputFrom = "Unexpected use of"
const compilerAgreementCombinedOutputTo = "PLANTED witness unexpected use of"
const compilerAgreementCombinedOutputFile = "rules/no-bitwise/rule.ts"

func compilerAgreementCombinedOutputNative(t *testing.T) string {
	t.Helper()
	options := native.Options{Sanitize: true, Jobs: 1}
	flags := append(native.Flags(options), "Split=false", "Jobs=1", compilerAgreementCombinedOutputFile, compilerAgreementCombinedOutputFrom, compilerAgreementCombinedOutputTo)
	inputs := compilerAgreementInputs(t, "compiler-agreement-combined-output-mutant-v1", flags)
	inputs.Files = append(append([]string(nil), inputs.Files...), "stage1/cohere/lint/compiler_stage1_combined_sanitized_test.go", "stage1/cohere/lint/lint_test.go")
	product := compilerAgreementProduct(t, inputs, func(directory string) error {
		source, err := os.MkdirTemp("", "compiler-agreement-output-mutant-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(source)
		copyPort(t, source, compilerAgreementCombinedOutputFrom, compilerAgreementCombinedOutputTo, compilerAgreementCombinedOutputFile)
		prepareRegistry(t, source)
		program, err := load.Load([]string{filepath.Join(source, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		c := native.C(lowered)
		if err := os.WriteFile(filepath.Join(directory, "lint.c"), []byte(c), 0644); err != nil {
			return err
		}
		return native.Build(c, filepath.Join(directory, "lint"), options)
	})
	return filepath.Join(product, "lint")
}

func TestProduct_CompilerAgreementCombinedOutputMutant(t *testing.T) {
	t.Parallel()
	compilerAgreementMeasureProduct(t, func(t *testing.T) string {
		binary := compilerAgreementCombinedOutputNative(t)
		requireLintArtifacts(t, filepath.Dir(binary), "lint.c")
		requireLintExecutable(t, binary)
		return binary
	})
}

func compilerAgreementCombinedOutputWitness(t *testing.T, part int) {
	t.Helper()
	binary := compilerAgreementCombinedOutputNative(t)
	got, want := compilerAgreementCheckerGroupRun(t, part, binary)
	red := false
	for _, d := range compilerAgreementCombinedGroups(t)[part].rules {
		if d.Slug == "no-bitwise" {
			red = true
		}
	}
	equal := bytes.Equal(got.output, want.output)
	if equal == red {
		t.Fatalf("output mutant group %d agreement=%t, want agreement=%t: %s", part, equal, !red, difference(got.output, want.output))
	}
	t.Logf("output mutant group %d: agreement=%t; no-bitwise owner=%t", part, equal, red)
}

var compilerAgreementCombinedWitnessWrapper = regexp.MustCompile(`(?m)^func TestCompilerAndStage1AgreeCombinedOutputWitness_([0-9]+)\(t \*testing.T\) \{\s*t.Parallel\(\)\s*compilerAgreementCombinedOutputWitness\(t, ([0-9]+)\)\s*\}`)
var compilerAgreementCombinedWitnessDeclaration = regexp.MustCompile(`(?m)^func TestCompilerAndStage1AgreeCombinedOutputWitness_[0-9]+\(`)

func TestCompilerAndStage1AgreeCombinedGroupsUnion(t *testing.T) {
	t.Parallel()
	descriptors := prepareRegistry(t, ".")
	groups := compilerAgreementCombinedGroups(t)
	if err := compilerAgreementCheckerGroupUnion(groups, descriptors); err != nil {
		t.Fatal(err)
	}
	// Exercise the same validator on actual dropped and duplicated memberships.
	for _, mutant := range []string{"dropped", "duplicated", "empty"} {
		broken := append([]compilerAgreementCheckerGroup(nil), groups...)
		broken[0].rules = append([]registry.Descriptor(nil), groups[0].rules...)
		switch mutant {
		case "dropped":
			broken[0].rules = broken[0].rules[1:]
		case "duplicated":
			broken[0].rules = append(broken[0].rules, broken[0].rules[0])
		case "empty":
			broken[0].rules = nil
		}
		if err := compilerAgreementCheckerGroupUnion(broken, descriptors); err == nil {
			t.Fatalf("%s membership survived checker group union", mutant)
		}
	}
	source, err := os.ReadFile("compiler_stage1_combined_sanitized_test.go")
	if err != nil {
		t.Fatal(err)
	}
	wrappers := compilerAgreementCombinedWitnessWrapper.FindAllSubmatch(source, -1)
	if len(wrappers) != len(groups) || len(compilerAgreementCombinedWitnessDeclaration.FindAll(source, -1)) != len(wrappers) {
		t.Fatalf("%d output witness wrappers for %d groups", len(wrappers), len(groups))
	}
	seen := map[int]bool{}
	for _, m := range wrappers {
		part, _ := strconv.Atoi(string(m[1]))
		runs, _ := strconv.Atoi(string(m[2]))
		if part != runs || part >= len(groups) || seen[part] {
			t.Fatalf("invalid output witness group %s runs %s", m[1], m[2])
		}
		seen[part] = true
	}
	t.Logf("union: %d registered rules in exactly one of %d consecutive checker groups; each group has an output witness", len(descriptors), len(groups))
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_00(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 0)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_01(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 1)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_02(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 2)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_03(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 3)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_04(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 4)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_05(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 5)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_06(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 6)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_07(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 7)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_08(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 8)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_09(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 9)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_10(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 10)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_11(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 11)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_12(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 12)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_13(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 13)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_14(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 14)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_15(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 15)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_16(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 16)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_17(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 17)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_18(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 18)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_19(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 19)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_20(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 20)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_21(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 21)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_22(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 22)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_23(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 23)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_24(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 24)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_25(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 25)
}

func TestCompilerAndStage1AgreeCombinedOutputWitness_26(t *testing.T) {
	t.Parallel()
	compilerAgreementCombinedOutputWitness(t, 26)
}
