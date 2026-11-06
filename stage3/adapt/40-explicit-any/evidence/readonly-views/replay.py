import hashlib,json,os,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[5]; tree=Path(sys.argv[1]); out=Path(sys.argv[2]); out.mkdir(parents=True,exist_ok=True)
def snapshot():
    return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in tree.rglob('*') if p.is_file() and '.git' not in p.relative_to(tree).parts}
records=[]
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for adapter in sorted((repo/'stage3/adapt').glob('*/adapt.cjs')):
    before=snapshot()
    with (out/(adapter.parent.name+'.log')).open('w') as log:
        run=subprocess.run(['node',str(adapter),str(tree)],env=env,stdout=log,stderr=subprocess.STDOUT)
    after=snapshot()
    changed=sorted(k for k in before.keys()|after.keys() if before.get(k)!=after.get(k))
    row=dict(adaptation=adapter.parent.name,exit=run.returncode,changed_files=changed);records.append(row)
    (out/'replay.json').write_text(json.dumps(records,indent=2)+'\n')
    print(json.dumps(row),flush=True)
    if run.returncode or changed: sys.exit(1)
