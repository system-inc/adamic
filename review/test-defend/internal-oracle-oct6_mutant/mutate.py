import pathlib,subprocess,json,os,time,difflib
p=pathlib.Path('/tmp/def-oct6');r=pathlib.Path('/workspace/adamic');scope=json.loads((p/'scope.json').read_text());patterns=scope['patterns'];cp='internal/native/runtime/input.c';gp='internal/lower/census_small.go';originals={f:(r/f).read_text() for f in [cp,gp]}
result_statement='\t\treturn &Refused{Where: l.program.Where(overload), What: label + " result " + l.checker.TypeToString(promised) + " cannot be served by implementation result " + l.checker.TypeToString(produced), Fix: "make the implementation result covariant with every overload result"}'
parameter_start=originals[gp].index('\tdeclared, served := overload.Parameters(), implementation.Parameters()');parameter_end=originals[gp].index('\n\tpromised :=',parameter_start);parameter_block=originals[gp][parameter_start:parameter_end]
menu=[('D01','TestAPromptComesBeforeTheRead',cp,'\t// What was printed comes first, as on Node: the file may be stdin, waiting on a prompt.\n\tadamic_output_flush();','\t// What was printed comes first, as on Node: the file may be stdin, waiting on a prompt.','drop flush before blocking read'),('D02','TestOverloadContractRulings',gp,result_statement,'','drop incompatible overload-result refusal'),('D03','TestOverloadContractRulings',gp,'\t\tordinal++','\t\tordinal += 2','change overload ordinal increment from 1 to 2'),('D04','TestOverloadContractRulings',gp,parameter_block,'','drop whole overload parameter-check loop and its local declaration')]
(p/'menu.json').write_text(json.dumps([{'id':i,'target':t,'file':f,'line':originals[f][:originals[f].index(a)].count('\n')+1,'change':c} for i,t,f,a,b,c in menu],indent=2))
results=json.loads((p/'matrix.json').read_text())
for i,t,file,a,b,change in menu:
 if i in {x['id'] for x in results}:continue
 original=originals[file]; src=r/file
 try:
  assert original.count(a)==(2 if i=='D03' else 1),(i,original.count(a)); src.write_text(original.replace(a,b,1))
  if file.endswith('.go'):subprocess.run(['gofmt','-w',str(src)],check=True)
  (p/(i+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),src.read_text().splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
  validate=['go','vet','./internal/lower/'] if file.endswith('.go') else ['clang','-fsyntax-only','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I','internal/native/runtime',file]
  with (p/(i+'-validate.log')).open('w') as f:subprocess.run(validate,cwd=r,stdout=f,stderr=subprocess.STDOUT,check=True)
  row={'id':i,'target':t,'file_line':file+':'+str(original[:original.index(a)].count('\n')+1),'change':change,'validate_command':' '.join(validate),'runs':[],'failed':[],'passed':[],'evidence':[]}
  for n,pat in enumerate(patterns):
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pat];env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':f'/tmp/def-oct6/cache/{i}'};start=time.monotonic()
   with (p/f'{i}-{n}.log').open('w') as f:q=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
   events=[]
   for line in (p/f'{i}-{n}.log').read_text().splitlines():
    try:events.append(json.loads(line))
    except:pass
   failures=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test') and '/'not in e['Test']]; passes=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/'not in e['Test']]
   row['failed']+=failures;row['passed']+=passes;row['evidence'] += [e.get('Output','').strip() for e in events if e.get('Test','').split('/')[0]==t and ('output_test.go:'in e.get('Output','') or 'overload_contract_test.go:'in e.get('Output',''))]
   row['runs'].append({'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+f' > {i}-{n}.log 2>&1','exit':q.returncode,'wall_seconds':time.monotonic()-start,'top_failures':failures,'top_passes':passes,'native_subcases':[e['Test'] for e in events if e.get('Action') in ['pass','fail'] and e.get('Test','').startswith('TestNativeAgreesWithNode/')]})
  results.append(row);(p/'matrix.json').write_text(json.dumps(results,indent=2));print(i,row['failed'],flush=True)
 finally:src.write_text(original)
