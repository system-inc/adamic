from pathlib import Path
import json,subprocess,time
out=Path(__file__).resolve().parent;root=out.parents[2]
path='internal/oracle/testdata/records_refused/records_prototype_read.a';p=root/path;original=p.read_bytes();header,_,body=original.partition(b'\n');selection=out/'mutant-path.json';selection.write_text(json.dumps([path]));rows=[]
try:
 for name,source in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately wrong reason that is absent from this diagnostic\n'+body)]:
  p.write_bytes(source);start=time.monotonic()
  args=['python3',str(out/'scan.py'),name,str(selection)]
  with (out/(name+'.log')).open('w') as log:result=subprocess.run(args,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=180)
  data=json.loads((out/(name+'.json')).read_text());assert result.returncode==1 and data['rows'][0]['status']=='refused' and data['failure'],data
  rows.append({'mutant':name,'command':args,'exit':result.returncode,'seconds':round(time.monotonic()-start,3),'catcher':'unchanged pinned Gate.aCheck','observed_status':'refused','expected':data['rows'][0]['expected']})
finally:p.write_bytes(original)
assert p.read_bytes()==original
(out/'header-mutants.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows),flush=True)
