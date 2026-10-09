import pathlib,subprocess,time,json,os,statistics,re
r=pathlib.Path('/workspace/adamic');o=pathlib.Path('/tmp/u009');names=json.loads((o/'names.json').read_text());ms=json.loads((o/'mutants.json').read_text())
def run(cmd,label,env=None):
 t=time.monotonic()
 with open(o/(label+'.log'),'w') as f:x=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
 z=dict(name=label,command=cmd,seconds=time.monotonic()-t,exit=x.returncode)
 with open(o/'commands.jsonl','a') as f:f.write(json.dumps(z)+'\n')
 return z
# Correct generated switch value selectors, before successful build or matrices.
f=r/'cmd/adamic-meter/main.go';s=f.read_text();s=s.replace('func() string { if os.Getenv("ADAMIC_MUTANT") == "M8"','func() int { if os.Getenv("ADAMIC_MUTANT") == "M8"');f.write_text(s)
f=r/'cmd/adamic-meter/returns.go';s=f.read_text();s=s.replace('func() string { if os.Getenv("ADAMIC_MUTANT") == "M18" { return diagnosticCount(before, "7030") != 0 }; return diagnosticCount(before, "7030") == 0 }()', 'func() bool { if os.Getenv("ADAMIC_MUTANT") == "M18" { return diagnosticCount(before, "7030") != 0 }; return diagnosticCount(before, "7030") == 0 }()');f.write_text(s)
run(['gofmt','-w','cmd/adamic-meter/main.go','cmd/adamic-meter/optional.go','cmd/adamic-meter/returns.go'],'switch-gofmt-corrected')
x=run(['go','test','-c','-o',str(o/'meter.test'),'./cmd/adamic-meter/'],'switch-build-corrected');assert x['exit']==0
(o/'switch.diff').write_text(subprocess.check_output(['git','diff','--','cmd/adamic-meter'],cwd=r,text=True))
for m in ms:
 env=os.environ.copy();env['ADAMIC_MUTANT']=m['id'];x=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-meter/','-run','.'],m['id']+'-matrix',env)
 events=[]
 for line in (o/(m['id']+'-matrix.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 observed={x['Test'] for x in events if x.get('Test') in names and x['Action'] in ['pass','fail','skip']}
 if observed!=set(names):
  for name in names:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-meter/','-run','^'+name+'$'],m['id']+'-'+name,env)
print('matrices complete')
