from pathlib import Path
import os,subprocess,time,json
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
records=[]
for mid,args in [('M02',[]),('M07',['build',str(out/'sanitize-witness.a'),'-o','/tmp/u007-sanitize-witness.wasm','--target','wasm32-wasi','--sanitize'])]:
 for selector in ['',mid]:
  env=dict(os.environ,ADAMIC_MUTANT=selector,ADAMIC_NATIVE_SPLIT='u007-'+(selector or 'baseline')+'-witness')
  cmd=['timeout','90','/tmp/u007-adamic']+args
  started=time.monotonic()
  result=subprocess.run(cmd,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
  record=dict(mutant=mid,selector=selector or 'baseline',command='ADAMIC_MUTANT='+repr(selector)+' '+' '.join(cmd),exit=result.returncode,stdout=result.stdout,stderr=result.stderr,seconds=time.monotonic()-started)
  records.append(record)
  (out/'survivor-witnesses.json').write_text(json.dumps(records,indent=2)+'\n')
  print(mid,selector or 'baseline',result.returncode,repr(result.stdout),repr(result.stderr),flush=True)
