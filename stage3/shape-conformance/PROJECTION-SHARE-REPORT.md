Built: record-property and constant-element allocation flow, with visible-store joins and conservative unknown frontiers.
Commits: merge checkpoint c1f4c5a70bd7d98fd543f4eb7ff96191be4a338b; implementation commit is in branch history.
Commands: full lower/IR tests, vet, 43 unchanged graph-region counts, native fixtures against Node, and independent census audits pass.
Mutants: four projection guards, six measurement guards, forged host/body proofs, and native wrong-shape/readiness erasures caught.
Limits: no tsc cast is proven free; callbacks, dynamic keys, classes, whole-array and transitive contracts remain unproven.

Every census number in this report is measured on a checker-rejected program.

| Outcome | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Conforms and ready (free) | 0 | 0 | 0 |
| Conforms but not proven ready | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 1 | 1 |
| Unknown: unsupported flow | 605 | 326 | 931 |
| Unknown: diagnosed body/dependency | 1153 | 851 | 2004 |
| Total | 1758 | 1178 | 2936 |

The merged checkpoint had 1668 unsupported sites and 1267 diagnosed sites. Tracking projections moves exactly 737 sites (591 tagged, 146 untagged) from unsupported to diagnosed dependencies: unsupported falls to 931 and diagnosed rises to 2004. Diagnostics stay at 315 and the adapted source hashes stay identical. This is newly visible provenance, not new checker failures. Allocation schemas increase from 589 to 1779. All 2936 ledger sites still map exactly. No field obligations are reported as conforms-if: unresolved whole-array intrinsic and augmented-field certificates remain unknown.

The largest unsupported pattern in the requested original 1148-site bucket was property/element producer flow: 536 sites (377 tagged, 159 untagged). That is the pattern implemented here. After the main merge, callback boundaries became more frequent; callbacks remain outside this change. The flow query now follows record properties, constant string keys and constant non-relative array elements, joining allocation initializers and every visible store. It retains unknown for opaque receivers, absent initial fields, spreads, untracked constructors, array mutations and opaque callable mutation. Known stores alone do not establish initialization on every path.

The diagnosed split is 44 own-function sites (9 tagged, 35 untagged) and 1960 dependency sites (1144 tagged, 816 untagged). Own-function counts did not move from the merge checkpoint; dependency counts grew by 737. Every diagnosed row carries raw checker diagnostics and scope, with zero missing provenance. Codes count once per site/scope and overlap:

| Diagnostic | Own function sites | Dependency sites |
|---|---:|---:|
| TS2322 | 7 | 1656 |
| TS2339 | 0 | 1641 |
| TS2345 | 32 | 1792 |
| TS2375 | 7 | 1637 |
| TS2379 | 8 | 1632 |
| TS2412 | 10 | 1749 |
| TS2488 | 7 | 1633 |
| TS2532 | 8 | 1740 |
| TS2538 | 0 | 1658 |
| TS2556 | 6 | 1629 |
| TS2722 | 6 | 1629 |
| TS2769 | 9 | 1635 |
| TS18046 | 0 | 1619 |
| TS18048 | 7 | 1736 |

These are adaptation obligations involving assignments, arguments, optionality/nullability, missing properties, iterable requirements and callable signatures. Fixing the underlying diagnostics can remove these boundaries, but does not establish the remaining field and readiness proofs. Context-insensitive joins spread diagnosed callers and opaque stores across shared values; tighter call-context or independently checked statement analysis would need a separate soundness argument. No rejected producer is trusted.

Host values remain unknown, even when another diagnosed dependency determines the primary bucket. Counts below count distinct cast sites per host value and overlap:

| Host value requiring library metadata | Sites |
|---|---:|
| `host.getPackageJsonInfoCache?.()?.getPackageJsonInfo` | 1 |
| `host.getSourceFile` | 1615 |
| `host.getSourceFileByPath` | 1615 |
| `host.readFile` | 617 |
| `host: ModuleResolutionHost` | 617 |

The host handoff includes target types and source locations in projection/latent-host-values.json; full frontier details are retained in projection/latent-share.json.gz. JSON.parse, host.getBuildInfo and readTextFile from earlier checkpoints remain outstanding metadata requests; the native readTextFile fixture still has the documented unsupported-representation gap. No new host execution agreement is claimed.

Named production functions: reachingAllocations adds projection-aware queries without changing graph.follow; projectionIndex indexes allocation fields and stores; projectedSources joins property/element producers. The graph owns lazy projection caches; graph-region ownership traversal and all 43 fixture counts remain unchanged. Existing hooks certifyAllocationFields, certifiedCheckedCast and eraseProvenViewChecks consume these queries. The adapter frontier uses a visited worklist and preserves raw diagnosed-store and host provenance.

New proven-property.a has zero checked casts and reads, pinned by an IR assertion and held to Node/native output. New nonconforming-property.a joins a conforming child and a number-valued boolean field: its view remains and fails with exit 70 naming ready, boolean and number. Erasing it produces a wrong value and the native mutant assertion catches it. Existing wrong-shape and readiness mutants also pass. Unit witnesses cover visible store joins, opaque aliases, recursive receivers, missing initialization and array mutation.

Four valid Go projection mutants drop stores, ignore opaque receivers, ignore missing initialization and ignore array mutation; semantic query assertions catch each. Six valid measurement mutants ignore whole-array contracts, exhaust the frontier budget, drop diagnostic provenance, ignore field type, ignore readiness and analyze diagnosed bodies; independent control assertions catch each. Forged host/body certifications are rejected by the corpus audit. Measurement controls contain 20 casts: five free, one readiness-only, two conforms-if and twelve unknown; these are separate from tsc.

Full lower/IR package tests pass in 35.483s/26.296s. Uncached native eraser fixtures and the 43 graph-count guards pass in 12.410s; vet passes. Logs are logs/projection-*.log and logs/latent-*.log. The full repository gate was not run. Initial recursive census attempts were stopped for excessive traversal or killed with exit 137; those attempts are not measurements. The final visited-worklist run completed and passed independent mapping/provenance audits.

Production erasure still conservatively disables mutable/callable programs and does not certify post-store readiness. Projection depth and frontier work limits return unknown rather than discarding inputs. Whole-array intrinsic properties, callbacks, dynamic keys, constructors and nested field contracts require additional certificates; the five array targets that reach literal allocations stay unknown rather than pretending array methods are absent.

Artifacts: projection/latent-share-summary.json, projection/site-comparison.jsonl.gz, projection-diagnostic-share.json and projection/latent-controls.json.gz. Reproduce with latent/make-overlay.py, go build -overlay, the analysis tool on the merged adapted root and site map, then latent/audit.py and latent/export-artifacts.py. The tool refuses production loading and output IR; checker-rejected code is never emitted.

Commands (outputs were written to the named log files):

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
go test ./internal/lower ./internal/ir -count=1 > stage3/shape-conformance/logs/projection-packages.log 2>&1
go test ./internal/oracle -run '^(TestCheckedViewShapeErasure|TestShapeGraphCountSnapshot)$' -count=1 -v > stage3/shape-conformance/logs/projection-final-oracle.log 2>&1
go vet ./internal/lower ./internal/ir ./internal/oracle ./stage3/shape-conformance/latent/tool > stage3/shape-conformance/logs/projection-vet.log 2>&1
python3 stage3/shape-conformance/projection-mutants.py > stage3/shape-conformance/logs/projection-mutants.log 2>&1
python3 stage3/shape-conformance/latent/make-overlay.py /tmp/shape-projection-final-overlay
go build -buildvcs=false -overlay=/tmp/shape-projection-final-overlay/overlay.json -o /tmp/shape-projection-final ./stage3/shape-conformance/latent/tool
/tmp/shape-projection-final /tmp/shape-conformance-merged-adapted /tmp/shape-merged-mapped.json /tmp/shape-projection-final-result.json > stage3/shape-conformance/logs/projection-final-census.log 2>&1
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-projection-final-result.json /tmp/shape-merged-mapped.json /tmp/shape-conformance-merged-adapted > stage3/shape-conformance/logs/projection-final-audit.log 2>&1
python3 stage3/shape-conformance/latent/mutants.py /tmp/shape-projection-latent-mutants > stage3/shape-conformance/logs/projection-latent-mutants.log 2>&1
```

Exact merged-to-projection transitions are recorded in projection/merged-transitions.json.
