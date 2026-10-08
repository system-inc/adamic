Meter fix 49bdfe0e replaces the refusal text substitutions with checked Go AST instrumentation.
Evidence run 20261008T012240Z.UPFX2e completed with exit 0 in default single-compiler mode.
Compiler 719bb8d9 is acdf5c10 (area 3b255125 merged with enum compiler 41231d51), followed only by the meter fix.
Four AST tests and all 23 meter tests pass; missing-function, changed-walker, planted-NotYet, body-range, misattribution, IR-output and loader-output mutants are caught.
This is a measurement on a checker-rejected program; no native/oracle gate was run. The landing run was canceled as requested.

Whole program: main: 2/79; area: 2/79
Own file: main: 56/79; area: 57/79

Source refs: main f4efdd2369311d1420aa53fdf5c1a55bdda811d4; area 3b25512566206bc603b93264e8d072c55e075d64.

Compared with 20261007T230852Z.jzKzWs:

| Tree | Checked enum object views | NotYet | Refused |
| --- | --- | --- | --- |
| main | 90 -> 0 | 1560 -> 1085 (-475) | 3735 -> 2682 (-1053) |
| area | 90 -> 0 | 1558 -> 1084 (-474) | 3894 -> 2746 (-1148) |

No file moved from pass to fail for whole-program checking, own-file checking, or lowering. Own-file fail-to-pass: performanceCore.ts on both trees; tracing.ts on area only.

The old checked-view group has zero events and no renamed checked-view group appears. A separate Refused reason remains: an unproven singleton numeric enum member field in an object view, 8 main / 6 area (OWNER BLANK). Zero checked-view events do not mean all enum object views compile.

SkippedDependency: 3 -> 4 on both trees. Error: 0 on both trees. Panic: 2 -> 0 main; 2 -> 1 area. The area panic is moduleNameResolver.ts:977:1 during lowering: runtime error: invalid memory address or nil pointer dereference. See latent-errors.json.

Counts changed with both compiler/source versions and repaired instrumentation; they are observations, not changes attributable solely to the enum fix. Diagnosed bodies remain skipped; findings are not exhaustive blockers or compiled programs.

The first attempt failed because the newer compiler requires the exact @types/node 25.3.3 dependency. npm ci --prefix stage3/api installed the existing lockfile without tracked changes. Its produced artifacts are preserved under prior-attempt/full-run.

Toolchain: Node 24.19.0; Go 1.27.1; clang 20.1.8; nproc=5. Reused enum scratch setup completed in 105.372s (Go 0.045s, Node 0.055s, markdown 0.153s, submodules 0.168s, clang 0.392s, Go build 105.009s, cache warm 105.275s). The fix checkout setup reached Go 0.039s, Node 0.040s, markdown 0.142s, clang 0.451s, submodules 337.537s, then exited 1 at go list with empty stderr; no specific cause is claimed. Logs retain both setups.

Production compiler sources are identical to the enum merge. The actual census binary build records revision 719bb8d9e6bebed2a92543bd86c49243b20cef37 and vcs.modified=true from untracked meter run directories; tracked sources were clean. The latent binary uses buildvcs=false, with selected compiler provenance retained separately.

The full owner-ranked top-10 tables and unowned ledger are in report.md/report.json. Raw census and latent JSONL remain local and are excluded from the evidence commit.
