Built: prove Array.isArray predicates over unknown and preserve declared array members at builtin and exact-wrapper call sites.
Branch: codex/host-blockers, based on cd220dd5; main was not merged.
Checks: full lower package, uncached host and Array.from oracles, all recorded counts, vet and formatting passed.
Mutants: predicate tested undefined instead of its argument; declared-member narrowing returned the checker's any[] view. Both caught and restored.
Limits: any remains refused; opaque unknown supports arrayness and narrowed length; tuple erasure and unknown element reads remain named NotYet cases.

The adapted core signature is:

```typescript
function isArray(value: unknown): value is readonly unknown[] {
    return Array.isArray(value);
}
```

The predicate's body proves arrayness, with no element contract stronger than unknown
and no mutable array view. A union guard retains its declared array member rather
than the library signature's any[]. Both direct builtin tests and the exact wrapper
recover that contract. Tests reject fabricated number element predicates, mutable
unknown[] predicates, constant-true predicates, any parameters and shadowed Array.

Unknown parameters and local bindings use an opaque boxed input. Null and undefined
share the missing representation only in this restricted context: both fail the array
brand test. Other observations are refused before emission. Tuple inputs are refused
because their current native object storage lacks the array brand. Unknown array
element reads cannot guess whether the original slots hold numbers or references;
array length is supported. Existing Array.from inputs proven undefined and catches
proven to hold Error retain their existing representations.

Validation, with /workspace/adamic-tools/env.sh sourced and output redirected to logs:

```text
go test ./internal/lower -count=1
  PASS, 22.531s; /tmp/host-array-lower-final.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(host_.*|array_from.*)' -count=1
  PASS, 4.254s; /tmp/host-array-oracle-final.log
go test ./internal/oracle -run TestCountsAreRecorded -count=1
  PASS, 28.216s; /tmp/host-array-counts-final.log
go vet ./...
  PASS, no diagnostics; /tmp/host-array-vet-final.log
gofmt -l [changed Go files]; git diff --check
  PASS, no output
```

Node, emitted JavaScript, sanitized native and release native agree for number and
string arrays, empty arrays, strings, objects with a length, null, undefined, numbers
and booleans. Both unknown parameters and unknown locals are tested, as are indexed
reads from declared number-array unions. The new recorded-count fixture balances
50 allocations and 50 frees.

Typed-array constructors are absent from this worker's reviewed base. The review
fixture host_array_typed_predicate.a is validated in the isolated checkout
/workspace/host-array-typed-review at codex/typed-arrays 692f34ff979fc4c9fbd8f01506872970749cc308,
with this predicate lowering integrated. No typed-array implementation or main merge
was added to codex/host-blockers. Uint8Array, Int32Array and Float64Array, including an
unknown binding holding Uint8Array, are false through both builtin and wrapper calls.
The same ordinary-value tests run in that fixture.

```text
ADAMIC_GATE_UNCACHED=1 go test -buildvcs=false ./internal/oracle -run 'TestNativeAgreesWithNode/review/host_array_typed_predicate' -count=1
  PASS, 1.529s; /tmp/host-array-typed-uncached.log
```

Mutants were run by /tmp/host-array-mutants.py. The predicate mutant changed the
wrapper's runtime tested value to undefined: both emitted backends disagreed with
Node's stdout after successful compilation; the native release mutant also failed
at runtime. The narrowing mutant removed declared-member recovery, so the positive
fixture failed lowering with "an array of any". Neither was caught by a C warning.
Logs: /tmp/host-array-mutants-final.log and /tmp/host-array-mutant-{predicate,narrowing}.log.

The full repository test gate was not run. The original any signature, tuple array
brands, opaque element reads, unknown returns, and general unknown observations are
not covered as working features. No full parser or host fixture 25 build was claimed.
The prior setup completed in 50.778s (build 50.657s); nproc was 5, CPU quota 4.
