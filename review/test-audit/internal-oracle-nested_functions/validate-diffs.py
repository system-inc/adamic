from pathlib import Path
import json,subprocess,time
p=Path(__file__).resolve().parent
root=Path.cwd();plan=json.loads((p/'replay-plan.json').read_text());scratch=Path('/tmp/u064-standalone');scratch.mkdir(exist_ok=True)
originals={m['file']:subprocess.check_output(['git','show','HEAD:'+m['file']],text=True) for m in plan}
results=json.loads((p/"standalone-builds.json").read_text())
completed={x["id"] for x in results if x["exit"]==0}
for m in plan:
 if m["id"] in completed: continue
 mapping={}
 for i,(file,original) in enumerate(originals.items()):
  path=scratch/(str(i)+'.go');path.write_text(original.replace(m['old'],m['new']) if file==m['file'] else original)
  mapping[str(root/file)]=str(path)
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':mapping}))
 t=time.monotonic()
 with (p/('build-standalone-'+m['id']+'.log')).open('w') as f:
  cmd=['timeout','90','go','test','-overlay',str(overlay),'-c','-o','/tmp/u064-standalone.test','./internal/oracle/'] if m['kind']=='witness' else ['timeout','90','go','build','-overlay',str(overlay),'./'+str(Path(m['file']).parent)]
  r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 result=dict(id=m['id'],exit=r.returncode,seconds=time.monotonic()-t);results.append(result)
 (p/'standalone-builds.json').write_text(json.dumps(results,indent=2));print(result,flush=True)
 if r.returncode: print((p/('build-standalone-'+m['id']+'.log')).read_text(),flush=True)
