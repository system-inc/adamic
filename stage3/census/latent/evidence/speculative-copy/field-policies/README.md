# Build-checked speculative field policies

This follow-up completes the private-field rule for every exported IR struct field. The canonical field-by-field actions and reasons are in statecopy/ir_fields.go and statecopy/POLICY.md. The overlay build checks names, declared types and scalar immutability before generating either the Function fast copier or the general IR mutable-field indexes. There is no runtime Function field allowlist.

The complete latent Go test tree passes. One guard requires an unnamed exported field to be refused as Function.AddedLater; the existing private-table guard requires Program.addedLater; the populated methodTables case still passes. The Boundary ownership probe verifies deep copying and preservation of aliases inside the cloned schema graph.

The drop-Boundary mutant compiles the generator with that table entry removed, then generation refuses Function.Boundary at internal/ir/ir.go:200 before publishing a copier. Its command exits 1. Restoring the shallow private branch makes the methodTables probe fail, also exit 1. Both mutant sources and logs are in the raw archive, under names containing mutant.

Full stage3/apply.sh --build-product finishes with exit 0. Piece 01 runs audit_deferred.py --project /workspace/9k2-adapted/src/tsc/tsc.ts --mode speculative --baseline /workspace/scratch/9k-base-census --seconds 1200. The four child runs share that 20-minute deadline; they complete in about 82 seconds total. No partial adapted tree or pinned unadapted source is substituted. The src hash manifest matches before and after the audit.

Result: **0 state-copy panic findings and 0 unhandled state-copy panics**. There is one entry-file record and **6 units: 1 import declaration, 2 expression statements, 3 if statements, 0 function declarations**. Imported program context remains loaded; this result is for the requested entry root, not a claim of one independently attempted unit for every imported declaration.

Main's speculative run aborts on methodTables before publishing a file record. There are **0 shared speculative unit records**, so the empty difference list is not an equivalence proof. The separate first-error runs both complete all 6 units and their entire JSONL files are **byte-identical**. piece-01.json records exact timings, counts, baseline pin and binary hashes. The original synthetic fixture mode still has its previously recorded legacy import-route witness failure on main; the new real-project mode counts and audits the requested entry directly.

Raw observations, generated overlay, complete apply log, tests, mutants and adapted entry are packed in piece-01.tar.gz. piece-01.sha256 lists unpacked files.
