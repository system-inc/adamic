import json,re
from pathlib import Path
p=Path(__file__).resolve().parent
# When invoked in scratch, read the repository evidence directory.
if not (p/'plan.json').exists():p=Path('/workspace/adamic/review/test-defend/internal-oracle-node_fs_file')
plan=json.loads((p/'plan.json').read_text());matrix={};timings={}
def parse(f):
 out={'passed':[],'failed':[],'skipped':[],'errors':[],'seconds':None}
 for l in f.read_text(errors='replace').splitlines():
  try:x=json.loads(l)
  except:continue
  a=x.get('Action');t=x.get('Test')
  if t and '/'not in t and a in ['pass','fail','skip']:out[{'pass':'passed','fail':'failed','skip':'skipped'}[a]].append(t)
  if x.get('OutputType')=='error':out['errors'].append(x['Output'].strip())
  if not t and a in ['pass','fail']:out['seconds']=x.get('Elapsed')
 return out
for f in p.glob('*.log'):
 v=parse(f)
 if v['seconds'] is not None:timings[f.name]=v['seconds']
for id in ['D1','D2','D3']:matrix[id]=parse(p/(id+'-expanded.log'))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'timings.json').write_text(json.dumps(timings,indent=2)+'\n')
rows=[]
for name,sub,id in [('TestCheckedNonNullAdamicRefusal','TestNonNullAdamicRefusal','D2'),('TestPossibleNonNullAdamicAssertionsAreRefused','TestCheckedNonNullAdamicRefusal','D3'),('TestNonNullAdamicRefusal','TestCheckedNonNullAdamicRefusal','D1')]:
 m=next(x for x in plan if x['mutant']==id);regex="^Test(CheckedNonNull.*|ImpossibleNonNull.*|PossibleNonNull.*|MigratedNonNull.*|NonNull.*|ParserNonNull.*|StatementsSmallRulings|FractionalPowersReachRuntime|ReviewPrograms.*)$";command="ADAMIC_DEFENSE_MUTANT="+id+" ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/nonnull-defense/cache/"+id+"-expanded timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run '"+regex+"' > "+id+"-expanded.log 2>&1"
 unique=matrix[id]['failed']==[name] and len(matrix[id]['passed'])==15
 rows.append({'test':name,'package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':sub,'defense':'defended' if unique else 'cannot-judge','unique_mutant':id+' '+m['file']+':'+str(m['line']) if unique else None,'attempts':[{'mutant':id,'file_line':m['file']+':'+str(m['line']),'change':m['menu']+'; '+m['new'],'rows_failed':matrix[id]['failed']}],'code_under_test':'Adamic non-null refusal policy and source-position reporting','oracle':'self: hand-written Refused kind, What/Fix text'+(' and exact diagnostic positions' if id=='D1' else ''),'evidence':command+'; '+'; '.join(matrix[id]['errors']),'bounded':True,'matrix_rows':sorted(matrix[id]['passed']+matrix[id]['failed']),'outside_matrix':'Unknown; whole package exceeded 90 seconds.'})
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
