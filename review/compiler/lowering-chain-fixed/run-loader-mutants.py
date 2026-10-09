import json, subprocess, time
from pathlib import Path
repo=Path.cwd();root=repo/'review/compiler/lowering-chain-fixed/loader-mutants';root.mkdir(exist_ok=True)
rows=json.loads((root.parent/'loader-mutants.json').read_text())
# Also prove the new mixed-file boundary rejects the old selection behavior.
rows.append(['ordinary-project-old-selection','project_loader.go', 'if config != nil && len(config.ProjectReferences()) == 0 {', 'if config != nil && false {','TestMixedAdamicAndTypeScriptWithoutReferences','mixed'])
for name,file,needle,replacement,test,witness in rows:
 path=repo/'internal/load'/file;source=path.read_text();assert source.count(needle)==1,(name,source.count(needle))
 directory=root/name;directory.mkdir(exist_ok=True)
 changed=directory/(path.name+'.txt');changed.write_text(source.replace(needle,replacement,1))
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(changed)}}))
 start=time.monotonic()
 with (directory/'test.log').open('w') as log:
  result=subprocess.run(['go','test','-p','1','-parallel','4','-timeout','90s','-overlay',str(overlay),'./internal/load','-run','^'+test+'$','-count=1','-v'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 output=(directory/'test.log').read_text()
 if result.returncode==0 or '--- FAIL:' not in output or witness not in output or '[build failed]' in output:raise SystemExit(name+': not caught as expected: '+output[-3000:])
 print(name+': caught %.2fs'%(time.monotonic()-start),flush=True)
