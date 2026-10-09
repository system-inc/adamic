import pathlib,subprocess,json,difflib,time,os
p=pathlib.Path(__file__).resolve().parent; base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip();(p/'base.txt').write_text(base+'\n')
audit='origin/test-audit/internal-lower-module_namespace'
for n in ['REPORT.txt','report.json','rows.json','mutant-plan.json','matrix.json','reached-functions.txt']:
 (p/('prior-'+n)).write_bytes(subprocess.check_output(['git','show',audit+':review/test-audit/internal-lower-module_namespace/'+n]))
menu=[dict(mutant='D01',file='internal/lower/namespaces_call_graph.go',old='entry.low = min(entry.low, edge.low)',new='entry.low = min(entry.low, edge.index)',change='use child discovery index instead of transitive low link',target='TestNamespaceCallGraphCycleUnion'),dict(mutant='D02',file='internal/lower/namespaces.go',old='if declaration.Kind == ast.KindEnumDeclaration {\n\t\t\t\t\treturn l.notYet(read,',new='if declaration.Kind == ast.KindEnumDeclaration && l.result != nil {\n\t\t\t\t\treturn l.notYet(read,',change='gate enum-specific initialization refusal on prepared IR state',target='TestNamespaceEnumInitializationIndependentOfModuleAnalysis')]
(p/'menu.json').write_text(json.dumps(menu,indent=2)+'\n');results=[]
for r in menu:
 path=pathlib.Path(r['file']);clean=path.read_text();assert clean.count(r['old'])==1;r['file_line']=r['file']+':'+str(clean[:clean.index(r['old'])].count('\n')+1)
 changed=clean.replace(r['old'],r['new']);diff=''.join(difflib.unified_diff(clean.splitlines(True),changed.splitlines(True),fromfile='a/'+r['file'],tofile='b/'+r['file']));dp=p/(r['mutant']+'.diff');dp.write_text(diff)
 subprocess.run(['git','apply','--check',str(dp)],check=True);subprocess.run(['git','apply',str(dp)],check=True)
 try:
  with (p/(r['mutant']+'-vet.log')).open('w') as log:
   vet=subprocess.run(['timeout','120','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT);assert vet.returncode==0
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/lower-namespace-defend-cache/'+r['mutant'];r['command']='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > '+r['mutant']+'.log 2>&1'
  start=time.monotonic()
  with (p/(r['mutant']+'.log')).open('w') as log:ret=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],env=env,stdout=log,stderr=subprocess.STDOUT)
  r['wall_seconds']=time.monotonic()-start;r['exit']=ret.returncode
  events=[]
  for l in (p/(r['mutant']+'.log')).read_text().splitlines():
   try:events.append(json.loads(l))
   except ValueError:pass
  r['rows_failed']=sorted({e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')});r['rows_passed']=sorted({e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']});r['skipped']=[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')];r['failures']=[{'test':e.get('Test'),'line':e.get('Output','').strip()} for e in events if e.get('Action')=='output' and e.get('Test','').split('/')[0] in r['rows_failed'] and '.go:' in e.get('Output','')];r['binary_seconds']=[e.get('Elapsed') for e in events if not e.get('Test') and e.get('Action')=='fail']
  results.append(r);(p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(r['mutant'],r['rows_failed'],r['wall_seconds'],flush=True)
 finally:subprocess.run(['git','apply','-R',str(dp)],check=True)
