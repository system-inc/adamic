import pathlib,subprocess,json,concurrent.futures,os
root=pathlib.Path('/workspace/adamic');out=root/'review/compiler/chain-slice-8/admission-final'
rows=[r for r in json.loads((out/'census.json').read_text()) if r['newly_admitted']]
def run(cmd):
 try:
  p=subprocess.run(cmd,cwd=root,capture_output=True,timeout=30)
  return dict(exit=p.returncode,stdout=p.stdout.decode(errors='replace'),stderr=p.stderr.decode(errors='replace'))
 except subprocess.TimeoutExpired:return dict(timeout=True)
def check(pair):
 i,r=pair;p=pathlib.Path(r['absolute_path']);d=out/f'observed-{i:02}';d.mkdir(exist_ok=True)
 (d/'source.a.txt').write_bytes(p.read_bytes())
 node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(p)])
 js=run(['/tmp/chain-slice-8-checked','js',str(p)])
 (d/'program.mjs').write_text(js.get('stdout',''))
 javascript=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(d/'program.mjs')]) if js.get('exit')==0 else js
 binary=pathlib.Path('/tmp/adamic-gate')/f'chain-slice-8-observed-{i}'
 build=run(['/tmp/chain-slice-8-checked','build',str(p),'-o',str(binary),'--sanitize'])
 native=run([str(binary)]) if build.get('exit')==0 else build
 row=dict(path=r['path'],Node=node,JavaScript=javascript,native=native,build=build)
 row['agrees']=all(o.get('exit')==node.get('exit') and o.get('stdout')==node.get('stdout') for o in [javascript,native])
 (d/'outputs.json').write_text(json.dumps(row,indent=2)+'\n')
 return row
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
 for row in pool.map(check,enumerate(rows)):
  results.append(row);(out/'observations.json').write_text(json.dumps(results,indent=2)+'\n')
  print(row['path'],row['agrees'],flush=True)
