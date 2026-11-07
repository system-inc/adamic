# Build-checked speculative field policies

This follow-up completes the private-field rule for every exported IR struct field. The canonical field-by-field actions and reasons are in statecopy/ir_fields.go and statecopy/POLICY.md. The overlay build checks names, declared types and scalar immutability before generating either the Function fast copier or the general IR mutable-field indexes. There is no runtime Function field allowlist.

The complete latent Go test tree passes. One guard requires an unnamed exported field to be refused as Function.AddedLater; the existing private-table guard requires Program.addedLater; the populated methodTables case still passes. The once-measured audit probe verified deep copying and preservation of aliases inside the cloned schema graph; it is historical behavioral evidence, not a current gate test.

The drop-Boundary mutant compiles the generator with that table entry removed, then generation refuses Function.Boundary at internal/ir/ir.go:200 before publishing a copier. Its command exits 1. Restoring the shallow private branch makes the methodTables probe fail, also exit 1. Both mutant sources and logs are in the raw archive, under names containing mutant.

Full stage3/apply.sh --build-product finishes with exit 0. Piece 01 runs audit_deferred.py --project /workspace/9k2-adapted/src/tsc/tsc.ts --mode speculative --baseline /workspace/scratch/9k-base-census --seconds 1200. The four child runs share that 20-minute deadline; they complete in about 82 seconds total. No partial adapted tree or pinned unadapted source is substituted. The src hash manifest matches before and after the audit.

Result: **0 state-copy panic findings and 0 unhandled state-copy panics**. There is one entry-file record and **6 units: 1 import declaration, 2 expression statements, 3 if statements, 0 function declarations**. Imported program context remains loaded; this result is for the requested entry root, not a claim of one independently attempted unit for every imported declaration.

Main's speculative run aborts on methodTables before publishing a file record. There are **0 shared speculative unit records**, so the empty difference list is not an equivalence proof. The separate first-error runs both complete all 6 units and their entire JSONL files are **byte-identical**. piece-01.json records exact timings, counts, baseline pin and binary hashes. The original synthetic fixture mode still has its previously recorded legacy import-route witness failure on main; the new real-project mode counts and audits the requested entry directly.

Raw observations, generated overlay, complete apply log, tests, mutants and adapted entry are packed in piece-01.tar.gz. piece-01.sha256 lists unpacked files.

## Gate proof without a build subprocess

The gate now starts with t.Parallel(), assembles the copier and generated policies, and type-checks them in process with go/types and go/importer's source importer against the real internal/ir source. It creates no temporary module, runs no go command, and sets no environment variables. The mutant path is read once at package initialization. For trimmed binaries, the source importer's source-root configuration is resolved from the installed Go tree on PATH once before any tests start; parallel tests only read it.

An AST ownership assertion requires a fresh zero struct, a private-policy/continue guard dominating each field assignment, a generated privateEmpty answer that only skips its field, and the shared-pointer guard before pointer copying. It checks every irPrivateFields policy, including methodTables. The historical old-branch-mutant.go.txt, selected with LATENT_STATECOPY_MUTANT as before, fails this AST assertion before type checking, so stale helper signatures cannot be the reason it turns red.

Measured behavior is retained from the earlier audit at 0375b8c12: its once-run probe populated Program.methodTables using MethodsNamed, observed the speculative copy's cache was nil, and observed the original cache stayed populated. The historical run is retained in piece-01.tar.gz (9k2-all-tests.log), with its exact probe source in that commit. The current gate does not execute that behavioral probe or launch go test. The adapted-entry audit above remains unchanged.

Validation: the whole statecopy package and the whole cmd/adamic-gate/parallelcensus package pass. All four introduced test names occur once repo-wide and call t.Parallel() first. A standalone trace of the final probe records exactly one execve, the test binary itself, and no subprocess. The historical mutant exits 1 at the AST assertion. gate-proof.json and the packed logs retain these checks. The standalone binary was compiled before tracing, outside the test, and was run from its package directory after sourcing the toolchain environment.

Compiler census: the gate's derived compiler-dependencies map still gives statecopy all seven compiler packages, including internal/ir and internal/lower, which cover its actual compiler source reads. This checkout's legacy cloud/fast-gate/compiler-dependencies.json has no literal statecopy entry; internal/compilercensus.Dependencies supplies the replacement map, as its command documents. No legacy map or unrelated serial baseline entry was edited.
