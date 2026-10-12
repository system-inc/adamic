package lint

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// One rule over checker.ts under ASan is the indivisible sanitized case, so each rule's case is its own test, named by
// the rule's slug: no numbered shard selects it (its unit is compilerAgreementRuleOwned), and no hash group or list
// position decides which test runs it.
const compilerAgreementRuleOwned = -1
const compilerAgreementCheckerSanitizedPrefix = "TestCompilerAndStage1Agree_CheckerSanitized_"
const compilerAgreementSanitizedPlantedRule = "no-debugger"

var compilerAgreementCheckerSanitizedDeclaration = regexp.MustCompile(`(?m)^func ` + compilerAgreementCheckerSanitizedPrefix)
var compilerAgreementCheckerSanitizedWrapper = regexp.MustCompile(`(?m)^func ` + compilerAgreementCheckerSanitizedPrefix + `(\w+)\(t \*testing.T\) \{\s*t.Parallel\(\)\s*;?\s*compilerAgreementCheckerSanitized\(t, "([a-z0-9-]+)"\)\s*\}`)

func compilerAgreementUnit(c compilerAgreementCase, side int) int {
	if c.key == compilerAgreementCheckerKey && c.rule != "all" && side == 4 {
		return compilerAgreementRuleOwned
	}
	return compilerAgreementOwner(c)*compilerAgreementSides + side
}

// compilerAgreementCheckerSanitizedOwners reads the per-rule wrappers from source and maps each rule's name to the test
// that owns its sanitized checker.ts case. Every registered rule has exactly one wrapper, named by its slug, and every
// declared test of the family has the wrapper's shape, so none runs outside the union.
func compilerAgreementCheckerSanitizedOwners(source []byte, descriptors []registry.Descriptor) (map[string]string, error) {
	wrappers := compilerAgreementCheckerSanitizedWrapper.FindAllSubmatch(source, -1)
	if declared := len(compilerAgreementCheckerSanitizedDeclaration.FindAll(source, -1)); len(wrappers) != declared {
		return nil, fmt.Errorf("%d of %d sanitized checker tests have the wrapper shape", len(wrappers), declared)
	}
	names := map[string]string{}
	for _, d := range descriptors {
		names[d.Slug] = d.Name
	}
	owners := map[string]string{}
	for _, m := range wrappers {
		test, slug := compilerAgreementCheckerSanitizedPrefix+string(m[1]), string(m[2])
		name, ok := names[slug]
		if !ok {
			return nil, fmt.Errorf("%s names unregistered rule %s", test, slug)
		}
		if test != compilerAgreementCheckerSanitizedPrefix+strings.ReplaceAll(slug, "-", "_") {
			return nil, fmt.Errorf("%s runs rule %s", test, slug)
		}
		if owners[name] != "" {
			return nil, fmt.Errorf("rule %s is owned by %s and %s", slug, owners[name], test)
		}
		owners[name] = test
	}
	for _, d := range descriptors {
		if owners[d.Name] == "" {
			return nil, fmt.Errorf("rule %s has no sanitized checker test", d.Slug)
		}
	}
	return owners, nil
}

// compilerAgreementCheckerSanitized runs one rule over checker.ts on the sanitized native binary and compares its bytes
// with the Go oracle's answer for the same row.
func compilerAgreementCheckerSanitized(t *testing.T, slug string) {
	t.Helper()
	name := ""
	for _, d := range prepareRegistry(t, ".") {
		if d.Slug == slug {
			name = d.Name
		}
	}
	if name == "" {
		t.Fatalf("sanitized checker test names unregistered rule %s", slug)
	}
	var owned []compilerAgreementCase
	for _, c := range compilerAgreementCases(t) {
		if compilerAgreementUnit(c, 4) == compilerAgreementRuleOwned && c.rule == name {
			owned = append(owned, c)
		}
	}
	if len(owned) != 1 {
		t.Fatalf("rule %s owns %d sanitized checker.ts cases, want 1", name, len(owned))
	}
	if os.Getenv("ADAMIC_COMPILER_AGREEMENT_PROBE") == "1" {
		want, got := []byte("case 0\nfixed\tunchanged\n"), []byte("case 0\nfixed\tunchanged\n")
		if owned[0].key == compilerAgreementCheckerKey && owned[0].rule == compilerAgreementSanitizedPlantedRule {
			got = []byte("case 0\nfixed\tdisagreement\n")
		}
		compilerAgreementCompare(t, got, want)
		return
	}
	path := manifest(t, []string{owned[0].path + "\t" + owned[0].rule})
	setupStarted := time.Now()
	oracle := compilerAgreementGoOracle(t)
	binary := compilerAgreementNative(t, true)
	t.Logf("preparation: %.3fs", time.Since(setupStarted).Seconds())
	// Own work is logged, not asserted, as in compilerAgreementShard: execute's testguard budget and ceiling bound a hang.
	workStarted := time.Now()
	defer func() { t.Logf("own work: %.3fs", time.Since(workStarted).Seconds()) }()
	want := execute(t, "", oracle, "--manifest", path)
	got := execute(t, "", binary, "--manifest", path)
	compilerAgreementCompare(t, got.output, want.output)
	t.Logf("side 4: %s, oracle %s, rule %s", got.duration, want.duration, name)
}

func TestProduct_CompilerAgreementLowered(t *testing.T) {
	t.Parallel()
	compilerAgreementMeasureProduct(t, compilerAgreementLowered)
}

func TestCompilerAndStage1Agree_6484(t *testing.T) {
	t.Parallel()
	t.Skip("awaits devtools/heavy-units: pending #5ggwf8c; combined sanitized checker.ts/all run moved to that heavy-unit task, not dropped; #m9es0rv adds the heavy census class and budget exemption")
}

func compilerAgreementCheckUnits(t *testing.T, cases []compilerAgreementCase) {
	t.Helper()
	var source []byte
	for _, path := range []string{"compiler_stage1_split_test.go", "compiler_stage1_sanitized_groups_test.go"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source = append(append(source, '\n'), data...)
	}
	units := map[int]bool{}
	old := regexp.MustCompile(`(?m)^func TestCompilerAndStage1Agree_([0-9]+)\(t \*testing.T\)`).FindAllSubmatch(source, -1)
	if len(old) != testCompilerAndStage1AgreeShards {
		t.Fatalf("enumerated %d original shards, want %d", len(old), testCompilerAndStage1AgreeShards)
	}
	for _, m := range old {
		index, err := strconv.Atoi(string(m[1]))
		if err != nil || index < 0 || index >= testCompilerAndStage1AgreeShards || units[index] {
			t.Fatalf("invalid original shard %s", m[1])
		}
		units[index] = true
	}
	descriptors := prepareRegistry(t, ".")
	owners, err := compilerAgreementCheckerSanitizedOwners(source, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	// The per-rule check must refuse a missing wrapper, one that runs another rule, and one naming no registered rule.
	first := compilerAgreementCheckerSanitizedWrapper.FindSubmatch(source)
	wrapper := func(test, slug string) []byte {
		return []byte(fmt.Sprintf("\nfunc %s%s(t *testing.T) {\n\tt.Parallel()\n\tcompilerAgreementCheckerSanitized(t, %q)\n}\n", compilerAgreementCheckerSanitizedPrefix, test, slug))
	}
	for planted, broken := range map[string][]byte{
		"missing":      bytes.Replace(source, first[0], nil, 1),
		"wrong rule":   append(bytes.Replace(source, first[0], nil, 1), wrapper(string(first[1]), compilerAgreementSanitizedPlantedRule)...),
		"unregistered": append(append([]byte(nil), source...), wrapper("planted_rule", "planted-rule")...),
	} {
		if _, err := compilerAgreementCheckerSanitizedOwners(broken, descriptors); err == nil {
			t.Fatalf("a %s sanitized checker wrapper survived the union check", planted)
		}
	}
	seen := map[string]bool{}
	owned := map[string]int{}
	for _, c := range cases {
		for side := 0; side < compilerAgreementSides; side++ {
			unit := compilerAgreementUnit(c, side)
			pair := fmt.Sprintf("%s\t%s\t%d", c.key, c.rule, side)
			if seen[pair] {
				t.Fatalf("duplicate coverage for %s", pair)
			}
			seen[pair] = true
			sanitizedChecker := c.key == compilerAgreementCheckerKey && c.rule != "all" && side == 4
			if unit == compilerAgreementRuleOwned {
				if !sanitizedChecker || owners[c.rule] == "" {
					t.Fatalf("%s/%s side %d left the numbered shards for no rule test", c.key, c.rule, side)
				}
				owned[owners[c.rule]]++
				continue
			}
			if !units[unit] {
				t.Fatalf("missing coverage for %s/%s side %d: unit %d absent", c.key, c.rule, side, unit)
			}
			if sanitizedChecker {
				t.Fatalf("sanitized checker rule %s is in numbered shard %d, not its own test", c.rule, unit)
			}
			if c.key == compilerAgreementCheckerKey && c.rule != "all" && unit != compilerAgreementOwner(c)*compilerAgreementSides+side {
				t.Fatalf("checker rule %s moved on side %d", c.rule, side)
			}
			if c.key == compilerAgreementCheckerKey && c.rule == "all" && unit != 6480+side {
				t.Fatalf("checker/all moved on side %d", side)
			}
		}
	}
	for _, test := range owners {
		if owned[test] != 1 {
			t.Fatalf("%s owns %d sanitized checker.ts cases, want 1", test, owned[test])
		}
	}
	t.Logf("union: %d checker rules covered exactly once under ASan, each by its own %s<slug>; checker/all stays on sides 0–3; side 4 is pending #5ggwf8c", len(owned), compilerAgreementCheckerSanitizedPrefix)
}

// The probe child runs the whole family with a disagreement planted in one rule's checker.ts case: only the test that
// owns that case may fail, and every other rule's test must pass.
func TestCompilerAndStage1AgreeSanitizedPlantedDisagreement(t *testing.T) {
	t.Parallel()
	_ = compilerAgreementCases(t)
	source, err := os.ReadFile("compiler_stage1_sanitized_groups_test.go")
	if err != nil {
		t.Fatal(err)
	}
	descriptors := prepareRegistry(t, ".")
	owners, err := compilerAgreementCheckerSanitizedOwners(source, descriptors)
	if err != nil {
		t.Fatal(err)
	}
	owner := owners[compilerAgreementSanitizedPlantedRule]
	command, cancel := compilerAgreementCommand(os.Args[0], "-test.run=^"+compilerAgreementCheckerSanitizedPrefix, "-test.skip=^TestProduct_", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_COMPILER_AGREEMENT_PROBE=1")
	output, err := command.CombinedOutput()
	cancel()
	if err == nil || bytes.Count(output, []byte("--- FAIL:")) != 1 || !bytes.Contains(output, []byte("--- FAIL: "+owner+" ")) {
		t.Fatalf("planted sanitized disagreement not caught exactly once, by %s: %v\n%s", owner, err, output)
	}
	if passed := bytes.Count(output, []byte("--- PASS: "+compilerAgreementCheckerSanitizedPrefix)); passed != len(descriptors)-1 {
		t.Fatalf("%d sanitized checker tests passed beside the owner, want %d\n%s", passed, len(descriptors)-1, output)
	}
	t.Logf("planted sanitized fixed-output disagreement in %s's checker.ts case caught only by %s", compilerAgreementSanitizedPlantedRule, owner)
}

func TestCompilerAndStage1Agree_CheckerSanitized_adamic_no_type_predicate(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "adamic-no-type-predicate")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_boundary_no_global_container(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "base-boundary-no-global-container")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_no_console(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "base-consistency-no-console")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_no_hand_built_declared_error(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "base-consistency-no-hand-built-declared-error")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_require_pagination_argument_name(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "base-consistency-require-pagination-argument-name")
}

func TestCompilerAndStage1Agree_CheckerSanitized_eqeqeq(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "eqeqeq")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_classes_per_file(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "max-classes-per-file")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_depth(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "max-depth")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_nested_callbacks(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "max-nested-callbacks")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_abbreviated_identifier(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-abbreviated-identifier")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_ambiguous_identifier(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-ambiguous-identifier")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_enum(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-enum")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_for_in(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-for-in")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_long_line_comment(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-long-line-comment")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_multiline_arrow_function(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-multiline-arrow-function")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_shouting(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-shouting")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_single_line_jsdoc(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-single-line-jsdoc")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_stuttering_name(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-no-stuttering-name")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_require_type_suffix(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "nexus-consistency-require-type-suffix")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_await_in_loop(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-await-in-loop")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_bitwise(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-bitwise")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_caller(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-caller")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_continue(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-continue")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_debugger(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-debugger")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_div_regex(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-div-regex")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_duplicate_case(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-duplicate-case")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-empty")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_character_class(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-empty-character-class")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_pattern(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-empty-pattern")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_static_block(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-empty-static-block")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_ex_assign(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-ex-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_iterator(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-iterator")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_labels(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-labels")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_multi_assign(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-multi-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_negated_condition(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-negated-condition")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_new(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-new")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_octal_escape(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-octal-escape")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_plusplus(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-plusplus")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_proto(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-proto")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_return_assign(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-return-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_script_url(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-script-url")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_self_compare(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-self-compare")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_sequences(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-sequences")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_sparse_arrays(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-sparse-arrays")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_template_curly_in_string(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-template-curly-in-string")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_underscore_dangle(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-underscore-dangle")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unexpected_multiline(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unexpected-multiline")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unneeded_ternary(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unneeded-ternary")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_finally(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unsafe-finally")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_negation(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unsafe-negation")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_optional_chaining(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unsafe-optional-chaining")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unused_private_class_members(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-unused-private-class-members")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_catch(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-useless-catch")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_concat(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-useless-concat")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_constructor(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-useless-constructor")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_var(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-var")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_warning_comments(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-warning-comments")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_with(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "no-with")
}

func TestCompilerAndStage1Agree_CheckerSanitized_prefer_destructuring(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "prefer-destructuring")
}

func TestCompilerAndStage1Agree_CheckerSanitized_prefer_template(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "prefer-template")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_forward_ref_uses_ref(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "react-forward-ref-uses-ref")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_jsx_no_comment_textnodes(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "react-jsx-no-comment-textnodes")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_find_dom_node(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "react-no-find-dom-node")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_is_mounted(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "react-no-is-mounted")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_redundant_should_component_update(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "react-no-redundant-should-component-update")
}

func TestCompilerAndStage1Agree_CheckerSanitized_require_yield(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "require-yield")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_adjacent_overload_signatures(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-adjacent-overload-signatures")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_init_declarations(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-init-declarations")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_method_signature_style(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-method-signature-style")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_no_this_alias(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-no-this-alias")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_no_wrapper_object_types(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-no-wrapper-object-types")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_prefer_literal_enum_member(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "typescript-prefer-literal-enum-member")
}

func TestCompilerAndStage1Agree_CheckerSanitized_unicode_bom(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "unicode-bom")
}

func TestCompilerAndStage1Agree_CheckerSanitized_vars_on_top(t *testing.T) {
	t.Parallel()
	compilerAgreementCheckerSanitized(t, "vars-on-top")
}
