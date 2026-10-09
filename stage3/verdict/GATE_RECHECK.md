Merged origin/main 45487a80 into codex/stage3-verdict-harness at a9da725e.
The authored .a fixture already has the correct in-place TS2322 header.
A-check passes; lane unit tests pass 32/32; verdict unit tests pass 34/34; apply checkout test passes 1/1.
A wrong-header mutant and stand-in B's one-byte diagnostic mutant are both caught.
The remote gate log was unavailable locally; no source failure was reproduced in these package checks.

The only branch-owned .a differing from main is fixtures/case-host/input.a.
Its in-place `adamic c` result is exactly:

```
stage3/verdict/fixtures/case-host/input.a:5:7: error TS2322: Type 'string' is not assignable to type 'number'.
```

Its existing first line, `// a-check: type error TS2322`, matches that result.
No header, program, Node expected output, or suite count was changed.
The reproduction copies Gate.aCheck verbatim from devtools/fast-gate
5e9bc866c3cca817b697705146d42ca23d67b574. Its wrong-header mutant replaces
TS2322 with TS9999, gets a-check exit 1, and restores the original bytes in
finally. The restored file passes again. Stand-in A also passes all four
case-host configurations and the authored diagnostic fixture; stand-in B
changes exactly one stdout diagnostic byte and fails 0/1, with identical
stderr and exit status. [Evidence](evidence/gate-recheck/fixture-proof.json).
The fixture inventory in counts.md remains unchanged.

Commands, all with stdout/stderr redirected to files:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > setup.log 2>&1
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/verdict-gate
export PYTHONDONTWRITEBYTECODE=1
python3 stage3/verdict/evidence/gate-recheck/a-check-reproduction.py > a-check.log 2>&1
python3 -m unittest discover -s stage3/lane -p 'test_*.py' > lane-tests.log 2>&1
STAGE3_VERDICT_UPSTREAM=/tmp/stage3-verdict-auto-upstream/upstream \
  python3 -m unittest discover -s stage3/verdict -p 'test_*.py' > verdict-tests.log 2>&1
python3 stage3/test_apply.py > apply-tests.log 2>&1
STAGE3_VERDICT_NODE_TSC=/tmp/stage3-verdict-adapted/built/local/tsc.js \
  python3 stage3/verdict/prove_groups.py /tmp/verdict-834-before.json \
  stage3/verdict/selection.json /tmp/stage3-verdict-auto-upstream/upstream \
  /workspace/scratch/verdict-gate/fixture --group case-host > fixture.log 2>&1
```

Setup's successful timing lines were Node 0.039 s, Go 0.064 s, Markdown
0.269 s, submodules 0.372 s, clang 0.411 s, Go build 27.885 s,
test binaries deferred 27.997 s, build cache 27.998 s, total 28.028 s.
`nproc` is 5; cgroup CPU quota is 400000/100000. [Setup log](evidence/gate-recheck/setup.log).

The initial local apply checkout and lane clone failed because the separate,
RAM-backed /tmp volume was full. The package checks succeeded using workspace
scratch. Two full lane attempts lost an upstream test worker before its API
baseline was generated; memory.events recorded two OOM kills. One attempt
overlapped setup's Go linking; the other also had substantial retained tmpfs
captures. Automatic approval review rejected deleting older captures because
of their potential evidentiary value. They were instead preserved by moving
them to /workspace/scratch/verdict-gate/preserved. This freed tmpfs capacity
without deleting evidence. These are local observations, not an attribution
of the unavailable remote gate failure.

After preserving the older captures, the same unmodified adapted tree completed
all upstream tests with eight workers, no test filter, and unchanged lane
expectations: 106,366 passing, one failing, zero pending. The sole failure is
the expected Public APIs acknowledgement, with only api/typescript.d.ts in
the baseline diff. Oracle exit 1 is required for that sanctioned failure;
the lane guard exits 0 and prints PASS stage3 landing lane.
No additional OOM kills occurred during this successful oracle retry.

```sh
# Preserve the failed oracle output, then reuse the already applied tree.
bash stage3/oracle/run.sh /workspace/scratch/verdict-gate/lane-serial/adapted-tree \
  /workspace/scratch/verdict-gate/lane-serial/oracle > oracle-retry.log 2>&1
python3 stage3/lane/check.py /workspace/scratch/verdict-gate/lane-serial \
  > lane-check-retry.log 2>&1
```

The retained lane execution.json describes its original apply and failed
oracle invocation; its wall_seconds is not the retry duration. Both oracle
invocations exit 1, but only the successful retry has complete evidence and
passes the lane guard. oracle-report.json records the retry's actual duration
and full counts. No execution record or expected baseline was manufactured.
See [lane report](evidence/gate-recheck/lane-report.json),
[oracle report](evidence/gate-recheck/oracle-report.json), and their raw logs.
No full Adamic gate or native final-command benchmark was rerun. The branch
update is the requested main merge plus reproducible check evidence; no
unobserved source defect is claimed fixed.
