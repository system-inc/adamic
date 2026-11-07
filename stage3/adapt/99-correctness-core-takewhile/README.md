# takeWhile: measurement before construction

Task b5w3ycg, wave row 11. Main base:
`f0271c6a7b2e92900eb7046d8246cfb89f52cfb6`.
Pinned TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Continuation in progress. The earlier sparse-witness stop is historical in
PREVIOUS-STOP.md. The user clarified that equivalence is required on all lane
inputs tsc supplies, established by a complete count. A sparse accepted hole
is an off-count difference if its counted frequency is zero. No adapter has
been built yet. COUNT.json will record completed canonical pieces.

The scratch-only original implementation records every predicate argument and
answer, the accepted indices missing under `i in array`, the original result,
and the proposed push loop's replay. The replay uses recorded predicate answers
and checks exact Object.is element identity and argument order, so the real
predicate is called only once. Input slots must remain unchanged. Result length,
prototype, own keys and element identities are compared before encoding raw
JSON. Seed probes deliberately plant accepted holes, reversed identities, input
mutation and extra predicate calls; they are excluded from the oracle count.

All unchanged lane manifest units run through stage3/lane/shard.py with its
normal identity, API, baseline and count guards, four units at a time and the
runner-supported 90-second bound. Each piece has at most 20 units. Raw calls,
reports and observations are archived with internal SHA256SUMS and adjacent
archive hashes. reconcile-count.py independently checks every member, reconstructs
the call classification and verifies the complete lane inventory and results.
An unchanged main stage3/lane/run.sh control checks the same identity multiset
and raw baseline diff, without adding calls to the census.

Static full-src caller audit includes transformers/ts.ts:1086 and
services/refactors/moveToFile.ts:252 in the pinned source. The adapted audit is
retained with the measurement evidence. 99s leaves takeWhile unchanged;
30-indexed-reads owns only its existing indexed-read assertion, not the whole
owner. The initial ownership search and witnesses are in PREVIOUS-STOP.md.

The first baseline apply failed when a concurrent setup npm ci briefly removed
the shared API parser. That log is retained separately. Setup and all package
installs completed before the fresh baseline retry. The failed apply is excluded
from all measurement and proof claims.

## Construction checkpoint

The complete count gate has passed and the adapter is now built. The preceding
measurement-in-progress status is historical; final proof results will replace it.
The existing indexed-read assertion from adaptation 30 remains at the element read.
