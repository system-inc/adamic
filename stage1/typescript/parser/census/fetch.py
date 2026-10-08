"""Fetch pinned public corpus, without installs, hooks, or repository scripts."""
import pathlib,subprocess,concurrent.futures,json
base=pathlib.Path(__file__).resolve().parent
out=pathlib.Path('/workspace/parser-corpus');out.mkdir(exist_ok=True)
def fetch(line):
 repo,sha=line.split();p=out/(repo.replace('/','__')+'-'+sha[:12])
 try:
  p.mkdir(exist_ok=True)
  for args in [['init','-q'],['fetch','--depth=1','https://github.com/'+repo+'.git',sha],['-c','core.hooksPath=/dev/null','checkout','--detach','-q','FETCH_HEAD']]:
   r=subprocess.run(['git','-C',str(p)]+args,capture_output=True,text=True,timeout=600)
   if r.returncode:raise RuntimeError(r.stderr[-1000:])
  files=sorted(str(f) for f in p.rglob('*') if f.is_file() and f.suffix in ('.ts','.tsx') and '.git' not in f.parts)
  print(repo,len(files),flush=True)
  return {'repo':repo,'sha':sha,'path':str(p),'files':files}
 except Exception as e:
  print(repo,'FAILED',str(e),flush=True);return {'repo':repo,'sha':sha,'error':str(e)}
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:rows=list(pool.map(fetch,(base/'public-pins.tsv').read_text().splitlines()))
(base/'fetch-results.json').write_text(json.dumps(rows,indent=2)+'\n')
(base/'public.manifest').write_text(''.join(f+'\n' for row in rows for f in row.get('files',[])))
