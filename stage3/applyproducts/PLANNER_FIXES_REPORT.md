Added the apply non-Go read inventory with a successful cold access trace.
Fixed product 99's omitted top-level adaptation metadata and proved its key mutant fails.
Branch: typescript/stage3-apply-products; new tip is this report's containing commit.
Four cold product commands and input checks pass; individual results follow.
Lane follow-up measurements and both delivered tip SHAs are in the lane follow-up report.

# Planner follow-up, apply

READS.md contains the complete package read inventory, recipe reconciliation and
exceptions. evidence/planner-reads contains the complete successful trace,
387 repository path/operation pairs, parser and logs. The trace predates the
metadata correction; product 10's executed recipe and read set are unchanged.

## Commands and exits

All commands from the repository root unless noted. Toolchain:
`export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh`
exit 0. Setup timing lines: node 0.019s, Go 0.022s, submodules 0.060s,
markdown ready 0.079s, clang 0.156s, shared cache 0.879s, Go build 17.113s,
build cache warm 17.308s, done 17.333s. nproc=5, cpu.max=400000 100000,
so CPU quota is four cores. source /workspace/adamic-tools/env.sh for all Go commands.
Node v24.19.0, npm 11.9.0, Go 1.27.1, Python 3.12; no toolchain changed.

`timeout 90 go test -c -o /workspace/scratch/hefhych-review/apply.test ./stage3/applyproducts`
exit 0. From stage3/applyproducts, with ADAMIC_TYPESCRIPT_SOURCE=/tmp/hefhych-typescript:
`ADAMIC_BUILD_CACHE_DIR=/workspace/scratch/hefhych-review/trace-cache-fast timeout 90 strace -f --seccomp-bpf -yy -s 4096 -e trace=openat,execve,chdir,fchdir -o /workspace/scratch/hefhych-review/apply-fast.trace /workspace/scratch/hefhych-review/apply.test -test.run '^TestProduct_Stage3Adapted10$' -test.count=1 -test.timeout=90s -test.v`
exit 0; cold traced test 43.06s. The first unfiltered ptrace attempt hit the
outer 90s limit, exit 124. It is preserved as unfiltered-timeout.log.gz;
seccomp filtering avoids stops on unrelated syscalls.

Final cold cache: /tmp/hefhych-review-cache, initially absent. One fresh Go
process per boundary, previous boundaries prepared, no concurrent builds:
`ADAMIC_BUILD_CACHE_DIR=/tmp/hefhych-review-cache ADAMIC_TYPESCRIPT_SOURCE=/tmp/hefhych-typescript timeout 90 go test ./stage3/applyproducts -run '^TestProduct_Stage3AdaptedNN$' -count=1 -timeout 90s -v`
(replace NN by each row). Full logs final-ANN.log are archived below.

| Product | Test wall | Exit | Cache observation |
|---|---:|---:|---|
| Adapted10 | 25.16s | 0 | miss, predecessors hit |
| Adapted40 | 48.54s | 0 | miss, predecessors hit |
| Adapted70 | 50.68s | 0 | miss, predecessors hit |
| Adapted99 | 25.53s | 0 | miss, predecessors hit |

An earlier measurement on /workspace's overlay filesystem, before the final
recipe correction and while proof work ran, gave 32.19s, 58.46s and 65.26s
for 10/40/70. The 65.26s measurement fails the 60s law. That sequence was
stopped and is not the final proof. Final measurements above use the original
wave-1 temporary filesystem and a fresh cache. Logs retain both observations.
Products only clone the local pinned source through file:// and read installed
packages. No git fetch or npm install command runs inside them. Network namespace
is unavailable on this instance (wave-1 proof); no download appears in worker logs.

`timeout 90 go test ./stage3/applyproducts -run 'Tracks|Deadline' -count=1 -timeout 90s`
exit 0. `timeout 150 python3 -B stage3/applyproducts/prove_inputs.py` exit 0;
its four real recipe source.json omission mutants each exit 1 with
"changed source pin reused product key". Product 99 metadata omission mutant:
replace `index == 3 || entry.IsDir()` with `entry.IsDir()` in products.go,
then `timeout 90 go test ./stage3/applyproducts -run '^TestStage3Adapted99TracksMetadata$' -count=1 -timeout 90s`;
exit 1 with the same key-change assertion. The fixture changes the actual
checker-options-evidence.json in a private copy and computes both keys through
the real recipe. Recipe restored; positive key suite passes.

This follow-up changes no adapter or payload bytes. The previous whole-tree
comparison remains the payload reference; the lane report records the resulting
new product-99 payload identity. Planner read-line mutants are Planner's next
step; no claim is made that Loom's unit keys have already been changed here.
The 150 remaining Go shards, full lane rerun and original producer baseline
are outside this follow-up. Before push, the committed branch is checked by
the prescribed lane checks, and both package tests and test_apply.py are rerun.

Final verification: `timeout 90 go test ./stage3/applyproducts ./stage3/applyproducts/fetch -count=1 -timeout 90s` exit 0 (1.003s package wall);
`timeout 90 python3 -B stage3/test_apply.py` exit 0 (25 fixtures);
`timeout 90 go vet ./stage3/applyproducts ./stage3/applyproducts/fetch` exit 0.
The new product-99 tree and retained plain-main tree were independently rehashed
with prove_tree.tree_hash under timeout 90; exit 0, both
`b122537d07680ddad789bf9e543ac807efdeb6fe6748b49d325e83764306901b`.
The hash includes sorted non-.git names, full modes, symlink targets and bytes.
`git diff --check` exit 0. The prescribed committed lane-check command is:
`timeout 90 bash -c 'git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -'`.
Its result is included in the final delivery report after this commit.
