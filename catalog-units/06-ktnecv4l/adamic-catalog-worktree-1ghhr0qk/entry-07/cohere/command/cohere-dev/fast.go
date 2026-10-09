package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/corpus"
)

// The fast tier is the edit loop (#nxgt2ca, ruled by @system_cohere on Kirk's two-tier call):
//
//	go run ./command/cohere-dev test --fast
//
// It runs every package of the module except the ones landingGateOnly names, and Go's test cache decides
// what actually runs: a package whose inputs an edit did not touch answers from the cache, and one whose
// tests import the edited code or read the edited file runs. That is the measured dependency set, kept by
// Go, so nothing here maps edits to packages. The corpus walks are left out inside their package by the
// variable fastTierVariable names, which their helpers read.
//
// It takes no slot: it is the edit loop and has to stay instant. It prints the machine's load when it
// starts, so a fast run beside a held full gate is visible. It is never the landing gate, and says so last.

// fastTierVariable tells a package's tests that the fast tier is running them, so a test the landing gate
// alone runs can skip itself. The corpus walks in high_level_intermediate_representation read it.
const fastTierVariable = "COHERE_FAST_TIER"

// fastBudgetCPUSeconds is what a package, or part of one, must cost in CPU, user plus system, measured alone
// with -count=1 and its test binary already built, to join landingGateOnly. Anything at or under it stays
// in the edit loop. TestEveryLandingGateOnlyEntryStillEarnsItsPlace holds every entry's recorded cost above
// it.
//
// CPU and not wall, because wall moves with the machine's load and CPU does not (#1qbez1f). With a 2s wall
// budget beside it, the registry's ESLint corpus passed two gates at 2.5s and 3.5s under load and failed the
// third at 1.1s on a quieter machine, while its CPU read 10.5 to 12.7s throughout. A gate whose verdict
// depends on who else is running is a flaky test.
//
// CPU is steadier than wall, not load-invariant. A heavily parallel package's CPU inflates under contention:
// command/cohere read 926 to 933s of CPU at load 105 to 166 against 49.4s at load about 10, and
// crosscompile 36 to 41s against 26. Every change load made was upward, so a reading at low load is the
// low one.
const fastBudgetCPUSeconds = 15

// fastEvictionCPUSeconds is what an entry, measured again now, must fall to before
// TestEveryLandingGateOnlyEntryStillEarnsItsPlace sends it back to the edit loop: a third under the budget
// to join, because a live measurement scatters and one threshold judged both ways flips on the scatter.
// hir's corpus walks, a difference of two runs, read 14.9s in one of build's land runs, 18.5s at load 10
// and 24.8s at load 13 to 50, all on 2026-10-05, and failed build's land at 14.9 against a single bar of
// 15. Under 10 is a change in what the tests do, not in who else is running.
const fastEvictionCPUSeconds = 10

// landingGateEntry is one thing the fast tier leaves to the full gate.
type landingGateEntry struct {
	// pattern is the package, as `go test` names it from the module root.
	pattern string

	// skippedBy is the variable its tests read to leave themselves out, for an entry that is part of a
	// package rather than all of it; empty for a whole package.
	skippedBy string

	// reads is the corpus its expensive tests read, for an entry whose cost is that corpus (internal/corpus);
	// the zero Corpus for one that reads none. Without the corpus those tests skip and the entry costs
	// nothing, which would read as an entry that got cheap.
	reads corpus.Corpus

	// wallSeconds and cpuSeconds are its cost measured alone with -count=1 on a quiet machine, when it was
	// added; for part of a package, the package run in full less the package run in the fast tier.
	// TestEveryLandingGateOnlyEntryStillEarnsItsPlace measures both again and judges the CPU; the wall is
	// what the fast tier prints, for a reader, and judges nothing.
	wallSeconds, cpuSeconds float64

	reason string
}

// landingGateOnly is everything the fast tier leaves to the full gate, each measured on 2026-10-05 as its
// entry says.
var landingGateOnly = []landingGateEntry{
	{
		// At load about 10.
		pattern:     "./command/cohere",
		wallSeconds: 9.2, cpuSeconds: 49.4,
		reason: "it imports every package, so every code edit reruns it, and about 260 of its runs are the real " +
			"binary type-checking a project",
	},
	{
		// At load about 10.
		pattern:     "./internal/release/crosscompile",
		wallSeconds: 2.4, cpuSeconds: 26.0,
		reason: "it type-checks all of cohere for each of the six release targets",
	},
	{
		pattern:   "./internal/lint/ecmascript/high_level_intermediate_representation",
		skippedBy: fastTierVariable,
		reads:     corpus.Structure,
		// At load about 10: the package in full took 12.8s and 22.3s of CPU, in the fast tier 1.8s and 3.8s.
		wallSeconds: 11.0, cpuSeconds: 18.5,
		reason: "the corpus walks lower hundreds of real files, again in each of about 45 tests; the package's " +
			"other tests stay in the fast tier",
	},
}

// fastTest runs the fast tier in the environment given and returns go test's exit code.
func fastTest(base []string, arguments []string) int {
	fmt.Fprintf(os.Stderr, "cohere-dev: fast tier, at load %s\n", machineLoad())

	packages, err := modulePackages()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 1
	}
	skippedPackages := map[string]bool{}
	for _, entry := range landingGateOnly {
		if !packages[entry.pattern] {
			fmt.Fprintf(os.Stderr, "cohere-dev: the fast tier's landing-gate list names %s, which is not a package of this "+
				"module; fix landingGateOnly in command/cohere-dev/fast.go\n", entry.pattern)
			return 2
		}
		if entry.skippedBy == "" {
			skippedPackages[entry.pattern] = true
		}
	}

	selected := []string{}
	for pattern := range packages {
		if !skippedPackages[pattern] {
			selected = append(selected, pattern)
		}
	}
	sort.Strings(selected)

	environment, uncovered, err := testEnvironment(base)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cohere-dev: %v\n", err)
		return 2
	}
	command := exec.Command("go", append(append([]string{"test"}, arguments...), selected...)...)
	command.Env = append(environment, fastTierVariable+"=1")
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	code := 0
	if err := command.Run(); err != nil {
		exited, ok := err.(*exec.ExitError)
		if !ok {
			fmt.Fprintf(os.Stderr, "cohere-dev: running go test: %v\n", err)
			return 1
		}
		code = exited.ExitCode()
	}

	fmt.Fprintln(os.Stderr, "cohere-dev: the fast tier left these to the landing gate:")
	for _, entry := range landingGateOnly {
		what := entry.pattern
		if entry.skippedBy != "" {
			what += "'s corpus walks"
		}
		fmt.Fprintf(os.Stderr, "  %s (%.1fs alone): %s\n", what, entry.wallSeconds, entry.reason)
	}
	printUncovered(os.Stderr, uncovered)
	fmt.Fprintln(os.Stderr, "fast tier: not the landing gate. Run cohere-dev test ./... through the lock before landing on main.")
	return code
}

// modulePackages is every package of the module, named as `go test` names it from the module root.
//
// The root module's only: the fast tier is not the landing gate, so it leaves the other modules go.work
// names (the shims, static_single_assignment) to the gate and to `cohere-dev test ./...` (workspace.go).
func modulePackages() (map[string]bool, error) {
	return modulePackagesIn("")
}

// modulePackagesIn is modulePackages for the module at root, or the working directory's when it is empty.
func modulePackagesIn(root string) (map[string]bool, error) {
	command := exec.Command("go", "list", "-f", "{{.ImportPath}} {{.Module.Path}}", "./...")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("listing the module's packages: %w", err)
	}
	packages := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		relative := strings.TrimPrefix(fields[0], fields[1])
		packages["."+relative] = true
	}
	return packages, nil
}

// machineLoad is the one-minute load average, or "unknown" where it cannot be read.
func machineLoad() string {
	output, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		if contents, readError := os.ReadFile("/proc/loadavg"); readError == nil {
			if fields := strings.Fields(string(contents)); len(fields) > 0 {
				return fields[0]
			}
		}
		return "unknown"
	}
	// "{ 3.12 2.90 2.71 }"
	if fields := strings.Fields(strings.Trim(strings.TrimSpace(string(output)), "{}")); len(fields) > 0 {
		return fields[0]
	}
	return "unknown"
}
