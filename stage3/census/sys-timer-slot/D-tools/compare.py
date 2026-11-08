import sys
sys.dont_write_bytecode=True
from run import *
import collections
result={};sets={}
for label in ['H','D']:
 rows=[json.loads(l) for l in (S/(label+'-latent.jsonl')).read_text().splitlines()];sites={}
 for f in rows[1]['findings']:
  if f['kind'] not in ['Refused','NotYet']:continue
  relative=f['where'].replace(str(S/label)+'/','');file,line,col=relative.rsplit(':',2);line=int(line);mapped=line+2 if file=='src/compiler/sys.ts' and line>=57 else line
  row={'where':file+':'+str(mapped)+':'+col,'actualWhere':relative,'kind':f['kind'],'phase':f['phase'],'reason':f['reason']};sites[(row['where'],row['kind'],row['phase'],row['reason'])]=row
 sets[label]=sites;result[label]={'checkerRejected':rows[0]['checker_rejected'],'diagnostics':rows[0]['diagnostics'],'walls':len(sites),'unitStatuses':dict(collections.Counter(u['status'] for u in rows[1]['units'])),'units':rows[1]['units'],'sites':list(sites.values())}
a=sets['H'];b=sets['D'];result['gone']=[a[k] for k in sorted(a.keys()-b.keys())];result['new']=[b[k] for k in sorted(b.keys()-a.keys())];result['kept']=[b[k] for k in sorted(b.keys()&a.keys())]
result['providerUnitD']=next(u for u in result['D']['units'] if u.get('name')=='getNodeSystem')
result['providerLocations']=[{'mainLine':line,'actualLine':line-2,'head2':[r for r in a.values() if r['where']==f'src/compiler/sys.ts:{line}:13'],'D':[r for r in b.values() if r['where']==f'src/compiler/sys.ts:{line}:13']} for line in [1501,1502,1599,1600]]
(S/'comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({'H':result['H']['walls'],'D':result['D']['walls'],'gone':result['gone'],'new':result['new'],'kept':len(result['kept'])},indent=2))
