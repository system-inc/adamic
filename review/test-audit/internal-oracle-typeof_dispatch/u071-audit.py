import pathlib,subprocess,json,os,time,difflib,re
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-oracle-typeof_dispatch';tests=json.loads((p/'requested-tests.json').read_text());groups=json.loads((p/'row-members.json').read_text());witness=tests[:6]+tests[-2:];production=['TestRequiredViewFieldOperandOnce','TestNarrowedFieldUsesSharedReadiness','TestViewFieldInheritedStaticReadiness','TestDefaultTaggedSourceViews'];(p/'witness-tests.json').write_text(json.dumps(witness,indent=2));(p/'production-tests.json').write_text(json.dumps(production,indent=2))
menu=[dict(id='M1',file='internal/lower/non_null.go',old='" is null or undefined"',new='" is absent"',kind='change constant',change='change non-null assertion diagnostic suffix'),dict(id='M2',file='internal/native/view_fields.go',old='cString(property.Name), e.cache(), property.Of, cString(name), cString(property.View))',new='cString(property.View), e.cache(), property.Of, cString(name), cString(property.Name))',kind='swap arguments',change='swap checked-field lookup name and diagnostic name'),dict(id='M3',file='internal/javascript/view_fields.go',old='e.values(property.ViewAllowed))',new='"")',kind='change option',change='drop ordinary checked-field literal restrictions'),dict(id='M4',file='internal/native/emit_branches.go',old='present = "false"',new='present = "true"',kind='change constant',change='treat literal undefined/null as present in coalescing')]
paths=set(m['file'] for m in menu)|{'internal/lower/lower.go','internal/native/emit.go','internal/javascript/javascript.go','internal/oracle/oracle_test.go','internal/oracle/wasi_shards_test.go'};bases={f:(r/f).read_text() for f in paths};helpers=[];runs=[];env=os.environ.copy()
def diff(f,new):return ''.join(difflib.unified_diff(bases[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
for m in menu:
 matches=[x.start() for x in re.finditer(re.escape(m['old']),bases[m['file']])];assert len(matches)==1;(p/(m['id']+'.diff')).write_text(diff(m['file'],bases[m['file']].replace(m['old'],m['new'])));m['lines']=[bases[m['file']][:off].count('\n')+1 for off in matches]
(p/'menu.json').write_text(json.dumps(menu,indent=2))
pr=[dict(id='PLower',file='internal/lower/lower.go',signature='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {',answer='return nil, nil',tests=production[1:]),dict(id='PC',file='internal/native/emit.go',signature='func C(program *ir.Program) string {',answer='return ""',tests=production),dict(id='PJS',file='internal/javascript/javascript.go',signature='func JavaScript(program *ir.Program) string {',answer='return ""',tests=production),dict(id='PShards',file='internal/oracle/wasi_shards_test.go',signature='func wasiFixtureShards() [][]int {',answer='return nil',tests=['TestWASIShardUnion'])]
for x in pr:
 x['line']=bases[x['file']][:bases[x['file']].index(x['signature'])].count('\n')+1;new=bases[x['file']].replace(x['signature'],x['signature']+'\n\temptyAnswer := true\n\tif emptyAnswer { '+x['answer']+' }\n',1);(p/(x['id']+'.diff')).write_text(diff(x['file'],new))
(p/'probes.json').write_text(json.dumps(pr,indent=2))
wf='internal/oracle/oracle_test.go';sig='func disagreement(oracle run, native run) string {';weakened=bases[wf].replace(sig,sig+'\n\tacceptEverything := true\n\tif acceptEverything { return "" }\n',1);(p/'W1.diff').write_text(diff(wf,weakened))
sf='internal/oracle/wasi_shards_test.go';statement='\tfor ordinal, row := range rows {\n\t\tshards[ordinal%len(shards)] = append(shards[ordinal%len(shards)], row)\n\t}\n';assert statement in bases[sf];setup=bases[sf].replace(statement,'',1);(p/'S1.diff').write_text(diff(sf,setup));(p/'special-edits.json').write_text(json.dumps([dict(id='W1',file=wf,line=bases[wf][:bases[wf].index(sig)].count('\n')+1,change='return agreement without comparing',tests=witness),dict(id='S1',file=sf,line=bases[sf][:bases[sf].index(statement)].count('\n')+1,change='drop entire fixture distribution loop; avoids unused ordinal/row variables',tests=['TestWASIShardUnion'])],indent=2))
def run(cmd,log,id=None):
 ev=env.copy()
 if id:ev.update(ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u071/cache/'+id)
 start=time.monotonic()
 with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,env=ev,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start,environment={k:ev[k] for k in ['ADAMIC_MUTANT','ADAMIC_BUILD_CACHE_DIR','ADAMIC_ORACLE_WASI','WASI_SYSROOT'] if k in ev}));(p/'audit-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True);return q.returncode
def test(ts,log,id):return run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(ts)+')$'],log,id)
try:
 for m in menu:
  f=m['file'];(r/f).write_text(bases[f].replace(m['old'],m['new']));assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],m['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 for x in pr:
  f=x['file'];(r/f).write_text(bases[f].replace(x['signature'],x['signature']+'\n\temptyAnswer := true\n\tif emptyAnswer { '+x['answer']+' }\n',1));assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],x['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 (r/wf).write_text(weakened);assert run(['timeout','90','go','vet','./internal/oracle/'],'W1-vet.log')==0;test(witness,'W1.log','W1');(r/wf).write_text(bases[wf])
 (r/sf).write_text(setup);assert run(['timeout','90','go','vet','./internal/oracle/'],'S1-vet.log')==0;test(['TestWASIShardUnion'],'S1.log','S1');(r/sf).write_text(bases[sf])
 for package in ['lower','native','javascript','oracle']:
  f=r/('internal/'+package+'/u071_mutant'+('_test' if package=='oracle' else '')+'.go');f.write_text('package '+package+'\nimport "os"\nfunc auditUnitMutant()string{return os.Getenv("ADAMIC_MUTANT")}\nfunc auditUnitString(id,a,b string)string{if auditUnitMutant()==id{return a};return b}\n');helpers.append(f);(p/(package+'-switch-helper.go.txt')).write_text(f.read_text())
 for f,base in bases.items():
  for m in menu:
   if m['file']!=f:continue
   replacement={'M1':'auditUnitString("M1", " is absent", " is null or undefined")','M2':'cString(auditUnitString("M2", property.View, property.Name)), e.cache(), property.Of, cString(name), cString(auditUnitString("M2", property.Name, property.View)))','M3':'auditUnitString("M3", "", e.values(property.ViewAllowed)))','M4':'present = auditUnitString("M4", "true", "false")'}[m['id']];base=base.replace(m['old'],replacement)
  for x in pr:
   if x['file']==f:base=base.replace(x['signature'],x['signature']+'\n if auditUnitMutant() == "'+x['id']+'" { '+x['answer']+' }\n',1)
  (r/f).write_text(base);(p/(f.replace('/','_')+'.switch.txt')).write_text(base)
 assert run(['gofmt','-w']+list(bases)+[str(f.relative_to(r)) for f in helpers],'switch-gofmt.log')==0
 assert run(['timeout','90','go','test','-c','-o','/tmp/u071-test','./internal/oracle/'],'switch-build.log')==0
 for m in menu:
  id=m['id'];test(tests,id+'.log',id);es=[]
  for line in (p/(id+'.log')).read_text().splitlines():
   try:es.append(json.loads(line))
   except:pass
  term={e['Test'] for e in es if e.get('Action') in ['pass','fail','skip'] and e.get('Test') in tests}
  if term!=set(tests):
   for t in tests:test([t],id+'-alone-'+t+'.log',id)
 for x in pr:
  for t in x['tests']:test([t],x['id']+'-'+t+'.log',x['id'])
finally:
 for f,s in bases.items():(r/f).write_text(s)
 for f in helpers:
  if f.exists():f.unlink()
test(tests,'final-baseline.log',None);run(['go','vet','./internal/oracle/'],'final-vet.log');run(['git','diff','--check'],'source-diff-check.log')
