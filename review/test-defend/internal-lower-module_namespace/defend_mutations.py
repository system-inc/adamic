import pathlib,subprocess,json,time,os,difflib
p=pathlib.Path('review/test-defend/internal-lower-module_namespace');(p/'diffs').mkdir(exist_ok=True)
plan=[
 ('D01','internal/lower/load_time_reads.go','return !l.provenModuleReads[node] && l.checked(local)','return (!l.provenModuleReads[node] || node.Kind == ast.KindPropertyAccessExpression) && l.checked(local)','TestModuleNamespaceInitializedReadProof','retain readiness for qualified property reads but not plain imported identifiers'),
 ('D02','internal/lower/namespaces.go','if namespaceOwnThis(declaration) {','if false {','TestCallableNamespaceReceiverStaysLoud','disable the receiver refusal for function and namespace declaration merging, keeping member-function refusal'),
 ('D03','internal/lower/namespaces.go',' || ast.GetSourceFileOfNode(declaration).IsDeclarationFile','', 'TestNamespaceAmbientContextsDoNotExecute','drop declaration-file ambient authority while keeping syntax and inherited flags'),
 ('D04','internal/lower/modules.go','l.result.Locals[local].NamespaceVar = true','l.result.Locals[local].NamespaceVar = false','TestParserFactoryBindingHoisting','clear destructured namespace var metadata'),
 ('D05','internal/lower/namespaces.go','"reading an enum before its runtime initialization; move the call after the enum declaration"','"reading a value before its runtime initialization; move the call after the enum declaration"','TestNamespaceEnumInitializationIndependentOfModuleAnalysis','change namespace preflight reachable-enum diagnostic while independent enum preflight stays intact'),
 ('D06','internal/lower/namespaces_call_graph.go','if cached := g.functions[function]; cached != nil {','if cached := g.functions[function]; cached != nil && (cached.active || len(cached.reaches) != 0) {','TestNamespaceCallGraphLinearWork','do not reuse completed empty reach sets, retaining active-cycle and nonempty memoization'),
 ('D07','internal/lower/namespaces_call_graph.go','member.reaches = reaches','member.reaches = member.reads','TestNamespaceCallGraphCycleUnion','replace completed component union with each member direct reads; also exercises early-class reach propagation')]
items=[]
for ident,file,old,new,target,reason in plan:
 original=pathlib.Path(file).read_text(); assert old in original
 line=original[:original.index(old)].count('\n')+1
 changed=original.replace(old,new,1)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+file,tofile='b/'+file))
 (p/'diffs'/(ident+'.diff')).write_text(diff)
 items.append(dict(mutant=ident,file=file,file_line=f'{file}:{line}',old=old,new=new,target=target,change=reason))
(p/'mutant-plan.json').write_text(json.dumps(items,indent=2))
results=[]
for item in items:
 ident=item['mutant'];file=pathlib.Path(item['file']); original=file.read_text()
 try:
  file.write_text(original.replace(item['old'],item['new'],1))
  vcmd=['go','vet','./internal/lower/'];t=time.monotonic()
  with (p/'logs'/('vet-'+ident+'.log')).open('w') as log: vet=subprocess.call(vcmd,stdout=log,stderr=log)
  item['vet_exit']=vet;item['vet_seconds']=time.monotonic()-t
  if vet: raise RuntimeError('vet failed '+ident)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-module-namespace/cache/'+ident)
  t=time.monotonic()
  with (p/'logs'/(ident+'.log')).open('w') as log: code=subprocess.call(cmd,env=env,stdout=log,stderr=log)
  elapsed=time.monotonic()-t
  events=[]
  for s in (p/'logs'/(ident+'.log')).read_text().splitlines():
   try: events.append(json.loads(s))
   except ValueError:pass
  fails=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
  passes=sorted(e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test'])
  errors=[e for e in events if e.get('OutputType')=='error']
  results.append(dict(item,command=cmd,env={'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},exit=code,wall_seconds=elapsed,rows_failed=fails,rows_passed=passes,errors=errors,package_result=[e for e in events if e.get('Action') in ['fail','pass'] and not e.get('Test')]))
  (p/'matrix.json').write_text(json.dumps(results,indent=2))
  print(ident,code,round(elapsed,2),fails,flush=True)
 finally:file.write_text(original)
