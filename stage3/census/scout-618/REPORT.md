# sys.ts:618 scout (checkpoint; measurements pending)

Provisional recommendation: retain the cached string canonicalizer and use a typed
wrapper with a direct `as Path` on its string result. This preserves the original
canonicalization operation; `toPath` also normalizes paths. No measured candidate
counts are available at this checkpoint. No adaptation or compiler change.

Base: e71fab21338c77ed365f56647da147f149cf8114. Isolated detached worktree
created at this base, then this new report branch. TypeScript v6.0.3 source pin
confirmed as 050880ce59e30b356b686bd3144efe24f875ebc8. Recursive submodules
initialized and stage3/api npm ci completed. Scratch: /tmp/scout-618.

At this base BRANDS-VOID.md does not exist: BRANDS-UNDEFINED.md explicitly
supersedes the October void ruling. Phantom members retain their names and become
undefined. Rule 103 handles PathPathComponents as a generic interface extending
Array<T>, with the required undefined __pathComponensBrand. This is an array-brand
rule, distinct from primitive Path's intersection.

Source inspection: castProof first calls phantomCast, whose phantomAssignable
compares primitive bases. Path's undefined member passes phantomField. Thus a
direct string-result `as Path` has an existing admission path. This is distinct
from brandMintOf, which recognizes an interface extending one base. A full
path.ts census is still required to verify lowering rather than only scan admission.

Rule 99's README and owners.json were read. The recursive-watcher owner is pinned
whole. Its after text still contains `let timerToUpdateChildWatches: any;`.
Future work would amend this owner in rule 99; this scout changes neither pin.

## Reader inventory (whole rule-99 after owner)

There are five call expressions plus one function-value use:

- `createDirectoryWatcher`: `dirPath` keys cache.get/set/delete,
  callbackCache.add/remove, callback dispatch, and child-watch updates. Captured
  callbacks and close retain this key; updates use it in the pending map too.
- `toPathInLink`: passes the function to getRelativePathFromDirectory as its
  string canonicalizer; the returned Path is consumed as a string comparison
  key, and combinePaths produces a string filename.
- Both recursive `invokeCallbacks` branches: canonicalize `link` for dirPath;
  this keys cache/links and participates in callback-root equality, prefix,
  separator, and length tests. One branch carries an InvokeMap, the other a
  single filename. Callback payloads remain filenames, not branded keys.
- `removeChildWatches`: canonicalized childWatcher.dirName keys cache.get before
  recursively removing its children.
- `isInPath`: canonicalized path is used only for string.includes(searchPath)
  after the direct test fails and case-sensitive mode is excluded.

No consumer observes the phantom brand. The Path-consuming boundaries are
map/multimap keys and the dirPath parameters. The function-value consumer and
includes consumer need only strings.

## Completed baseline measurements

Fresh full apply exited 0. The adapted sys.ts contains rule 99's exact whole
`after` owner once. Its cast is at 618:33 and its timer local at 615.
Stock TypeScript 6.0.3, all 79 recursively enumerated compiler roots, strict,
noUncheckedIndexedAccess, exactOptionalPropertyTypes, ES2024, ESNext/Bundler,
verbatimModuleSyntax, forced modules and pinned @types/node: **0 diagnostics**.

The guarded full single-file CommonJSBundle census has a completed relation
trailer, no overflow, and checker_rejected=false (the census's generic
measurement label still says “measured on a checker-rejected program”). Counts
are unique (kind, where, reason, text), combined scan and lowering; boundaries
are counted separately. Baseline sys.ts: **41 NotYet, 19 Refused, 55 Boundary**.

The full path.ts census completed: **13 NotYet, 3 Refused, 16 Boundary**, no
errors/panics. `toPath` at adapted 764:1 is attempted with no finding. Its return
cast at 768 is therefore accepted in actual lowering, not merely absent from a
scan. The rule is castProof -> phantomCast -> phantomAssignable, which compares
Path's string base after phantomBase admits its undefined, string-keyed,
non-prototype brand member. This is the primitive rule in docs/rulings.md:29–31,
35–37, not the interface/array brandMintOf rule. Rule 103's generic interface
permits the separate PathPathComponents array brand; it does not mint Path.

The timer local is **not settled**. Full baseline already finds NotYet
`reading timerToUpdateChildWatches` at 764:13 (truth test before cancellation),
782:17 (while condition), and 809:117 (log interpolation), each with a failed
statement Boundary at 764:9, 782:9 and 809:9 respectively. These are latent
nested-unit observations; they do not imply the ordinary first-stop compiler
reaches them. Adaptation 40's TIMER-SLOT.md covers only the solution builder's
`timerToBuildInvalidatedProject`; rule 99's concrete host methods do not retype
this local.
