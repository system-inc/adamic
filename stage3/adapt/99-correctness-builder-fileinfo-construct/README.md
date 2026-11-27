# Build the cast branch's file info fresh

Compiler's Nov 27 pick on #gbr6vs2, for Adamic's accept, covers
`toBuilderStateFileInfoForMultiEmit` at upstream and adapted `builder.ts:2262`.
When a multi-file build info's file info carries a string signature, tsc hands the
build info's own object back as `fileInfo as BuilderState.FileInfo`. That is an
unchecked cast from `IncrementalMultiFileEmitBuildInfoBuilderStateFileInfo`, whose
`signature` is `string | false | undefined`, to `BuilderState.FileInfo`, whose
`signature` is a writable `string | undefined`. The adaptation builds the result
the way the function's third branch already does:
`{ version: fileInfo.version, signature: fileInfo.signature, affectsGlobalScope: fileInfo.affectsGlobalScope, impliedFormat: fileInfo.impliedFormat }`.
The unchecked cast leaves tsc's source, so the site needs no checked view.

It is JavaScript-changing in two ways, both measured over the whole lane oracle
before it was built:

- Identity. The result is a fresh object, not the build info's. J-8ddpt0g froze every
  result and tagged it by branch: 13,191 results, 5,548 from this branch, 14,966
  `signature` writes, 0 on a result, and the frozen oracle matched the lane exactly.
  J-4dma35b ran the oracle with this change in place and it matched exactly. By
  reading, every reader of a result or its build-info object reads fields only
  (`builder.ts:2290`, `:2407`, `:387`, `:440`).
- Key presence. A build info's file info omits `affectsGlobalScope` and
  `impliedFormat` when they are undefined: of the 5,548 inputs, 5,051 carry only
  `version` and `signature`, 433 add `affectsGlobalScope`, 61 add `impliedFormat`,
  and 3 carry all four (J-4dma35b). The fresh object carries all four, present and
  undefined where the input omitted them, as `BuilderState.FileInfo` declares them
  required and as the third branch builds them. No input carries a fifth key, so
  nothing is dropped.

The adapter pins the whole `toBuilderStateFileInfoForMultiEmit` declaration by kind,
name, before and after text, and SHA-256. The change sits inside the function body.
The snapshot test compares stock public declaration emit byte for byte.

From the repository root, run:

    NODE_PATH="$PWD/stage3/api/node_modules" node stage3/adapt/99-correctness-builder-fileinfo-construct/test.cjs TREE

The test reads both snapshots and runs both function bodies in Node on a string
input, the cast branch in each key order the lane hands it, and the built branch.
It checks each of the four fields is the same value before and after, the
serialized result is the same, and the input is left alone. On the cast branch it
also checks the result is a fresh object carrying the four keys in order, and that
a write to the result doesn't reach the build info. Loom runs the sealed lane, the
`builder.ts` census and the restoring patch.
