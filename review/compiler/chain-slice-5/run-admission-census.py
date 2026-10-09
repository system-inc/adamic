import concurrent.futures,json,pathlib,subprocess,time
root=pathlib.Path.cwd();out=root/'review/compiler/chain-slice-5'
paths=json.loads((out/'admission-paths.json').read_text())
def observe(path):
 row={'path':path}
 for side,binary in [('main','/tmp/chain-slice-5-main-adamic'),('slice','/tmp/chain-slice-5-adamic')]:
  start=time.monotonic()
  with (out/'logs'/('admission-'+side+'.log')).open('a') as log:
   try:
    p=subprocess.run([binary,'c',path],stdout=subprocess.DEVNULL,stderr=subprocess.PIPE,text=True,timeout=20)
    row[side]={'exit':p.returncode,'diagnostic':p.stderr,'seconds':round(time.monotonic()-start,3)}
   except subprocess.TimeoutExpired:row[side]={'exit':124,'diagnostic':'census timeout'}
   log.write(json.dumps({'path':path,**row[side]})+'\n')
 return row
results=[]
with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:
 for i,row in enumerate(pool.map(observe,paths)):
  results.append(row)
  if (i+1)%100==0:print('Observed',i+1,'programs',flush=True)
(out/'admission-results.json').write_text(json.dumps(results,indent=2)+'\n')
new=[r['path'] for r in results if r['main']['exit']!=0 and r['slice']['exit']==0]
regressed=[r for r in results if r['main']['exit']==0 and r['slice']['exit']!=0]
timeouts=[r['path'] for r in results if any(r[s]['exit']==124 for s in ['main','slice'])]
(out/'admission-delta.json').write_text(json.dumps({'scope':'Every recorded counts fixture and every added member .a file; finite corpus, not a universal proof over arbitrary programs','programs':len(results),'newly_admitted':new,'regressions':regressed,'timeouts':timeouts},indent=2)+'\n')
print('Newly admitted',len(new),'regressions',len(regressed),'timeouts',len(timeouts),flush=True)
assert not timeouts
