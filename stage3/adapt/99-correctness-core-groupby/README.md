# core-groupby: type-only cast removal

Task b5w3ycg, wave rows 4 and 5. Rule class: type-only.
Site: F4, the mapped-result return assertion in groupBy.

Both declared overloads stay unchanged. The implementation and its local result use Record<string, T[]> consistently, so return result needs no assertion. The overloads continue to expose the mapped optional-key result and type-predicate result to callers. This is a type-only implementation annotation change.

Adaptation 30 pins values[i] by expression and occurrence 2 of 3 in core.ts, adding the existing non-null assertion. That site remains byte-for-byte unchanged. This adapter pins the whole owner after adaptation 30 and all earlier directories. The paired replay compares adaptation 30 and every other adapter with and without this directory.

The adapter accepts exactly the complete before or after owner text, including
all overloads, and refuses drift, duplicate or nested owners, and partial edits.
It preserves line endings and performs zero writes on an accepted replay.
It checks complete core.ts JavaScript byte equality with comments kept and
removed before any write. No compiler overlay or refusal suppression is used
for the ordinary 79-root types check. The census overlay is measurement only.

Proof collection is in progress. PROOF.md and the retained evidence will state
completed checks and any limitations before this branch is offered for review.
