# Scanner source proof

This driver imports the adapted scanner.ts and creates its scanner exactly as
parser.ts:1444 does: `createScanner(ScriptTarget.Latest, true)`. It calls `scan()`
until EOF for each input. Parser-directed rescans of slash, template or JSX
syntax are outside this lexical scanner proof.

Each line has six tab-separated fields: SyntaxKind name, full start, start, end,
token flags, and JSON-escaped token text. Positions are UTF-16 code units. EOF
is included, trivia is skipped, and token text is escaped so a multiline token
still occupies one output line. Reverse enum aliases follow stock TypeScript
6.0.3. Files are sorted by relative path, recursively, with every regular file
under src/compiler included, even JSON and generated files. files.json records
that order; each EOF separates files without adding non-token output.

After integrating stage3-base and stage3-type-imports at a3ef0dc:

```sh
source /workspace/adamic-tools/env.sh
stage3/drivers/scanner/run.sh /tmp/scanner-proof > /tmp/scanner-proof.log 2>&1
```

The output directory must be new. The runner builds a scratch copy of the apply
pipeline with adaptations 00, 10, 50 and 51. Adaptation 20 is deliberately excluded
per the October 7 instruction. apply.sh still constructs the tree and measures
the patch set. The runner regenerates diagnostics after adaptation 10 changes
the generator's type import. `--tree <already-applied-tree>` reuses an existing
scratch tree; `--compiler <binary>` selects an integrated feature compiler.
`--node-only` omits native compilation; `--inputs <tree>` selects a fixed corpus
for comparing the scanner before and after a source adaptation.

Node runs this same main.a through stock TypeScript's transpileModule, with
verbatim module syntax. This handles upstream's enums and namespaces without
using Adamic output. The loader resolves upstream .js requests to their .ts
sources and initializes the compiler barrel before importing the driver.
Starting the ESM graph directly at scanner.ts exposes parser.ts's load-time read
of textToKeywordObj before scanner initialization. Barrel initialization is an
oracle bootstrap, not evidence that native module initialization is implemented.

The runner retains separate stdout, stderr, reports, source closure edges and
input manifests. A failed native build exits 1 and leaves its full diagnostics;
it never reports a native comparison pass. When compilation succeeds it runs
the native binary with the same input paths and diffs every token byte.

Before the native build, the same `diff -u` comparison passes a copied Node
control, then rejects a scratch Node output with only the first token's end
incremented by one. This proves comparison sensitivity even while native is
blocked; the copied control is explicitly not native output.

BLOCKERS.md states the observed gate and the stopping point. Evidence retains
all checker diagnostics in returned order. It does not claim an exhaustive
Refused/NotYet traversal while checking still fails.
