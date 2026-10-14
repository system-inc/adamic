# Mutable runtime storage

## Uint16Array extension, October 7, 2026

This unit adds no mutable static runtime storage. Uint16Array uses the existing
counted typed-array header and owner pattern, with a two-byte `uint16_t` buffer
allocated per owning array. A subarray retains its ultimate owner and points
into that buffer. The generic iterator retains the array. Existing heap freeing
releases owners, buffers and iterators; no new global cache or static counter is
introduced. This entry records this extension, not an inventory of earlier units.
