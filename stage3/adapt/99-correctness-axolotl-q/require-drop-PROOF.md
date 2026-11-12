# 99q require-test amendment evidence

Task #nqp4p7g, 99q half; Task #00qn5az. Base is main
730e7c26e590af0c57c545ba0cb905b09e4bee54. TypeScript is pinned to v6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8. Node is v24.19.0 on Linux x64.
This amendment changes JavaScript. The earlier PROOF.md describes original
99q's type-only edits, not this amendment.

## Scope and pins

Only H05's after text changes: remove the typeof require conjunct and terminate
the process.browser line with a semicolon. Its after SHA-256 is
876351683806aed43e71d4dcc9039c80399490ac01fc1c9f2c8b3c65486e2748.
Both adapt.cjs and owners.json carry it. They were the only files naming the
old after hash. The exact-owner gates still refuse drift before writing.
The adapter's JavaScript comparison restores H05 solely for that comparison,
retaining the type-only check on the other three owners. The contract now
expects javascriptChanging: true.

The pre-edit stage3/adapt grep found these outside references:

- 30-indexed-reads: whole-files.cjs pins the ambient header and function marker,
  not the function body; both remain unchanged. Its README and whole-sites.json
  describe that earlier header adaptation.
- 99-cli-perf-hooks-import: adapt.cjs pins the performance guard, which remains
  `if (isNodeLikeSystem())`. Its historical witness assumes the require test
  exists and has a recorded row removing it in memory. The source witness here
  follows its owner-extraction/full-performanceCore approach without that old
  assertion. No CLI pin changes.
- 99-correctness-axolotl-r/main-runtime-proof.json is historical equivalence
  evidence, not a gate for this new JavaScript-changing amendment.
- 99-correctness-core-multimap/owners.json names the sys import, not this owner.
- 65-temporary-node-builtins/evidence, 99-correctness-axolotl-k's before/after
  JavaScript, and 32-indexed-reads-program/record-tables-measure-plan.json are
  historical snapshots, not pins on the current function body.
- ACHECK.md records a historical compiler refusal.

None requires an outside adaptation edit. No harness pin is added.

## Accepted whole-lane count

Adamic ruling (1), #nqp4p7g, accepts this measurement as evidence, following
adaptation 47's meta-spread-lane.json and the tracing-assert PROOF.md precedent.
It is not a counter checked automatically on future lane runs.

| Observation | Count |
| --- | ---: |
| Lane cases passing generateCpuProfile in any form | 1 |
| enableCPUProfiler function entries | 0 |
| getNodeSystem function-entry control | 9 |

require-drop-counts.json retains all nine per-process records, the actual input
record and test title, command/phase results, and the generated case's passing
outcome. The input is the showConfig case quoted in README.md. It passes
`["--generateCpuProfile", "someString"]` directly to parseCommandLine, converts
the result to TSConfig, writes the baseline, and returns without calling
executeCommandLine. Its 0 profiler entries do not make the real CLI path
unreachable. A real command with --showConfig and --generateCpuProfile enters
the profiler before executeCommandLineWorker handles showConfig. The bare
require throws `require is not defined` under the ES-module oracle. The ruled
load divergence is backend_divergences.json rows 2 and 3,
node-without-inspector, #cn4yapb.

The literal, case-insensitive input search covered all 19,585 files under
upstream tests/cases, plus src/testRunner and src/harness, 19,905 files total.
It found 0 literal occurrences. Inspection of the reference baselines exposed
the generated option case, whose option name comes from ts.optionDeclarations.
Scratch-only parser probes then recorded actual command-line arguments and
parsed raw compilerOptions, including response-file parseStrings calls.
The complete run recorded exactly one input case, not zero.

After full apply, a complete scratch source copy received one call at the first
statement of getNodeSystem and one at the first statement of enableCPUProfiler.
Both call a preload emitter that synchronously writes a JSON counter file named
by PID after every entry. Each of the nine processes recorded one control and
zero profiler entries; summing those files after completion gives 9 and 0.
The preload also exposes the emitter to compiler code evaluated in a VM, as in
stage3/census/debug-assertion-level/REPORT.md on the reference census branch.
The parser observations are the only other source probes. All 747 source files
were compared: exactly sys.ts and commandLineParser.ts differ, two calls each.
The repository adapter and ordinary applied tree contain no probes.

The unchanged stage3/oracle/run.sh ran all runners, eight workers, no filter,
with the ordinary timeout. Install and build exited 0; tests/oracle exited 1.
There were 106,366 passes, one sanctioned Public APIs failure, and zero pending.
The only baseline difference was api/typescript.d.ts. The unchanged lane
checker returned PASS. Mocha's worker and host events contain 212,732 passes
and two failures, all multiplicities even; the showConfig case passes once
when divided by two. An initial build overlapping a corrected probe anchor was
discarded before this complete measurement; no partial counts were spliced in.

Raw measurement logs, preload, instrumenter, per-process counters and oracle
output are retained under /tmp/99q-proof in the task workspace. The accepted
measurement output is measured-oracle, not measurement-setup-attempt.

## Require audit

require-drop-requires.json classifies every `require(` occurrence in all 79
compiler TS files: 32 comment occurrences, 2 strings, 4 actual calls, 0 dead
calls. The calls are:

| Applied source | Path and handling |
| --- | --- |
| sys.ts:1593, source-map-support | Returned tryEnableSourceMapsForHost method; existing catch ignores failure |
| sys.ts:1615, modulePath | Returned system.require method; existing catch returns the error record |
| sys.ts:1646, inspector | Returned enableCPUProfiler method; bare require outside a try/catch, covered by the cited ruling; measured entries 0 |
| tracing.ts:66, fs | startTracing; existing try/catch rethrows the tracing-specific error; not an isNodeLikeSystem caller |

getNodeSystem is reached first through the now-true guard. Its fs, path and os
bindings are static node: imports from adaptation 93; crypto is a static import
from 99-wasm-crypto-import. Its remaining requires are the three deferred methods
above. performanceCore's other isNodeLikeSystem call uses CLI's static
node:perf_hooks import, with its existing try block intact. Those are the only
two callers of isNodeLikeSystem. No other unhandled bare require is opened.

The whole-tree grep also includes upstream corpus/baseline text, documentation,
shipped CommonJS code, build scripts and the Node test harness. The compressed
require-drop-whole-requires.json.gz records every hit by file, line, column,
category and context; its summary states the exact totals. Root .git,
node_modules and built output are excluded, while nested corpus node_modules
are included. Fixture source and baselines are input strings to the harness;
infrastructure calls run in their own CommonJS or virtual-host context, not on
the compiler's newly enabled ES-module branch. Compiler JSON diagnostic strings
are included in the whole-tree totals in addition to the TS-only audit above.

## Remaining validation

The final validation results and identity/baseline counts are recorded in
require-drop-validation.json. The restoring mutant script is
require-drop-mutant.cjs, and require-drop-mutant.json records its run. It
restores the require line only in a temporary copy of H05's after text,
recomputes that temporary owner hash, applies it to the real pre-99q snapshot,
and observes false through oracle/node.mjs. The finally block removes the
mutation. The script's recorded witness lines name Node v24.19.0 and false.

Reproduction uses the environment printed by cloud/setup.sh and
GOPROXY='https://proxy.golang.org|direct'. Setup completed with exit 0, including
submodules. Tools: Go 1.27.1, Node v24.19.0, clang 20.1.8. The lane fixture check
uses original 99q's documented command:

```sh
go test -count=1 -parallel=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/
```

The whole upstream lane is independently covered by the complete oracle runs;
the Go command above checks the lane's fixtures rather than rerunning its 398
shards. No production Go file, checker option, per-test or oracle timeout, expected test outcome,
public API sanction, or outside adaptation is changed.


## Final results

Both main and tip ordinary oracle runs have 106,366 passing, one sanctioned
failing and zero pending tests. Both unchanged lane checks return PASS.
The exact title/outcome multisets match, including multiplicity: zero additions
and removals. The normalized sorted event/title/count JSONL has SHA-256
61b98459713bf84d52f00222ecbafc36adfe60279ef9cbcd665c4778d1e83c63.
The measuring run has the same multiset, so its probes preserve case identity.

Each of main, tip and the measuring run reads **60,933 reference baselines**,
covering **60,931 distinct reference files**. Their complete path/byte/hash
multisets, including read multiplicity, are identical. These are observed reads
of expected baseline files, not a count of unexecuted files in the corpus.
The only output baseline mismatch is the sanctioned api/typescript.d.ts.
require-drop-identity.json retains these counts and comparisons.

Both public declarations are byte-identical; require-drop-public-dts.json
records lengths and hashes. All 79 compiler roots pass adamic types with exit 0
and zero diagnostics. Both full applies exit 0. The 99q replay has zero files,
zero edits and no writes; its contract proves unchanged bytes/mtime and 29
refusals before any write, plus foreign edits and both newline styles.
generate_shards --check passes.

In addition to the fixture command above, the **unfiltered** Go command passes:

```sh
go test -count=1 -parallel=2 -timeout=60m ./stage3/lane/
```

For the Go product, stage3/typescript/package.json and package-lock.json were
byte-identical to the oracle tree's. Its locked dependency install was linked
temporarily at stage3/typescript/node_modules, and the link was removed after
the run. A standalone reproduction can install that lockfile with
`npm ci --prefix stage3/typescript --no-audit --no-fund`.

This includes all 398 manifest shards, using its independently built ordinary
adapted-tree product, with no counter instrumentation. Package time is
2784.981 seconds, including product preparation. No successful subsets are
spliced together. The restoring mutant returns false on Node v24.19.0, and
its temporary mutation is removed. require-drop-validation.json collects every
result and command. New proof data and prose were finalized after these checks;
the adapter code, owner pins, and generated runtime source stayed fixed.
