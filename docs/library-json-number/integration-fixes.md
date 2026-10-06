# Integration's four JSON.stringify probes

## 1. Generic bodies

The supplied A and B are preserved in `library_json_generic.a` and
`library_json_generic_shared.a`. Node prints `"xx"` then `[1,2]` for A,
and `3` then `[1,2]` for B. Both now receive a compile-time NotYet naming
per-instantiation container metadata. JSON types are resolved with `l.concrete`.
Generic union bodies also need different cache keys: a scalar union can otherwise
share its descriptor with a later container union. Until that metadata is part
of the instantiation key, generic union serialization is conservatively refused.

`ADAMIC_GATE_UNCACHED=1 go test -v -count=1 ./internal/oracle -run
TestLibraryJSONGenericInstantiations` passed (0.275s). Removing the concrete
resolution admitted both programs, and the same check failed both (0.264s).
The mutant was restored. These fixtures are compile-time refusals, so they have
no native allocation counts. Their Node results are checked independently.

## 2. NUL in the gap

Exact probe C is `library_json_gap_nul.a`, registered with the shared oracle
and allocation-count gate without editing integration's `oracle_test.go`.
Native now cuts the first-ten-code-unit gap at NUL, but preserves V8's
multiline path when that cutoff leaves an empty indentation string. The
runtime comment names the departure from ECMA-262 25.5.2's gap step and
SerializeJSONProperty indentation. Node's length is 17.

The uncached full oracle comparison for this fixture passed (8.695s).
Removing only the cutoff compiled and exited 0 with empty stderr, leaked
nothing, and failed only the Node stdout comparison: native length 27,
Node length 17 (8.849s). The mutant was restored.
