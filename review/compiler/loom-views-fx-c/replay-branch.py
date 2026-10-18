import concurrent.futures, hashlib, json, os, pathlib, re, subprocess, sys, time
root=pathlib.Path(__file__).resolve().parents[3]
evidence=root/'review/compiler/loom-views-fx-c'
cli=pathlib.Path('/tmp/loom-views-c')
def run(cmd):
 start=time.monotonic()
 try:
  p=subprocess.run(list(map(str,cmd)),cwd=root,capture_output=True,timeout=90)
  return dict(command=list(map(str,cmd)),exit=p.returncode,stdout=p.stdout.decode(errors='replace'),stderr=p.stderr.decode(errors='replace'),seconds=round(time.monotonic()-start,3))
 except subprocess.TimeoutExpired:
  return dict(command=list(map(str,cmd)),exit='timeout',stdout='',stderr='',seconds=90)
def probe(pair):
 item,source=pair
 dest=evidence/'branch-probes'/(source.stem+'-'+hashlib.sha256(source.read_bytes()).hexdigest()[:12])
 dest.mkdir(parents=True,exist_ok=True)
 node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',source])
 js=run([cli,'js',source]); (dest/'generated.mjs').write_text(js['stdout'])
 jsrun=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',dest/'generated.mjs']) if js['exit']==0 else None
 native=run([cli,'build',source,'-o',dest/'native'])
 nativerun=run([dest/'native']) if native['exit']==0 else None
 def same(x): return x is not None and all(x[k]==node[k] for k in ('exit','stdout','stderr'))
 row=dict(item=item,source=str(source.relative_to(root)),node=node,js_compile={k:v for k,v in js.items() if k!='stdout'},js=jsrun,native_compile=native,native=nativerun,agrees=same(jsrun) and same(nativerun))
 (dest/'results.json').write_text(json.dumps(row,indent=2)+'\n')
 print(item,source.name,'agree' if row['agrees'] else 'DIFF',flush=True)
 return row

pairs=[]
manifest=json.loads((evidence/'baseline-manifest.json').read_text())
for entry in manifest:
 if entry['item'] in [133, 135, 143, 145, 154]: pairs.append((entry['item'],root/entry['source']))
with concurrent.futures.ThreadPoolExecutor(max_workers=1) as pool: rows=list(pool.map(probe,sorted(pairs)))
(evidence/'branch-results.json').write_text(json.dumps(rows,indent=2)+'\n')
