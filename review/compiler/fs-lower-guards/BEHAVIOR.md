Replaced all four fs IR guards with programs compared against source Node for task #5x6wpqm.
Delivery branch: compiler/fs-lower-guards; this follow-up is test-only.
Each restored row passed in a separate process; source and JavaScript backend stdout and exit code agree.
Exact M09, M11, M12 and M13 diffs each failed the rewritten row.
No fs IR assertions remain; native fs execution and whole-package tests were not run in this follow-up.

The compiler/lower-agree branch was not published when checked with git ls-remote origin refs/heads/compiler/lower-agree. The local lowersAndAgreesWithNode helper follows class_static_guard_test.go: it writes a .a source, runs that source through oracle/node.mjs, lowers it, writes javascript.JavaScript(program), and runs the generated program through the same runner. Each process owns a separate t.TempDir working directory, so filesystem state starts empty for both and parallel tests cannot interfere. Node processes have 10-second deadlines. The source observation is pinned, then backend stdout and exit code are compared. Exception stderr is not compared because generated source locations differ.

Behavior observed:
- openSync('x','r'): writes a file, opens and closes it, reads back and prints opened bytes; exit 0.
- rmSync('x',{retryDelay:100}): prints present before removal and missing afterward; exit 0.
- existsSync('x'): prints missing, present after writing, and missing after unlinking; exit 0.
- statSync('missing'): prints missing, then throws before the final print; oracle/node.mjs normalizes the exception to exit 70.

All four differences are observable through behavior. No targeted IR exception is needed. The existing parseInt IR test was retained and rerun; its original Node oracle evidence is unchanged.

Mutant results:
- M09: source Node succeeds and prints opened bytes, but lowering refuses string open flags as numeric fs open flags.
- M11: source Node succeeds and demonstrates removal, but lowering refuses retryDelay:100 as outside driver defaults.
- M12: source Node exits 0 and prints missing/present/missing; backend exits 70 before printing anything because stat replaces exists.
- M13: source Node exits 70 after printing missing; backend exits 0 and additionally prints stat returned instead of throwing.

Commands from repository root, after source /workspace/adamic-tools/env.sh:
- timeout 120 go test ./internal/lower -run '^TestFS(OpenStringFlagsLower|RemoveDefaultRetryDelayLowers|ExistsOperation|StatThrowsByDefault)$' -count=1 -parallel=4 -timeout=90s -json > review/compiler/fs-lower-guards/behavior-clean.json 2>&1.
- For each exact audit diff: git apply review/compiler/fs-lower-guards/Mxx.diff; timeout 120 go test ./internal/lower -run '^NAME$' -count=1 -timeout=90s -json > review/compiler/fs-lower-guards/behavior-Mxx.json 2>&1; git apply -R the same diff.
- Restored rows: GOMAXPROCS=4 timeout 120 go test ./internal/lower -run '^NAME$' -count=1 -timeout=90s -json > review/compiler/fs-lower-guards/behavior-restored-NAME.json 2>&1.
- Lane checks after commit: git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -. Output retained in behavior-lane.log.

Separate process wall seconds, including test-process setup, on a four-CPU quota:
- TestFSOpenStringFlagsLower: 2.526s, exit 0.
- TestFSRemoveDefaultRetryDelayLowers: 2.585s, exit 0.
- TestFSExistsOperation: 2.327s, exit 0.
- TestFSStatThrowsByDefault: 2.378s, exit 0.
- TestParseIntMapUsesIndexRadix: 1.975s, exit 0.

No persistent fixtures were added, so counts.md is unchanged. Only a _test.go file and review/ evidence changed. Exact mutant diffs remain .diff files; production sources were restored.

Main c0a7667b merged without conflicts; focused behavioral rows and the retained parseInt guard passed again on the merged tip (behavior-merged.json). Lane checks passed: lane checks 1.7 s: gofmt and tools on 3 Go files, t.Parallel on 2 test packages; vet 2 packages.
