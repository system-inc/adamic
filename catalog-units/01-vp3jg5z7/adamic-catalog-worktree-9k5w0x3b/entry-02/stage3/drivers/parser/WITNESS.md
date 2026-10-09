# Stage1 parser witness comparison

The stage1 parser cannot emit the full dump: it lacks node flags and parse
diagnostics. witness.a compares its available preorder projection: kind, pos,
end and identifier/private-identifier text. Both sides use UTF-16 offsets.
Only EndOfFile is renamed EndOfFileToken; other kind differences are preserved.
Literal text and the three JSON files are excluded.

On the adaptation-10 tree, all 78 TypeScript files completed with exit 0 and
empty stderr. Both parsers emitted 915,578 nodes. There are 639 changed node
records in 24 files. Every difference is a kind difference; positions, ends,
identifier text, preorder and node population match. The changes are:

| createSourceFile kind | Stage1 kind | Nodes |
| --- | --- | ---: |
| ExpressionWithTypeArguments | TypeReference | 635 |
| OmittedExpression | BindingElement | 4 |

[evidence/witness-every-difference.json.gz](evidence/witness-every-difference.json.gz)
contains every differing record, its file and both node ordinals. The
[report](evidence/witness-report.json) contains per-file counts and hashes.
This comparison preserves the differences rather than normalizing them away.

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs stage3/drivers/parser/witness.a /tmp/parser-adapted10 /tmp/parser-node10-final/manifest > /tmp/parser-witness.dump 2> /tmp/parser-witness.stderr
python3 stage3/drivers/parser/compare-witness.py /tmp/parser-node10-final/node.dump /tmp/parser-witness.dump /tmp/parser-witness-comparison > /tmp/parser-witness-comparison.log 2>&1
```

The comparison reports differences without failing its process, so it can
save the exhaustive inventory. A self-comparison produced zero hunks. A
planted first-Identifier end increment in that equal projection produced
exactly one hunk in one file; the assertion checked those counts. This is
separate from run.sh's actual Node-node mutation, which must fail byte equality.

This is a Node comparison of the stage1 TypeScript port, not native execution
of stage1 and not a proof of the pending createSourceFile slice.

## Why the two kinds differ

These are deliberate representations in pinned typescript-go, faithfully
followed by stage1. Neither example demonstrates a stage1 parser bug relative
to its Go oracle. For the stage3 proof, tsc 6.0.3 is the oracle: its kind names
are the required bytes. Replacing the driver with stage1 cannot produce the
same full dump, and an unmodified stage1 projection differs at these kinds.

### Heritage: ExpressionWithTypeArguments versus TypeReference

The shortest trimmed corpus line containing one of these 635 changed nodes
is `src/compiler/types.ts:3951` (37 characters):

```typescript
export interface JSDoc extends Node {
```

A reduced complete single-line reproducer is:

```typescript
interface I extends N{}
```

The changed node occupies [19, 21), including the space before N:
`ExpressionWithTypeArguments 19 21` on tsc, `TypeReference 19 21` on stage1
and Go. The newline following the source affects only SourceFile/EOF ends.

TypeScript 6.0.3's parser.ts:8207 parseHeritageClause always calls
parseExpressionWithTypeArguments, which finishes a createExpressionWithTypeArguments
node at :8223. This applies to interface extends and class implements as well
as class extends.

Pinned Go parser.go:1835 explicitly selects parseTypeHeritageClauseElement
when isTypeHeritageClause(:1847) sees interface extends or class implements.
That helper(:1852) first parses the expression form, verifies a valid entity-name
expression, converts the name, and finishes NewTypeReferenceNode at :1859.
It retains the expression form for invalid heritage expressions and class
extends. This conditional conversion establishes a deliberate representation
change; a blind global kind rename would be wrong.

Stage1 statements.ts:912 builds TypeReference for interface extends;
classDeclaration at :506 uses this.type() for implements and the expression
path for extends. parser.ts:559 finishes that ordinary TypeReference.
Thus stage1 follows Go's distinction on these valid inputs.

### Binding holes: OmittedExpression versus BindingElement

The shortest trimmed line among the four changed nodes is
`src/compiler/executeCommandLine.ts:412` (42 characters):

```typescript
.map(([, synonyms]) => synonyms.join("/"))
```

A reduced complete single-line reproducer is:

```typescript
let[,x]=a;
```

The hole occupies the zero-width range [4, 4): `OmittedExpression 4 4` on
tsc, `BindingElement 4 4` on stage1 and Go.

TypeScript parser.ts:7583 parseArrayBindingElement detects a comma and returns
createOmittedExpression before parsing an ordinary binding element.
Go parser.go:1659 explicitly leaves dotDotDotToken, name and initializer nil
for a comma (comment: "These are all nil for a missing element"), then always
finishes NewBindingElement at :1670. Stage1 parser.ts:813 likewise creates an
empty BindingElement for the comma. This is deliberate in Go, not accidental
loss of an identifier in stage1.

The array-expression control `[,x];` emits OmittedExpression on all three
implementations. The change applies to a binding-pattern hole, not all array
holes. Go's expression parser uses NewOmittedExpression at parser.go:5559;
stage1's corresponding expression path is parser.ts:1278.

## Reduced observations and check

Five ASCII sources were run through the adapted tsc source on Node, stage1
on Node, and the unmodified Go parser pinned above. All exited 0 with empty
stderr; tsc reported zero parse diagnostics on every source.

| Source | tsc changed node | Stage1 / Go node |
| --- | --- | --- |
| interface I extends N{} | ExpressionWithTypeArguments | TypeReference |
| let[,x]=a; | OmittedExpression | BindingElement |
| class C implements N{} | ExpressionWithTypeArguments | TypeReference |
| class C extends N{} | ExpressionWithTypeArguments | ExpressionWithTypeArguments |
| [,x]; | OmittedExpression | OmittedExpression |

Stage1's original rich whole-tree protocol was byte-identical to Go for all
five cases. The simpler stage3 projection had exactly three changed records,
only their kind names, and 37 nodes on each side. Source strings, hashes,
exit results and source pins are in evidence/difference-cases.json; all
three implementations' records are retained in evidence/difference-cases-*.dump.

To prove this new Go comparison can fail, a scratch copy of stage1 changed
only interface heritage's bases.push TypeReference to ExpressionWithTypeArguments.
The mutant completed on Node with exit 0 and empty stderr. It differed from
Go in exactly one node kind, [19, 21), while all other records matched;
evidence/difference-cases-mutant.log retains the disagreement. No stage1 or
production compiler source was edited. Native execution was not rerun.

Reproduction uses testdata/oracle.go through a Go overlay, as stage1's tests
do. The overlay maps the virtual cohere/TypeScript/tsc/adamic_parser_oracle.go
to stage1/typescript/parser/testdata/oracle.go, leaving the Go parser unchanged.
Build with go build -buildvcs=false -overlay=<overlay.json> -o <oracle> <virtual-file>.
Run <oracle> --manifest <absolute-source-paths> --whole and compare with
node --disable-warning=ExperimentalWarning oracle/node.mjs
stage1/typescript/parser/main.ts --manifest <absolute-source-paths> --whole.
For the stage3 projection, use the existing staged parser-proof-main.a with
node.mjs and witness.a with oracle/node.mjs, passing the same scratch tree and
relative manifest. Inputs are exactly the strings in difference-cases.json.
