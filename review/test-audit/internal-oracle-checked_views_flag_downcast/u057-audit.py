import pathlib,subprocess,json,os,time,difflib,re,shutil
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/internal-oracle-checked_views_flag_downcast';rows=json.loads((p/'requested-rows.json').read_text());witness=[rows[i] for i in [0,1,2,6,7,8,9,10]];production=[t for t in rows if t not in witness];(p/'witness-rows.json').write_text(json.dumps(witness,indent=2));(p/'production-rows.json').write_text(json.dumps(production,indent=2));cov=(p/'functions-coverage.txt').read_text();(p/'reached-functions.txt').write_text('\n'.join(l for l in cov.splitlines() if not re.search(r'\s0\.0%$',l) and not l.startswith('total:'))+'\n')
menu=[dict(id='M1',file='internal/lower/view_lazy.go',old='if demanded {',new='if !demanded {',kind='flip condition',change='invert demand for unsupported checked-view reads'),dict(id='M2',file='internal/lower/interface_cast.go',old='"a checked field alias requiring an optional, accessor, or representation conversion"',new='"a checked field alias requiring an optional, accessor, or representation adaptation"',kind='change constant',change='change canonical optional-read boundary wording in both paths'),dict(id='M3',file='internal/native/view_fields.go',old='e.cache(), property.Of, cString(name), cString(property.View))',new='e.cache(), ir.Boolean, cString(name), cString(property.View))',kind='change option',change='require boolean physical storage for ordinary checked fields'),dict(id='M4',file='internal/javascript/view_fields.go',old='e.values(property.ViewAllowed))',new='"")',kind='change option',change='drop literal restrictions from ordinary checked fields')]
paths=set(m['file'] for m in menu)|{'internal/lower/lower.go','internal/native/emit.go','internal/javascript/javascript.go','internal/oracle/oracle_test.go'};bases={f:(r/f).read_text() for f in paths};helpers=[];runs=[];env=os.environ.copy()
def diff(f,new):return ''.join(difflib.unified_diff(bases[f].splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
for m in menu:
 matches=[x.start() for x in re.finditer(re.escape(m['old']),bases[m['file']])];assert matches;m['lines']=[bases[m['file']][:off].count('\n')+1 for off in matches];assert len(matches)==(2 if m['id']=='M2' else 1);(p/(m['id']+'.diff')).write_text(diff(m['file'],bases[m['file']].replace(m['old'],m['new'])))
(p/'menu.json').write_text(json.dumps(menu,indent=2))
pr=[dict(id='PLower',file='internal/lower/lower.go',signature='func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {',answer='return nil, nil',rows=production),dict(id='PC',file='internal/native/emit.go',signature='func C(program *ir.Program) string {',answer='return ""',rows=[t for t in production if t!=rows[5]]),dict(id='PJS',file='internal/javascript/javascript.go',signature='func JavaScript(program *ir.Program) string {',answer='return ""',rows=[t for t in production if t!=rows[5]])]
for x in pr:
 x['line']=bases[x['file']][:bases[x['file']].index(x['signature'])].count('\n')+1;new=bases[x['file']].replace(x['signature'],x['signature']+'\n\temptyAnswer := true\n\tif emptyAnswer { '+x['answer']+' }\n',1);(p/(x['id']+'.diff')).write_text(diff(x['file'],new))
(p/'probes.json').write_text(json.dumps(pr,indent=2))
wf='internal/oracle/oracle_test.go';sig='func disagreement(oracle run, native run) string {';weakened=bases[wf].replace(sig,sig+'\n\tacceptEverything := true\n\tif acceptEverything { return "" }\n',1);(p/'W1.diff').write_text(diff(wf,weakened));(p/'W1.json').write_text(json.dumps(dict(id='W1',file=wf,line=bases[wf][:bases[wf].index(sig)].count('\n')+1,change='disable agreement comparison',rows=witness),indent=2))
def run(cmd,log,id=None):
 ev=env.copy()
 if id:ev.update(ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u057/cache/'+id)
 start=time.monotonic()
 with (p/log).open('w') as f:q=subprocess.run(cmd,cwd=r,env=ev,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,exit=q.returncode,wall=time.monotonic()-start,environment={k:ev[k] for k in ['ADAMIC_MUTANT','ADAMIC_BUILD_CACHE_DIR'] if k in ev}));(p/'audit-runs.json').write_text(json.dumps(runs,indent=2));print(log,q.returncode,round(runs[-1]['wall'],3),flush=True);return q.returncode
try:
 for m in menu:
  f=m['file'];(r/f).write_text(bases[f].replace(m['old'],m['new']));assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],m['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 for x in pr:
  f=x['file'];new=bases[f].replace(x['signature'],x['signature']+'\n\temptyAnswer := true\n\tif emptyAnswer { '+x['answer']+' }\n',1);(r/f).write_text(new);assert run(['timeout','90','go','vet','./'+str(pathlib.Path(f).parent)+'/'],x['id']+'-vet.log')==0;(r/f).write_text(bases[f])
 (r/wf).write_text(weakened);assert run(['timeout','90','go','vet','./internal/oracle/'],'W1-vet.log')==0
 run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(witness)+')$'],'W1.log','W1');(r/wf).write_text(bases[wf])
 for package in ['lower','native','javascript']:
  f=r/('internal/'+package+'/u057_mutant.go');f.write_text('package '+package+'\nimport "os"\nfunc auditUnitMutant()string{return os.Getenv("ADAMIC_MUTANT")}\nfunc auditUnitBool(id string,b bool)bool{if auditUnitMutant()==id{return !b};return b}\nfunc auditUnitString(id,a,b string)string{if auditUnitMutant()==id{return a};return b}\nfunc auditUnitInt(id string,a,b int)int{if auditUnitMutant()==id{return a};return b}\n');helpers.append(f);(p/(package+'-switch-helper.go.txt')).write_text(f.read_text())
 for f,base in bases.items():
  for m in menu:
   if m['file']!=f:continue
   replacement={'M1':'if auditUnitBool("M1", demanded) {','M2':'auditUnitString("M2", "a checked field alias requiring an optional, accessor, or representation adaptation", "a checked field alias requiring an optional, accessor, or representation conversion")','M3':'e.cache(), auditUnitInt("M3", int(ir.Boolean), int(property.Of)), cString(name), cString(property.View))','M4':'auditUnitString("M4", "", e.values(property.ViewAllowed)))'}[m['id']];base=base.replace(m['old'],replacement)
  for x in pr:
   if x['file']==f:base=base.replace(x['signature'],x['signature']+'\n if auditUnitMutant() == "'+x['id']+'" { '+x['answer']+' }\n',1)
  if f!=wf:(r/f).write_text(base);(p/(f.replace('/','_')+'.switch.txt')).write_text(base)
 assert run(['gofmt','-w']+[f for f in bases if f!=wf]+[str(f.relative_to(r)) for f in helpers],'switch-gofmt.log')==0
 assert run(['timeout','90','go','test','-c','-o','/tmp/u057-test','./internal/oracle/'],'switch-build.log')==0
 for m in menu:
  id=m['id'];run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(rows)+')$'],id+'.log',id)
  es=[]
  for line in (p/(id+'.log')).read_text().splitlines():
   try:es.append(json.loads(line))
   except:pass
  term=[e for e in es if e.get('Action') in ['pass','fail','skip'] and 'Test' in e and '/' not in e['Test']]
  if len(term)!=len(rows):
   for t in rows:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+t+'$'],id+'-alone-'+t+'.log',id)
 for x in pr:
  for t in x['rows']:run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+t+'$'],x['id']+'-'+t+'.log',x['id'])
finally:
 for f,s in bases.items():(r/f).write_text(s)
 for f in helpers:
  if f.exists():f.unlink()
run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(rows)+')$'],'final-baseline.log')
run(['go','vet','./internal/oracle/'],'final-vet.log')
run(['git','diff','--check'],'source-diff-check.log')
