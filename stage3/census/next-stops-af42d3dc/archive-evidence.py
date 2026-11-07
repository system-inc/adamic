import hashlib,json,tarfile
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;core=repo/'stage3/census/core-sites-af42d3dc';scratch=Path('/workspace/scratch/after-wave')
def pack(dest,members):
 paths={name:path for name,path in members};assert len(paths)==len(members)
 manifest=[{'name':name,'bytes':path.stat().st_size,'sha256':hashlib.sha256(path.read_bytes()).hexdigest()} for name,path in sorted(paths.items())]
 archive=dest/'raw-streams.tar.gz'
 with tarfile.open(archive,'w:gz',compresslevel=9) as t:
  for name,path in sorted(paths.items()):t.add(path,arcname=name,recursive=False)
 (dest/'archive-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (dest/'raw-streams.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  raw-streams.tar.gz\n')
 print(dest.name,'archive_bytes',archive.stat().st_size,'members',len(paths))
nextmembers=[(p.name,p) for p in sorted(out.glob('group-*.jsonl'))]
nextmembers += [('logs/'+p.name,p) for p in sorted(scratch.glob('*.log'))]
nextmembers += [('statuses/'+p.name,p) for p in sorted(scratch.glob('*-status.json'))]
nextmembers += [('processes/'+p.name,p) for p in sorted(scratch.glob('*-pid.json'))]
nextmembers += [('tsc/'+p.name,p) for p in sorted((scratch/'artifacts/tsc').glob('*')) if p.is_file()]
nextmembers += [('node-readiness/'+p.relative_to(scratch/'artifacts/node-readiness').as_posix(),p) for p in sorted((scratch/'artifacts/node-readiness').rglob('*')) if p.is_file()]
for overlay in ['latent-overlay','entry-overlay']:
 nextmembers += [(overlay+'/'+p.name,p) for p in sorted((scratch/overlay).iterdir()) if p.is_file()]
if (scratch/'entry-unit-times.jsonl').exists():nextmembers.append(('entry-unit-times.jsonl',scratch/'entry-unit-times.jsonl'))
pack(out,nextmembers)
pack(core,[(n,scratch/n) for n in ['raw.jsonl','unit-times.jsonl','core-census.log','ast-map.json','analysis.log','emitter-map.log','site-comparison.log','core-census-status.json']])
# Raw root streams and copied preparation logs remain in the archive only in the final commit.
for p in list(out.glob('group-*.jsonl'))+list(out.glob('*.log')):p.unlink()
