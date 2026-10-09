Starting commit: 2b1be38362046455e0e5454d8b0e674a8630e95d

CUT: native emitter and memory planners, build flags/runtime feature selection, runtime/case.c and version macros.
Oracle: executed Node case mapping, version and optional methods; actual clang diagnostics for three witnesses; hand-written IR/C expectations otherwise.

The exhaustive reached Go function inventory is reached-functions.txt (nonzero coverage rows). The C case entry adamic_string_to_upper/lower reaches convert, map_point, simple, full, within, size_at, decode_at, encoded_size, encode, cased_beyond, allocation and length helpers. Generated fixture C also reaches the runtime selected by emitted expressions; this audit mutates only the inventoried native sources.

Menu frozen before any mutation run in menu.json. Three C source rebuilds, all others Go switch builds. Bounded matrix is names.json; whole package exceeded 90 seconds. C callers of the generic emitter elsewhere are outside the matrix and remain unknown.
