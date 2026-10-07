from pathlib import Path
import subprocess,json,hashlib
root=Path('/workspace/lint-cost-baseline'); records={}
commands=[('Go',[str(root/'oracle')]),('today',[str(root/'scanner')]),('kind_table',['/workspace/lint-cost-prototype/scanner']),('all_listeners_table',['/workspace/lint-cost-all-listeners/scanner']),('Node_prototype',['node','--disable-warning=ExperimentalWarning','/workspace/adamic/oracle/node.mjs','/workspace/lint-cost-prototype/main.ts'])]
for label,cmd in commands:
 out=root/(label+'-parity.stdout');err=root/(label+'-parity.stderr')
 with out.open('wb') as o,err.open('wb') as e:r=subprocess.run(cmd+['--manifest',str(root/'compiler.txt')],stdout=o,stderr=e)
 data=out.read_bytes();assert r.returncode==0 and not err.read_bytes(),(label,r.returncode,err.read_bytes()[:1000])
 if label=='Go':want=data
 else:assert data==want,(label,'byte mismatch')
 records[label]={'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest(),'exit':r.returncode,'stderr_bytes':err.stat().st_size};print(label,records[label],flush=True)
Path('/workspace/lint-cost-parity.json').write_text(json.dumps(records,indent=2)+'\n')
