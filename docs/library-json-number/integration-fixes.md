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
