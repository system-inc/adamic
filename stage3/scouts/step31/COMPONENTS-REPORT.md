Built: stock Node checker diagnostics and emitter text drivers, stable per-project caches and a native process comparator.
Commits: parent 46ca102be59235974da4945d11abdff8b4298f7a; same branch codex/step31-scout; delivery SHA is in the final dispatch.
Observed: both pieces run all 301 acceptance and 6262 selected compiler cases; actual source Node output equals the stock bundle.
Mutants: actual checker/emitter bodies, eight record fields, cache integrity, library identity and independent process streams are exercised.
Not covered: native execution, native ownership, whole-program linking, the full upstream suite or a new own-file census/meter run.

This follow-up implements the unanswered checker side and the emitter side from
the initial scout. All changes remain under stage3/scouts/step31. The pinned
TypeScript source is 050880ce59e30b356b686bd3144efe24f875ebc8 (6.0.3), adapted
at the initial scout's main 45487a8. The original measurement stays in REPORT.md;
no new 78/78 claim is made. The request selection and source hashes come from
the existing acceptance/verdict manifests through prepare-corpus.py, unchanged.

checker-dump.cjs creates a real Program and calls getPreEmitDiagnostics. It
records each input file, including clean files, plus global/library diagnostics.
Records retain code, UTF-16 start/length, category, recursive message chains and
related locations, with a total diagnostic sort. This is tsc's public pre-emit
API and includes option, syntax, global and declaration diagnostics as well as
checker semantic diagnostics. It does not reduce diagnostics to a verdict.

emitter-dump.cjs calls program.emit and records every writeFile text exactly,
with virtual output path, BOM flag, source-file names, emitSkipped and returned
emit diagnostics. The selected corpus produced 6,556 output files and zero emit
diagnostics. This does not mean the corpus is type-correct: the checker produced
20,274 diagnostics. The additional fixture explicitly exercises declarations,
JS, JS maps and declaration maps. Actual emitted bytes inside strings retain
line endings, final-newline presence and Unicode separators.

component-compare.py validates the Node request and output hashes before handing
the exact golden bytes to compare.py. Native stdout, stderr and exit are captured
separately. Both native slots remain unrun; comparator success fixtures use
explicitly labelled Python protocol test doubles. The source comparisons really
execute TypeScript's checker and emitter through stock transpilation of the
adapted source tree, rather than forwarding to the stock compiler bundle.

Measured complete subprocess wall time, Node 24.19.0, nproc=5 with four CPUs
of cgroup quota. These are single observations, not performance guarantees:

| Piece | Full 6563 | Unchanged cache, 0 selected | One changed project, 1 selected |
|---|---:|---:|---:|
| checker | 19.826s | 0.315s | 1.021s |
| emitter | 22.157s | 0.315s | 0.980s |

The one-change benchmark prepends a real comment to the tiny project's first
source. Both diagnostic spans and emitted text change; exactly tiny is selected,
with 6,562 observations reused. Each unchanged run validates and copies the full
current cache but emits an empty request/output pair. One-change time includes
fresh library parsing in its new Node process, hashing and filesystem writes.
Per-project timings in the manifest exclude shared initialization and disk
writes. Acceptance/upstream timing sums are derived from the full run, not
independently warmed benchmark runs:

- checker: acceptance 4.225s; upstream 14.889s.
- emitter: acceptance 4.370s; upstream 17.105s.

Full stable-output fingerprints:

| Piece | Bytes | Records | SHA256 |
|---|---:|---:|---|
| checker | 7374390 | 19691 | d3c4bbd20d1b60b1c78b8cc58a50f7a9911626c17e2a54536570653afc94b4dc |
| emitter | 6175615 | 19682 | b72f09302786fb06b8f2a0a9a938ecf92690e454dd89baa852eaa300a2c4f00e |

Node cache identity covers project source/options/script kind, exact compiler
and library bytes, Node version and driver bytes. Changed-only native comparison
uses the cached delta's request.json and golden.stdout. A native build change
still requires the full cached baseline or a justified affected-input list;
unchanged input alone cannot prove unchanged behavior in a new implementation.
Full Node goldens need not be recomputed for that comparison.

Commands actually run, with output written to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-checker-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export STEP31_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
bash stage3/scouts/step31/run-components.sh /tmp/step31-adapted /tmp/step31-components-final > /tmp/step31-components-final.log 2>&1
```

The final focused command passed fixtures, fresh-versus-reused library checks,
the independent standard-host comparison, all semantic/comparison mutants, full
stock measurements and exact source-versus-stock comparisons for both pieces.
Its benchmark commands and all wall/driver/per-project measurements are recorded
in evidence/components/benchmark.json and the compressed manifests. Earlier
standalone development runs completed the same full stock corpus in 19.97s
(checker driver time) and 22.76s (emitter driver time). The first integrated
reporter failed after successful Node emission because Python splitlines treated
U+2028 as a record separator. It now splits only on LF; the final run and a
Unicode fixture regression pass. The initial failure log is retained.

Setup succeeded: Node ready 0.015s, Go ready 0.017s, markdown ready 0.046s,
submodules ready 0.049s, clang ready 0.115s, Go build ready 27.343s,
cache warm 27.415s, done 27.437s. nproc=5, cpu.max=400000 100000, 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0; environment path
/workspace/adamic-tools/env.sh. JS/Python/shell syntax and whitespace checks
passed. Existing binder fixtures and independent process checks still pass.
No whole-package tests, integration gate, npm suite or PR were run/opened.

Every mutation/probe and its observed catcher:

| Mutation, each actually run | Catcher and observed outcome |
|---|---|
| actual checker.checkTypeAssignableToAndOptionallyElaborate returns true | nested assignment TS2322 disappears; bytes differ from stock; Node exit 0, empty stderr |
| actual emitter.emitVariableDeclaration omits emitInitializer | emitted text differs from stock; Node exit 0, empty stderr |
| checker code +1 | stdout comparison fails only, comparison exit 1 |
| checker start +1 | stdout comparison fails only, comparison exit 1 |
| checker length +1 | stdout comparison fails only, comparison exit 1 |
| checker nested message text gains mutant | recursive chain bytes fail only, comparison exit 1 |
| emitter one text byte changes | stdout comparison fails only, comparison exit 1 |
| emitter one output removed | output-population bytes fail only, comparison exit 1 |
| emitter BOM flag flips | stdout comparison fails only, comparison exit 1 |
| emitter emitSkipped flips | stdout comparison fails only, comparison exit 1 |
| checker and emitter library identities change, separately | compatibility check invalidates all observations; each recomputes 2/2 fixture projects |
| checker and emitter cached shards gain a byte, separately | reuse hash check rejects each, driver exit 1 |
| checker and emitter aggregate golden bytes change, separately | comparator integrity check rejects each before native command, exit 1 |
| checker and emitter selected request bytes change, separately | request integrity check rejects each before native command, exit 1 |
| comparator stdout one byte changes | only stdout differs, exit 1 |
| comparator stderr gains one byte | only stderr differs, exit 1 |
| comparator process exit changes from 0 to 1 | only exit differs, exit 1 |
| checker input string changes to number | exactly one project selected; cached output equals fresh libraries and no assignment TS2322 remains |
| emitter initializer expression changes 41+1 to 40+2 | exactly one project selected; emitted bytes equal fresh-library run |

The eight record-field producers all themselves exit zero with empty stderr.
Malformed-output, timeout, crash or assertion failure is not counted as a kill
of the two semantic source mutants. Fixture counts are refreshed in counts.md:
checker-use.a 31 AST nodes/10 symbols/two checker diagnostics;
checker-values.a 24/10/zero; emitter-text.a 22/5/zero. No native allocations,
leak counts, native mutant kills or internal/oracle registrations are claimed.

Evidence includes both complete compressed Node goldens, project manifests,
benchmark timings, small fixture goldens, source/comparator mutant logs, setup
log and source-versus-stock result files. Large scratch caches remain in /tmp
and are regenerated by run-components.sh. COMPONENTS.md defines how the native
side consumes the same host capabilities, pinned libraries and wire format.

The original language/ownership questions for @system_adamic remain undecided
in REPORT.md. This unit supplies the observations needed for those decisions.
