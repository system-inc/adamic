# Concurrency scaling

Stacked on 137afa5. Claims: runtime/string_index.c and share.c for lazy shared string indexes; runtime/parallel.c for adaptive range sizes; runtime/adamic.h and object.c plus runtime slot-cache declarations and the one declaration hook in emit_objects.go for packed atomic caches. Parallel C harnesses and Go tests, measurement scripts and this report will hold the proofs and measurements. No lower or JavaScript changes.

## Implementation

Sharing now traverses and marks the graph without counting string units or building indexes. The first worker that needs a string's UTF-16 view builds a complete private candidate and publishes its pointer with a release CAS. Readers acquire that pointer. A losing builder frees both its BMP view and its candidate. Length, checkpoints and the BMP view are published together; the original plain length field remains immutable after sharing. Shared cursors remain disabled. Empty, short and ASCII shared strings use a small length-only candidate. Unshared strings retain their existing caches and cursor.

Warm shared length and BMP reads are inline. The string index layout is internal to the C runtime; moving it into adamic.h lets a BMP read use one acquired pointer without another runtime call. string_decode_impl.h and string_slice_impl.h use the same published length helper. The three-argument parallelMap ABI and adamic_share ABI are unchanged.

Each scope chooses `max(1, min(256, items / (8 * threads)))`. Stealing and range claiming use that scope's grain, including nested calls. At 4,096 items and 16 threads this is 32 rather than 256; at a million items and four threads it remains 256. `adamic_parallel_grain` is a private runtime observation in parallel.h, not a compiler ABI.

A slot cache is now one naturally aligned, lock-free 64-bit word: the shape address in the low 48 bits and the slot index in the high 16. Reads and writes are relaxed atomics. Field and method lookup use one snapshot; method lookup does not reread the cache after another thread can replace it. A compile-time assertion requires lock-free 64-bit atomics, and a runtime check refuses a shape address outside 48 bits with `adamic: panic: shape address exceeds 48 bits`. Slots greater than 65,535 bypass the cache. The emitted declaration and the caches in map.c and exceptions.c are ordinary static caches; library_object.c's local initializer is adjusted to the new representation.

## Proofs

The existing parallel memory, map, nested-map, exception, lifecycle, panic and million-element harnesses run under ASan/UBSan, sanitized slabs, counted, malloc and TSan builds. The scaling boundary harness additionally checks that sharing leaves caches cold, that empty/short/ASCII/indexed strings read correctly, the grain choices, large-slot fallback, and the address guard. Alternating shapes stress both field caches and dispatch between an object's closure field and a class method.

A snapshot-only hook holds four builders immediately before CAS publication. This forces exactly three losing candidates without adding instrumentation to production. It passes each build variant; omitting the losing-copy frees fails LeakSanitizer. Darwin uses the existing unsanitized malloc build under `leaks --atExit -- <binary>`; this Linux worker cannot run macOS tests.

The first full gate exposed a scheduling-sensitive new mutant: a plain read of the cache pointer was not always observed racing with publication. That detector was replaced by non-atomic publication in the forced-four-builder harness. The final report records the repeated proof and reruns rather than treating the original mutant as sufficient.

Measurements and final gate results follow below.
