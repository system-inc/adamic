Preserved both for-in guards and added Node-held property-presence and array-hole witnesses.
Base b410340dc8f889b5799c3bc519117c63def3aa24; replay merge 1aa37cf4; first kind bb393fb8.
All three examples replay; final focused oracle passes in 0.697s and counts refresh in 22.431s.
Both guard mutants fail exact-stop assertions: origin bypass accepts; array bypass reaches the origin stop.
No root site is newly lowered: 10 origin sites and 1 array site retained for a ruling; refused witnesses cannot run through backends.

## Origin kind: refused for a ruling, 10 root sites

The diagnostic remains NotYet, not Refused: this unit does not decide the language ruling.
The conservative assumption is that a structural type is not proof of the runtime
property set. docs/0.1.md refuses shape mutation, expandos and prototype mutation,
and leaves reflection as a later question. The existing fixed-literal reflection
is retained. No production compiler source or shared hook changes.

Reviewed every assigned origin site in the exact adapted source:

| Site | Required evidence beyond a structural type |
| --- | --- |
| core.ts:1289:23 | MapLike parameter, arbitrary own keys |
| core.ts:1314:23 | MapLike or array parameter, including array holes |
| debug.ts:432:28 | enum object passed as a structural parameter |
| factory/nodeFactory.ts:6137:25 | SourceFile parameter, copies extra properties into another object |
| commandLineParser.ts:2788:24 | OptionsBase parameter, actual optional property presence |
| commandLineParser.ts:2975:24 | CompilerOptions parameter, actual optional property presence |
| commandLineParser.ts:4265:23 | CompilerOptions parameter, actual optional property presence |
| moduleNameResolver.ts:432:27 | package JSON dictionary keys |
| moduleNameResolver.ts:2428:23 | peerDependencies dictionary keys |
| moduleSpecifiers.ts:927:23 | MapLike paths parameter, arbitrary own keys |

Ruling requested: permit dynamic dictionaries/property presence and sparse-array
reflection with a representation contract, or adapt these uses to explicit Maps
and explicit field copies. A local guard bypass does not supply that contract.
Interprocedural origin proof for safe parameters is also possible future work;
this unit neither establishes such a proof nor rejects it by design.

The new .a fixture prints `stable` for a missing property and `stable|missing`
for a present undefined property. TestForInOptionsRuling independently pins Node's
output and the exact origin diagnostic. It is registered from its own _test.go.
It cannot run through the backends while lowering intentionally stops. Existing
library_for_in.a, library_for_in_keys.a and library_for_in_live.a pass source Node,
backend Node, release native, ASan/UBSan and the leak check, uncached.

## Exact commands and observations

Source preparation used `git archive 9a1f14c5 stage3` into a scratch snapshot,
then its `bash stage3/apply.sh /tmp/notyet-for-in-census-input`. All 81 source
byte lengths and SHA-256 hashes match the replay census manifest. Installed
lockfile-pinned dependencies with `npm ci --prefix stage3/api --ignore-scripts
--no-audit --no-fund`. Compiler branch base is the newest fetched area/compiler;
the unit-specific base overrides the generic main-base instruction. Main is
already an ancestor of this branch at the first push.

Every build/test shell sources `/workspace/adamic-tools/env.sh`. Output goes
directly to logs. Commands:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestForInOptionsRuling|TestNativeAgreesWithNode/internal/oracle/testdata/library_for_in' -count=1 -v
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/notyet-for-in-origin-mutant/overlay.json ./internal/oracle -run '^TestForInOptionsRuling$' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
```

Positive oracle: exit 0, 10.894s. Origin mutant: exit 1, 0.178s,
`want exact origin stop ..., got <nil>`. The mutant changes only forIn's
plainEnumerableObject guard to `if false && ...` through a scratch Go overlay;
no compilation/sanitizer failure counts as its kill. Counts: exit 0, 21.378s,
counts.md unchanged because the added fixture intentionally does not lower.

Each replay uses `go run ./stage3/census/latent/replay -project
/tmp/notyet-for-in-census-input/src/tsc/tsc.ts -where
/tmp/notyet-for-in-census-input/src/compiler/commandLineParser.ts:LINE:24
-kind NotYet -reason 'for...in without a proven fixed plain-object origin
(arrays, prototypes and absent synthetic fields cannot be enumerated soundly)'`.
2788 before, 2975 and 2788 after all exit 0 and reproduce that exact stop.
Full command timings: 14.959s, 2.506s and 2.496s respectively.
These are selected-unit observations on a checker-rejected project, not native
tsc compilation. The other eight origin sites were source-reviewed, not replayed.

Setup succeeded without workaround. Timing lines: Node 0.027s, Go 0.040s,
clang 0.249s, markdown dependency install step 0.994s and ready 1.077s,
submodules 16.447s, Go build 221.748s, tests deferred 221.858s, cache warm
221.859s, done 221.886s. nproc=5; cgroup CPU quota 400000/100000.
Go 1.27.1, clang 20.1.8, Node 24.19.0.

No whole package test or full gate was run. Evidence logs are in evidence/.

## Array kind: refused for a ruling, 1 root site

factory/nodeFactory.ts:7538:23 in mergeTokenSourceMapRanges copies present
properties from sourceRanges to destRanges. Replacing it with an element loop
would copy a hole as undefined and may change the destination's existing entry.
The fixture distinguishes present undefined from a hole: Node prints `0|1|2`
then `0|2`. TestForInArrayRuling pins that output and the exact array diagnostic.
This is a sparse-property representation question, not a choice of loop syntax.
No expansion of the language's fixed-shape contract is made in this unit.
The stop remains NotYet while the ruling is pending; arrays with a proven dense
origin could be future proof work, but the census site's parameter is not proved dense.

Before and after replay commands use the same project as above, with `-where
/tmp/notyet-for-in-census-input/src/compiler/factory/nodeFactory.ts:7538:23
-kind NotYet -reason 'for...in over an array (holes and own enumerable properties
are not represented; use for...of for elements)'`. Both exit 0 and reproduce the
exact array stop. Total times: 2.274s before, 3.043s after.

Array mutant uses a scratch Go overlay replacing only forIn's array test with
`if false && (l.checker.IsArrayType(proven) || checker.IsTupleType(proven))`.
Command: `ADAMIC_GATE_UNCACHED=1 go test
-overlay=/tmp/notyet-for-in-array-mutant/overlay.json ./internal/oracle
-run '^TestForInArrayRuling$' -count=1 -v`. Exit 1, 0.305s; the test expected
the array stop and got the origin stop. This proves diagnostic classification,
not an unsound acceptance: the second guard still protects the program.
The unmutated array witness passed in 0.178s.

Final commands before landing:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestForIn(Options|Array)Ruling|TestNativeAgreesWithNode/internal/oracle/testdata/(library_for_in|for_in_.*_refused)' -count=1 -v
go vet ./internal/oracle
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
git diff --check
```

All exit 0. Oracle 0.697s, counts 22.431s, vet and diff check no output.
counts.md is unchanged. The first-kind push succeeded. Current main advanced
from d65e2d5e to ef3141e9 during work; landing validation follows its merge.
