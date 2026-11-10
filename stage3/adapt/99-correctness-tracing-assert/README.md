# stopTracing boolean assertion

Task baobab. Base: origin/main e24f2600a5d8a8521c5128827cad395d47c81c81.
Pinned TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

This JavaScript-changing adaptation changes only the first stopTracing assertion
argument from tracing to !!tracing. The message and every other source byte stay
unchanged. tracing can hold only tracingEnabled or undefined, and Debug.assert
reads its argument only as if (!expression), so the assertion outcome is preserved.
The asserts-expression narrowing is lost. Nothing after the call reads tracing;
its next occurrence assigns undefined. Adamic types accepts all 79 compiler roots
with zero errors before and after the edit.

Adaptations 32, 47, 62, 63 and 65 edit tracing.ts. Adaptation 32 owns indexed reads
and the optional TraceRecord field; 47 owns the startTracing loader catch; 62 owns
dumpTypes; 63 owns dumpLegend; 65 owns Node builtin declarations. No active pin
covers stopTracing whole or its exact assert expression, so this is a new owner.
Historical whole-file hashes are evidence rather than active input pins.

adapt.cjs pins the complete composed stopTracing implementation before and after,
including CRLF line endings and adaptation 32's existing non-null assertion.
It refuses drift, mixed edits or an unreviewed owner, and a second run writes
nothing. Stock emission, with comments on and off, admits only the argument edit.
Declaration emission is identical with stripInternal on and off.

witness.cjs runs against two normal built trees. It starts server tracing in a
temporary directory and stops it, then separately stops without starting. Outcomes
and assertion messages must be byte-identical. Run it with Node and the before
and after tree paths as its two arguments.

PROOF.md records the complete lane count, positive controls, witness, restoring
mutant, next wall, every JavaScript hunk and runtime export comparison. Raw logs,
counters, census streams and measurement scripts live only on the separate
branch typescript/tracing-assert-evidence.
