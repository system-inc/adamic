from pathlib import Path
import subprocess,os,json,time
p=Path('review/test-defend/internal-lower-class_inheritance');f=Path('internal/lower/class_generic_calls.go')
menu=[('D1','\tif err := l.nominalTypeArguments(declaration, targets, mapper, node); err != nil {\n\t\treturn nil, true, err\n\t}\n'),('D2','\tl.typeMapper, l.genericInstances = mapper, bucket\n')]
for mid,old in menu:
 original=f.read_text();assert original.count(old)==1
 try:
  f.write_text(original.replace(old,''));(p/(mid+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',str(f)]))
  with (p/(mid+'-vet.log')).open('w') as log:subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/inheritance-defense/cache/'+mid;start=time.monotonic()
  with (p/(mid+'.log')).open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[json.loads(l) for l in (p/(mid+'.log')).read_text().splitlines() if l.startswith('{')];d={a:sorted({e['Test'] for e in events if e.get('Action')==a and e.get('Test') and '/' not in e['Test']}) for a in ['pass','fail','skip']};d.update(exit=r.returncode,wall_seconds=time.monotonic()-start,file=str(f),line=original[:original.index(old)].count('\n')+1);(p/(mid+'-results.json')).write_text(json.dumps(d,indent=2)+'\n');print(mid,d['fail'],d['wall_seconds'],flush=True)
 finally:f.write_text(original)
