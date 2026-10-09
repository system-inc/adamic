"""Read every prior parser evidence file and package measured scout provenance."""
from pathlib import Path
from collections import Counter,defaultdict
import gzip
import hashlib
import json
import shutil
import subprocess

here=Path(__file__).resolve().parent
repo=here.parents[2]
prior=repo/'stage3/drivers/parser/evidence'
index=[]
for p in sorted(prior.rglob('*')):
    if not p.is_file():continue
    raw=p.read_bytes()
    data=gzip.decompress(raw) if p.suffix=='.gz' else raw
    index.append({'file':p.relative_to(repo).as_posix(),'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),
                  'lines':data.count(b'\n')})
(here/'evidence/prior-evidence-index.json').write_text(json.dumps(index,indent=2)+'\n')
print('Read prior evidence files',len(index),'expanded bytes',sum(x['bytes'] for x in index))
closure=json.loads((here/'evidence/value-closure.json').read_text())
constructs=json.loads((here/'evidence/constructs.json').read_text())
files={r['file'] for r in constructs['files']}
summary=[];sites=[]
raw=Path('/tmp/step24-latent.jsonl').read_bytes()
for line in raw.splitlines():
    row=json.loads(line)
    if 'file' not in row:continue
    file=row['file'].split('/step24-adapted/',1)[-1]
    if file not in files:continue
    unique={}
    for f in row['findings']:
        if f['kind'] not in ['NotYet','Refused']:continue
        for k in ['where','text','unit']:
            if k in f:f[k]=f[k].replace('/tmp/step24-adapted/','')
        identity=tuple(f.get(k,'') for k in ['kind','where','reason','text'])
        unique[identity]=f
    reasons=Counter((f['kind'],f['reason']) for f in unique.values())
    summary.append({'file':file,'counts':dict(Counter(f['kind'] for f in unique.values())),
                    'reasons':[{'kind':k,'reason':r,'count':n} for (k,r),n in reasons.most_common()]})
    sites.extend(unique.values())
record={'main':subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip(),
        'measurement':'measured on a checker-rejected program','counts_scope':'whole-file unique latent findings for files containing parser-only reachable declaration spans; not restricted to reached spans',
        'scope_limits':'first lowering failure per attempted unit; diagnosed bodies skipped; no IR or backend output',
        'files':summary,'sites':sites}
(here/'evidence/latent.json').write_text(json.dumps(record,indent=2)+'\n')
with gzip.open(here/'evidence/latent-raw.jsonl.gz','wb') as stream:stream.write(raw)
for source,name in [('/tmp/step24-depth-final.json','depth.json.gz'),('/tmp/step24-setup.log','setup.log.gz'),
 ('/tmp/step24-apply.log','apply.log.gz'),('/tmp/step24-latent-audit.log','latent-audit.log.gz'),
 ('/tmp/step24-latent-run.log','latent-run.log.gz'),('/tmp/step24-parser-final.log','parser-run.log.gz'),
 ('/tmp/step24-parser-final/report.json','parser-report.json.gz')]:
    with gzip.open(here/'evidence'/name,'wb') as stream:stream.write(Path(source).read_bytes())
# Keep the existing oracle mutant results, without full dumps.
shutil.copyfile('/tmp/step24-parser-final/jsdoc-mutants/report.json',here/'evidence/jsdoc-report.json')
print(json.dumps([{'file':r['file'],**r['counts']} for r in summary],indent=2))
