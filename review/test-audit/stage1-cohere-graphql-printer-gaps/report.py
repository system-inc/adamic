from pathlib import Path
import json,statistics,shlex
p=Path('review/test-audit/stage1-cohere-graphql-printer-gaps');scope=json.loads((p/'scope.json').read_text());rows=scope['rows'];m=json.loads((p/'matrix.json').read_text());c=json.loads((p/'checks.json').read_text());t=json.loads((p/'timings.json').read_text());med={r:statistics.median(x['seconds'] for x in t if x['test']==r) for r in rows}
def failure(mid,row):
 for l in (p/(mid+'.log')).read_text().splitlines():
  try:e=json.loads(l)
  except:continue
  if e.get('Test')==row and e.get('OutputType')=='error':return e['Output'].strip()
def command(mid):
 x=next(x for x in m+c if x['id']==mid);return ('ADAMIC_MUTANT='+mid+' ' if mid.startswith('M') else '')+shlex.join(x['command'])+' > '+mid+'.log 2>&1'
res=[]
for r,kind,oracle,okind,mid,kills,probes,vacuous in [(rows[0],'sacred','Node executes the fixture and checks exit 0 and ready\\n; Lower must return self-written Refused.What substring. Refused.Where is not checked: M03 passes with a different refusal location.',['external-run','self'],'M02',['M01','M02'],['P01'],False),(rows[1],'witness','Self-written expected owner, planted case ID and exactly one disagreement. Node transports synthetic answers; it is not a semantic printer oracle.','self','W01',[],[],None),(rows[2],'setup-check','Self-written fixed shard count, complete union with unchanged cases, and unchanged owner after growth.','self','S01',[],['P02'],False)]:
 d=dict(test=r,package='stage1/cohere/graphql/printer',file='stage1/cohere/graphql/printer/gaps_test.go',seconds=med[r],oracle=oracle,oracle_kind=okind,kills=kills,unique_kills=kills,last_proven_fail=mid+': '+failure(mid,r),verdict=kind,subsumed_by=[],mutants_in_matrix=3 if r==rows[0] else 0,probe_kills=probes,subsumer_seconds=None,vacuous=vacuous,bounded=True,matrix_rows=rows,evidence=command(mid)+'; '+failure(mid,r))
 if r==rows[1]:d['witness_kills']=['W01']
 if r==rows[2]:d['setup_kills']=['S01','S02']
 res.append(d)
(p/'results.json').write_text(json.dumps(res,indent=2)+'\n')
plan=json.loads((p/'plan.json').read_text());checks=json.loads((p/'check-plan.json').read_text());table=[]
for x in plan+checks+[{'id':'P01','file':'internal/lower/lower.go','line':20,'change':'Lower returns nil, nil at entry','kind':'empty-answer probe'}]:
 mid=x['id'];runs=next(q for q in m+c if q['id']==mid);line=x['line']+(1 if mid=='M03' else 0)
 change=x.get('change') or ('drop return nil from ordinary-field exception' if mid=='M03' else x['before']+' -> '+x['after'])
 table.append(dict(id=mid,file=x['file'],line=line,change=change,kind=x.get('kind',x.get('menu')),failed_rows=runs['failed_rows']))
(p/'changes.json').write_text(json.dumps(table,indent=2)+'\n')
runwall=sum(x['wall'] for x in t+m)+sum(x.get('wall',0) for x in c);vetwall=sum(x['vet_wall'] for x in c)
report='\n'.join(['Unit u097: all three requested names exist at '+scope['base']+'.','Constructor: sacred within the bounded matrix; witness: witness; growth: setup-check.','Three production mutants: two caught, one survivor with changed refusal location.','Median binary seconds: 0.458, 0.201, 0.009; nproc 5 (cgroup quota 4).','Whole package cooked at 90.823 seconds; package and repository uniqueness remain unknown.'])+'\n\n'+json.dumps(res,indent=2)+'\n\n| ID | Origin/main file:line | Change | Failed rows |\n|---|---|---|---|\n'
for x in table:report+='| '+x['id']+' | '+x['file']+':'+str(x['line'])+' | '+x['change']+' ('+x['kind']+') | '+(', '.join(x['failed_rows']) or '[]')+' |\n'
report+='''
M03 survivor witness: control.log reports nestedInner.ts:5:29; M03.log reports nestedConstructor.ts:6:9, with the same Refused.What. All three assigned rows pass M03. This is changed, unguarded refusal-location behavior within this matrix, not an equivalent candidate. Other package rows may catch it.

W01 is the weakened-comparison witness check, S01 and S02 are construction faults, and P01/P02 are probes. They are separate from production kills. S02/P02 break the witness's construction precondition; those failures do not establish witness strength. Only W01 establishes its verdict. Probe P02 is judged only for the construction entry called by growth. The disagreement witness has no applicable production-entry empty probe, so vacuity is null. No families were grouped: the three rows check different things.

Clarifications, costs, and limits:

- The advertised stage1 port location is misleading for this slice. ConstructorGap calls Go lower.Lower directly and inspects a refusal. Its code under test is the compiler, not a .a printer. The other two rows test the suite's comparison and construction. No port, fixture, Node implementation, or oracle source was mutated.
- The constructor row pins a known compiler gap rather than demonstrating semantic correctness. Both production kills remove the expected refusal even though Node still executes the valid fixture. Its self-written message oracle accepts a different refusal with the same wording under M03. It also does not check Node stderr.
- The full clean package hit its 90-second binary timeout, with no preceding assertion failure observed. The narrowed clean control and three standalone timing runs per assigned row passed. Kills and unique_kills are bounded to exactly the listed matrix_rows. No package uniqueness or repository uniqueness is claimed. potential-lowering-callers.txt lists excluded callers for replay.
- TestPrinterThroughput skipped in the whole-package baseline behind ADAMIC_GRAPHQL_PRINTER_BENCH. It is outside this slice. None of the three assigned rows skipped. Prettier was enabled for baseline with ADAMIC_GRAPHQL_PRETTIER pointing at the installed external dependencies.
- Warm tools passed the env.sh check; setup was skipped. npm ci ran in stage3/api before baseline (387 ms npm-reported). The fresh optional oracle directory was installed before baseline with pinned Prettier 3.9.6 and GraphQL 17.0.2 (1 second npm-reported). It used npm install to create dependencies and a lock, rather than npm ci in that newly created directory.
- The source reach inventory contains 254 positive-coverage lowering functions, saved before mutants were chosen. Three menu mutants cover lastFieldAssignment and useOfThis; this meets approximately three per production row, rather than treating the witness and construction rows as production tests. It is not exhaustive path coverage.
- Every standalone diff passed git apply --check --cached against the starting commit and an isolated go vet overlay. M03 retains the empty if because its condition still uses field, so the deletion compiles. P01 drops the full Lower body and now-unused imports; W01/P02 replace full helper bodies to avoid unreachable-code vet failures. selector.diff captures the compiled scratch switch, which was restored after runs.
- Source-relative failure lines in overlay logs move when a body or loop is deleted. W01 failure is origin gaps_test.go:195; S01 is :205; S02 union failure is origin :209 (log :205), and its witness-precondition failure is origin :195 (log :191); P02 count failure is origin :205 (log :197). Mutation locations in the table are all against the starting origin/main.
- No native rebuild is required for the assigned compiler-refusal row because it never emits or executes a native product. Each production run nevertheless used its own ADAMIC_BUILD_CACHE_DIR. The switched Go source was compiled once through Go's cache; later modes reuse it. Command wall minus binary elapsed is only an upper bound on build/driver overhead, not a measured native rebuild time.
- /usr/bin/time was absent during final restoration verification. That command exited 127; verification was rerun with Python's monotonic timer and saved to restored.json/restored.log.

'''
report+=f'Timings: setup 0 s; dependency commands report 0.387 s and 1 s. Three-run timing commands total {sum(x["wall"] for x in t):.3f} s. Switched control and production/probe matrix commands total {sum(x["wall"] for x in m):.3f} s, binary elapsed total {sum(x["seconds"] for x in m):.3f} s. Comparison/construction/probe runs total {sum(x.get("wall",0) for x in c):.3f} s. Seven isolated vet builds total {vetwall:.3f} s. Full baseline binary elapsed 90.823 s. Coverage and final restoration logs are also included. Entire work was about 15 minutes including inspection and evidence preparation.\n'
report+='Not covered: full-package mutant matrix, repo-wide replay, all lowering paths, native port correctness and throughput. All source scratch changes restored. Evidence contains replayable diffs, source reach inventory, fixed plans, drivers, timings, JSON matrices, raw logs and results.\n'
(p/'REPORT.md').write_text(report)
(p/'survivors.json').write_text(json.dumps([dict(id='M03',before='nestedInner.ts:5:29',after='nestedConstructor.ts:6:9',evidence=['control.log','M03.log'],equivalent=False,bounded=True)],indent=2))
