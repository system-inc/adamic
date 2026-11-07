# takeWhile builds its proven prefix

Task b5w3ycg, wave row 11. Complete, under the counted established-shape ruling.
Main base: `f0271c6a7b2e92900eb7046d8246cfb89f52cfb6`. TypeScript 6.0.3 pin:
`050880ce59e30b356b686bd3144efe24f875ebc8`.

Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts.
The user's continuation clarifies the scope: the same values on every input tsc
gives, established by a complete lane count. The accepted sparse-hole witness
is outside that counted input class, with frequency 0. The ledger records both
the ruling and this difference. docs/0.1.md:137 describes native array append
and existing-index writes; native arrays cannot represent a hole this way.

## Change and ownership

Site F42 is the prefix cast at adapted core.ts:2595:16, Refused "a cast the
runtime can't check" (internal/lower/cast_proof.go:188). Both overloads stay
unchanged. The implementation creates `const result: U[] = []`, reads each
leading element once, breaks when its predicate rejects, pushes the now-narrowed
U, increments index, and returns result. The cast disappears. The bounded read
retains adaptation 30's existing `!`; no new cast or non-null assertion is added.
Undefined input still returns undefined. Rejection still stops the first time.

Read 99s first. It leaves takeWhile unchanged. The required pre-write search of
all adapt.cjs, owners and sites JSON found only 30-indexed-reads/sites.json's
indexed-read site, not a whole-owner takeWhile pin. ownership.txt is retained.
No sibling owner, import, compiler backend, lane runner or expectation is edited.

The adapter pins the complete before/after owner, including both overloads, in
adapt.cjs and owners.json, with exact normalized texts and SHA-256 hashes.
Changed, half-adapted, missing, duplicate, moved and nested owners refuse before
any write. Parse failures refuse too. Foreign source is preserved byte for byte,
including mixed line endings outside the owner. First apply is one file and one
edit; replay is zero writes with bytes and mtime unchanged. No whole-file hash
is an adaptation gate. The real lexical directory sort places this after all
99-correctness-axolotl-* adapters.

## Complete count before construction

All 398 unchanged lane manifest units ran once in the canonical count through
the normal shard identity, API, baseline and count guards. No test was filtered.
There are 106,366 passes, one expected API failure, and zero skips. Twenty
pieces have at most 20 units, four concurrent units, and the runner-supported
90-second deadline; the largest canonical piece was 33.583 seconds.

| Input class | Calls |
| --- | ---: |
| Undefined array | 0 |
| Empty array | 0 |
| All accepted | 0 |
| None accepted | 215 |
| Rejection after an accepted prefix | 5 |
| Accepted range contains a hole | 0 |
| Total | 220 |

| Caller in the complete adapted source | Predicate | Calls |
| --- | --- | ---: |
| compiler/transformers/ts.ts:1086 | isExportOrDefaultModifier | 88 |
| services/refactors/moveToFile.ts:252 | isPrologueDirective | 132 |

Every call records the input slots, predicate argument order and answers,
accepted indices missing under `i in array`, and both results. The scratch
original body is unchanged inside its measurement wrapper. The candidate
replays the original predicate answers against an input snapshot, checking
actual Object.is argument identities without calling tsc's predicate twice.
Both bodies run in the same VM realm. Input slots remain unchanged in every
call. Result length, prototype, own keys, slot presence and element identity
match on all 220 calls. Result differences: 0. Predicate-sequence differences:
0. Accepted-range-hole calls: 0. Input-slot mutations: 0.

Independent reconciliation reopens every archive, verifies both checksum
layers, reconstructs raw call classes and identities, and checks every report
against the unchanged inventory. The unchanged stage3/lane/run.sh whole control
passes and has exactly the same test identity multiset and raw baseline diff.
All 746 applied source files match the uninstrumented counted baseline. Stock
TypeScript printer comparisons also prove that the two measured bodies match
the two actual full-applied owners. build-adapter.json timestamps construction
after the complete verified count. COUNT.json and COUNT-VERIFIED.json retain
the full result. Recorder seeds are excluded from the canonical count.

## Node witnesses and the named off-count difference

witness.cjs extracts both actual full-applied owners and executes their erased
bodies. actual-witness.json records all seven classes and predicate sequences.
Undefined, empty, all accepted, none accepted, middle rejection, and a rejected
sparse hole have equal results and sequences. An accepted sparse hole has the
same predicate sequence but a different property layout:

`new Array(1)` with `value => value === undefined` yields length 1 on both sides.
The original slice preserves the missing property 0. The push loop creates own
property 0 containing undefined. `0 in result` changes from false to true.
This witness is retained, not relabeled as equal. The complete count finds 0
calls in this accepted-hole class. Both observed input classes also have equal
Node witnesses. witness.json retains the initial witness, and PREVIOUS-STOP.md
retains the superseded stop before the user's scope clarification.

## Proof results

See PROOF.json, comparison.json, JAVASCRIPT.md and CENSUS.md.

| Proof | Result |
| --- | --- |
| Complete main and main-plus-ours applies | Both exit 0 |
| Owner contract | 10 no-write refusals; foreign text and line endings preserved |
| Own replay | Zero edits and writes; bytes and mtime preserved |
| Every adapter's paired replay | All other outcomes identical; only ours applies versus no-op |
| Final replay source comparison | Byte-identical on both sides |
| Adamic types, all 79 sorted roots together | Exit 0, zero diagnostics, before and after |
| Source comparison | Only pinned takeWhile owner changes across all 746 src files |
| Full-core JavaScript, comments kept and removed | Only takeWhile changes |
| All nine emitted compiler JavaScript files | Only _tsc.js and typescript.js differ, only at takeWhile |
| typescript.d.ts and tsserverlibrary.d.ts | Both byte-identical |
| Full core.ts census | 59 to 58 sites; F42 gone; no new site anywhere |
| Original-owner mutant | F42 and the complete original site set return |
| Complete built adapted lane | 398 units pass normal guards; identities and baseline match count |

Paired replay records ordinary main's existing adapter refusals on both sides;
it does not pretend every older adapter is independently replayable. Their
complete stdout/stderr and exits are in replay.json, with the matching nonzero
adapter names in replay-summary.json. Adamic types prints proven declarations;
its successful logs are not empty logs. The site meter still defers takeWhile's
generic body, so no successful generic lowering is claimed.

## Reproduction and evidence

Setup used `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`,
`npm ci --prefix stage3/api`, the pinned TypeScript source through
ADAMIC_TYPESCRIPT_SOURCE, and the locked stage3/typescript build packages.
The proof drivers use the recorded /workspace paths and fresh output folders.
Before is a full apply from an isolated checkout of the frozen main base.
prepare-count.py, count-pieces.py and reconcile-count.py run measurement before
adapter construction; whole-control.py runs the unchanged whole-lane control.
prepare-proof.py, run-proofs.py, compare.cjs, check-types.py, census.py,
analyze-census.py, replay.py and after-lane.py reproduce the remaining proofs.

count-01 through count-20 archives hold raw canonical calls, reports, observations
and logs. proof.tar.gz holds setup, both applies, whole control, emission hunks,
public declaration hashes, raw site streams, original-owner mutant, source and
binary manifests, contract results, recorder seeds and the scratch-only no-output
overlay. adapted-lane.tar.gz holds every adapted lane report and observation.
Each archive has internal SHA256SUMS and an adjacent SHA-256 file. All retained
Go sources are .go.txt. No compiler source overlay is installed in production.
Evidence changes product keys; the adapter and both measured source manifests
remain frozen and the exact owner transformation is unchanged.

Preparation-only failures are retained separately and excluded: the first
baseline apply raced a queued API reinstall; the fresh retry ran after setup
and installs completed. A first shard preparation collided with its package
symlink and recorded zero calls. A superseded host-realm recorder also recorded
zero calls. A proof script initially shadowed Python's types module; it was
renamed before the accepted checks. The template contract exposed mixed line
ending changes outside an owner; the final adapter preserves those bytes and
passes the same test. No failed result or expectation was relabeled.

The prior pushed stop commit 28106340 is preserved. This continuation adds one
implementation commit, with plain pushes and no push to main. The separate
post-apply checkpoint is a0ba7618e1e2e4ffe3164fd7d3b392134b64833b on
`typescript/core-takewhile-apply-checkpoint`. All task commands have exited;
the final live-process audit is retained with the proof evidence.
