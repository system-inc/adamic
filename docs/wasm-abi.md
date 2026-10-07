# Adamic WebAssembly ABI

Version: **1**. Target: wasm32-wasi, WASI Preview 1, single-threaded.
This contract describes the compiler-generated boundary, independent of Node.

## Build and discovery

```sh
adamic build --target wasm32-wasi --reactor entry.a -o out.wasm --abi-json out.json
adamic build --target wasm32-wasi entry.a -o out.wasm --reactor \
  --export first --export second --abi-json out.json
```

`--reactor` is accepted immediately after `--target wasm32-wasi` or after the
output path. Export selection and ABI JSON options follow the output path.

**Release flags.** wasm32-wasi builds default to clang `-Oz` for generated C
and the runtime, and `-Wl,--strip-debug` at link. Debug custom sections are
removed; the `name` section remains for profiles. Counted builds use the same
WASI defaults and retain their counters. Native builds keep their existing
flags. An optional post-step is `wasm-opt -Oz -g out.wasm -o optimized.wasm`
(`-g` preserves function names). Binaryen is a separate installation and is not included in the WASI SDK.

The reactor exports every entry-module `export function` in the crossing set.
Repeatable `--export` selects only those names and refuses missing names or an
unsupported signature, including its failing field path. Automatic selection
omits signatures outside the crossing set. Export aliases, dependency exports,
function-valued variables, generics, optional/default/rest parameters, literal
parameter types and classes are outside this boundary. Function declaration
names must currently be unique across the lowered modules; ambiguous names are
refused rather than guessed. The language's usual lowering refusals still apply.

A wasm build with crossing exports automatically becomes a reactor. With no
crossing exports and no reactor option it remains a command. `--reactor` permits
an empty export set. `_initialize` runs runtime constructors and module top level
once, through `adamic_module_start`; calls do not run `main`. Module globals remain
alive for the instance's lifetime.

Each application export is named `adamic_export_<name>`. `handleRequest` is an
ordinary export with the same rules. For its existing `(string) -> string`
signature, the earlier `adamic_request`/response helper API remains available for
existing hosts; its input decoder retains its earlier replacement semantics.

## Crossing types and wasm signatures

Parameters and results can be `number`, `boolean`, `string`, `readonly number[]`,
`readonly string[]`, and readonly records described by object aliases or interfaces.
Every record field must be required, readonly and itself in this set. A root
record is depth 1; a record field is depth 2; a record inside that is refused.
Arrays of records, unions, tuples, maps, callable objects, index signatures and
optional fields are refused. A result can additionally be `void`.

| Type | Parameter | Result |
| --- | --- | --- |
| number | f64 | f64 |
| boolean | i32, 0 or 1 | i32, 0 or 1 |
| string | i32 pointer, i32 byte length | i32 result handle |
| readonly number[] | i32 pointer, i32 element count | i32 result handle |
| readonly string[] | i32 pointer, i32 byte length | i32 result handle |
| readonly record | i32 pointer, i32 byte length | i32 result handle |
| void | unavailable | no wasm result |

Arguments appear in source parameter order. Each buffer parameter expands into
two adjacent wasm arguments. Direct number-array input is contiguous,
8-byte-aligned little-endian f64, without a count prefix. `adamic_alloc` returns
sufficiently aligned memory. All pointer, count, size and handle values are
unsigned 32-bit bit patterns; JS hosts use `>>> 0` when reading them.

## Schema encoding

There are no tags. Both sides must use the exact generated signature. Each
buffer starts at offset 0. Alignment is relative to that buffer's start, including
inside nested records; it does not restart at a field or a nested record.
Padding bytes written by the reference host and compiler are zero.

| Type | Schema encoding |
| --- | --- |
| number | pad to next offset divisible by 8, then f64 LE |
| boolean | one byte, 0 or 1 |
| string | u32 LE byte length, then those bytes |
| number[] | u32 LE element count, pad to 8, then count f64 LE values |
| string[] | u32 LE element count, then count schema strings |
| record | each field's schema encoding in the JSON fields array's declaration order |

A top-level **string parameter or string result is raw bytes**, without the
schema length prefix. A number-array parameter is likewise the direct form
above. Other buffer parameters and array/record results use schema encoding.

Strings use UTF-8 for Unicode scalar values. Adamic output uses WTF-8 to preserve
UTF-16 code units: a lone U+D800 becomes bytes `ed a0 80`, and the reference
host decodes it back to the lone U+D800 code unit. It does not replace it with
U+FFFD. Input also accepts this WTF-8 extension, enabling exact JS round trips
of lone surrogates. The reference host joins valid surrogate pairs into one
four-byte code point. Do not use TextEncoder/TextDecoder for lossless strings:
TextEncoder replaces lone input surrogates, and TextDecoder rejects their WTF-8
bytes. Malformed or noncanonical input bytes panic rather than enter the runtime
as malformed string storage. The old request API continues to replace malformed
UTF-8 as documented in docs/wasm.md.

### Worked examples

These bytes apply to schema values unless the direct form is specified.

* `number` 1: offset 0, `00 00 00 00 00 00 f0 3f`. Direct scalar calls pass f64 1
  instead of allocating this buffer. NaN and negative zero retain IEEE semantics.
* `boolean` true: `01`. Direct scalar calls pass i32 1.
* `string` "A": `01 00 00 00 41`. A direct string argument/result is only `41`;
  its separate byte length is 1. An empty direct string has byte length 0.
* `readonly number[]` [1, -0]: `02 00 00 00`, four zero padding bytes, then
  `00 00 00 00 00 00 f0 3f` and `00 00 00 00 00 00 00 80`. Direct input omits
  the first eight bytes and passes element count 2. An empty schema number array
  is the zero count plus four padding bytes, even when there are no elements.
* `readonly string[]` ["A", ""]: `02 00 00 00 01 00 00 00 41 00 00 00 00`.
* Record `{ readonly enabled: boolean; readonly score: number; readonly text: string }`
  with `{enabled: true, score: 1, text: "A"}`: byte `01` at offset 0; seven zero
  padding bytes at offsets 1..7; f64 1 at 8..15; `01 00 00 00 41` at 16..20.
  A containing `{ readonly inner: RecordType }` has exactly those same bytes:
  nested records add neither a tag nor a length nor a new alignment origin.
* A void result has no bytes, handle or wasm result.

Record encoding follows declaration order, regardless of the runtime object's
construction order. The JSON schema is the authority for generators, including
inherited interface properties. An empty record has zero bytes.

## Signature JSON

`--abi-json out.json` writes UTF-8 JSON after a successful build:

```json
{
  "version": 1,
  "exports": [
    {
      "name": "serve",
      "parameters": [{"name": "input", "type": {"kind": "string"}}],
      "returns": {
        "kind": "record",
        "fields": [
          {"name": "ok", "type": {"kind": "boolean"}},
          {"name": "body", "type": {"kind": "string"}}
        ]
      }
    }
  ]
}
```

Exports are sorted by name. Parameter arrays and record field arrays are ordered;
JSON object key order carries no meaning. Type `kind` is one of `number`,
`boolean`, `string`, `number[]`, `string[]`, `record`, or result-only `void`.
Only a record has `fields`; an empty record may omit `fields`, meaning `[]`.
There are no compiler-internal function IDs in this file. The export name in the
wasm table is obtained by prepending `adamic_export_` to `name`.

## Ownership and instance lifecycle

The module exports `memory` and:

```text
adamic_alloc(size: i32) -> i32
adamic_free(pointer: i32) -> void
adamic_result_bytes(handle: i32) -> i32
adamic_result_length(handle: i32) -> i32
adamic_result_release(handle: i32) -> void
```

The host allocates input buffers, writes their bytes, calls synchronously, and
frees every input exactly once after the call. Inputs are borrowed for the call;
the runtime makes owned values as needed. An empty input still uses a valid
allocation, with length/count 0. Keep number-array buffers aligned to 8.

A non-scalar return is an opaque owned result handle. Its bytes are independent
of all temporary Adamic values. Read its pointer and byte length, decode or copy
the bytes, then release the handle exactly once. Zero-length results still have
a handle; their byte pointer can be zero. Releasing invalid/stale handles or
input pointers is host misuse. There is no implicit release and no garbage
collector. All per-call temporary runtime values and statement regions are
released before normal return; the result bytes and handle survive until host
release. Values deliberately retained in module globals survive with the module.

Any allocation or exported call may grow linear memory and detach old JS views.
Reacquire `memory.buffer` after every allocating call, including subsequent
argument allocations. Decode result bytes before releasing their handle. Do not
retain a wasm-memory view across a call or a release. Pointer offsets remain
valid through memory growth while their allocation remains owned.

A counted build retains `adamic_live()` and `adamic_regions()`. Live counts track
Adamic values, not host allocations or ABI result buffers, so memory measurements
are also necessary to prove result release. Compare live counts with the initialized
module baseline if module top level retains dynamic values.

An uncaught throw or panic exits through WASI **proc_exit 70**. The host must
**discard the instance**; it must not call cleanup exports or reuse that instance.
An engine trap also requires discarding it. The reference wrapper marks trapped
calls failed and refuses further calls. Do not execute calls concurrently or
reenter the same instance. Initialize a new instance exactly once before calling.

## Imports and reference host

Provide the WASI Preview 1 imports listed by `WebAssembly.Module.imports(module)`.
The tested fixture imports `fd_close`, `fd_prestat_get`, `fd_prestat_dir_name`,
`fd_seek`, `fd_write` and `proc_exit`, all from `wasi_snapshot_preview1`. Programs
using filesystem or other runtime services can import additional WASI functions;
a signature alone does not determine imports. No filesystem preopens are implicit.
Output imports must have real host behavior; proc_exit must terminate the instance.

`internal/native/wasm/exports.mjs` provides `wrapExports(instance, table)` and
`instantiateExports(module, table, imports, initialize)`. The latter takes a
compiled WebAssembly.Module; initialization defaults to `_initialize()`. With
Node WASI, pass `instance => wasi.initialize(instance)` so WASI binds memory
before constructors run. Wrap an already initialized instance with `wrapExports`.
The returned functions are ordinary synchronous JS functions. Encoding/decoding
helpers implement the documented schema, and cleanup runs even if decoding fails.
The host requires inputs matching the table; it is not a general JS type validator.
