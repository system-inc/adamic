import json,pathlib,collections
R=pathlib.Path('/workspace/scratch/core-wave')
def read(name):
 rows=[json.loads(x) for x in (R/('census-'+name+'.jsonl')).read_text().splitlines()]
 assert len(rows)==2 and rows[0]['checker_rejected']==False and not rows[0]['diagnostics'],rows[0]
 d=rows[1];sites=set()
 for f in d['findings']:
  if f['kind'] in ['Refused','NotYet']:
   sites.add((f['where'].split('core.ts:')[1],f['kind'],f['phase'],f['reason']))
 return sites,{'units':len(d['units']),'findings':len(d['findings']),'distinctSites':len(sites),'diagnostics':0,'deferredUnits':sum(x['status']=='deferred_generic_instantiation' for x in d['units'])}
base,bm=read('before');summary={'before':bm,'adaptations':{}}
for name,owner in [('clone','clone'),('groupby','groupBy')]:
 after,am=read(name);mutant,mm=read('mutant-'+name);gone=sorted(base-after);added=sorted(after-base)
 record={'after':am,'mutant':mm,'removed':gone,'added':added,'mutantRestoresBaselineSites':mutant==base}
 summary['adaptations'][name]=record
 assert not added,record
 assert len(gone)==1 and gone[0][1:]==('Refused','refusal_scan',"a cast the runtime can't check"),record
 assert mutant==base,record
(R/'census-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
