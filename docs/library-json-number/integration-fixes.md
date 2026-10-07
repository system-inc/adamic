# Integration's four JSON.stringify probes

## 1. Generic bodies

The supplied A and B are preserved under `docs/library-json-number/probes/` in
`library_json_generic.a` and
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

## 3. Runtime shapes

Exact probe D is `library_json_shapes.a`, registered with the shared oracle and
counts gate. The uncached Node/native/backend/release/leak comparison passed
(0.444s). All three requested runtime-line mutants compiled and were caught:

| Mutant | Check that failed |
|---|---|
| Remove the replacer NULL guard | UBSan: member access within NULL array in keys |
| Remove arrays from scalar's NULL-to-undefined list | UBSan: member access within NULL array in write_value |
| Read a union closure as a Map | Node stdout comparison; both executions exit 0 with empty stderr |

The script `run-stringify-mutants.py` runs them independently, records complete
logs under `/tmp/library-json-number-shapes-mutants/`, and restores the runtime
in a finally block. No mutated source is committed.

## 4. Unreachable duplicate-literal check

Removed the lowering duplicate-key branch and the test case that returned
success on TS1117 before reaching it. TypeScript continues to reject repeated
literal property names; that checker behavior is not claimed as evidence for
a lowering guard. No runtime behavior changed.

The uncached focused suite passed all eleven stringify fixtures, both generic
probes and all six remaining refusal probes (1.713s). The six refusal probes
and two generic probes also passed with their own root filter (0.315s).

The first full Linux gate caught a fixture-location error: flow tests glob all
root oracle fixtures as compilable programs. The two deliberately refused
generic probes now live under this unit's docs, and their dedicated test still
compares Node output and requires the compile-time refusal. No flow test was
changed or weakened.
