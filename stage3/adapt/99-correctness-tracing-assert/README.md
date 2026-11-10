# stopTracing boolean assertion

Task baobab. Base: origin/main e24f2600a5d8a8521c5128827cad395d47c81c81.
Pinned TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

Owner audit: adaptations 32, 47, 62, 63 and 65 edit tracing.ts. Adaptation
32 edits indexed reads and the optional trace record field; 47 edits the
startTracing loader catch; 62 owns dumpTypes; 63 owns dumpLegend; 65 owns
Node builtin declarations. No active pin covers stopTracing whole or its
Debug.assert(tracing, "Tracing is not in progress") expression. Historical
whole-file hashes are evidence, not active input pins.

The intended edit changes only the first assert argument to !!tracing.
Verification is pending. Raw evidence belongs on typescript/tracing-assert-evidence.
