# 79: truthful exit codes

System.exit returns number | undefined under the Oct 19 #se1zcgr ruling.
Two compiler files change only annotations and emit byte-identical JavaScript.
Two dedicated harness exit stubs replace noop; two consumers use object tests.
The census loses exactly 20 sys.exit reads, with no new read: 491 to 471.
The full lane passes: 106,366 passing, one sanctioned API failure, zero pending.

Branch: typescript/stage3-adapt79-exit-code, cut directly from origin/main
23ff023b9ed986834d20404d43e2c363667edf9b. Adaptation 79 was free.
Starting material: blocked branch 12432d746daee34aa8ebd2fe88f104fc744882ec.
The marker-only branch is separate; neither marker nor global noop is changed.
The source pin is TypeScript 6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8. Stock TypeScript comes from
stage3/api's locked install. The adapter follows 47-host-errors and
76-truthful-casts, resolves all 20 calls to System.exit with the checker,
validates every planned edit before writing, and is idempotent.

## Rewrites

Locations are before and after in the tree after adaptations 1 through 76;
none of these edits inserts a newline. [exit.patch](proofs/exit.patch) records
the complete source diff, so the surrounding functions can be reviewed.

| File:line | Before | After |
| --- | --- | --- |
| src/compiler/sys.ts:1432 | System.exit: void | number \| undefined |
| src/compiler/sys.ts:1588 | Node exit: void | undefined |
| src/compiler/executeCommandLine.ts:760 | void \| SolutionBuilder<…> \| WatchOfConfigFile<…> | number \| undefined \| SolutionBuilder<…> \| WatchOfConfigFile<…> |
| src/harness/fakesHosts.ts:131 | exit: void | number \| undefined |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts:1283 | exit: void | number \| undefined |
| src/harness/harnessLanguageService.ts:563 | exit = ts.noop | exit = (_exitCode?: number): number \| undefined => undefined |
| src/testRunner/unittests/tsserver/session.ts:41 | exit: ts.noop | exit: (_exitCode?: number): number \| undefined => undefined |
| src/testRunner/unittests/helpers/tscWatch.ts:78 | watchOrSolution | watchOrSolution: typeof watchOrSolution === "object" ? watchOrSolution : undefined |
| src/testRunner/unittests/tscWatch/incremental.ts:77 | if (closer) | if (typeof closer === "object") |

The complete 15-definition inventory is in
[testdata/exit-implementations.json](testdata/exit-implementations.json).
Node's exit method falls through to undefined when profiler shutdown is
asynchronous; otherwise process.exit does not return. It is not never.
The two throwing harness methods accept numeric-returning replacements, so
number | undefined is their mutable host signature. Four test overrides assign
and return their optional numeric argument. The two noop stand-ins return
undefined, and the dedicated stubs retain that result. harnessIO forwards to
System.exit but its separate void IO interface discards the result.
The other named exit functions belong to the traversal state machine, timer,
or server session and do not implement System.exit. Timer.exit: noop is left
alone. No other noop-as-System.exit implementation exists in the source inventory.

Only the two stubs and two consumer files have changed transpileModule output.
The other two harness files change annotations alone. The consumer functions
retain every other token. No test input, baseline reference, test registration,
runner, or assertion is changed by this adaptation.

## Public API

The sole added pair is at built/local/typescript.d.ts:8490:

```diff
-        exit(exitCode?: number): void;
+        exit(exitCode?: number): number | undefined;
```

stage3/lane/generate-api.cjs reconstructs that exact System member using
[api-owner.json](proofs/api-owner.json); the generated sanctioned-api.json
retains every existing owner and adds one adaptation-79 owner. The lane unit
test updates its exact declaration count from 222 to 223 and requires that owner. The normalization
key is member:["ts","System","exit"]#0. No method conversion or other API
change is sanctioned. The exact emitted API and baseline diff are checked
against current main with only that replacement. The lane's normalizer also
checks all composed declarations against the manifest.

## Census

The census script is byte-for-byte from 2670c4e7, SHA-256
8eeac14d3eaf0e6484b39a8b33a4be1ce572855585c0e917470f2aae7e7a2f06.
The main census has 491 pure-void reads; this tree has 471, with no new site.
Union-with-void reads fall from 16 to 12, with no new site. The marker read at
checker.ts:23401:28 stays present. Each removed call now has checker type
number | undefined; [results.json](proofs/results.json) records expressions,
locations and before/after types. The 20 executeCommandLine.ts locations are:

```text
579:16 584:16 589:16 594:16 599:16 605:20 613:24 620:24 632:24 643:20
663:24 667:20 707:20 833:16 839:16 845:16 850:16 899:12 938:12 965:12
```

## Proof commands

Every test writes directly to a log. Raw logs and census JSON are retained
compressed under proofs/logs. The scratch source control is /tmp/adapt78-before,
whose stage3 inputs match this branch's main tip exactly; compare.py checks
that Git difference. /tmp/exit79-proof-tree is its isolated source copy with
locked node_modules linked. No new .a program is introduced.

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh > /tmp/exit79-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
P=stage3/adapt/79-exit-code

timeout 120 python3 stage3/lane/generate-api.py /tmp/adapt78-upstream > /tmp/exit79-sanction.log 2>&1
timeout 120 node "$P/adapt.cjs" /tmp/exit79-proof-tree > /tmp/exit79-adapt.log 2>&1
timeout 60 node "$P/proofs/inventory.cjs" /tmp/adapt78-before /tmp/exit79-inventory.json > /tmp/exit79-inventory.log 2>&1
timeout 300 node "$P/proofs/check.cjs" /workspace/adamic /tmp/adapt78-before /tmp/exit79-proof-tree /tmp/exit79-census /tmp/exit79-check.json > /tmp/exit79-check.log 2>&1
timeout 150 node "$P/proofs/check-scope.cjs" /tmp/adapt78-before /tmp/exit79-proof-tree /tmp/exit79-scope-mutant "$P/proofs/source-scope-mutant.json" > /tmp/exit79-scope-mutant.log 2>&1
timeout 120 node "$P/proofs/source-scope.cjs" /tmp/adapt78-before /tmp/exit79-proof-tree "$P/proofs/source-scope.json" > /tmp/exit79-source-scope.log 2>&1
timeout 30 node "$P/proofs/host-witness.cjs" /tmp/adapt78-before /tmp/exit79-proof-tree /tmp/exit79-hosts.json > /tmp/exit79-hosts.log 2>&1
timeout 180 node "$P/proofs/diagnostics.cjs" /tmp/adapt78-before /tmp/exit79-proof-tree /tmp/exit79-diagnostics.json > /tmp/exit79-diagnostics.log 2>&1
timeout 300 node "$P/proofs/check-diagnostics.cjs" /tmp/adapt78-before /tmp/exit79-proof-tree /tmp/exit79-diagnostic-mutant /tmp/exit79-diagnostic-mutant.json > /tmp/exit79-diagnostic-mutant.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream PYTHONDONTWRITEBYTECODE=1 timeout 300 python3 stage3/test_apply.py > /tmp/exit79-apply-tests.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 180 python3 stage3/lane/test_check.py > /tmp/exit79-lane-check-tests.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adapt78-upstream timeout 1800 bash stage3/lane/run.sh /tmp/exit79-lane > /tmp/exit79-lane.log 2>&1
timeout 180 node "$P/proofs/census.cjs" /workspace/adamic /tmp/exit79-lane/adapted-tree /tmp/exit79-lane-census /tmp/exit79-lane-census.json > /tmp/exit79-lane-census.log 2>&1
timeout 60 python3 "$P/proofs/compare.py" /workspace/adamic /tmp/adapt78-before /tmp/exit79-lane "$P/proofs/lane-comparison.json" > /tmp/exit79-comparison.log 2>&1
```

The first uncached census/check run used timeout 900; each census subprocess
has its own 180-second limit. Subsequent runs validate all root and prelude
hashes before reusing exact observations. Setup: Node ready 0.059s, Go 0.080s,
markdown dependencies 0.169s, submodules 0.172s, clang 0.517s, shared cache
2.156s, Go build 31.972s, test binaries deferred 32.136s, build cache warm
32.140s, total 32.207s. nproc is 5;
cgroup CPU quota is 4. Compiler transpile checks pass with comments retained
and removed. Diagnostics: 171 before, 171 after, none introduced or removed.
Source scope: 746 files, 738 unchanged; exactly four harness files change emitted JavaScript.
Apply tests: 23 passed in 27.243s. Lane checker tests: 37 passed in 109.868s,
including the existing malformed observation/API/count mutants. The adapter
rerun makes zero edits.

## Mutants

Each experiment restores its source in finally before the next check.

- Drop only the System.exit rewrite: all 20 exact void reads return, 471 to 491.
- Widen System.exit to unknown: exact checker return-type guard rejects it.
- Widen System.exit to any: the same guard rejects it.
- Add a fresh void-valued read: census detects exactly one new site.
- Append a runtime console.log to sys.ts: transpile byte comparison rejects it.
- Append a runtime console.log to executeCommandLine.ts: byte comparison rejects it.
- Restore the truthiness close guard: real Node return 1 throws TypeError.
- Drop the object selection before runWatchBaseline: code 1 incorrectly survives.
- Return _exitCode from the language-service stub: code 1 violates undefined behavior.
- Return _exitCode from the server stub: code 1 violates undefined behavior.
- Add a compiler-project number/string assignment error: diagnostics check rejects TS2322.
- Append an unrelated comment to core.ts: the exact source-set check rejects it.
- Append an extra line to the API diff: exact baseline byte comparison rejects it.

The upstream ledger records the real Node counterexample. It observes failure
of the old narrowing for an explicit input; it does not claim the upstream
suite itself already reaches that input in the close consumer.

## Full lane result

```text
PASS stage3 landing lane [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]
```

Observed: 106,366 passing, one sanctioned API baseline failure, zero pending.
All runners, no test filter, eight workers. Install passed in 7.586s; the full
compiler and harness build passed in 66.684s; tests took 700.565s. Total lane
wall time including numbered apply: 1167.759s. No passing-count sanction changed.

The composed API has 223 declarations versus main's 222. The actual diff has
195 added and 194 removed lines versus main's 194 and 193. Its 51,093 bytes
are exactly the diff reconstructed from main's built API after replacing only
System.exit. SHA-256:
5333568a661aea24fc8af3149e9f4d2dfca70731842461c677f385b092adc4f0.
The nine built compiler JavaScript files are byte-identical to main. The
harness run.js changes as expected. The lane reproduces the isolated proof's
exact eight-file source patch, with no other source change.

A fresh run of the original census on the lane's actual built tree reports
471 pure-void reads and 12 union-with-void reads. Its sites, union sites and
all 79 source root hashes exactly match the isolated proof. This is recorded
in [lane-census.json](proofs/lane-census.json), with raw JSON and logs retained.
The two required mutants and eleven additional proof mutants are all caught.
The 37 existing lane checker cases also pass, including their planted invalid
counts, missing/extra/changed declarations, API comments and malformed diffs.
No new .a files require a-check headers. No internal/ path is changed.
