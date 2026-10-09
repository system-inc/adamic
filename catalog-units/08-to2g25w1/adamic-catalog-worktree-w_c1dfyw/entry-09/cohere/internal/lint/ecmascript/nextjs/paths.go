// Package nextjs answers which Next.js file a path is, for the rules that gate on it.
//
// Eight rules in the `@next/next` family decide "is this the document file" and upstream answers it
// four different ways: a shared `is_document_page` that splits on `pages`, two byte-equivalent
// hand-rolled basename tests, a third basename test that also accepts a `_document/index.*`
// directory, and one rule with no gate at all.
//
// **That is drift upstream has accumulated rather than a specification to reproduce**, and a
// per-rule port would bake it in permanently: eight rules, four predicates, and no guard that could
// tell you they disagreed, because each passes its own fixtures.
//
// So this package is a deliberate divergence from the port target, ruled 2026-08-23. Fidelity is to
// what a rule decides rather than to how it obtains what it needs, and all eight are deciding the
// same thing.
package nextjs

import (
	"strings"
)

// IsDocumentFile reports whether a path is the Next.js custom document.
//
// The union of every spelling upstream uses, because they are four answers to one question:
//
//	pages/_document.tsx           every spelling agrees
//	src/pages/_document.tsx       every spelling agrees
//	pages/_document/index.tsx     only no_head_import_in_document accepts it
//	components/_document.tsx      the basename spellings accept it, is_document_page does not
//	pages/_documentation.tsx      nothing accepts it, and the trailing dot is why
//
// The trailing dot on `_document.` is load-bearing and is the one piece of upstream precision worth
// keeping: without it `_documentation.tsx` reads as the document, which is a real ESLint false
// positive that oxc fixed.
//
// The `pages` requirement is deliberately dropped. `is_document_page` requires it and three sibling
// rules do not, and requiring it would make this narrower than three of the eight rules it serves.
// A file named `_document.tsx` outside a pages directory is still a custom document by name, and a
// rule that ignored it would be silent on a real misplacement.
//
// So this predicate is wider than the shared helper on two paths and narrower on one, which is worth
// stating because a reader checking only one direction will conclude it drifts:
//
//	components/_document.tsx     wider   the basename spellings accept it, is_document_page does not
//	pages/_document/index.tsx    wider   only the widest spelling accepts it
//	pages/_documentation.tsx     NARROWER  is_document_page accepts it, this does not
//
// The last one is the trailing dot doing its work, and it is the direction that matters most: the
// shared helper tests `starts_with("/_document")` with no dot, so a file merely beginning with the
// word reads as the document. A research pass reported this comparison with the direction inverted,
// claiming we report that path and upstream exempts it. Checked at the source: the opposite is true,
// and the fixture asserting our answer was already correct.
func IsDocumentFile(filePath string) bool {
	baseName, parentName := splitPath(filePath)

	if strings.HasPrefix(baseName, "_document.") {
		return true
	}
	// The directory form, which only the widest upstream spelling accepts: `_document/index.tsx`
	// names the same thing as `_document.tsx` and a rule seeing one should see the other.
	return strings.HasPrefix(baseName, "index") && strings.HasPrefix(parentName, "_document")
}

// IsInApplicationDirectory reports whether a path sits under a Next.js app directory.
//
// A substring test rather than a segment test, matching upstream's `is_in_app_dir`, and **the
// looseness has a live cost on our own trees rather than a theoretical one.**
//
// `www-connected-app/` contains `app/`, so every file in that project answers true and any rule
// gating on this is silent across the whole checkout. A research pass found it while measuring
// `no-before-interactive-script-outside-document`: four checkouts report a real violation and two do
// not, and the two are the ones whose directory name ends in the word.
//
// Reproduced rather than tightened anyway, and that is a decision rather than an oversight.
// Upstream's rules gate on a routing model, and narrowing this to a path segment would make our
// rules silent on trees upstream lints. **Fixing it is a divergence that needs its own evidence**,
// and the evidence a future reader would want is which of the eight gating rules actually changes
// answer on a real tree.
//
// Recorded here so the next person to read this sees a measured cost rather than a theoretical
// looseness, and does not have to rediscover which projects it silences.
func IsInApplicationDirectory(filePath string) bool {
	return strings.Contains(filePath, "app/") || strings.Contains(filePath, "app\\")
}

// splitPath returns a path's base name and its parent directory's name.
//
// Both separators are handled because every upstream predicate handles both, and a path arriving in
// the Windows shape must answer the same as its POSIX twin. Splitting on the last separator of
// either kind is what makes that true without asking which platform produced the path.
func splitPath(filePath string) (baseName string, parentName string) {
	baseName = filePath
	if index := lastSeparator(filePath); index >= 0 {
		baseName = filePath[index+1:]
		parent := filePath[:index]
		parentName = parent
		if parentIndex := lastSeparator(parent); parentIndex >= 0 {
			parentName = parent[parentIndex+1:]
		}
	}
	return baseName, parentName
}

// lastSeparator returns the index of the final path separator of either kind, or -1.
func lastSeparator(filePath string) int {
	posix := strings.LastIndexByte(filePath, '/')
	windows := strings.LastIndexByte(filePath, '\\')
	if windows > posix {
		return windows
	}
	return posix
}

// IsDocumentPage reports whether a path is the custom document, answering the way oxc's shared
// `is_document_page` answers rather than the way IsDocumentFile above does.
//
// Both predicates exist on purpose and a reader arriving at one should be sent to the other, so the
// difference is stated here rather than left to be discovered:
//
//	IsDocumentFile   the union of every upstream spelling, for the family of rules whose own
//	                 upstream spellings disagree and whose corpora do not pin the disagreement
//	IsDocumentPage   oxc's single shared helper, byte for byte, for the one rule whose corpus
//	                 asserts the difference as a test
//
// `no-document-import-in-page` is that rule. Its fourth failing case is
// `src/pages/user/_document.tsx`, a file literally named `_document.tsx` that upstream deliberately
// **reports**, and its pass list includes `pages/_documentation.tsx`-shaped exemptions by way of a
// prefix test that carries no trailing dot. IsDocumentFile answers the opposite on both, and it is
// right to: it serves rules whose corpora never vote on those paths. Here the corpus votes, so the
// vote wins.
//
// Measured against the release oxlint binary rather than modelled, on all eleven corpus paths plus
// the two disagreements:
//
//	src/pages/user/_document.tsx   reports upstream    IsDocumentFile exempts it
//	components/_document.tsx       reports upstream    IsDocumentFile exempts it
//	pages/_documentation.tsx       exempt upstream     IsDocumentFile reports it
//
// The mechanism is a split on the bare word `pages` keeping the LAST segment, so only the immediate
// child of a `pages` directory is the document and a path with no `pages` in it at all is never the
// document. `pagesapp/src/pages/_document.js` is upstream's own probe of the `.last()`: splitting
// gives `["", "app/src/", "/_document.js"]` and the last one exempts.
//
// Both separators are tested, unconditionally, exactly as upstream does. ESLint uses the platform's
// separator plus the posix one, so on a Unix host it tests the forward slash twice and never tests
// the backslash. oxc is strictly more correct there and oxc is the port target.
//
// One further ESLint divergence is reproduced in oxc's direction rather than ESLint's: ESLint
// carries a leading emptiness guard, so a path ENDING in `pages` splits to an empty last segment and
// exempts, where oxc reports. Unreachable in practice, since a lintable file needs an extension, but
// it is asserted in the tests below so the choice is visible rather than incidental.
func IsDocumentPage(filePath string) bool {
	segments := strings.Split(filePath, "pages")
	page := segments[len(segments)-1]
	return strings.HasPrefix(page, "/_document") || strings.HasPrefix(page, `\_document`)
}

// IsInPagesDirectory reports whether a path is a Pages Router route rather than an API route.
//
// Two questions in one answer, because upstream asks them together: the file must sit under a
// `pages` directory at all, and the segment *directly* under `pages` must not be `api`. A file that
// is under no `pages` directory answers false, which is the same false a file under `pages/api`
// gets, and both mean "this rule does not apply here".
//
//	pages/index.tsx           true
//	src/pages/index.tsx       true
//	pages/api/user.ts         false   an API route, which returns data rather than rendering
//	pages/blog/api/x.tsx      true    only the top-level api directory is Next's API routes
//	pages/api                 false   the segment after pages is api, with nothing after it
//	mypages/index.tsx         false   a segment test, not a substring test
//	components/Thing.tsx      false
//
// # Why a segment test here when IsInApplicationDirectory above is a substring test
//
// The two look inconsistent and are faithful to two different upstream spellings. oxc's `should_run`
// for this walks path *components* and compares each to `pages` and `api` exactly, so a directory
// named `mypages` does not match and `apiary` does not exempt. Its sibling `is_in_app_dir` really is
// a bare `contains`. Reproducing each as written is why one is loose and one is tight.
//
// The `@next/eslint-plugin-next` original spells it a third way, by splitting the filename on the
// literal text `pages` and asking whether the remainder's directory starts with `/api`. That is
// wrong in two directions and oxc fixed both: `mypages/x.tsx` runs under ESLint because the split
// matches inside a longer word, and `pages/apiary/x.tsx` is exempted because `"/apiary"` starts with
// `"/api"`. oxc is the port target and it is also simply right, so no note is needed at any call
// site.
//
// Only the FIRST `pages` segment decides, matching oxc's walk: it sets a flag on the first match and
// returns on the very next component. So `pages/api/pages/x.tsx` is false, because the segment after
// the first `pages` is `api`, and the second `pages` is never consulted.
//
// Both separators are handled here, as everywhere in this package. Paths reaching a rule have
// already been normalized to forward slashes, so the backslash arm cannot currently fire; it is kept
// because every other predicate in this package accepts both and a caller should not have to know
// which of them normalizes.
func IsInPagesDirectory(filePath string) bool {
	segments := splitSegments(filePath)
	for index, segment := range segments {
		if segment != "pages" {
			continue
		}
		// The component immediately after `pages` decides, and a `pages` with nothing after it is
		// not a route directory at all. oxc reaches its `return` only when a next component exists.
		if index+1 >= len(segments) {
			return false
		}
		return segments[index+1] != "api"
	}
	return false
}

// splitSegments breaks a path on separators of either kind, dropping empty segments.
//
// Empty segments are dropped so that a leading separator, a trailing one, or a doubled one cannot
// shift what counts as "the component after pages". The harness roots every fixture path, so a
// leading empty segment is the ordinary case rather than a malformed one.
func splitSegments(filePath string) []string {
	segments := make([]string, 0, 8)
	start := 0
	for index := 0; index <= len(filePath); index++ {
		if index == len(filePath) || filePath[index] == '/' || filePath[index] == '\\' {
			if index > start {
				segments = append(segments, filePath[start:index])
			}
			start = index + 1
		}
	}
	return segments
}
