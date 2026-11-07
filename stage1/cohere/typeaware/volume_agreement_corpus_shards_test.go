package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// volumeAgreementCompilerGroups is 32, not 16: at 16, checker.ts (3,077 KB) and types.ts (476 KB) shared ASan group 9,
// 25.3 s of own work alone on one core (Home), while no other ASan group took 9 s. At 32 checker.ts keeps only
// _namespaces/ts.ts (3 KB), in group 25 (shards _025 plain and _057 ASan); the union is parametric in the count.
const volumeAgreementCompilerGroups = 32

// volumeAgreementSplitFile is the one compiler file too heavy for an ASan unit by itself: every rule over checker.ts
// took 31.6 s of the ASan native alone (Pupa, load 7.5), against 3.5 s for the Go oracle and 0.5 s for the program
// load, and held TestVolumeAgreementCompiler_057 at 59 to 63 s under load. Under ASan it leaves group 25 (_057 keeps
// _namespaces/ts.ts) and runs as volumeAgreementCheckerParts, each part some of the native's walks (volume_suite.ts
// --rules) against the oracle's same subjects. The gap this leaves: ASan sees checker.ts's rule families in separate
// processes, never together. Plain TestVolumeAgreementCompiler_025 still runs every rule over checker.ts in one
// process, and every other file runs every rule together under ASan too.
const volumeAgreementSplitFile = "src/compiler/checker.ts"

// volumeAgreementCheckerPartCount is how many TestVolumeAgreementCompilerChecker_ tests there are, one a part.
const volumeAgreementCheckerPartCount = 5

// volumeAgreementCheckerParts group volume_suite.ts's walks, each listed whole, by their ASan seconds over checker.ts
// alone (Golem, wall, the oracle's own after it): Rules 9.5 (2.0), Volume 13.2 (2.2), Shadow 6.4 (1.4), Assignment 7.2
// (1.3), Void 6.9 (1.3), Returns 5.4 (0.8), Unbound 11.4 (1.2), against 32.6 (3.7) for every walk at once. About 4.5 s
// of each is the native's own fixed cost (parsing checker.ts and linking its nodes under ASan), so the light walks
// share a part. Volume is one walk, its 13.2 s the floor of a part until its rules are gated one by one inside it.
var volumeAgreementCheckerParts = [][]string{
	{
		"@typescript-eslint/no-unsafe-unary-minus",
		"@typescript-eslint/related-getter-setter-pairs",
		"@typescript-eslint/no-unsafe-declaration-merging",
		"@typescript-eslint/no-unsafe-argument",
		"@typescript-eslint/restrict-plus-operands",
		"@typescript-eslint/no-unnecessary-boolean-literal-compare",
	},
	{
		"@typescript-eslint/prefer-nullish-coalescing",
		"@typescript-eslint/switch-exhaustiveness-check",
		"@typescript-eslint/no-unsafe-enum-comparison",
		"@typescript-eslint/no-unsafe-type-assertion",
		"@typescript-eslint/no-unsafe-member-access",
	},
	{"@typescript-eslint/no-shadow", "@typescript-eslint/no-unsafe-assignment"},
	{"@typescript-eslint/no-confusing-void-expression", "@typescript-eslint/consistent-return"},
	{"@typescript-eslint/unbound-method"},
}

// volumeAgreementOracleSubjects are the rules oracle_volume.go runs, read from its subject list: rules.NoShadow is
// @typescript-eslint/no-shadow.
func volumeAgreementOracleSubjects(test *testing.T) []string {
	test.Helper()
	data, readError := os.ReadFile("testdata/oracle_volume.go")
	if readError != nil {
		test.Fatal(readError)
	}
	_, list, ok := strings.Cut(string(data), "subjects := []rule.Rule{")
	list, _, closed := strings.Cut(list, "}")
	if !ok || !closed {
		test.Fatal("oracle_volume.go has no subject list")
	}
	var subjects []string
	for _, match := range regexp.MustCompile(`rules\.([A-Za-z]+)`).FindAllStringSubmatch(list, -1) {
		name := regexp.MustCompile(`[A-Z]`).ReplaceAllStringFunc(
			match[1],
			func(letter string) string { return "-" + strings.ToLower(letter) },
		)
		subjects = append(subjects, "@typescript-eslint/"+strings.TrimPrefix(name, "-"))
	}
	return subjects
}

// volumeAgreementPartsError says how parts fail to name each subject exactly once.
func volumeAgreementPartsError(parts [][]string, subjects []string) error {
	seen := map[string]int{}
	for _, part := range parts {
		if len(part) == 0 {
			return fmt.Errorf("empty part")
		}
		for _, name := range part {
			seen[name]++
		}
	}
	for _, subject := range subjects {
		if seen[subject] != 1 {
			return fmt.Errorf("subject %s occurs %d times in the parts", subject, seen[subject])
		}
		delete(seen, subject)
	}
	for name := range seen {
		return fmt.Errorf("part rule %s is no oracle subject", name)
	}
	return nil
}

type volumeAgreementCorpusFile struct{ key, path string }

func volumeAgreementCorpusOwner(key string, groups int) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % groups
}
func volumeAgreementCorpusFiles(test *testing.T, kind string) (string, []volumeAgreementCorpusFile) {
	test.Helper()
	var paths []string
	var root, config string
	switch kind {
	case "compiler":
		root = os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
		if root == "" {
			test.Skip("ADAMIC_TYPESCRIPT_SOURCE is not configured")
		}
		var absolutePathError error
		root, absolutePathError = filepath.Abs(root)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		paths = pinnedCompilerFiles(test, root)
		config = filepath.Join(root, "src/compiler/tsconfig.json")
	default:
		test.Fatalf("unknown corpus %s", kind)
	}
	seen := map[string]bool{}
	var files []volumeAgreementCorpusFile
	for _, path := range paths {
		key, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		key = filepath.ToSlash(key)
		if seen[key] {
			test.Fatalf("duplicate %s corpus root %s", kind, key)
		}
		seen[key] = true
		files = append(files, volumeAgreementCorpusFile{key, path})
	}
	if len(files) == 0 {
		test.Fatalf("empty %s corpus", kind)
	}
	return config, files
}
func volumeAgreementCorpusUnion(test *testing.T, kind string, groups int) {
	test.Helper()
	_, files := volumeAgreementCorpusFiles(test, kind)
	if len(volumeAgreementCheckerParts) != volumeAgreementCheckerPartCount {
		test.Fatalf(
			"enumerated checker parts %d differ from declared %d",
			len(volumeAgreementCheckerParts),
			volumeAgreementCheckerPartCount,
		)
	}
	if err := volumeAgreementPartsError(volumeAgreementCheckerParts, volumeAgreementOracleSubjects(test)); err != nil {
		test.Fatal(err)
	}
	for _, mode := range []string{"plain", "asan"} {
		seen := map[string]int{}
		for shard := 0; shard < groups; shard++ {
			for _, file := range files {
				if volumeAgreementCorpusOwner(file.key, groups) == shard && volumeAgreementGrouped(file.key, mode) {
					seen[file.key]++
				}
			}
		}
		if mode == "asan" {
			// Each part checks the split file over its own rules, so together they check it once.
			if seen[volumeAgreementSplitFile] != 0 {
				test.Fatalf("%s asan groups hold %s", kind, volumeAgreementSplitFile)
			}
			seen[volumeAgreementSplitFile] = 1
		}
		if len(seen) != len(files) {
			test.Fatalf("%s %s union=%d want %d", kind, mode, len(seen), len(files))
		}
		for _, file := range files {
			if seen[file.key] != 1 {
				test.Fatalf("%s %s file %s occurs %d times", kind, mode, file.key, seen[file.key])
			}
		}
		test.Logf(
			"%s %s counted union: %d files exactly once in %d groups and %d parts of %s",
			kind,
			mode,
			len(seen),
			groups,
			len(volumeAgreementCheckerParts),
			volumeAgreementSplitFile,
		)
	}
}

// volumeAgreementGrouped says whether a group shard of mode checks the file: the split file's ASan run is its parts'.
func volumeAgreementGrouped(key, mode string) bool {
	return mode != "asan" || key != volumeAgreementSplitFile
}
func runVolumeAgreementCorpusShard(test *testing.T, kind string, groups, unit int) {
	test.Helper()
	// Live enumeration and required product fetches precede this unit's work clock.
	config, files := volumeAgreementCorpusFiles(test, kind)
	group := unit % groups
	name := "volume"
	mode := "plain"
	if unit >= groups {
		name = "volume-asan"
		mode = "asan"
	}
	var paths []string
	for _, file := range files {
		if volumeAgreementCorpusOwner(file.key, groups) == group && volumeAgreementGrouped(file.key, mode) {
			paths = append(paths, file.path)
		}
	}
	test.Logf("%s %s group %d: %d/%d file roots", kind, mode, group, len(paths), len(files))
	if len(paths) == 0 {
		test.Log("own work=0s")
		return
	}
	volumeAgreementCorpusCompare(test, kind, mode, fmt.Sprintf("%03d", group), name, config, files, paths, nil)
}

// runVolumeAgreementCheckerPart checks the split file under ASan over one part's rules, in every program root.
func runVolumeAgreementCheckerPart(test *testing.T, kind string, part int) {
	test.Helper()
	config, files := volumeAgreementCorpusFiles(test, kind)
	var paths []string
	for _, file := range files {
		if file.key == volumeAgreementSplitFile {
			paths = append(paths, file.path)
		}
	}
	if len(paths) != 1 {
		test.Fatalf("%s corpus holds %s %d times", kind, volumeAgreementSplitFile, len(paths))
	}
	rules := volumeAgreementCheckerParts[part]
	test.Logf("%s asan %s part %d: %s", kind, volumeAgreementSplitFile, part, strings.Join(rules, " "))
	volumeAgreementCorpusCompare(
		test,
		kind,
		"asan",
		fmt.Sprintf("checker-%03d", part),
		"volume-asan",
		config,
		files,
		paths,
		rules,
	)
}

// volumeAgreementCorpusCompare runs the oracle and the native over paths, every file a program root, with only rules
// when it names some, and requires the same bytes.
func volumeAgreementCorpusCompare(
	test *testing.T,
	kind, mode, label, name, config string,
	files []volumeAgreementCorpusFile,
	paths, rules []string,
) {
	test.Helper()
	oracle := volumeAgreementFetch(test, "oracle")
	binary := volumeAgreementFetch(test, name)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() {
		test.Logf("own work=%.6fs", time.Since(started).Seconds())
		if timeoutContext.Err() != nil {
			test.Error("corpus shard own work exceeded 60s")
		}
	}()
	h := volumeGuardPreparationHarness(test)
	h.ctx = timeoutContext
	manifest := h.write(fmt.Sprintf("%s-%s-%s.manifest", kind, mode, label), strings.Join(paths, "\n")+"\n")
	var allPaths []string
	for _, file := range files {
		allPaths = append(allPaths, file.path)
	}
	all := h.write("program-roots.manifest", strings.Join(allPaths, "\n")+"\n")
	args := []string{config, manifest, "--program-roots", all}
	if len(rules) != 0 {
		args = append(args, "--rules", h.write("rules.manifest", strings.Join(rules, "\n")+"\n"))
	}
	truth := volumeGuardMust(h, kind+"-"+mode+"-go", exec.Command(oracle, args...))
	got := volumeGuardMust(h, kind+"-"+mode+"-native", exec.Command(binary, args...))
	if len(got.stderr) != 0 || !bytes.Equal(got.stdout, truth.stdout) {
		test.Fatalf("%s %s mismatch byte %d: %s", kind, mode, firstDifference(got.stdout, truth.stdout), got.stderr)
	}
	test.Logf(
		"%s %s: %s; Go %.3fs, native %.3fs",
		kind,
		mode,
		summary(truth.stdout),
		truth.elapsed.Seconds(),
		got.elapsed.Seconds(),
	)
}
func TestVolumeAgreementCompilerUnion(test *testing.T) {
	test.Parallel()
	volumeAgreementCorpusUnion(test, "compiler", volumeAgreementCompilerGroups)
}

// A dropped or a repeated part fails the union, as a part naming a rule the oracle doesn't run does.
func TestVolumeAgreementCheckerPartsUnion(test *testing.T) {
	test.Parallel()
	subjects := volumeAgreementOracleSubjects(test)
	if len(subjects) != 16 {
		test.Fatalf("oracle_volume.go names %d subjects, want 16: %q", len(subjects), subjects)
	}
	if err := volumeAgreementPartsError(volumeAgreementCheckerParts, subjects); err != nil {
		test.Fatal(err)
	}
	for i := range volumeAgreementCheckerParts {
		dropped := append(
			append([][]string{}, volumeAgreementCheckerParts[:i]...),
			volumeAgreementCheckerParts[i+1:]...,
		)
		if volumeAgreementPartsError(dropped, subjects) == nil {
			test.Fatalf("parts without part %d accepted", i)
		}
		repeated := append(append([][]string{}, volumeAgreementCheckerParts...), volumeAgreementCheckerParts[i])
		if volumeAgreementPartsError(repeated, subjects) == nil {
			test.Fatalf("parts with part %d twice accepted", i)
		}
	}
	if volumeAgreementPartsError(
		append(append([][]string{}, volumeAgreementCheckerParts...), []string{"@typescript-eslint/no-such-rule"}),
		subjects,
	) == nil {
		test.Fatal("a part naming no subject accepted")
	}
	test.Logf(
		"%d parts name the oracle's %d subjects exactly once; every planted drop and repeat fails",
		len(volumeAgreementCheckerParts),
		len(subjects),
	)
}
func TestVolumeAgreementCompiler_000(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 0)
}
func TestVolumeAgreementCompiler_001(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 1)
}
func TestVolumeAgreementCompiler_002(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 2)
}
func TestVolumeAgreementCompiler_003(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 3)
}
func TestVolumeAgreementCompiler_004(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 4)
}
func TestVolumeAgreementCompiler_005(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 5)
}
func TestVolumeAgreementCompiler_006(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 6)
}
func TestVolumeAgreementCompiler_007(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 7)
}
func TestVolumeAgreementCompiler_008(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 8)
}
func TestVolumeAgreementCompiler_009(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 9)
}
func TestVolumeAgreementCompiler_010(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 10)
}
func TestVolumeAgreementCompiler_011(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 11)
}
func TestVolumeAgreementCompiler_012(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 12)
}
func TestVolumeAgreementCompiler_013(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 13)
}
func TestVolumeAgreementCompiler_014(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 14)
}
func TestVolumeAgreementCompiler_015(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 15)
}
func TestVolumeAgreementCompiler_016(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 16)
}
func TestVolumeAgreementCompiler_017(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 17)
}
func TestVolumeAgreementCompiler_018(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 18)
}
func TestVolumeAgreementCompiler_019(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 19)
}
func TestVolumeAgreementCompiler_020(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 20)
}
func TestVolumeAgreementCompiler_021(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 21)
}
func TestVolumeAgreementCompiler_022(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 22)
}
func TestVolumeAgreementCompiler_023(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 23)
}
func TestVolumeAgreementCompiler_024(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 24)
}
func TestVolumeAgreementCompiler_025(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 25)
}
func TestVolumeAgreementCompiler_026(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 26)
}
func TestVolumeAgreementCompiler_027(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 27)
}
func TestVolumeAgreementCompiler_028(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 28)
}
func TestVolumeAgreementCompiler_029(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 29)
}
func TestVolumeAgreementCompiler_030(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 30)
}
func TestVolumeAgreementCompiler_031(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 31)
}
func TestVolumeAgreementCompiler_032(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 32)
}
func TestVolumeAgreementCompiler_033(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 33)
}
func TestVolumeAgreementCompiler_034(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 34)
}
func TestVolumeAgreementCompiler_035(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 35)
}
func TestVolumeAgreementCompiler_036(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 36)
}
func TestVolumeAgreementCompiler_037(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 37)
}
func TestVolumeAgreementCompiler_038(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 38)
}
func TestVolumeAgreementCompiler_039(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 39)
}
func TestVolumeAgreementCompiler_040(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 40)
}
func TestVolumeAgreementCompiler_041(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 41)
}
func TestVolumeAgreementCompiler_042(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 42)
}
func TestVolumeAgreementCompiler_043(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 43)
}
func TestVolumeAgreementCompiler_044(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 44)
}
func TestVolumeAgreementCompiler_045(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 45)
}
func TestVolumeAgreementCompiler_046(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 46)
}
func TestVolumeAgreementCompiler_047(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 47)
}
func TestVolumeAgreementCompiler_048(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 48)
}
func TestVolumeAgreementCompiler_049(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 49)
}
func TestVolumeAgreementCompiler_050(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 50)
}
func TestVolumeAgreementCompiler_051(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 51)
}
func TestVolumeAgreementCompiler_052(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 52)
}
func TestVolumeAgreementCompiler_053(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 53)
}
func TestVolumeAgreementCompiler_054(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 54)
}
func TestVolumeAgreementCompiler_055(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 55)
}
func TestVolumeAgreementCompiler_056(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 56)
}
func TestVolumeAgreementCompiler_057(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 57)
}
func TestVolumeAgreementCompiler_058(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 58)
}
func TestVolumeAgreementCompiler_059(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 59)
}
func TestVolumeAgreementCompiler_060(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 60)
}
func TestVolumeAgreementCompiler_061(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 61)
}
func TestVolumeAgreementCompiler_062(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 62)
}
func TestVolumeAgreementCompiler_063(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCorpusShard(test, "compiler", volumeAgreementCompilerGroups, 63)
}
func TestVolumeAgreementCompilerChecker_000(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCheckerPart(test, "compiler", 0)
}
func TestVolumeAgreementCompilerChecker_001(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCheckerPart(test, "compiler", 1)
}
func TestVolumeAgreementCompilerChecker_002(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCheckerPart(test, "compiler", 2)
}
func TestVolumeAgreementCompilerChecker_003(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCheckerPart(test, "compiler", 3)
}
func TestVolumeAgreementCompilerChecker_004(test *testing.T) {
	test.Parallel()
	runVolumeAgreementCheckerPart(test, "compiler", 4)
}
