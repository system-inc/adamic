# 99z: elementAt keeps its existing fallback

The count-authorized fallback is implemented. All corpus, type, emission,
replay and native-census checks below pass. One requested assertion remains
false: the original and fallback do not agree for unobserved fractional offsets
and NaN. Those witness results are retained as a failing check, not reported as
all-class equality. The clarification asked during Step A remains pending.

## Ruling and ledger entry

Status: JavaScript-changing adaptation under Adamic's road (b), task b5w3ycg,
Nov 6 21:49Z, admitted only on the completed count. Lib es2022 is refused.
Adamic's lib ruling, task axochk, Nov 1 16:06Z: "tsc is checked against the lib
its own tsconfig declares." The upstream lib remains es2020.

Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Base: origin/main d0279c2bb with 99r amendment 980870fa retained, on the same
feature branch with plain pushes. No compiler or lane-runner source is changed.
Location: core.ts:1076, elementAt's top-level exported const initializer.

The patch drops the feature-test condition and native-call arm, retaining the
existing fallback arrow with its exact body. Two blank source lines preserve
later refusal coordinates. No new cast, local, truncation or lib change is added.
Both original as any casts disappear. The adapter reports javascriptChanging
true and pins the complete exact before/after owner and hashes in owners.json.
A missing, duplicate, nested, changed or half-adapted owner refuses before any
write. LF and CRLF survive, foreign owners survive, and replay preserves bytes
and mtime with zero writes. The 12 owner-contract checks pass.

This is an internal helper and both public declarations are byte-identical.
Named off-count difference: native at truncates fractional offsets and treats
NaN as zero; the unchanged fallback does neither. On [10,20], 0.5, -0.5 and NaN
return 10 before and undefined after. No issue has been filed upstream.

## Count before editing

COUNT-ELEMENTAT.md and ELEMENTAT-COUNT.json retain the completed Step A checkpoint.
All 398 lane shards ran before the initializer edit; every raw call records its
offset and array length. The whole stage3/lane/run.sh control passed, and its
exact test multiset and raw baseline diff matched that shard union.

| Class | Calls |
| --- | ---: |
| In-range non-negative integer | 414 |
| In-range negative integer | 1,138 |
| Out-of-range integer | 217 |
| Non-integer | 0 |
| Array undefined | 167 |

Total: 1,936. Observed offsets are only -1, 0 and 1. All exceptional caller sites
and their counts are in the Step A report and JSON. Seeded calls, diagnostic
retries and the failed 30-second deadline attempt are explicitly excluded.

## Checks

- Both complete before and after applies exit 0.
- Paired replay on independent complete copies: all other adaptation outcomes
  are identical; only 99z applies once before and performs zero edits after.
- Adamic types receives all 79 sorted compiler roots together before and after;
  both exit 0 with no diagnostics. The compiler is ordinary refreshed main.
- Complete core JavaScript, comments kept and removed, changes only in the
  initializer's full span. All nine emitted compiler JavaScript files are
  checked: only _tsc.js and typescript.js differ, at that same initializer.
- Both typescript.d.ts and tsserverlibrary.d.ts are byte-identical.
- All 398 fallback lane shards pass the normal identity, baseline and API
  guards: 106,366 passing tests, one sanctioned API failure, zero skipped.
  The full identity multiset and raw baseline diff match the counted before run.
- The original-initializer mutant returns core.ts:1076:100 in both ordinary
  binder rows. The fallback returns core.ts:1326:26 in both rows.
- The paired per-file ordinary census measures all 79 roots in groups of at
  most 20, each committed and pushed. Its inspector is byte-identical to the
  ordinary tool's; batching changes only the command argument loop.

The fallback lane uses the runner-supported 90-second measurement deadline;
all semantic, identity, baseline and API guards remain unchanged. Every oracle
piece is below 25 minutes. Checkpoint writes are serialized with the native
census; completed lane pieces are individually committed and plainly pushed.

## Ordinary per-file census

| Roots | Before | After |
| ---: | --- | --- |
| 73 | Refused `core.ts:1076:100` | Refused `core.ts:1326:26` |
| 2 | accepted `-` | accepted `-` |
| 1 | NotYet `debug.ts:243:59` | NotYet `debug.ts:243:59` |
| 1 | NotYet `semver.ts:71:9` | NotYet `semver.ts:71:9` |
| 1 | NotYet `commandLineParser.ts:293:35` | NotYet `commandLineParser.ts:293:35` |
| 1 | NotYet `diagnosticInformationMap.generated.ts:13:45` | NotYet `diagnosticInformationMap.generated.ts:13:45` |

The 73's new source is getOwnValues, core.ts:1326:26:

```ts
values.push((collection as MapLike<T>)[key]!);
```

castProof constructs "a cast the runtime can't check" at
internal/lower/cast_proof.go:188. Its union-write gate is :294 and refusal return
:295; castUnionWrites begins at :340 and returns false at :349 when a member
property is absent or cannot receive the union's write. MapLike<T> has no named
length property, while T[] does. This branch explanation is source inspection,
not an instrumented compiler trace. next-stop.json and cast_proof.go.txt retain
the exact source. No Adamic compiler-source overlay or refusal suppression is used; the batch
tool overlay changes only its entry point and keeps inspect byte-identical.

## Node witnesses and remaining requirement

The Node program extracts both actual complete owners and toOffset from the
full-applied trees. It checks that the fallback body is exact, runs each on the
same inputs, and records results and Object.is equality. All four observed
classes agree. Both infinities also agree. Fractions and NaN on [10,20] differ:

| Offset | Original | Fallback | Equal |
| --- | --- | --- | --- |
| 0.5 | 10 | undefined | no |
| -0.5 | 10 | undefined | no |
| NaN | 10 | undefined | no |

The witness process exits 1 for this unmet all-class equality requirement.
PROOF-ELEMENTAT.json records allRequestedChecksPass false for that reason.
There is no Math.trunc change, because the requested body must remain unchanged
and a truncation repair needs its own ruling. No empty-array example is used to
hide the nonempty-array counterexamples. Acceptance of these off-count witness
differences is the outstanding clarification, not a claimed passing assertion.

## Evidence

The original MEASURE-ELEMENTAT.md and COUNT-ELEMENTAT.md are historical checkpoints.
Raw Step A pieces, native census groups, fallback lane pieces and Step B checks
are packed in tar.gz archives with internal SHA256SUMS and adjacent archive
SHA-256 files. The Step B archive retains full types logs, exact JavaScript
hunks (with a display copy in elementat-javascript-hunk.md), both declaration hashes, paired replay outcomes, Node witnesses, binary
hashes and the exact lowering source. Every retained Go source is .go.txt.
All reproduction scripts are in this directory. No message or issue was sent
upstream. All task worker processes are stopped at closeout.
