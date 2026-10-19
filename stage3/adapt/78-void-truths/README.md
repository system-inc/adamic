Blocked proposal: both ruled rewrites are implemented, but the full upstream build fails in six harness consumers.
The unchanged census falls from 491 to 470 reads: all 21 targets removed, no new site.
All six changed files transpile identically and all ten built JavaScript bundles match the control byte for byte.
All 23 apply tests, the Adamic witness and eleven census/type/erasure mutants pass.
The required 106,366-pass full oracle is not established; the manifest proposal and exact blockers are retained here.

# Void truths

Base: origin/main `141fec7702d4aed21429724b2afd7efcd980f277`.
Branch: `typescript/stage3-adapt78-void-truths`. Number 78 was free.
Evidence read first: `2670c4e7`, `stage3/census/void-reads/README.md` and its
JSON table. That corpus observed numeric exit values 0, 1, 2 and 4, and true
from the marker forwarding call. Observations do not prove arbitrary public
host implementations; the adapted signature states the host contract.

Run `node adapt.cjs <tree>` with stage3/api's stock TypeScript 6.0.3 on NODE_PATH.
The adapter uses the stock parser and checker, resolves all 20 return calls to
System.exit, identifies the marker by checker symbol identity, validates its
writers and reference count, plans every edit before writing, and checks
full-file transpileModule equality with comments both retained and removed.
A second run must write zero files. apply.sh discovers it in numerical order.
Nothing under internal/ changes.

## Rewrites

Locations below are in the pre-78 fully adapted tree, and remain the same
lines afterward. No runtime text is changed.

| File:line | Before | After | Reason |
|---|---|---|---|
| src/compiler/sys.ts:1432 | System.exit(exitCode?: number): void | System.exit(exitCode?: number): number \| undefined | Returning test hosts return the optional numeric argument. |
| src/compiler/sys.ts:1588 | nodeSystem.exit(...): void | nodeSystem.exit(...): undefined | The profiler's asynchronous path returns before process.exit; the other path terminates. |
| src/compiler/core.ts:1794 | noop(_?: unknown): void | noop(_?: unknown): undefined | Empty body; used directly as exit by language-service and server test hosts. Internal, absent from typescript.d.ts. |
| src/compiler/executeCommandLine.ts:760 | void \| SolutionBuilder<...> \| WatchOfConfigFile<...> | number \| undefined \| SolutionBuilder<...> \| WatchOfConfigFile<...> | Forward the newly visible result; this export is internal and absent from typescript.d.ts. |
| src/harness/fakesHosts.ts:131 | exit(...): void | exit(...): number \| undefined | Throwing implementation yields no result, but the mutable method must also allow replacement by returning test callbacks. |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1283 | exit(...): void | exit(...): number \| undefined | Same contract; changing to never would reject the existing numeric-returning replacements. |
| src/compiler/checker.ts:2127 | ((onlyUnreliable: boolean) => void) \| undefined | ((onlyUnreliable: boolean) => boolean) \| undefined | Every installed handler returns true or forwards the saved boolean handler. |

The harness/testRunner files therefore do require type-only changes. Their
statements stay byte-identical. The four assigning overrides need no edits.
No host is made to discard its actual result or return an invented value.

## All exit implementations

[testdata/exit-implementations.json](testdata/exit-implementations.json) retains
all 15 actual definitions/overrides found by an AST walk of every .ts file
under upstream src, including harness, testRunner, server and library files.
It includes full implementation bodies and the pre-78 adapted locations.

| System host implementation | Actual completion |
|---|---|
| compiler/sys.ts:1588 nodeSystem | process termination, or undefined while profiler stop finishes asynchronously |
| harness/fakesHosts.ts:131 | throws processExitSentinel after storing exitCode |
| testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1283 | throws Error after storing exitCode |
| testRunner/unittests/helpers/tsc.ts:55 | returns assigned exitCode, number or undefined |
| testRunner/unittests/tsc/cancellationToken.ts:57 | same |
| testRunner/unittests/tscWatch/incremental.ts:52 | same |
| testRunner/unittests/tscWatch/listFilesOnly.ts:12 | same |
| harness/harnessLanguageService.ts:563, testRunner/unittests/tsserver/session.ts:41 | ts.noop returns undefined |
| harness/harnessIO.ts:139 | forwards ts.sys.exit; its separate IO interface still discards the result through void |

The remaining definitions are different contracts: compiler/factory/utilities.ts
exit returns the numeric traversal stack index; compiler/performance.ts exit
returns undefined (or Debug.fail throws), and nullTimer.exit is noop;
server/session.ts's empty exit returns undefined and nodeServer.ts's override
terminates with process.exit(0). They are inventoried, not retagged as System
hosts. System's number | undefined admits every completed host result. never
alone would lose the deferred Node path and returning overrides; number alone
would lose the omitted argument and no-op hosts. void would hide actual values;
any or unknown would promise less than the established contract.

proofs/hosts.cjs executes extracted source implementations on Node for omitted
codes and 0, 1, 2, 4, 97. Its profiler scheduling is an explicit model of
immediate/deferred continuations, not a real inspector session. Both throwing
methods and all four assigning overrides are extracted, rather than recreated.

## Every marker reader

The marker has 15 symbol references. The direct calls at checker.ts:2130 and
2136 are expression statements, so their results remain discarded. Conditions
at 2129, 2135, 23331, 23397 and 23423 read the presence of the callback, not a
void call result. originalHandler's typeof declaration at 23395 follows the
refinement. Saves at 23398 and 25042 retain the callback; restorations at 23424
and 25058 retain its type. The wrapper installed at 23399 forwards its saved
handler at 23401. The callback installed at 25043 returns the true assignment
in either branch. The declaration itself is at 2127. The existing non-null
assertion remains; this unit does not claim a new proof of its lifetime.

## Public API

Exactly this declaration changes:

```diff
-        exit(exitCode?: number): void;
+        exit(exitCode?: number): number | undefined;
```

[proofs/api-addition.diff](proofs/api-addition.diff) records its exact hunk.
[proofs/sanctioned-api.json](proofs/sanctioned-api.json) reconstructs the existing
222 normalized sanctioned declarations from main's proof generator, then adds
only System.exit, making 223. It preserves prior allowances from 20, 30, 32,
33, 40, 70 and 75. No reference baseline is accepted or edited by 78.

The brief first permits stage3/lane/sanctioned-api.json but finally requires
every commit to touch only this directory. The latter restriction is followed:
the proposed full manifest and its exact patch are delivered here, and lane's
unchanged API predicate validates it in isolation (not a full lane pass). The repository's default manifest
remains unchanged; its unextended verdict must reject the new declaration.

## Proof commands and results

The measured results below include failures; this proposal is not landing-ready. Every test writes
to a log, and every long command has an explicit timeout. The original census
script in proofs/census.cjs is byte-for-byte from 2670c4e7, including its original
discard treatment; no language-rule change is smuggled into the census. Its
fixture roots and library overlays are outside internal/.

### Census and mutants

The unchanged census SHA-256 is
`8eeac14d3eaf0e6484b39a8b33a4be1ce572855585c0e917470f2aae7e7a2f06`.
The control reproduces exactly 491 sites, 16 union-with-void sites and 480,940
runtime expressions. After 78: 470 sites, 12 union-with-void sites and the same
expression count. All remaining pure-void sites have the same file, line,
column, expression and form as before. There are no new union-with-void sites.
No refined-discard adjustment is applied to the borrowed census.

The twenty exit sites, each now number | undefined, are at
executeCommandLine.ts **579:16, 584:16, 589:16, 594:16, 599:16, 605:20,
613:24, 620:24, 632:24, 643:20, 663:24, 667:20, 707:20, 833:16, 839:16,
845:16, 850:16, 899:12, 938:12, 965:12**. The marker forwarding call at
checker.ts **23401:28** is now boolean. Locations are identical before/after.
[proofs/results.json](proofs/results.json) records every expression and type.

| Mutant | Catch |
|---|---|
| Restore System.exit's original void declaration | Census returns to 490, restoring all 20 exit sites at exactly those locations; marker remains repaired. |
| Restore marker callback's void return | Census becomes 471, restoring only checker.ts:23401:28. |
| Widen System.exit to any | Adapter rejects the checker's return type before writing. |
| Widen System.exit to unknown | Same exact-type rejection. |
| Append one new void read | Census becomes 471, adding checker.ts:54437:34, adaptation78VoidMutant(). |
| Append runtime console.log independently to each of the six changed files | Each full-file transpileModule byte comparison fails, both with and without comments. Six catches. |

All mutants run only in an isolated source copy or virtual text. Production
oracle sources never receive them. The final source copy is restored. The Node
witness prints `0 1 2 4 undefined` and `true true true true` on two lines.
The `.a` witness passes `adamic types` and `adamic c`, each exit 0, in place.
Per a-check-headers/README.md, a clean program requires no expected-error
header. Its body hash and measured status are in proofs/a-check.json.

The first proof run hit its 360-second limit during the marker mutant. That
isolated source was restored with the adapter before the final run. The final
720-second run passed, reusing only census records whose source root hashes
and prelude hash match, under the pinned census bytes and TypeScript 6.0.3.
The first stock diagnostic comparison ran before candidate Node declarations
were installed and failed with missing-node diagnostics; this prerequisite
is now explicit and the installed comparison is reported separately.

### Full-build blocker

The real `stage3/lane/run.sh` completes apply (exit 0), then `npm run build`
fails (exit 1). Compiler and service declaration builds complete; the tests'
declaration build reports exactly these six errors:

| Location | Diagnostic | Exposed assumption |
|---|---|---|
| src/harness/incrementalUtils.ts:655:5 | TS2322 | service.verifyDocumentRegistry was inferred from noop as returning undefined; withIncrementalVerifierCallbacks returns void. |
| src/testRunner/parallel/shared.ts:85:28 | TS2352 | The existing noop-to-ExclusiveSuiteFunction cast promises Suite. |
| src/testRunner/parallel/shared.ts:87:53 | TS2352 | The existing noop-to-PendingTestFunction cast promises Test. |
| src/testRunner/parallel/shared.ts:88:22 | TS2352 | The existing noop-to-ExclusiveTestFunction cast promises Test. |
| src/testRunner/unittests/helpers/tscWatch.ts:78:13 | TS2322 | WatchOrSolution excludes the now-visible numeric result. |
| src/testRunner/unittests/tscWatch/incremental.ts:77:32 | TS2339 | `if (closer)` excludes zero but does not exclude a nonzero number before `.close()`. |

These are real full-project failures, not waived diagnostics. No gate-passing
truthful type-only repair has been established for these dependent consumers.
In particular, broadening WatchOrSolution alone cannot justify calling close
on a number. The existing harness casts also demonstrate why globally refining
noop needs a wider review than the compiler-root census. No additional hiding
assertion, any/unknown widening, checker-option relaxation, or runtime guard
has been added to make this build look green.

Consequently the standard candidate oracle does not reach tests, and the
required 106,366 passing / one sanctioned failure result is NOT claimed.
The control's 840-second outer limit also expires before its report is written;
its remaining independently grouped test workers were explicitly terminated.
Its partial test log is retained, without invented totals. An API-only run on
the existing built bundles is separately identified below and does not replace
the required full build/oracle.

### Commands and measured results

All paths below are the actual scratch paths. Commands run from the repository
unless a working directory is stated. `source /workspace/adamic-tools/env.sh`
selects Node 24.19.0, Go 1.27.1 and clang 20.1.8. Proof commands set
`NODE_PATH=/workspace/adamic/stage3/api/node_modules`.

| Command | Result |
|---|---|
| `export GOPROXY='https://proxy.golang.org\|direct'; timeout 600 bash cloud/setup.sh > /tmp/adapt78-setup.log 2>&1` | exit 124 during cold cache warming |
| `export GOPROXY='https://proxy.golang.org\|direct'; timeout 840 bash cloud/setup.sh > /tmp/adapt78-setup-retry.log 2>&1` | exit 0, 462.491s; nproc 5, cpu.max 400000 100000 |
| `timeout 180 npm ci --prefix stage3/api > /tmp/adapt78-api-install.log 2>&1` | exit 0, stock TypeScript 6.0.3 |
| `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream timeout 600 bash stage3/apply.sh /tmp/adapt78-before > /tmp/adapt78-before-apply.log 2>&1` | exit 0, main control through adaptation 76 |
| `timeout 90 node stage3/adapt/78-void-truths/adapt.cjs /tmp/adapt78-proof-tree > /tmp/adapt78-adapt.log 2>&1` | exit 0, seven edits / six files, JavaScript identical |
| `timeout 720 node stage3/adapt/78-void-truths/proofs/check.cjs /workspace/adamic /tmp/adapt78-before /tmp/adapt78-proof-tree /tmp/adapt78-census /workspace/adamic/stage3/adapt/78-void-truths/proofs/results.json > /tmp/adapt78-check-final.log 2>&1` | exit 0, idempotence, 21 targets, five census/type mutants and six erasure mutants |
| `timeout 60 node stage3/adapt/78-void-truths/proofs/hosts.cjs /tmp/adapt78-proof-tree stage3/adapt/78-void-truths/proofs/hosts.json > /tmp/adapt78-hosts.log 2>&1` | exit 0, nine extracted implementation/model observations |
| `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream PYTHONDONTWRITEBYTECODE=1 timeout 300 python3 stage3/test_apply.py > /tmp/adapt78-apply-tests.log 2>&1` | exit 0, 23 tests, 89.762s |
| `timeout 300 go build -o /tmp/adapt78-adamic ./cmd/adamic > /tmp/adapt78-adamic-build-retry.log 2>&1` | exit 0 after successful setup; an earlier redundant cold build was terminated to reduce contention |
| `timeout 30 /tmp/adapt78-adamic types stage3/adapt/78-void-truths/testdata/values.a > /tmp/adapt78-a-check.log 2>&1` | exit 0, clean |
| `timeout 30 /tmp/adapt78-adamic c stage3/adapt/78-void-truths/testdata/values.a > /tmp/adapt78-a-lower.log 2>&1` | exit 0, clean; no expected-error header required |
| `timeout 180 node stage3/adapt/78-void-truths/proofs/diagnostics.cjs /tmp/adapt78-before /tmp/adapt78-lane/adapted-tree /workspace/adamic/stage3/adapt/78-void-truths/proofs/diagnostics.json > /tmp/adapt78-diagnostics-installed.log 2>&1` | exit 0, compiler-project diagnostics 171 -> 171, none introduced; this excludes the harness project |
| `timeout 60 node stage3/adapt/78-void-truths/proofs/api.cjs /tmp/adapt78-before /tmp/adapt78-lane/adapted-tree /workspace/adamic/stage3/adapt/78-void-truths/proofs/api.json > /tmp/adapt78-api-proof.log 2>&1` | exit 0, exactly System.exit changed, ten built JS artifacts identical, reference unchanged, API widening mutant caught |
| `timeout 90 python3 stage3/adapt/78-void-truths/proofs/sanction.py /workspace/adamic /tmp/adapt78-upstream /workspace/adamic/stage3/adapt/78-void-truths/proofs/sanctioned-api.json > /tmp/adapt78-sanction.log 2>&1` | exit 0, prior 222 normalized declarations preserved, exactly one added |
| `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream timeout 1200 bash stage3/lane/run.sh /tmp/adapt78-lane > /tmp/adapt78-lane.log 2>&1` | exit 1, 881.591s; apply 0, install 0, build 1; six harness errors, no test phase |
| `timeout 840 bash stage3/oracle/run.sh /tmp/adapt78-before /tmp/adapt78-control/oracle --workers 2 --limit-seconds 720 > /tmp/adapt78-control-oracle.log 2>&1` | outer exit 124; no complete report or totals |

The check runner invokes the retained census as
`node proofs/census.cjs REPOSITORY TREE SCRATCH OUTPUT`, each subprocess with
a 180-second hard limit and its own log. The roots/prelude hashes and all raw
before/after/mutant tables are retained under proofs/logs/. The final source
hashes in the isolated proof and the numbered apply output match for all six
changed files. proofs/rewrites.patch records every exact before/after line.

Setup retry timing lines: Node 0.117s, Go 0.217s, markdown 0.421s (step 0.032s),
submodules 0.493s, clang 1.437s, shared cache 3.189s, Go build 461.720s,
test binaries deferred 462.066s, cache warm 462.086s, done 462.491s.
Initial setup: Node 0.162s, Go 0.183s, clang 0.832s, markdown 1.757s
(install step 1.387s), submodules 312.374s, shared cache 325.019s, then timeout.
Both complete logs are retained. No full Adamic repository gate or native tsc
execution is claimed.

### Actual API-only observation

From `/tmp/adapt78-lane/adapted-tree`:

```sh
source /workspace/adamic-tools/env.sh
timeout 180 node /workspace/adamic/stage3/oracle/run-prepared.mjs /tmp/adapt78-lane/adapted-tree --workers=4 --light=false --runners=unittest --tests='Public APIs' --no-colors > /tmp/adapt78-api-tests-final.log 2>&1
```

Observed exit 1: **30 passing, 1 failing**, only the expected typescript.d.ts
acknowledgment. This uses already built bundles and explicitly bypasses build;
it is not a replacement full oracle. Earlier invocations used the wrong cwd
(no test file) and then the plural runner name (zero tests); neither is counted
as validation. The final log is retained.

The unchanged lane `api_check` predicate was invoked on this actual reference,
actual local snapshot, and the exactly reconstructed API baseline.diff. It
passes with the proposed manifest: 223 composed / 223 sanctioned declarations,
28 reference declarations, 195 added and 194 removed diff lines. The main
manifest mutant fails with exactly
`unsanctioned API declaration: member:["ts","System","exit"]#0`.
This is API-only validation, not a successful `check_results` full-lane verdict.

proofs/api-diff-delta.json separately compares the control build's actual API
to this observed snapshot. Every old added/removed line is preserved; the only
two additional diff lines are the exact System.exit pair shown above. The
control had 194 added / 193 removed lines; this proposal has 195 / 194.
The built declaration is line 8490 before and after. The actual API-only diff
is proofs/api-baseline.diff. The exact manifest patch is
proofs/sanctioned-api.patch and is deliberately not applied outside this unit.

Decision on the unresolved scope: retain the six build errors as blockers and
ship this branch as a proposal with evidence. A full 106,366-pass oracle and a
landing-ready adaptation have not been produced. The branch is not ready for
integration. No PR is opened, and main is not changed.
