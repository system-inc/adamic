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
