// Command vendor_fixtures re-pulls React's error-named compiler fixtures at a chosen commit.
//
// The corpus under `internal/react_conformance/testdata/fixtures` is vendored rather than fetched at
// test time, so this tool is what makes the pin reproducible: it is how the vendored tree was
// produced and how it gets moved to a new sha. It writes nothing unless every count it expects
// holds, because the whole value of a vendored corpus is that a human can review the diff, and a
// diff produced by a partial fetch is a review of the wrong thing.
//
// # The counting trap this tool is built around
//
// The research pass this corpus comes from reported 221 error fixtures. The real number is 325. The
// cause was a grep anchored at the fixture root, which excluded the nine subdirectories that hold
// 103 of them — wrong by a third, and it read exactly like a clean result. It was caught only
// because a parallel investigation returned a different number.
//
// So this tool does not filter by path prefix at any point. It walks the full tree, keys fixtures
// by their path relative to the fixture root, and reports the top-level and nested split separately
// so the number that moved last time is visible rather than summed away.
//
// # Why the git-tree API and not the contents API
//
// The GitHub contents API caps at 1,000 entries and silently repeats pages rather than erroring,
// which produced a plausible-but-wrong file count during the research pass. The git-tree API with
// `recursive=1` reports `truncated`, and this tool refuses to proceed when it is true. That check
// is the difference between a measurement and a guess.
//
//	go run ./internal/lint/rules/react/tools/vendor_fixtures -sha <commit> -out internal/react_conformance/testdata/fixtures
//
// Pass -dry-run to report the counts without writing.
package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// upstreamFixtureDirectory is the fixture root inside facebook/react.
const upstreamFixtureDirectory = "compiler/packages/babel-plugin-react-compiler/src/__tests__/fixtures/compiler"

// expectedFixtureCount is what a correct run finds at the pinned sha.
//
// Asserted rather than reported, because a vendoring tool that writes whatever it found cannot tell
// a corpus that shrank from a fetch that failed halfway. When bumping the pin, expect this to need
// updating and read the diff before changing it.
const expectedFixtureCount = 325

// expectedCleanFixtureCount is how many fixtures carrying `preserveMemoizationPragma` expect no
// error at all, at the pinned sha.
//
// It is asserted exactly rather than as "more than zero", and that is the whole point of the number.
// A corpus of error fixtures cannot measure false positives: over-reporting produces more findings
// and reads as success. The clean fixtures are the only population where silence is the right answer
// and firing is a defect, so a run that quietly collects none restores exactly the blind spot this
// change exists to close -- and it would report success while doing it.
//
// The exact count also guards a failure a nonzero check cannot see. Collecting fixtures by copying
// them into one directory loses any whose basenames collide, and four of these do:
// `optional-member-expression-as-memo-dep.js` and three siblings appear in more than one
// subdirectory. That collection lost four, returned 66, and errored on nothing. A guard asserting
// "some clean fixtures exist" passes on 66 as happily as on 70.
//
// Changing this number is meant to be a deliberate act with the diff read, not a threshold nudged
// until a run goes green.
const expectedCleanFixtureCount = 70

// preserveMemoizationPragma marks the fixtures that exercise manual-memoization preservation.
//
// This is the one non-error population vendored, rather than upstream's whole 1,486, because a
// corpus is only worth vendoring if a human can review the diff that lands it. These are the
// fixtures `preserve-manual-memoization` is scored against; the rest are other rules' business and
// would be vendored when someone needs them for the same stated reason.
const preserveMemoizationPragma = "validatePreserveExistingMemoizationGuarantees"

func main() {
	sha := flag.String("sha", "", "facebook/react commit to vendor from (required)")
	out := flag.String("out", "", "directory to write the fixture pairs into (required unless -dry-run)")
	dryRun := flag.Bool("dry-run", false, "report the counts without writing")
	expect := flag.Int("expect", expectedFixtureCount, "fixture pairs a correct run finds; a mismatch aborts")
	expectClean := flag.Int("expect-clean", expectedCleanFixtureCount, "clean fixture pairs a correct run finds; a mismatch aborts")
	flag.Parse()

	if *sha == "" {
		fail("-sha is required; vendoring from a moving branch is the thing this corpus exists not to do")
	}
	if *out == "" && !*dryRun {
		fail("-out is required unless -dry-run")
	}

	names, err := errorFixtureNames(*sha)
	if err != nil {
		fail("listing the tree: %v", err)
	}

	topLevel, nested := 0, 0
	subdirectories := map[string]int{}
	for _, name := range names {
		if !strings.Contains(name, "/") {
			topLevel++
			continue
		}
		nested++
		subdirectories[path.Dir(name)]++
	}

	fmt.Printf("sha              %s\n", *sha)
	fmt.Printf("error fixtures   %d  (%d top-level, %d across %d subdirectories)\n",
		len(names), topLevel, nested, len(subdirectories))
	for _, directory := range sortedKeys(subdirectories) {
		fmt.Printf("                   %-40s %d\n", directory, subdirectories[directory])
	}

	// A run that found nothing nested is the exact shape of the depth-anchored defect, and it would
	// otherwise write a corpus two thirds the size and report success.
	if nested == 0 {
		fail("found no fixtures in subdirectories at all, which is the signature of a path-anchored filter rather than an empty tree")
	}
	if len(names) != *expect {
		fail("found %d error fixtures, expected %d; if upstream really changed, read the diff and pass -expect", len(names), *expect)
	}

	clean, err := cleanFixtureNames(*sha)
	if err != nil {
		fail("listing the clean fixtures: %v", err)
	}
	fmt.Printf("clean fixtures   %d  (carry %s and expect no error)\n", len(clean), preserveMemoizationPragma)

	// Asserted exactly rather than as a floor. A run collecting zero clean fixtures restores the
	// blind spot this population exists to close -- a corpus of error fixtures cannot measure false
	// positives -- and it would report success while doing it. See `expectedCleanFixtureCount` for
	// why a nonzero check is not enough.
	if len(clean) != *expectClean {
		fail("found %d clean fixtures, expected %d; if upstream really changed, read the diff and pass -expect-clean",
			len(clean), *expectClean)
	}

	if *dryRun {
		fmt.Println("dry run, nothing written")
		return
	}

	written, err := writeFixtures(*sha, append(append([]string(nil), names...), clean...), *out)
	if err != nil {
		fail("writing fixtures: %v", err)
	}
	fmt.Printf("wrote            %d files (%d pairs) into %s\n", written, written/2, *out)
}

// errorFixtureNames returns the error-named inputs, relative to the fixture root.
func errorFixtureNames(sha string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/facebook/react/git/trees/%s?recursive=1", sha)
	body, err := get(url)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var tree struct {
		Truncated bool `json:"truncated"`
		Tree      []struct {
			Path string `json:"path"`
			Type string `json:"type"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(body).Decode(&tree); err != nil {
		return nil, err
	}

	// The contents API silently repeats pages when it caps; the tree API says so instead. A
	// truncated tree is missing files with no way to know which, so this refuses rather than
	// vendoring a corpus of unknown completeness.
	if tree.Truncated {
		return nil, fmt.Errorf("the git tree came back truncated, so the listing is incomplete and any count from it is wrong")
	}

	var names []string
	for _, entry := range tree.Tree {
		if entry.Type != "blob" {
			continue
		}
		name, found := strings.CutPrefix(entry.Path, upstreamFixtureDirectory+"/")
		if !found || strings.HasSuffix(name, ".expect.md") {
			continue
		}
		base := path.Base(name)
		if strings.HasPrefix(base, "error.") || strings.HasPrefix(base, "todo.error.") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// cleanFixtureNames returns the inputs carrying `preserveMemoizationPragma` whose expectation holds
// no error, relative to the fixture root.
//
// It reads the tarball rather than the tree listing, because the pragma is in the file's contents
// and an expectation's error block is in its pair's. Neither question can be answered from a list of
// paths, which is why this population was invisible to a tool built to filter on names.
func cleanFixtureNames(sha string) ([]string, error) {
	contents, err := fixtureContents(sha)
	if err != nil {
		return nil, err
	}

	var names []string
	for name, source := range contents {
		if strings.HasSuffix(name, ".expect.md") || !strings.Contains(string(source), preserveMemoizationPragma) {
			continue
		}
		expectation, paired := contents[strings.TrimSuffix(name, path.Ext(name))+".expect.md"]
		// An input with no expectation is not a judgement about it either way, so it is not
		// evidence of cleanliness and is left out rather than assumed silent.
		if !paired {
			continue
		}
		// Upstream writes the errors under a `## Error` heading, and its absence is what "expects no
		// error" means. Anchored to a line start so a fixture merely discussing the word in prose is
		// not mistaken for one that reports.
		if !strings.Contains("\n"+string(expectation), "\n## Error") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

// fixtureContents pulls the tarball once and returns every fixture file, keyed by its path relative
// to the fixture root.
//
// Memoised, because two callers need it -- the clean-fixture selection reads contents to find the
// pragma, and the write pass needs the same bytes. Fetching twice would double a 10MB download and,
// worse, admit the possibility of the two passes disagreeing about what upstream holds.
var fixtureContentsCache map[string][]byte

func fixtureContents(sha string) (map[string][]byte, error) {
	if fixtureContentsCache != nil {
		return fixtureContentsCache, nil
	}

	body, err := get(fmt.Sprintf("https://api.github.com/repos/facebook/react/tarball/%s", sha))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	gzipReader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	contents := map[string][]byte{}
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// The tarball root is a generated `owner-repo-sha` directory, so the first path segment is
		// dropped rather than matched.
		_, rest, cut := strings.Cut(header.Name, "/")
		if !cut {
			continue
		}
		name, under := strings.CutPrefix(rest, upstreamFixtureDirectory+"/")
		if !under {
			continue
		}
		file, err := io.ReadAll(tarReader)
		if err != nil {
			return nil, err
		}
		contents[name] = file
	}

	fixtureContentsCache = contents
	return contents, nil
}

// writeFixtures writes each pair out of the fetched tarball.
//
// One tarball rather than 650 blob requests: it is a single round trip, it cannot be rate-limited
// halfway into a corpus, and it gives byte-identical content to what a clone would.
func writeFixtures(sha string, names []string, out string) (int, error) {
	wanted := map[string]bool{}
	for _, name := range names {
		wanted[name] = true
		wanted[strings.TrimSuffix(name, filepath.Ext(name))+".expect.md"] = true
	}

	contents, err := fixtureContents(sha)
	if err != nil {
		return 0, err
	}

	found := map[string]bool{}
	written := 0
	for name := range wanted {
		file, present := contents[name]
		if !present {
			continue
		}
		destination := filepath.Join(out, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(destination, file, 0o644); err != nil {
			return written, err
		}
		found[name] = true
		written++
	}

	// Every input must have arrived with its expectation. Upstream pairs them 1:1 with zero missing
	// at the pinned sha, so a gap here means the fetch dropped something, and a corpus quietly
	// short a few pairs still scores and still looks complete.
	var missing []string
	for name := range wanted {
		if !found[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return written, fmt.Errorf("%d files listed in the tree never appeared in the tarball, first few: %v",
			len(missing), missing[:min(5, len(missing))])
	}
	return written, nil
}

func get(url string) (io.ReadCloser, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	// Unauthenticated GitHub allows 60 requests an hour, which two runs of this tool can exhaust.
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("%s: %s", url, response.Status)
	}
	return response.Body, nil
}

func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func fail(format string, arguments ...any) {
	fmt.Fprintf(os.Stderr, "react/vendor_fixtures: "+format+"\n", arguments...)
	os.Exit(1)
}
