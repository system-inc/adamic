import pathlib,json,re,subprocess,os,time
p=pathlib.Path('review/test-audit/internal-oracle-stack');(p/'logs').mkdir(parents=True,exist_ok=True);names=['TestLongArgumentsLeaveTheStackItsLimit','TestSmallStacksStillPanic'];rows=[];listing=pathlib.Path('/tmp/u069-list.log').read_text().splitlines()
for n in names:
 assert n in listing,n
 f=pathlib.Path('internal/oracle/stack_test.go');s=f.read_text();m=re.search(r'^func '+n+r'\(',s,re.M);assert m;rows.append(dict(test=n,file=str(f),line=s[:m.start()].count('\n')+1))
for x in ['baseline','list','npm']:(p/'logs'/(x+'.log')).write_bytes(pathlib.Path('/tmp/u069-'+x+'.log').read_bytes())
def run(cmd,log,env=None):
 start=time.monotonic()
 with log.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 es=[]
 for s in log.read_text().splitlines():
  try:es.append(json.loads(s))
  except:pass
 return r.returncode,es,time.monotonic()-start
pattern='^('+'|'.join(names)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]
rc,es,w=run(cmd,p/'logs'/'selected-baseline.log');assert rc==0,'red selected baseline';(p/'selected-baseline.json').write_text(json.dumps(dict(command=' '.join(cmd),wall_seconds=w,seconds=next(e['Elapsed'] for e in reversed(es) if e['Action']=='pass' and not e.get('Test')))))
for r in rows:
 vals=[];walls=[]
 for k in range(3):
  rc,es,w=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+r['test']+'$'],p/'logs'/('timing-'+r['test']+'-'+str(k)+'.log'));assert rc==0;vals.append(next(e['Elapsed'] for e in reversed(es) if e['Action']=='pass' and not e.get('Test')));walls.append(w)
 r.update(timings=vals,seconds=sorted(vals)[1],wall_seconds=walls)
(p/'rows.json').write_text(json.dumps(rows,indent=2))
env=dict(os.environ,ADAMIC_GATE_UNCACHED='1');cmd=['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/'coverage.out'),'./internal/oracle/','-run',pattern]
rc,es,w=run(cmd,p/'logs'/'coverage.log',env);assert rc==0
with (p/'coverage-functions.txt').open('w') as out:subprocess.run(['go','tool','cover','-func='+str(p/'coverage.out')],stdout=out,check=True)
s=(p/'coverage-functions.txt').read_text().splitlines();(p/'reached-functions.txt').write_text('\n'.join(l for l in s if re.search(r'\s([\d.]+)%$',l) and float(re.search(r'\s([\d.]+)%$',l)[1])>0)+'\n')
(p/'runtime-reach.txt').write_text('Runtime call chain from emitted stack macro, not Go-covered: native/runtime/stack.c find_stack_limit -> ADAMIC_CHECK_STACK (adamic.h) -> adamic_stack_overflow -> adamic_panic -> flush, write_all, ADAMIC_COUNT_REPORT, _exit. Constructor invokes getrlimit and __builtin_frame_address. Oracle helpers: cacheProbe, lowered, sanitized, sourceIdentity, native.C, native.Build, identity, bounded, executeWith, disagreement, rememberRun.\n')
print(json.dumps(rows,indent=2))
