from pathlib import Path
import json,os,re,subprocess
root=Path('/workspace/adamic/review/compiler/c-emission-identity')
manifest=json.loads(Path('/tmp/cemit-manifest.json').read_text())
largest=sorted(manifest,key=lambda r:Path('/tmp/cemit-final-cut',r['Name']+'.c').stat().st_size,reverse=True)[:3]
results=[]
for row in largest:
 single=Path('/tmp/cemit-one.json');single.write_text(json.dumps([row]))
 for trial in range(3):
  for side in (['parent','cut'] if trial%2==0 else ['cut','parent']):
   load_before=Path('/proc/loadavg').read_text().strip()
   log=root/f"time-{row['Name']}-{side}-{trial+1}.log"
   env=dict(os.environ,CEMIT_NATIVE_C='1',GOMAXPROCS='4')
   with log.open('w') as f:
    child=subprocess.run(['/tmp/cemit-'+side+'-compiler-v2',str(single),'/tmp/cemit-timed-'+side],cwd='/workspace/cemit-'+side,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=90)
   if child.returncode:raise RuntimeError(str(log))
   seconds=float(re.search(r'emission=([0-9.]+)',log.read_text())[1])
   record=dict(program=row['Name'],side=side,trial=trial+1,native_C_seconds=seconds,load_before=load_before,load_after=Path('/proc/loadavg').read_text().strip())
   results.append(record);(root/'timings.json').write_text(json.dumps(results,indent=2)+'\n')
   print(record,flush=True)
