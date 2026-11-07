package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"
)

// The port always matches Go. External printers retain their own exact bytes or
// parser errors; these outcomes are compared, never normalized into Go output.
func comparePrinterLibrary(test *testing.T, name string, result oracles.Run, specs, family string, embedded bool) {
	test.Helper()
	comparePrinterLibraryCases(test, name, result, specs, family, embedded, true)
}

// A unit checks every applicable pinned outcome; the parent checks the full census.
func comparePrinterLibraryCases(
	test *testing.T,
	name string,
	result oracles.Run,
	specs, family string,
	embedded, census bool,
) {
	test.Helper()
	if result.ExitCode != 0 || len(result.Stderr) != 0 {
		test.Fatalf("%s: exit %d stderr %s", name, result.ExitCode, result.Stderr)
	}
	data, err := os.ReadFile(specs)
	if err != nil {
		test.Fatal(err)
	}
	var cases []oracles.PrinterCase
	if err := json.Unmarshal(data, &cases); err != nil {
		test.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(result.Stdout), "\n"), "\n")
	if len(lines) != len(cases) {
		test.Fatalf("%s: %d answers, want %d", name, len(lines), len(cases))
	}
	records := oracles.UpstreamOutcomes(test)
	seen := make(map[string]bool)
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	differences, refusals := 0, 0
	for index, item := range cases {
		status, want := "ok", item.Want
		for _, record := range records {
			if record.Family != family || !strings.HasSuffix(item.Label, record.Label) || record.Source != item.Source {
				continue
			}
			if record.Go != item.Want {
				test.Fatalf("%s: recorded Go outcome changed", record.Label)
			}
			seen[record.Label] = true
			errText := record.PrettierError
			want = record.Prettier
			if embedded {
				want, errText = record.Embedded, record.EmbeddedError
			}
			if errText != "" {
				status, want = "error", errText
				refusals++
			} else {
				differences++
			}
			if want == "" {
				test.Fatalf("%s: missing %s outcome", record.Label, name)
			}
			break
		}
		expected := status + "\t" + escape.Replace(want)
		if lines[index] != expected {
			test.Errorf("%s %s: %s", name, item.Label, firstDifference(lines[index], expected))
		}
	}
	for _, record := range records {
		if census && record.Family == family && !seen[record.Label] {
			test.Errorf("%s: recorded upstream case absent: %s", name, record.Label)
		}
	}
	test.Logf(
		"%s: %d cases; %d exact upstream text differences, %d exact parser refusals, all other bytes match Go",
		name,
		len(cases),
		differences,
		refusals,
	)
}

func TestTSCCorpusUpstreamDifferences(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_TS_PRETTIER")
	if library == "" {
		test.Skip("set ADAMIC_TS_PRETTIER to prettier@3.9.6; #xq2ecw6 (setup --gate-inputs) installs it")
	}
	for _, family := range []string{"expressions", "statements"} {
		var cases []oracles.PrinterCase
		for _, record := range oracles.UpstreamOutcomes(test) {
			if record.Family == family {
				cases = append(cases, oracles.PrinterCase{Label: record.Label, Source: record.Source, Want: record.Go})
			}
		}
		data, _ := json.Marshal(cases)
		specs := filepath.Join(test.TempDir(), "cases.json")
		if err := os.WriteFile(specs, data, 0644); err != nil {
			test.Fatal(err)
		}
		npm, _ := filepath.Abs("testdata/expressions.mjs")
		comparePrinterLibrary(
			test,
			"npm Prettier",
			execute(test, nil, "node", npm, library, specs),
			specs,
			family,
			false,
		)
		fork, _ := filepath.Abs("testdata/embedded.mjs")
		bundles, _ := filepath.Abs(filepath.Join(oracles.Repository, "cohere/internal/format/prettier/bundles"))
		comparePrinterLibrary(
			test,
			"embedded Prettier",
			execute(test, nil, "node", fork, bundles, specs),
			specs,
			family,
			true,
		)
	}
}
