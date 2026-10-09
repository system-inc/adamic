package differential

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/native"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// TestDumpDifferences is TestCorpora's working copy for a printer port: the same comparison against the
// same oracle cache, but every difference is written out whole instead of three samples per language,
// so a porter can group the failures by construct and read the expected output beside its own.
//
// Off by default. Driven by TestCorpora's variables plus:
//
//	COHERE_FORMAT_DUMP        directory to write into: <corpus>/summary.txt, and <corpus>/files/<path>.expected
//	                          and .actual for every differing file
//	COHERE_FORMAT_EXTENSIONS  comma-separated extensions to compare, e.g. .ts,.tsx (default: every file)
//	COHERE_FORMAT_PRINT_WIDTH overrides each corpus's printWidth
//
// The print width override is a stress test, not the acceptance number. Our trees are already formatted
// at their own width, so most files are identical to their formatted form and only a few hundred test a
// layout decision. At another width nearly every file is rewritten, which exercises the breaking logic
// the way freshly written code will. The oracle cache keys on the options, so each width warms its own.
//
// The candidate is always native. It reads the cache TestCorpora warms under the same identity, so a
// dump costs the candidate's time once the corpus has been measured.
func TestDumpDifferences(t *testing.T) {
	t.Parallel()
	roots, output := os.Getenv("COHERE_FORMAT_CORPORA"), os.Getenv("COHERE_FORMAT_DUMP")
	if roots == "" || output == "" {
		t.Skip("set COHERE_FORMAT_CORPORA and COHERE_FORMAT_DUMP to dump every difference")
	}
	extensions := map[string]bool{}
	for _, extension := range strings.Split(os.Getenv("COHERE_FORMAT_EXTENSIONS"), ",") {
		if extension = strings.TrimSpace(extension); extension != "" {
			extensions[extension] = true
		}
	}

	bundles, err := prettier.Bundles()
	if err != nil {
		t.Fatal(err)
	}
	digest, err := prettier.DigestBundles(bundles.Files)
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("COHERE_FORMAT_CACHE")
	if directory == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			t.Fatal(err)
		}
		directory = filepath.Join(base, "cohere", "prettier-oracle")
	}
	enumerator, err := prettier.New(formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}

	pending := strings.Split(roots, ":")
	seen := map[string]bool{}
	for len(pending) > 0 {
		root := expandHome(strings.TrimSpace(pending[0]))
		pending = pending[1:]
		if root == "" || seen[root] {
			continue
		}
		seen[root] = true

		// Nested repositories regardless of root's own ignore lists, before anything can skip root, as
		// TestCorpora finds them (#k6vebep).
		nestedRepositories, err := formatfiles.NestedRepositoriesBelow(root)
		if err != nil {
			t.Fatalf("finding the repositories nested in %s: %v", root, err)
		}
		for _, nested := range nestedRepositories {
			pending = append(pending, filepath.Join(root, nested))
		}

		resolution, err := formatoptions.Resolve(root)
		if errors.Is(err, formatoptions.ErrPrettierConfigRemains) {
			// A repository still carrying Prettier config, pinned to its own adoption (www-ahra-ai). The
			// differential's TestCorpora refuses this for a named corpus; here it is named and skipped.
			t.Logf("skipping %s, not yet adopted: %v", root, err)
			continue
		}
		if err != nil {
			t.Fatalf("resolving the format options for %s: %v", root, err)
		}

		enumeration, err := enumerator.Enumerate(root)
		if errors.Is(err, formatoptions.ErrPrettierConfigRemains) {
			t.Logf("skipping %s, not yet adopted: %v", root, err)
			continue
		}
		if err != nil {
			t.Fatalf("enumerating %s: %v", root, err)
		}
		files := enumeration.Files
		if len(extensions) > 0 {
			files = nil
			for _, file := range enumeration.Files {
				if extensions[filepath.Ext(file)] {
					files = append(files, file)
				}
			}
		}

		options := resolution.Options
		if width := os.Getenv("COHERE_FORMAT_PRINT_WIDTH"); width != "" {
			if _, err := fmt.Sscan(width, &options.PrintWidth); err != nil {
				t.Fatalf("COHERE_FORMAT_PRINT_WIDTH %q: %v", width, err)
			}
		}
		newOracle := func() (Formatter, error) { return prettier.New(options) }
		newCandidate := func() (Formatter, error) { return native.Formatter{Options: options}, nil }
		// The identity TestCorpora uses, so both read and warm one cache.
		cache := &OracleCache{
			Directory: directory,
			Identity:  digest + "|" + string(bundles.Origin) + "|filepath|" + describeOptions(options),
		}

		report, err := Compare(root, files, newOracle, newCandidate, runtime.NumCPU(), cache)
		if err != nil {
			t.Fatalf("comparing %s: %v", root, err)
		}
		t.Log(report.Summary())

		// The whole root names the directory, because nested corpora share basenames (three are nexus).
		corpus := filepath.Join(output, strings.ReplaceAll(strings.Trim(root, string(filepath.Separator)), string(filepath.Separator), "_"))
		var summary strings.Builder
		summary.WriteString(report.Summary())
		results := append([]FileResult(nil), report.Files...)
		sort.Slice(results, func(left, right int) bool { return results[left].Path < results[right].Path })
		oracle, err := newOracle()
		if err != nil {
			t.Fatal(err)
		}
		for _, result := range results {
			if result.Outcome != Different && result.Outcome != CandidateRefused {
				continue
			}
			relative, _ := filepath.Rel(root, result.Path)
			fmt.Fprintf(&summary, "%s\t%s\t%s\n", result.Outcome, relative, firstLine(result.Detail))
			if result.Outcome != Different {
				continue
			}
			source, err := os.ReadFile(result.Path)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := cache.format(oracle, result.Path, string(source))
			if err != nil {
				t.Fatal(err)
			}
			actual, err := native.Formatter{Options: options}.Format(result.Path, string(source))
			if err != nil {
				t.Fatal(err)
			}
			base := filepath.Join(corpus, "files", relative)
			if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(base+".expected", []byte(expected), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(base+".actual", []byte(actual), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.MkdirAll(corpus, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(corpus, "summary.txt"), []byte(summary.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
