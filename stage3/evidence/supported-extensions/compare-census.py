import collections,json
from pathlib import Path
p=Path(__file__).resolve().parent
roots={'before':'/workspace/scratch/supported-extensions-before/','after':'/workspace/scratch/supported-extensions-candidate/','mutant':'/workspace/scratch/supported-extensions-mutant/'}
def read(label):
 records=[json.loads(line) for line in (p/('census-'+label+'.jsonl')).read_text().splitlines()]
 assert records[0]['latent_mode']=='full' and not records[0]['diagnostics']
 assert records[-1]['status']=='relation_report','incomplete census'
 files=[r for r in records if r.get('file')];assert len(files)==1
 file=files[0];assert file['file'].endswith('/src/compiler/utilities.ts')
 def normalize(s):return s.replace(roots[label],'')
 findings=file['findings']+records[-1].get('findings',[])
 sites={(normalize(f['where']),f['kind'],f['reason']) for f in findings if f['kind'] not in ['Boundary','SkippedDependency']}
 phases={(normalize(f['where']),f['kind'],f['reason'],f['phase']) for f in findings if f['kind'] not in ['Boundary','SkippedDependency']}
 return {'sites':sites,'phases':phases,'units':len(file['units']),'events':len(findings),'statuses':dict(collections.Counter(u['status'] for u in file['units']))}
a,b,m=[read(label) for label in ['before','after','mutant']]
target=('src/compiler/utilities.ts:10005:158','Refused',"a cast the runtime can't check")
assert target in a['sites'];assert target not in b['sites'];assert target in m['sites']
assert not b['sites']-a['sites'],sorted(b['sites']-a['sites'])
assert a['sites']-b['sites']=={target},sorted(a['sites']-b['sites'])
assert m['sites']==a['sites']
result={'removed':sorted(a['sites']-b['sites']),'added':sorted(b['sites']-a['sites']),'restoringMutantMatchesBaseline':True,'phaseRemoved':sorted(a['phases']-b['phases']),'phaseAdded':sorted(b['phases']-a['phases']),'measurements':{label:{k:v for k,v in data.items() if k not in ['sites','phases']} for label,data in zip(['before','after','mutant'],[a,b,m])}}
(p/'census-delta.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
