package json

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The profiled -O2 -g binaries must answer the full corpus, not just the profile sample.
func TestProfileSnapshotsAgree(test *testing.T) {
	test.Parallel()
	snapshots := os.Getenv("ADAMIC_JSON_PROFILE_BINARIES")
	if snapshots == "" {
		test.Skip("set ADAMIC_JSON_PROFILE_BINARIES to profile snapshot binaries")
	}
	cases := corpusCases(test)
	answers, _ := cohereAnswers(test, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		test.Fatal(err)
	}
	for _, binary := range filepath.SplitList(snapshots) {
		compare(test, binary, execute(test, nil, binary, "--cases", path), expected, cases)
	}
	test.Logf(
		"%d profiled snapshots agree on %d texts, %d answer bytes",
		len(filepath.SplitList(snapshots)),
		len(cases),
		len(expected),
	)
}

// Not parallel: where its product is missing, mutantPortProduct builds it, and native.Build writes the shared user
// cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestCachedWidthMutantIsCaught(test *testing.T) {
	cases := []textCase{{"probe.json", `["` + strings.Repeat("wide", 25) + `",1]`}}
	answers, _ := cohereAnswers(test, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(test.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		test.Fatal(err)
	}
	mutation := cachedWidthMutant
	entry := filepath.Join(portDirectory(test, &mutation), "main.ts")
	binary := filepath.Join(mutantPortProduct(test, cachedWidthMutant, native.Options{}), "port")
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native release", execute(test, nil, binary, "--cases", path)},
		{"Node", onNode(test, entry, "--cases", path)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			test.Fatalf("%s mutant must execute successfully: %+v", side.name, side.result)
		}
		if string(side.result.stdout) == expected {
			test.Fatalf("%s failed to catch zero cached width", side.name)
		}
		test.Logf("%s byte comparison caught zero cached width", side.name)
	}
}
