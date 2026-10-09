import pathlib,subprocess,json,time,os,difflib
p=pathlib.Path('review/test-defend/stage1-typescript-scanner');base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip();(p/'base.txt').write_text(base+'\n');file='stage1/typescript/scanner/scanner.ts'
menu=[('D01',"first === -1 ? 'JsxTextAllWhiteSpaces' : 'JsxText'","first !== -1 ? 'JsxTextAllWhiteSpaces' : 'JsxText'",'flip JSX whitespace token classification'),('D02','else if(code === 47 && !characterClass)','else if(code === 47 && characterClass)','flip regex terminator character-class condition'),('D03','this.template(true);','this.template(false);','change template rescan diagnostic option true to false')]
(p/'menu.json').write_text(json.dumps(menu,indent=2))
records=[]
for ident,old,new,change in menu:
 original=subprocess.check_output(['git','show',base+':'+file],text=True);assert original.count(old)==1
 modified=original.replace(old,new,1);patch=p/(ident+'.diff');patch.write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 rec=dict(mutant=ident,file_line=file+':'+str(original[:original.index(old)].count('\n')+1),change=change,old=old,new=new)
 subprocess.run(['git','apply','--check',str(patch)],check=True);subprocess.run(['git','apply',str(patch)],check=True)
 subprocess.run(['git','add',file],check=True)
 with (p/(ident+'-commit.log')).open('w') as log:subprocess.run(['git','commit','-m','Plant scanner defense mutant '+ident+' for clean corpus replay'],stdout=log,stderr=subprocess.STDOUT,check=True)
 rec['variant_commit']=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
 try:
  env=os.environ.copy();env.update(ADAMIC_BUILD_CACHE_DIR='/workspace/scanner-defend-tmp/cache/'+ident,ADAMIC_SCANNER_PROFILE_DIR='/workspace/scanner-defend-tmp/artifacts/'+ident,ADAMIC_SCANNER_PROFILE_SNAPSHOTS='/workspace/scanner-defend-tmp/artifacts/'+ident)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/scanner/','-run','.'];rec['command']='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' ADAMIC_SCANNER_PROFILE_DIR='+env['ADAMIC_SCANNER_PROFILE_DIR']+' ADAMIC_SCANNER_PROFILE_SNAPSHOTS='+env['ADAMIC_SCANNER_PROFILE_SNAPSHOTS']+' '+' '.join(cmd)+' > '+ident+'.log 2>&1';rec['env']={k:env[k] for k in ['ADAMIC_BUILD_CACHE_DIR','ADAMIC_SCANNER_PROFILE_DIR','ADAMIC_SCANNER_PROFILE_SNAPSHOTS','ADAMIC_TYPESCRIPT_SOURCE','ADAMIC_SCANNER_BENCH']}
  start=time.monotonic()
  with (p/(ident+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  rec['exit']=r.returncode;rec['wall_seconds']=time.monotonic()-start
  events=[]
  for s in (p/(ident+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except ValueError:pass
  rec['tests_failed']=sorted({e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']});rec['tests_passed']=sorted({e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']})
  def row(n):
   if n.startswith('TestScannerAgreesWithTypescriptGo_'):return 'TestScannerAgreesWithTypescriptGo family'
   if n.startswith('TestProduct_Scanner'):return 'TestProduct_Scanner family'
   return n
  rec['rows_failed']=sorted({row(n) for n in rec['tests_failed']});rec['rows_passed']=sorted({row(n) for n in rec['tests_passed']}-{row(n) for n in rec['tests_failed']})
  rec['failures']=[dict(test=e['Test'],line=e['Output'].strip()) for e in events if ('TestScannerAgreesWithTypescriptGo_' in e.get('Test','') or e.get('Test')=='TestProfileSnapshotsAgree') and 'Output' in e and '.go:' in e['Output'] and ('native:' in e['Output'] or 'Node:' in e['Output'] or 'release:' in e['Output'])]
  rec['build_evidence']=[e for e in events if e.get('Test','').startswith('TestProduct_Scanner') and e.get('Action')=='pass' or e.get('Test')=='TestProfileArtifacts' and e.get('Action')=='pass']
  records.append(rec);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(rec),flush=True)
 finally:
  subprocess.run(['git','apply','-R',str(patch)],check=True);subprocess.run(['git','add',file],check=True)
  with (p/(ident+'-restore.log')).open('w') as log:subprocess.run(['git','commit','-m','Restore scanner after defense mutant '+ident],stdout=log,stderr=subprocess.STDOUT,check=True)
