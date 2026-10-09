import pathlib,json,subprocess,time
p=pathlib.Path('review/test-audit/internal-oracle-private_generic_mutant');rows=json.loads((p/'scope.json').read_text());pattern=(p/'scope-pattern.txt').read_text()
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/oracle,./internal/lower,./internal/native,./internal/javascript','-coverprofile',str(p/'reached.cover'),'./internal/oracle/','-run',pattern]
with (p/'coverage.log').open('w') as log:subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
with (p/'functions.txt').open('w') as log:subprocess.run(['go','tool','cover','-func',str(p/'reached.cover')],stdout=log,stderr=subprocess.STDOUT)
for row in rows:
 times=[]
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row['test']+'$'];start=time.monotonic()
  file=p/(row['test']+f'-time{i+1}.log')
  with file.open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT)
  ev=[json.loads(l) for l in file.read_text().splitlines() if l.startswith('{')];last=next((e for e in reversed(ev) if e.get('Action') in ['pass','fail'] and 'Test' not in e),{})
  times.append({'exit':r.returncode,'seconds':last.get('Elapsed'),'wall':time.monotonic()-start,'command':' '.join(cmd)})
 (p/(row['test']+'-times.json')).write_text(json.dumps(times,indent=2))
