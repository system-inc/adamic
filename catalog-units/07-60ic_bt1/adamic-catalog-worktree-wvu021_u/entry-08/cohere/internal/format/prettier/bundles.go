package prettier

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

/*
 * Where the Prettier bundles come from, and why the answer is bytes rather than a path.
 *
 * This replaced a function that returned a directory. That shape made disk the default and embedding
 * the special case, which is backwards once embedding is the norm, and it left the guard in
 * `internal/release` vouching for a directory while the engine read somewhere else. Returning bytes
 * plus their provenance means a caller cannot ask where the bundles are without also being told
 * which mode produced them.
 *
 * The bundles are committed under bundles/ rather than pinned from upstream. That is not a departure
 * from how this repo handles generated artifacts: typescript-go embeds 108 committed `libs/*.d.ts`
 * from its pinned checkout, which is the same pattern. It works there because upstream commits its
 * build output. Our fork does not -- `/dist*` is gitignored in `system/prettier` and `git ls-files
 * dist/prettier` returns zero, measured from inside that repository -- so a pin would give us 9,343
 * source files and a Node build step, which is the dependency vendoring exists to remove. The commit
 * happens here because upstream declines to do it there.
 *
 * Checking that claim about typescript-go costs a paragraph and it is worth it, because the obvious
 * probe returns the wrong answer with no error. `git ls-files TypeScript/tsc/internal/bundled/libs`
 * from this repository returns 0 and exit 0. The same command from inside the submodule returns 108.
 * A submodule's contents never appear in the parent's index, so both readings are correct answers to
 * different questions asked in identical words: does the parent track these (no, and it never will)
 * against does the submodule track these (yes). If a path crosses a submodule boundary, `git
 * ls-files` from outside is answering about the pin rather than the contents, and `git submodule
 * status` names every boundary in one line.
 */

// bundleFS holds the vendored bundles.
//
// The eight paths are named rather than globbed, and that is load-bearing. `go:embed bundles` would
// take whatever happens to be in the directory, so a bundle deleted from the tree would produce a
// binary that builds and then fails at runtime on the one language whose plugin went missing. Named
// paths make the same mistake a compile error, which is the missing-binary rule applied one layer
// down: absence fails loudly rather than degrading into something that half works.
//
// Keep this list and BundleFiles in the same order and the same membership. TestEmbedMatchesBundleFiles
// holds them together, because a divergence here is silent in the direction that matters -- an embed
// naming a file BundleFiles does not load is dead weight, and BundleFiles naming a file the embed
// does not carry is a binary that cannot format.
//
//go:embed bundles/standalone.js
//go:embed bundles/plugins/estree.js
//go:embed bundles/plugins/typescript.js
//go:embed bundles/plugins/babel.js
//go:embed bundles/plugins/postcss.js
//go:embed bundles/plugins/markdown.js
//go:embed bundles/plugins/graphql.js
//go:embed bundles/plugins/yaml.js
var bundleFS embed.FS

// BundleOrigin says which copy of the bundles a BundleSource carries.
//
// It is set explicitly by whichever branch produced the bytes, never inferred from the filesystem
// afterwards. Inferring it is how the last defect in this package happened: two paths to one
// artifact, each believing it knew where the other was reading.
type BundleOrigin string

const (
	// Embedded means the bytes came from this binary, which is the normal case.
	Embedded BundleOrigin = "Embedded"

	// Disk means ForkPathVariable was set and the bytes came from a checkout.
	Disk BundleOrigin = "Disk"
)

// BundleSource is the bundles and where they came from.
//
// Path is empty when Origin is Embedded, deliberately. An empty string is unusable rather than
// misleading, so a guard that forgets to branch on Origin fails visibly instead of quietly running
// `git -C ""` against the current directory and reporting on the wrong repository.
type BundleSource struct {
	Files  map[string][]byte
	Origin BundleOrigin
	Path   string
}

// Bundles returns the bundle bytes and their provenance.
//
// The mode switch is one condition in one place: ForkPathVariable set means load from disk instead
// of the embedded copy. That is a different meaning from what the variable used to carry -- it used
// to answer "where is the fork", back when disk was the only mode -- and the change is the point.
// A developer pointing at a live fork is the only caller in Disk mode.
//
// There is no caching. `New` is called once per process (cmd/cohere/format.go:149, reached from one
// production call site), so the read is not on a hot path, and a cached read cannot see a fork that
// changed mid-run -- which is exactly the case Disk mode exists to serve. Caching here would
// optimize the mode that does not need it at the cost of correctness in the mode that does.
func Bundles() (BundleSource, error) {
	forkPath := strings.TrimSpace(os.Getenv(ForkPathVariable)) // Go whitespace: a path from the environment, which no JavaScript tool reads.
	if forkPath == "" {
		return embeddedBundles()
	}
	return diskBundles(filepath.Join(forkPath, "dist", "prettier"))
}

// embeddedBundles reads the bundles compiled into this binary.
//
// A missing name here is a broken build rather than a broken machine, since `go:embed` would have
// refused to compile. It is still checked, because the check costs nothing and the alternative is
// trusting that this function and the directive above never drift.
func embeddedBundles() (BundleSource, error) {
	files := make(map[string][]byte, len(BundleFiles))
	for _, name := range BundleFiles {
		content, err := fs.ReadFile(bundleFS, "bundles/"+name)
		if err != nil {
			return BundleSource{}, fmt.Errorf("the embedded prettier bundle %s is missing: %w", name, err)
		}
		files[name] = content
	}
	return BundleSource{Files: files, Origin: Embedded}, nil
}

// diskBundles reads the bundles from a fork checkout.
//
// Every name in BundleFiles must be present or this fails and says which. No partial load: an engine
// that quietly loaded seven of eight would format the eighth's files by falling through to no parser
// at all, and the failure would look like a file type nobody formats rather than a broken checkout.
func diskBundles(directory string) (BundleSource, error) {
	files := make(map[string][]byte, len(BundleFiles))
	for _, name := range BundleFiles {
		content, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(name)))
		if err != nil {
			return BundleSource{}, fmt.Errorf(
				"reading prettier bundle %s from %s: %w\nUnset %s to use the embedded bundles",
				name, directory, err, ForkPathVariable,
			)
		}
		files[name] = content
	}
	return BundleSource{Files: files, Origin: Disk, Path: directory}, nil
}
