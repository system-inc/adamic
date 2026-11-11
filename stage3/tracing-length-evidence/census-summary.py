from pathlib import Path
from run import raw
import json,re
sides=['before','after','mutant-pop','mutant-popAll'];reports={}
for side in sides:
 assert json.loads((raw/(side+'-census.json')).read_text())['exit']==0
 rows=[json.loads(s) for s in (raw/(side+'-census.jsonl')).read_text().splitlines()]
 assert rows[0]['latent_mode']=='full';assert rows[-1].get('status')=='relation_report'
 row=next(r for r in rows if r.get('file','').endswith('/src/compiler/tracing.ts'))
 unitNames={u['where']:u.get('name') for u in row['units']}
 for name in ['pop','popAll']:assert any(u.get('name')==name and u['status']=='attempted' for u in row['units'])
 findings=[]
 for f in row['findings']:
  if f['kind']=='Boundary':continue
  g={k:v for k,v in f.items() if k in ['kind','where','reason','text','unit','phase']}
  for k in ['where','text','unit']:
   if k in g:g[k]=g[k].replace(str(Path(row['file']).parents[2])+'/','')
  g['owner']=unitNames.get(f['unit']);findings.append(g)
 reports[side]={'findings':findings,'pop':[f for f in findings if f['owner']=='pop'],'popAll':[f for f in findings if f['owner']=='popAll']}
for owner,line in [('pop',164),('popAll',171)]:
 original=reports['before'][owner];assert len(original)==1 and original[0]['kind']=='NotYet' and original[0]['where'].endswith(':'+str(line)+':9') and original[0]['reason']=='assigning a field of a value'
 # Any new stop introduced by the guard is recorded, not hidden.
 restored=reports['mutant-'+owner][owner];assert any(f['kind']=='NotYet' and f['reason']=='assigning a field of a value' for f in restored)
 other='popAll' if owner=='pop' else 'pop';assert not any(f['reason']=='assigning a field of a value' for f in reports['mutant-'+owner][other])
for side in sides:assert any(f['where'].endswith(':72:9') and f['reason']=='assigning a field of a value' for f in reports[side]['findings'])
def unique(side):return {(f['kind'],f['where'],f['reason'],f['text']) for f in reports[side]['findings']}
def toBefore(value):
 return re.sub(r'(src/compiler/tracing\.ts:)([0-9]+)(:)',lambda m:m[1]+str(int(m[2])-1 if int(m[2])>=165 else int(m[2]))+m[3],value)
afterMapped={(kind,toBefore(where),reason,toBefore(text)) for kind,where,reason,text in unique('after')}
new=sorted(afterMapped-unique('before'))
summary={'mode':'LATENT_FULL=1; LATENT_ONLY_FILE=adapted tracing.ts; LATENT_ASSERT_NO_OUTPUT=1','reports':reports,'newlyVisibleFindingsInTracing':new,'newlyVisibleStopsInPopOrPopAll':[f for owner in ['pop','popAll'] for f in reports['after'][owner]],'startTracing72Retained':True,'deltaCoordinateNormalization':'The added guard inserts one line; after positions at line 165 or later map back one line for delta comparison only. Raw and owner findings retain actual positions.'}
(raw/'census-summary-report.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({'popBefore':reports['before']['pop'],'popAfter':reports['after']['pop'],'popAllBefore':reports['before']['popAll'],'popAllAfter':reports['after']['popAll'],'bothIndependentRestoringMutantsReproduceOriginal':True,'newlyVisibleFindingsInTracing':new},indent=2))
