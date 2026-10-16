package json

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Upstream parity is one product of its pins (Loom's ruling on #testaudit, #fqvxk72). Its build answers the whole
// corpus three times, with json's Go oracle, with cohere's audit binary and with the pinned Prettier, and holds each Go
// answer set to Prettier's: every case that differs must be allowed (jsonUpstreamAllowed) and its block must be the
// checked-in report's, and no other case may differ. Any of that failing fails the build, so a moved pin that breaks
// parity is a red product unit. Its key names the three pins (cohere, the corpus and Prettier) and the spec files, and
// TestUpstreamRepositoryCorpusParity reads its reports.
//
// It replaces two families that made these comparisons over the same corpus: 2,048 top shards (json's Go oracle
// against Prettier) and 16 split shards (cohere's audit binary against Prettier, over sampledCorpusCases, which is the
// whole corpus or a subset of it). Each checked the same allowed list, the same known report and that every known
// difference is a corpus case, and so does this, over every case for both Go answer sets.

// The pinned Prettier the gate installs at ADAMIC_JSON_PRETTIER, which adamic's tree doesn't hold: its version and the
// sha256 of the package-lock.json npm installed it from. The build refuses an installed Prettier that doesn't match.
type prettierPin struct {
	Version           string `json:"version"`
	PackageLockSHA256 string `json:"packageLockSha256"`
}

func readPrettierPin(root string) (prettierPin, error) {
	var pin prettierPin
	encoded, err := os.ReadFile(filepath.Join(root, "stage1/cohere/json/testdata/prettier.pin"))
	if err != nil {
		return pin, err
	}
	if err := json.Unmarshal(encoded, &pin); err != nil {
		return pin, fmt.Errorf("testdata/prettier.pin: %w", err)
	}
	if pin.Version == "" || len(pin.PackageLockSHA256) != 64 {
		return pin, fmt.Errorf("testdata/prettier.pin names version %q and package-lock.json sha256 %q", pin.Version, pin.PackageLockSHA256)
	}
	return pin, nil
}

// jsonUpstreamParityFiles are the verdict's spec: the comparison (this file), the generated cases and the allowed list
// (audit_test.go), the known report, the Prettier driver and the two pins the repository holds.
var jsonUpstreamParityFiles = []string{
	"stage1/cohere/json/upstream_parity_test.go",
	"stage1/cohere/json/audit_test.go",
	"stage1/cohere/json/known-upstream-differences.txt",
	"stage1/cohere/json/testdata/library.mjs",
	"stage1/cohere/json/testdata/prettier.pin",
	"stage1/cohere/json/testdata/corpus.pin",
}

// jsonUpstreamParityInputs keys the verdict by the cohere pin, both Go oracle binaries' bytes (as the class-order answer
// keys its oracle), the corpus pin and the Prettier pin read from root, and the spec files. Keying reads the repository
// and the oracles, never the installed Prettier.
func jsonUpstreamParityInputs(root, cohereCommit, corpusCommit string, oracle, audit []byte) (buildcache.Inputs, error) {
	pin, err := readPrettierPin(root)
	if err != nil {
		return buildcache.Inputs{}, err
	}
	return buildcache.Inputs{
		Name:  "json upstream parity verdict",
		Files: jsonUpstreamParityFiles,
		Flags: []string{
			"cohere " + cohereCommit,
			fmt.Sprintf("Go oracle binary sha256 %x", sha256.Sum256(oracle)),
			fmt.Sprintf("cohere audit binary sha256 %x", sha256.Sum256(audit)),
			"repository JSON at " + corpusCommit,
			"Prettier " + pin.Version + " package-lock.json sha256 " + pin.PackageLockSHA256,
		},
		Toolchain: []string{buildcache.Tool("node", "--version")},
	}, nil
}

// The known report, by case: three lines each, two cases.
type jsonUpstreamReport struct {
	text   string
	blocks map[string]string
}

func readJSONUpstreamReport(path string) (jsonUpstreamReport, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		return jsonUpstreamReport{}, err
	}
	lines := strings.Split(strings.TrimSuffix(string(encoded), "\n"), "\n")
	if len(lines) != 6 {
		return jsonUpstreamReport{}, fmt.Errorf("known upstream report has %d lines, want 6", len(lines))
	}
	report := jsonUpstreamReport{text: string(encoded), blocks: map[string]string{}}
	for i := 0; i < len(lines); i += 3 {
		if _, exists := report.blocks[lines[i]]; exists {
			return jsonUpstreamReport{}, fmt.Errorf("known upstream report repeats %s", lines[i])
		}
		report.blocks[lines[i]] = strings.Join(lines[i:i+3], "\n") + "\n"
	}
	return report, nil
}

func jsonUpstreamBlock(id string, left, right answer) string {
	if sameAnswer(left, right) {
		return ""
	}
	return fmt.Sprintf("%s\nGo: %q error=%q\nPrettier: %q error=%q\n", id, left.Output, left.Error, right.Output, right.Error)
}

// jsonUpstreamVerdict holds one Go answer set to Prettier's over cases and returns the report it makes. Every
// difference must be allowed and equal to the known report's block, every known difference must be a case, and the
// report, in corpus order, must be the known report's bytes, and no case may be repeated; every failure is named by side
// and case.
func jsonUpstreamVerdict(side string, cases []textCase, goAnswers, prettierAnswers []answer, known jsonUpstreamReport) (string, error) {
	if len(goAnswers) != len(cases) || len(prettierAnswers) != len(cases) {
		return "", fmt.Errorf("%s: %d Go and %d Prettier answers for %d cases", side, len(goAnswers), len(prettierAnswers), len(cases))
	}
	var failures []string
	ids := map[string]bool{}
	var report strings.Builder
	for i, item := range cases {
		if ids[item.Name] {
			failures = append(failures, "repeated case "+item.Name)
		}
		ids[item.Name] = true
		block := jsonUpstreamBlock(item.Name, goAnswers[i], prettierAnswers[i])
		if block != "" && !jsonUpstreamAllowed(item.Name) {
			failures = append(failures, "unexpected upstream difference: "+item.Name)
		} else if block != known.blocks[item.Name] {
			failures = append(failures, "case "+item.Name+": upstream difference identity or answer changed; update the checked-in report")
		}
		report.WriteString(block)
	}
	for id := range known.blocks {
		if !ids[id] {
			failures = append(failures, "known difference missing from corpus: "+id)
		}
	}
	if len(failures) == 0 && report.String() != known.text {
		failures = append(failures, "the report's blocks are the known report's but not in its order")
	}
	if len(failures) > 20 {
		failures = append(failures[:20], fmt.Sprintf("and %d more", len(failures)-20))
	}
	if len(failures) > 0 {
		return "", fmt.Errorf("%s against Prettier: %s", side, strings.Join(failures, "; "))
	}
	return report.String(), nil
}

// buildJSONUpstreamParity answers cases with the three oracles at once and writes each Go side's report, which the
// verdict has already held to the known one. The cases and answers stay in a scratch directory: the product is the
// reports.
func buildJSONUpstreamParity(cases []textCase, oracle, audit, library string) func(string) error {
	return func(directory string) error {
		root, err := filepath.Abs(repository)
		if err != nil {
			return err
		}
		pin, err := readPrettierPin(root)
		if err != nil {
			return err
		}
		lock, err := os.ReadFile(filepath.Join(library, "package-lock.json"))
		if err != nil {
			return fmt.Errorf("ADAMIC_JSON_PRETTIER: %w", err)
		}
		var installed struct {
			Version string `json:"version"`
		}
		manifest, err := os.ReadFile(filepath.Join(library, "node_modules/prettier/package.json"))
		if err == nil {
			err = json.Unmarshal(manifest, &installed)
		}
		if err != nil {
			return fmt.Errorf("ADAMIC_JSON_PRETTIER: %w", err)
		}
		if digest := fmt.Sprintf("%x", sha256.Sum256(lock)); digest != pin.PackageLockSHA256 || installed.Version != pin.Version {
			return fmt.Errorf("ADAMIC_JSON_PRETTIER %s installs Prettier %s from a package-lock.json with sha256 %s; testdata/prettier.pin names %s and %s", library, installed.Version, digest, pin.Version, pin.PackageLockSHA256)
		}
		known, err := readJSONUpstreamReport("known-upstream-differences.txt")
		if err != nil {
			return err
		}
		scratch, err := os.MkdirTemp("", "adamic-json-upstream-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		casesPath, protocolPath := filepath.Join(scratch, "cases.json"), filepath.Join(scratch, "cases.txt")
		encoded, err := json.Marshal(cases)
		if err != nil {
			return err
		}
		input, _ := protocol(cases, make([]answer, len(cases)))
		if err := os.WriteFile(casesPath, encoded, 0o644); err != nil {
			return err
		}
		if err := os.WriteFile(protocolPath, []byte(input), 0o644); err != nil {
			return err
		}
		cohere, err := filepath.Abs("../../../cohere")
		if err != nil {
			return err
		}
		readAnswers := func(name, path string) ([]answer, error) {
			encoded, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			var answers []answer
			if err := json.Unmarshal(encoded, &answers); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			return answers, nil
		}
		var goAnswers, auditAnswers, prettierAnswers []answer
		errs := make([]error, 3)
		var group sync.WaitGroup
		group.Add(3)
		go func() {
			defer group.Done()
			result, err := executeResult(nil, oracle, "--cases", protocolPath)
			if err == nil && (result.exitCode != 0 || len(result.stderr) != 0) {
				err = fmt.Errorf("exit %d: %s", result.exitCode, result.stderr)
			}
			if err == nil {
				goAnswers, err = jsonPortAnswers(string(result.stdout), len(cases))
			}
			if err != nil {
				errs[0] = fmt.Errorf("Go oracle: %w", err)
			}
		}()
		go func() {
			defer group.Done()
			answersPath := filepath.Join(scratch, "audit.json")
			command := exec.Command(audit, "-test.v", "-test.run=^TestAdamicJSONAudit$")
			command.Dir = filepath.Join(cohere, "internal/format/javascript")
			command.Env = append(os.Environ(), "ADAMIC_JSON_CASES="+casesPath, "ADAMIC_JSON_ANSWERS="+answersPath)
			if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
				errs[1] = fmt.Errorf("cohere audit: %v\n%s", err, output)
				return
			}
			auditAnswers, errs[1] = readAnswers("cohere audit", answersPath)
		}()
		go func() {
			defer group.Done()
			answersPath := filepath.Join(scratch, "prettier.json")
			// The whole corpus in one process, with the stack the deepest nesting needs.
			command := toolchain.Command("node", "--stack-size=4096", "testdata/library.mjs", library, casesPath, answersPath)
			if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
				errs[2] = fmt.Errorf("Prettier: %v\n%s", err, output)
				return
			}
			prettierAnswers, errs[2] = readAnswers("Prettier", answersPath)
		}()
		group.Wait()
		if err := errors.Join(errs...); err != nil {
			return err
		}
		for _, side := range []struct {
			name, file string
			answers    []answer
		}{
			{"json's Go oracle", "go-oracle.txt", goAnswers},
			{"cohere's audit binary", "cohere-audit.txt", auditAnswers},
		} {
			report, err := jsonUpstreamVerdict(side.name, cases, side.answers, prettierAnswers, known)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, side.file), []byte(report), 0o644); err != nil {
				return err
			}
		}
		summary := fmt.Sprintf("%d cases; %d known differences, from both Go oracles; Prettier %s\n", len(cases), len(known.blocks), pin.Version)
		return os.WriteFile(filepath.Join(directory, "summary.txt"), []byte(summary), 0o644)
	}
}

var jsonUpstreamParity struct {
	once      sync.Once
	directory string
}

// jsonUpstreamParityProduct is the verdict's directory. Without ADAMIC_JSON_PRETTIER it skips, as every Prettier
// comparison here does.
func jsonUpstreamParityProduct(t *testing.T) string {
	t.Helper()
	library := os.Getenv("ADAMIC_JSON_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_JSON_PRETTIER for the separate upstream report")
	}
	jsonUpstreamParity.once.Do(func() {
		if err := gatesample.Validate(); err != nil {
			t.Fatal(err)
		}
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		oracle, audit := portMatchesProduct(t, "oracle"), upstreamParityOracleProduct(t)
		binaries := make([][]byte, 2)
		for i, path := range []string{oracle, audit} {
			if binaries[i], err = os.ReadFile(path); err != nil {
				t.Fatal(err)
			}
		}
		inputs, err := jsonUpstreamParityInputs(root, corpusfiles.CohereCommit, repositoryCorpusCommit, binaries[0], binaries[1])
		if err != nil {
			t.Fatal(err)
		}
		// The whole corpus, not a gate sample, read before the build: listing the tracked JSON reads .git, which a
		// traced build refuses.
		jsonUpstreamParity.directory = buildcache.Product(t, inputs, buildJSONUpstreamParity(corpusCases(t), oracle, audit, library))
	})
	if jsonUpstreamParity.directory == "" {
		t.Fatal("the upstream parity verdict failed earlier in this process")
	}
	return jsonUpstreamParity.directory
}

func TestProduct_JSONUpstreamParityVerdict(t *testing.T) {
	t.Parallel()
	jsonUpstreamParityProduct(t)
}

// Prettier is a separate upstream report. Only the two named disagreements are known; an added difference or a
// closed one fails the verdict's build until the report is updated explicitly. This reads the verdict and holds both
// reports to the checked-in one again, and writes the report to ADAMIC_JSON_REPORT when that is set.
func TestUpstreamRepositoryCorpusParity(t *testing.T) {
	t.Parallel()
	directory := jsonUpstreamParityProduct(t)
	known, err := readJSONUpstreamReport("known-upstream-differences.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"go-oracle.txt", "cohere-audit.txt"} {
		report, err := os.ReadFile(filepath.Join(directory, file))
		if err != nil {
			t.Fatal(err)
		}
		if string(report) != known.text {
			t.Fatalf("the verdict's %s is not known-upstream-differences.txt", file)
		}
		if path := os.Getenv("ADAMIC_JSON_REPORT"); path != "" && file == "go-oracle.txt" {
			if err := os.WriteFile(path, report, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	summary, err := os.ReadFile(filepath.Join(directory, "summary.txt"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("upstream parity verdict: %s", strings.TrimSpace(string(summary)))
}

// Each pin and each spec file moves the verdict's key: the cohere pin and the oracles' bytes, the corpus pin and
// corpus.pin, the Prettier pin, the allowed list, the known report and the Prettier driver. A pin is a flag, so it moves
// the name key (what a runner finds the verdict by) and the declared key (what this machine alone builds it by); a
// file moves the declared key, and the traced read set the runner's. Keying builds nothing.
func TestJSONUpstreamParityRekeys(t *testing.T) {
	t.Parallel()
	repositoryRoot, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// A stand-in tree holding only the declared files, so each edit is made to a copy.
	stand := t.TempDir()
	for _, name := range jsonUpstreamParityFiles {
		content, err := os.ReadFile(filepath.Join(repositoryRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(stand, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(stand, name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	type keyed struct{ nameKey, declared string }
	keys := func(cohereCommit, corpusCommit string, oracle []byte) keyed {
		t.Helper()
		inputs, err := jsonUpstreamParityInputs(stand, cohereCommit, corpusCommit, oracle, []byte("audit"))
		if err != nil {
			t.Fatal(err)
		}
		declared, err := buildcache.Key(stand, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return keyed{buildcache.NameKey(stand, inputs), declared}
	}
	other := func(commit string) string {
		if commit[39] == '1' {
			return commit[:39] + "2"
		}
		return commit[:39] + "1"
	}
	base := keys(corpusfiles.CohereCommit, repositoryCorpusCommit, []byte("oracle"))
	for _, move := range []struct {
		name string
		key  keyed
	}{
		{"cohere pin", keys(other(corpusfiles.CohereCommit), repositoryCorpusCommit, []byte("oracle"))},
		{"Go oracle binary", keys(corpusfiles.CohereCommit, repositoryCorpusCommit, []byte("oracle2"))},
		{"corpus pin", keys(corpusfiles.CohereCommit, other(repositoryCorpusCommit), []byte("oracle"))},
	} {
		if move.key.nameKey == base.nameKey || move.key.declared == base.declared {
			t.Fatalf("moving the %s kept name key %s or key %s", move.name, base.nameKey[:12], base.declared[:12])
		}
	}
	edit := func(name, from, to string, nameKeyMoves bool) {
		t.Helper()
		path := filepath.Join(stand, name)
		original, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(original), from) != 1 {
			t.Fatalf("%s holds %q %d times, want once", name, from, strings.Count(string(original), from))
		}
		if err := os.WriteFile(path, []byte(strings.Replace(string(original), from, to, 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		edited := keys(corpusfiles.CohereCommit, repositoryCorpusCommit, []byte("oracle"))
		if err := os.WriteFile(path, original, 0o644); err != nil {
			t.Fatal(err)
		}
		if edited.declared == base.declared || (edited.nameKey != base.nameKey) != nameKeyMoves {
			t.Fatalf("editing %s: key %s to %s, name key %s to %s (should move: %t)", name, base.declared[:12], edited.declared[:12], base.nameKey[:12], edited.nameKey[:12], nameKeyMoves)
		}
		t.Logf("editing %s moved key %s to %s", name, base.declared[:12], edited.declared[:12])
	}
	edit("stage1/cohere/json/testdata/prettier.pin", "70695602", "70695603", true)
	edit("stage1/cohere/json/audit_test.go", `strings.HasPrefix(name, "generated/16/")`, `strings.HasPrefix(name, "generated/15/")`, false)
	edit("stage1/cohere/json/known-upstream-differences.txt", "generated/16/package.json", "generated/15/package.json", false)
	edit("stage1/cohere/json/testdata/library.mjs", "texts in", "cases in", false)
	edit("stage1/cohere/json/testdata/corpus.pin", `"count": 1113`, `"count": 1114`, false)
	if keys(corpusfiles.CohereCommit, repositoryCorpusCommit, []byte("oracle")) != base {
		t.Fatal("restoring every file did not restore the key")
	}
}

// The verdict fails on a changed Prettier answer, wherever it lands: on a case that agreed (an unlisted difference),
// on a known difference (its block changed), or closing one (a known difference gone). The answers are the known
// report's own, read back, beside two cases on which Go and Prettier agree.
func TestJSONUpstreamVerdictCatchesChangedAnswers(t *testing.T) {
	t.Parallel()
	known, err := readJSONUpstreamReport("known-upstream-differences.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := []textCase{{"stage1/cohere/json/package.json", "{}"}}
	goAnswers := []answer{{Output: "{}\n"}}
	prettierAnswers := []answer{{Output: "{}\n"}}
	lines := strings.Split(strings.TrimSuffix(known.text, "\n"), "\n")
	side := func(line, prefix string) answer {
		t.Helper()
		// Go: "<output>" error="<error>", each quoted by %q.
		rest, found := strings.CutPrefix(line, prefix)
		output, after, cut := strings.Cut(rest, " error=")
		if !found || !cut {
			t.Fatalf("malformed known line %q", line)
		}
		var value answer
		if value.Output, err = strconv.Unquote(output); err != nil {
			t.Fatal(err)
		}
		if value.Error, err = strconv.Unquote(after); err != nil {
			t.Fatal(err)
		}
		return value
	}
	for i := 0; i < len(lines); i += 3 {
		cases = append(cases, textCase{lines[i], "x"})
		goAnswers = append(goAnswers, side(lines[i+1], "Go: "))
		prettierAnswers = append(prettierAnswers, side(lines[i+2], "Prettier: "))
	}
	cases = append(cases, textCase{"generated/0/probe.json", "[]"})
	goAnswers = append(goAnswers, answer{Output: "[]\n"})
	prettierAnswers = append(prettierAnswers, answer{Output: "[]\n"})
	report, err := jsonUpstreamVerdict("baseline", cases, goAnswers, prettierAnswers, known)
	if err != nil || report != known.text {
		t.Fatalf("the known answers' verdict: %v; report equals the known one: %t", err, report == known.text)
	}
	for _, mutant := range []struct {
		name, want string
		index      int
		answer     answer
	}{
		{"an agreeing case's answer changed", "unexpected upstream difference: stage1/cohere/json/package.json", 0, answer{Output: "{ }\n"}},
		{"an agreeing case refused", "unexpected upstream difference: generated/0/probe.json", len(cases) - 1, answer{Error: "refused"}},
		{"a known difference's answer changed", "case generated/16/probe.json: upstream difference identity or answer changed", 1, answer{Error: "another message"}},
		{"a known difference closed", "case generated/16/package.json: upstream difference identity or answer changed", 2, goAnswers[2]},
	} {
		changed := append([]answer(nil), prettierAnswers...)
		changed[mutant.index] = mutant.answer
		if _, err := jsonUpstreamVerdict("mutant", cases, goAnswers, changed, known); err == nil || !strings.Contains(err.Error(), mutant.want) {
			t.Fatalf("%s: verdict %v, want %q", mutant.name, err, mutant.want)
		}
	}
	repeated := append(append([]textCase(nil), cases...), cases[0])
	if _, err := jsonUpstreamVerdict("repeated", repeated, append(goAnswers, goAnswers[0]), append(prettierAnswers, prettierAnswers[0]), known); err == nil || !strings.Contains(err.Error(), "repeated case stage1/cohere/json/package.json") {
		t.Fatalf("a repeated case: %v", err)
	}
	if _, err := jsonUpstreamVerdict("missing", cases[:len(cases)-2], goAnswers[:len(cases)-2], prettierAnswers[:len(cases)-2], known); err == nil || !strings.Contains(err.Error(), "known difference missing from corpus: generated/16/package.json") {
		t.Fatalf("a known difference dropped from the corpus: %v", err)
	}
	t.Log("an unlisted difference, a changed known block, a closed difference and a missing known case each fail the verdict")
}

var upstreamParityOracleBuild struct {
	once sync.Once
	path string
}

// upstreamParityOracleProduct is cohere's internal/format/javascript test binary with testdata/cohere_side_test.go
// added by an overlay, keyed by what go says the build compiles (GoTest). The overlay names this checkout's paths, so
// it sits in a temporary directory: it builds the product and is never part of it (#tqrqx60).
func upstreamParityOracleProduct(t *testing.T) string {
	t.Helper()
	upstreamParityOracleBuild.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		side, err := filepath.Abs("testdata/cohere_side_test.go")
		if err != nil {
			t.Fatal(err)
		}
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/internal/format/javascript/adamic_json_audit_test.go"): side}})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			t.Fatal(err)
		}
		upstreamParityOracleBuild.path = buildcache.GoTest(t, "cohere", "oracle", "./internal/format/javascript", []string{"-overlay=" + path})
	})
	if upstreamParityOracleBuild.path == "" {
		t.Fatal("upstream oracle preparation failed")
	}
	return upstreamParityOracleBuild.path
}

func TestProduct_JSONUpstreamOracle(t *testing.T) {
	t.Parallel()
	upstreamParityOracleProduct(t)
}
