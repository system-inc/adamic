import json,re,subprocess,shlex
from pathlib import Path
out=Path(__file__).parent
base='5a94465370ea0172e25aae8c13478e59e40c2bb4';tip='f75b244b339ac7a0efca162dfc96805547763fe4'
rows=[json.loads(l) for l in (out/'suites.jsonl').read_text().splitlines()]
latest={r['package']:r for r in rows}
mutants=json.loads((out/'mutants-results.json.txt').read_text()) if (out/'mutants-results.json.txt').exists() else []
finish=json.loads((out/'finish-results.json.txt').read_text()) if (out/'finish-results.json.txt').exists() else []
counts=json.loads((out/'counts-rows.json.txt').read_text())
text=f'''Built: re-cut compiler slice 3 onto fetched main, preserving cached checker builds and main's release linker selection.
Commit: one commit on {base}; source {tip}; the final delivery SHA is in the final response.
Commands and outputs: full recursive suites were run for all 14 touched package roots; 13 package roots have passing full runs; markdown remains red. Final vet and guards pass; all attempts are recorded below.
Mutants: {sum(r['caught'] for r in mutants)}/{len(mutants)} recorded production mutants caught; both header mutants and all six direct assignment source mutants caught.
Not covered: no ahra/Loom/full gate, macOS or WASI run; full markdown validation remains red from the recorded cold setup and disk failures.

## Delivery and assumptions

Requested branch: compiler/loom-r-slice-3-ac-m. Delivered as compiler/loom-r-slice-3-ac-m-wip: the code conflict is resolved, but the mandatory full markdown run is not green and its over-budget leaves were not rerun. Conservative assumption: use the authorized WIP form for this validation blocker rather than present a normal delivery as ready. Base: {base}, the newest origin/main returned by the initial fetch, newer than 5179838b3bd78efe21c87cc359b6adf918fd0a09. Source: origin/compiler/loom-r-slice-3-ac at {tip}.

Applied only the source commit's own change with git cherry-pick --no-commit, never a main-to-source diff. The final message preserves the complete original message, adds the required Re-cut line and Co-Authored-By trailer, and records the assumptions. Conservative assumption: the fetched main snapshot is the delivery base; main is not chased during validation. The latest request for complete recursive suites takes precedence over the earlier focused-only and 30-minute first-push workflow. Existing main setup tests were kept intact rather than changed to hide cold-run failures. No main behavior was intentionally reverted beyond the source change's replacement of repeated checker-runtime compilation with its cached runtime.

A local candidate commit was necessary before finishing the full suites: main's corpus selector refuses staged GAPS.md changes against HEAD. Final evidence is folded into that same unpushed commit; the delivery still has exactly one parent commit on main.

## Conflicts and automatic merges

One conflicted hunk, internal/native/tsgo.go BuildTSGo: main compiled the runtime's C units fresh and selected linkFlags(tsgoFlags(...), options), including Linux's lld for release ThinLTO. The source reused splitTSGoRuntime, linked relative runtime/archive names, and retained Darwin's separate main.o build and fixed mtime. Kept the source's cache and relative paths, both build/runtime prefix maps, its runtime/header validation, compiler-token limits and Darwin path. Applied main's linkFlags to a clone of flags at the final link only; linker-only flags do not reach Darwin's -c compilation. conflicts.diff and tsgo-resolution.patch preserve the hunk and resolution.

Checked internal/ir/ir.go's automatic merge: Function retains both the source's SourceLength and main's CheckedUnionNarrow. Also reviewed internal/lower/functions.go and cycles.go: source-length recording and library Function recognition coexist with main's newer function lowering and static-global condition. auto-merges.patch records these checks. No other code conflict occurred. counts.md was reset to main's version and generated, not hand-merged.

## Runtime clearance

Extra runtime C/header lines beyond the original commit's own delta: NONE. Signed added/removed lines are identical to the source's own delta in adamic.h, closure.c and union.c. runtime-clearance.json.txt records the exact comparison. Main's newer runtime code stays. No static or runtime feature set was added by this resolution.

## Counts

On Linux with house clang 20.1.8, ran exactly once:

    go test ./internal/oracle -run TestCountsAreRecorded -count=1 -json -parallel=4 -timeout=10m -args -update-counts

PASS, 116.660 package seconds. Started with main's table. Changed 33 rows: 22 new and 11 existing; every key is one the source commit itself changes. Thirty-two measured rows equal the source's recorded rows. json_shorthand.a measures 20/20/8/29/4/0 here, versus the source's 18/18/8/27/4/0. Observation: main's JSON runtime has advanced; attributing those extra allocations to a particular newer change was not separately proved. The Linux measurements are retained. No unrelated row changed.

{(out/'counts-rows.md').read_text()}
## Full package tests

GOMAXPROCS=4 and affinity to four CPUs. Each command had an outer hard limit, -count=1, -json, -parallel=4 and -timeout=11m. Later recursive commands use -p 1 so oracle-owner and reader packages do not compete on the same four-CPU share. Main's compiler-free oracles were prepared with their TestProduct_ functions before subsequent reader runs. suites.jsonl holds every exact command; compressed logs and per-test seconds are alongside it. Failed-node counts include both family parents and leaves.

| Package | Attempt | Exit | Wall seconds | Failed nodes | Leaves >=60s |
|---|---:|---:|---:|---:|---:|
'''
attempts={}
for r in rows:
 p=r['package'];attempts[p]=attempts.get(p,0)+1
 text+=f"| {p} | {attempts[p]} | {r['exit']} | {r['seconds']} | {len(r['failed'])} | {len(r['overdue'])} |\n"
text+='''
Initial environment failures are preserved, not credited as compiler mismatches:

- IR first failed exporting Go archives to a full root disk; the complete retry passed. Its first raw log was overwritten by the retry; initial-disk-failure.txt and suites.jsonl retain what was observed.
- JSON's pinned historical corpus contains four /workspace/adamic tokens whose original evidence file is no longer tracked. On this builder those happen to name the current tree, so its product relocation check refused them. Built the identical pinned product in an isolated relocated worktree without changing the guard or corpus bytes, then verified the original checkout reads that same portable product (key 05bb4737819b). Both logs are retained. New review data uses .json.txt so it does not change the declared JSON source census.
- Selector initially reported an unsettled Go answers owner; its owner was prepared before retry.
- Scanner's Git helper strips the session network route and its automatic clone could not reach GitHub. Provisioned the exact v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8 with the session route, sparsely retaining src/compiler, and used its supported ADAMIC_TYPESCRIPT_SOURCE setting.
- Oracle's first full run completed with 76 failed nodes after its separate /home/agent/.cache/adamic/gate result cache filled root. Inspected failures say no space left on device. The attempt to stop it found it had already exited. Preserved this unit's result-cache files in a byte-verified archive on task scratch, then retried with ADAMIC_GATE_UNCACHED=1, a task-local XDG_CACHE_HOME, and the old build cache retained as an explicit read source.
- Full markdown ran before the candidate commit and hit the dirty-GAPS corpus guard, disk exhaustion and cold setup overruns. It remains an unresolved full-suite result; no over-budget leaf was rerun and no tests were deleted or loosened.

Timing exceptions: native's existing TestSplitTSGoAgreesUnit00 parent took 167.89s including a cold checker archive (126.89s) and source product (2.86s). The leaf-only summary excludes that parent's setup, so it is stated explicitly. Markdown's over-budget leaves were TestMarkdownListLayout_Setup 81.89s, TestProduct_WholeDocumentPreflightGoProducts 83.44s, TestProduct_MarkdownblocksTextFormatter 68.13s, and TestProduct_GrainMarkdownFormatter 88.62s. These are not represented as satisfying the 60-second requirement. They were not rerun. Do not treat this report as a fully green Loom certification.

## Focused checks, guards and mutants

The checker cache's real Node/native cold/warm witnesses passed: new TestTSGoCacheShippedRelease 4.96s, normal TestTSGoCacheRelease 5.59s, sanitized TestTSGoCacheSanitized 8.72s. They compare program output to source Node, cold/warm binary bytes, absence of builder paths, and the cached archive's identity/mtime.

All 22 changed Adamic programs pass the unchanged pinned a-check predicate using this branch's compiler: 16 clean, six refused with matching first-line reasons, zero NotYet/type-error/unclassified. Both recorded header mutants (removed header and wrong reason) remain genuinely refused and make that predicate exit 1. after.json.txt and mutants.json.txt hold observations.

The carried assignment-proof check.py still asserted the redundant pass headers the source removed. Changed only that assertion to require no a-check header on these six clean programs. Its six proofs compile, match Node, have zero stock TypeScript 6.0.3 diagnostics, and all six source statement/input mutants fail their independent Node stdout goldens. assignment-check.log and assignment-proof-logs/ preserve the outputs. No program body was changed by this re-cut.

IR's sorted-map and call-target reader guards, lower's registered-agreement-program guard, native's statics guard, and oracle's runtime-set witness guard are included in the full suites. Naming and parallelism guards were also run explicitly. No new Go package was introduced, so no compiler-consumer dependency entry was needed. All changed Go files pass gofmt; product/test diff whitespace is clean. Historical review patches/logs carried verbatim from the source contain pre-existing whitespace, so whole-diff whitespace output is not used as product evidence.

| Final focused command | Exit | Wall seconds |
|---|---:|---:|
'''
for r in finish:text+=f"| {shlex.join(r['command']).replace('|', chr(92)+'|')} | {r['exit']} | {r['seconds'] if r['seconds'] is not None else 'not measured'} |\n"
text+='\nRecorded production mutants use Go overlays; sources are .go.txt. A behavioral catch requires a test failure, excluding build/clang failures and timeouts.\n\n| Mutant | Catcher | Caught | Wall seconds |\n|---|---|---|---:|\n'
for r in mutants:
 command=r['command'];test=command[command.index('-run')+1];text+=f"| {r['name']} | {test} | {r['caught']} | {r['seconds']:.3f} |\n"
text+='''
The eight embedded source mutants are the six TestAssignmentProof*Mutant tests, TestCallSpreadEvaluatedTwiceMutant, and TestJSONStringifyShorthandWrongBindingMutant, covered by the full oracle suite. The final vet run passed after removing only completed task-owned reproducible build products to recover space. Two earlier vet attempts and the first linker-negative attempt hit ENOSPC during Go compilation, and are not credited. All 14 production mutants are caught after the six unfinished ones were retried; completed overlay archive variants were evicted from task scratch between those retries to bound disk use.

A separate linker-selection overlay omits only linkFlags from the resolved final checker link; any link failure is classified as structural/toolchain evidence, not a behavioral compiler mutant. Its successful negative retry fails TestTSGoCacheShippedRelease after 1.08s: GNU ld cannot load LLVMgold.so; the positive witness with main's lld selection passed.

Needed before a green submission: the markdown owner must address the cold setup/product grain and obtain a complete clean-tree run with adequate disk. Its four over-budget functions were not split or rerun in this lowering unit. No runtime clearance is needed for extra resolution lines, because there are none.

## Toolchain and limitations

GOPROXY was exported as https://proxy.golang.org|direct before bash cloud/setup.sh, then /workspace/adamic-tools/env.sh was sourced. nproc=5; cgroup quota is 400000/100000, four CPUs. Go 1.27.1, Node 24.19.0, clang 20.1.8. Setup succeeded; timing lines:

'''
text+='```text\n'+(out/'setup.log').read_text()+'```\n'
text+='\nNo ahra CLI, Loom submission, PR, main push or force push was performed. No new code was copied from cohere. Full markdown validation and its cold setup budgets remain unproved/failed as stated above. Original source evidence is preserved under review/compiler/loom-r-slice-3-ac/ and review/compiler/loom-r-slice-3/.\n'
(out/'REPORT.md').write_text(text)
# Every top-level test in a touched test file, with all its observed descendant timings.
files=subprocess.check_output(['git','diff-tree','--no-commit-id','--name-only','-r',tip,'--','*_test.go'],text=True).splitlines();wanted={}
for file in files:
 for name in re.findall(r'^func (Test\w+)\(',Path(file).read_text(),re.M):wanted[('github.com/system-inc/adamic/'+str(Path(file).parent),name)]=file
observed={}
for p in sorted(out.glob('*-seconds.json*'),key=lambda p:p.stat().st_mtime):
 for r in json.loads(p.read_text()):
  key=(r['package'],r['test'].split('/')[0])
  if key in wanted:observed[(r['package'],r['test'])]=r
(out/'touched-tests-seconds.json.txt').write_text(json.dumps(list(observed.values()),indent=2)+'\n')
print('Report refreshed;',len(rows),'suite attempts,',len(mutants),'production mutants,',len(observed),'touched-test observations')
