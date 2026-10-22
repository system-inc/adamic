import json,subprocess
from pathlib import Path
out=Path(__file__).resolve().parent
meta=json.loads((out/'base.json').read_text());checks=json.loads((out/'checks.json').read_text());counts=json.loads((out/'counts.json').read_text());events=[];skips=[]
for p in sorted(out.glob('*.log')):
 for line in p.read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if not isinstance(e,dict):continue
  if 'Test' in e and e.get('Action') in ('pass','fail','skip'):
   r={'log':p.name,'package':e.get('Package'),'test':e['Test'],'seconds':e.get('Elapsed',0),'result':e['Action']}
   (skips if e['Action']=='skip' else events).append(r)
groups=set()
for e in events:
 parts=e['test'].split('/')
 for n in range(1,len(parts)):groups.add((e['log'],e['package'],'/'.join(parts[:n])))
for e in events:e['leaf']=(e['log'],e['package'],e['test']) not in groups
leaves=[e for e in events if e['leaf']];large=[e for e in leaves if e['seconds']>=60]
(out/'test-times.json').write_text(json.dumps(events,indent=2));(out/'skips.json').write_text(json.dumps(skips,indent=2))
assert not large,large
assert all(r['exit']==(1 if r['label'].startswith('mutant-') or r['label']=='oracle-all' and any(x['label']=='oracle-all-fixed' and x['exit']==0 for x in checks) or r['label']=='other-touched-all' and any(x['label']=='ir-all-fixed' and x['exit']==0 for x in checks) else 0) for r in checks)
(out/'test-times.tsv').write_text('log\tpackage\ttest\tseconds\tresult\tleaf\n'+''.join(f"{e['log']}\t{e['package']}\t{e['test']}\t{e['seconds']}\t{e['result']}\t{e['leaf']}\n" for e in events))
mutants=[e for e in events if e['result']=='fail' and e['log'].startswith('mutant-')];(out/'mutant-catchers.json').write_text(json.dumps(mutants,indent=2))
report=f'''Built: {out.name}, preserving current main and applying the source tip's own change as one commit.
Commits: source {meta['source']}; base {meta['base']}; delivery SHA in final response.
Commands and outputs: final whole touched-package suites, counts regeneration, reader/map guard, parallel guard and vet passed; initial failures and exact argv/exits are retained in checks.json.
Mutants: current-tree lowering and/or a-check header proofs are recorded below; no compiler-build failure counts as a catcher.
Not covered: full repository gate, ahra, Loom submission or opt-in lanes that the package suites skipped (skips.json).

Fetched main and source refs. The newest fetched origin/main at branch creation is the base above. Applied only the source tip's own diff using git cherry-pick --no-commit, and independently captured the requested explicit merge-base merge-tree result. No diff against the source branch's whole history was applied. All main work outside the branch's required changes is retained.

'''
report+=(out/'resolution.md').read_text()+'\n\n'
report+='Counts: took the complete main table, never hand-merged it, then ran the full TestCountsAreRecorded -update-counts on Linux with house clang 20.1.8 after setup compiled the resolved tree. counts.patch and counts.json list the entire delta. Only this change\'s rows moved; every other main row is preserved. The precise added/deleted lines are below.\n\n'
report+='Runtime clearance: runtime.json compares source-own and recut added/deleted lines under internal/native/runtime. Additional resolution lines in runtime *.c or *.h: NONE. No runtime approval is needed for an extra resolution line.\n\n'
report+='Validation: whole-package go test commands are in checks.json. Each uses GOMAXPROCS=4, -parallel 4, -p 1, -count=1, -timeout 12m and JSON output, with an outer per-command hard deadline. Focused guards use -timeout 90s. ADAMIC_GATE_UNCACHED=1 bypasses oracle observation caches. Every requested test command actually ran; no tests were removed. Output goes directly to evidence logs. All measured leaf tests are below 60 seconds; test-times.json/tsv retains every leaf and grouping parent, including failures intentionally produced by mutants. Native/oracle suites include Node agreement, sanitizers, leaks and runtime witnesses. gofmt and Go-source whitespace checks pass. The full internal/docs whitespace check retains one inherited fixture EOF blank line, documented in the resolution notes. Vet covers the changed compiler packages. Opt-in skips are preserved and explicitly listed, not claimed covered.\n\n'
if (out/'native-leaf-audit.json').exists():
 report+='Timing detail: on this worker\'s first native run, TestSplitTSGoAgreesUnit00 was a 163.58s grouping parent because it built the shared checker C archive before its child comparisons. Its child leaves were under 60s. The initial collector mistakenly classified that grouping parent as a leaf; corrected the classification and retained both observations. The whole native suite exited zero. The group was verified again with its products prepared, matching Loom\'s product ordering; native-split-after-products in checks.json records the successful bounded run. No test was deleted, weakened or given a larger leaf budget.\n\n'
report+='Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh under a 600-second hard limit; sourced /workspace/adamic-tools/env.sh. nproc=5 and cpu.max=400000 100000 (four CPUs). House clang 20.1.8, Go 1.27.1 and Node v24.19.0. Timing lines:\n\n```text\n'+'\n'.join(s for s in (out/'setup.log').read_text().splitlines() if s.startswith('setup:'))+'\n```\n\n'
report+='Assumptions: use the latest fetched origin/main separately for each branch; retain main\'s newer infrastructure and apply only the source-own change. The newest request explicitly requires whole touched-package suites, superseding the earlier focused-only worker instruction. No cohere code was copied. No PR, ahra command or Loom submission was made. One commit is pushed only to this new own branch, without force. The commit message preserves the original message and adds the requested re-cut line and Ahra co-author trailer.\n\n'
report+='Commands and observed outcomes:\n\n'+'\n'.join(f"- {r['label']}: exit {r['exit']}, {r['seconds']}s." for r in checks)+'\n\n'
report+='Counts added lines:\n\n```text\n'+'\n'.join(counts.get('added',[]))+'\n```\n\nCounts removed lines:\n\n```text\n'+'\n'.join(counts.get('removed',[]))+'\n```\n\n'
report+='Lowering mutant catchers:\n\n'+'\n'.join(f"- {r['log']}: {r['test']}, {r['seconds']}s." for r in mutants)+'\n\n'
if (out/'header-mutants.json').exists():report+='Header mutant results:\n\n```json\n'+(out/'header-mutants.json').read_text()+'\n```\n\n'
report+=f"Measured leaf count: {len(leaves)}; maximum observed leaf seconds: {max(e['seconds'] for e in leaves)}.\n"
(out/'REPORT.md').write_text(report)
print('Report ready; max leaf',max(e['seconds'] for e in leaves))
