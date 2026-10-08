# Step 16 measurement validation

The census and audit are measurement-only; neither emits a runnable compiler program.

`python3 docs/step-16-generics/audit.py --test > /tmp/scout-audit.log 2>&1`: four tests passed. Distinct roots are counted by attempted unit position, independent of duplicate findings and diagnostic sites. Collision classification requires an enclosing binder with the diagnostic spelling.

`python3 docs/step-16-generics/audit.py --mutant > /tmp/scout-audit-mutant.log 2>&1`: exited 1. Counting observations instead of distinct roots produced 4 where the independent assertion requires 2.

`python3 docs/step-16-generics/audit.py --mutant-selection > /tmp/scout-selection-mutant.log 2>&1`: exited 1. Selecting every reason incorrectly included concrete `Type`, lexical `reading T`, and concrete `SourceFile` diagnostics; the negative selection assertion caught it.

`python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" > /tmp/scout-parameters-test.log 2>&1`: passed the independent same-name concrete/generic AST probe. Adding `--mutant` disables AST parent links; it exited 1 because both generic cast uses disappeared, caught by the expected-two assertion.

`python3 docs/step-16-generics/audit.py --mutant-markdown > /tmp/scout-markdown-mutant.log 2>&1`: exited 1. Splitting an escaped union pipe as a table separator produced 11 columns instead of 10, caught by the independent table-row assertion.

Setup initially failed while a checkout overlapped its cache warm build: undefined `usesNodeModules`, `nodeTypesIndex`, `starCollision`, and `nodePrelude`. Repeating setup on the stable requested base succeeded. Setup timing lines: Go 0.066 s; Node 0.063 s; submodules 0.148 s; markdown skipped after lock validation, step 0.011 s, ready 0.211 s; clang 0.495 s; go build ready 103.641 s; test binaries deferred 103.819 s; cache warm 103.824 s; done 103.877 s. `nproc` printed 5, CPU quota 4, memory limit 17.6 GB. Toolchain: Go 1.27.1, Node 24.19.0, clang 20.1.8.

No whole-package test or full gate was run for this documentation unit.

## Fixture unit

`go test ./internal/oracle -run 'TestStep16GenericOutcomes|TestStep16IdentityMutant|TestNativeAgreesWithNode/stage3/fixtures/generics/' -count=1 -v > /tmp/scout-fixtures-test.log 2>&1`: PASS, 21.287 s. All eight exact baseline outcomes passed; the two supported fixtures matched source Node through both backends and release C, with sanitizer and leak checks. The valid numeric identity result mutant was caught by stdout in both backends, with empty sanitizer/leak reports.

`go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts > /tmp/scout-fixtures-counts.log 2>&1` initially failed after 81.977 s: host fixtures could not find pinned @types/node 25.3.3 under stage3/api/node_modules. `npm ci --prefix stage3/api --ignore-scripts > /tmp/scout-node-types.log 2>&1` installed the three lockfile-pinned packages without changing the lockfile. The same counts command redirected to `/tmp/scout-fixtures-counts-retry.log` then passed in 64.068 s. The generated table adds the two supported fixture rows, removes the stale taste/17_binder_flow row already registered unsupported on this base, and moves an existing logical-and row to registry order without changing its numbers.

Independent source Node runs for all eight fixtures exited 0 with empty stderr. Their exact stdout and source hashes are retained in fixture-baseline.json. `python3 docs/step-16-generics/audit.py --test > /tmp/scout-fixture-audit.log 2>&1` also passed. No full package or full gate was run.
