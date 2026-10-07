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
lexicographic relative-path order. Every input, including JSON, is parsed as
ScriptKind.TS, ScriptTarget.Latest, with parent links disabled. The source-file
name is the relative manifest name. Each file header is followed by preorder
nodes and that file's parse diagnostics in parser order. Positions and ends
are UTF-16 offsets. Nodes print canonical kind name, pos, end, integer flags,
and cooked identifier/private-identifier/literal text. Backslash and non-ASCII
code units are escaped as \\uXXXX, so one node or diagnostic occupies one line.
No parent, transform, list or JSDoc-tag fields are included. forEachChild is the
ordinary compiler traversal, not a separate traversal of attached JSDoc tags.

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

The declaration-slice proof is pending the shared scanner tool. See
BLOCKERS.md for its required sequence and WITNESS.md for the completed
stage1 projection comparison.
