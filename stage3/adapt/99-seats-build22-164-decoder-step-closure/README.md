# Step the source-map decoder directly in appendSourceMap

Family 164 on the Nov 22 meter, #bbssdcy. The for-of in
`src/compiler/sourcemap.ts:184:27` iterates the decoder object created at
183:33. The decoder starts at 464:1. This applies Adamic's Nov 22 ruling
and her Nov 23 07:21:37Z answer covering the harness. This applies rows 206, 207 and 290,
the last two as they stand. At ac44317c, Adamic's ruling counts the
iterator and prelude rows as 205, 206, 288, 289 and 290.

`appendSourceMap` takes the `step` closure from the file-local
`createMappingDecoder` factory. Each loop iteration calls it once and
checks `done` before reading `value`. The existing end-position break
and start-position continue retain their order. The step runs the same
decoder body against the same captured state. A yielded mapping has
`done: false`, as the original `while (!done && ...)` body guaranteed.
The completion result retains family 156's own `value: undefined`.

The file-local factory also returns three state-reading closures.
The exported internal `decodeMappings` builds the iterator around them.
Its `next()` method calls the same step closure. Its `pos`, `error` and
`state` getters retain their reads through those closures, and
`[Symbol.iterator]()` still returns the iterator itself.
`appendSourceMap` reaches the step without constructing an iterator. The wrapper
has the original members and no return hook. Its declaration and
`MappingsDecoder` are marked `@internal`.

The five callers on the input snapshot are:

- `appendSourceMap` at `sourcemap.ts:183:33`, the direct step consumer.
- `getDecodedMappings` at `sourcemap.ts:742:29`, which consumes
  `arrayFrom(decoder, processMapping)` and reads `decoder.error` twice.
- `SourceMapDecoder.initializeSourceMapDecoding` at
  `src/harness/sourceMapRecorder.ts:24:23`. Its helpers call `next()`,
  take `value` on yield, read `error` and `state` on done, and read `pos`
  independently to check completion.
- `getSourceMapRecord` and `getSourceMapRecordWithSystem` in the same
  harness file at 305:27 and 340:27. Both drive `next()` by hand and read
  `value` only on a yield.

The adapter locates the two functions through the TypeScript 6.0.3 AST.
The adjacent text files pin their complete input on family 156's output.
A second run accepts its own output. An unexpected or partly edited
snapshot fails before writing.

## Manual pin

Use a full apply output containing
`refs/stage3/99-seats-build22-164-decoder-step-closure-before`, or the
full family 156 output directly.

```sh
NODE_PATH="$PWD/stage3/api/node_modules" node \
  stage3/adapt/99-seats-build22-164-decoder-step-closure/test.cjs \
  /tmp/adapted-tsc /tmp/family-164-wrapper.json
```

The test copies `sourcemap.ts` into an OS temporary tree, runs the
production adapter, and loads stock CommonJS transpilation of each side
in a fresh Node VM. It checks the wrapper's exact own keys, self-iterator,
hand-driven steps and getters, including an additional step after done.
It also runs the actual append loop on three maps and compares emitted
mappings and code-unit reads. The mapper's malformed input exercises
error logging on two lookups. The JSON path receives all traces, and a
differing row exits 1.
