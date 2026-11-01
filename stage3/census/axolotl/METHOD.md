# Method and scope

This is a measurement. No fixes were applied to compiler source, adaptations, checker options, or the prelude.

Measured main tip: `a23ba096388ceb3774b94d2db70009316d90b0cf`.
Baseline main: `141ad1d244d6a2d9c5403858086d505dfdef5fd5`.
Upstream source: TypeScript `v6.0.3`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`.
The measurement branch starts at the fetched main tip. Main was not pushed.

## Where the counts come from

`stage3/census/tool/main.go` walks every file under the adapted `src/compiler`, loads each file separately with `load.Load([]string{path})`, then loads the sorted `.ts` and `.a` paths together with `load.Load(paths)`. The last record is the compiler-directory whole-program observation. Its `CheckError.Diagnostics` slice length becomes `whole_program_diagnostics` in `stage3/meter/report.py:report`. This count is separate from the fraction of single-root loads that pass the checker.

For own-file counts, `report` visits diagnostics from every record, including the aggregate record. It parses only the primary `file:line:column: error TS...` prefix, resolves that file, and adds the entire formatted diagnostic string to a set for that file when the file is under the compiler directory. Elaborations do not own diagnostics. The own-file sum is the sum of those set sizes. It is neither the number of messages in each root's load nor merely the aggregate diagnostic count. Global and external diagnostics are tracked separately.

The scratch Go extractor calls the production `load.Load` for the same sorted source-root singletons and aggregate roots, retaining `CheckError.Diagnostics`. It is built using a Go file overlay replacing only the census command's main file in the build view. Its source is `/workspace/scratch/extract.go`, outside the checkout, and is not committed. The independent ordinary census covers the three non-source inputs as well. Their extension-gate errors do not contribute diagnostics. The extractor never calls lowering. The ordinary census calls lowering only after a successful load, following the meter.

`stage3/census/latent/tool/main.go` is a stub without an overlay. `stage3/census/latent/make_overlay.py` substitutes `main.go.txt`, exposes `LatentLoad`, and retains rejected programs and raw diagnostic spans. The overlay calls the same loader and diagnostic collection, and does not change compiler options. Its aggregate header repeats the diagnostics; latent body ownership (`LatentOwnDiagnosticsIn`) is a separate rule and is not the meter's own-file rule. Full latent lowering is unnecessary to measure these checker counts and was not run here.

## Effective loader configuration

`internal/load/load.go:compilerOptions` sets:

```json
{
  "strict": true,
  "noUncheckedIndexedAccess": true,
  "exactOptionalPropertyTypes": true,
  "erasableSyntaxOnly": false,
  "verbatimModuleSyntax": true,
  "allowImportingTsExtensions": true,
  "noEmit": true,
  "module": "ESNext",
  "moduleDetection": "force",
  "moduleResolution": "bundler",
  "target": "ES2024",
  "lib": ["lib.es2024.d.ts"],
  "types": []
}
```

`noImplicitReturns` and `noFallthroughCasesInSwitch` are not enabled. No project tsconfig is read. The host is the real source filesystem, wrapped with the embedded ES libraries, Adamic's prelude, Node library handling, regexp declarations, and filesystem caching. Every load adds the prelude root. The program is single-threaded. When a resolved graph uses Node builtins, `usesNodeModules` triggers a rebuild with the embedded `@types/node` 25.3.3 index, a Node-compatible prelude, and `typeRoots: ["bundled:///node/node_modules/@types"]`. This graph-dependent rebuild matters when comparing singleton loads with the aggregate.

`Program.diagnostics` returns syntactic diagnostics alone if any exist; otherwise it combines config parsing, program, global, bind, and semantic diagnostics. It formats each with `Program.formatDiagnostic`, including the complete indented message chain, then sorts the strings. Positions are one-based; columns count UTF-16 code units. The report preserves full diagnostic messages, including chains. Source lines refer to the adapted tree, not the pristine upstream or Go files.

## Classification standard

Each unique own-file diagnostic receives exactly one cause. Whole-program membership is recorded independently so the overlapping populations are not added together.

The existing rulings in `docs/rulings.md` under compiler strictness explicitly keep tsc under the strict options and require truthful source adaptations. Optional properties that are explicitly assigned undefined must admit that value or omit the property. Indexed reads must carry a proof or handle absence. Caught JavaScript values are unknown and need a guard or truthful normalization. Those established rules do not require a new language ruling. A proposed adaptation is a repair direction, not proof that a particular edit is sound or sufficient; downstream work must audit callers and preserve behavior.

A missing third-party module is not automatically a wrong global prelude. Adamic's bundled Node declarations cover Node builtins, not every optional package. A trusted host API must be pinned and audited before its declaration can be treated as an axiom, as required by `docs/0.1.md`'s ambient declaration rule.
