package release

import (
	"encoding/json"
	"strings"
	"testing"
)

// The defect these guard is quiet by construction. A manifest whose `os` or `cpu` disagrees with
// the binary beside it installs on the wrong machine or on none, and either way what a user sees is
// "no binary for this platform" — the same message a genuinely unsupported platform produces. So
// the fields are asserted against the same target that builds the binary, which is the only way the
// two cannot drift.

func TestPlatformManifestDeclaresTheOperatingSystemAndArchitectureNpmResolvesBy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		target                  Target
		expectedName            string
		expectedOperatingSystem string
		expectedArchitecture    string
	}{
		{Target{"darwin", "arm64"}, "@system-inc/cohere-darwin-arm64", "darwin", "arm64"},
		{Target{"darwin", "amd64"}, "@system-inc/cohere-darwin-x64", "darwin", "x64"},
		{Target{"linux", "amd64"}, "@system-inc/cohere-linux-x64", "linux", "x64"},
		{Target{"windows", "amd64"}, "@system-inc/cohere-win32-x64", "win32", "x64"},
		{Target{"windows", "arm64"}, "@system-inc/cohere-win32-arm64", "win32", "arm64"},
	}

	for _, testCase := range cases {
		manifest := decodeManifest(t, mustPlatformManifest(t, testCase.target, "1.2.3"))

		if manifest["name"] != testCase.expectedName {
			t.Errorf("%s published as %v, wanted %s", testCase.target, manifest["name"], testCase.expectedName)
		}
		if got := singleString(t, manifest["os"]); got != testCase.expectedOperatingSystem {
			t.Errorf("%s declared os %q, wanted %q — npm would install it on the wrong machine", testCase.target, got, testCase.expectedOperatingSystem)
		}
		if got := singleString(t, manifest["cpu"]); got != testCase.expectedArchitecture {
			t.Errorf("%s declared cpu %q, wanted %q — npm would install it on the wrong machine", testCase.target, got, testCase.expectedArchitecture)
		}
	}
}

func TestPlatformManifestMatchesWhatTheLauncherLooksFor(t *testing.T) {
	t.Parallel()

	// The launcher builds the package name from Node's process.platform and process.arch, which are
	// npm's spelling. This asserts the published name is the one it will ask for, for every target
	// we ship. A mismatch here is a package that publishes, installs, and is never found.
	for _, target := range Targets {
		manifest := decodeManifest(t, mustPlatformManifest(t, target, "1.2.3"))

		expected := PlatformPackagePrefix + singleString(t, manifest["os"]) + "-" + singleString(t, manifest["cpu"])
		if manifest["name"] != expected {
			t.Errorf(
				"%s publishes as %v, but the launcher on that machine asks for %s",
				target, manifest["name"], expected,
			)
		}
	}
}

func TestPlatformManifestClaimsNoCommand(t *testing.T) {
	t.Parallel()

	// Only the dispatcher owns the `cohere` name. A platform package with a `bin` field would race
	// it for the same link in node_modules/.bin, and which one wins depends on install order.
	for _, target := range Targets {
		manifest := decodeManifest(t, mustPlatformManifest(t, target, "1.2.3"))
		if _, claimed := manifest["bin"]; claimed {
			t.Errorf("%s declares a bin field, which would fight the dispatcher for node_modules/.bin/cohere", target)
		}
	}
}

func TestDispatcherDependsOnEveryPlatformExactly(t *testing.T) {
	t.Parallel()

	manifest := decodeManifest(t, mustDispatcherManifest(t, "1.2.3"))

	optional, ok := manifest["optionalDependencies"].(map[string]any)
	if !ok {
		t.Fatalf("the dispatcher declares no optionalDependencies, so no binary would ever install")
	}

	for _, target := range Targets {
		pinned, listed := optional[target.PackageName()]
		if !listed {
			t.Errorf("%s is built but never depended on, so it would never install", target.PackageName())
			continue
		}
		// A range would let a resolver pair this dispatcher with a binary built from other rules.
		if pinned != "1.2.3" {
			t.Errorf("%s pinned as %v rather than the exact version", target.PackageName(), pinned)
		}
	}

	if len(optional) != len(Targets) {
		t.Errorf("the dispatcher depends on %d platform packages but %d are built", len(optional), len(Targets))
	}
}

// The published name is a contract with code outside this repository, so it is pinned as a literal
// rather than compared against the constant that produces it.
//
// Structure tells a consumer whose gate found no binary to run `pnpm add -D @system-inc/cohere`. A
// constant renamed here alone would still pass every other test, publish cleanly, and leave that
// instruction naming a package that does not exist, which is how this name came to be decided: the
// error text and the registry disagreed, and the error text is what people read.
func TestDispatcherPublishesUnderTheNameStructureTellsPeopleToInstall(t *testing.T) {
	t.Parallel()

	manifest := decodeManifest(t, mustDispatcherManifest(t, "1.2.3"))
	if manifest["name"] != "@system-inc/cohere" {
		t.Errorf("the dispatcher publishes as %v, but Structure tells people to install @system-inc/cohere", manifest["name"])
	}

	// Every platform package lives in the same scope, so one scope's access settings govern the
	// whole release and no binary publishes somewhere the dispatcher's owner does not control.
	for _, target := range Targets {
		if !strings.HasPrefix(target.PackageName(), PackageScope+"/") {
			t.Errorf("%s publishes outside %s", target.PackageName(), PackageScope)
		}
	}
}

// Both command names must be installed, and both must reach the same launcher.
//
// Two entries rather than one because `v` is what gets typed on the loop and `cohere` is what reads
// in a script, and because a single letter is short enough to collide with something already on a
// machine. A consumer who hits that keeps a working command instead of a broken install.
//
// Asserted per name rather than by counting. An earlier version of this test checked only `cohere`,
// which would have passed unchanged if `v` had silently stopped being declared: the guarantee it
// stated was true and the guarantee anyone reading it believed was not.
func TestDispatcherOwnsBothCommandNames(t *testing.T) {
	t.Parallel()

	manifest := decodeManifest(t, mustDispatcherManifest(t, "1.2.3"))

	binaries, ok := manifest["bin"].(map[string]any)
	if !ok {
		t.Fatalf("the dispatcher declares no bin, so installing it would create no command at all")
	}

	for _, name := range []string{ShortCommandName, FullCommandName} {
		if binaries[name] != launcherRelativePath {
			t.Errorf("the %q command points at %v, wanted %s", name, binaries[name], launcherRelativePath)
		}
	}

	if len(binaries) != 2 {
		// A third name would be a command nobody documented, installed on every consumer's machine.
		t.Errorf("the dispatcher declares %d commands, wanted exactly the two named ones: %v",
			len(binaries), binaries)
	}
}

func TestLauncherHoldsTheNoFallbackRule(t *testing.T) {
	t.Parallel()

	// The launcher is generated as a string, so a compiler cannot check it. These assert the
	// substance rather than the wording: that it exits non-zero on a missing binary, and that no
	// bare command name appears as something it could run instead.
	launcher := DispatcherLauncher()

	for _, placeholder := range []string{"__PLATFORM_PACKAGE_PREFIX__", "__PACKAGE_SCOPE__", "__OVERRIDE_VARIABLE__", "__CHECKSUMS_FILE_NAME__"} {
		if strings.Contains(launcher, placeholder) {
			t.Fatalf("the launcher shipped with %s unreplaced, so it would resolve or check nothing", placeholder)
		}
	}
	if !strings.Contains(launcher, "'"+PlatformPackagePrefix+"'") {
		t.Errorf("the launcher never names the %s prefix, so it cannot find a platform package", PlatformPackagePrefix)
	}
	if !strings.Contains(launcher, BinaryOverrideVariable) {
		t.Errorf("the launcher never reads %s, so a local build cannot be pointed at", BinaryOverrideVariable)
	}
	if !strings.Contains(launcher, "process.exit(1)") {
		t.Errorf("the launcher has no non-zero exit, so a missing binary could report success")
	}

	// The historical defect, spelled out: resolving to a bare name and letting the shell find
	// something. If this ever appears, the launcher can spawn a stranger.
	for _, forbidden := range []string{"spawnSync('cohere'", `spawnSync("cohere"`, "PATH"} {
		if strings.Contains(launcher, forbidden) {
			t.Errorf("the launcher contains %q, which is the fallback-to-PATH bug this rule forbids", forbidden)
		}
	}
}

func TestLauncherDoesNotReportSuccessForASignalledRun(t *testing.T) {
	t.Parallel()

	// spawnSync reports a null status when the child is killed by a signal, and null is falsy — so
	// the obvious implementation exits zero for a run that was terminated. That is a green gate over
	// a killed process.
	launcher := DispatcherLauncher()
	if !strings.Contains(launcher, "result.status === null") {
		t.Errorf("the launcher never checks for a null status, so a signalled run would exit zero")
	}
}

func mustPlatformManifest(t *testing.T, target Target, version string) []byte {
	t.Helper()
	encoded, err := PlatformManifest(target, version)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func mustDispatcherManifest(t *testing.T, version string) []byte {
	t.Helper()
	encoded, err := DispatcherManifest(version)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

// decodeManifest parses generated JSON, failing the test if it is not valid.
func decodeManifest(t *testing.T, encoded []byte) map[string]any {
	t.Helper()
	manifest := map[string]any{}
	if err := json.Unmarshal(encoded, &manifest); err != nil {
		t.Fatalf("the generated manifest is not valid JSON, so npm would reject the package: %v", err)
	}
	return manifest
}

// singleString reads a JSON array that should hold exactly one string.
func singleString(t *testing.T, value any) string {
	t.Helper()
	values, ok := value.([]any)
	if !ok || len(values) != 1 {
		t.Fatalf("expected a one-element array, got %#v", value)
	}
	text, ok := values[0].(string)
	if !ok {
		t.Fatalf("expected a string, got %#v", values[0])
	}
	return text
}
