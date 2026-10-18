import json,pathlib
p=pathlib.Path(__file__).resolve().parent
commands=json.loads((p/'commands.json').read_text())
summary=json.loads((p/'test-summary.json').read_text())
counts=json.loads((p/'counts-row-changes.json').read_text())
focused=json.loads((p/'focused-test-seconds.json').read_text())
mutants=[('omit-boundary-inputs','TestClosureTargetsKnownCallerDoesNotHideUnknownCaller','known caller hid unbounded caller'),('unknown-is-empty','TestClosureTargetsBoundOnlyProvenValues','bounded without proof'),('omit-captures','TestClosureCycleDirect','want cycle refusal'),('drop-return-target','TestClosureTargetsFollowParametersReturnsAndMapValues','want both returned targets'),('omit-map-captures','TestClosureCycleMap','want cycle refusal'),('type-matched-captures','TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd','adamic/cycle-capable')]
for name,test,assertion in mutants:
    text=(p/(name+'.log')).read_text()
    assert assertion in text and 'build failed' not in text and '--- FAIL:' in text,name
required={'focused','counts-update','affected-fixtures','instantiation-find','mutants','admission-mutant','readers','vet','consumers','flow-corpus','table-and-families'}
assert required <= {c['name'] for c in commands if c['exit']==0}
assert all(x['seconds']<60 for x in focused)
assert all(c['exit']==0 for c in commands if c['name'] not in ('focused-initial','counts-unmatched'))
assert all(v['fail']==0 for v in summary.values())
assert len(json.loads((p/'lower-tests.json').read_text()))==298
assert len([c for c in commands if c['name'].startswith('lower-shard-')])==15
assert 'flow-corpus.log' in summary
text='''Rebased the closure-target cycle rule onto pinned main, retaining main's newer lowering guards.
Commits: delivery SHA in final response; base dbdf977f5b8b0979d414ebe6f20e7aadd312e309; source 20fde75fa96b48f3dea2e3612b418acae666332f.
Checks: 35 focused witnesses/guards, all 298 lowering top-level tests in 15 bounded shards, full flow corpus, readers, affected fixtures, counts and vet pass.
Mutants: all six source-report mutants fail at their intended assertions; every production file is restored.
Not covered: full gate, unrelated oracle/consumer leaves, WASI, benchmarks, full pristine-source cycle ledger or admission census; one existing mixed-union array child skips pending views-v3. No ahra, PR or Loom submission.

Applied only the source commit's own change using `git cherry-pick -n 20fde75fa96b48f3dea2e3612b418acae666332f` onto the exact pinned base. No conflicts or conflicted hunks occurred. Main's generated async cycle identity audit remains intact and TestAsyncGeneratedIdentityCannotBeClaimedBySource passes in the complete lowering run. The source's compiler patch passes reverse-apply checking on the rebased tree. There are no additional production/compiler edits. The cache's private contents map remains SortedMap, as in the source.

Conservative assumption: unknown closure producers remain cycle-capable. The function-slot decision uses actual producer flow and captured cells; no function-type compatibility guess supplies a bounded target set. The previously stale cwd refusal expectation is already corrected in the source change. Its exact program is stage3/namespace-init-sys/cwd.a: memoize(() => process.cwd()), a string cache, Debug.enabled initialization, and a comparison with process.cwd(). The old type-matched rule incorrectly put the returned memoizer closure into callback's target set and refused callback as cycle-capable. The actual callback captures no environment; the returned closure owns callback and a string cache, which cannot reach it. Node, JavaScript, release native and ASan/UBSan plus leak checks agree on true. Unsupported realpath remains independently stopped at node:fs.native; executable namespace reads before initialization remain rejected.

Counts were restored from pinned main and regenerated once successfully on Linux with house clang 20.1.8 using the documented unanchored -run TestCountsAreRecorded -args -update-counts command. The first anchored attempt selected no tests and wrote no table; its failed writer output is preserved. Regeneration adds eleven rows, changes six existing rows and removes none. All numeric changes from the source report match. Main's newer instantiation_find.a is the sole additional changed row: 13/13/28/31/7/0 becomes 13/13/22/25/7/0. Observation: allocations/frees/peak/regions stay unchanged and six retain/release pairs disappear. Inference: the now bounded getPattern callback flow improves borrowing. Its dedicated TestInstantiationFind independently passes uncached Node/JavaScript/release/sanitized native and leak agreement. The five source-report affected existing fixtures also pass uncached; their actual fixture bodies ran, not just family wrappers.

Column order below: allocations/frees/retains/releases/peak/regions.

| Fixture | Pinned main | Rebased |
|---|---|---|
'''
for row in counts:
    render=lambda x:'new' if x is None else '/'.join(map(str,x))
    text+=f"| {row['fixture']} | {render(row['before'])} | {render(row['after'])} |\n"
text+='''
Runtime C: the source's own delta and this rebased delta under internal/native/runtime both contain zero changed lines. No runtime lines need clearance. See runtime-delta-comparison.json. No new runtime feature set or compiler-consuming package was added; existing declarations and exact runtime witnesses remain, and TestRuntimeSetsAreEachUsed passes. New refused fixtures remain under cycle_closure_targets_refused/, outside flow's top-level fixture glob. No cohere code was copied.

Every recorded production test result below passes, with the existing mixed-union child skip noted above. Full per-test timings are in test-seconds.json; the 35 own focused events are also listed below. Counts' successful regeneration contains 1102 passing events. The flow corpus runs with ADAMIC_UNIT_BUDGET=1, GOMAXPROCS=4 and -parallel=4. Full lowering is selected from go test -list and split into 20-top-level-test shards, each with a 90-second Go timeout and 120-second outer limit. Setup builds every package; consumer tests focus on the call-target, borrowing, loop-call, flow and JavaScript paths, with the full flow corpus additionally exercised. Unrelated package tests are left to Loom as requested.

| Log | Passing events | Skipped events | Maximum event seconds |
|---|---:|---:|---:|
'''
for log,s in sorted(summary.items()): text+=f"| {log} | {s['pass']} | {s['skip']} | {s['max_seconds']} |\n"
text+='''
Every mutant was caught by an assertion, never by a build error or timeout. Sources are saved only as .patch files. The runners use a 90-second Go timeout, 120-second subprocess limit, and restore originals in finally blocks.

| Mutant | Catcher | Assertion |
|---|---|---|
'''
for name,test,assertion in mutants:text+=f'| {name} | {test} | {assertion} |\n'
text+='\nOwn focused test timings:\n\n| Test | Seconds |\n|---|---:|\n'
for x in focused:text+=f"| {x['test']} | {x['seconds']} |\n"
text+='''
The initial focused attempt reached its 180-second outer limit during cold Go compilation, with no test events; the subsequent attempt passed all 35 events. The unmatched counts attempt measured and wrote no rows. Both attempts and successful retries are recorded rather than hidden. No test leaf exceeded 60 seconds, and no over-budget leaf was rerun.

Toolchain setup passed. nproc=5; cgroup cpu.max=400000 100000 (four CPUs); Go 1.27.1, Node 24.19.0, clang 20.1.8. GOPROXY=https://proxy.golang.org|direct; source /workspace/adamic-tools/env.sh; GOMAXPROCS=4 for checks. Setup timing lines:

```text
'''
text+='\n'.join(line for line in (p/'setup.log').read_text().splitlines() if line.startswith('setup:'))+'\n```\n'
text+='''
All test output goes directly to the named log, never through a pipe. commands.json records commands, exits and wall seconds, including both unsuccessful attempts. Fetch/apply/setup commands:

```text
timeout 120 git fetch origin
timeout 120 git fetch origin 20fde75fa96b48f3dea2e3612b418acae666332f
git checkout -b compiler/loom-r-cycle-targets dbdf977f5b
timeout 120 git cherry-pick -n 20fde75fa96b48f3dea2e3612b418acae666332f
git restore --source=dbdf977f5b --staged --worktree internal/oracle/counts.md
export GOPROXY='https://proxy.golang.org|direct'
timeout 600 bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
export GOMAXPROCS=4
```

Verification commands and outputs:

```text
'''
for c in commands:text+=f"{c['command']}\n  exit={c['exit']} wall={c['seconds']}s log={c['name']}.log\n"
text+='```\n\nMutation runner source files record each exact inner go test command and mutation site. The final source-delta-retained.json and runtime-delta-comparison.json were rechecked after all mutations. Strict whitespace checks pass for docs/internal; patch evidence preserves unified-diff context whitespace. Historical review files inherited from the source commit remain historical; this directory is the evidence for this rebase.\n'
(p/'REPORT.md').write_text(text)
print('wrote REPORT.md')
