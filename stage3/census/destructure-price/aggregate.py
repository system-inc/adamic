from pathlib import Path
import json,collections
root=Path('/workspace/scratch/destructure-price-f2989762/r2');report=json.loads((root/'clean-oracle/report.json').read_text());meta=json.loads((root/'clean-lane.json').read_text())
assert report['counts']=={'passing':106366,'failing':1,'pending':0},report['counts']
assert report['baseline_diffs']==['api/typescript.d.ts'],report['baseline_diffs']
assert report['phases']['install']['exit']==0 and report['phases']['build']['exit']==0 and report['phases']['tests']['exit']==1
rows=[json.loads(p.read_text()) for p in (root/'clean-processes').glob('*.json')];workers=[r for r in rows if any(a.startswith('--config=') for a in r['argv']) and r['argv'][1].endswith('/run.js')]
assert len(workers)==8,len(workers)
enter=collections.Counter();calls=collections.Counter()
for row in rows:enter.update(row['enter']);calls.update(row['calls'])
sites=json.loads(Path('/tmp/destructure-price-scratch/ast-sites.json').read_text());pub={p['name'] for p in json.loads(Path('/tmp/destructure-price-scratch/public-types.json').read_text())['names'] if p['typescript']}
heat=[];price=0
for site in sites:
 key=site['file']+':'+str(site['line']);public=site['type'].split('<')[0] in pub
 item={**site,'site':key,'public':public,'enclosingRuns':enter[key],'members':[{**m,'calls':calls[key+'::'+m['name']]} for m in site['members']]}
 item['extraCalls']=sum(m['calls'] for m in item['members'] if m['method']) if public else 0;price+=item['extraCalls'];heat.append(item)
assert enter['src/compiler/emitter.ts:1215']==enter['src/compiler/emitter.ts:1275']
result={'certified':True,'controls':'controls-result.json','workload':report['counts'],'workerSnapshots':len(workers),'processSnapshots':len(rows),'scope':'whole oracle install/build/tests; separate controls excluded','publicSites':sum(h['public'] for h in heat),'publicWrappers':sum(m['method'] for h in heat if h['public'] for m in h['members']),'extraCallsPerLaneRun':price,'totalCountedMemberCalls':sum(calls.values()),'measurement':meta,'sites':heat}
(root/'heat.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:result[k] for k in ['publicSites','publicWrappers','extraCallsPerLaneRun','totalCountedMemberCalls','workerSnapshots']}))
