// Command cohere-release stages every published package for one version: a binary package per
// platform, and the dispatcher package that resolves between them.
//
// It exists as a Go command rather than a shell script so that the platform list, the npm `os` and
// `cpu` fields, the launcher, and the binary that gets built all read from one set of definitions
// in `internal/release`. The classic packaging bug is a matrix in CI that drifts from the manifest
// it publishes, producing a package that installs on the wrong platform or on none — and either one
// looks like an unsupported platform rather than like the mistake it is.
//
// It stages and never publishes. Publishing is `npm publish` against the staged directories, run
// deliberately, because an accidental publish cannot be taken back.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/system-inc/cohere/internal/release/packaging"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "cohere-release: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	version := flag.String("version", "", "the version to stamp and publish, like 1.0.0; anything below 1.0.0, or not semver, is refused")
	outputDirectory := flag.String("output", "dist", "where to stage the packages")
	moduleDirectory := flag.String("module", ".", "the root of the cohere module")
	only := flag.String("only", "", "build just these platforms, comma separated, as os/arch")
	signingIdentity := flag.String("signing-identity", "", "the macOS codesigning identity; unsigned when empty")
	keychainProfile := flag.String("notary-profile", "", "the notarytool keychain profile; not notarized when empty")
	swiftScratch := flag.String("swift-scratch", "", "where SwiftPM builds the Swift engine; the user cache directory when empty, never inside --output")
	flag.Parse()

	if *version == "" {
		// Refused rather than defaulted. A release stamped "dev" is a published artifact that
		// cannot say what it is, which is precisely what the version stamping exists to prevent.
		return fmt.Errorf("--version is required: a release that cannot name itself is not a release")
	}

	moduleRoot, err := filepath.Abs(*moduleDirectory)
	if err != nil {
		return fmt.Errorf("resolving the module directory: %w", err)
	}
	outputRoot, err := filepath.Abs(*outputDirectory)
	if err != nil {
		return fmt.Errorf("resolving the output directory: %w", err)
	}

	targets, err := selectTargets(*only)
	if err != nil {
		return err
	}

	fmt.Printf("staging cohere %s into %s\n", *version, outputRoot)

	signing := release.Signing{Identity: *signingIdentity, KeychainProfile: *keychainProfile}

	result, err := release.Build(release.Options{
		ModuleDirectory: moduleRoot,
		OutputDirectory: outputRoot,
		Version:         *version,
		Targets:         targets,
		Signing:         signing,

		SwiftScratchDirectory: *swiftScratch,
	})
	if err != nil {
		return err
	}

	printSummary(result, targets, signing)
	return nil
}

// selectTargets parses the --only filter, defaulting to every shipped platform.
//
// An unrecognized platform is an error rather than a silent no-op. Building nothing and reporting
// success is the same shape as every other defect this tool exists to prevent: a run that did
// nothing must never look like a run that had nothing to do.
func selectTargets(only string) ([]release.Target, error) {
	if strings.TrimSpace(only) == "" {
		return nil, nil
	}

	targets := []release.Target{}
	for _, requested := range strings.Split(only, ",") {
		requested = strings.TrimSpace(requested)
		if requested == "" {
			continue
		}

		found := false
		for _, target := range release.Targets {
			if target.String() == requested {
				targets = append(targets, target)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("%q is not a shipped platform; the shipped platforms are %s", requested, describeTargets(release.Targets))
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("--only matched no platforms")
	}
	return targets, nil
}

// describeTargets lists platforms the way they are written on the command line.
func describeTargets(targets []release.Target) string {
	names := make([]string, len(targets))
	for index, target := range targets {
		names[index] = target.String()
	}
	return strings.Join(names, ", ")
}

// printSummary reports what was staged, with sizes.
//
// The sizes print because they are the cheapest available check that the binaries are real. Six
// platform binaries within a megabyte of each other is a matrix that worked; one that is a fraction
// of its siblings is a build that failed into something that still packages perfectly.
//
// Every size says which build produced it. A stripped release binary and a plain `go build` of the
// same commit differ by roughly 18 MB — measured at 43.1 against 61.9 — so a bare megabyte figure
// is not a measurement, it is two possible measurements sharing a label. Two people quoting sizes
// from different builds spent real time reconciling numbers that were both correct.
func printSummary(result release.Result, requested []release.Target, signing release.Signing) {
	fmt.Println()
	for _, staged := range result.Packages {
		if staged.SizeInBytes == 0 {
			fmt.Printf("  %-32s %s\n", staged.Name, staged.Directory)
			continue
		}
		fmt.Printf("  %-32s %6.1f MB  %s\n", staged.Name, float64(staged.SizeInBytes)/(1<<20), staged.Directory)
	}
	fmt.Printf("\nsizes are stripped release builds: %s\n", release.DescribeBuild())

	platformCount := len(requested)
	if platformCount == 0 {
		platformCount = len(release.Targets)
	}

	fmt.Printf("\nstaged %d platform packages and the dispatcher.\n", platformCount)
	fmt.Println(release.DescribeSigningState(signing))

	if requested != nil {
		// A partial release published as if it were whole leaves the missing platforms resolving to
		// nothing, and the dispatcher's loud failure gets reported as a bug against a version that
		// looked fine everywhere else. Saying so here is the only place it can be caught.
		fmt.Printf("this is a partial build (--only): do not publish it as a release.\n")
	}
}
