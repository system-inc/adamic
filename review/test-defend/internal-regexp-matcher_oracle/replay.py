import pathlib,subprocess,time,json,os
root=pathlib.Path('/workspace/adamic'); p=root/'review/test-defend/internal-regexp-matcher_oracle'
mutants=[('D1','internal/regexp/properties_standin.go','return PropertySet{}, &UnavailablePropertyError{property, unicode.Version}','return PropertySet{}, nil','provider: drop unavailable-property error'),('D2','internal/regexp/matcher.go','if unicodeMode(p.flags) && position > 0','if p.flags.Unicode && position > 0','loops: change Unicode mode option, omit v surrogate rewind'),('D3','internal/regexp/matcher.go','s.repeats = slices.Clone(s.repeats)','','step: drop repeat-register snapshot'),('D4','internal/regexp/matcher.go','failed = !(s.pos == len(input) || i.flags.Multiline && after)','failed = !(s.pos <= len(input) || i.flags.Multiline && after)','step: change end-assertion equality bound')]
original_source=(root/'internal/regexp/matcher.go').read_text()
start=original_source.index('\t\tif !p.anchored && !p.flags.Sticky {')
end=original_source.index('\t\tcaps :=',start)
mutants=mutants+[('D5','internal/regexp/matcher.go',original_source[start:end],'','step cost: drop complete prefix and first-character search fast path')]
results=[]
for ident,file,old,new,why in mutants:
 target=root/file; original=target.read_text(); assert original.count(old)==1,(ident,original.count(old));line=original[:original.index(old)].count('\n')+1
 try:
  target.write_text(original.replace(old,new)); subprocess.run(['gofmt','-w',file],cwd=root,check=True)
  (p/(ident+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file],cwd=root))
  with (p/(ident+'-vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/regexp/'],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0,ident
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','.'];env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/regexp-defense/cache/'+ident;t=time.monotonic()
  with (p/(ident+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[json.loads(s) for s in (p/(ident+'.log')).read_text().splitlines() if s.startswith('{')];failed=sorted({e['Test'] for e in events if e['Action']=='fail' and 'Test' in e and '/' not in e['Test']});passed=sorted({e['Test'] for e in events if e['Action']=='pass' and 'Test' in e and '/' not in e['Test']})
  record=dict(mutant=ident,file_line=file+':'+str(line),change=why,old=old,new=new,rows_failed=failed,rows_passed=passed,exit=r.returncode,wall=time.monotonic()-t,command='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+ident+'.log 2>&1')
  results.append(record);(p/'attempts.json').write_text(json.dumps(results,indent=2));print(ident,failed,flush=True)
 finally:target.write_text(original)
