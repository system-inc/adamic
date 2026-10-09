from pathlib import Path
import json,re,statistics,shlex
p=Path('review/test-audit/stage1-cohere-gitignore');g=json.loads((p/'groups.json').read_text());m=json.loads((p/'matrix-runs.json').read_text());sp=json.loads((p/'special-runs.json').read_text());inventory=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]
names={'answers':'TestThePortAnswersAsGoCohereAndGitDo comparison family','witnesses':'TestThePortAnswersAsGoCohereAndGitDo mutant witness family','path':'TestPathAnswersAsGosPathPackage','gaps':'TestEachGapStandsWhereGapsMdSaysItDoes','planted':'TestPortAnswerShardPlantedDisagreement','growth':'TestPortAnswerShardAssignmentSurvivesCorpusGrowth'}
def events(file):
 r=[]
 for s in (p/file).read_text().splitlines():
  try:
   v=json.loads(s)
   if isinstance(v,dict):r.append(v)
  except:pass
 return r
def elapsed(file):
 for e in events(file):
  z=re.search(r'^(?:ok\s+|FAIL\s+).*\t([\d.]+)s',e.get('Output',''))
  if z:return float(z[1])
def error(file):
 for e in events(file):
  if e.get('OutputType')=='error':return e['Output'].strip()
 return None
def command(records,name):
 r=next(x for x in records if x['name']==name);return ' '.join(shlex.quote(k+'='+v) for k,v in r['env'].items())+' '+shlex.join(r['command'])+' > '+str(p/(name+'.log'))+' 2>&1'
med={k:statistics.median([elapsed('timing-'+k+'-'+str(i)+'.log') for i in [1,2,3]]) for k in g}
rows=[];matrix=[names[x] for x in g if x!='witnesses']
for k in g:
 file='stage1/cohere/gitignore/'+{'answers':'answers_shards_test.go; stage1/cohere/gitignore/gitignore_test.go','witnesses':'answers_shards_test.go','path':'path_test.go','gaps':'gaps_test.go','planted':'answers_shards_test.go','growth':'answers_shards_test.go'}[k]
 r=dict(test=names[k],package='stage1/cohere/gitignore',file=file,seconds=med[k],oracle='',oracle_kind='self',kills=[],unique_kills=[],last_proven_fail=None,verdict='',subsumed_by=[],mutants_in_matrix=4 if k in ['answers','path','gaps'] else 0,probe_kills=[],subsumer_seconds=None,vacuous=None,bounded=k in ['answers','path','gaps','witnesses'],matrix_rows=matrix if k in ['answers','path','gaps'] else [names[k]],evidence='')
 if k in ['answers','witnesses']:r['members']=[s for s in inventory if re.search(g[k],s)]
 if k=='answers':
  r.update(oracle='Go cohere and git check-ignore actually run; Git v2.51.0 t0008/t3070 scripted expectations corroborate answers. Coverage and nonzero counts are self checks. The extended clean baseline checked the script values against the port and Go cohere.',oracle_kind=['external-run','external-authority','self'],kills=['M1','M2','M3','M4'],unique_kills=['M2','M3'],verdict='sacred',probe_kills=['P1'],vacuous=False)
  log='M4-answers.log';r['last_proven_fail']='M4: '+error(log);r['evidence']='git apply '+str(p/'diffs/M4.diff')+'; '+command(m,'M4-answers')+'; '+error(log);r['unknown']=['M1: comparison shard _006 timed out; _002 fallback was interrupted. No package-wide uniqueness claim.']
 elif k=='path':
  r.update(oracle='Go standard-library path.Clean, Base and Dir are called to construct expected bytes; Node also runs the port. Built-in wrong-port subcases were checked separately with W1.',oracle_kind='external-run',kills=['M1','M4'],verdict='subsumed',subsumed_by=[names['answers']],subsumer_seconds=med['answers'],probe_kills=['P2'],vacuous=False)
  r['last_proven_fail']='M4: '+error('M4-path.log');r['evidence']=command(m,'M4-path')+'; '+error('M4-path.log');r['subsumption_basis']='2 caught production mutants; a hint, not a deletion recommendation';r['witness_kills']=['W1']
 elif k=='gaps':
  r.update(oracle='Node runs every fixture and must match handwritten GAPS.md stdout; native stdout must match the same value. All ten gaps are closed.',oracle_kind=['external-run','self'],verdict='untrue',probe_kills=['P3'],vacuous=False)
  r['evidence']=command(m,'M4-gaps')+'; ok github.com/system-inc/adamic/stage1/cohere/gitignore '+str(elapsed('M4-gaps.log'))+'s. P3: '+error('P3-gaps.log');r['limitation']='Only M4 is a reached compiler mutation. Coverage proves expression.go:508 executed. It survived this row; P3 empty-program probe fails. Untrue is limited to this frozen sample.'
 elif k=='witnesses':
  r.update(oracle='Built-in wrong ports must disagree with executed Go cohere, and with git where applicable. W1 disables the byte comparator; witness expectations then fail.',oracle_kind=['external-run','self'],verdict='witness',witness_kills=['W1'])
  r['last_proven_fail']='W1: '+error('W1-largest.log');r['evidence']=command(sp,'W1-largest')+'; '+error('W1-largest.log');r['timing_configuration']='Median of the three standard-corpus family runs, each skipping _026. The opt-in boundary member is timed separately.'
  r['boundary_seconds']=statistics.median([elapsed('largest-baseline.log'),elapsed('largest-timing-2.log'),elapsed('largest-timing-3.log')]);r['unknown']=[];r['witness_members_proven']=11;r['narrowed_evidence']='W1 shard _017: '+error('W1-narrow-017.log')
 elif k=='planted':
  r.update(oracle='Handwritten requirement: exactly one shard detects a planted disagreement, and it is portBucket(patterns/planted).',verdict='witness',witness_kills=['W1'])
  r['last_proven_fail']='W1: '+error('W1-planted.log');r['evidence']=command(sp,'W1-planted')+'; '+error('W1-planted.log')
 elif k=='growth':
  r.update(oracle='Handwritten construction invariant: adding a real-tree query preserves every existing query owner.',verdict='setup-check',construction_kills=['S1'])
  r['last_proven_fail']='S1: '+error('S1-growth.log');r['evidence']=command(sp,'S1-growth')+'; '+error('S1-growth.log')
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
mutants=json.loads((p/'plan.json').read_text())
for mu in mutants:
 mu['failed_rows']=[r['test'] for r in rows if mu['id'] in r['kills']];mu['diff']='diffs/'+mu['id']+'.diff';mu['equivalent_candidate']=False
(p/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
times={'initial':json.loads((p/'initial.json').read_text()),'timing_wall':sum(x['wall'] for x in json.loads((p/'timings.json').read_text())),'matrix_wall':sum(x['wall'] for x in m),'special_wall':sum(x['wall'] for x in sp),'standalone_mutant_build_wall':sum(x['wall'] for x in m if '-build-' in x['name']),'native_product_fetches':[]}
for f in p.glob('*.log'):
 for e in events(f.name):
  z=re.search(r'build (gitignore-port-\S+) (\S+) (miss|hit) ([\d.]+)',e.get('Output',''))
  if z:times['native_product_fetches'].append(dict(log=f.name,product=z[1],key=z[2],outcome=z[3],seconds=float(z[4])))
(p/'costs.json').write_text(json.dumps(times,indent=2)+'\n')
report='Audited 33 listed tests as 6 rows at 3774cd5ec6b2d8a8c06ac3277e4c7b58c3b58334; nproc=5.\nClean full package cooked at 90 seconds; all bounded clean groups passed.\nVerdicts: 1 sacred, 1 subsumed, 1 untrue in the sampled matrix, 2 witnesses, 1 setup-check.\nFour production mutants were caught; all three applicable empty-answer probes were caught.\nEvidence is on test-audit/stage1-cohere-gitignore under review/test-audit/stage1-cohere-gitignore/.\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n| Id | Origin/main file:line | Change | Failed rows |\n|---|---|---|---|\n'
for mu in mutants:report+='| '+mu['id']+' | '+mu['file']+':'+str(mu['line'])+' | `'+mu['from']+'` becomes `'+mu['to']+'` | '+', '.join(mu['failed_rows'])+' |\n'
report+='\nSurvivors: none among M1 through M4 in the bounded matrix. No claim about unsampled behavior.\n\nW1 is a harness weakening, S1 a permitted construction break, and P1 through P3 are probes. Their diffs are separate and do not support production uniqueness or subsumption.\n\nBrief issues and costs:\n\n- The listed shard prefix combines answer comparisons, setup/union coverage, and wrong-port witnesses in one checker. Grouping all as one production row would hide witness behavior. We split comparisons and witnesses, and retain construction coverage with the comparison family.\n- The full clean package and the M1 comparison group cooked at 90 seconds. M1 has a proven assertion kill in _014 and a timed-out _006; other kills outside the observed set remain unknown.\n- The timeout did not kill M1\'s Node process. The saved process probe shows parent PID 1 and 95.8% CPU after 12 minutes. We terminated its process group. Matrix and weakened-check costs after M1 were inflated; the three standard clean timing runs preceded this orphan.\n- W1\'s cold witness-family run also cooked, after nine witness shards had failed. The opt-in boundary was run separately, and the remaining witness shard failed on its separate rerun.\n- The 100 MiB opt-in changes the whole corpus, not just its one witness. Standard family timing excludes this optional member; its three separate enabled timings are reported in boundary_seconds. A full family median with the enlarged corpus was not measured.\n- The stage1 rebuild cap conflicts with three mutants per row. We used four frozen production mutants and separate entry probes. Native validation and cache miss/fetch times are retained in costs.json; multiple native products are required by the tests\' own built-in mutants.\n- The gap row has a distinct compiler entry. It reaches M4 but still passes. This is one reached compiler mutation, not evidence that all ten compiler features are unfailable. Its empty-program probe fails.\n- P1/P2 are module entries rather than functions, so their empty answer is implemented by deleting module execution after the helper declarations. P3 returns an empty IR program from Lower at entry under a probe-only selector.\n- Coverage lists every reached Go lowering function for the gap row. The TypeScript function inventory is a static call inventory, not instrumented branch coverage; Matcher.root and Matcher.directory are not called by the drivers.\n- No repo-wide uniqueness run was attempted. Central replay has standalone no-switch M1 through M4 diffs, validated against the starting commit and built natively; M4 also passed go vet.\n\nSetup/build/run timings:\n\n- Warm env worked; setup skipped. Tool validation '+str(times['initial']['tools']['seconds'])+' s; npm ci '+str(times['initial']['npm']['seconds'])+' s; nproc 5.\n- Standalone native mutant build commands: '+str(round(times['standalone_mutant_build_wall'],3))+' s total, including Go command compilation. Individual timings are in matrix-runs.json.\n- Three-run timing commands: '+str(round(times['timing_wall'],3))+' s wall total; production matrix commands: '+str(round(times['matrix_wall'],3))+' s wall total; witness/probe commands: '+str(round(times['special_wall'],3))+' s wall total. These phases include compilation and execution, so they are not pure compiler times.\n- Not covered: full-package green completion, the timed-out M1 shard, a full enlarged-corpus family timing, all compiler gap behaviors beyond M4, other packages, or repo-wide uniqueness.\n'
(p/'report.md').write_text(report)
print(json.dumps(dict(medians=med,rows=[(r['test'],r['verdict']) for r in rows],costs={k:v for k,v in times.items() if k not in ['native_product_fetches','initial']}),indent=2))
