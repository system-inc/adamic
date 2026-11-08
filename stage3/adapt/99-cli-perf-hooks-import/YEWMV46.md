# PerformanceTime.timeOrigin, task yewmv46

This is an amendment to 99-cli-perf-hooks-import, not a new adaptation.
Base and census compiler: `cd42fd2126b63e1647f6ee2e6c3dd2c87bdda2fe`. Its lowering code is identical
to 97e3615c; their differences are Lower's tests and pin. Upstream source is
TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Ownership and exact source

Every adapt.cjs and every owner/site JSON was searched. The only existing
member owner is this adapter. The other match is
47-host-errors/evidence/optional-owners.json:3264: PerformanceHooks.performanceTime
has the type PerformanceTime. It does not pin this interface or timeOrigin.
The amended owners.json pins the complete interface's exact CRLF text and
SHA-256 before and after. The after pin includes readonly. Five drifted or
half-applied cases fail before writing; a second run writes zero files and
preserves mtime. An unrelated function is preserved.

The base already emits readonly. Against the actual base, the only source
change is consumption of the blank line following the first import while the
node:perf_hooks import is added. This preserves the pinned range at lines
13 through 28. Stock 6.0.3 emits byte-identical JavaScript with comments kept
and removed for this spacing change. The readonly comparison and census use
a completed adapted tree with only `readonly` removed; the mutant is a second
copy with exactly the same writable source bytes.

## Count

The final stock 6.0.3 language-service find-references pass is count-after-node.log.
It loads 603 non-lib src roots, including all harness and testRunner files,
and 824 source/declaration files. It uses the complete module-resolution
host and explicit pinned Node types. Its seeds are PerformanceTime.timeOrigin
and @types/node/perf_hooks.d.ts:94's Performance.timeOrigin. The local Performance
inherits the first member; Node's Performance supplies the imported value.
All source types assignable to PerformanceTime are examined for member symbols,
class/object providers, computed member accesses and writes through aliases.

**Writes: 0. Reads: 2.** The complete read list is:

- performanceCore.ts:72:28, `typeof performance.timeOrigin`.
- performance.ts:184:48, `timeorigin = perfHooks.performance.timeOrigin`.

Assignments, compound assignments, destructuring targets, increments/decrements
and alias member writes have zero timeOrigin write references. All eight
compatible-receiver member accesses are reads. All ten Object.defineProperty,
Reflect.defineProperty, Reflect.set or Object.assign candidates in src were
inspected; none targets a performance value. An unrestricted src/tests spelling
scan finds no occurrences in tests, harness or testRunner. src/lib contains two
further readonly timeOrigin declarations, in DOM and worker library definitions.

Source object literals and classes providing PerformanceTime: **0**. There is
no source-created timeOrigin field to adapt. tryGetPerformance obtains the Node
performance object from perfHooks or the ambient browser performance value;
its two returned literals wrap that value rather than implement PerformanceTime.
tryGetPerformanceHooks assigns the value to hooks.performanceTime after its
number/function checks. Readonly permits an ordinary-field provider as well.
Node v24.19.0's descriptor has a getter and no setter; assignment throws TypeError.

## Type-only checks

| Check | Result |
| --- | --- |
| Stock 6.0.3, writable and readonly, Adamic strictness/module/target options | 0 diagnostics on each side |
| Stock 6.0.3, exact types-empty Adamic options plus its embedded Node index | 0 diagnostics on each side |
| Stock 6.0.3, upstream compiler options | 0 diagnostics on each side |
| Upstream npm run build, actual base / writable / readonly | Exit 0 on all three |
| Adamic types, all 79 roots, actual base / writable / readonly | Exit 0 on all three, 0 checker diagnostics |
| Nine emitted compiler JavaScript files, all three builds | All byte-identical |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical; PerformanceTime is not public |
| typescript.internal.d.ts, actual base to amendment | Byte-identical, because the base already has readonly |
| typescript.internal.d.ts, writable to readonly | Only line 7971 changes as below |

```diff
-        timeOrigin: number;
+        readonly timeOrigin: number;
```

comparison.json and baseline-comparison.json hold every output hash; the exact
internal declaration diff is retained. No public declaration line moves.

## Full census and mutant

All runs load all 79 roots, report checker_rejected=false and zero diagnostics,
set LATENT_FULL=1 and LATENT_ONLY_FILE to the complete adapted performanceCore.ts,
and keep LATENT_ASSERT_NO_OUTPUT=1. Every one of its 12 units is attempted.

The production host table still has pending Runtime/Platforms clearances and
performance's pending model. Ordinary table results are also retained: 10, 9,
and 10 findings for writable, readonly and mutant. They cannot expose the
requested Provided-row declaration error while that table stop wins.

The requested Provided-row proof therefore uses exactly routedPerformanceRows
from cd42fd21's internal/lower/host_binding_route_test.go: only performance's
native and wasm32 columns become Provided, and its model path is node:perf_hooks.
performance-row.go.txt installs that fixture only in the scratch measurement
overlay when LATENT_YEWMV46_PROVIDED=1. No production source, host table,
lowering decision, file selector, attempt guard or no-output guard is changed.

| Variant | Units | Findings | Declaration row |
| --- | --- | --- | --- |
| Writable | 12 | 6 | NotYet at 28:15 |
| Readonly | 12 | 4 | Gone; nothing new |
| Mutant, readonly removed | 12 | 6 | The same NotYet returns |

Row: unit performanceCore.ts:28:1, phase refusal_scan, kind NotYet,
location performanceCore.ts:28:15:

> the host binding performance, declared as Performance, whose field timeOrigin is writable where its model export node:perf_hooks.performance's is read only,

The full text ends `yet`. The related invariant-mutable Refused at 39:21 also
disappears. The other four findings are unchanged. The mutant restores the
entire original finding set, including the declaration row.

## Apply, replay and equality unit

Both full keyed applies exit 0. All 746 source files agree apart from the
single consumed import separator in performanceCore.ts. The post-apply checkpoint
is `9e50f3ea4cb346985e9c54d69b60244db7b1f685`, pushed on typescript/core-performance-time-apply-checkpoint.
All 97 paired replay outcomes match after normalizing checkout paths,
including unchanged refusals from 14 earlier adapters on already
fully adapted trees. This is a replay equivalence result, not a claim that those
older adapters all accept a second run. The complete outcomes and names are in
replay.json and replay-summary.json. The replayed sources retain only this
amendment's import separator difference.

The stage3 unit is TestPerformanceTimeYewmv46AdaptedProductMatchesLowerPin.
It calls adaptedtree.Product, as existing consumers do, with the product built
by TestProduct_Stage3AdaptedTree ahead of the unit. It is t.Parallel and starts
no Go or compiler subprocess. The proof runs the compiled unit in
ADAMIC_BUILD_CACHE=read with a 30-second test deadline and traces process starts.

The comparison replaces only CRLF with LF on the adapted side. It joins exactly
lines 13 through 28 including their newline bytes, and compares against Lower's
pin without trimming, other whitespace normalization or a trailing-newline
allowance. Lower's pin stays unchanged. unit.py checks three red mutants under
that same test name: one pin byte, readonly removed from the adapted product,
and one space changed in those 16 source lines. It restores every byte and
checks green again. The measured unit results are recorded in LEDGER.md after
this evidence archive freezes the adaptation inputs and the final product key
has been built; adding result files here would itself rekey that product.

## Evidence and reproduction

proof-yewmv46.tar.gz contains raw logs, child exit records, reference details,
output hashes/diff, complete census rows and timings, replay outcomes, input
manifests, the exact Go unit as equality-test.go.txt, and the proof drivers.
Its SHA-256 is in proof-yewmv46.sha256. The archive excludes compiler binaries,
source checkouts and generated .go overlays; its Go evidence uses .go.txt.

The drivers use /workspace/adamic and /workspace/scratch/yewmv46. Recreate the
base worktree at cd42fd21 and the pinned upstream checkout, source
/workspace/adamic-tools/env.sh, run bash cloud/setup.sh and npm ci --prefix
stage3/api, and install stage3/typescript's pinned build dependencies. Set
ADAMIC_TYPESCRIPT_SOURCE to that upstream pin and NODE_PATH to stage3/api/node_modules.
The recorded run.py command arrays reproduce both --build-product applies.
proofs.py drives the writable/readonly checks; count.cjs uses stock findReferences;
compare.py and compare-main.py compare outputs; make_overlay.py plus the recorded
provided-overlay.json mapping and performance-row.go.txt reproduce the census;
provided.py checks all three runs; replay.py reproduces the paired replay.
For a cold local cache, first settle TestProduct_Stage3AdaptedTree through
internal/buildcache/cmd/traced with ADAMIC_BUILD_STORE=traced; read mode
deliberately does not use the untraced local cache. The ledger records the
exact preparation command and its result. unit.py then reads that prepared
product and runs the equality unit and its mutants in read-only product mode. All wrappers report child exit codes,
cap output silence at 30 minutes, and clean their process groups.
