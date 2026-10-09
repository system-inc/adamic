from pathlib import Path
import subprocess,json,os,time,difflib
p=Path('review/test-defend/internal-oracle-method_values_mutant');base={f:Path(f).read_text() for f in ['internal/lower/module_namespace.go','internal/lower/load_time_reads.go','internal/lower/namespaces.go']}
plan=[('D01','internal/lower/load_time_reads.go','index < positions[file]','index > positions[file]','module provider order: flip bound'),('D02','internal/lower/module_namespace.go','!load.IsLibrary(declaration.AsSourceFile())','load.IsLibrary(declaration.AsSourceFile())','module namespace library exclusion: flip condition'),('D03','internal/lower/load_time_reads.go','return !l.provenModuleReads[node] && l.checked(local)','return l.provenModuleReads[node] && l.checked(local)','module runtime readiness: flip proven condition'),('D04','internal/lower/namespaces.go','body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: true}})','body = append(body, ir.Assign{Local: l.namespaceReadyLocal(node), Value: ir.BooleanConstant{Value: false}})','namespace completion: change ready constant'),('D05','internal/lower/namespaces.go','b.parameters[len(checks)]','b.parameters[len(checks)-1]','qualified read result: off-by-one parameter bound'),('D06','internal/lower/namespaces.go','parent.Kind == ast.KindModuleBlock || module','parent.Kind != ast.KindModuleBlock || module','namespace member binding scope: flip condition')]
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
plain='^(Test.*Namespace.*|TestImportCycle.*)$'
family='^TestNativeAgreesWithNode$/^(internal|stage3)$/^(oracle|namespace-live-export|namespace-init-sys|fixtures)$/^(testdata|live[.]a|realpath-control[.]a|cwd-control[.]a|namespaces)$/^(library_method_values.*|module_namespace_reads|namespaces.*|e4eec87_f2.*|narrowed_union_valid[.]a|08_parser_jsdoc_nested[.]a)$'
(p/'matrix-selectors.json').write_text(json.dumps([plain,family],indent=2)+'\n')
results=[]
for mid,f,old,new,meaning in plan:
 assert base[f].count(old)==1,(mid,base[f].count(old))
 modified=base[f].replace(old,new)
 diff=''.join(difflib.unified_diff(base[f].splitlines(True),modified.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (p/(mid+'.diff')).write_text(diff);Path(f).write_text(modified)
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/method-defend/cache/'+mid;env['ADAMIC_GATE_UNCACHED']='1'
 start=time.monotonic()
 try:
  with open(p/(mid+'-vet.log'),'w') as log:vet=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  row={'mutant':mid,'file_line':f+':'+str(base[f][:base[f].index(old)].count('\n')+1),'change':meaning,'vet_exit':vet,'runs':[]}
  if vet==0:
   for k,s in enumerate([plain,family]):
    cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',s]
    logpath=p/(mid+f'-{k}.log')
    with open(logpath,'w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
    events=[]
    for line in logpath.read_text().splitlines():
     try:events.append(json.loads(line))
     except:pass
    row['runs'].append({'command':'ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'exit':r.returncode,'failed':[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],'passed':[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],'skipped':[e['Test'] for e in events if e.get('Action')=='skip' and e.get('Test')],'cooked':any('test timed out' in e.get('Output','') for e in events),'diagnostics':[e for e in events if e.get('Action')=='output' and '.go:' in e.get('Output','') and (e.get('Test','').startswith('TestModuleNamespaceReadsMatchNode') or e.get('Test')=='TestNamespaceLiveExportBoundary')]})
  row['wall_seconds']=round(time.monotonic()-start,3);results.append(row);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(mid,vet,[r['failed'][:4] for r in row['runs']],flush=True)
 finally:Path(f).write_text(base[f])
