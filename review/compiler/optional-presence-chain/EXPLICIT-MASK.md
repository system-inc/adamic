Built: made adamic_object_view_store's metadata mask explicit with & 255 instead of relying on unsigned char conversion.
Commit: one follow-up on compiler/optional-presence-chain after c014d545.
Validation: optional-presence fixtures, both backends, native sanitizers, recorded counts and lane checks pass.
Mutants: runtime, generated-JavaScript and original source mutants were rerun and caught; construction optionality mutant was caught too.
Remaining: full gate was not run; runtime cleared c014d545 and requested no reclearance for this follow-up.

Production change: internal/native/runtime/object.c:430. No behavior, object layout or count row changed. No new tests or fixtures were added.

Commands:

- `ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run 'TestOptionalField|TestNativeAgreesWithNode/internal/oracle/testdata/optional_field_' -count=1 -v -timeout 90s`: green, 6.749 s. explicit-mask-focus.log. Includes Node comparisons, release and sanitized native, JavaScript, presence/deletion/reinsertion, enumeration, reserving-copy state, checked writes and required-string exit-70 expectations.
- `ADAMIC_GATE_UNCACHED=1 timeout 360 python3 review/compiler/optional-presence-chain/explicit-mask-run-mutants.py`: all source mutants caught. explicit-mask-source-mutants.log and individual logs in explicit-mask-source-mutants/.
- `ADAMIC_GATE_UNCACHED=1 timeout 130 python3 review/compiler/optional-presence-chain/destination-mutant.py`: construction optionality mutant caught by Node in both backends. explicit-mask-destination-mutant.log and explicit-mask-metadata.log.
- `timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$/^fixtures$/internal/oracle/testdata/optional_field_' -count=1 -timeout 90s -args -update-counts`: green, 20.466 s. All 15 measured optional rows match the baseline. Restored the complete, unchanged counts.md. explicit-mask-counts.log and explicit-mask-counts-check.txt.
- Required integration lane checks: 4.5 s; gofmt and tools on 272 Go files, t.Parallel on 20 test packages, vet 20 packages. explicit-mask-lane.log. Repeated after commit before pushing.

The initial combined run also selected TestCountsAreRecorded without update-counts. Its fixture leaves passed, but the driver rejected comparing its incomplete selected table with the complete baseline. The focused fixture run and explicit count measurement above resolve that invocation issue. The initial log is preserved in explicit-mask-count-selection.log.

Mutant catches: static key list, missing reserved storage, wrong presence/order/deletion, copy presence/readiness/representation state, wrong dynamic descriptor types, dropped checked store, missing presence publication, missing write guard, absent number read as zero, optional boolean decoding, union boxing, omitted finite literal guard, borrowed string self-write without retain (ASan heap-use-after-free), optional string rejection, required string optional-check omission, boolean reservation omission, checked/plain origin bypass, unproven hasOwn shape, binder fixture registration, literal undefined field representation, and construction optionality omission. Runtime semantic mutants were caught by Node or pinned backend expectations; the ownership mutants were caught by ASan. The complete output and exact mutation sites are in the tests and the source runner.

Setup used GOPROXY=https://proxy.golang.org|direct. Build 50.753 s, deferred tests 50.996 s, cache 50.997 s, total 51.025 s; nproc 5, CPU quota 4. Other setup timing lines are recorded in explicit-mask-setup.log. Every test command had a hard timeout. No full packages or full gate were run, and no PR was opened.
