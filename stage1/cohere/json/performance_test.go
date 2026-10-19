package json

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// The profiled -O2 -g binaries must answer the full corpus, not just the profile sample.
func TestProfileSnapshotsAgree(t *testing.T) {
	t.Parallel()
	snapshots := os.Getenv("ADAMIC_JSON_PROFILE_BINARIES")
	if snapshots == "" {
		t.Skip("set ADAMIC_JSON_PROFILE_BINARIES to profile snapshot binaries")
	}
	cases := corpusCases(t)
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	for _, binary := range filepath.SplitList(snapshots) {
		compare(t, binary, execute(t, nil, binary, "--cases", path), expected, cases)
	}
	t.Logf("%d profiled snapshots agree on %d texts, %d answer bytes", len(filepath.SplitList(snapshots)), len(cases), len(expected))
}

// The zero-cached-width mutant, which TestCachedWidthMutantIsCaught runs natively and on Node.
var cachedWidthMutant = printerMutation{"zero cached width", "doc.ts", "width: stringWidth(text)", "width: 0"}

// The mutant's native release build is a keyed product, built ahead by this test and nothing else, so the test that
// runs it compiles no C (#tze35sy, Loom's build law): built inside TestCachedWidthMutantIsCaught it was 33 to 40 s of
// json#1/23's time on Cloud and Home, Oct 18.
//
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONCachedWidthMutantRelease(t *testing.T) {
	buildcache.RequireArtifacts(t, cachedWidthMutantRelease(t), buildcache.Artifact{Name: "release"})
}

// cachedWidthMutantRelease is the mutant port lowered in this process and built by clang without sanitizers: keyed,
// as json's lowered port is, by the compiler and the port's sources, with the mutation, the native options and clang.
func cachedWidthMutantRelease(t *testing.T) string {
	t.Helper()
	options := native.Options{}
	inputs := jsonLoweredPortInputs()
	inputs.Name = "json cached-width mutant native release"
	inputs.Files = append(inputs.Files, "stage1/cohere/json/performance_test.go")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%s:%q=>%q", cachedWidthMutant.file, cachedWidthMutant.from, cachedWidthMutant.to), fmt.Sprintf("options=%+v", options))
	inputs.Flags = append(append(inputs.Flags, native.Flags(options)...), jsonBuildEnvironment("ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return buildcache.Product(t, inputs, func(directory string) error {
		source := filepath.Join(directory, "source")
		if err := os.Mkdir(source, 0755); err != nil {
			return err
		}
		mutation := cachedWidthMutant
		if err := buildJSONLoweredPortMutation(source, &mutation); err != nil {
			return err
		}
		c, err := os.ReadFile(filepath.Join(source, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(c), filepath.Join(directory, "release"), options)
	})
}

// Not parallel: on a machine whose cache lacks the product, cachedWidthMutantRelease builds it natively, and
// native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestCachedWidthMutantIsCaught(t *testing.T) {
	cases := []textCase{{"probe.json", `["` + strings.Repeat("wide", 25) + `",1]`}}
	answers, _ := cohereAnswers(t, cases, false)
	input, expected := protocol(cases, answers)
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	mutation := cachedWidthMutant
	entry := filepath.Join(portDirectory(t, &mutation), "main.ts")
	binary := filepath.Join(cachedWidthMutantRelease(t), "release")
	for _, side := range []struct {
		name   string
		result run
	}{
		{"native release", execute(t, nil, binary, "--cases", path)},
		{"Node", onNode(t, entry, "--cases", path)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s mutant must execute successfully: %+v", side.name, side.result)
		}
		if string(side.result.stdout) == expected {
			t.Fatalf("%s failed to catch zero cached width", side.name)
		}
		t.Logf("%s byte comparison caught zero cached width", side.name)
	}
}
