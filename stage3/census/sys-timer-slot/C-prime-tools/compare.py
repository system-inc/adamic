import sys
sys.dont_write_bytecode=True
from run import *
import collections
base=json.loads((ROOT/'stage3/census/sys-timer-slot/D-census-comparison.json').read_text())['H'];rows=[json.loads(l) for l in (S/'CP-latent.jsonl').read_text().splitlines()];sites={}
for f in rows[1]['findings']:
 if f['kind'] not in ['Refused','NotYet']:continue
 actual=f['where'].replace(str(S/'CP')+'/','');file,line,col=actual.rsplit(':',2);line=int(line);mapped=line+2 if file=='src/compiler/sys.ts' and line>=57 else line
 row={'where':file+':'+str(mapped)+':'+col,'actualWhere':actual,'kind':f['kind'],'phase':f['phase'],'reason':f['reason']};sites[(row['where'],row['kind'],row['phase'],row['reason'])]=row
key=lambda r:(r['where'],r['kind'],r['phase'],r['reason']);before={key(r):r for r in base['sites']};after=sites
result={'head2':base,'Cprime':{'checkerRejected':rows[0]['checker_rejected'],'diagnostics':rows[0]['diagnostics'],'unitStatuses':dict(collections.Counter(u['status'] for u in rows[1]['units'])),'walls':len(sites),'units':rows[1]['units'],'sites':list(sites.values())},'gone':[before[k] for k in sorted(before.keys()-after.keys())],'new':[after[k] for k in sorted(after.keys()-before.keys())],'kept':[after[k] for k in sorted(after.keys()&before.keys())]}
result['providerUnit']=next(u for u in rows[1]['units'] if u.get('name')=='getNodeSystem');result['providers']=[{'mainLine':n,'actualLine':n-2,'head2':[r for r in before.values() if r['where']==f'src/compiler/sys.ts:{n}:13'],'Cprime':[r for r in sites.values() if r['where']==f'src/compiler/sys.ts:{n}:13']} for n in [1501,1502,1599,1600]]
(S/'comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({'H':len(before),'CP':len(after),'gone':result['gone'],'new':result['new'],'kept':len(result['kept'])},indent=2))
