# Gather TypeScript declarations

The claim is **tsc's own code, unchanged, gathered by a tool**. This tool uses stock TypeScript 6.0.3's checker, starts from exported entry symbols, and follows value references to their declarations at top-level declaration granularity. Functions (including overloads), variable statements, enums, namespaces and classes are kept whole. Required interfaces and type aliases come along verbatim, with type-only imports where appropriate. Declaration text and attached comments are copied directly from the original source, without printing or rewriting an AST. Only import lines are rewritten, to point at declaring modules. The output mirrors source paths and file names.

Run:

```sh
SLICE_TYPESCRIPT=/path/to/typescript/lib/typescript.js \
  bash stage3/slice/run.sh /path/to/adapted /path/to/new-slice \
  src/compiler/scanner.ts:createScanner \
  src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind
```

The destination must not exist. `slice.json` records each copied span's original and output UTF-16 offsets, byte count and SHA-256. Counts include export facade records, which are retained when a module namespace is used as a value. Line counts sum copied spans, including attached whitespace. Multiple entry symbols are supported for the parser worker.

This is a declaration gatherer, not an initialization-cycle repair. A namespace value or dynamic module lookup conservatively reaches all its members. In particular, the whole `Debug` namespace contains `(ts as any)[enumName]`, reaching the compiler barrel and its exports. Under the requested whole top-level namespace rule the scanner still reaches 78 files. Redirecting its imports also exposes a runtime initialization cycle: `binder.ts` reads `Debug.attachFlowNodeDebugInfo` while `Debug` is undefined. The strict scanner slice therefore has not passed the Node oracle; do not treat it as a native scanner proof.

A smaller `core.ts:compareValues` entry ran on Node with output `-1, 1, 0, -1` (one per line). Removing its reached `compareComparableValues` implementation caused `ReferenceError: compareComparableValues is not defined`. Byte audits compared every recorded declaration span against its source. Namespace-member slicing would change the requested granularity and is pending clarification.

The tool does not suppress checker diagnostics, alter function bodies, or invent exports. Inputs are the already adapted source, specifically adaptations 10 and 50 for scanner; adaptation 20 is excluded. Any new temporary semantic adaptation must be separately listed and pushed before implementation, and apply to slice files only.
