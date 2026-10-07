# Gather TypeScript declarations

The claim is **tsc's own code, unchanged, gathered by a tool**. This tool uses stock TypeScript 6.0.3's checker, starts from exported entry symbols, and follows value references to their declarations at top-level declaration granularity. Functions (including overloads), variable statements, enums and classes are kept whole. Namespace members are gathered separately: reached members retain their exact bytes, surrounded by the namespace's original header and closing brace. Required interfaces and type aliases come along verbatim, with type-only imports where appropriate. Declaration text and attached comments are copied directly from the original source, without printing or rewriting an AST. The only rewrites are import lines, plus the namespace wrapper keeps only the members that are reached. Type imports point at declaring modules. Value imports retain their original binding modules, including the _namespaces barrels, so rewriting bindings does not add evaluation edges into a partially initialized cycle. Original export facades are copied verbatim. The output mirrors source paths and file names.

Run:

```sh
SLICE_TYPESCRIPT=/path/to/typescript/lib/typescript.js \
  bash stage3/slice/run.sh /path/to/adapted /path/to/new-slice \
  src/compiler/scanner.ts:createScanner \
  src/compiler/types.ts:ScriptTarget src/compiler/types.ts:SyntaxKind
```

Audit copied spans with `node stage3/slice/verify.cjs /path/to/new-slice`. The audit reads the original tree recorded in the manifest, compares actual UTF-8 bytes and checks every hash.

The destination must not exist. `slice.json` records each copied span's original and output UTF-16 offsets, byte count and SHA-256. Counts distinguish code declarations from export facade records. Original ordered runtime imports and re-exports are retained as side-effect imports; modules with no reached code still carry that import graph. Type-only edges do not execute. slice-entry.a offers the original compiler-barrel bootstrap; the scanner and parser drivers retain their own original entry roots. Line counts sum copied spans, including attached whitespace. Multiple entry symbols are supported for the parser worker.

The scanner member slice reaches eight files and 89 code declarations, plus two namespace-wrapper spans: 7,127 code-span lines. Its preserved evaluation graph has 78 modules, including 70 with no reached code. With 77 original export-facade records, the manifest has 168 spans and 7,284 copied-span lines. Its Node output is byte-identical to the full tree's 509,014 tokens (SHA-256 `c1a9f239790e158cc4471aa6c9273ff678cb32e5890b3d4c95e077ee90c61b0f`). The former Debug initialization cycle is absent. Reached Debug members are `isDebugging`, `fail`, `assert`, and `assertEqual`. `fail` and the assertions build failure messages; no enum formatting functions, dynamic `formatEnum`, or debug-info registration members are reached. Static namespace access follows the selected member; bare or dynamic namespace values conservatively gather exported members.

A smaller `core.ts:compareValues` entry ran on Node with output `-1, 1, 0, -1` (one per line). Removing its reached `compareComparableValues` implementation caused `ReferenceError: compareComparableValues is not defined`. Byte audits compared every recorded declaration span against its source. Namespace-member gathering is explicitly authorized.

The tool does not suppress checker diagnostics, alter function bodies, or invent exports. Inputs are the already adapted source, specifically adaptations 10 and 50 for scanner; adaptation 20 is excluded. Any new temporary semantic adaptation must be separately listed and pushed before implementation, and apply to slice files only.

The scanner proof used `bash stage3/drivers/scanner/run.sh OUTPUT --tree SLICE --inputs FIXED_FULL_CORPUS --node-only`, followed by `diff -u FULL_TREE_NODE_STDOUT OUTPUT/node.stdout`. The fixed corpus has 81 regular compiler input files. The slice itself has eight compiler modules; it scans the full corpus. The end-offset mutant is caught by the runner's diff. Deleting unicodeESNextIdentifierStart from a scratch slice caused Node ReferenceError during scanIdentifier; verify.cjs also rejects that omission.

The raw gathered slice passes the byte audit. Temporary scanner adaptations are separate edits after gathering and intentionally invalidate original span hashes; do not describe their edited statements as verbatim. Their README plans and baseline status are in stage3/adapt and scanner/BLOCKERS.md. Earlier checked-read and public-interface repairs failed the full baseline and were superseded by narrower repairs.

Use `--why src/compiler/checker.ts:createTypeChecker` (repeatable) after the
entries to print one shortest declaration-reference chain, with each original
reference location, expression, type/value classification and namespace
expansion. Chains use a breadth-first walk from all supplied entries; namespace
expansion is one reference edge. Results are JSON on stderr and in slice.json.
A declaration absent from the closure reports reached: false. Select names with
file:declaration; namespace members may use Debug.formatSyntaxKind. Overloads
with the same name select a shortest chain to one retained declaration.

Static namespace member access is followed through parentheses, as/type
assertions, non-null assertions and satisfies wrappers. Those wrappers preserve
the runtime namespace identity. Resolve the member on the original namespace,
even if an any/Record cast masks the checker's property symbol. Bare and dynamic
namespace uses still retain every export, even after an earlier static use.
Unread export facades do not count as references to every exported value.
No runtime condition is assumed false and no reachable declaration is dropped.

`--conservative-namespaces` reproduces the previous over-expansion for diagnosis;
`--reference-only` omits evaluation scaffolding and is not a runtime proof mode.
The default preserves the original ordered runtime graph and bindings.
verify.cjs checks both copied bytes and the ordered import prefixes.

Both final Node proofs are green on adaptations 10 and 50 (20 excluded), using
the original 81-file corpus. Parser output is 35,456,964 bytes, SHA256
2014ef06f9db928b50d001787c44e490bbdbd8ecd9e912d3f676d58148ffefc5.
Its driver entries reach 1,988 code declarations in 26 files, 41,677 code-span
lines; with facades, 2,077 copied spans and 41,834 lines. The 78-module graph
has 52 modules without reached code. Reversing only the barrel's generated
evaluation imports, leaving declaration bytes intact, fails on Node for both
slices and is independently caught by the import-order audit. See WHY.md and
evidence/evaluation-proof.json for exact counts, chains, commands and mutants.

Slice-only temporary adaptations live in `stage3/slice/adapt/`, never in the
full-tree adaptation directory. After gathering, the tool applies the scanner
profile in number order: 52–57, 59, 80–82, 85. Retired adaptations 58, 83, 84 and
plan-only 86 are excluded. Parser adaptations 60–64 remain full-tree adaptations
and are already present in the slice input. The parser profile has no additional
slice edits. `--no-adapt` emits only the verbatim gathered tree.

The gathered code is tsc's own code, unchanged, gathered by a tool: only import
lines and namespace wrappers selecting reached members are rewritten. Temporary
adaptations are a separate, explicit step after gathering. The pre-adaptation
files are retained under `.verbatim/`; `verify.cjs` checks their original source
spans and the final adapted file hashes. `slice.json` records the applied profile.
