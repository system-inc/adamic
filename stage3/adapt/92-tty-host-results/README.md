# 92: truthful Node TTY results

Task #jk7tph7, head B1. This is a JavaScript-changing adaptation: only the two
Node system method bodies in src/compiler/sys.ts change. System keeps
writeOutputIsTTY?(): boolean and getWidthOfTerminal?(): number. No public API
change is intended; the oracle must retain exactly main's sanctioned API diff.

The independent base is main f2cf1ebb169994b3d27b9d5fc9760c5b2d7d8c6a.
The upstream is TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
Head A is 92d1abe8f28073fd9bd242138017f4111aa02b73; this branch does not carry
its loader changes or stdout-write. It follows adaptation 91's adapter, target
hashes, scope/idempotence proofs, proof logs and serial control/after oracle.

## Readers and Node equivalence

On the fully adapted main tree, generateOptionOutput at executeCommandLine.ts:265
and getHeader at :534 are the only getWidthOfTerminal callers. Both already use sys.getWidthOfTerminal?.() ?? 0.
Moving that fallback into the method gives exactly the same value at each read.
At :268, zero selects the compact option layout instead of the >=80-column
layout; at :540 it omits the wide TS icon/header alignment. No width is invented.

The sole writeOutputIsTTY caller is defaultIsPretty, executeCommandLine.ts:166.
Its consumers read the result only by truthiness: shouldBePretty :171,
createColors :189-190, updateReportDiagnostic :154, createReportErrorSummary
:903, createBuilderStatusReporter in tsbuildPublic.ts:289, and
createWatchStatusReporter in watch.ts:194. The builder/watch call chain is
executeCommandLine.ts:859, :889 and :1008. Replacing undefined with false keeps
all these branches and selected output functions the same. NO_COLOR/FORCE_COLOR
precedence is unchanged. These positions were checked on the adapted tree.

The adapter pins the complete before/after sys.ts hashes and unchanged System
signatures. It validates both targets before writing, and a rerun writes zero
files. It changes only the two return expressions. Emitted JavaScript is
intentionally different in those two bodies; baseline and oracle equality,
not JavaScript byte identity, is the acceptance criterion.

## Proofs

Bounded checker, consumer-equivalence, scope and mutation proofs and full serial
main/with-92 oracles are recorded under proofs/. The adapter and bounded proofs
were pushed before accepting the long comparison; final oracle evidence follows
by fast-forward. No expected count, API sanction or baseline was updated.

The required mutant undoes both bodies with head A's exact served tty.d.ts.
TestTTYHostAxiomTypecheck must report TS2322 at sys.ts:1531:13 and :1534:13;
restoring 92 removes those two errors. The checker must also admit the adapted
bodies against the published declarations. See proofs/typecheck.mutant.md.

## Bounded preparation results

Both full numbered applies exit 0. The official apply tool regenerated
stage3/patch-set.md: row 92-tty-host-results is 1 file, 2 added lines, 2 removed.
The same generation refreshes older stale table rows and removes its old prose
suffix; none was hand-merged. The actual main control's total is 85 files,
5,523 added / 5,481 removed; with 92 it is 85 files, 5,525 / 5,483.

TestTTYSourceScope inventories all 746 source files and finds only sys.ts
changed. Removing the two expression edits restores its exact source bytes.
Comments-retained and comments-removed transpilation differ only at those two
return expressions. The System interface text is byte-identical. The adapter's
idempotence run reports zero files, and its moved-target guard rejects an added
newline before writing. TestTTYConsumerEquivalence passes 63 cases using the
actual method bodies, both actual width readers, defaultIsPretty and createColors.
A wrong numeric fallback (1) makes that witness fail.

With locked dependencies installed, the ordinary compiler check is:

| Tree | Published @types/node | Head A served tty.d.ts |
| --- | ---: | ---: |
| Main control | 0 | 2 TS2322 |
| With 92 | 0 | 0 |

With Adamic's stricter options, the existing 140 diagnostics stay unchanged:
main is 140 / 142, with-92 is 140 / 140. The two additions are exactly
sys.ts:1531:13 and :1534:13. The required undone mutant exits 1 with those two
errors; the restored check exits 0. The served Node file is exported through A's
real applicator, not a second correction list. hashes and commands are retained.

Both upstream builds exit 0 and their built/local/typescript.d.ts bytes match.
No expected count, API sanction, harness source or fixture status is changed.
Main's current lane requires eight workers, so the complete serial oracle runs
use eight and the unmodified lane checker (91's historical run used four).

## Complete serial oracle and public API

The current main lane pins eight workers. Both final full runs used
--workers 8 --limit-seconds 1800, all runners and no filter. Both upstream
builds finished before the first suite. No setup or other test workload ran
beside either final suite. The actual unmodified stage3/lane/check.py CLI says
PASS stage3 landing lane for both trees. The sole failure in each raw oracle is
the existing sanctioned public-API baseline test; neither count nor sanction
was changed.

| Oracle | Passing | Sanctioned API failure | Pending | Wall seconds |
| --- | ---: | ---: | ---: | ---: |
| Main control | 106,366 | 1 | 0 | 313.431 |
| With 92 | 106,366 | 1 | 0 | 321.55 |

The complete baseline.diff is byte-identical: 51,093 bytes,
SHA-256 `5333568a661aea24fc8af3149e9f4d2dfca70731842461c677f385b092adc4f0`.
The emitted public typescript.d.ts is also byte-identical: 590,861 bytes,
SHA-256 `477a56fb7b35a5f268be6131441c67d237b4646a9f8d2e1f50dc163879450051`. The unchanged normalizer accepts exactly
223 sanctioned declarations. Appending an extra baseline-diff line makes the
same lane's byte checker fail; the original bytes are restored and the full
lane CLI passes again. See proofs/oracle.mutant.md.

The final built compilers were also observed through Node v24.19.0 with stdout
on a pipe. Main's raw width and TTY results are undefined/undefined; with 92 they
are 0/false. Applying the actual readers gives 0/false in both. This witnesses
the deliberate JavaScript change and the reader equivalence on real artifacts.

After the suites, the 746-file scope check and all 63 reader cases pass again,
and the adapter's final rerun writes zero files. Both final suites leave every
cgroup memory-event counter unchanged.

## Final run preparation

Each accepted final run starts with only the stock harness's generated,
untracked .parallelperf.json cache reset, like a fresh tree. No harness source,
worker count, fixture, baseline, or expected outcome changes. The main host
reads that cache at src/testRunner/parallel/host.ts:204 and writes it at :600.
Final logs show the stock size-based packing and unchanged counters.

The official apply command regenerated the patch table, including older stale
rows; there is no hand merge. No internal/load, library_node_host.go, fixture
status, oracle implementation, lane checker, expected count or API sanction is
changed by B1. Compressed raw logs use the same layout as adaptation 91.
