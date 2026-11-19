# internal host contracts: SolutionBuilderHostBase, CompilerHost, IncrementalCompilationOptions

Type-only, refusal family 5 (#pncjpqy, Compiler's split to Apple). Accepted by Compiler under the delegation. 6 sites stop today on "optional property P in T absent from structural source S, which can hide fields" (rulings.md:71). Each edit states a truthful, owned contract, certified as adaptation 72 is: the compiled program judges every relation and write at its own site, so a view the contract doesn't cover is still refused there, not trusted.

The evidence for each candidate, its construction trace and its measured stock result, is Correctness's scout, typescript/family5-scout 464502ff, `stage3/typescript/family5-scout/REPORT.md`.

## C09b (G09, G28, 4 sites)

`createSolutionBuilderHostBase` (tsbuildPublic.ts:294) obtains `createProgramHost` and adds known reporting functions. The default objects omit `getGlobalTypingsCacheLocation`, and every installed getter, reporter and parse-config callback is callable with the stated signature (program.ts:1966; tsbuildPublic.ts:596, :598). The members go on SolutionBuilderHostBase, since the scout measured the broader ProgramHost placement (C09) at TS2320.

- `src/compiler/tsbuildPublic.ts`, interface `SolutionBuilderHostBase`: adds `/** @internal */ getGlobalTypingsCacheLocation?: () => string | undefined;`.
- `src/compiler/tsbuildPublic.ts`, interface `SolutionBuilderHostBase`: adds `/** @internal */ onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void;`.

## C10 (G27, 1 site)

CompilerHost is made by `createCompilerHostWorker` in program.ts, and the watch and config helpers install a DiagnosticReporter-compatible callback or omit it. program.ts:5131 reads it with a `returnUndefined` fallback. CompilerHost has no nominal-class constructor, so the exemption is no premise here: an `@internal` optional callable states the contract.

- `src/compiler/types.ts`, interface `CompilerHost`: adds `/** @internal */ onUnRecoverableConfigFileDiagnostic?: (diagnostic: Diagnostic) => void;`.

## C11 (G21, 1 site)

`performIncrementalCompilation` (watch.ts:987) forwards its input to `createIncrementalProgram`. The CLI and controller constructors omit `createProgram`, and every supplied override is the same `CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram>` factory watchPublic.ts:132-151 uses. watch.ts already imports `CreateProgram`, so the member names it directly where the scout wrote an `import()` type; the type is the same.

- `src/compiler/watch.ts`, interface `IncrementalCompilationOptions`: adds `createProgram?: CreateProgram<EmitAndSemanticDiagnosticsBuilderProgram>;`.

## How it is held

- `adapt.cjs` finds every owner through the parsed source (an interface by name, a function with a body by name), never by a coordinate, and refuses before any write a missing, duplicate or drifted owner, or a set half made. A second run writes nothing. CRLF is kept.
- `contract.cjs TREE` checks the edits under LF and CRLF, the zero-edit replay with mtime kept, and the refusals: each owner renamed, each replaced text drifted, and the set half made.
- `check.cjs BEFORE AFTER` runs on two complete applies, one without this directory. It checks three things: only the named files differ under src, and undoing these edits gives the before files; stock 6.0.3 reports the same diagnostics over the 79 src/compiler roots; and each changed file transpiles to byte-identical JavaScript at ES2020 and ESNext.
- No published API changes: every member is `@internal` or sits on an internal interface, and every annotation is on a private function.
