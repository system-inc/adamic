# Resolved caller classification

Counts are calls, not identifier matches. Full calls, arguments and declarations are in the caller JSON.

## getDiagnostics

- `src/compiler/checker.ts:1925`: argument; result `DiagnosticWithLocation[]`.
- `src/compiler/checker.ts:49737`: bound; result `DiagnosticWithLocation[]`.
- `src/compiler/checker.ts:49761`: returned; result `Diagnostic[]`.
- `src/compiler/emitter.ts:789`: PropertyAssignment; result `Diagnostic[]`.
- `src/compiler/program.ts:2821`: bound; result `DiagnosticWithLocation[]`.
- `src/compiler/program.ts:3278`: bound; result `DiagnosticWithLocation[]`.
- `src/compiler/program.ts:3280`: argument; result `DiagnosticWithLocation[]`.
- `src/compiler/programDiagnostics.ts:157`: member read; result `Diagnostic[] | undefined`.
- `src/testRunner/unittests/diagnosticCollection.ts:23`: bound; result `Diagnostic[]`.
- `src/testRunner/unittests/diagnosticCollection.ts:55`: bound; result `Diagnostic[]`.
- `src/testRunner/unittests/diagnosticCollection.ts:76`: bound; result `Diagnostic[]`.

## getSupportedExtensionsWithJsonIfResolveJsonModule

- `src/compiler/commandLineParser.ts:3980`: bound; result `readonly string[][]`.
- `src/compiler/program.ts:1588`: bound; result `readonly Extension[][]`.
- `src/compiler/utilities.ts:10151`: argument; result `readonly string[][]`.
- `src/services/stringCompletions.ts:734`: ConditionalExpression; result `readonly string[][]`.

## convertOrRepopulateDiagnosticMessageChainArray

- `src/compiler/builder.ts:544`: PropertyAssignment; result `DiagnosticMessageChain[] | undefined`.
- `src/compiler/builder.ts:550`: PropertyAssignment; result `DiagnosticMessageChain[] | undefined`.
- `src/compiler/builder.ts:553`: bound; result `DiagnosticMessageChain[] | undefined`.

## convertOptionsFromJson

- `src/compiler/commandLineParser.ts:3776`: discarded; result `CompilerOptions | TypeAcquisition | undefined`.
- `src/compiler/commandLineParser.ts:3789`: discarded; result `CompilerOptions | TypeAcquisition | undefined`.
- `src/compiler/commandLineParser.ts:3794`: returned; result `WatchOptions | undefined`.

Diagnostic collection: 5 filename calls (all filename property reads); 6 no-argument calls. Filename results feed diagnostic return/concatenation/directive filtering; no-argument results feed one return, emitter output, recomputation iteration, and three equality assertions.

Extensions: 1 enum-array input at program.ts feeds flatten and fileExtensionIs; 3 string-array inputs feed flatten/extension-priority logic or completion extension results. None writes an inner array at these sites.

Builder: all 3 calls pass chain.next as T[], with the same (T) repopulation callback; 2 install the returned array as next on newly constructed diagnostic chains, and 1 compares returned identity with chain.next before reusing the original chain.

Options: 2 calls mutate and retain a supplied default object while discarding the result; 1 passes undefined and returns the resulting watch options.
