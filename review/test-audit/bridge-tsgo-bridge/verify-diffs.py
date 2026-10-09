from pathlib import Path
import subprocess,time,json
root=Path.cwd(); p=root/'review/test-audit/bridge-tsgo-bridge'; plan=json.loads((p/'plan.json').read_text()); commit='b3f83786a0ea00c23d47d774d3b6c98f9dd71336'
originals={entry['file']:subprocess.check_output(['git','show',commit+':'+entry['file']],text=True) for entry in plan}
for file,text in originals.items(): (root/file).write_text(text)
results=[]
try:
 for entry in plan:
  id=entry['id']; patch=p/'diffs'/(id+'.diff'); start=time.monotonic(); check=subprocess.run(['git','apply','--check',str(patch)],capture_output=True,text=True)
  if check.returncode: raise RuntimeError(check.stderr)
  subprocess.run(['git','apply',str(patch)],check=True)
  package='./'+str(Path(entry['file']).parent)+'/'
  command=['timeout','90','go','vet',package]
  with (p/(id+'-vet.log')).open('w') as f: result=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT)
  seconds=time.monotonic()-start; results.append(dict(id=id,applies=True,command=' '.join(command),exit=result.returncode,wall_seconds=seconds)); (p/'diff-validation.json').write_text(json.dumps(results,indent=2)+'\n'); print(id,result.returncode,round(seconds,3),flush=True)
  (root/entry['file']).write_text(originals[entry['file']])
finally:
 for file,text in originals.items(): (root/file).write_text(text)
