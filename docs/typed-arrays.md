# Typed arrays runtime contract

October 7, 2026: Uint8Array, Int32Array and Float64Array only. The C contract is
`internal/native/runtime/adamic.h`; reads return `adamic_maybe_number`, agreed
with the compiler so lowering can reuse plain-array read paths.

`adamic_typed_array` begins with `adamic_heap`, followed by element `kind`,
fixed element `length`, `data` and `owner`. Data is one flat buffer with 1, 4 or
8 bytes per element, never boxed. An owning header has NULL owner; a subarray
has its own count, points into the original buffer, and retains its ultimate
owning header. Releasing the last header/view releases the buffer. Unlike
string sharing, even short or empty subarrays are views, never copies.

| Entry point | Contract |
|---|---|
| `adamic_typed_array_new(kind, length)` | Owned, zero-filled; JavaScript ToIndex of length, with RangeError panic for invalid length |
| `adamic_typed_array_from_numbers(kind, numbers)` | Owned; borrowed plain number[] copied with element conversion |
| `adamic_typed_array_get(array, index)` | Present number at an existing integer index; absent (`present = false`, number initialized to 0) otherwise; numeric -0 addresses index 0 |
| `adamic_typed_array_check_write(array, index)` | Compiler write check; rejects absent indexes |
| `adamic_typed_array_set(array, index, value)` | Checks index itself, then converts and writes |
| `adamic_typed_array_length(array)` | Element count as double |
| `adamic_typed_array_fill(array, value, start, end, has_start, has_end)` | Converts value, uses relative clamped indexes; returns borrowed self |
| `adamic_typed_array_set_from(array, source, offset, has_offset)` | Same kind only; offset defaults to 0, ToIntegerOrInfinity, negative or insufficient room panics with RangeError; overlap behaves as a temporary copy |
| `adamic_typed_array_subarray(array, start, end, has_end)` | Owned view; relative clamped indexes, omitted end is length; start defaults to 0 at call site |
| `adamic_typed_array_iterate(array)` | Owned counted iterator retaining array |
| `adamic_typed_array_iterator_next(iterator, value)` | Reads next current element; true with output value, false when exhausted |

All arguments are borrowed. Constructors, subarray and iterate return one
reference that the compiler releases. Iterators must be released on normal,
early and exceptional exits. Fill returns no additional count. Optional numeric
arguments use explicit presence flags; an explicit undefined is lowered as absent.

Uint8 writes use ToUint8: truncate and wrap modulo 256, NaN and either infinity
become 0. Int32 writes truncate and wrap modulo 2^32 into signed range. Float64
stores the double unchanged, including negative zero and NaN. Out-of-range reads
are undefined. Out-of-range writes panic with the existing plain-array shape:
`adamic: panic: index <index> is outside an array of length <length>` and exit 70.
Node silently drops typed-array writes outside the buffer; the JavaScript backend
must insert this same Adamic check before its write.

ArrayBuffer, DataView, other element types, constructing views over another
view's buffer, and sharing across parallel tasks remain NotYet in the compiler.
This does not prohibit subarray on a subarray: it retains the ultimate owner
without growing an owner chain. No buffer property or buffer constructor is exposed.
