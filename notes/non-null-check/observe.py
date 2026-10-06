import concurrent.futures, json, subprocess, os, sys
from pathlib import Path
repository=Path(__file__).resolve().parents[2]
os.chdir(repository)
base=Path(sys.argv[1] if len(sys.argv)>1 else '/tmp/non-null-check-replay').resolve()
base.mkdir(parents=True,exist_ok=True)
def command(args):
    p=subprocess.run(args,capture_output=True,text=True)
    return dict(command=args,exit=p.returncode,stdout=p.stdout,stderr=p.stderr)
def run(path):
    name=path.stem
    node=command(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(path)])
    build=command(['go','run','./cmd/adamic','build',str(path),'-o',str(base/name)])
    native=command([str(base/name)]) if build['exit']==0 else None
    js=command(['go','run','./cmd/adamic','js',str(path)])
    backend=None
    if js['exit']==0:
        emitted=base/(name+'.mjs')
        emitted.write_text(js['stdout'])
        backend=command(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(emitted)])
    if js['exit']==0: js['stdout']='[generated JavaScript saved in '+str(emitted)+']'
    result=dict(program=str(path.relative_to(repository)),source=path.read_text(),node=node,build=build,native=native,javascript=js,backend=backend)
    (base/(name+'.json')).write_text(json.dumps(result,indent=2)+'\n')
    same=native is not None and backend is not None and all((x['exit'],x['stdout'],x['stderr'])==(node['exit'],node['stdout'],node['stderr']) for x in [native,backend])
    print(name, 'AGREES' if same else 'DIFF' if native else 'BUILD FAILED',flush=True)
    if native is None: print(build['stderr'],flush=True)
    return result
paths=sorted(list((repository/'internal/oracle/testdata').glob('coverage_non_null_*.a'))+list((repository/'notes/non-null-check').glob('*.a')))
if len(sys.argv)>2: paths=[p for p in paths if p.stem.startswith(sys.argv[2])]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
    results=list(pool.map(run,paths))
(base/'observations.json').write_text(json.dumps(results,indent=2)+'\n')
