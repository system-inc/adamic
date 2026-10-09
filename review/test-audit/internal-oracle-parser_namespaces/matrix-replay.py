import pathlib,json,subprocess,time,os,shlex
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');s=json.loads((p/'scope.json').read_text());prod=s['family']+s['names'][-3:];primary='^('+'|'.join(prod)+')$';family='^TestNativeAgreesWithNode$/^(internal|stage3)$/^(oracle|namespace-parser-stops)$/^(testdata|native-namespace-object-receiver\\.a|native-namespace-class\\.a|native-callable-namespace\\.a)$/^(namespace_method_receiver\\.a|namespace_class_registration\\.a|namespace_callable_properties\\.a|predicate_refusals)$/^oct8_predicates_(p14_overload_optional_chain_container|p27_overload_parameter_rebound)\\.a$';records=[];envbase=os.environ.copy();envbase['ADAMIC_GATE_UNCACHED']='1';envbase.pop('PREDICATE_MISCOMPILE_BASELINE',None)
def run(id,label,regex):
 env=envbase.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u067/cache/'+(id or 'base');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex];t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=id,label=label,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd),exit=r.returncode,wall=time.monotonic()-t));(p/'matrix-commands.json').write_text(json.dumps(records,indent=2));print(label,r.returncode,round(records[-1]['wall'],2),flush=True);return (p/(label+'.log')).read_text()
run('','switched-control',primary)
if records[-1]['exit']!=0:raise SystemExit('red switched control')
for i in range(1,4):
 run('','family-baseline.'+str(i),family)
 if records[-1]['exit']!=0:raise SystemExit('red shared baseline')
for x in json.loads((p/'manifest.json').read_text()):
 id=x['id']
 if id.startswith('P'):
  for n in [s['rows'][0]]+s['names'][-3:]:run(id,id+'.'+n.replace(' ','_'),'^('+'|'.join(s['family'])+')$' if n==s['rows'][0] else '^'+n+'$')
  continue
 if id.startswith('W'):
  for n in s['witnesses']:run(id,id+'.'+n,'^'+n+'$')
  continue
 out=run(id,id+'.primary',primary)
 if any(json.loads(line).get('Output', '').lstrip().startswith('panic:') for line in out.splitlines() if line.startswith('{')) or records[-1]['exit']==124:
  for n in [s['rows'][0]]+s['names'][-3:]:run(id,id+'.'+n.replace(' ','_'),'^('+'|'.join(s['family'])+')$' if n==s['rows'][0] else '^'+n+'$')
 run(id,id+'.family',family)
