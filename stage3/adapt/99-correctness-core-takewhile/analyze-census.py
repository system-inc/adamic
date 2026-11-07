import json,re
from pathlib import Path
root=Path('/workspace/scratch/takewhile');raw=root/'raw';out=Path('/workspace/adamic/stage3/adapt/99-correctness-core-takewhile');pin=json.loads((out/'owners.json').read_text())[0]
texts={side:(root/side/'src/compiler/core.ts').read_text() for side in ['before','after','mutant']}
start=texts['before'].count('\n',0,texts['before'].index(pin['before']))+1
ends={side:start+(pin['after'] if side=='after' else pin['before']).count('\n') for side in texts}
delta=ends['after']-ends['before']
records={};measurements={}
for side in texts:
    rows=[json.loads(line) for line in (raw/(side+'-census.jsonl')).read_text().splitlines()]
    measurements[side]=rows[0];assert rows[0].get('diagnostics')==[],rows[0]
    records[side]=next(r for r in rows if r.get('file','').endswith('/core.ts'))
def key(finding,side):
    where=finding['where'];match=re.search(r'core.ts:(\d+):(\d+)$',where)
    if match:
        line,column=map(int,match.groups())
        if side=='after' and line>ends['after']:line-=delta
        where=f'core.ts:{line}:{column}'
    else: where=where.replace(str(root/side),'TREE')
    return (where,finding['kind'],finding['reason'],finding['phase'])
sets={side:{key(f,side) for f in records[side]['findings'] if f['kind']!='Boundary'} for side in records}
removed=sets['before']-sets['after'];added=sets['after']-sets['before']
target=[f for f in records['before']['findings'] if f['kind']=='Refused' and f['reason']=="a cast the runtime can't check" and re.search(r'core.ts:(\d+):',f['where']) and start<=int(re.search(r'core.ts:(\d+):',f['where'])[1])<=ends['before']]
assert len({key(f,'before') for f in target})==1,target
assert removed=={key(f,'before') for f in target},removed
assert not added,added
assert sets['mutant']==sets['before'],sets['mutant']^sets['before']
units={side:[u for u in record['units'] if u.get('name')=='takeWhile'] for side,record in records.items()}
result={'scope':'complete core.ts, exact LATENT_FULL/LATENT_ONLY_FILE/LATENT_ASSERT_NO_OUTPUT protocol','measurements':measurements,'uniqueSites':{side:len(s) for side,s in sets.items()},'removedSites':[list(k) for k in sorted(removed)],'newSites':[],'lineDelta':delta,'takeWhileUnits':units,'mutantRestoresOriginalFullSiteSet':True,'genericBodyLoweringClaim':False}
(out/'CENSUS.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
