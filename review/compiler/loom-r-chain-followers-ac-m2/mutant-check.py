from pathlib import Path
import subprocess,json,time
out=Path('review/compiler/loom-r-chain-followers-ac-m2');root=Path.cwd();(out/'mutants').mkdir(exist_ok=True)
rows=[]
cases=[('drop-main-preflight','internal/lower/refusals.go','l.knownLoweringGap(node)','error(nil)','./internal/oracle','^TestMiscompile2AErrorSpread$'),('drop-phantom-refusal','internal/lower/refusals.go','l.phantomRefusal(node)','error(nil)','./internal/lower','^TestPhantomRefusals$'),('drop-callback-specialization','internal/lower/callback_widening.go','l.contextualGenericType(node, l.checker.GetTypeAtLocation(node))','l.checker.GetTypeAtLocation(node)','./internal/lower','^TestGenericFunctionValuesHaveSeparateInstances$')]
for name,file,old,new,pkg,pattern in cases:
 original=(root/file).read_text();assert old in original
 replacement=root/out/'mutants'/(name+'.go.txt');replacement.write_text(original.replace(old,new,1))
 overlay=root/out/'mutants'/(name+'.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(replacement)}}))
 command=['go','test',pkg,'-overlay',str(overlay),'-run',pattern,'-count=1','-timeout','90s','-parallel','4','-json'];start=time.monotonic()
 with (out/('mutant-'+name+'.log')).open('w') as f:r=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,timeout=180)
 events=[];text=(out/('mutant-'+name+'.log')).read_text()
 for line in text.splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if e.get('Action')=='fail' and 'Test' in e:events.append(e['Test'])
 row=dict(name=name,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3),catchers=events);rows.append(row);(out/'mutant-results.json').write_text(json.dumps(rows,indent=2));print(row,flush=True)
 assert r.returncode==1 and events and '[build failed]' not in text and 'panic: test timed out' not in text
