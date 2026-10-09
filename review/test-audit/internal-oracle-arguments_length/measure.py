import pathlib,json,re,subprocess,os,time
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');(p/'logs').mkdir(parents=True,exist_ok=True)
names='TestArgumentsLengthTypeScriptSource TestArgumentsLengthWrongSlotMutant TestGateCacheGeneratedC TestGateCacheToolchains TestGateCacheNodeInputs TestGateCacheAtomicEvidence TestGateCacheUncached TestCallTargetThrowAgreesWithNode TestDirectClosureCallAgreesWithNode TestCheckedCastFlushesOutput TestUncheckableCastAdmission TestCheckedCastRunsNoCatchOrFinally TestCheckedCastFailureContract TestCheckedCastMutants'.split();listing=pathlib.Path('/tmp/u055-list.log').read_text().splitlines();rows=[]
for n in names:
 assert n in listing,n
 for f in pathlib.Path('internal/oracle').glob('*_test.go'):
  m=re.search(r'^func '+n+r'\(',f.read_text(),re.M)
  if m:rows.append(dict(test=n,file=str(f),line=f.read_text()[:m.start()].count('\n')+1));break
for x in ['baseline','list','npm']:(p/'logs'/(x+'.log')).write_bytes(pathlib.Path('/tmp/u055-'+x+'.log').read_bytes())
(p/'rows.json').write_text(json.dumps(rows,indent=2))
def run(cmd,log):
 start=time.monotonic()
 with log.open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
 es=[]
 for s in log.read_text().splitlines():
  if s.startswith('{'):
   try:es.append(json.loads(s))
   except:pass
 return r.returncode,es,time.monotonic()-start
pattern='^('+'|'.join(names)+')$'
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern]
rc,es,wall=run(cmd,p/'logs'/'selected-baseline.log');(p/'selected-baseline.json').write_text(json.dumps(dict(exit=rc,wall_seconds=wall,command=' '.join(cmd)),indent=2));assert rc==0,'red selected baseline'
for row in rows:
 vals=[];walls=[]
 for k in range(3):
  rc,es,w=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row['test']+'$'],p/'logs'/('timing-'+row['test']+'-'+str(k)+'.log'));assert rc==0,row['test'];vals.append(next(e['Elapsed'] for e in reversed(es) if e['Action']=='pass' and not e.get('Test')));walls.append(w)
 row.update(timings=vals,seconds=sorted(vals)[1],wall_seconds=walls);(p/'rows.json').write_text(json.dumps(rows,indent=2));print(row['test'],row['seconds'],flush=True)
family=[]
for k in range(3):
 rc,es,w=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^(TestCallTargetThrowAgreesWithNode|TestDirectClosureCallAgreesWithNode)$'],p/'logs'/('timing-family-'+str(k)+'.log'));assert rc==0;family.append(next(e['Elapsed'] for e in reversed(es) if e['Action']=='pass' and not e.get('Test')))
(p/'family-times.json').write_text(json.dumps(dict(timings=family,seconds=sorted(family)[1])))
cmd=['timeout','120','go','test','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/'coverage.out'),'./internal/oracle/','-run',pattern]
rc,es,w=run(cmd,p/'logs'/'coverage.log');assert rc==0,'coverage'
with (p/'coverage-functions.txt').open('w') as out:subprocess.run(['go','tool','cover','-func='+str(p/'coverage.out')],stdout=out,check=True)
s=(p/'coverage-functions.txt').read_text().splitlines();(p/'reached-functions.txt').write_text('\n'.join(l for l in s if re.search(r'\s([\d.]+)%$',l) and float(re.search(r'\s([\d.]+)%$',l)[1])>0)+'\n')
