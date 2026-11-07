import json
from pathlib import Path
s=Path('/workspace/scratch/core-extend')
def load(side):
 rows=[json.loads(x) for x in (s/('sites-'+side+'.jsonl')).read_text().splitlines()]
 assert len(rows)==2 and rows[0]['diagnostics']==[] and not rows[0]['checker_rejected']
 return rows[1]
def findings(r):return {(x['where'].split('/src/compiler/')[-1],x['kind'],x['reason'],x['phase']) for x in r['findings'] if x['kind'] in ['Refused','NotYet']}
before,after,mutant=map(load,['before','after','mutant']);a,b,m=map(findings,[before,after,mutant]);expected={'core.ts:1506:29','core.ts:1509:14','core.ts:1509:36','core.ts:1515:14','core.ts:1515:36','core.ts:1526:14'}
removed=a-b;added=b-a;result={'checkerDiagnostics':0,'beforeDistinctFindings':len(a),'afterDistinctFindings':len(b),'removed':[dict(zip(['where','kind','reason','phase'],x)) for x in sorted(removed)],'added':[dict(zip(['where','kind','reason','phase'],x)) for x in sorted(added)],'mutantRestoresBaseline':m==a,'ownerCoverage':{side:[u for u in r['units'] if any(f':{line}:1' in u['where'] for line in [1505,1523])] for side,r in [('before',before),('after',after),('mutant',mutant)]}}
(s/'site-comparison.json').write_text(json.dumps(result,indent=2)+'\n');assert not added,added;assert {x[0] for x in removed}==expected,removed;assert len(removed)==6;assert m==a
print(json.dumps({k:v for k,v in result.items() if k!='ownerCoverage'},indent=2))
