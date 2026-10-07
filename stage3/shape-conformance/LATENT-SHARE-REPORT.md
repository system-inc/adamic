Built: per-function allocation-flow share over every original ledger site, measured on a checker-rejected program.
Commits: production eraser 59db5d63; lane 1 609ed395 is already merged through 40ad020d; this measurement is recorded in branch history.
Commands: exact AST mapping, corpus/control measurement, independent diagnostic-span audit and production graph/eraser regression pass.
Mutants: ignoring field types, readiness or diagnosed bodies is caught by independent controls; forged host/body certifications are caught by the corpus audit.
Limits: proof coverage is zero certified sites, not proof that every runtime shape fails; unsupported flows and host values retain views.

Every census number in this report is **measured on a checker-rejected program**.
The corpus remains rejected with **830 diagnostics**, but the measurement runs
per function, excluding bodies with diagnostic byte-span overlap. Ordinary Load,
LoadOverlay and lower.Lower are disabled in the measurement binary. It imports
no backend and checks both output guards before analyzing source.

| Outcome, measured on a checker-rejected program | Tagged | Untagged | Total |
| --- | ---: | ---: | ---: |
| Conforms and ready (free) | 0 | 0 | 0 |
| Conforms, readiness not proven (readiness checks only) | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 1 | 1 |
| Unknown: flow the graph cannot see or certify | 727 | 421 | 1,148 |
| Unknown: diagnosed body/dependency | 1,031 | 756 | 1,787 |
| Total | 1,758 | 1,178 | 2,936 |

No corpus site is certified conforms-if, so there is no certified conditional
field list to report for tsc. This is a completed conservative proof-coverage
measurement, replacing the earlier unmeasured checkpoint. It is not a claim of
successful source lowering or that every unknown allocation is nonconforming.
Before and after certified free counts are both zero. The baseline audit had
only declared-origin hints; the new measurement actually builds and queries the
shared allocation graph on eligible source bodies.

Of the diagnosed totals, **346 tagged and 175 untagged** sites lie directly in
excluded function bodies. Another **685 tagged and 581 untagged** depend on diagnosed
callees, incoming arguments from skipped bodies or diagnosed module initializers.
Unknown reason buckets use diagnosed dependency first, then host metadata, then
unsupported flow, so they partition the ledger. Host provenance can overlap the
diagnosed bucket and is retained separately. Skipped call-site syntax can taint
incoming arguments, but its body is never analyzed for allocation certificates.
Signature-only diagnostics do not by themselves exclude a body.

The immutable 2,936 locations all map to adapted casts by exact AST text and
lexical owner, with equal occurrence counts; there are **zero ambiguous/unmapped
sites**. Offsets are byte offsets for the native checker, converted from the stock
AST's UTF-16 positions. The independent auditor checks raw diagnostic/body ranges,
site identity, every category total, host exclusion and non-vacuous allocation sets.
The analysis records **573 allocation schemas** with declared field types and
initializer readiness. These are analysis-IR site identities, assigned by the
same graphAllocationSites/assignAllocationSites machinery; they are not claimed
to be ids from successfully lowered executable tsc IR.

The source adapter creates identity-only IR producers for declarations, assignments,
conditionals, direct known function calls, parameters and returns. Dynamic method
and closure dispatch stay unknown. Its solving, producer
joins and allocation numbering are the existing compiler graph API, not a copied
flow engine. Actual record property syntax establishes presence; an interface
property list does not. Staged fields reuse uninitializedInitializer. Source
assertions cannot certify payloads. Missing property/element producer edges,
callbacks, default/opaque parameters, generic substitutions, constructors,
array/intrinsic certificates, callable bodies, mutation effects and transitive
object certificates remain explicit unknowns. No post-store readiness proof or
full executable IR is claimed for the source adapter. Its conservative initializer
subset is tested independently of the production eraser's CFG readiness hook.

`latent-flow-reasons.json` lists the overlapping exact frontier detail counts.
These reveal additional limits of the current graph/source adapter rather than
assigning a declared interface an invented runtime shape. Every retained unknown
continues to require lane 1's views.

Host values for the library handoff, measured on a checker-rejected program:

| Host value | Original cast | Target | Primary bucket |
| --- | --- | --- | --- |
| `host.getPackageJsonInfoCache?.()?.getPackageJsonInfo(...)` | moduleNameResolver.ts:931:25, `entry as PackageJsonInfo` | PackageJsonInfo | Host metadata |
| `host.getBuildInfo(...)` and `JSON.parse(...)` | builder.ts:1182:15 | IncrementalBuildInfo | Diagnosed dependency, with host provenance also recorded |

`latent-host-values.json` records the exact paths and source details. Package-json
cache results need runtime metadata for their structural fields and nested values;
build-info parsing requires validated/reified metadata for the returned structure.
Metadata enables checked reads; it does not make these host origins statically
certified allocations, and this lane leaves them unknown.

The additional native library fixture remains `readTextFile(...)` in
stage3/interface-downcasts/lane3/host.a. It is outside the 2,936-site denominator.
Its advertised Error fields are kind:string and message:string. Native execution
retains the view and stops with exit 70: `field read failed: kind is not a string;
expected string, found unsupported representation`. Source Node prints true.
The runtime-produced object's field type/readiness metadata needs a library fix.
No agreement or static erasure is claimed for that host fixture.

The small production hooks are newAllocationFlowGraph, allocationFlowGraph.follow,
allocationFlowGraph.ReachingAllocations, assignAllocationSites, certifyAllocationFields,
certifiedCheckedCast, eraseProvenViewChecks and rewriteShapeStatements. The new
measurement entry point is MeasureLatentShapes in a scratch-only overlay; it does
not alter production lowering. Hook details and the production scalar subset are
recorded in [FLOW-ERASER-REPORT.md](FLOW-ERASER-REPORT.md).

The control corpus is intentionally checker-rejected and written as .a source.
Its **12 cast controls** include two free proofs (factory/parameter/return flow
and a clean nested function in a diagnosed parent), one staged
readiness-only proof, one conforms-if proof and eight unknowns. The conditional
control names **ready**, declared **number**, expected **boolean**. Unknowns cover
host returns and asserted host fields, diagnosed bodies and incoming calls,
callback escape, forged scalar assertions, a wider mutable alias and a captured
value from a diagnosed parent. A nested function is excluded only for diagnostic
overlap with its own body; diagnoses of an outer body do not exclude a clean
nested body. Host detection uses checker receiver types, including named Host
interfaces, rather than treating every AST variable named host as a boundary.
Source measurement mutants compile valid measurement binaries: ignoring field
assignability incorrectly certifies the numeric payload free; ignoring readiness
incorrectly certifies the staged slot free; ignoring diagnosed-body exclusion
analyzes a skipped incoming call. The independent control assertions catch all
three. Artifact mutants adjust headline counts before forging a host or diagnosed
certification; the auditor catches the host provenance or raw diagnostic-body
boundary, not merely a stale count total. Existing native/JavaScript eraser
wrong-shape and ignored-readiness runtime mutants remain covered by the production
regression command.

The adapted tree was prepared with stage3/apply.sh at /tmp/shape-conformance-adapted,
using the pinned upstream source and existing stage-3 adaptations. The earlier
REPORT.md records its command. The per-file source hashes in the summary pin this
measurement independently of checker acceptance.

Reproduction, from the repository root after sourcing the toolchain:

```sh
node stage3/shape-conformance/latent/map-sites.cjs /tmp/shape-conformance-adapted /tmp/shape-latent-mapped.json
python3 stage3/shape-conformance/latent/make-overlay.py /tmp/shape-latent-overlay-audited
gofmt -w /tmp/shape-latent-overlay-audited/*.go
go build -buildvcs=false -overlay=/tmp/shape-latent-overlay-audited/overlay.json -o /tmp/shape-latent-audited ./stage3/shape-conformance/latent/tool
/tmp/shape-latent-audited /tmp/shape-conformance-adapted /tmp/shape-latent-mapped.json /tmp/shape-latent-result.json
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-latent-result.json /tmp/shape-latent-mapped.json /tmp/shape-conformance-adapted
```

All test/measurement output was captured to logs. `logs/latent-map.log`,
`latent-build.log`, `latent-run.log`, `latent-audit.log`, `latent-fixture-audit.log`,
`latent-mutants.log`, and `latent-artifact-mutants.log` preserve observations.
The independent .a controls and helper mapper are under latent/. Prepare controls:

```sh
mkdir -p /tmp/shape-latent-fixtures/src/compiler
cp stage3/shape-conformance/latent/fixtures/main.a /tmp/shape-latent-fixtures/src/compiler/main.a
node stage3/shape-conformance/latent/fixture-sites.cjs /tmp/shape-latent-fixtures/src/compiler/main.a /tmp/shape-latent-fixture-sites.json
/tmp/shape-latent-audited /tmp/shape-latent-fixtures /tmp/shape-latent-fixture-sites.json /tmp/shape-latent-fixture-result.json
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-latent-fixture-result.json /tmp/shape-latent-fixture-sites.json /tmp/shape-latent-fixtures --fixtures
python3 stage3/shape-conformance/latent/mutants.py /tmp/shape-latent-mutants
python3 stage3/shape-conformance/latent/artifact-mutants.py /tmp/shape-latent-result.json /tmp/shape-latent-mapped.json /tmp/shape-conformance-adapted
python3 stage3/shape-conformance/latent/export-artifacts.py /tmp/shape-latent-result.json /tmp/shape-latent-mapped.json /tmp/shape-latent-fixture-result.json /tmp/shape-conformance-adapted
```

Capture command output in log files, as the recorded runs did. Full raw measurements, body maps and
control results are latent-share.json.gz, latent-site-map.json.gz and
latent-controls.json.gz. Summary source hashes pin every adapted input.
The updated site-comparison.jsonl.gz retains baseline reasons alongside every
measured after row. No original site categories are reclassified.

The ordinary production regression runs uncached
`go test ./internal/oracle -run '^TestShapeGraphCountSnapshot$|^TestCheckedViewShapeErasure$' -count=1 -v`:
all 43 graph-region fixture count rows remain pinned to the pre-refactor merge,
and the factory/nonconforming/staged executions and erasure mutants pass.
`go vet ./internal/lower ./internal/ir ./stage3/shape-conformance/latent/tool` passes.
Production compiler source was not changed for the measurement. The full repository
gate was not rerun, and host runtime agreement is still excluded for the documented
metadata gap.

Fetched origin/codex/interface-downcasts is 609ed395; `git merge` reports already
up to date because 40ad020d merged it. The production eraser is already wired to
lane 1's contracts through certifiedCheckedCast and to shared readiness through
eraseProvenViewChecks. The measurement ran on the 71d7e491-based production tree. A final refresh found
origin/main had advanced to 48c05d09; integration of that larger compiler/cohere
change follows this pinned measurement commit.
Only codex/shape-conformance is pushed; no PR is opened.

Initial toolchain setup completed in 58.329s; nproc=5, CPU quota=4. The
individual timing lines and the initial failed dependency-conflict attempt are
retained in REPORT.md. This measurement used /workspace/adamic-tools/env.sh and
GOPROXY=https://proxy.golang.org|direct.
