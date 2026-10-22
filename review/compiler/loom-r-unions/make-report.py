import csv,hashlib,json,re,subprocess
from pathlib import Path
e=Path('review/compiler/loom-r-unions');source='2d9d5504a1d85dee817bd7702282e73adc31e378';base='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
focused=json.loads((e/'focused-results.json').read_text());mutants=[json.loads(l) for l in (e/'mutants/results.jsonl').read_text().splitlines()];consumers=[json.loads(l) for l in (e/'consumer-results.jsonl').read_text().splitlines()];last={(r['package'],r['kind']):r for r in consumers};selected=json.loads((e/'selected-consumers.json').read_text());packages=json.loads((e/'consumers.json').read_text());recoveries={}
for name,package in [('estree','stage1/cohere/estree'),('comments','stage1/cohere/lint/helpers/comments')]:
 rows=json.loads((e/(name+'-retry.json')).read_text());recoveries[package]=rows
 assert all(r['exit']==0 for r in rows) and len(rows)==2
assert len(focused)==14 and all(r['exit']==0 and not r['overgrain'] for r in focused)
assert len(mutants)==22 and all(r['killed'] for r in mutants)
assert len(packages)==49 and all(last[(r['package'],'build')]['exit']==0 for r in packages)
covered_by_focused={'internal/flow','internal/fresh','internal/ir','internal/lower','internal/oracle'}
for p in selected:
 if p not in covered_by_focused and p not in recoveries and p!='stage1/cohere/typeaware':assert last[(p,'test')]['exit']==0,p
guard=json.loads((e/'typeaware-deadline-guard.json').read_text());assert guard['exit']==0
limit=last[('stage1/cohere/typeaware','test')];assert limit['exit']==2 and 'test timed out after 1m0s' in (e/'stage1-cohere-typeaware-consumer.log').read_text()
checks=json.loads((e/'final-checks.json').read_text());assert all(r['exit']==0 for r in checks)
declared=json.loads((e/'declared-reader-results.json').read_text());assert len(declared)==12 and all(r['exit']==0 for r in declared if r.get('package')!='stage3/lane')
lane=json.loads((e/'lane-pinned-source-retry.json').read_text());assert lane['exit']==0
counts=json.loads((e/'counts-update.json').read_text());assert counts['exit']==0
assert all(hashlib.sha256(Path(p).read_bytes()).hexdigest()==h for p,h in json.loads((e/'code-hashes.json').read_text()).items())
assert not subprocess.check_output(['git','diff','HEAD','--','internal/native/runtime'])
# Every completed Go test event, including failed setup attempts, is retained with its log.
events=[]
for p in sorted(e.rglob('*.log')):
 if p.name.endswith('-build.log') or p.name.endswith('-build.first.log'):continue
 text=p.read_text(errors='replace');parsed=[]
 for line in text.splitlines():
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:
   parsed.append(dict(log=str(p.relative_to(e)),package=r.get('Package',''),test=r['Test'],seconds=r.get('Elapsed',0),result=r['Action'],measurement='Go test elapsed'))
 if not parsed:
  for state,test,seconds in re.findall(r'^\s*--- (PASS|FAIL|SKIP): (\S+) \(([0-9.]+)s\)',text,re.M):parsed.append(dict(log=str(p.relative_to(e)),package='',test=test,seconds=float(seconds),result=state.lower(),measurement='Go test elapsed'))
 events.extend(parsed)
assert all(r['seconds']<60 for r in events),[r for r in events if r['seconds']>=60]
events.append(dict(log='stage1-cohere-typeaware-consumer.log',package='stage1/cohere/typeaware',test='TestInspectRequestRefusals',seconds=limit['seconds'],result='timeout',measurement='process wall; Go alarm 60 seconds'))
with (e/'test-seconds.csv').open('w') as f:
 writer=csv.DictWriter(f,fieldnames=['log','package','test','seconds','result','measurement']);writer.writeheader();writer.writerows(events)
(e/'test-seconds.json').write_text(json.dumps(events,indent=2)+'\n')
(e/'final-results.json').write_text(json.dumps(dict(focused=focused,mutants=mutants,consumer_attempts=consumers,consumer_recoveries=recoveries,final_checks=checks,counts_update=counts,typeaware_deadline_guard=guard,typeaware_limit=limit,declared_readers=declared,lane_pinned_source_retry=lane),indent=2)+'\n')
counts_events=[r for r in events if r['log']=='counts-update.log'];moves=json.loads((e/'counts-diff.json').read_text());assert len(moves)==17
catchers={}
for m in mutants:
 errors=[]
 for line in (e/'mutants'/(m['topic']+'.log')).read_text().splitlines():
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('OutputType')=='error':errors.append(r.get('Output','').strip())
 catchers[m['topic']]=errors
(e/'mutant-catchers.json').write_text(json.dumps(catchers,indent=2)+'\n')
lines=[
'Built: #49yq1ah step 17 unions applied as one commit on exactly the requested main base.',
'Commits: source '+source+'; base '+base+'; delivery SHA is the pushed compiler/loom-r-unions branch tip.',
'Commands and outputs: all 14 original focused command groups passed; 49 imported plus 11 declared reader packages compiled, other selected checks and all reader vet passed; typeaware request check hit its 60-second limit; counts regenerated once.',
'Mutants: all 22 recorded source mutants caught by their named assertions, plus the embedded IR field-check removal mutant caught by its baseline/unchecked-output comparison.',
'Not covered: completion of TestInspectRequestRefusals (60-second timeout, never rerun); exhaustive package tests/full gate, ahra loom check, WASI, general callable/object-kind unions, optional writes/delete, hidden never-array lowering.',
'',
'Exact-base application: git fetch origin, checkout -b compiler/loom-r-unions dbdf977f5b, and git cherry-pick --no-commit '+source+'. The source SHA was absent after the ordinary fetch; an explicit bounded fetch of that SHA recovered it. Only its own parent-to-commit change was applied. No comparison from main to the source tip was applied. Main was not advanced or rebased to a newer base.',
'',
'Conflicts and resolutions: none. Git automatically merged internal/lower/expression.go, internal/lower/object.go and internal/oracle/counts.md. Main\'s library-spread infrastructure remains in object.go beside the new checked union field reads. There are no conflicted hunks. The counts auto-merge was discarded in favor of main\'s complete table before the single successful regeneration; no counts row was hand-merged.',
'',
'One extra fixture-layout correction is required by this unit: rename scout_nullable_storage/ to scout_nullable_storage_refused/ and update nullableObjectStorageOutcome\'s reader. All six .a sources are byte-for-byte unchanged, all six Node/NotYet tests remain, and no flow or freshness test is edited. Conservative assumption, also stated in the commit message: keep unsupported shared-storage cases explicitly refused and use bounded representative checks for each reader package rather than an exhaustive package/full-gate run.',
'',
'Runtime C lines for clearance: none. Both git diff '+source+'^ '+source+' -- internal/native/runtime and the delivery diff against the pinned main base are empty. runtime-comparison.json records both empty diffs. The three added native/union.go lines are Go emission, retained from the source change. Main\'s runtime work is preserved.',
'',
'Counts regeneration: '+str(counts['seconds'])+' seconds command wall, exit 0, Linux with house clang 20.1.8. Exact argv is counts-update.json and output counts-update.log. All '+str(len(counts_events))+' Go pass records (including shard parents) completed, maximum measured leaf '+str(max(r['seconds'] for r in counts_events))+' seconds. Regeneration measured the full generated counts family once, not just the new fixtures. Independent TestCountsAreRecordedTable then passed. There are exactly 17 changed rows: 16 newly admitted fixture rows and one existing row. The latter, taste_stage3_representations.a, changes retains/releases 11/34 to 13/36. The current-tree counterfactual disables only the new checked string-field route and restores 11/34, and scout_boxed_scalar_fields.a returns to 95/150 from 114/169. That control passed; counts-cause.log preserves the observation. The counterfactual retains main\'s newer library-spread code.',
'',
'| Counts fixture | Before | After | Cause |','|---|---|---|---|']
def numbers(row):return 'absent' if row is None else '/'.join(x.strip() for x in row.split('|')[2:-1])
for r in moves:lines.append(f"| {r['fixture']} | {numbers(r['before'])} | {numbers(r['after'])} | {'new admitted fixture' if r['before'] is None else 'new checked string-field route; counterfactual verified'} |")
lines += ['',
'Focused validation uses the exact source-report selections. run-focused.py records all argv, exits, command wall seconds and Go leaf events in focused-results.json. The fixtures run source Node, native release/ASan/UBSan/leak checks and backend Node as their harness specifies; checked stale-flow fixtures retain the inserted-panic assertions. Coverage includes nullable strings and objects, scalar field tags, optional boolean class fields, all six storage refusals, RegExp adapter, changed stage1 gap leaves, flow/freshness remainders and native ownership consumer. Guards include SortedMap IR storage, exact runtime-feature witnesses, registered native agreement rows, parallel tests and TestCallTargetReaders (-timeout 90s). No new IR map, package, runtime feature set or build product was introduced by this recut; no new compiler dependency declaration is needed. Product/test sources add no fixed build-directory or machine-path literals. No cohere implementation was copied.',
'',
'| Focused command group | Exit | Command seconds | Completed Go test records |','|---|---:|---:|---:|']
for r in focused:lines.append(f"| {r['label']} | {r['exit']} | {r['seconds']:.3f} | {len(r['events'])} |")
lines += ['',
'All 49 touched or direct-reader packages were discovered from go list production, test and external-test imports. consumers.json and selected-consumers.json name them and their checks. go test -c compiled every package; three are compile-only, and the remaining 46 have selected test coverage (five core packages reuse the focused checks instead of rerunning them). A further 11 compiler readers declared in cloud/fast-gate/compiler-dependencies.json, including CLI/source readers without direct compiler imports, were compiled and checked with the bounded selections in declared-readers.json. All selected checks passed except the typeaware request timeout. These are representative exact test selections, not whole-package runs. Each additional consumer test has a 60-second Go limit and a 65-second outer kill; Go compilation has its own 180-second bound. Every completed test leaf in all logs, including setup failures and mutants, is below 60 seconds; the one timed-out check is explicitly red. The maximum completed leaf is '+str(max(r['seconds'] for r in events if r['result']!='timeout'))+' seconds. TestInspectRequestRefusals hit its 60-second alarm (60.050 seconds process wall) and was not rerun. Test-binary compilation and command wall time are distinguished from test-leaf setup; some commands exceed 60 seconds because they compile Go before tests start. No over-budget test was rerun.',
'',
'Cold setup recovery, with original failed logs retained:',
'',
'- ESTree initially stopped because its Go oracle owner was unprepared. TestProduct_EstreeGoOracle in stage1/cohere/estree/oracles passed, then TestAcceptanceGrammar passed (48.81 seconds leaf elapsed). estree-retry.json records owner/consumer command times.',
'- The root overlay filled (32 GB; /tmp is a separate 8.8 GB filesystem). A comments check, two compile-profile checks and four package builds failed from ENOSPC; the first consumer driver then stopped while opening a new log. Prior-unit scratch trees and one immutable cached export product were relocated to /tmp, preserving their contents and the cache path via a symlink. The detached worktree metadata was repaired. space-recovery.json and cache-relocation.json record the exact moves. No source, test or runtime code was changed for storage recovery.',
'- Only failed or unrecorded packages resumed. Comments then exposed its separate required Go oracle owner. TestProduct_CommentsOracle passed (19.72 seconds leaf), then TestJsxParserGapIsExplicit passed. All original attempts remain in consumer-results.jsonl and *.first.log; the comments owner recovery is in comments-retry.json. Every successful final reader check is accounted for in final-results.json.\n- Typeaware TestInspectRequestRefusals timed out at its 60-second alarm while preparing checker archives and native probes (60.050 seconds process wall). The log records checker preparation 13.505s, wrong-kind archive 11.93s, unknown-question archive 11.38s, plus several 3.7s clang steps. No union assertion failure preceded the timeout. This does not establish behavior on the base or prove the timing is independent of this change. It was never rerun or weakened. The separate TestTypeAwareUnitDeadline passed; this is additional bounded package coverage, not completion of the timed-out request check.\n- stage3/lane TestApplyPython initially failed in 0.54s because its ordinary-apply test requires the pinned TypeScript gate checkout. Providing ADAMIC_TYPESCRIPT_SOURCE for that command with the existing verified v6.0.3 checkout (050880ce59e30b356b686bd3144efe24f875ebc8) resolved it. The retry passed without changing product/test source; lane-pinned-source-retry.json records its command seconds.',
'',
'Vet passed for all 49 touched/direct-reader packages ('+str(next(r['seconds'] for r in checks if r['label']=='all-consumer-vet'))+' seconds, exit 0), and separately for all 11 additional declared readers; changed-package vet also passed as the original focused group. gofmt and source whitespace checks are checked before committing. The catalog 08 patch applies cleanly with git apply --check; the full catalog runner was not run.',
'',
'Every recorded source mutant and its actual catcher follows. All mutations were rebuilt as Go overlays over the current files, preserving main additions; no stale whole-file overlay from the old source was used. Patches and mutant sources are .patch/.go.txt. The first preparation stopped before executing a mutant because a one-line seam was ambiguous; three lines of unchanged context resolved it. No compiler warning, build failure or Go panic is counted as a semantic kill.',
'',
'| Mutant | Catcher | Command seconds | Failure observed |','|---|---|---:|---|']
for r in mutants:
 names=[v['Test'] for v in r['catchers'] if '/' in v['Test'] or len(r['catchers'])==1]
 error=catchers[r['topic']][0] if catchers[r['topic']] else 'see mutant log'
 error=re.sub(r'^\s*\w+\.go:\d+:\s*','',error).replace('|','\\|')
 lines.append(f"| {r['topic']} | {'; '.join(names)} | {r['seconds']:.3f} | {error} |")
lines += ['',
'The 23rd mutant is TestNullableStringFieldCheckMutant\'s IR field-check removal. The checked baseline panics with exit 70, while the unchecked JavaScript mutant prints undefined and exits 0. This independent behavior comparison passed in oracle-focus.log. Source-mutant catcher errors and exact leaf seconds are also in mutant-catchers.json and test-seconds.csv/json. Every completed focused, consumer, owner, count, counterfactual and mutant leaf\'s seconds are preserved there, including original failed setup attempts.',
'',
'Setup: export GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Completed without failure in 85.831 seconds. nproc=5; cgroup cpu.max=400000 100000, four CPU shares. GOMAXPROCS=4 and ADAMIC_GATE_UNCACHED=1 were used for validation. Timing lines:',
'', '```text']
lines += [l for l in (e/'setup.log').read_text().splitlines() if l.startswith('setup:')]
lines += ['```','','No PR, ahra CLI check or Loom submission was made. Only compiler/loom-r-unions is pushed, without force.']
(e/'REPORT.md').write_text('\n'.join(lines)+'\n')
print('packages',len(packages),'mutants',len(mutants),'recorded tests',len(events),'max leaf',max(r['seconds'] for r in events))
