from pathlib import Path
import json,subprocess,time
p=Path('review/test-audit/internal-oracle-wasi');scratch=Path('/workspace/u072-tmp/verify');scratch.mkdir(exist_ok=True)
plan=json.loads((p/'plan.json').read_text());files=sorted(set(m['file'] for m in plan)|{'internal/native/emit.go'})
base={f:subprocess.check_output(['git','show','HEAD:'+f],text=True) for f in files}
results=[]
for mid in [m['id'] for m in plan]+['P01']:
 replacements={};values=base.copy()
 if mid=='P01':values['internal/native/emit.go']=values['internal/native/emit.go'].replace('return cProgram(program, -1)','return ""')
 else:
  m=next(m for m in plan if m['id']==mid);values[m['file']]=values[m['file']].replace(m['before'],m['after'])
 for i,(f,text) in enumerate(values.items()):
  out=scratch/f'{mid}-{i}.go';out.write_text(text);replacements[str(Path(f).resolve())]=str(out)
 stub=scratch/'stub.go';stub.write_text('package native\n');replacements[str(Path('internal/native/audit_selector.go').resolve())]=str(stub)
 overlay=scratch/f'{mid}.json';overlay.write_text(json.dumps({'Replace':replacements}))
 with open(p/f'{mid}-verify.log','w') as log:
  apply=subprocess.run(['git','apply','--check','--cached',str(p/(mid+'.diff'))],stdout=log,stderr=subprocess.STDOUT)
  command=['go','vet','-overlay',str(overlay),'./internal/native/'];start=time.monotonic();r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT)
 results.append(dict(id=mid,apply_status=apply.returncode,compile_status=r.returncode,seconds=time.monotonic()-start,command=command))
 (p/'compile-status.json').write_text(json.dumps(results,indent=2))
 if apply.returncode or r.returncode:raise SystemExit(mid+' verification failed')
