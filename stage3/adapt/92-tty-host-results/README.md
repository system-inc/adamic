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

On the fully adapted main tree, executeCommandLine.ts:265 and :534 are the only
getWidthOfTerminal callers. Both already use sys.getWidthOfTerminal?.() ?? 0.
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
main/with-92 oracles are recorded under proofs/. The initial push precedes the
long runs; their actual counts and outcome will be added without changing the
adapter or widening its scope. No expected count, API sanction or baseline is
updated to make a failing run green.

The required mutant undoes both bodies with head A's exact served tty.d.ts.
TestTTYHostAxiomTypecheck must report TS2322 at sys.ts:1531:13 and :1534:13;
restoring 92 removes those two errors. The checker must also admit the adapted
bodies against the published declarations. See proofs/typecheck.mutant.md.
