package nextjs

import "testing"

// The cases where upstream's four spellings disagree are the ones worth pinning, because they are
// the reason this package exists. Each row records which upstream spellings would have said what,
// so a reader can see the drift this replaced rather than take the collapse on faith.
func TestIsDocumentFile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filePath string
		want     bool
	}{
		// Every upstream spelling agrees on these two.
		{"the ordinary document", "pages/_document.tsx", true},
		{"under a source directory", "src/pages/_document.tsx", true},

		// Only the widest spelling accepts the directory form. It names the same file.
		{"the directory form", "pages/_document/index.tsx", true},
		{"the directory form with a source directory", "src/pages/_document/index.jsx", true},

		// The basename spellings accept this and `is_document_page` does not, because it requires a
		// `pages` segment. Dropped deliberately: three of the eight rules do not require it, and
		// requiring it here would make this narrower than the rules it serves.
		{"outside a pages directory", "components/_document.tsx", true},

		// The trailing dot, which is the one piece of upstream precision worth keeping. Without it
		// this reads as the document, and that is a real ESLint false positive oxc fixed.
		{"a longer name sharing the prefix", "pages/_documentation.tsx", false},
		// Upstream's directory case uses a bare `starts_with("_document")` with no trailing dot,
		// unlike its basename case. So this DOES match, and it is recorded as upstream's behavior
		// rather than corrected: a fixture asserting the tidier answer was written first here and
		// was wrong, which is the trap of trusting what a rule should do over what it does.
		{"a longer directory name still matches, as upstream", "pages/_documentation/index.tsx", true},

		{"an ordinary page", "pages/index.tsx", false},
		{"an index outside a document directory", "pages/settings/index.tsx", false},
		{"a component", "components/Thing.tsx", false},
		{"an empty path", "", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsDocumentFile(testCase.filePath); got != testCase.want {
				t.Fatalf("%q: want %v, got %v", testCase.filePath, testCase.want, got)
			}
		})
	}
}

// Every upstream predicate handles both separators, so a path in the Windows shape must answer the
// same as its POSIX twin. A port handling one would be silent on half the platforms upstream lints.
func TestIsDocumentFileHandlesBothSeparators(t *testing.T) {
	t.Parallel()
	pairs := [][2]string{
		{"pages/_document.tsx", `pages\_document.tsx`},
		{"src/pages/_document.tsx", `src\pages\_document.tsx`},
		{"pages/_document/index.tsx", `pages\_document\index.tsx`},
		{"pages/_documentation.tsx", `pages\_documentation.tsx`},
	}
	for _, pair := range pairs {
		posix, windows := pair[0], pair[1]
		if IsDocumentFile(posix) != IsDocumentFile(windows) {
			t.Fatalf("%q and %q disagree: %v against %v",
				posix, windows, IsDocumentFile(posix), IsDocumentFile(windows))
		}
	}
}

// A substring test rather than a segment test, matching upstream. The looseness is reproduced rather
// than tightened, and the case that shows it is asserted so nobody "fixes" it later without seeing
// what they are changing.
func TestIsInApplicationDirectory(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"an app route", "app/page.tsx", true},
		{"nested under app", "src/app/settings/page.tsx", true},
		{"the windows shape", `src\app\settings\page.tsx`, true},
		// A near-miss that does NOT match, recorded because it is the case a reader assumes the
		// substring test catches: `application/` contains `app` but not `app/`, so the separator in
		// the needle is doing real work.
		{"a directory whose name merely starts with app", "application/thing.tsx", false},
		// The looseness that IS real: a segment ending in `app` answers true.
		// Not theoretical: this is our own project directory, so every file in that checkout
		// answers true and any rule gating on this is silent across it. Asserted so the cost is
		// visible rather than rediscovered.
		{"our own www-connected-app checkout", "www-connected-app/source/Thing.tsx", true},
		{"a directory ending in app", "src/myapp/thing.tsx", true},
		{"a pages route", "pages/index.tsx", false},
		{"an empty path", "", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsInApplicationDirectory(testCase.filePath); got != testCase.want {
				t.Fatalf("%q: want %v, got %v", testCase.filePath, testCase.want, got)
			}
		})
	}
}

// The tree this gates measured 37 `pages/` directories in source that are not Pages Router at all,
// a PascalCase component convention plus one holding translations. None of them is a document, and
// a predicate that answered otherwise would put a false-positive surface under every rule that
// gates this way.
func TestOurOwnPagesConventionIsNotADocument(t *testing.T) {
	t.Parallel()
	for _, filePath := range []string{
		"libraries/structure/source/modules/account/pages/AccountPage.tsx",
		"libraries/structure/source/modules/account/pages/index.tsx",
		"modules/finance/pages/FinanceOverviewPage.tsx",
	} {
		if IsDocumentFile(filePath) {
			t.Fatalf("%q is our own pages convention and must not read as a Next document", filePath)
		}
	}
}

// The corpus of `no-document-import-in-page` is this predicate's specification, so every path it
// names is asserted here as well as in the rule. The three rows where this and IsDocumentFile
// disagree are the reason both exist, and each was confirmed against the release oxlint binary
// rather than modelled from the Rust.
func TestIsDocumentPage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filePath string
		want     bool
	}{
		// Upstream's seven passing paths.
		{"the ordinary document", "pages/_document.js", true},
		{"the document in TypeScript", "pages/_document.tsx", true},
		{"the page suffix convention", "pages/_document.page.tsx", true},
		{"the directory form", "pages/_document/index.js", true},
		{"the directory form in TypeScript", "pages/_document/index.tsx", true},
		// Upstream's own probe of the LAST split segment. Splitting on the bare word gives
		// ["", "app/src/", "/_document.js"], and the last one exempts.
		{"a directory merely beginning with the word pages", "pagesapp/src/pages/_document.js", true},

		// Upstream's four failing paths.
		{"outside any pages directory", "components/test.js", false},
		{"an ordinary page", "pages/test.js", false},
		{"a page nested below pages", "src/pages/user/test.tsx", false},
		// The case that decides which predicate the rule may use. Named exactly `_document.tsx`
		// and still reported upstream, because only the immediate child of a pages directory
		// counts. IsDocumentFile answers true here.
		{"a document not immediately under pages", "src/pages/user/_document.tsx", false},

		// The other two disagreements with IsDocumentFile, both measured on the binary.
		// No trailing dot in upstream's prefix test, so the longer word still reads as the
		// document and the import is exempt. IsDocumentFile answers false here, deliberately.
		{"a longer name sharing the prefix", "pages/_documentation.tsx", true},
		// No pages segment at all, so never the document however it is named. IsDocumentFile
		// answers true here, deliberately.
		{"a document outside pages entirely", "components/_document.tsx", false},

		// The Windows separator branch, which upstream tests unconditionally and no upstream
		// fixture exercises.
		{"the windows shape", `pages\_document.js`, true},
		{"the windows directory form", `pages\_document\index.tsx`, true},

		// Reproduced in oxc's direction rather than ESLint's. ESLint carries a leading emptiness
		// guard so a path ENDING in pages exempts; oxc has none, so the empty last segment fails
		// both prefix tests and the file is treated as an ordinary page. Unreachable in practice
		// because a lintable file needs an extension, asserted so the choice is visible.
		{"a path ending in the word pages", "pages", false},
		{"a nested path ending in the word pages", "x/pages", false},

		{"an empty path", "", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsDocumentPage(testCase.filePath); got != testCase.want {
				t.Fatalf("%q: want %v, got %v", testCase.filePath, testCase.want, got)
			}
		})
	}
}

// The two predicates are not interchangeable and a reader who assumes they are would break a
// corpus. Asserted rather than left to the doc comments, so a later tidy-up that collapses them
// fails loudly instead of silently changing which files a rule runs on.
func TestTheTwoDocumentPredicatesDisagreeOnPurpose(t *testing.T) {
	t.Parallel()
	disagreements := []struct {
		filePath     string
		wantPage     bool
		wantFile     bool
		whyItMatters string
	}{
		{"src/pages/user/_document.tsx", false, true,
			"upstream's fourth failing case for no-document-import-in-page"},
		{"components/_document.tsx", false, true,
			"no pages segment, so the shared helper never treats it as the document"},
		{"pages/_documentation.tsx", true, false,
			"the shared helper's prefix test carries no trailing dot"},
	}
	for _, disagreement := range disagreements {
		if got := IsDocumentPage(disagreement.filePath); got != disagreement.wantPage {
			t.Fatalf("IsDocumentPage(%q): want %v, got %v (%s)",
				disagreement.filePath, disagreement.wantPage, got, disagreement.whyItMatters)
		}
		if got := IsDocumentFile(disagreement.filePath); got != disagreement.wantFile {
			t.Fatalf("IsDocumentFile(%q): want %v, got %v (%s)",
				disagreement.filePath, disagreement.wantFile, got, disagreement.whyItMatters)
		}
	}
}

// The four rows separating oxc's component walk from the ESLint original's string surgery are the
// reason this is a segment test, and each was pinned against the release oxlint binary rather than
// reasoned about. ESLint would answer the opposite on all four.
func TestIsInPagesDirectory(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filePath string
		want     bool
	}{
		{"an ordinary page", "pages/index.tsx", true},
		{"under a source directory", "src/pages/index.tsx", true},
		{"the harness shape, rooted", "/pages/index.tsx", true},
		{"deeply nested under pages", "pages/blog/posts/x.tsx", true},

		// Next's API routes, which return data rather than rendering, so a data fetching function
		// has no meaning there.
		{"an api route", "pages/api/user.ts", false},
		{"a nested api route", "pages/api/v1/user.ts", false},

		// Only the segment DIRECTLY after pages exempts. A deeper api directory is an ordinary
		// route directory that happens to be named api, and upstream runs on it.
		{"an api directory nested deeper", "pages/blog/api/x.tsx", true},

		// A segment test rather than a prefix test. ESLint exempts this because it asks whether the
		// remaining directory startsWith `/api`, which is a real upstream defect oxc fixed.
		{"a directory whose name begins with api", "pages/apiary/x.tsx", true},

		// A segment test rather than a substring test. ESLint runs on this because it splits the
		// filename on the literal text `pages` and finds it inside the longer word.
		{"a directory whose name ends in pages", "mypages/index.tsx", false},
		{"a directory whose name begins with pages", "pagesx/index.tsx", false},

		// Nothing after `pages`, so there is no component for the walk to judge. ESLint runs on
		// this because path.parse's dir is `/` rather than `/api`.
		{"a path ending at pages", "pages", false},
		{"a path ending at pages with a separator", "src/pages/", false},

		// The FIRST pages segment decides and the second is never consulted, matching oxc's walk,
		// which returns on the component immediately after its first match.
		{"a pages directory nested inside an api route", "pages/api/pages/x.tsx", false},

		{"outside any pages directory", "components/Thing.tsx", false},
		{"an app router page", "app/page.tsx", false},
		{"an empty path", "", false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsInPagesDirectory(testCase.filePath); got != testCase.want {
				t.Fatalf("%q: want %v, got %v", testCase.filePath, testCase.want, got)
			}
		})
	}
}

// Both separators, as everywhere in this package. Paths reaching a rule are already normalized to
// forward slashes, so the backslash arm cannot fire in production; it is asserted because a caller
// should not have to know which layer normalizes, and because upstream's own corpus ships a
// backslash path whose verdict on a non-Windows host is an accident of PathBuf rather than a
// decision.
func TestIsInPagesDirectoryHandlesBothSeparators(t *testing.T) {
	t.Parallel()
	pairs := [][2]string{
		{"pages/index.tsx", `pages\index.tsx`},
		{"pages/api/user.ts", `pages\api\user.ts`},
		{"src/pages/index.tsx", `src\pages\index.tsx`},
		{"mypages/index.tsx", `mypages\index.tsx`},
	}
	for _, pair := range pairs {
		posix, windows := pair[0], pair[1]
		if IsInPagesDirectory(posix) != IsInPagesDirectory(windows) {
			t.Fatalf("%q and %q disagree: %v against %v",
				posix, windows, IsInPagesDirectory(posix), IsInPagesDirectory(windows))
		}
	}
}
