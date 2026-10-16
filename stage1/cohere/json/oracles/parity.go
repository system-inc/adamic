// Package oracles owns json's compiler-free Go reference products.
// Both test binaries resolve relative paths from stage1/cohere/json.
package oracles

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/internal/toolchain"
)

type PrettierPin struct {
	Version           string `json:"version"`
	PackageLockSHA256 string `json:"packageLockSha256"`
}

func readPrettierPin(root string) (PrettierPin, error) {
	var pin PrettierPin
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

var upstreamParityFiles = []string{
	"stage1/cohere/json/upstream_parity_test.go",
	"stage1/cohere/json/oracles/parity.go",
	"stage1/cohere/json/oracles/corpus.go",
	"stage1/cohere/json/oracles/oracles.go",
	"stage1/cohere/json/audit_test.go",
	"stage1/cohere/json/known-upstream-differences.txt",
	"stage1/cohere/json/testdata/library.mjs",
	"stage1/cohere/json/testdata/prettier.pin",
	"stage1/cohere/json/testdata/corpus.pin",
}

// UpstreamParityInputFiles is shared with the parent json tests.
func UpstreamParityInputFiles() []string { return append([]string(nil), upstreamParityFiles...) }

// UpstreamParityInputs is shared with the parent json tests.
func UpstreamParityInputs(root, cohereCommit, corpusCommit string, oracle, audit []byte) (buildcache.Inputs, error) {
	pin, err := readPrettierPin(root)
	if err != nil {
		return buildcache.Inputs{}, err
	}
	return buildcache.Inputs{
		Name:  "json upstream parity verdict oracle-owner-v1",
		Files: upstreamParityFiles,
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

// ReadUpstreamReport is shared with the parent json tests.
func ReadUpstreamReport(path string) (string, map[string]string, error) {
	r, err := readJSONUpstreamReport(path)
	return r.text, r.blocks, err
}

func upstreamBlock(id string, left, right ComparisonAnswer) string {
	if SameAnswer(left, right) {
		return ""
	}
	return fmt.Sprintf("%s\nGo: %q error=%q\nPrettier: %q error=%q\n", id, left.Output, left.Error, right.Output, right.Error)
}

func upstreamVerdict(side string, cases []ComparisonCase, goAnswers, prettierAnswers []ComparisonAnswer, known jsonUpstreamReport) (string, error) {
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
		block := upstreamBlock(item.Name, goAnswers[i], prettierAnswers[i])
		if block != "" && !upstreamAllowed(item.Name) {
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

// UpstreamVerdict is shared with the parent json tests.
func UpstreamVerdict(side string, cases []ComparisonCase, goAnswers, prettierAnswers []ComparisonAnswer, text string, blocks map[string]string) (string, error) {
	return upstreamVerdict(side, cases, goAnswers, prettierAnswers, jsonUpstreamReport{text, blocks})
}

func upstreamAllowed(name string) bool {
	return strings.HasPrefix(name, "generated/16/")
}

func buildJSONUpstreamParity(cases []ComparisonCase, oracle, audit, library string) func(string) error {
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
		input, _ := Protocol(cases, make([]ComparisonAnswer, len(cases)))
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
		readAnswers := func(name, path string) ([]ComparisonAnswer, error) {
			encoded, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			var answers []ComparisonAnswer
			if err := json.Unmarshal(encoded, &answers); err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
			return answers, nil
		}
		var goAnswers, auditAnswers, prettierAnswers []ComparisonAnswer
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
				goAnswers, err = ParseAnswers(string(result.stdout), len(cases))
			}
			if err != nil {
				errs[0] = fmt.Errorf("Go oracle: %w", err)
			}
		}()
		go func() {
			defer group.Done()
			answersPath := filepath.Join(scratch, "audit.json")
			command := toolchain.CommandIn("", audit, "-test.v", "-test.run=^TestAdamicJSONAudit$")
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
			answers    []ComparisonAnswer
		}{
			{"json's Go oracle", "go-oracle.txt", goAnswers},
			{"cohere's audit binary", "cohere-audit.txt", auditAnswers},
		} {
			report, err := upstreamVerdict(side.name, cases, side.answers, prettierAnswers, known)
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

// UpstreamParity is shared with the parent json tests.
func UpstreamParity(t *testing.T) string {
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
		oracle, audit := FormatterOracle(t), AuditOracle(t)
		binaries := make([][]byte, 2)
		for i, path := range []string{oracle, audit} {
			if binaries[i], err = os.ReadFile(path); err != nil {
				t.Fatal(err)
			}
		}
		inputs, err := UpstreamParityInputs(root, corpusfiles.CohereCommit, RepositoryCorpusCommit, binaries[0], binaries[1])
		if err != nil {
			t.Fatal(err)
		}
		// The whole corpus, not a gate sample, read before the build: listing the tracked JSON reads .git, which a
		// traced build refuses.
		jsonUpstreamParity.directory = oracleProduct(t, inputs, buildJSONUpstreamParity(CorpusCases(t), oracle, audit, library))
	})
	if jsonUpstreamParity.directory == "" {
		t.Fatal("the upstream parity verdict failed earlier in this process")
	}
	return jsonUpstreamParity.directory
}
