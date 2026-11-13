"""Detach only the bounded outer runner so a tool reconnect cannot orphan its group."""
import json,os,subprocess,sys
from pathlib import Path
base=Path(__file__).resolve().parent;repo=base.parents[2];mode=os.environ['CENSUS_MODULE_FORMAT']
first,last=map(int,sys.argv[1:]);name=f'group-{first}-{last}'
assert not (base/mode/(name+'-status.json')).exists()
args=[sys.executable,str(base/'run.py'),name,sys.executable,str(base/'group.py'),str(first),str(last)]
with (base/mode/(name+'-outer.log')).open('wb') as log:
 p=subprocess.Popen(args,cwd=repo,stdout=log,stderr=log,start_new_session=True)
(base/mode/(name+'-launch.json')).write_text(json.dumps(dict(pid=p.pid,command=args,detached_outer_runner=True),indent=2)+'\n')
print(name,'bounded outer runner pid',p.pid)
