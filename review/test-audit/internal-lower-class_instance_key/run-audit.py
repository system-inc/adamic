import pathlib,json,subprocess,os,time,re
p=pathlib.Path('review/test-audit/internal-lower-class_instance_key'); menu=json.loads((p/'menu.json').read_text()); rows=json.loads((p/'rows.json').read_text()); originals={m['file']:pathlib.Path(m['file']).read_text() for m in menu}; timings=json.loads((p/"timings.json").read_text()) if (p/"timings.json").exists() else {}
os.environ.update(ADAMIC_CYCLE_LEDGER_ROOT='/tmp/u030/typescript',ADAMIC_CYCLE_LEDGER_OUTPUT='/tmp/u030/cycle-ledger.json',OPTIONAL_WIDENING_CONFIG='/tmp/u030/typescript/src/compiler/tsconfig.json',OPTIONAL_WIDENING_OUTPUT='/tmp/u030/optional-widening.json')
def run(args,name):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as f: result=subprocess.run(args,stdout=f,stderr=subprocess.STDOUT)
 timings[name]={'wall':round(time.monotonic()-start,3),'exit':result.returncode,'command':args}; (p/'timings.json').write_text(json.dumps(timings,indent=2)+'\n'); print(name,timings[name]['wall'],result.returncode,flush=True)
 return result.returncode
try:
 for row in rows:
  for i in range(1,4):
   if row+'-'+str(i) in timings: continue
   run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],row+'-'+str(i))
 for m in menu:
  if m['id']+'-vet' in timings and timings[m['id']+'-vet']['exit']==0: continue
  f=pathlib.Path(m['file']); f.write_text(originals[m['file']].replace(m['old'],m['new'],1)); (p/(m['id']+'.diff')).write_text(subprocess.check_output(['git','diff','--',m['file']],text=True)); code=run(['timeout','120','go','vet','./internal/lower/'],m['id']+'-vet'); f.write_text(originals[m['file']]); assert code==0,m['id']
 # One selector, lazy expressions preserve original short-circuit safety.
 switched=dict(originals)
 for m in menu:
  mid,old,new=m['id'],m['old'],m['new']
  if m['kind']=='drop': replacement='if !auditMutation("'+mid+'") {\n'+old+'}\n'
  elif m['kind']=='probe': replacement=old+'\nif auditMutation("'+mid+'") { '+ ('return nil, nil' if mid=='P1' else 'return 0, false')+' }'
  else:
   prefix,suffix,oldexpr,newexpr='', '', old,new
   if mid=='M05': prefix='fresh := '; oldexpr=old[len(prefix):]; newexpr=new[len(prefix):]
   if mid in ['M15','M17']: prefix='if '; suffix=' {'; oldexpr=old[3:-2]; newexpr=new[3:-2]
   if mid=='M18': prefix='\t\t\t\t\twritten = '; oldexpr=old[len(prefix):]; newexpr=new[len(prefix):]
   if mid=='M19': prefix='\treturn '; suffix='\n'; oldexpr=old[len(prefix):-1]; newexpr=new[len(prefix):-1]
   typ='string' if mid=='M20' else 'bool'
   replacement=prefix+'func() '+typ+' { if auditMutation("'+mid+'") { return '+newexpr+' }; return '+oldexpr+' }()'+suffix
  assert old in switched[m['file']],mid
  switched[m['file']]=switched[m['file']].replace(old,replacement,1)
 for file,source in switched.items(): pathlib.Path(file).write_text(source)
 helper=pathlib.Path('internal/lower/audit_u030_mutant.go'); helper.write_text('package lower\nimport "os"\nfunc auditMutation(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\n')
 (p/'selector.txt').write_text(subprocess.check_output(['git','diff','--','internal/lower'],text=True)+'\n# Supplemental selector helper: internal/lower/audit_u030_mutant.go\n'+helper.read_text())
 assert run(['timeout','120','go','test','-c','-o','/tmp/u030/lower.test','./internal/lower/'],'switch-build')==0
 for m in menu:
  mid=m['id']; os.environ['ADAMIC_MUTANT']=mid; os.environ['ADAMIC_BUILD_CACHE_DIR']='/tmp/u030/cache/'+mid
  if m['kind']=='probe':
   relevant=rows[:]
   if mid=='P1': relevant=[r for r in rows if r not in ['TestClockGenericReturnsT01RejectsNullBeforeBody','TestClockGenericReturnsT01RejectsIndexBeforeBody']]
   else: relevant=['TestClockGenericReturnsT01RejectsNullBeforeBody','TestClockGenericReturnsT01RejectsIndexBeforeBody']
   for row in relevant: run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],mid+'-'+row)
   continue
  code=run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],mid)
  log=(p/(mid+'.log')).read_text()
  if any(json.loads(line).get('Output','').startswith('panic:') for line in log.splitlines() if line.startswith('{')) and 'test timed out' not in log:
   for row in rows: run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],mid+'-'+row)
  elif 'test timed out' in log or code==124:
   run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^('+'|'.join(rows)+')$'],mid+'-bounded')
finally:
 for file,source in originals.items(): pathlib.Path(file).write_text(source)
 helper=pathlib.Path('internal/lower/audit_u030_mutant.go')
 if helper.exists(): helper.unlink()
