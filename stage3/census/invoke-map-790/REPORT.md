# InvokeMap annotation evidence (#4pc8wd8)

Final implementation: a1f279e55e39e06d450da64d413d21f508017396 on base 8e3a784aa5a601f57342c07b68f145fe56252c21.
TypeScript source 050880ce59e30b356b686bd3144efe24f875ebc8.

Full single-file sys.ts CommonJSBundle census: 0 new, 0 changed, 1 gone.
The sole gone observation is the 790:87 invariant-mutable Refused in phase
refusal_scan; census-delta.json and census-delta.txt list its full content.
The existing adapted Array.isArray guard at 715 has no finding on either side;
upstream's 717 cast was already replaced before this amendment. No reader changes.

Stock 6.0.3: 79 roots, 0 diagnostics on each side. Adamic types: both exit 0.
All 79 emitted JavaScript files are byte-identical (archives and hash manifests).
Upstream hereby dts succeeds on each side; typescript.d.ts and
tsserverlibrary.d.ts are byte-identical. Existing witnesses and whole-owner
contract outputs are byte-identical. comparisons.json records these assertions.

Restoring mutant passes: its entire full census equals main byte-for-byte,
including the 790:87 refusal_scan observation. The amendment is restored.
Full unfiltered lane PASS: 106,366 passing, one sanctioned Public API row,
zero pending. The sole failure is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`;
baseline_diffs is exactly `["api/typescript.d.ts"]`. Apply, install and build
exit 0; tests and oracle exit 1 as required; lane exits 0. Eight workers, all
runners, no filter, no timeout or exception. `lane/` retains every receipt
and raw phase log (the adapted tree is omitted). All 79 lane compiler roots
match the exact amendment bytes used by the other proofs.

The supplemental legacy census audit passes. The supplemental full_audit.py
fixture fails on unchanged main because it expects string/number condition
refusals while actual output contains only var; its raw runs and failure are retained.
Main apply completed all adaptations but its cache tar failed when concurrent
proof builds changed build files. This failed invocation is retained, not claimed
as a passing apply. Amendment lane apply runs isolated until process exit.

No Loom submission.
