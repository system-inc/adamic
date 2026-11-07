import hashlib,json,os,shutil,subprocess
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'));rows=[]
for label in ['before','after']:
 tree=s/('checkpoint6-'+label);shutil.copytree(s/'checkpoint30-before/src',tree/'src')
 if label=='after':
  r=subprocess.run(['node',str(repo/'stage3/adapt/99-correctness-core-levenshtein/adapt.cjs'),str(tree)],env=env,capture_output=True,text=True);assert r.returncode==0,r.stderr;assert json.loads(r.stdout)['edits']==1
 def manifest():return {str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
 frozen=manifest();r=subprocess.run(['node',str(repo/'stage3/adapt/30-indexed-reads/adapt.cjs'),str(tree)],env=env,capture_output=True,text=True);assert r.returncode==0,r.stderr;assert frozen==manifest();rows.append({'side':label,'exit':r.returncode,'stdout':r.stdout,'stderr':r.stderr,'sourceUnchanged':True})
assert {k:v for k,v in rows[0].items() if k!='side'}=={k:v for k,v in rows[1].items() if k!='side'}
(s/'checkpoint30-six.json').write_text(json.dumps({'identical':True,'bothSuccessful':True,'rows':rows},indent=2)+'\n');print('adapter 30 checkpoint replay succeeds identically with all eight authorized sites')
