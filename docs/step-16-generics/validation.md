# Step 16 measurement validation

The census and audit are measurement-only; neither emits a runnable compiler program.

`python3 docs/step-16-generics/audit.py --test > /tmp/scout-audit.log 2>&1`: four tests passed. Distinct roots are counted by attempted unit position, independent of duplicate findings and diagnostic sites. Collision classification requires an enclosing binder with the diagnostic spelling.

`python3 docs/step-16-generics/audit.py --mutant > /tmp/scout-audit-mutant.log 2>&1`: exited 1. Counting observations instead of distinct roots produced 4 where the independent assertion requires 2.

`python3 docs/step-16-generics/audit.py --mutant-selection > /tmp/scout-selection-mutant.log 2>&1`: exited 1. Selecting every reason incorrectly included concrete `Type`, lexical `reading T`, and concrete `SourceFile` diagnostics; the negative selection assertion caught it.

`python3 docs/step-16-generics/parameters_test.py "$HOME/.cache/adamic-stage3/api/node_modules/typescript" > /tmp/scout-parameters-test.log 2>&1`: passed the independent same-name concrete/generic AST probe. Adding `--mutant` disables AST parent links; it exited 1 because both generic cast uses disappeared, caught by the expected-two assertion.

`python3 docs/step-16-generics/audit.py --mutant-markdown > /tmp/scout-markdown-mutant.log 2>&1`: exited 1. Splitting an escaped union pipe as a table separator produced 11 columns instead of 10, caught by the independent table-row assertion.

Setup initially failed while a checkout overlapped its cache warm build: undefined `usesNodeModules`, `nodeTypesIndex`, `starCollision`, and `nodePrelude`. Repeating setup on the stable requested base succeeded. Setup timing lines: Go 0.066 s; Node 0.063 s; submodules 0.148 s; markdown skipped after lock validation, step 0.011 s, ready 0.211 s; clang 0.495 s; go build ready 103.641 s; test binaries deferred 103.819 s; cache warm 103.824 s; done 103.877 s. `nproc` printed 5, CPU quota 4, memory limit 17.6 GB. Toolchain: Go 1.27.1, Node 24.19.0, clang 20.1.8.

No whole-package test or full gate was run for this documentation unit.
