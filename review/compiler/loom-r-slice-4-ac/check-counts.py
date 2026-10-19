import json,os,re,subprocess,time
from pathlib import Path
p=Path('review/compiler/loom-r-slice-4-ac');scratch=Path('/tmp/adamic-slice4-counts');scratch.mkdir(exist_ok=True);env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR=str(scratch/'cache'));rows=[]
recorded=Path('internal/oracle/counts.md').read_text()
for name in ['generator','delegated_generator']:
 path='stage3/fixtures/iteration/'+name+'.a';binary=scratch/name;start=time.monotonic()
 command=['/tmp/adamic-slice4-ac','build',path,'-o',str(binary),'--count']
 with (p/(name+'-count-build.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env=env,timeout=45)
 assert r.returncode==0,(name,r.returncode)
 out=p/(name+'-count-stdout.log');err=p/(name+'-count-stderr.log')
 with out.open('w') as stdout,err.open('w') as stderr:r=subprocess.run([str(binary)],stdout=stdout,stderr=stderr,timeout=15)
 assert r.returncode==0;actual=re.search(r'adamic: counts: allocations (\d+) frees (\d+) retains (\d+) releases (\d+) peak (\d+) regions (\d+)',err.read_text());assert actual
 values=list(map(int,actual.groups()));line=next(l for l in recorded.splitlines() if l.startswith('| '+path+' |'));want=[int(n.strip()) for n in line.split('|')[2:-1]];assert want==values,(path,want,values)
 rows.append(dict(path=path,command=command,result='pass',recorded=want,measured=values,seconds=time.monotonic()-start))
(p/'direct-counts.json').write_text(json.dumps(rows,indent=2)+'\n');print(rows)
