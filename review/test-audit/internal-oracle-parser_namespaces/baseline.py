import pathlib,json,subprocess,time,os,shlex
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');members=['TestParserNamespaceReceiver','TestParserNamespaceClass','TestParserCallableNamespace'];witnesses=['TestParserCallableNamespaceMutant','TestParserNamespaceClassRegistrationMutant','TestParserNamespaceReceiverMutant'];names=members+witnesses+['TestPredicateDirectionCountsAreRecorded','TestPredicateMiscompileRefusals','TestInUnionNarrowingIsRefused'];rows=['TestParserNamespace family']+names[3:];(p/'scope.json').write_text(json.dumps(dict(names=names,rows=rows,family=members,witnesses=witnesses),indent=2));env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env.pop('PREDICATE_MISCOMPILE_BASELINE',None);records=[]
def run(label,cmd):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(label=label,command='ADAMIC_GATE_UNCACHED=1 '+shlex.join(cmd),exit=r.returncode,wall=time.monotonic()-t));(p/'baseline-commands.json').write_text(json.dumps(records,indent=2));print(label,r.returncode,round(records[-1]['wall'],2),flush=True);return r.returncode
if run('bounded-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(names)+')$']):raise SystemExit('RED baseline')
for n in rows:
 regex='^('+'|'.join(members)+')$' if n==rows[0] else '^'+n+'$'
 for i in range(1,4):
  if run(n.replace(' ','_')+'.'+str(i),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex]):raise SystemExit('RED row')
run('coverage',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower','-coverprofile='+str(p/'coverage.out'),'./internal/oracle/','-run','^('+'|'.join(names)+')$'])
