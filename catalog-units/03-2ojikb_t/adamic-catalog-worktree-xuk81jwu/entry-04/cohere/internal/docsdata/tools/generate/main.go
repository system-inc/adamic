// Command generate writes the website's data (internal/docsdata) into the module's docs/data/ directory.
// Run it from the module root:
//
//	go run ./internal/docsdata/tools/generate                   write the files, examples read back
//	go run ./internal/docsdata/tools/generate -capture          write the files, examples recaptured
//	go run ./internal/docsdata/tools/generate -check            exit 1 if any is stale, examples recaptured
//	go run ./internal/docsdata/tools/generate -check-rendering  exit 1 if a file rendered from the
//	                                                            committed examples is stale
//
// cli.json is read from the binary's own help, so every run builds ./command/cohere into a temporary
// directory. To recapture is to rerun the lint tests with COHERE_DOCS_CAPTURE set, which takes minutes.
//
// -check always recaptures, because a check computes from source: one that read the committed examples
// back compared them with themselves and could never find one stale (#d7vx76v). It belongs where minutes
// are fine, the release's module check and CI, never the per-edit test run, whose cache a fresh capture
// directory would bust for all of internal/lint/... The examples come back from the committed files only
// in the modes that say so: a write without -capture, and -check-rendering, which covers every file
// rendered from them and says on every run that the examples themselves were not recaptured.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/cohere/internal/docsdata"
	"github.com/system-inc/cohere/internal/docsdata/capture"
)

// capturedPackages are the tests that assert rule cases through the harness.
const capturedPackages = "./internal/lint/..."

// verbs are the subcommands with a flag set of their own.
var verbs = []string{"rename"}

// notRecaptured is what a run that read the examples back says, every time, so nothing it prints reads as
// a verdict on the examples themselves.
const notRecaptured = "the examples were read back from " + docsdata.ExamplesDirectory + ", not recaptured from the lint tests; -check recaptures them"

// sources is where a run reads what Build takes beyond the module's files: the binary's help and the lint
// tests' captured cases. A test swaps both, so it can run a check without building the binary or the suite.
type sources struct {
	root string
	// help returns what the binary's flag sets print. capture runs the lint tests with their cases
	// recorded into directory.
	help    func(scratch string) ([]byte, map[string][]byte, error)
	capture func(directory string) error
}

var moduleSources = sources{root: ".", help: binaryHelp, capture: captureLintTests}

// errStale is a check that found stale files, which it has already named: exit 1 with nothing more to say.
var errStale = errors.New("stale")

func main() {
	check := flag.Bool("check", false, "recapture the examples from the lint tests, then report every stale file and exit 1, writing nothing (minutes)")
	checkRendering := flag.Bool("check-rendering", false, "report every stale file rendered from the committed examples and exit 1, writing nothing; the examples are not recaptured")
	recapture := flag.Bool("capture", false, "rerun the lint tests with "+capture.Variable+" set and pick the examples afresh")
	flag.Parse()

	if err := run(*check, *checkRendering, *recapture); err != nil {
		if !errors.Is(err, errStale) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func run(check bool, checkRendering bool, recapture bool) error {
	if check && checkRendering {
		return fmt.Errorf("-check and -check-rendering are two depths of one check; name one")
	}
	if checkRendering && recapture {
		return fmt.Errorf("-check-rendering reads the committed examples back; to recapture them, run -check")
	}
	scratch, err := os.MkdirTemp("", "cohere-docsdata-")
	if err != nil {
		return err
	}
	defer removeScratch(scratch)

	if check || checkRendering {
		stale, err := staleFiles(moduleSources, scratch, check)
		if err != nil {
			return err
		}
		for _, path := range stale {
			fmt.Printf("stale: %s\n", path)
		}
		if checkRendering {
			fmt.Println(notRecaptured)
		}
		if len(stale) > 0 {
			// Returned rather than exiting here, so the scratch directory is still removed.
			return errStale
		}
		if check {
			fmt.Println("every file is current, the examples recaptured from the lint tests")
		}
		return nil
	}

	inputs, err := gather(moduleSources, scratch, recapture)
	if err != nil {
		return err
	}
	files, err := docsdata.Build(inputs)
	if err != nil {
		return err
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		existing, _ := os.ReadFile(path)
		if bytes.Equal(existing, files[path]) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, files[path], 0o644); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", path)
	}
	orphans, err := docsdata.OrphanedExamples(".", files)
	if err != nil {
		return err
	}
	for _, path := range orphans {
		if err := os.Remove(path); err != nil {
			return err
		}
		fmt.Printf("removed %s\n", path)
	}
	if !recapture {
		fmt.Println(notRecaptured)
	}
	return nil
}

// staleFiles is a check: every file under source.root that differs from what Build produces, plus every
// example file Build no longer produces. With recapture, the examples are the lint tests' own cases, so a
// committed example that no longer matches its rule's tests is stale; without it, they are the committed
// files, and only what is rendered from them is checked.
func staleFiles(source sources, scratch string, recapture bool) ([]string, error) {
	inputs, err := gather(source, scratch, recapture)
	if err != nil {
		return nil, err
	}
	files, err := docsdata.Build(inputs)
	if err != nil {
		return nil, err
	}
	return docsdata.Stale(source.root, files)
}

// gather reads every source Build takes. Without recapture, the examples are the committed files.
func gather(source sources, scratch string, recapture bool) (docsdata.Inputs, error) {
	inputs, err := docsdata.SourceInputs(source.root)
	if err != nil {
		return inputs, err
	}
	if inputs.Help, inputs.VerbHelp, err = source.help(scratch); err != nil {
		return inputs, err
	}
	if !recapture {
		return inputs, nil
	}

	captured := filepath.Join(scratch, "captured")
	if err := source.capture(captured); err != nil {
		return inputs, err
	}
	records, err := docsdata.ReadCaptures(captured)
	if err != nil {
		return inputs, err
	}
	if len(records) == 0 {
		// A capture that recorded nothing would leave every example orphaned, and a check reading that as a
		// verdict would be checking nothing.
		return inputs, fmt.Errorf("the lint tests ran under capture and recorded no case")
	}
	var withOptions int
	inputs.Examples, withOptions = docsdata.PickExamples(records)
	fmt.Printf("captured %d records; %d distinct cases ran with options, counted for message ids and never shown\n", len(records), withOptions)
	return inputs, nil
}

// captureLintTests runs the lint tests with each asserted case recorded into directory.
//
// No -count=1, as no gate run passes it: Go keys a cached test result on the environment the test reads,
// and the harness reads COHERE_DOCS_CAPTURE, which names a fresh temporary directory every run. So each
// package that records misses the cache and reruns, and a package that never reads the variable records
// nothing and may stay cached.
func captureLintTests(directory string) error {
	test := exec.Command("go", "test", capturedPackages)
	test.Env = append(os.Environ(), capture.Variable+"="+directory)
	test.Stdout, test.Stderr = os.Stderr, os.Stderr
	if err := test.Run(); err != nil {
		return fmt.Errorf("the lint tests failed under capture, so their cases are not all asserted: %w", err)
	}
	return nil
}

// binaryHelp builds the cohere binary and returns what its flag sets print: the top level's, and each
// verb's.
func binaryHelp(scratch string) ([]byte, map[string][]byte, error) {
	binary := filepath.Join(scratch, "cohere")
	build := exec.Command("go", "build", "-o", binary, "./command/cohere")
	build.Stdout, build.Stderr = os.Stderr, os.Stderr
	if err := build.Run(); err != nil {
		return nil, nil, fmt.Errorf("building ./command/cohere for its help: %w", err)
	}

	help, err := helpOutput(binary, "-help")
	if err != nil {
		return nil, nil, err
	}
	verbHelp := map[string][]byte{}
	for _, verb := range verbs {
		if verbHelp[verb], err = helpOutput(binary, verb, "-help"); err != nil {
			return nil, nil, err
		}
	}
	return help, verbHelp, nil
}

// helpOutput runs the binary with arguments ending in -help and returns the help it printed, without the
// flag package's `Usage of <path>:` line, which names the temporary binary, or the `cohere: flag: help
// requested` a verb's flag set adds when it returns the request as an error.
func helpOutput(binary string, arguments ...string) ([]byte, error) {
	command := exec.Command(binary, arguments...)
	command.Dir = filepath.Dir(binary)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := command.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return nil, err
		}
	}
	var kept []string
	for _, line := range strings.Split(output.String(), "\n") {
		if strings.HasPrefix(line, "Usage of ") || line == "cohere: flag: help requested" {
			continue
		}
		kept = append(kept, line)
	}
	return []byte(strings.Join(kept, "\n")), nil
}

// removeScratch removes the temporary directory's files one by one, by name, and then the directories.
func removeScratch(scratch string) {
	var directories []string
	_ = filepath.WalkDir(scratch, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			directories = append(directories, path)
			return nil
		}
		return os.Remove(path)
	})
	for index := len(directories) - 1; index >= 0; index-- {
		_ = os.Remove(directories[index])
	}
}
