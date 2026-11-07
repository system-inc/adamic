import os,signal,subprocess,shutil,json,hashlib
from pathlib import Path
s=Path('/workspace/scratch/core-extend');out=Path('/workspace/adamic/stage3/adapt/99-correctness-core-extend')
rows=[r.strip().split(None,1) for r in subprocess.check_output(['ps','-eo','pid,args'],text=True).splitlines()[1:]]
for pid,args in rows:
 if args=='python3 /workspace/scratch/core-extend/pack-pieces.py':os.kill(int(pid),signal.SIGTERM)
excluded=s/'excluded-fixture-copy-attempt';excluded.mkdir()
for p in [s/'count-summary.json',s/'count-results',s/'count-pieces.log',s/'counted-inputs.json',s/'pack-pieces.log',*s.glob('piece-*')]:
 if p.exists():shutil.move(str(p),str(excluded/p.name))
for p in [*out.glob('count-piece-*'),out/'COUNT-PROGRESS.json']:
 if p.exists():shutil.move(str(p),str(excluded/p.name))
a={str(p.relative_to(s/'before/tests')) for p in (s/'before/tests').rglob('*') if p.is_file()};b={str(p.relative_to(s/'counted/tests')) for p in (s/'counted/tests').rglob('*') if p.is_file()}
(excluded/'missing-fixtures.json').write_text(json.dumps(sorted(a-b),indent=2)+'\n')
(excluded/'REASON.txt').write_text('Excluded attempt: recursive copy ignore_patterns removed nested node_modules and built fixture directories under tests. Two project shards detected the missing fixtures. The owner implementations emit identical JavaScript. Complete tests are restored from the full baseline apply; all 398 units are rerun, and fixtures are included in the immutable input manifest.\n')
shutil.rmtree(s/'counted/tests');shutil.copytree(s/'before/tests',s/'counted/tests',symlinks=True)
a={str(p.relative_to(s/'before/tests')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (s/'before/tests').rglob('*') if p.is_file()};b={str(p.relative_to(s/'counted/tests')):hashlib.sha256(p.read_bytes()).hexdigest() for p in (s/'counted/tests').rglob('*') if p.is_file()};assert a==b
(s/'fixture-copy-verification.json').write_text(json.dumps({'pass':True,'files':len(a),'missingInExcludedAttempt':len(json.loads((excluded/'missing-fixtures.json').read_text()))},indent=2)+'\n');print('restored complete fixture corpus',len(a),flush=True)
