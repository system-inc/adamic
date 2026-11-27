# Use local stubs with the receiving callable signatures

Adamic ruled this adaptation's shape on Nov 22 and 23, 2026 (#bbssdcy,
family 34): the resolver's throwing slot at `emitter.ts:1175:23` and
19 shared-stub slots in the `nullTransformationContext` initializer
(13 `noop`, 6 `notImplemented`). She accepts the head after its review
and Loom's unit census at main. The Nov 22
ruling on #bbssdcy requires tracing readers and preserving arity and
the thrown message. Adamic's first Nov 23 answer ruled the transformer
noop in, and extended the rule to `enableSubstitution`. Her second Nov 23
answer extended the rule to the remaining shared noop and throwing slots.
The adapter locates the `EmitResolver` object `notImplementedResolver`
and the transformer initializer, checks the exact slot snapshots, and
checks stock TypeScript's emitted JavaScript against the expected changes.

[stub-readers.cjs](stub-readers.cjs) covers the resolver,
[slot-to-baseline.cjs](slot-to-baseline.cjs) covers the transformer
slots at `:679` and `:689`, and [remaining-stubs.cjs](remaining-stubs.cjs)
covers the other 17 transformer slots.

`docs/rulings.md` row 347 describes value and reference storage. Rows 363 and
364 describe Weak storage and row 365 includes callable slots in storage
comparison. Row 392 says “NotYet means intended support.” Row 402 gives the
meaning-preserving meter adaptation constraint.

## Checker references

Source citations below use this adapter's after snapshot at the pin.

The recorded TypeScript 6.0.3 checker and language service trace is in
[stub-readers.cjs](stub-readers.cjs). It resolves the wrapper call to the
shared declaration through the stub's import and traces the changed slot,
its contextual property, and the object carrying the slot.

- `src/compiler/program.ts:2656` passes `notImplementedResolver` into
  `emitFiles` with `onlyBuildInfo` set to `true`. The emitter only emits
  build info on that path.
- `src/compiler/transformers/ts.ts:2743` reads the `EmitResolver` property
  by calling it, `resolver.getConstantValue(node)`. The emitter's separate
  printer helper and the `TypeChecker` property have different symbols.
- `src/compiler/emitter.ts:1172` initializes the slot. The checker reports
  the declaration in `src/compiler/types.ts:5934` and the real resolver's
  implementation in `src/compiler/checker.ts:51439`.

The checker trace covers readers in the TypeScript source tree. Those
readers only call the slot or pass its containing object. On Nov 23,
Adamic ruled that a declaration file cannot promise a member function's
`name` or identity. Her second answer rests on review-a/10's
outside-consumer record. Node ran both functions and printed the following.

```text
notImplemented shared length = 0
notImplemented shared message = "Not implemented"
notImplemented local length = 0
notImplemented local message = "Not implemented"
```

## Transformer slot

Adamic's round 2 ruling on #bbssdcy limits the identity condition to
readers reachable from the slot. The whole-src checker trace is
[slot-to-baseline.cjs](slot-to-baseline.cjs).

- `src/compiler/transformer.ts:679` initializes the slot. The context
  is the default argument of `visitEachChild` at
  `src/compiler/visitorPublic.ts:598`. The visitor passes the context into
  its syntax-kind dispatch at line 604.
- `src/services/textChanges.ts:1380` copies the context by spread.
  Line 1389 passes that resulting context to `visitEachChild`.
  Lines 1382 and 1383 read the factory's flags and base factory.
- The slot has two value readers through `TransformationContext`.
  `src/compiler/visitorPublic.ts:399` calls it with
  `(LexicalEnvironmentFlags.InParameters, true)`. Line 414 calls it
  with `(LexicalEnvironmentFlags.InParameters, false)`. Both calls
  discard the void result. Neither transfers the function.
- `src/server/editorServices.ts:1432` initializes
  `ProjectService.baseline` directly from the shared `noop` declaration.
  `src/harness/projectServiceStateLogger.ts:81` writes a locally created
  logging arrow. Line 73 compares the service field with the shared
  `noop`. The initialization and logging arrow are the checker-resolved
  writes. The harness and virtual-file-system readers call the field.

The checker trace finds no in-tree route from the transformer slot to
`service.baseline` and no identity, name, or length reader in the covered
checker-assignable flows. Its indirect lookup scan excludes `any` and
`unknown` receivers and checks a fixed list of reflective operations on
the first argument. Node
measured shared and local length as 1. The default on the second parameter
keeps the local function's arity at 1.


## Substitution slot

The Nov 23 ruling extends the noop adaptation to
`src/compiler/transformer.ts:689:25`. The same whole-src trace includes
all 38 harness roots and follows the contextual `enableSubstitution`
property. Its 30 runtime reads call the function directly and discard
the void result. The separate shorthand at transformer.ts:291
installs `transformNodes`'s own function. The context default and
spread paths above carry the slot into the visitor. The checker trace
finds no in-tree route to `service.baseline` and no identity, name, or length
reader in the checker-assignable flows covered by the same scan. Node
measured shared and local arity as 1.

## Remaining shared stubs

The Nov 23 follow-up on #bbssdcy applies the same rule to every remaining
`noop` and `notImplemented` slot in the initializer. The whole-src
[remaining-stubs.cjs](remaining-stubs.cjs) trace checks the contextual
property symbols, accepting only direct calls and the declarations or
initializations of those properties. It loads 713 roots, including 38
from the harness. The context and spread-alias flows and the bounded
indirect lookup scan above apply to the same receiver here. The trace
reports no runtime references for the two boolean predicates on
`TransformationContext`; the identically named emitter hooks belong to
other objects.

Node evaluates the actual local arrows alongside the original shared
functions. All 11 noops keep length 1 and return undefined. All six
throwing stubs keep length 0 and throw `Error: Not implemented`. Four
zero-argument noop slots keep the shared optional parameter. Two
throwing predicates use a defaulted parameter to keep length 0.

The transformer-file census moves unit `src/compiler/transformer.ts:669:1`
from its Round 5 first stop at `:690:29` to no finding in the unit.
Rounds 4 and 5 listed `:690:29` (`enableEmitNotification: noop`) as
this unit's exposed stop. Round 6 cleared it.

## Exposed stops

The current inventory from Round 6's census of unit
`src/compiler/transformer.ts:669:1` holds no exposed stop, and
[exposed-stops.jsonl](exposed-stops.jsonl) is empty.

For each stop a later census exposes in this directory's units, add a
row here with its family and unit, copy the census record's own `text`
verbatim, and add the record to `exposed-stops.jsonl`. At main, the callable
diagnostic follows d36ba425 (#bjtwhg8) and names the compared types, the
slot, and its two representations. A slot boxed in one view and unboxed
in the other describes the diagnostic family. The row carries the
specific finding from that run. Historically, the old pin used
“one keeps something weakly that the other keeps strongly.”

| Family | Unit | Census finding | Disposition |
| --- | --- | --- | --- |

## Replay the checker evidence

With the locked packages installed in `stage3/api`, pass a tree that
`stage3/apply.sh` wrote to [replay.cjs](replay.cjs).

```sh
node stage3/adapt/99-family34-local-stubs/replay.cjs /path/to/applied-tree
```

Replay checks the exposure fixture against the README rows and runs the
three traces on the source archived from this adapter's `-after` ref.
Later adapters leave this snapshot untouched. Earlier adapters may move
traced lines while preserving the recorded readers and behavior.

At main 934baf81, two later adapters change traced lines in the final
tree, which the replay does not read.
`99-seats-scout24-whole-built-hosts` stores the shared `noop` as the
`setModifiedTime` and `deleteFile` fallbacks at final
`src/compiler/tsbuildPublic.ts` lines 352, 353, 401 and 402, as tsc's
own host fallbacks in that file do.
`99-seats-build22-62-node-array-range` gives the `context.factory` line
at final `src/compiler/visitorPublic.ts:449` new text. Neither reads a
stub slot or an object carrying one, compares `noop`, or reads its
`name` or `length`. Adamic read both on Nov 26 (#qv62557).

Replay compares each transcript byte for byte with its record
(`stub-readers.expected.txt`, `slot-to-baseline.expected.txt`, or
`remaining-stubs.expected.txt`) after masking the numeric line and column
coordinates of source paths as `<coordinate>`.
