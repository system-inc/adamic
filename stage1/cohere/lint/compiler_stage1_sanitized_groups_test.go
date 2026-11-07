package lint

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// One rule over checker.ts under ASan is the indivisible sanitized case, so each rule's case is its own test, named by
// the rule's slug: no numbered shard selects it (its unit is compilerAgreementRuleOwned), and no hash group or list
// position decides which test runs it.
const compilerAgreementRuleOwned = -1
const compilerAgreementCheckerSanitizedPrefix = "TestCompilerAndStage1Agree_CheckerSanitized_"
const compilerAgreementSanitizedPlantedRule = "no-debugger"

var compilerAgreementCheckerSanitizedDeclaration = regexp.MustCompile(
	`(?m)^func ` + compilerAgreementCheckerSanitizedPrefix,
)
var compilerAgreementCheckerSanitizedWrapper = regexp.MustCompile(`(?m)^func ` +
	compilerAgreementCheckerSanitizedPrefix +
	`(\w+)\(t \*testing.T\) \{\s*t.Parallel\(\)\s*;?\s*compilerAgreementCheckerSanitized\(t, "([a-z0-9-]+)"\)\s*\}`)

// compilerAgreementSplit names the owner sides heavy enough to split, each into Parts sub-shards
// (compilerAgreementPart). The first six are side 4 under ASan, their cases dealt by a hash of the case's key and rule,
// so a case keeps its sub-shard as the corpus grows: their numbered side-4 shards ran 15.2 to 26.5 s at an 8-CPU share
// over 72 to 110 cases, and every other side-4 shard at most 14.6 (#3tf18zm, 45 test logs over lint's 20 units, Oct
// 14). 1215 is a fix bucket, _6079 at 19.1 s, whose program.ts alone is 14.1 s on one core: two parts put it beside
// transformers/destructuring.ts and a small file, about 12 s, and its fourth file apart, and no count up to its four
// files gives program.ts a part to itself without an empty one. The fix buckets whose one file alone is over 15 s
// (types.ts, utilities.ts, parser.ts) stay whole. 1303 and 1310 are checker.ts's per-rule groups 23 and 30 on side 1,
// the lowered JavaScript under Node, one part for each rule: _6516 ran 30.2 s over six rules and _6551 23.4 s over
// three (#mq768wp, Home's units 35 and 36 of 4a769ad6's run, Oct 19), and on a loaded Golem a part of one rule took 5.7
// and 13.7 s, so the hash's best cut, three and two parts, left two of them at 23 and 24 s. Each part is
// TestCompilerAndStage1Agree_<shard>_<part>, and the owner's numbered wrapper for that side is gone. Checker/all (owner
// 1312) uses consecutive measured-price groups instead (compilerAgreementCheckerSideGroups), and its one original row
// expands on every side: sides 1 and 4 into its groups, and sides 2 and 3 into compilerAgreementCheckerSideParts parts
// of consecutive groups, as the whole file under every rule ran 64.7, 56.2 and 53.1 s on sides 1 to 3, _6561 to _6563,
// over Loom's 60 s budget at an 8-CPU share (#mq768wp, unit 72 of future-9f780ec59f99-1, Oct 22). So no side runs
// every rule over checker.ts in one process any more: an interaction between rules in different parts goes unseen
// there, as it already did on side 4. Every other file still runs all its rules together, in one case, on every side.
// 1281, 1284, 1285, 1294, 1301 and 1306 are checker.ts's per-rule groups 1, 4, 5, 14, 21 and 26 on side 1, split one
// rule to a part as 1303 and 1310 are: _6506 ran 61.2 and 55.0 s over five rules, _6406 52.2 and 47.2 over six, _6426
// 42.2 over six, _6421 37.0 over five, _6471 36.2 over five and _6531 35.7 over three, against Loom's 60 s budget at an
// 8-CPU share (#mq768wp, Eve's units of future-f648d7372a15-1 and future-29df76653c00-1, Oct 22). A future heavy owner
// side joins by number.
var compilerAgreementSplit = []struct{ Owner, Side, Parts int }{
	{289, 4, 4},
	{293, 4, 4},
	{385, 4, 4},
	{673, 4, 4},
	{1009, 4, 4},
	{1215, 4, 2},
	{compilerAgreementCheckerAllOwner, 1, compilerAgreementCheckerGroupCount},
	{compilerAgreementCheckerAllOwner, 2, compilerAgreementCheckerSideParts},
	{compilerAgreementCheckerAllOwner, 3, compilerAgreementCheckerSideParts},
	{compilerAgreementCheckerAllOwner, 4, compilerAgreementCheckerGroupCount},
	{1281, 1, 6},
	{1284, 1, 5},
	{1285, 1, 6},
	{1294, 1, 5},
	{1301, 1, 5},
	{1303, 1, 6},
	{1306, 1, 3},
	{1310, 1, 3},
}

// compilerAgreementCheckerSideParts is how many parts checker/all's sides 2 and 3 run in, two of the 28 groups to a
// part. Each part pays its own parse of checker.ts and its own live Go answer, so a side takes no more parts than it
// needs. Run all at once at GOMAXPROCS=8 on a quiet Pupa (3e5fde87, Oct 22), sides 2 and 3 at fourteen parts took 2.7
// to 12.7 s, and side 1 at fourteen 5.9 to 26.5 s, its part of groups 12 and 13 the 26.5, so side 1 runs every group
// as a part, as side 4 does: its 28 parts, run alone, took 5.2 to 15.2 s.
const compilerAgreementCheckerSideParts = 14

// compilerAgreementMaxParts bounds a split owner's parts, so every sub-unit has its own number past the shards'.
const compilerAgreementMaxParts = 32

// compilerAgreementSplitParts is how many sub-shards owner's side is split into, or 0 when it isn't.
func compilerAgreementSplitParts(owner, side int) int {
	for _, split := range compilerAgreementSplit {
		if split.Owner == owner && split.Side == side {
			return split.Parts
		}
	}
	return 0
}

// compilerAgreementPart deals a split owner's case to one of parts, most by a sha256 of its key and rule. Not by fnv,
// as the owner is: the owner's cases share their key's and rule's fnv buckets, and fnv-1a's low bits follow from its
// input's low bits, so fnv's modulo put every case of four of the five owners in one part (Cloud, Oct 14).
func compilerAgreementPart(agreementCase compilerAgreementCase, parts int) int {
	// Combined checker groups already carry their production-order part.
	if agreementCase.group > 0 {
		return agreementCase.group - 1
	}
	// A checker.ts per-rule case is one lint of the whole file, the indivisible case on Node as under ASan, so each of
	// its owner's rules is a part of its own. A rule added to the group is a part past Parts, which the union refuses.
	if agreementCase.ordinal > 0 {
		return agreementCase.ordinal - 1
	}
	sum := sha256.Sum256([]byte(agreementCase.key + "\t" + agreementCase.rule))
	return int(binary.BigEndian.Uint64(sum[:8]) % uint64(parts))
}

// compilerAgreementSubUnit numbers a split shard's sub-shard past every numbered shard.
func compilerAgreementSubUnit(shard, part int) int {
	return testCompilerAndStage1AgreeShards + shard*compilerAgreementMaxParts + part
}

func compilerAgreementUnit(agreementCase compilerAgreementCase, side int) int {
	if agreementCase.key == compilerAgreementCheckerKey &&
		agreementCase.rule != "all" &&
		agreementCase.group == 0 &&
		side == 4 {
		return compilerAgreementRuleOwned
	}
	owner := compilerAgreementOwner(agreementCase)
	if parts := compilerAgreementSplitParts(owner, side); parts > 0 {
		return compilerAgreementSubUnit(owner*compilerAgreementSides+side, compilerAgreementPart(agreementCase, parts))
	}
	return owner*compilerAgreementSides + side
}

// compilerAgreementSubShardWrapper is a sub-shard's test as the file spells it: the shard and part it carries, and the
// shard and part it hands compilerAgreementSubShard, which must say the same.
var compilerAgreementSubShardWrapper = regexp.MustCompile(
	`(?m)^func TestCompilerAndStage1Agree_([0-9]+)_([0-9]+)\(t \*testing\.T\) \{\n\tt\.Parallel\(\)\n\tcompilerAgreementSubShard\(t, ([0-9]+), ([0-9]+)\)\n\}$`,
)
var compilerAgreementSubShardDeclaration = regexp.MustCompile(`(?m)^func TestCompilerAndStage1Agree_[0-9]+_[0-9]+\(`)

// compilerAgreementSubShardUnits reads the sub-shard wrappers from source: every split owner side has exactly one
// wrapper per part, each runs the shard and part its name says, and none names an owner side the list doesn't split.
func compilerAgreementSubShardUnits(source []byte) (map[int]bool, error) {
	wrappers := compilerAgreementSubShardWrapper.FindAllSubmatch(source, -1)
	if declared := len(compilerAgreementSubShardDeclaration.FindAll(source, -1)); len(wrappers) != declared {
		return nil, fmt.Errorf("%d of %d sub-shard tests have the wrapper shape", len(wrappers), declared)
	}
	units := map[int]bool{}
	for _, wrapper := range wrappers {
		shard, _ := strconv.Atoi(string(wrapper[1]))
		part, _ := strconv.Atoi(string(wrapper[2]))
		if string(wrapper[1]) != string(wrapper[3]) || string(wrapper[2]) != string(wrapper[4]) {
			return nil, fmt.Errorf(
				"TestCompilerAndStage1Agree_%s_%s runs shard %s part %s",
				wrapper[1],
				wrapper[2],
				wrapper[3],
				wrapper[4],
			)
		}
		if part >= compilerAgreementSplitParts(shard/compilerAgreementSides, shard%compilerAgreementSides) {
			return nil, fmt.Errorf("TestCompilerAndStage1Agree_%d_%d is no split owner side's part", shard, part)
		}
		unit := compilerAgreementSubUnit(shard, part)
		if units[unit] {
			return nil, fmt.Errorf("TestCompilerAndStage1Agree_%d_%d is declared twice", shard, part)
		}
		units[unit] = true
	}
	splits := map[int]bool{}
	for _, split := range compilerAgreementSplit {
		shard := split.Owner*compilerAgreementSides + split.Side
		if split.Side < compilerAgreementFirstSide || split.Side >= compilerAgreementSides || splits[shard] {
			return nil, fmt.Errorf("owner %d side %d is no side, or is split twice", split.Owner, split.Side)
		}
		splits[shard] = true
		if split.Parts < 2 || split.Parts > compilerAgreementMaxParts {
			return nil, fmt.Errorf(
				"owner %d side %d is split into %d parts, want 2 to %d",
				split.Owner,
				split.Side,
				split.Parts,
				compilerAgreementMaxParts,
			)
		}
		for part := range split.Parts {
			if !units[compilerAgreementSubUnit(shard, part)] {
				return nil, fmt.Errorf(
					"owner %d's side %d has no TestCompilerAndStage1Agree_%d_%d: its cases would never run",
					split.Owner,
					split.Side,
					shard,
					part,
				)
			}
		}
	}
	return units, nil
}

// compilerAgreementSubShardPlanted is the case the sub-shards' probe plants a disagreement in for a split owner: its
// first case, in corpus order.
func compilerAgreementSubShardPlanted(test *testing.T, owner int) compilerAgreementCase {
	test.Helper()
	for _, agreementCase := range compilerAgreementCases(test) {
		if compilerAgreementOwner(agreementCase) == owner {
			return agreementCase
		}
	}
	test.Fatalf("split owner %d holds no case", owner)
	return compilerAgreementCase{}
}

// compilerAgreementCheckerSanitizedOwners reads the per-rule wrappers from source and maps each rule's name to the test
// that owns its sanitized checker.ts case. Every registered rule has exactly one wrapper, named by its slug, and every
// declared test of the family has the wrapper's shape, so none runs outside the union.
func compilerAgreementCheckerSanitizedOwners(
	source []byte,
	descriptors []registry.Descriptor,
) (map[string]string, error) {
	wrappers := compilerAgreementCheckerSanitizedWrapper.FindAllSubmatch(source, -1)
	if declared := len(compilerAgreementCheckerSanitizedDeclaration.FindAll(source, -1)); len(wrappers) != declared {
		return nil, fmt.Errorf("%d of %d sanitized checker tests have the wrapper shape", len(wrappers), declared)
	}
	names := map[string]string{}
	for _, descriptor := range descriptors {
		names[descriptor.Slug] = descriptor.Name
	}
	owners := map[string]string{}
	for _, wrapper := range wrappers {
		test, slug := compilerAgreementCheckerSanitizedPrefix+string(wrapper[1]), string(wrapper[2])
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
	for _, descriptor := range descriptors {
		if owners[descriptor.Name] == "" {
			return nil, fmt.Errorf("rule %s has no sanitized checker test", descriptor.Slug)
		}
	}
	return owners, nil
}

// compilerAgreementCheckerSanitized runs one rule over checker.ts on the sanitized native binary and compares its bytes
// with the Go oracle's answer for the same row.
func compilerAgreementCheckerSanitized(test *testing.T, slug string) {
	test.Helper()
	name := ""
	for _, descriptor := range prepareRegistry(test, ".") {
		if descriptor.Slug == slug {
			name = descriptor.Name
		}
	}
	if name == "" {
		test.Fatalf("sanitized checker test names unregistered rule %s", slug)
	}
	var owned []compilerAgreementCase
	for _, agreementCase := range compilerAgreementCases(test) {
		if compilerAgreementUnit(agreementCase, 4) == compilerAgreementRuleOwned && agreementCase.rule == name {
			owned = append(owned, agreementCase)
		}
	}
	if len(owned) != 1 {
		test.Fatalf("rule %s owns %d sanitized checker.ts cases, want 1", name, len(owned))
	}
	if os.Getenv("ADAMIC_COMPILER_AGREEMENT_PROBE") == "1" {
		want, got := []byte("case 0\nfixed\tunchanged\n"), []byte("case 0\nfixed\tunchanged\n")
		if owned[0].key == compilerAgreementCheckerKey && owned[0].rule == compilerAgreementSanitizedPlantedRule {
			got = []byte("case 0\nfixed\tdisagreement\n")
		}
		compilerAgreementCompare(test, got, want)
		return
	}
	path := manifest(test, []string{owned[0].path + "\t" + owned[0].rule})
	setupStarted := time.Now()
	oracle := lintGoOracleProduct(test)
	binary := compilerAgreementNative(test, true)
	test.Logf("preparation: %.3fs", time.Since(setupStarted).Seconds())
	// Own work is logged, not asserted, as in compilerAgreementShard: execute's testguard budget and ceiling bound a
	// hang.
	workStarted := time.Now()
	defer func() { test.Logf("own work: %.3fs", time.Since(workStarted).Seconds()) }()
	want := execute(test, "", oracle, "--manifest", path)
	got := execute(test, "", binary, "--manifest", path)
	compilerAgreementCompare(test, got.output, want.output)
	test.Logf("side 4: %s, oracle %s, rule %s", got.duration, want.duration, name)
}

func TestProduct_CompilerAgreementLowered(test *testing.T) {
	test.Parallel()
	compilerAgreementMeasureProduct(test, func(test *testing.T) string {
		directory := compilerAgreementLowered(test)
		requireLintArtifacts(test, directory, "lint.c", "lint.mjs")
		return directory
	})
}

func compilerAgreementCheckUnits(test *testing.T, cases []compilerAgreementCase) {
	test.Helper()
	var source []byte
	for _, path := range []string{"compiler_stage1_split_test.go", "compiler_stage1_sanitized_groups_test.go"} {
		data, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		source = append(append(source, '\n'), data...)
	}
	units := map[int]bool{}
	old := regexp.MustCompile(`(?m)^func TestCompilerAndStage1Agree_([0-9]+)\(t \*testing.T\)`).FindAllSubmatch(
		source,
		-1,
	)
	// Every owner's sides from compilerAgreementFirstSide on have a wrapper; side 0 has none, and a split owner side
	// has its sub-shards' instead.
	want := testCompilerAndStage1AgreeShards/compilerAgreementSides*(compilerAgreementSides-compilerAgreementFirstSide) - len(
		compilerAgreementSplit,
	)
	if len(old) != want {
		test.Fatalf("enumerated %d original shards, want %d", len(old), want)
	}
	for _, m := range old {
		index, parseError := strconv.Atoi(string(m[1]))
		if parseError != nil ||
			index < 0 ||
			index >= testCompilerAndStage1AgreeShards ||
			units[index] ||
			index%compilerAgreementSides < compilerAgreementFirstSide {
			test.Fatalf("invalid original shard %s", m[1])
		}
		units[index] = true
	}
	for _, split := range compilerAgreementSplit {
		if units[split.Owner*compilerAgreementSides+split.Side] {
			test.Fatalf(
				"split owner %d keeps its numbered side-%d wrapper beside its sub-shards",
				split.Owner,
				split.Side,
			)
		}
	}
	subUnits, compilerAgreementSubShardUnitsError := compilerAgreementSubShardUnits(source)
	if compilerAgreementSubShardUnitsError != nil {
		test.Fatal(compilerAgreementSubShardUnitsError)
	}
	for unit := range subUnits {
		units[unit] = true
	}
	// The sub-shard check must refuse a missing part, a wrapper that runs another part, and a part beyond its owner's.
	firstSub := compilerAgreementSubShardWrapper.FindSubmatch(source)
	subWrapper := func(shard, part, runsShard, runsPart string) []byte {
		return []byte(fmt.Sprintf(
			"\nfunc TestCompilerAndStage1Agree_%s_%s(t *testing.T) {\n\tt.Parallel()\n\tcompilerAgreementSubShard(t, %s, %s)\n}\n",
			shard,
			part,
			runsShard,
			runsPart,
		))
	}
	for planted, broken := range map[string][]byte{
		"missing": bytes.Replace(source, firstSub[0], nil, 1),
		"wrong part": append(
			bytes.Replace(source, firstSub[0], nil, 1),
			subWrapper(string(firstSub[1]), string(firstSub[2]), string(firstSub[1]), "9")...,
		),
		"extra part": append(
			append([]byte(nil), source...),
			subWrapper(string(firstSub[1]), "9", string(firstSub[1]), "9")...,
		),
	} {
		if _, compilerAgreementSubShardUnitsError := compilerAgreementSubShardUnits(
			broken,
		); compilerAgreementSubShardUnitsError == nil {
			test.Fatalf("a %s sub-shard wrapper survived the union check", planted)
		}
	}
	descriptors := prepareRegistry(test, ".")
	owners, compilerAgreementCheckerSanitizedOwnersError := compilerAgreementCheckerSanitizedOwners(source, descriptors)
	if compilerAgreementCheckerSanitizedOwnersError != nil {
		test.Fatal(compilerAgreementCheckerSanitizedOwnersError)
	}
	// The per-rule check must refuse a missing wrapper, one that runs another rule, and one naming no registered rule.
	first := compilerAgreementCheckerSanitizedWrapper.FindSubmatch(source)
	wrapper := func(test, slug string) []byte {
		return []byte(fmt.Sprintf(
			"\nfunc %s%s(t *testing.T) {\n\tt.Parallel()\n\tcompilerAgreementCheckerSanitized(t, %q)\n}\n",
			compilerAgreementCheckerSanitizedPrefix,
			test,
			slug,
		))
	}
	for planted, broken := range map[string][]byte{
		"missing": bytes.Replace(source, first[0], nil, 1),
		"wrong rule": append(
			bytes.Replace(source, first[0], nil, 1),
			wrapper(string(first[1]), compilerAgreementSanitizedPlantedRule)...,
		),
		"unregistered": append(append([]byte(nil), source...), wrapper("planted_rule", "planted-rule")...),
	} {
		if _, compilerAgreementCheckerSanitizedOwnersError := compilerAgreementCheckerSanitizedOwners(
			broken,
			descriptors,
		); compilerAgreementCheckerSanitizedOwnersError == nil {
			test.Fatalf("a %s sanitized checker wrapper survived the union check", planted)
		}
	}
	seen := map[string]bool{}
	owned := map[string]int{}
	subCases := map[int]int{}
	for side := compilerAgreementFirstSide; side < compilerAgreementSides; side++ {
		for _, agreementCase := range compilerAgreementSideCases(test, cases, side) {
			unit := compilerAgreementUnit(agreementCase, side)
			if agreementCase.group > 0 && (agreementCase.key != compilerAgreementCheckerKey ||
				agreementCase.group > compilerAgreementSplitParts(compilerAgreementCheckerAllOwner, side) ||
				unit != compilerAgreementSubUnit(
					compilerAgreementCheckerAllOwner*compilerAgreementSides+side,
					agreementCase.group-1,
				)) {
				test.Fatalf("invalid combined checker group %d on side %d: unit %d", agreementCase.group, side, unit)
			}
			pair := fmt.Sprintf("%s\t%s\t%d", agreementCase.key, agreementCase.rule, side)
			if seen[pair] {
				test.Fatalf("duplicate coverage for %s", pair)
			}
			seen[pair] = true
			sanitizedChecker := agreementCase.key == compilerAgreementCheckerKey &&
				agreementCase.rule != "all" &&
				agreementCase.group == 0 &&
				side == 4
			if unit == compilerAgreementRuleOwned {
				if !sanitizedChecker || owners[agreementCase.rule] == "" {
					test.Fatalf(
						"%s/%s side %d left the numbered shards for no rule test",
						agreementCase.key,
						agreementCase.rule,
						side,
					)
				}
				owned[owners[agreementCase.rule]]++
				continue
			}
			if !units[unit] {
				test.Fatalf(
					"missing coverage for %s/%s side %d: unit %d absent",
					agreementCase.key,
					agreementCase.rule,
					side,
					unit,
				)
			}
			if subUnits[unit] {
				subCases[unit]++
			}
			if sanitizedChecker {
				test.Fatalf(
					"sanitized checker rule %s is in numbered shard %d, not its own test",
					agreementCase.rule,
					unit,
				)
			}
			// A checker rule stays on its owner's side, in the numbered shard or one of its sub-shards.
			if shard := compilerAgreementOwner(agreementCase)*compilerAgreementSides +
				side; agreementCase.key == compilerAgreementCheckerKey &&
				agreementCase.rule != "all" &&
				agreementCase.group == 0 &&
				unit != shard &&
				(!subUnits[unit] || (unit-testCompilerAndStage1AgreeShards)/compilerAgreementMaxParts != shard) {
				test.Fatalf("checker rule %s moved on side %d", agreementCase.rule, side)
			}
			if agreementCase.key == compilerAgreementCheckerKey && agreementCase.rule == "all" {
				test.Fatalf("checker/all runs whole on side %d, not in its parts", side)
			}
		}
	}
	// A part that holds no case is a test that runs nothing, the sign of a hash that doesn't spread its owner's cases.
	for unit := range subUnits {
		if subCases[unit] == 0 {
			test.Fatalf(
				"TestCompilerAndStage1Agree_%d_%d holds no case",
				(unit-testCompilerAndStage1AgreeShards)/compilerAgreementMaxParts,
				(unit-testCompilerAndStage1AgreeShards)%compilerAgreementMaxParts,
			)
		}
	}
	for _, testCase := range owners {
		if owned[testCase] != 1 {
			test.Fatalf("%s owns %d sanitized checker.ts cases, want 1", testCase, owned[testCase])
		}
	}
	test.Logf(
		"union: %d checker rules covered exactly once under ASan, each by its own %s<slug>; checker/all runs in consecutive groups on every side",
		len(owned),
		compilerAgreementCheckerSanitizedPrefix,
	)
	test.Logf(
		"union: %d split owner sides' cases each in exactly one of their %d sub-shards",
		len(compilerAgreementSplit),
		len(subUnits),
	)
}

// The probe child runs the whole family with a disagreement planted in one rule's checker.ts case: only the test that
// owns that case may fail, and every other rule's test must pass.
func TestCompilerAndStage1AgreeSanitizedPlantedDisagreement(test *testing.T) {
	test.Parallel()
	_ = compilerAgreementCases(test)
	source, readError := os.ReadFile("compiler_stage1_sanitized_groups_test.go")
	if readError != nil {
		test.Fatal(readError)
	}
	descriptors := prepareRegistry(test, ".")
	owners, compilerAgreementCheckerSanitizedOwnersError := compilerAgreementCheckerSanitizedOwners(source, descriptors)
	if compilerAgreementCheckerSanitizedOwnersError != nil {
		test.Fatal(compilerAgreementCheckerSanitizedOwnersError)
	}
	owner := owners[compilerAgreementSanitizedPlantedRule]
	command, ctx, cancel := compilerAgreementCommand(
		os.Args[0],
		"-test.run=^"+compilerAgreementCheckerSanitizedPrefix,
		"-test.skip=^TestProduct_",
		"-test.timeout=90s",
		"-test.v",
	)
	command.Env = append(os.Environ(), "ADAMIC_COMPILER_AGREEMENT_PROBE=1")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	cancel()
	if captureError == nil ||
		bytes.Count(output, []byte("--- FAIL:")) != 1 ||
		!bytes.Contains(output, []byte("--- FAIL: "+owner+" ")) {
		test.Fatalf(
			"planted sanitized disagreement not caught exactly once, by %s: %v\n%s",
			owner,
			captureError,
			output,
		)
	}
	if passed := bytes.Count(
		output,
		[]byte("--- PASS: "+compilerAgreementCheckerSanitizedPrefix),
	); passed != len(descriptors)-1 {
		test.Fatalf(
			"%d sanitized checker tests passed beside the owner, want %d\n%s",
			passed,
			len(descriptors)-1,
			output,
		)
	}
	test.Logf(
		"planted sanitized fixed-output disagreement in %s's checker.ts case caught only by %s",
		compilerAgreementSanitizedPlantedRule,
		owner,
	)
}

func TestCompilerAndStage1Agree_CheckerSanitized_adamic_no_type_predicate(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "adamic-no-type-predicate")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_boundary_no_global_container(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "base-boundary-no-global-container")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_no_console(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "base-consistency-no-console")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_no_hand_built_declared_error(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "base-consistency-no-hand-built-declared-error")
}

func TestCompilerAndStage1Agree_CheckerSanitized_base_consistency_require_pagination_argument_name(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "base-consistency-require-pagination-argument-name")
}

func TestCompilerAndStage1Agree_CheckerSanitized_eqeqeq(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "eqeqeq")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_classes_per_file(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "max-classes-per-file")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_depth(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "max-depth")
}

func TestCompilerAndStage1Agree_CheckerSanitized_max_nested_callbacks(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "max-nested-callbacks")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_abbreviated_identifier(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-abbreviated-identifier")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_ambiguous_identifier(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-ambiguous-identifier")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_enum(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-enum")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_for_in(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-for-in")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_long_line_comment(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-long-line-comment")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_multiline_arrow_function(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-multiline-arrow-function")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_shouting(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-shouting")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_single_line_jsdoc(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-single-line-jsdoc")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_stuttering_name(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-no-stuttering-name")
}

func TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_require_type_suffix(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "nexus-consistency-require-type-suffix")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_await_in_loop(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-await-in-loop")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_bitwise(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-bitwise")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_caller(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-caller")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_continue(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-continue")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_debugger(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-debugger")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_div_regex(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-div-regex")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_duplicate_case(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-duplicate-case")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-empty")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_character_class(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-empty-character-class")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_pattern(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-empty-pattern")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_empty_static_block(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-empty-static-block")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_ex_assign(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-ex-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_iterator(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-iterator")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_labels(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-labels")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_multi_assign(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-multi-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_negated_condition(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-negated-condition")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_new(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-new")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_octal_escape(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-octal-escape")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_plusplus(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-plusplus")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_proto(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-proto")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_return_assign(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-return-assign")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_script_url(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-script-url")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_self_compare(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-self-compare")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_sequences(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-sequences")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_sparse_arrays(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-sparse-arrays")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_template_curly_in_string(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-template-curly-in-string")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_underscore_dangle(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-underscore-dangle")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unexpected_multiline(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unexpected-multiline")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unneeded_ternary(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unneeded-ternary")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_finally(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unsafe-finally")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_negation(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unsafe-negation")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unsafe_optional_chaining(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unsafe-optional-chaining")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unused_private_class_members(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unused-private-class-members")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_catch(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-useless-catch")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_concat(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-useless-concat")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_constructor(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-useless-constructor")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_var(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-var")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_warning_comments(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-warning-comments")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_with(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-with")
}

func TestCompilerAndStage1Agree_CheckerSanitized_prefer_destructuring(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "prefer-destructuring")
}

func TestCompilerAndStage1Agree_CheckerSanitized_prefer_template(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "prefer-template")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_forward_ref_uses_ref(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "react-forward-ref-uses-ref")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_jsx_no_comment_textnodes(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "react-jsx-no-comment-textnodes")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_find_dom_node(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "react-no-find-dom-node")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_is_mounted(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "react-no-is-mounted")
}

func TestCompilerAndStage1Agree_CheckerSanitized_react_no_redundant_should_component_update(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "react-no-redundant-should-component-update")
}

func TestCompilerAndStage1Agree_CheckerSanitized_require_yield(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "require-yield")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_adjacent_overload_signatures(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-adjacent-overload-signatures")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_init_declarations(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-init-declarations")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_method_signature_style(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-method-signature-style")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_no_this_alias(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-no-this-alias")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_no_wrapper_object_types(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-no-wrapper-object-types")
}

func TestCompilerAndStage1Agree_CheckerSanitized_typescript_prefer_literal_enum_member(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "typescript-prefer-literal-enum-member")
}

func TestCompilerAndStage1Agree_CheckerSanitized_unicode_bom(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "unicode-bom")
}

func TestCompilerAndStage1Agree_CheckerSanitized_vars_on_top(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "vars-on-top")
}

// The probe child runs every split owner side's sub-shards with a disagreement planted in each owner's first case: only
// the sub-shard compilerAgreementPart deals that case to may fail, once for each owner side, and every other part must
// pass. Checker/all's groups are left out, as its sub-shards compare a measured group, not a probed case.
func TestCompilerAndStage1AgreeSubShardPlantedDisagreement(test *testing.T) {
	test.Parallel()
	var shards, owners []string
	parts := 0
	for _, split := range compilerAgreementSplit {
		if split.Owner == compilerAgreementCheckerAllOwner {
			continue
		}
		shard := split.Owner*compilerAgreementSides + split.Side
		shards = append(shards, strconv.Itoa(shard))
		owners = append(owners, fmt.Sprintf(
			"TestCompilerAndStage1Agree_%d_%d",
			shard,
			compilerAgreementPart(compilerAgreementSubShardPlanted(test, split.Owner), split.Parts),
		))
		parts += split.Parts
	}
	command, ctx, cancel := compilerAgreementCommand(
		os.Args[0],
		"-test.run=^TestCompilerAndStage1Agree_("+strings.Join(shards, "|")+")_[0-9]+$",
		"-test.timeout=90s",
		"-test.v",
	)
	command.Env = append(os.Environ(), "ADAMIC_COMPILER_AGREEMENT_PROBE=1")
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	cancel()
	if captureError == nil || bytes.Count(output, []byte("--- FAIL:")) != len(owners) {
		test.Fatalf(
			"planted sub-shard disagreements not caught once for each of %d owner sides: %v\n%s",
			len(owners),
			captureError,
			output,
		)
	}
	for _, owner := range owners {
		if !bytes.Contains(output, []byte("--- FAIL: "+owner+" ")) {
			test.Fatalf("planted sub-shard disagreement not caught by %s\n%s", owner, output)
		}
	}
	if passed := bytes.Count(output, []byte("--- PASS: TestCompilerAndStage1Agree_")); passed != parts-len(owners) {
		test.Fatalf("%d sub-shards passed beside the owners, want %d\n%s", passed, parts-len(owners), output)
	}
	test.Logf(
		"planted disagreements in %d split owner sides each caught only by %s",
		len(owners),
		strings.Join(owners, ", "),
	)
}

func TestCompilerAndStage1Agree_6564_0(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 0)
}

func TestCompilerAndStage1Agree_6564_1(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 1)
}

func TestCompilerAndStage1Agree_6564_2(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 2)
}

func TestCompilerAndStage1Agree_6564_3(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 3)
}

func TestCompilerAndStage1Agree_6564_4(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 4)
}

func TestCompilerAndStage1Agree_6564_5(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 5)
}

func TestCompilerAndStage1Agree_6564_6(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 6)
}

func TestCompilerAndStage1Agree_6564_7(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 7)
}

func TestCompilerAndStage1Agree_6564_8(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 8)
}

func TestCompilerAndStage1Agree_6564_9(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 9)
}

func TestCompilerAndStage1Agree_6564_10(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 10)
}

func TestCompilerAndStage1Agree_6564_11(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 11)
}

func TestCompilerAndStage1Agree_6564_12(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 12)
}

func TestCompilerAndStage1Agree_6564_13(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 13)
}

func TestCompilerAndStage1Agree_6564_14(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 14)
}

func TestCompilerAndStage1Agree_6564_15(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 15)
}

func TestCompilerAndStage1Agree_6564_16(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 16)
}

func TestCompilerAndStage1Agree_6564_17(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 17)
}

func TestCompilerAndStage1Agree_6564_18(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 18)
}

func TestCompilerAndStage1Agree_6564_19(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 19)
}

func TestCompilerAndStage1Agree_6564_20(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 20)
}

func TestCompilerAndStage1Agree_6564_21(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 21)
}

func TestCompilerAndStage1Agree_6564_22(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 22)
}

func TestCompilerAndStage1Agree_6564_23(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 23)
}

func TestCompilerAndStage1Agree_6564_24(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 24)
}

func TestCompilerAndStage1Agree_6564_25(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 25)
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_cond_assign(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-cond-assign")
}

func TestCompilerAndStage1Agree_6564_26(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 26)
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_extra_bind(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-extra-bind")
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_unused_labels(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-unused-labels")
}

func TestCompilerAndStage1Agree_6564_27(test *testing.T) {
	test.Parallel()
	compilerAgreementSubShard(test, 6564, 27)
}

func TestCompilerAndStage1Agree_CheckerSanitized_no_useless_rename(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "no-useless-rename")
}

func TestCompilerAndStage1Agree_CheckerSanitized_default_param_last(test *testing.T) {
	test.Parallel()
	compilerAgreementCheckerSanitized(test, "default-param-last")
}
