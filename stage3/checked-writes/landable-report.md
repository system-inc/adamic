Merged step 10 checked wider writes with current main for #0fgf1cq.
Base c9514f34; main 031a1259; the delivery SHA is reported with the push.
Focused suite, a-check, stage3 fixtures, stage1 gap probes and refreshed counts pass.
The focused suite reruns all 36 contract mutants; flags-misfit.ts catches removal of the write check.
Twenty-two views-dependent sites remain refused; this unit adds no views representation.

The merge resolved nine conflicts: internal/ir/ir.go,
internal/lower/class_inheritance.go, internal/lower/class_static.go,
internal/lower/invariance.go, internal/lower/object.go,
internal/lower/readiness.go, internal/native/emit_statements.go,
internal/oracle/counts.md and internal/oracle/counts_test.go.
Contracts, record storage, lazy non-null readiness and both counts registrations
remain present. Two runtime Object.entries shape initializers now include their
empty contract member.

The Buffer control exposed incorrect overload pairing. Every receiving overload
now needs a producing overload with compatible arguments and covariant observable
results. Each candidate has its own visited pairs. A void callback discards its
result. Regression tests preserve the unsafe second-receiving-overload refusal,
prove reordered producing overloads, and accept discarded callback results.
The existing Buffer and collection callback programs are held to Node.

Four stage3 statuses changed after measurement. The safe import-cycle read and
its import-order stop now compile and pass the native sanitizer hook against
source Node. The generic namespace and erased-prototype method remain NotYet,
with their actual first diagnostics. Source programs and Node observations did
not change.

The measured relation census remains 471 checked, 18 proven, 26 refused,
zero unmatched, out of 515. The twenty-two waiting rows retain their original
waits-on reasons: fourteen boxed-union fields and one Set element wait on
V2/V5, four tuple slots and one constrained generic array wait on V3, and two
recursive references wait on V5/V6. Four independent method-parameter or erased
any relations remain refused. These are relation decisions on the pinned adapted
source, not complete native compiler execution.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct and the printed
/workspace/adamic-tools/env.sh. Timing lines: Node 0.049s, Go 0.067s,
submodules 0.155s, Markdown dependencies 0.166s, clang 0.405s,
go build 63.303s, test binaries deferred 63.524s, cache warm 63.526s,
done 63.572s. nproc: 5; cpu.max: 400000 100000.

Commands use the sourced toolchain and send all test output directly to files.
The first runtime validation failed at the two incomplete shape initializers;
those runs were discarded. Counts also exposed the overload pairing and the
void callback rule before their repairs. No failed run is counted as passing.

- Focused lower/native/driver/oracle suite: /tmp/landable-focus-ready.log; PASS.
- a-check for the fifteen changed .a files: /tmp/landable-acheck-delivery.log.
  Eleven refused controls and four proven controls; no failures.
- go test ./stage3/fixtures -count=1 -timeout 15m:
  /tmp/landable-fixtures-delivery.log; PASS.
- go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 15m:
  /tmp/landable-stage1-delivery.log; PASS.
- go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1
  -timeout 15m -args -update-counts: /tmp/landable-counts-ready.log; PASS.
- Pinned census: /tmp/landable-census-delivery.log; PASS, 471/18/26.

No full repository gate was run. The stock tsc latent controls retain their
previous recorded evidence; this merge runs their reduced checked-write controls.

Exact commands: landable-commands.txt. Counts were regenerated, including the
109 checked-write and proven-control rows absent from main. Existing rows are
measured on this merged compiler rather than copied from either parent.
