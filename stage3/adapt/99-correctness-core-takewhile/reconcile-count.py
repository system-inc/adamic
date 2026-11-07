import hashlib,json,tarfile,sys
from collections import Counter
from pathlib import Path
repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-takewhile';summary=json.loads((out/'COUNT.json').read_text());plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text())
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import names_digest
assert summary['complete']
units=[];counts=Counter();classes=Counter();predicates=Counter();calls=0;records=[];diff=b''
for piece in summary['pieces']:
    archive=out/f"count-{piece['piece']:02d}.tar.gz"
    expected=(archive.with_suffix('').with_suffix('.sha256')).read_text().split()[0]
    assert hashlib.sha256(archive.read_bytes()).hexdigest()==expected
    with tarfile.open(archive) as tar:
        prefix=f"count-{piece['piece']:02d}/"
        for line in tar.extractfile(prefix+'SHA256SUMS').read().decode().splitlines():
            digest,name=line.split('  ',1);assert hashlib.sha256(tar.extractfile(prefix+name).read()).hexdigest()==digest,name
        for unitname in piece['units']:
            units.append(unitname);unit=next(u for u in plan['units'] if u['name']==unitname)
            report=json.load(tar.extractfile(prefix+unitname+'/report.json'))
            assert report['status']=='pass' and report['counts']==unit['expect']
            observed=json.load(tar.extractfile(prefix+unitname+'/observations.json'))
            assert names_digest(observed)==unit['names_digest'];records+=observed
            counts.update(report['counts']);diff+=tar.extractfile(prefix+unitname+'/oracle/baseline.diff').read()
        for member in tar.getmembers():
            if member.isfile() and member.name.startswith(prefix+'calls/') and member.name.endswith('.jsonl'):
                for line in tar.extractfile(member):
                    row=json.loads(line);calls+=1;predicates[row['predicateName']]+=1
                    assert row['sameResult'] and row['sameSequence'] and row['stableInput'],row
                    assert row['before']==row['after'] and row['predicateArguments']==row['replayPredicateArguments'],row
                    assert not row['acceptedRangeHasHole'] and row['acceptedRangeHoles']==[],row
                    accepted=sum(bool(answer) for answer in row['answers'])
                    if not row['input'].get('undefined'):
                        assert all(row['input']['slots'][i]['present'] for i in range(accepted)),row
                    category='undefined' if row['input'].get('undefined') else 'empty' if row['input']['length']==0 else 'all accepted' if all(row['answers']) else 'none accepted' if not row['answers'][0] else 'middle rejection'
                    classes[category]+=1
assert units==summary['completedUnits']==[u['name'] for u in plan['units']]
assert counts==Counter(plan['counts'])==Counter(summary['oracleCounts'])
assert calls==summary['calls'] and classes==Counter(summary['classes'])
result={'complete':True,'units':len(units),'calls':calls,'classes':dict(classes),'predicates':dict(predicates),'acceptedRangeHoleCalls':0,'resultDifferences':0,'predicateSequenceDifferences':0,'unstableInputs':0,'oracleCounts':dict(counts),'namesDigest':names_digest(records),'baselineDiffSha256':hashlib.sha256(diff).hexdigest(),'archivesAndEveryMemberVerified':True}
root=Path('/workspace/scratch/takewhile');(root/'raw/count-observations.json').write_text(json.dumps(records)+'\n');(root/'raw/count-baseline.diff').write_bytes(diff)
(out/'COUNT-VERIFIED.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result),flush=True)
