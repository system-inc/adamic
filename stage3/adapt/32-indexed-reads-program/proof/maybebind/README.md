# Historical kpp0f8s stop at 997331ef1

The [continuation on main 12fec995](continuation/README.md) supersedes this
historical stopping report. Amended 99x at 5fa773b7 resolves the blocker.

Base origin/main is 322f80565edaa978d928ca559803c578ed67c105. The upstream pin
is 050880ce59e30b356b686bd3144efe24f875ebc8. This is a partial result, not a
completed seven-check delivery. checks.json records every requested check.

| Check | Result |
| --- | --- |
| 1. Full applies with --build-product | Both fresh final runs exit 0 |
| 2. Paired replay, continuing after refusals | All 81 outcomes identical, empty diff |
| 3. Adamic types | All 79 sorted roots together, exit 0, zero diagnostics |
| 4. Nine emitted JavaScript files | Seven byte-identical; _tsc.js and typescript.js each contain exactly twelve rewrites |
| 4. Public declarations | Both byte-identical |
| 5. Mutants | Half owner and foreign callee refuse without writes; one restored site is reported exactly |
| 6. Whole lane and 398 shards | Not run; preparation terminated at the composition blocker |
| 7. CW051 | Three configurations measured; 32 plus 99x is blocked before measurement |

The unchanged 99x adapter from a6d5f494e46262f805e5be76a4d7c108fe6a3a6c
applies successfully to the unchanged main tree. It refuses main plus this fold
with `handed-off maybeBind site inventory drift`. Its adapt.cjs compares the
remaining handed-off calls with inventory.left exactly: six checker/program
calls remain, while the twelve watch calls have correctly been rewritten by 32.
No 99x source was changed. Updating that handoff assertion requires work outside
the permitted directory. The user explicitly permits stopping with a reason.
No combined-tree latent result, oracle PASS, or shard coverage is claimed.
The exact refusal is in logs.tar.gz as 32-99x-composition.log.

For each of performBuild, createWatchOfConfigFile and
createWatchOfFilesAndCompilerOptions, main's first lowering stop is
core.ts:1533:12, an optional call inside maybeBind. With this fold it is
watch.ts:856:19, the optional binding at system.realpath. Main plus 99x keeps
the baseline stop. All nine measured units were attempted with no checker
diagnostics. These are first-error measurement results, not native-build claims.
latent-summary.json retains the exact reasons and attempting units; the raw
records, selections and timings are in latent-raw.tar.gz. The production
compiler is unchanged. The measurement-only selector changes which units are
attempted and uses the repository's existing first-error overlay and no-output
guards. Its Go source is latent-selector.go.txt.

replay-before.json and replay-after.json retain exact refusals, written files
and edits for every adapter, including baseline refusals. In particular 32
refuses `handoff producer runtime drift: appendSourceMap` on both complete
applied trees, before reaching watch-owner validation. The independent pre-32
contract run and the successful full apply exercise the new watch code.
replay-records.diff is empty. The passive filesystem journal and source/mtime
manifests are packed in replay.tar.gz, with raw member hashes in
replay-files.sha256 and an archive hash in replay.sha256. The replay copies
contain the complete applied source trees and Git history; subsequent npm
install/build outputs are excluded, matching the pre-build full-apply inputs.

The emitted check normalizes only the twelve ledgered expressions and their
introduced temporary declarations. No surviving binding was renamed in this
measurement. All remaining generated bindings have matching declarations and
symbol occurrence counts. The complete normalized ASTs agree in both changed
bundles; every other JavaScript product is byte-identical. Public declarations
are 593,502-byte typescript.d.ts and 866-byte tsserverlibrary.d.ts.
javascript-source-delta.json independently compares every source file and
requires precisely the twelve substitutions in watch.ts, including CRLF bytes.
The 27-file existing verifier also passes with comments kept and removed.

The half-owner mutant changes only watch.ts:761:32 and refuses before any source
write. The foreign-import mutant leaves the pinned owner text unchanged but
redirects the maybeBind import to noop; type-checker symbol resolution catches
it. A successful second run preserves source bytes and nanosecond mtimes.
The emitted mutant restores only watch.ts:761:32 to the original helper call.
The independent nine-file comparison exits 1 and reports only
createCompilerHostFromProgramHost / host.getDefaultLibLocation, in both bundles.
Its expected failing verdict is mutant-javascript-comparison.json.

Setup initially failed to fetch the nested TypeScript dependency. The exact pin
was already present locally, checked out successfully, and the setup retry
passed. The first baseline full apply was invalidated by an overlapping compiler
build that made tar report a changing output directory. The initial after apply
also exposed an incorrect assumption that all watch maybeBind imports became
unused; four unselected calls remain, and the import is now preserved. Both
attempts are retained in logs.tar.gz and superseded by the successful fresh
32-before-final and 32-after-final runs. Successful full applies finished before
their final compiler builds. Evidence and helper-script edits after those
applies do not change the five frozen transformation inputs hashed in
provenance.json.

Raw command logs are packed in logs.tar.gz and all archives have sha256 lists.
There are no raw per-shard files because no shard was run. No live task process
remains; stopped preparation processes have exited.
