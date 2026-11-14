# Combined recursive watcher amendment (#4pc8wd8, #frhzgrc)

Implementation: e53da7b28b7191e0dee15d87b67f80d859b83eef, one commit on
8e3a784aa5a601f57342c07b68f145fe56252c21. Both edits share one owner after-pin.

The only form tried for #frhzgrc was
`newChildWatches = [] as ChildDirectoryWatcher[]`. Stock TypeScript 6.0.3
accepts it, and the full sys.ts CommonJSBundle census accepts this assertion
on a fresh empty literal. No alternative or fallback form was needed.

Fresh main against the exact combined source: 0 new, 0 changed, 2 gone, both
Refused observations in refusal_scan, at 790:87 and 877:34. census-delta.txt
and census-delta.json list every full finding with phase and attempting unit.
The fresh main raw stream equals the earlier main stream byte-for-byte.
The combined-census stream is the accepted form trial, on exactly the final
source bytes. Source paths, binary, guards and module format match the prior
proof; the comparator maps only the changed type syntax byte offsets.

Combined stock check: 79 roots, 0 diagnostics. adamic types: 79 roots, exit 0.
Every one of the 79 emitted JavaScript files is byte-identical to main.
Both actual upstream hereby dts files, the unchanged witness, and the whole-owner
contract outputs are byte-identical to main. comparisons.json and retained
outputs record these checks. Baseline stock/declaration/witness/contract
artifacts from the identical main and toolchain remain in this directory's
parent; those established baseline bytes are reused for the combined comparison.

Both independent restoring mutants pass: each brings back exactly its own
refusal_scan site, with no other new, gone or changed finding. The combined
source is restored and only sys.ts differs in the source hashes.
The first fresh whole lane passed apply/install/build but stopped when a worker
exited without a test failure or complete counts. Cgroup OOM kills and its
16 GiB memory limit suggest memory pressure; this attempt is not a pass.
All failed receipts are retained under failed-lane/. Removing only task-owned
unused Go cache and duplicate completed source trees reduced idle memory from
7.8 GB to 3.9 GB. The identical full lane is retrying with eight workers, all
runners, and no filters or source/tool changes. Retry apply/build pass, and all
79 compiler inputs match the recorded combined source. Final receipts follow. No Loom submission.
