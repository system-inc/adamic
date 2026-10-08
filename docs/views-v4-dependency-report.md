Built: a lane5 net-change dependency probe on b8dd5b62; no production V4 admission is certified.
Commits: no compiler/views-v4 source commit or Train-slice marker; this report is on the separate plan branch.
Commands/results: three go build ./... probes and go vet ./internal/... exit 1 on missing prerequisite symbols.
Mutants: none run; an uncompilable compiler is not mutation evidence.
Uncovered: all V4 runtime certification, Set-domain pinned negatives, full comparisons and train delivery remain pending.

# Rehearsal dependency observations

The exact application base is b8dd5b62e, above views-v3 787cea7ac. The nine
prepared net patches remain in /tmp/views-v4-net. Only lane5's net change was
attempted. The first application was atomic and failed because the base lacks
callable scaffold files already present at the lane's fork. Adding the callable
scaffold only, and excluding absent native array helper files, exposes 22 file
conflicts. Each was saved under /tmp/views-v4-conflicts. The probe retained base
conflict sides; these are diagnostic choices, not final approved resolutions.

First build stops at missing ir.ArrayViewRead. Removing that incidental array
metadata hunk exposes graphArray/viewArrayReadOwner/heldIn in native, source-slot
and nominal Map certificates in JavaScript, and dictionary/Record helpers in
lowering. These inherited changes cannot be imported wholesale under V4's scope.

A second isolation removes every modified pre-existing shared file, retaining
only new callable files and witnesses. That build still fails. Besides the
expected own callable metadata hooks, its concrete missing dependencies include:

- lower/view_callables_aggregate.go: optionalViewWriteField and ir.ClosureOperands.
- lower/view_callables_boxing.go: viewArrayElementType for array payloads.
- lower/view_callables_read.go: prepareUntaggedCallableUnionRead.
- native/view_callables_signature.go: untaggedCallableUnionExpected and
  certifyUntaggedCallableRecorded.
- javascript/view_callables_signature.go: untaggedCallableUnionExpected and
  untaggedCallableRecorded.
- native/view_callables_discard.go: ir.Record in producer result disposal.

The own metadata errors CallableMasks, DiscardContract/DiscardView and payload
maps are not claimed as external blockers; V4 can add those minimal hooks.
The other named dependencies require a callable-only cut from their owning
implementations, keeping all unrelated source admissions refused. Missing names
alone do not prove permanent inseparability. No check is deleted or replaced
with fabricated producer metadata to make this build succeed.

Source linkage: cb6adb25e introduces boxed adaptation; 3e1a7f6f and ebb1f8a61
add aggregate propagation; 164dac2c adds array callable payloads. f339ca8d adds
complete higher-order producer checks on the existing untaggedCallableTargets
registry and certifyUntaggedCallableProducers. The latter lives in
lower/view_unions_untagged.go, introduced through 55b3fdef0/566aad67e; array type
helpers originate in 3b1263b65. These are dependencies of the supplied net changes,
not authorization to take entire V2/V3 commits. 2efc8949 remains dependent on
f339ca8d and the Set receiver certificate path; neither repair is certified here.

Every log is saved in views-v4-evidence/rehearsal-*.txt. The incomplete source
probe remains local on compiler/views-v4 and has no source commit. Its initial
net-change patch is saved at /tmp/views-v4-unresolved-net.patch. No rehearsal or
compiler/views-v4 push is made, because the required compiler build and vet
have not passed. Only this dependency evidence is pushed to codex/views-v4-plan.
No Node comparison, a-check, counts, stage1 gap move, platform-guard edit, full
oracle or new language ruling is claimed from these compile probes.
