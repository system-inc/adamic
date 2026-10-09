import subprocess,pathlib,difflib,json,time,os
p=pathlib.Path('review/test-defend/internal-load-export_collision')
base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip()
menu=[
('D01','TestATypeErrorFailsTheLoad','internal/load/load.go','\t\tmessage = diagnostic.Localize(english)\n','', 'drop localized diagnostic message assignment'),
('D02','TestLoadRefusesWhatIsNotAdamic','internal/load/load.go','if len(paths) == 0 {','if len(paths) < 0 {','off-by-one empty root bound'),
('D03','TestNodeLibraryRejectsDifferentVersion','internal/load/node_library.go','pin.Name != "@types/node" || pin.Version != NodeTypesVersion','pin.Name != "@types/node"','drop version comparison from refusal condition'),
('D04','TestTheProgramsInTheSpecLoad','internal/load/load.go','core.ScriptTargetES2024','core.ScriptTargetES5','change target option to ES5'),
('D05','TestNodeLibraryKeepsOfficialConsoleSignatures','internal/load/node_library.go','specifier != nil && strings.HasPrefix(specifier.Text(), "node:")','specifier != nil && strings.HasPrefix(specifier.Text(), "node:") && (statement.Kind != ast.KindImportDeclaration || statement.AsImportDeclaration().ImportClause == nil || !statement.AsImportDeclaration().ImportClause.IsTypeOnly())','change Node import recognition condition to exclude type-only imports'),
('D06','TestExplicitExportResolvesStarCollision','internal/load/load.go','AllowImportingTsExtensions: core.TSTrue','AllowImportingTsExtensions: core.TSFalse','disable explicit TypeScript-extension imports'),
('D07','TestExplicitExportResolvesStarCollision','internal/load/load.go','core.ModuleResolutionKindBundler','core.ModuleResolutionKindClassic','change module resolution to Classic'),
('D08','TestExplicitExportResolvesStarCollision','internal/load/load.go','core.ModuleKindESNext','core.ModuleKindCommonJS','change module kind to CommonJS'),
]
records=[]
for ident,target,file,old,new,change in menu:
 original=subprocess.check_output(['git','show',base+':'+file],text=True)
 assert original.count(old)==1,(ident,original.count(old))
 changed=original.replace(old,new,1)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (p/(ident+'.diff')).write_text(diff)
 line=original[:original.index(old)].count('\n')+1
 rec=dict(mutant=ident,target=target,file_line=f'{file}:{line}',change=change,old=old,new=new)
 subprocess.run(['git','apply','--check',str(p/(ident+'.diff'))],check=True)
 subprocess.run(['git','apply',str(p/(ident+'.diff'))],check=True)
 try:
  start=time.monotonic()
  with (p/(ident+'-vet.log')).open('w') as log:
   vet=subprocess.run(['timeout','120','go','vet','./internal/load/'],stdout=log,stderr=subprocess.STDOUT)
  rec['vet_exit']=vet.returncode;rec['vet_seconds']=time.monotonic()-start
  if vet.returncode==0:
   env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']=f'/workspace/load-defend-tmp/cache/{ident}'
   command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/load/','-run','.']
   start=time.monotonic()
   with (p/(ident+'.log')).open('w') as log: run=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT)
   rec['exit']=run.returncode;rec['wall_seconds']=time.monotonic()-start
   events=[]
   for text in (p/(ident+'.log')).read_text().splitlines():
    try:events.append(json.loads(text))
    except ValueError:pass
   rec['rows_failed']=sorted({e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']})
   rec['rows_passed']=sorted({e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']})
   rec['failures']=[e['Output'].strip() for e in events if e.get('Test','').split('/')[0]==target and 'Output' in e and '.go:' in e['Output']]
   rec['command']='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(command)+' > '+ident+'.log 2>&1'
  print(json.dumps(rec),flush=True)
  records.append(rec);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n')
 finally:subprocess.run(['git','apply','-R',str(p/(ident+'.diff'))],check=True)
