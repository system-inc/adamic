# Gather TypeScript declarations

The claim is **tsc's own code, unchanged, gathered by a tool**. This tool uses stock TypeScript 6.0.3's checker, starts from exported entry symbols, and follows value references to their declarations at top-level declaration granularity. Functions (including overloads), variable statements, enums and classes are kept whole. Namespace members are gathered separately: reached members retain their exact bytes, surrounded by the namespace's original header and closing brace. Required interfaces and type aliases come along verbatim, with type-only imports where appropriate. Declaration text and attached comments are copied directly from the original source, without printing or rewriting an AST. The only rewrites are import lines, to point at declaring modules, plus the namespace wrapper keeps only the members that are reached. The output mirrors source paths and file names.

Run:

```sh
SLICE_TYPESCRIPT=/path/to/typescript/lib/typescript.js \
  bash stage3/slice/run.sh /path/to/adapted /path/to/new-slice \
  src/compiler/scanner.ts:createScanner \
  src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind
```

Audit copied spans with `node stage3/slice/verify.cjs /path/to/new-slice`. The audit reads the original tree recorded in the manifest, compares actual UTF-8 bytes and checks every hash.

The destination must not exist. `slice.json` records each copied span's original and output UTF-16 offsets, byte count and SHA-256. Counts include export facade records, which are retained when a module namespace is used as a value. Line counts sum copied spans, including attached whitespace. Multiple entry symbols are supported for the parser worker.

The scanner member slice reaches eight files and 89 declarations, plus two namespace-wrapper spans: 7,127 copied-span lines. Its Node output is byte-identical to the full tree's 509,014 tokens (SHA-256 `c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f`). The former Debug initialization cycle is absent. Reached Debug members are `isDebugging`, `fail`, `assert`, and `assertEqual`. `fail` and the assertions build failure messages; no enum formatting functions, dynamic `formatEnum`, or debug-info registration members are reached. Static namespace access follows the selected member; bare or dynamic namespace values conservatively gather exported members.

A smaller `core.ts:compareValues` entry ran on Node with output `-1, 1, 0, -1` (one per line). Removing its reached `compareComparableValues` implementation caused `ReferenceError: compareComparableValues is not defined`. Byte audits compared every recorded declaration span against its source. Namespace-member gathering is explicitly authorized.

The tool does not suppress checker diagnostics, alter function bodies, or invent exports. Inputs are the already adapted source, specifically adaptations 10 and 50 for scanner; adaptation 20 is excluded. Any new temporary semantic adaptation must be separately listed and pushed before implementation, and apply to slice files only.

The scanner proof used `bash stage3/drivers/scanner/run.sh OUTPUT --tree SLICE --inputs FIXED_FULL_CORPUS --node-only`, followed by `diff -u FULL_TREE_NODE_STDOUT OUTPUT/node.stdout`. The fixed corpus has 81 regular compiler input files. The slice itself has eight compiler modules; it scans the full corpus. The end-offset mutant is caught by the runner's diff. Deleting unicodeESNextIdentifierStart from a scratch slice caused Node ReferenceError during scanIdentifier; verify.cjs also rejects that omission.

The raw gathered slice passes the byte audit. Adaptations 52-54 are separate edits after gathering and intentionally invalidate original span hashes; do not describe their edited statements as verbatim. Their README plans and baseline status are in stage3/adapt and scanner/BLOCKERS.md. Earlier checked-read and public-interface repairs failed the full baseline and were superseded by narrower repairs.
