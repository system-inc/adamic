Built empty never-array lowering toward roadmap step 30, measured with step 05’s hidden census.
Implementation commit: f81387f242baabae0bca55801cc68a801ca672bf; delivery branch codex/hidden-08-never-array.
Focused Node/native/JavaScript oracles, sanitizer checks, negative checks and counts refresh pass.
Mutant: insert one numeric element into the empty never literal; both backends print 1 instead of Node’s 0 and the focused oracle fails.
Not covered: whole-corpus native correctness or the full gate; the census is explicitly measured on a checker-rejected program.

The base is origin/compiler/area-next-fixtures at dcdbb9098f77f30ad41790c56df1bd63ad462b63. No other worker branch was merged. Only the delivery feature branch is pushed.

The checker still gives these arrays element type never. The lowering uses ordinary non-reference numeric storage because no present element can inhabit never. Empty identity, undefined reads (including readonly string and number views), pop, slice, concat, evaluation order and cleanup are held to Node. Writable widening to number[] and string[] remains Refused. No backend code changed.

Shared files touched: internal/lower/object.go and internal/oracle/counts.md. New files: internal/lower/hidden_never_array_test.go, internal/oracle/hidden_never_array_test.go, the two hidden_boundary_never_array .a fixtures, and this evidence directory.

Commands run, with all output redirected to logs:

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_never_array' -count=1 -v
go test ./internal/lower -run TestHiddenNeverArrayRejectsWritableWidening -count=1 -v
go test -overlay /workspace/hidden08-evidence/length-mutant-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_never_array[.]a$' -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
```

The normal oracle and negative commands exit 0. The mutant exits 1 from actual stdout disagreement, with successful compilation and otherwise matching exit status and stderr. The oracle runs source Node, the JavaScript backend and native with ASan/UBSan/leak checks. No whole-package test or full gate was run. The counts refresh passes in 75.182s after npm ci --prefix stage3/api installs the pinned Node declarations.

Every counts row change: the minimal fixture adds 3 allocations, 3 frees, 0 retains, 4 releases, high-water 3 and 0 region bytes. The observations fixture adds 17,17,8,33,5,0. logical_and_reference_maybe.a moves to registry order with its values unchanged (8,8,11,22,4,0). The stale stage3/fixtures/taste/17_binder_flow.a row is removed because the base already registers it with lowers:false; this is not a new regression.

Setup used GOPROXY=https://proxy.golang.org|direct. The first attempt was interrupted during the uncached TypeScript checkout; a reference worktree from the existing pinned repository repaired it without copying cohere code. The successful setup reports node ready 26.902s, Go ready 26.904s, submodules ready 26.974s, markdown dependencies ready 27.003s, clang ready 27.118s, Go build ready 279.128s and done 279.316s. nproc is 5; the cgroup grants four CPUs. The printed environment is /workspace/adamic-tools/env.sh.

The exact original replay on the area base exits 0 and reproduces utilities.ts:9049:23, NotYet, an array of never. After the fix the same selector exits 1 because its expected historical finding is absent: the owner utilities.ts:9041:1 is attempted and has no findings. replay-summary.json retains both records. This is a cleared statement, not a replay crash.

The hidden measurement uses the stock AST and method at census pin 388096e6a83a4e9d287fb827f793c599ba1bf0ad, and its adapted source. All 82 adapted file SHA-256 hashes match the pinned RESULT.json. utilities.ts hashes to ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef. Both ledgers contain all 79 compiler TypeScript files, with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1. Interrupted runs were resumed by skipping completed independent file attempts; the entire project was loaded unchanged, and completed records were preserved. The resumed metadata matches the original metadata exactly. No source was adapted differently for the after measurement.

The pinned hidden.py unions failed boundaries and checker skips, then subtracts independent coverage. The pinned audit.py checks that arithmetic independently with a byte mask. Both audits pass. Compressed raw ledgers, complete calculated results and the region intersection are retained here; their group totals are not reported as this unit’s revealed bytes.

Assigned utilities.ts intersection [356194,365984): old hidden intersection 9,790 bytes; new hidden intersection 0 bytes; difference 9,790 bytes. There is no remaining boundary inside this region. The next recorded boundary after it is utilities.ts:9370:12, an ElementAccessExpression, under the owner utilities.ts:9369:1, boundary [370572,370684). That boundary is outside this unit’s assigned region and remains uncovered.

Measurement commands (scratch overlay only):

```sh
bash stage3/apply.sh /tmp/hidden-adapted # from the census-pin evidence worktree
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/hidden08-evidence/overlay-after
go build -buildvcs=false -overlay=/workspace/hidden08-evidence/overlay-after/overlay.json -o /workspace/hidden08-evidence/census-after ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /workspace/hidden08-evidence/census-after /tmp/hidden-adapted/src/compiler /workspace/hidden08-evidence/after.jsonl
python3 /workspace/hidden08-evidence/hidden.py /tmp/hidden-adapted/src/compiler /workspace/hidden08-evidence/after.jsonl /workspace/hidden08-evidence/pinned-stock.json /workspace/hidden08-evidence/after-result.json --commit f81387f242baabae0bca55801cc68a801ca672bf
python3 /workspace/hidden08-evidence/hidden-audit.py /workspace/hidden08-evidence/after.jsonl /workspace/hidden08-evidence/pinned-stock.json /workspace/hidden08-evidence/after-result.json
```

Before uses the identical commands and overlay method with object.go frozen at the area base. Exact replay uses go run ./stage3/census/latent/replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/utilities.ts:9049:23 -kind NotYet -reason 'an array of never'. All command stdout and stderr were redirected to evidence logs.
