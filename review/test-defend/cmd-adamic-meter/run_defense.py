import pathlib,subprocess,json,time,os,difflib
p=pathlib.Path('review/test-defend/cmd-adamic-meter');base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip();(p/'base.txt').write_text(base+'\n')
menu=[
('D01','TestFixtureCorpusCountsEveryDiagnosticKind','cmd/adamic-meter/main.go','return []observation{{"Refused", normalize(refused.What), refused.Where}}, true','return []observation{{"Rejected", normalize(refused.What), refused.Where}}, true','change Refused report label to Rejected'),
('D02','TestOptionalAdaptationOwnsOnlyNamedRoots','cmd/adamic-meter/optional.go',' || !owned[ast.GetSourceFileOfNode(declaration).FileName().AsString()]','','drop external declaration ownership condition'),
('D03','TestAdaptOptionalPropertiesPreservesRuntimeAndDisk','cmd/adamic-meter/optional.go','\t\tvar target *checker.Type','\t\tif declaration.Type().Kind == ast.KindFunctionType { return }\n\t\tvar target *checker.Type','return early for callable property type declarations'),
('D04','TestOptionalAdaptationUTF16AndAdamicExtension','cmd/adamic-meter/optional.go','strings.HasSuffix(name, ".a.ts") && owned[strings.TrimSuffix(name, ".ts")]','strings.HasSuffix(name, ".a.ts")','drop ownership condition distinguishing a real .a.ts from an alias'),
('D05','TestImplicitReturnsNeedNoAdaptation','cmd/adamic-meter/main.go','if !errors.As(baselineErr, &checkError) {\n\t\treturn nil, nil, nil','if !errors.As(baselineErr, &checkError) {\n\t\treturn nil, nil, os.ErrInvalid','change the valid-baseline early return to an invalid-input refusal'),
('D06','TestReturnAdaptationLeavesUnprovenContracts','cmd/adamic-meter/returns.go','if includesUndefined {','if includesUndefined || typ.Flags()&checker.TypeFlagsUnknown != 0 {','change selection condition to admit unknown return contracts'),
('D07','TestReturnAdaptationOwnsOnlyRoots','cmd/adamic-meter/returns.go','file == nil || !owned[name]','file == nil','drop return declaration root ownership condition'),
('D08','TestOptionalAdaptationPreservesLiveMethodPlacement','cmd/adamic-meter/optional.go','\t\t\t\t\tselectOptional(c, c.GetSymbolAtLocation(assignment.Left), c.GetTypeAtLocation(assignment.Right), owned, selected)','\t\t\t\t\tif symbol := c.GetSymbolAtLocation(assignment.Left); symbol != nil && len(symbol.Declarations) != 0 && symbol.Declarations[0].Kind == ast.KindMethodDeclaration { release(); return overlay, 0, nil }\n\t\t\t\t\tselectOptional(c, c.GetSymbolAtLocation(assignment.Left), c.GetTypeAtLocation(assignment.Right), owned, selected)','return early from the whole adaptation on a live-method assignment'),
('D09','TestOptionalAdaptationRechecksPresenceNarrowing','cmd/adamic-meter/optional.go','if !relevant {\n\t\treturn overlay, 0, nil','if !relevant {\n\t\treturn nil, 0, nil','drop accumulated overlay when only non-optional diagnostics remain'),
]
records=[]
for ident,target,file,old,new,change in menu:
 original=subprocess.check_output(['git','show',base+':'+file],text=True);assert original.count(old)==1,(ident,original.count(old))
 modified=original.replace(old,new,1);diff=''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file));patch=p/(ident+'.diff');patch.write_text(diff)
 rec=dict(mutant=ident,target=target,file_line=file+':'+str(original[:original.index(old)].count('\n')+1),change=change,old=old,new=new)
 subprocess.run(['git','apply','--check',str(patch)],check=True);subprocess.run(['git','apply',str(patch)],check=True)
 try:
  start=time.monotonic()
  with (p/(ident+'-vet.log')).open('w') as log:vet=subprocess.run(['timeout','120','go','vet','./cmd/adamic-meter/'],stdout=log,stderr=subprocess.STDOUT)
  rec['vet_exit']=vet.returncode;rec['vet_seconds']=time.monotonic()-start
  if vet.returncode==0:
   env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/workspace/meter-defend-tmp/cache/'+ident
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-meter/','-run','.'];rec['command']='ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd)+' > '+ident+'.log 2>&1'
   start=time.monotonic()
   with (p/(ident+'.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
   rec['exit']=r.returncode;rec['wall_seconds']=time.monotonic()-start
   events=[]
   for s in (p/(ident+'.log')).read_text().splitlines():
    try:events.append(json.loads(s))
    except ValueError:pass
   rec['rows_failed']=sorted({e['Test'] for e in events if e.get('Action')=='fail' and 'Test' in e and '/' not in e['Test']});rec['rows_passed']=sorted({e['Test'] for e in events if e.get('Action')=='pass' and 'Test' in e and '/' not in e['Test']})
   rec['failures']=[e['Output'].strip() for e in events if e.get('Test','').split('/')[0]==target and 'Output' in e and '.go:' in e['Output']]
  records.append(rec);(p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps(rec),flush=True)
 finally:subprocess.run(['git','apply','-R',str(patch)],check=True)
