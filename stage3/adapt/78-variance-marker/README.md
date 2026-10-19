Adaptation 78 changes only outofbandVarianceMarkerHandler from void to boolean at checker.ts:2127.
Fresh census: 491 -> 490, exactly checker.ts:23401:28 removed, no new site; all 16 void-union sites unchanged.
Full lane PASS: 106,366 passing, one sanctioned API failure, zero pending; its diff is byte-identical to main.
Checker JavaScript, ten built JS artifacts and both public declaration artifacts are byte-identical; diagnostics remain 171 -> 171.
All 23 apply tests and six mutants pass; the blocked sys.exit branch and the sanctioned manifest are unchanged.

# Variance marker return truth

Branch: typescript/stage3-adapt78-marker, cut directly from current origin/main
141fec7702d4aed21429724b2afd7efcd980f277. Adaptation 78 is free on main.
The narrower directory name is stage3/adapt/78-variance-marker.

The single change, at the same line before and after:

```diff
-    var outofbandVarianceMarkerHandler: ((onlyUnreliable: boolean) => void) | undefined;
+    var outofbandVarianceMarkerHandler: ((onlyUnreliable: boolean) => boolean) | undefined;
```

The producer at checker.ts:25043 returns true from either assignment. The wrapper
installed at 23399 returns the saved handler's boolean at 23401. It keeps the
existing non-null assertion; this unit does not add a lifetime assertion.

The adapter uses stock TypeScript 6.0.3 from stage3/api, validates symbol identity,
all fifteen references, the four writers/restorations and both discarded direct
calls, then edits the parsed return-type span. It verifies full-file JavaScript
identity with comments both retained and removed before writing. The CLI is
`node adapt.cjs <tree>` and a second run must write zero files.

Every marker reference is accounted for: the declaration at 2127; presence
conditions at 2129, 2135, 23331, 23397 and 23423; discarded calls at 2130 and
2136; originalHandler's typeof at 23395; saves at 23398 and 25042; installations
at 23399 and 25043; restorations at 23424 and 25058. The forwarding call at
23401 is the one census read whose result type changes.

Only checker.ts changes in the external TypeScript source tree. System.exit,
executeCommandLine, noop, harness and testRunner source bytes stay unchanged.
No file under internal/ or stage3/lane is edited. No new .a program is added.

Proof commands, measured results and every mutant are recorded below.

## Measured local proofs

The fresh unchanged census from 2670c4e7 visits 480,940 expressions and records
491 void reads before, 490 after. Exactly checker.ts:23401:28 is removed:
`originalHandler!(onlyUnreliable)` is void before and boolean after. There are
no new sites, and all 16 union-with-void sites remain identical. Every other
745 file under upstream src is byte-identical to the current-main control.

Both full-file transpileModule comparisons pass:

| Comments | Emitted bytes | Before and after SHA-256 |
|---|---:|---|
| retained | 2,892,419 | 786ee138ff00765ab7c7ea7651de48df32c5981ec2c0810c68e302570d566f6e |
| removed | 2,354,716 | 3664a2f5fb2aba0e839a3153e6ff4d093b530db890b5b2433be6cd38e6e91aac |

Idempotence passes: a second adapter run writes zero files and preserves the
checker source bytes. The independent compiler-project check records **171 ->
171 diagnostics, none introduced**, with identical installed Node declarations
on both trees. This does not claim all existing compiler diagnostics are fixed.
All **23 apply tests pass**, 17.628 seconds.

The pre-78 control is /tmp/adapt78-before, already applied from the same current
main commit 141fec77. Its fresh census reproduces the original evidence exactly.
The isolated mutant source tree is /tmp/marker78-proof-tree. Mutants never touch
the source used by the full lane. The census script is byte-for-byte from
2670c4e7, SHA-256
8eeac14d3eaf0e6484b39a8b33a4be1ce572855585c0e917470f2aae7e7a2f06.
It is invoked without --adaptations, preserving its original discard treatment.

## Commands

Commands run from /workspace/adamic after sourcing
/workspace/adamic-tools/env.sh. Node is 24.19.0, Go is 1.27.1, clang is 20.1.8.
Node proof commands use
`NODE_PATH=/workspace/adamic/stage3/api/node_modules`, stock TypeScript 6.0.3.
All test output goes directly to logs, never through a pipe.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh > /tmp/marker78-setup.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream PYTHONDONTWRITEBYTECODE=1 timeout 300 python3 stage3/test_apply.py > /tmp/marker78-apply-tests.log 2>&1
timeout 120 node stage3/adapt/78-variance-marker/adapt.cjs /tmp/marker78-proof-tree > /tmp/marker78-adapt.log 2>&1
timeout 900 node stage3/adapt/78-variance-marker/proofs/check.cjs /workspace/adamic /tmp/adapt78-before /tmp/marker78-proof-tree /tmp/marker78-census /workspace/adamic/stage3/adapt/78-variance-marker/proofs/results.json > /tmp/marker78-proofs.log 2>&1
timeout 180 node stage3/adapt/78-variance-marker/proofs/diagnostics.cjs /tmp/adapt78-before /tmp/marker78-proof-tree /workspace/adamic/stage3/adapt/78-variance-marker/proofs/diagnostics.json > /tmp/marker78-diagnostics.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream timeout 1800 bash stage3/lane/run.sh /tmp/marker78-lane > /tmp/marker78-lane.log 2>&1
```

Setup, apply tests, adapter, local proof runner and diagnostics each exit 0.
Each census subprocess has its own 240-second limit and log. The full-lane
result and its exact verdict line are recorded below.

Setup timing lines: Node 0.056s, Go 0.091s, submodules 0.149s, markdown 0.163s
(validation step 0.023s), clang 0.507s, shared cache 1.744s, Go build 24.503s,
test binaries deferred 24.648s, cache warm 24.651s, done 24.703s. nproc is 5;
cgroup cpu.max is 400000 100000 (four CPUs). There is no setup failure or retry.

## Mutants

| Mutant | Result |
|---|---|
| Drop the boolean rewrite | Caught: exact original 491-site census restored, including only the removed forwarding read at checker.ts:23401:28. |
| Add one void initializer read | Caught: one extra site at checker.ts:54437:28, marker78VoidMutant(). |
| Append runtime console.log to checker.ts | Caught: full-file emitted bytes differ, with comments both retained and removed. |
| Change the reviewed producer's true assignment to false | Caught: adapter rejects writer drift before writing. This is a conservative source-provenance guard. |

The mutant runner restores the source in finally blocks and checks its final
bytes. Every listed mutant is actually run; results are in proofs/results.json.
No sys.exit mutant or unrelated rewrite is carried into this unit.

## Full lane and final comparison

The complete unfiltered lane exits **0**. Its exact result line is:

```text
PASS stage3 landing lane [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]
```

Upstream reports **106,366 passing, 1 failing, 0 pending**. The sole failure is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The underlying oracle exits 1 because of that deliberately sanctioned baseline;
the unchanged lane accepts precisely that failure and exits 0. Apply, npm
install and the full compiler/harness/test declaration build each exit 0.
No --prepared path, narrowed runner set, type-check bypass or new API allowance
is used.

The actual baseline.diff is **50,593 bytes**, SHA-256
`3da14ced9f2c7eea0ae8b4dabbe6067a91510b0ab0d950ddb83903a9eb50fa28`.
It is byte-for-byte equal to the diff reconstructed from the current-main
control's actual built typescript.d.ts and actual reference snapshot. This
also matches the original 2670c4e7 evidence hash. The main control is reused
from the preceding unit because origin/main remains exactly 141fec77; its
fresh source census is independently reproduced in this unit.

All ten built JavaScript files, typescript.d.ts and tsserverlibrary.d.ts match
that control byte for byte. The reference snapshot is unchanged. The full
numbered apply output changes only src/compiler/checker.ts compared with the
main control, by exactly the one annotation shown above. It matches the
isolated source proof. stage3/lane/sanctioned-api.json is byte-identical to main,
and the blocked branch still points to 12432d746daee34aa8ebd2fe88f104fc744882ec.

Final proof commands (same env.sh and NODE_PATH as above):

```sh
timeout 300 node stage3/adapt/78-variance-marker/proofs/check-diagnostics.cjs /tmp/adapt78-before /tmp/marker78-proof-tree /tmp/marker78-diagnostic-mutant /workspace/adamic/stage3/adapt/78-variance-marker/proofs/diagnostic-mutant.json > /tmp/marker78-diagnostic-mutant.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 120 python3 stage3/adapt/78-variance-marker/proofs/compare.py /workspace/adamic /tmp/adapt78-before /tmp/marker78-lane /workspace/adamic/stage3/adapt/78-variance-marker/proofs/lane-comparison.json > /tmp/marker78-comparison.log 2>&1
```

Both exit 0. The diagnostic mutant adds a number initialized with a string in
an isolated checker copy. The unchanged diagnostics check exits 1 with exactly
one new TS2322, **171 -> 172**. The runner restores the source and verifies it.
The baseline mutant appends extra text to the actual diff bytes and is rejected
by the exact comparison with main. These are the fifth and sixth mutants;
the four earlier mutants are listed in the table above. Every mutant was run.

The oracle phase times are install 6.657s, build 48.984s,
tests 633.886s; oracle wall 689.696s. Full lane wall 1009.899s.
The structured results, unchanged baseline diff and compressed complete logs
are retained under proofs/. There is no failed or abandoned validation run in
this marker-only unit. Final syntax and whitespace checks are also run before
the adaptation-only commit. No PR is opened; only the new marker branch is
pushed once. No native tsc execution or whole Adamic repository gate is claimed.
