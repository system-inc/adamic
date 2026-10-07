Built: merge main 48c05d09 with checked views, readiness and lane 3 hooks retained; fresh adapted-tree census.
Commits: previous census ffe5a90e; this merge is recorded in branch history.
Commands: full lower/IR packages, scoped uncached oracle, 43 graph-count guards, cast oracle, vet and independent census audits pass.
Mutants: native wrong-shape/readiness plus measurement type/readiness/body/provenance and host/body forgery mutants caught.
Limits: all tsc sites still need views; complete erasure and host runtime metadata remain unproven.

Every census number below is measured on a checker-rejected program.

| Outcome | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Free | 0 | 0 | 0 |
| Readiness checks only | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 1 | 1 |
| Unknown: unsupported flow | 1196 | 472 | 1668 |
| Unknown: diagnosed body/dependency | 562 | 705 | 1267 |
| Total | 1758 | 1178 | 2936 |

Compared with ffe5a90e, diagnostics drop from 830 to 315. Diagnosed sites shrink
by 520 (469 tagged and 51 untagged); all 520 move to unsupported flow. Free,
readiness-only, conforms-if and primary host counts do not move. All original
sites map exactly, with zero ambiguous/unmapped sites. Allocation schemas grow
from 573 to 589. The fresh adapted source is pinned by per-file hashes in
merged/latent-share-summary.json; no checker option was changed for measurement.
Main itself removed implicit-return/fallthrough rejection and erasable-only syntax
as it added those lowering mechanisms. The fresh stage-3 adaptations also differ.
These observations do not isolate the contributions of options versus adaptations.

44 sites are in directly diagnosed functions (9 tagged, 35 untagged), down from
521. 1223 depend on diagnosed producers (553 tagged, 670 untagged), down from
1266. Each row now retains raw diagnostic strings and whether its obligation
comes from its own function or a dependency. Identical diagnostic obligations
are deduplicated with a representative origin; graph source details retain the
other incoming frontier locations. Raw diagnostic membership is independently
validated, and zero diagnosed sites lack provenance.

Diagnostic-code rows overlap: each code is counted once per site/scope, even if
several dependencies carry it. The full messages and per-site causes are in
merged-diagnostic-share.json and merged/latent-share.json.gz.

| Diagnostic | Own function sites | Dependency sites |
|---|---:|---:|
| TS2322 | 7 | 1003 |
| TS2339 | 0 | 415 |
| TS2345 | 32 | 1143 |
| TS2375 | 7 | 992 |
| TS2379 | 8 | 986 |
| TS2412 | 10 | 981 |
| TS2488 | 7 | 985 |
| TS2532 | 8 | 1084 |
| TS2538 | 0 | 432 |
| TS2556 | 6 | 980 |
| TS2722 | 6 | 980 |
| TS2769 | 9 | 986 |
| TS18046 | 0 | 393 |
| TS18048 | 7 | 1118 |

Remaining obligations include incompatible assignments/arguments, missing
properties, exact optional fields, nullable values/calls, spread signatures and
iterator requirements. These are real checker obligations that adaptations can
remove. The large overlapping dependency counts also expose whole-program,
context-insensitive joins: one skipped caller can taint a shared parameter used
by many casts. Removing a diagnosis need not produce a free cast; this run
instead exposed 520 additional unsupported flows. A different approach to slice
only independent statements or specialize incoming call contexts would need its
own soundness proof; no diagnosed producer is assumed safe here.

Host provenance remains host.getPackageJsonInfoCache?.()?.getPackageJsonInfo,
host.getBuildInfo and JSON.parse. Build-info provenance overlaps the diagnosed
bucket; every such cast stays unknown. Library metadata requests and the extra
native readTextFile fixture are unchanged from LATENT-SHARE-REPORT.md.

Conflict resolutions keep main's enum/class/tag proof and import-cycle reads,
adding checkedViewTarget for structural view dispatch, retaining source readiness
labels and captured-cell checks, and preserving definite initialization support.
Both documentation checkpoints are kept. No cohere implementation was copied:
cohere and its nested TypeScript checkout use main's exact pinned revisions.
The production eraser hooks remain certifyAllocationFields, certifiedCheckedCast
and eraseProvenViewChecks. The merge-only dispatch helper is checkedViewTarget.

Validation logs are logs/merge2-*.log. Full lower/IR tests pass in 58.139s/52.554s.
The 43-count snapshot and native eraser/readiness mutants pass uncached in 13.666s;
main's cast fixtures pass against Node in 3.942s. Targeted package tests and vet
pass. All measurement source mutants build valid binaries and fail independent
semantic/provenance assertions, including dropping diagnostic provenance.
The full repository gate was not rerun.

Reproduce with the latent overlay and audit commands in LATENT-SHARE-REPORT.md,
using /tmp/shape-conformance-merged-adapted prepared with stage3/apply.sh and its
mapped ledger. Export with the optional final destination argument merged/;
diagnostic-share.py consumes the raw measurement and writes the code breakdown.
Both production Load and Lower remain disabled in the measurement binary.
