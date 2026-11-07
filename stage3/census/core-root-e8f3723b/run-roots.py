import json, os, subprocess, sys
from pathlib import Path
base=Path('/tmp/core-root-avqasqt')
tree=Path('/workspace/core-root-e8f3723b-adapted')
results=[]
for name in ['core','binder']:
    root=tree/'src/compiler'/(name+'.ts')
    raw=base/(name+'.jsonl')
    child=subprocess.run([sys.executable,str(base/'supervise.py'),name,str(base/'census'),str(root),str(raw)])
    print(name+' wrapper_exit='+str(child.returncode),flush=True)
    if child.returncode:
        sys.exit(child.returncode)
    rows=[json.loads(line) for line in raw.read_text().splitlines()]
    assert len(rows)==2,rows
    assert {k:v for k,v in rows[0].items() if k!='seconds'}=={k:v for k,v in rows[1].items() if k!='seconds'}
    results.append(rows[0])
(base/'results.json').write_text(json.dumps(results,indent=2)+'\n')
