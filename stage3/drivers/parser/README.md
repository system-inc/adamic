Current reference v2: **36,429,231 bytes**, SHA256
686a89adf8f215a92b3751b02b767fb062d6bc285d63bb4e363b60f16395d615.
Full adapted tree and slice outputs compare byte for byte, empty stderr.
reference.json is now the native acceptance target. The old 35,456,964-byte
SHA 2014ef06... reference is syntax-only historical evidence: dropping tags
left it unchanged. Blind-spot proof was pushed separately before this repair.

The full corpus emits 4,149 JSDoc nodes, 3,846 tag nodes, five type expressions
and 56 links. Its 81 jsDocDiagnostics lists are empty because the corpus uses
TS mode. jsdoc-inputs.json supplies directed JS cases, including a real malformed
@param type producing TS1110, and structured comment/link/type/tag text. The
same driver renders those cases; evidence/jsdoc-extended holds exact output.
Strict scratch checker accepts the extended driver without diagnostics.

Status pending native proof: **syntax tree and parse diagnostics identical on
Node, JSDoc unverified**. Extended JSDoc Node equality is an observed oracle
result; no native JSDoc result is claimed. When module-init-order lands, native
comparison uses the v2 reference, never the historical syntax-only SHA.

# Parser source proof driver

main.a calls createSourceFile and forEachChild from the adapted TypeScript
source tree. run.sh stages the driver at that tree's root, then runs exactly
that source on Node. node.mjs only transpiles module syntax, types, enums and
namespaces with stock TypeScript 6.0.3; it does not substitute the stock package's
parser for the adapted source parser.

```sh
source /workspace/adamic-tools/env.sh
bash /tmp/parser-apply10/stage3/apply.sh /tmp/new-parser-tree > /tmp/apply.log 2>&1
(cd /tmp/new-parser-tree && node scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json) > /tmp/generate.log 2>&1
bash stage3/drivers/parser/run.sh /tmp/new-parser-tree /tmp/new-parser-output > /tmp/driver.log 2>&1
```

The apply worktree must be pinned to a3ef0dc93d5b2a6cf58f74669c763dc83a1aad0e,
with adaptation 10 and no adaptation 20. Normal upstream regeneration is needed
after adaptation 10 changes the generator template. The generated JSON is an
additional corpus file, alongside the generated TypeScript diagnostic map.

The manifest includes every regular file recursively under src/compiler in
lexicographic relative-path order. Compiler corpus inputs, including JSON, are parsed as ScriptKind.TS; directed
.js inputs use ScriptKind.JS. ScriptTarget.Latest and disabled parent links
apply to both. The source-file
name is the relative manifest name. Each file header is followed by preorder
nodes and that file's parse diagnostics in parser order. Positions and ends
are UTF-16 offsets. Nodes print canonical kind name, pos, end, integer flags,
and cooked identifier/private-identifier/literal text. Backslash and non-ASCII
code units are escaped as \\uXXXX, so one node or diagnostic occupies one line.
The v2 traversal prints each node, then attached node.jsDoc entries, then
forEachChild children. Each JSDoc and tag has kind/pos/end/flags; tags include
tagName. Comments distinguish absence, string text and structured parts; link
parts print their text and name descendants. forEachChild on JSDoc traverses
tags, their type-expression nodes and all type descendants. Each file prints
parseDiagnostics and jsDocDiagnostics, including zero-length lists. Parents,
transform and other list metadata are not dumped.

kinds.a preserves TypeScript 6.0.3's canonical enum names, excluding First/Last
range aliases. The dump itself uses the adapted tree's node kinds and flags.

run.sh also stages a mutant that changes the first Identifier's actual end
by one. Both executions must finish successfully. cmp must return exactly 1,
and the report checks that exactly one line differs, by that end alone.
An optional third argument runs an existing native binary with the same tree
and manifest and compares both stdout and stderr byte for byte. Native build
support is measured separately in BLOCKERS.md; a Node-only run does not prove
native agreement.

The full dump is retained outside Git at
/tmp/parser-node10-final/node.dump, with the earlier equal run at
/tmp/parser-node10-regenerated/node.dump.
Only its byte size, SHA-256, corpus manifest and mutant evidence are committed.
All tests and runs write output to logs. No repository-wide gate is claimed.

The shared declaration slice and full tree match on Node. The native build
remains blocked; see BLOCKERS.md. WITNESS.md compares only the legacy ordinary
syntax projection and does not establish JSDoc coverage.
