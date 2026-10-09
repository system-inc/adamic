# Two writes measured on stock TypeScript 6.0.3

The flags assignment runs in both requested corpora, but its stock caller has
`Expression.flags: NodeFlags`, not the literal `NodeFlags.Synthesized` promise
used by the review's counterexample. No flags-domain violation was observed.
The parent-clearing assignment was not reached in either corpus. Both generic
API counterexamples are reachable on Node, and a later read sees each bad value.
These are observations on the pinned inputs, not a proof about every tsc run.

The upstream pin is `050880ce59e30b356b686bd3144efe24f875ebc8` (`v6.0.3`).
The starting Adamic main is `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`.
Only this directory is committed. The TypeScript checkout and its two added
observation calls are scratch artifacts; no Adamic or adaptation source changed.

| Population | Flags assignments | Flags violations | Later flags reads | Parent assignments / violations |
| --- | ---: | ---: | ---: | ---: |
| 301 acceptance projects | 61 | 0 | 199 | 0 / 0 |
| Upstream compiler and conformance tests | 171,296 | 0 | 1,340,306 | 0 / 0 |
| Narrow flags witness | 1 | 1 | 1 | 0 / 0 |
| Present parent witness | 0 | 0 | 0 | 1 / 1 |

Both positive witnesses have exactly one out-of-type read. Neither measured
corpus has an out-of-type read at these sites. See [counts.md](counts.md),
[inputs.csv](inputs.csv), [observations.json](observations.json), and
[acceptance-report.json](acceptance-report.json) for the recorded artifacts.
The compiler runner discovered 12,444 files and passed 89,838 tests; compiler
includes the conformance population in this upstream harness. This is the real
upstream harness, including option variants and its type/symbol baseline reads,
not a flat invocation of tsc on raw harness-controlled case files.

## What the counters mean

The sites are the stock assignments in `src/compiler/utilities.ts:10689` and
`:12365`. `instrument.py` checks the source commit and pristine utilities bytes,
then adds one observation immediately after each original store. It preserves
CRLFs and leaves the original assignment in place. `probe.cjs` is a Node preload
shared by the CLI and every upstream harness child process. Each process writes
its own JSON at normal exit, including explicit empty records for unreached sites.

At the parent site, undefined is outside `Node.parent: Node`, regardless of
whether the freshly updated clone's old parent was already undefined. This is
an incompatible assignment, not a claim that a present runtime parent was erased
on every visit. The witness starts with a present parent and shows that the
original remains parented while the returned clone has an undefined parent.
It calls `getSynthesizedDeepCloneWithReplacements` with no replacements, which
uses the shared worker containing this assignment.

At the flags site, a changed numeric value alone is insufficient evidence of a
type violation. Generic `T` is erased in JavaScript. The witness explicitly
registers its legal literal domain `[NodeFlags.Synthesized]` in a WeakMap;
writing `NodeFlags.None` then fails that domain. `typed-witness.a`, checked by
`typed.cjs` using the measured compiler's API, demonstrates stock TypeScript
accepts both narrowed holder promises and the generic calls with zero diagnostics.

For stock internal calls, `static.cjs` resolves the one implementation caller:
`checker.ts:41702`, `setNodeFlags(node, node.flags | NodeFlags.TypeCached)`.
The actual receiver is `Expression` and its declared flags domain is `NodeFlags`.
All bits 0 through 30 occur in that enum; the probe accepts their integral,
nonnegative combinations, as used by this bit-mask API. It does not invent a
literal domain from the runtime object's old flags. Arbitrary external callers
with narrower erased types require explicit domain registration; this probe
cannot discover those types from object values alone.

After a store, an enumerable/configurable getter observes reads of that field's
assigned value. A later store disarms that observation. The getter returns the
unchanged value, and the setter accepts every subsequent value. Provenance
lookups are excluded from read counts. Reads include compiler guards and harness
consumers, not just dereferences that could crash. The experiment changes field
descriptors and timing, so it is observational instrumentation, not a performance
measurement. All CLI golden comparisons and upstream baselines still pass.

## Examples and later consumers

The acceptance input `012_expandoFunctionExpressionsWithDynamicNames2` executes
the flags store twice, then reads flags four times. A sampled assignment writes
268435456 (`TypeCached`); its later reader is `markLinkedReferences` in the
checker. The upstream input `resolvingClassDeclarationWhenInBaseTypeResolution.ts`
also reaches the flags site; a sampled later reader is `isParseTreeNode`, called
by the harness's `TypeWriterWalker`. Full sampled stacks are in observations.json.
There is no corpus example or corpus later reader for the parent site because
its assignment count is zero.

Run either small program with the instrumented library:

```sh
node stage3/evidence/two-writes/flags.cjs "$TREE/built/local/typescript.js"
node stage3/evidence/two-writes/parent.cjs "$TREE/built/local/typescript.js"
```

They print, respectively:

```text
flags: Synthesized (16) -> None (0); later read = 0
parent: present on input; undefined on clone; later read = undefined
```

## Fixtures, mutants and commands actually run

`fixtures.cjs` first parses `const x = 1;` and asserts neither site was reached.
It then executes a compatible real flags store with the literal domain `[16]`.
For parent, the compatible predicate control calls the observer with a real
Node value, since the upstream site itself always writes undefined.

The following mutants were run and caught:

| Mutant | Catcher |
| --- | --- |
| Flags counter ignores its domain check | Compatible flags control: violations 1 instead of 0; exit 1 |
| Parent counter ignores its domain check | Compatible Node-parent predicate control: violations 1 instead of 0; exit 1 |
| Flags read observer disabled | Flags witness: out-of-type reads 0 instead of 1; exit 1 |
| Parent read observer disabled | Parent witness: out-of-type reads 0 instead of 1; exit 1 |
| Typed witness promises string flags | Stock 6.0.3 reports TS2322; the enclosing typed test exits 0 after asserting the error |
| Stock caller virtually narrowed to Synthesized | Static flags-domain assertion sees NodeFlags.Synthesized instead of NodeFlags; exit 1 |

The first four are runtime input modes, not source edits; their complete logs
are under [logs](logs). The last static mutant is a CompilerHost-only source
substitution, leaving the scratch checkout unchanged.

From a fresh stock checkout and the repository root:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/two-writes-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# TREE is a fresh stock v6.0.3 checkout at the commit above.
python3 stage3/evidence/two-writes/instrument.py "$TREE" > /tmp/two-writes-instrument.log 2>&1
(cd "$TREE" && npm ci) > /tmp/two-writes-install.log 2>&1
(cd "$TREE" && npm run build -- --no-typecheck) > /tmp/two-writes-build.log 2>&1
python3 stage3/evidence/two-writes/run.py "$TREE" /tmp/two-writes-final > /tmp/two-writes-final.log 2>&1
python3 stage3/evidence/two-writes/export.py /tmp/two-writes-final > /tmp/two-writes-export.log 2>&1
node stage3/evidence/two-writes/static.cjs "$TREE/built/local/typescript.js" "$TREE" --mutant > /tmp/two-writes-static-mutant.log 2>&1
# The preceding mutant must exit 1 with the flags-domain assertion.
```

The build passed both compiler and test type checks: npm warned that the
`--no-typecheck` argument was an unknown npm config, so it did not disable
Hereby's checks. `run.py` invokes the existing 301-project driver and
`npm test -- --runners=compiler --workers=4 --lint=false`, writing each command's
output to a log. It requires ordinary commands to exit 0 and each runtime mutant
to exit 1 with an AssertionError. The first exploratory run exposed a circular
JSON value in the compatible parent control; the serializer was corrected before
the passing measurements. A final census rerun used the provenance guard and
read-observer mutants; its counts are the recorded ones.

Setup timing lines were: Node ready 0.063s, Go ready 0.086s, clang ready 0.548s,
markdown dependencies ready 1.285s, submodules ready 18.578s, Go build ready
215.475s, build cache warm 215.580s, done 215.614s. `nproc` was 5; cgroup quota
was four CPUs. Setup printed `/workspace/adamic-tools/env.sh`, which was sourced.
Full setup and build logs are retained here.

Not covered: fourslash, server, unit, project and transpile runners; watch/build
modes; arbitrary external generic callers without registered domains; Windows;
native Adamic execution; or a whole repository gate. Zero parent executions do
not establish unreachability outside these corpora. These are local evidence
fixtures, so this directory's counts.md is refreshed; no internal oracle fixture
was registered and internal/oracle/counts.md is unchanged.
