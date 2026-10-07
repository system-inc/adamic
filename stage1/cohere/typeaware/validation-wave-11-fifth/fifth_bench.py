import subprocess,time,pathlib,json,statistics,hashlib,os
root=pathlib.Path('/workspace/wave-11-fifth-validation'); rows=[]
for population,config in [('repository','/workspace/adamic/tsconfig.json'),('compiler','/workspace/wave-11-typescript/src/compiler/tsconfig.json')]:
 truth=max(root.glob('*-'+population+'-go.stdout'),key=lambda p:p.stat().st_mtime).read_bytes()
 for i in range(3):
  for kind,binary in [('go','wave-11-fifth-oracle'),('native','wave-11-fifth')]:
   env=dict(os.environ,ADAMIC_TSGO_TIMING='1')
   start=time.perf_counter();r=subprocess.run([str(root/binary),config,str(root/(population+'.manifest'))],capture_output=True,env=env);elapsed=time.perf_counter()-start
   assert r.returncode==0,(kind,r.stderr)
   assert r.stdout==truth
   rows.append(dict(population=population,implementation=kind,round=i+1,seconds=elapsed,timing=r.stderr.decode().strip(),sha256=hashlib.sha256(r.stdout).hexdigest()))
  print(population,'round',i+1,'Go',rows[-2]['seconds'],'native',rows[-1]['seconds'],flush=True)
pathlib.Path('/workspace/wave-11-logs/fifth-timing.json').write_text(json.dumps(rows,indent=2)+'\n')
for pop in ['repository','compiler']:
 med={kind:statistics.median(x['seconds'] for x in rows if x['population']==pop and x['implementation']==kind) for kind in ['go','native']}
 print(pop,med,'native/go',med['native']/med['go'],flush=True)
