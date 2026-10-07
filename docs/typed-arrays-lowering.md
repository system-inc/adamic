# Typed arrays: compiler claim

Compiler owner: codex/typed-arrays. Runtime owns docs/typed-arrays.md and
internal/native/runtime/adamic.h on codex/typed-arrays-runtime. This claim
precedes implementation. Native emission waits for runtime's interface commit;
its published names and signatures supersede the provisional names below.

## IR and lowering

Add counted representations Uint8Array, Int32Array and Float64Array. Each
identifies the width and conversion rule without boxing elements. Add
TypedArrayNew (kind, length or number[] source), TypedArrayFill (receiver,
value, optional start/end), TypedArraySet (receiver, same-kind source, optional
offset), and TypedArraySubarray (receiver, optional begin/end). Preserve
JavaScript evaluation order, evaluating each receiver and argument once.

Reuse ArrayIndex with Element Number, SetIndex with Element Number, Length,
and ForOf. Index reads return MaybeNumber, including undefined past the end,
and reuse the existing optional-number narrowing and array-read paths.
Writes keep the existing loud index check and its text:
`index <index> is outside an array of length <length>`, exit 70.

## Runtime seam, pending its interface SHA

Expect a counted `adamic_typed_array *` for each representation, with a flat
buffer of its element width. Proposed operations (runtime's header decides
final spelling):

- `adamic_typed_array_new(kind, double length)`: owned, zero-filled result;
  JavaScript ToIndex length conversion and validation.
- `adamic_typed_array_from(kind, adamic_array *source)`: borrowed number[]
  input, owned result, conversion per element.
- `adamic_typed_array_at(array, double index)`: borrowed receiver,
  `adamic_maybe_number` result by value, no element allocation.
- `adamic_typed_array_set_index(array, double index, double value)`: borrowed
  receiver, checks bounds before storing with the kind's JavaScript conversion.
- `adamic_typed_array_length(array)`: borrowed receiver, numeric length.
- `adamic_typed_array_fill(array, value, start, end)`: borrowed receiver,
  mutates, JavaScript relative-index normalization; compiler returns receiver
  with ordinary ownership when the expression's result is kept.
- `adamic_typed_array_set(array, source, offset)`: both receivers borrowed,
  same kind, overlap-safe copy; JavaScript offset conversion and range failure.
- `adamic_typed_array_subarray(array, begin, end)`: borrowed receiver, owned
  counted view retaining the buffer owner, always shares, including nested and
  empty views. Omitted end means length; compiler preserves omitted arguments.

All ordinary reference paths retain/release the view through adamic_retain and
adamic_release. ForOf holds the receiver for the loop, loads doubles without
boxing, and releases it on normal and abrupt exits. Integer stores use ToUint8
or ToInt32; Float64 stores the double, including negative zero and infinities.

## JavaScript backend

TypedArrayNew emits `new Uint8Array(source)`, `new Int32Array(source)` or
`new Float64Array(source)`, where source is length or number[]. ArrayIndex
emits ordinary bracket access. SetIndex uses existing `adamicSetIndex`, which
checks the index then lets the JavaScript typed array convert the store.
Length emits `.length`; fill, set and subarray emit the corresponding built-in
method with the original optional arguments. ForOf uses JavaScript iteration.
Built-in subarray preserves shared writes in both directions.

## Refusals and proof

Give separate NotYet messages for ArrayBuffer, DataView, construction over
another view's buffer, resizable buffers, unsupported element types, and
sharing typed arrays across parallel tasks. Unsupported members and detached
methods must also stop during lowering, never become generic object reads.

Run Node/native/JavaScript oracles for zero fill, conversion edges (300, -1,
NaN, 2^31, -0, Infinity), reads, fill, overlapping same-kind set, relative
subarrays, nested-view lifetimes, alias writes and for...of. Pin the deliberate
out-of-range write stop separately, and pin Node's silently dropped write.
Run Workers primes and stats with Uint8Array and Float64Array using the sieve
at origin/codex/workers-bench-handler 0a3e651. Run missing-wrap,
missing-bounds-check and copying-subarray mutants and report what caught each.
