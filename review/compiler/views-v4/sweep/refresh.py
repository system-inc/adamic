import json,pathlib,subprocess,time,concurrent.futures
root=pathlib.Path('/workspace/adamic');out=root/'review/compiler/views-v4/sweep'
old=[json.loads(x) for x in (out/'tests.jsonl').read_text().splitlines()]
jobs=[r for r in old if r['name'].startswith('records-') or r['name'].startswith('gaps-stage1-cohere-json-') or r['name'].startswith('gaps-stage1-cohere-yaml-')]
def run(r):
 t=time.monotonic();name=r['name']
 with (out/('refresh-'+name+'.log')).open('w') as f:code=subprocess.run(['timeout','90',*r['command']],cwd=root,stdout=f,stderr=subprocess.STDOUT).returncode
 return dict(name=name,command=r['command'],status=code,seconds=round(time.monotonic()-t,3),log='refresh-'+name+'.log')
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool,(out/'refresh.jsonl').open('w') as f:
 for r in pool.map(run,jobs):f.write(json.dumps(r)+'\n');f.flush();print(r,flush=True)
