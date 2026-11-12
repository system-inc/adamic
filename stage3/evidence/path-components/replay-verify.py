import io,json,subprocess,tarfile
from pathlib import Path
out=Path(__file__).parent;replay=out/'replay40'
cmd=['git','-C','/workspace/scratch/path-components-after','archive','refs/stage3/40-explicit-any-after','src','tests/baselines/reference/api/typescript.d.ts','package.json']
r=subprocess.run(cmd,stdout=subprocess.PIPE);print('git archive child_exit='+str(r.returncode));assert r.returncode==0
changed=[];count=0
with tarfile.open(fileobj=io.BytesIO(r.stdout)) as archive:
 for entry in archive.getmembers():
  if entry.isfile():
   count+=1
   if archive.extractfile(entry).read()!=(replay/entry.name).read_bytes():changed.append(entry.name)
assert not changed,changed
(out/'replay40.json').write_text(json.dumps(dict(filesCompared=count,edits=0,reference='refs/stage3/40-explicit-any-after'),indent=2)+'\n');print('replay edits=0')
