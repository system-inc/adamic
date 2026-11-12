import io,json,subprocess,sys,tarfile
from pathlib import Path
out=Path(__file__).parent;tree=Path('/workspace/scratch/path-components-after')
cmd=['git','-C',str(tree),'archive','refs/stage3/40-explicit-any-after','src','tests/baselines/reference/api/typescript.d.ts','package.json']
r=subprocess.run(cmd,stdout=subprocess.PIPE);print('git archive child_exit='+str(r.returncode),flush=True)
if r.returncode:sys.exit(r.returncode)
replay=out/'replay40';replay.mkdir()
with tarfile.open(fileobj=io.BytesIO(r.stdout)) as archive:archive.extractall(replay,filter='data')
rule=next(r for r in json.loads(Path('/workspace/adamic/stage3/adapt/40-explicit-any/class-rules.json').read_text())['brands'] if r['id']==103)
assert rule['after'] in (replay/rule['file']).read_text(),'after-40 ref must already contain generic target'
(replay/'node_modules').symlink_to(tree/'node_modules',target_is_directory=True)
