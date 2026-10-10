import json,os,subprocess,time
from pathlib import Path
root=Path(__file__).resolve().parents[3]; ev=Path(__file__).resolve().parent
rows=[]
for name,p,b,a,pkg,test in json.loads((ev/'mutant-catalog.json').read_text()):
 d=ev/'mutants'/name;d.mkdir(parents=True,exist_ok=True);original=root/p;s=original.read_text();assert s.count(b)==1,(name,s.count(b));side=d/(original.name+'.txt');side.write_text(s.replace(b,a,1));overlay=d/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(original):str(side)}}))
 command=['timeout','90','go','test','-overlay',str(overlay),pkg,'-run',test,'-count=1','-timeout','80s','-v'];start=time.monotonic()
 with (d/'test.log').open('w') as log:r=subprocess.run(command,cwd=root,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','GOMAXPROCS':'4'},stdout=log,stderr=subprocess.STDOUT)
 text=(d/'test.log').read_text();caught=r.returncode==1 and '--- FAIL: Test' in text and '[build failed]' not in text and 'clang failed' not in text and 'panic: test timed out' not in text
 row=dict(name=name,exit=r.returncode,caught=caught,seconds=time.monotonic()-start,command=command);rows.append(row);print(name,caught,flush=True);(ev/'mutants-results.json').write_text(json.dumps(rows,indent=2)+'\n')
 if not caught:raise RuntimeError(name+' did not fail behaviorally')
