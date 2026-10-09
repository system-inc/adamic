package release

import (
	"os/exec"
	"strings"
	"testing"
)

// `Targets` is the largest hand-written claim in this package: six platforms, asserted rather than
// derived. Every other test here iterates it, which means they cohere consistency *with* the list
// and never the list itself — they would all pass green if it named one platform, or named a
// platform Go cannot build.
//
// That was measured, not supposed. Adding `plan9/sparc64` to `Targets` passed the entire suite, and
// a release would then have failed at the first cross-compilation of a platform we told npm we
// support. A field nobody derives drifts toward whatever was easiest to write.

func TestEveryTargetIsAPlatformGoCanActuallyBuild(t *testing.T) {
	t.Parallel()

	supported := supportedPlatforms(t)

	for _, target := range Targets {
		if !supported[target.String()] {
			t.Errorf(
				"%s is in Targets and `go tool dist list` does not offer it, so a release would fail at its cross-compilation",
				target,
			)
		}
	}
}

func TestTargetsCoverTheThreeOperatingSystemsWeClaimToShip(t *testing.T) {
	t.Parallel()

	// The manifests tell npm we support darwin, linux, and win32, and the launcher resolves a
	// package per platform. A target list that quietly lost one would publish a dispatcher whose
	// optional dependency for that platform does not exist, and the failure lands on whoever
	// installs there rather than on the release.
	//
	// Both architectures per system, because an arm64-only macOS release installs on nothing for an
	// Intel machine, which reads to that user as an unsupported platform.
	required := map[string]int{"darwin": 0, "linux": 0, "windows": 0}
	for _, target := range Targets {
		if _, claimed := required[target.GoOperatingSystem]; claimed {
			required[target.GoOperatingSystem]++
		}
	}

	for operatingSystem, count := range required {
		if count < 2 {
			t.Errorf("%s has %d targets; both amd64 and arm64 are needed to cover the machines people run", operatingSystem, count)
		}
	}
}

func TestTargetsHasNoDuplicates(t *testing.T) {
	t.Parallel()

	// Two entries for one platform stage the same package twice, and the second write wins silently.
	// Whichever manifest lands last is the one published, so a duplicate with a typo'd architecture
	// would publish a package whose contents and `cpu` field disagree.
	seen := map[string]bool{}
	for _, target := range Targets {
		if seen[target.String()] {
			t.Errorf("%s appears twice in Targets", target)
		}
		seen[target.String()] = true
	}
}

// supportedPlatforms asks the toolchain which platforms it can build for.
//
// This is the whole point of the file: the answer comes from `go tool dist list` rather than from a
// second list written here, because a hand-written expectation would drift exactly the way the one
// it checks does. The control below is what makes the result trustworthy — a probe that cannot
// return "absent" would pass every target including impossible ones.
func supportedPlatforms(t *testing.T) map[string]bool {
	t.Helper()

	output, err := exec.Command("go", "tool", "dist", "list").Output()
	if err != nil {
		t.Fatalf("asking the toolchain which platforms it supports: %v", err)
	}

	supported := map[string]bool{}
	for _, line := range strings.Split(string(output), "\n") {
		if platform := strings.TrimSpace(line); platform != "" {
			supported[platform] = true
		}
	}

	// Prove the probe can distinguish both answers before trusting it. A `dist list` that returned
	// nothing, or that this parsed wrongly, would make every target look supported and the check
	// would pass while measuring nothing.
	if len(supported) < 10 {
		t.Fatalf("the toolchain reported only %d platforms, which cannot be right", len(supported))
	}
	if !supported["linux/amd64"] {
		t.Fatalf("linux/amd64 is missing from the platform list, so it was not parsed correctly")
	}
	if supported["plan9/sparc64"] {
		t.Fatalf("an impossible platform reported as supported, so this check cannot return absent")
	}

	return supported
}
