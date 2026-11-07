# core-clone: type-only cast removal

Task b5w3ycg, wave rows 4 and 5. Rule class: type-only.
Site: F5, the (object as any)[id] read in clone.

The for-in variable has type Extract<keyof T, string>, so object[id] is already a checked indexed read of T. Removing the as any assertion requires no record constraint, no signature change, and no runtime operation. The existing any result is retained; this adaptation addresses F5 only.

Adaptation 41 documents the cast as unadapted. It does not pin or edit it. No earlier whole-owner or exact-target pin was found.

The adapter accepts exactly the complete before or after owner text, including
all overloads, and refuses drift, duplicate or nested owners, and partial edits.
It preserves line endings and performs zero writes on an accepted replay.
It checks complete core.ts JavaScript byte equality with comments kept and
removed before any write. No compiler overlay or refusal suppression is used
for the ordinary 79-root types check. The census overlay is measurement only.

Proof collection is in progress. PROOF.md and the retained evidence will state
completed checks and any limitations before this branch is offered for review.
