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
