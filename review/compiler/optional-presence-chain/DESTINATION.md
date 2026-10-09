Built: destination optional metadata restricts undefined writes to optional string slots, preserving required string storage.
Commit: one correction on compiler/optional-presence-chain, following 84d03d57.
Validation: focused Node/native sanitized/JavaScript comparisons, optional regressions, counts, CallTargetReaders and lane checks are green.
Mutants: dropping the destination optionality check is caught by the required-string exit-70 expectation; dropping construction optionality is caught by the optional Node fixture.
Remaining: shared gate and runtime review of these metadata hunks; no local blocker.

The chain had current presence and representation metadata, but no persistent declaration bit distinguishing a present optional string from a required string. IR Field.Optional now records the object's construction declaration. Missing optional reservations carry it too. The native shape type metadata uses bit 8 for it, independently of the low-byte representation tag and insertion-order presence. Shape deduplication includes this bit, and reserving copies preserve raw metadata. Ordinary objects gain no bytes. JavaScript carries the same declaration in a private WeakMap and preserves it through spread. Current presence or the alias's optional declaration cannot grant permission to store undefined into required string storage.

The required witness constructs { s: string }, hides the field behind a base interface and creates an admitted checked optional view. Source Node prints writing, stored, undefined and exits zero. Native release, sanitized native and JavaScript stop at the write, before stored is printed, with exit 70 and this exact expectation:

```
adamic: panic: field write failed: s expected string, found undefined
```

The fixture is registered as checked. Its explicit test pins each backend's diagnostic and pre-store stdout to writing, while holding Node to its separate successful expectation. Dropping only the optional check from production object.c and generated JavaScript produces Node's successful output in release, sanitized native and JavaScript; the required-stop expectation catches all three without a sanitizer failure. The optional witness remains held to Node, now including an undefined write through an ordinary spread copy. Its existing restore-rejection mutant is updated to the guarded condition. The construction metadata overlay sets the declaration false and is caught by the optional fixture's Node comparison in both backends, with expected string, found undefined. Mutant Go sources are stored as .go.txt.

Commands and evidence:

- ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run 'TestOptionalFieldCheckedView|TestOptionalFieldCopyState|TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_' -count=1 -v -timeout 90s: green, 25.851 s. New top-level leaves TestOptionalFieldCheckedViewRequiredString 0.60 s and TestOptionalFieldCheckedViewRequiredStringMutant 1.19 s. Existing string rejection mutant 1.17 s. destination-focus.log.
- ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/internal/oracle/testdata/optional_field_' -count=1 -timeout 90s -args -update-counts: green, 38.896 s. Unioned measured optional rows into the complete table. destination-counts.log.
- timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -v -timeout 90s: green, leaf 25.61 s. destination-readers.log.
- timeout 120 go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle: green. destination-vet.log.
- Construction metadata overlay: go test -overlay destination-metadata.json ./internal/oracle -run TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_checked_view_string_undefined.a$ -count=1 -v -timeout 90s fails as required, and the runner verifies the diagnostic and excludes build failures. destination-metadata.log and destination-metadata-result.log.
- Required lane checks: 4.3 s; gofmt and tools on 272 Go files, t.Parallel on 20 test packages, vet 20 packages. destination-lane.log. Repeated after commit before push.

Counts: the existing optional_field_checked_view_string_undefined.a row moves from 7 allocations, 7 frees, 10 retains, 15 releases, peak 5 to 10 allocations, 10 frees, 20 retains, 28 releases, peak 5 because the witness now creates and observes a spread copy and writes through it. Regions stay zero. No other measured existing row moved; no existing peak row moved. The new required-stop row is 2 allocations, 0 frees, 5 retains, 4 releases, peak 2, regions 0. Unmeasured rows are preserved. destination-counts-diff.md.

Setup: GOPROXY=https://proxy.golang.org|direct; Go 0.021 s, Node 0.022 s, markdown 0.070 s, submodules 0.092 s, clang 0.171 s, build 43.984 s, deferred tests 44.143 s, cache 44.144 s, total 44.175 s; nproc 5, CPU quota 4. destination-setup.log.

The attempted Reserved spread witness was rejected by the existing no-optional-widening rule; it was removed and that rule was preserved. Reserving-copy state regression fixtures pass, but the new optional-string fixture specifically exercises ordinary spread. No full packages, full gate, WASI or Darwin sweep was run. No PR was opened. This correction serves optional-presence-chain and step 10's checked-write misfit ruling.
