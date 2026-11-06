#!/usr/bin/env python3
"""Checkpoint one successful binary run as an independently pushed piece."""
import hashlib,io,json,pathlib,subprocess,sys,tarfile
p=pathlib.Path(__file__).resolve().parent
stem,label,variant,seconds,verification=sys.argv[1:6]
scratch=pathlib.Path('/workspace/scratch/v3-pieces')/stem
output=scratch/f'{label}-{variant}.jsonl';log=scratch/f'{label}-{variant}.log'
rows=list(map(json.loads,output.read_text().splitlines()))
assert len(rows)>1,'no complete file records'
meta=dict(stem=stem,label=label,variant=variant,elapsed_seconds_upper_bound=float(seconds),exit_verification=verification,
 files=[r['file'].split('-adapted-complete/',1)[1] for r in rows[1:]],output_sha256=hashlib.sha256(output.read_bytes()).hexdigest())
(scratch/f'{label}-{variant}-leg.json').write_text(json.dumps(meta,indent=2)+'\n')
assert meta['elapsed_seconds_upper_bound']<25*60,'piece exceeded 25 minutes'
files={output.name:output.read_bytes(),log.name:log.read_bytes(),'leg.json':json.dumps(meta,indent=2).encode()+b'\n'}
files['SHA256SUMS']=''.join(f'{hashlib.sha256(d).hexdigest()}  {n}\n' for n,d in sorted(files.items())).encode()
legs=p/'legs';legs.mkdir(exist_ok=True)
archive=legs/f'{stem}-{label}-{variant}.tar.gz'
with tarfile.open(archive,'w:gz') as t:
 for n,d in sorted(files.items()):
  info=tarfile.TarInfo(n);info.size=len(d);info.mtime=0;t.addfile(info,io.BytesIO(d))
index=p/'LEG-INDEX.json';data=json.loads(index.read_text()) if index.exists() else {}
meta['archive_sha256']=hashlib.sha256(archive.read_bytes()).hexdigest();meta['archive']=str(archive.relative_to(p))
data[f'{stem}-{label}-{variant}']=meta;index.write_text(json.dumps(data,indent=2)+'\n')
(p/'SHA256SUMS').write_text(''.join(f'{hashlib.sha256(x.read_bytes()).hexdigest()}  {x.relative_to(p)}\n' for x in sorted(p.rglob('*.tar.gz'))))
with (scratch/f'{label}-{variant}-publication.log').open('w') as stream:
 subprocess.run(['git','add',str(archive),str(index),str(p/'SHA256SUMS')],check=True,stdout=stream,stderr=stream,timeout=15)
 subprocess.run(['git','commit','-m',f'census: checkpoint views-v3 {stem} {label} {variant}'],check=True,stdout=stream,stderr=stream,timeout=30)
 subprocess.run(['git','push','origin','HEAD:refs/heads/typescript/views-v3-price'],check=True,stdout=stream,stderr=stream,timeout=60)
print(json.dumps(meta),flush=True)
