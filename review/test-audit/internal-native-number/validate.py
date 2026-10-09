from pathlib import Path
import json,subprocess,time
p=Path(__file__).resolve().parent;root=Path.cwd();plan=json.loads((p/'replay-plan.json').read_text());tmp=Path('/tmp/u051-standalone');tmp.mkdir(exist_ok=True);out=json.loads((p/"validations.json").read_text());done={x["id"] for x in out if all(e==0 for e in x["exits"])}
original=subprocess.check_output(['git','show','HEAD:internal/buildcache/buildcache.go'],text=True)
for m in plan:
 if m["id"] in done:continue
 r=subprocess.run(['git','apply','--check','--cached',str(p/'diffs'/(m['id']+'.diff'))],capture_output=True,text=True);assert r.returncode==0,(m['id'],r.stderr)
 text=subprocess.check_output(['git','show','HEAD:'+m['file']],text=True).replace(m['old'],m['new'],1);file=tmp/Path(m['file']).name;file.write_text(text)
 overlay=tmp/'overlay.json';cache_file=tmp/'original-buildcache.go';cache_file.write_text(text if m['file'].endswith('.go') else original)
 overlay.write_text(json.dumps({'Replace':{str(root/'internal/buildcache/buildcache.go'):str(cache_file)}}))
 package='internal/buildcache' if m['file'].endswith('.go') else 'internal/native'
 commands=[['timeout','90','go','vet','-overlay',str(overlay),'./'+package+'/']]
 if m['file'].endswith('.c'):
  commands.append(['timeout','90','clang','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-function','-Wno-unused-parameter','-Wno-unused-variable','-Wno-unused-but-set-variable','-I','internal/native/runtime','-fsyntax-only',str(file)])
 t=time.monotonic();exits=[]
 with (p/('validate-'+m['id']+'.log')).open('w') as f:
  for cmd in commands:
   f.write('COMMAND '+json.dumps(cmd)+'\n');f.flush();rr=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT);exits.append(rr.returncode)
 result=dict(id=m['id'],commands=commands,exits=exits,seconds=time.monotonic()-t);out.append(result);(p/'validations.json').write_text(json.dumps(out,indent=2));print(result['id'],exits,round(result['seconds'],3),flush=True)
