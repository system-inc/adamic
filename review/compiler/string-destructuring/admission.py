import concurrent.futures, json, pathlib, subprocess, sys, time
out = pathlib.Path(__file__).parent
mode, binary = sys.argv[1:]
names = json.loads((out/'admission-corpus.json').read_text())
def probe(name):
    start = time.monotonic()
    try:
        p = subprocess.run([binary, 'c', name], stdout=subprocess.DEVNULL, stderr=subprocess.PIPE, timeout=20)
        return dict(path=name, exit=p.returncode, diagnostic=p.stderr.decode(errors='replace')[:2500], seconds=round(time.monotonic()-start,3))
    except subprocess.TimeoutExpired:
        return dict(path=name, exit='timeout')
rows=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
    for row in pool.map(probe,names):
        rows.append(row)
        if len(rows)%200==0: print(mode,len(rows),'/',len(names),flush=True)
(out/('admission-'+mode+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
print(mode, 'admitted',sum(r['exit']==0 for r in rows),'timeouts',sum(r['exit']=='timeout' for r in rows),flush=True)
if mode=='after':
    before={r['path']:r for r in json.loads((out/'admission-before.json').read_text())}
    result=dict(new=[r['path'] for r in rows if r['exit']==0 and before[r['path']]['exit']!=0],regressions=[r['path'] for r in rows if r['exit']!=0 and before[r['path']]['exit']==0],timeouts=[r['path'] for r in rows if r['exit']=='timeout' or before[r['path']]['exit']=='timeout'])
    (out/'admission-delta.json').write_text(json.dumps(result,indent=2)+'\n')
    print(result,flush=True)
