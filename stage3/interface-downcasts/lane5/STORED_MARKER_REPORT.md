Built: stored erased-marker calls check actual zero-argument producer signatures and release discarded heap results.
Commits: follows 946b35c5; this group implements stored-marker calls without changing the closure ABI.
Commands/results: uncached callable source oracle passes 21.798s; IR/lower/native/JavaScript focused guards pass; vet passes; counts pass 4.186s.
Mutants: ten source mutants caught, including skipped stored-call arity checks and leaked dynamic string results; every file restored.
Limits: no new candidate-pair certification; 2 pairs/11 reads certified, 306 pairs/1492 reads remaining; observed marker results and mutable write-back remain refused.

Stored AnyFunction values with zero written arguments and discarded results now
check their independently recorded producer identity and arity before invocation.
The read itself continues to check callable presence and kind. The result is
released according to the producer's representation, never the erased view.
Unknown result metadata remains zero; known void is 254; the explicit discard
contract uses 255. Scalar/boxed parameter or valued-result variance stays exact.
Call metadata survives read optimization without introducing a synthetic field.

Node controls cover Boolean, number, dynamically allocated string, object, array
and void producers, a callable field followed by a stored invocation, and a
wrong-arity producer. The wrong-arity message is pinned in release native,
sanitized native and JavaScript as exit 70 with no producer output:

```
adamic: panic: field read failed: callback expected AnyFunction, found function with arity 1
```

The mutable write-back refusal still pins adamic/invariant-mutable. A marker call
whose result is observed remains unsupported. Positive reference-result witnesses
pass leak checks. Allocation counts append ten rows without reordering existing
rows. The full repository gate is not claimed.

Commands, all writing directly to logs:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Callable|StoredMarker)' -count=1 -timeout 10m
go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1 -timeout 10m -args -update-counts
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'TestViewCallable|TestPrepareViewCallable|Test.*CallTarget' -count=1 -timeout 10m
go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
python3 stage3/interface-downcasts/lane5/run-source-mutants.py
```

The canonical witness initially ran while the arity mutant was installed and
failed as expected under that mutant. It was rerun in the unmutated source gate
and passed; that overlap is not an independent regression. New certified pairs
will be recorded in the next group. No upstream cohere source was copied.
