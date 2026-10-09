Stage 3 now publishes full Mocha errors, original stacks and per-runnable timing beside its verdict.
Each failing test has one of five cause classes and linked first-40-line baseline diff previews.
The full adapted main lane remains PASS: 106,366 passing, one sanctioned API failure, zero pending.
Both real failure fixtures explain themselves, 40 lane tests pass, and the dropped-text mutant is caught.
Expected counts, reference outputs, inspect, api_check and the sanctioned-failure rule are unchanged.

The branch starts at origin/main 73352e874ddbda5a78c95c5c670860d43275e4d0.
Only stage3/oracle and stage3/lane are edited. The fast gate already publishes
report.json and verdict.json, which both embed full messages, stacks and diff
previews. failures.md is a readable companion beside verdict.txt in lane output.
Every cause count is appended to that line, including the sanctioned API failure.

| Run | Observed result | Cause | Evidence |
|---|---|---|---|
| Huge declaration, targeted 1 ms Mocha timeout | 0 passing, 1 failing; hook elapsed 1351.139 ms | timeout=1 | [report](evidence/errors/timeout/lane-report.json) |
| Protected1 type/symbol baseline with a scratch input addition | 0 passing, 1 failing; two changed files | baseline-content=1 | [report](evidence/errors/baseline/lane-report.json) |
| Fresh full adapted main tree | 106366 passing, 1 failing, 0 pending; lane PASS | baseline-content=1 | [report](evidence/errors/normal/lane-report.json) |
| Lane tests | 40/40, no skips | all five classes covered | [log](evidence/errors/lane-tests.log.gz) |
| Dropped-message mutant on real timeout evidence | message assertion rejects it; status and causes unchanged | text-only catch | [proof](evidence/errors/proof.log.gz) |

The timeout's full text includes `Timeout of 1ms exceeded` and the Mocha async
completion guidance. Elapsed time is measured from original Runnable.run to
the failure event with process.hrtime.bigint; it is not fabricated from the
limit. The observer records the original worker failure before upstream's
parallel replay, whose transport omits elapsed time. It calls the original run
and emit methods with the same arguments and preserves their return values.
Capture errors remain metadata, not acceptance decisions. A Node fixture also
checks exact zero-, one-, and multiple-payload argument lists and return values. A replay test covers
parallel root-suite titles beginning with an empty element.

The Protected1 fixture temporarily appends `const stage3ForcedMismatch = 123;`
to the scratch checkout's existing conformance input. It restores original
bytes in finally. References are untouched. The real harness produces both
Protected1.types and Protected1.symbols differences. Their respective first
12 and 9 lines are attached to the failed Correct type/symbol baselines test,
with the full 3,786-character error and original harnessIO stack top retained.
The normal sanctioned API error retains all 57,312 characters, alongside its
bounded preview. Full diffs and original per-process event records are saved
in the evidence directories.

Commands used, with all stdout/stderr directed to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /workspace/scratch/step12-setup.log 2>&1
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/step12 PYTHONDONTWRITEBYTECODE=1
bash stage3/lane/run.sh /workspace/scratch/step12/normal > normal-command.log 2>&1
bash stage3/oracle/run.sh TREE /workspace/scratch/step12/timeout-final \
  --runners compiler --tests hugeDeclarationOutputGetsTruncatedWithError.ts \
  --timeout 1 > timeout-final-command.log 2>&1
# In a scratch TREE, add the stated Protected1 input line and restore it in finally.
bash stage3/oracle/run.sh TREE /workspace/scratch/step12/baseline-observed \
  --runners conformance --tests 'Protected1.ts.*Correct type/symbol baselines' \
  > baseline-observed-command.log 2>&1
python3 -m unittest discover -s stage3/lane -p 'test_*.py' > tests-final.log 2>&1
python3 stage3/lane/prove_errors.py /workspace/scratch/step12/timeout-final \
  /workspace/scratch/step12/baseline-observed /workspace/scratch/step12/normal \
  > proof-final.log 2>&1
node --check stage3/oracle/observe-errors.cjs > observer-syntax.log 2>&1
```

Each targeted oracle's lane/ folder links its original oracle evidence and the
scratch adapted tree, with recorded platform and exits. The actual guard keeps
these subset observations red: they lack the sanctioned API snapshot and do
not satisfy the full-run counts or filter rules. Diagnostics survive that
missing-evidence error. No expected file or gate predicate is replaced for
these probes. The normal lane applied this branch's main tree fresh and ran
all runners with eight workers. After correcting event-title normalization,
the guard was replayed on the unchanged full-run logs and original events to
verify enrichment; upstream was not rerun or its counts changed. The retained
oracle report is the original live report; the lane report contains the final
replayed details and timing from the captured event.

The proof script mutates only the returned message, retaining timeout class,
counts and red status. Only the assertion for the full timeout text kills it.
Unit tests independently cover a passing sanctioned observation with its
message removed, keeping that observation's acceptance unchanged. Additional
fixtures cover baseline-missing, exceptions with stack top, other failures,
multiple linked files, preview truncation, captured failures without a final
reporter summary, missing API evidence, and metadata
collection errors that cannot affect acceptance. counts.md is refreshed.

Setup completed in 35.022 s: Node 0.027 s, Go 0.031 s, submodules 0.079 s,
Markdown dependencies 0.080 s, clang 0.171 s, Go build 34.869 s,
test binaries deferred 34.991 s, cache warm 34.993 s. nproc=5;
cgroup CPU quota is 400000/100000. [Raw setup log](evidence/errors/setup.log.gz).

Only Linux x64 / Node 24.19.0 was measured. The original remote four-failure
incident was not reproduced; the forced hook and baseline fixtures exercise
its relevant report paths. Historical reports without original event records
still retain reporter text and stacks, with elapsed_ms explicitly null. No
full Adamic gate or compiler performance measurement was run.
