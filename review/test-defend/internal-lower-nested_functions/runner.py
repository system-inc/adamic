from pathlib import Path
import subprocess,os,json,time
p=Path('review/test-defend/internal-lower-nested_functions')
menu=[('D1','internal/lower/borrow.go',' && !local.EnvironmentCell',''),('D2','internal/lower/nested_functions.go',None,None)]
for mid,name,old,new in menu:
 f=Path(name);original=f.read_text()
 if mid=='D2':
  start=original.index('\tfor index := range l.result.Functions {',original.index('func (l *lowering) finishNestedEnvironment'))
  end=original.index('\t// Preserve equal layouts',start);old=original[start:end];new=''
 assert original.count(old)==1
 try:
  f.write_text(original.replace(old,new));p.joinpath(mid+'.diff').write_bytes(subprocess.check_output(['git','diff','--',name]))
  with p.joinpath(mid+'-vet.log').open('w') as log:subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,check=True)
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/lower-defense/cache/'+mid;start=time.monotonic()
  with p.joinpath(mid+'.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env=env,stdout=log,stderr=subprocess.STDOUT)
  es=[json.loads(l) for l in p.joinpath(mid+'.log').read_text().splitlines() if l.startswith('{')]
  result={k:sorted({e['Test'] for e in es if e.get('Action')==k and e.get('Test') and '/' not in e['Test']}) for k in ['fail','pass','skip']};result.update(exit=r.returncode,wall_seconds=time.monotonic()-start,file=name,line=original[:original.index(old)].count('\n')+1)
  p.joinpath(mid+'-results.json').write_text(json.dumps(result,indent=2)+'\n');print(mid,json.dumps(result),flush=True)
 finally:f.write_text(original)
