import json
from pathlib import Path
base=Path(__file__).resolve().parent
prefixes=['/workspace/scratch/sourcefile-like-position-main-tree-2/','/workspace/scratch/sourcefile-like-position-trial-tree/']
def norm(value):
    if isinstance(value,str):
        for prefix in prefixes: value=value.replace(prefix,'<tree>/')
        return value
    if isinstance(value,list): return [norm(x) for x in value]
    if isinstance(value,dict): return {k:norm(v) for k,v in value.items()}
    return value
def read(side,name):
    records=[norm(json.loads(s)) for s in (base/f'{side}-{name}.jsonl').read_text().splitlines()]
    files=[r for r in records if 'file' in r and 'units' in r]
    assert len(files)==1
    return files[0]
def keys(rows,fields):
    return {json.dumps({k:r[k] for k in fields if k in r},sort_keys=True):r for r in rows}
report={}
for name in ['scanner','types']:
    a,b=[read(side,name) for side in ['main','trial']]
    assert a['file']==b['file']
    fields=['kind','where','reason','text']
    ak,bk=keys(a['findings'],fields),keys(b['findings'],fields)
    allfields=sorted(set().union(*(r.keys() for r in a['findings']+b['findings']))-{'measurement'})
    aa,bb=keys(a['findings'],allfields),keys(b['findings'],allfields)
    report[name]={'baseline_findings':len(a['findings']),'candidate_findings':len(b['findings']),
        'new':[bk[k] for k in sorted(bk.keys()-ak.keys())],
        'removed':[ak[k] for k in sorted(ak.keys()-bk.keys())],
        'new_or_changed_attempts':[bb[k] for k in sorted(bb.keys()-aa.keys())],
        'removed_or_changed_attempts':[aa[k] for k in sorted(aa.keys()-bb.keys())]}
(base/'finding-diff.json').write_text(json.dumps(report,indent=2)+'\n')
for name,d in report.items():
    print(name,'baseline',d['baseline_findings'],'candidate',d['candidate_findings'],'new',len(d['new']),'removed',len(d['removed']))
    for kind in ['new','removed']:
        for r in d[kind]: print(kind,r['where'],r['kind'],r['reason'])
