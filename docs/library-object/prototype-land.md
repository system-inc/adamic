# Object prototype landing

New branch: codex/library-object-prototype-land. Original codex/library-object-prototype remains at 55302e5a87cf189f0f77df5031684155cc355572. Rebased onto current origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c; fetched again before publication and main was unchanged. No force push or old-branch edits.

The only textual conflict was internal/oracle/counts.md. Counts were regenerated from the combined fixture registrations, retaining main's rows and adding the prototype fixtures in generated order. The first attempt failed because an integration dispatch conflict surfaced in library_object_keys.a. Main's class reflection intercepted primitive Object.keys receivers and refused them. A three-line hook in internal/lower/class_features.go delegates proven number, boolean, and string receivers to the library Object path; main's class and accessor reflection path remains intact. The existing keys fixture catches this regression, and class fixtures were included in the rerun.

Commands below ran on Linux after sourcing /workspace/adamic-tools/env.sh; all test output went directly to log files.

| Command | Final result |
|---|---|
| go test -count=1 -timeout 30m ./internal/lower ./internal/native ./internal/javascript | lower 30.847s; native 106.949s; JS no package tests |
| ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run 'TestObject\|TestNativeAgreesWithNode/internal/oracle/testdata/(library_object_\|object_prototype\|has_own\|class_)' | pass 17.730s |
| go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts | pass 24.421s |
| go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded | pass 17.772s |
| go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle | exit 0, no output |
| gofmt -l on the changed Go files; git diff --check | clean |

[Raw verification receipts](prototype-land-evidence/). Uncached oracle checks compare Node with sanitized native, release -O2, and JavaScript and check successful programs for leaks. The full repository gate was not rerun; this landing re-greened the affected packages, Object fixtures, class fixtures, and generated counts. Original prototype mutant evidence remains in prototype-review.md.

The static Object ancestor commits remain on this rebased branch because seat 1's older static landing is not yet on main. Land that dependency first and reconcile shared commit ancestry at integration. No work was done on library-object-3: it had already been pushed as 3aae593 before the cap arrived, with affected gates green but a documented inherited full-flow discovery failure. It has not been pushed again. New library work is stopped pending the backlog falling below the requested cap.
