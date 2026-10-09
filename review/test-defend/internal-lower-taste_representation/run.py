import pathlib,subprocess,json,os,time
root=pathlib.Path('/tmp/defend-lower-taste')
pairs=[('TestViewObjectWritesNeedSourceCertificate','TestDefaultTaggedInterfaceAdmission'),('TestMixedUnionContractGraph','TestMixedUnionContractRecursiveMember'),('TestUntaggedViewStructuralFallback','TestUntaggedViewMemberTags')]
coverage={}
for a,b in pairs:
 def blocks(n):
  return {x.split()[0]:int(x.split()[2]) for x in (root/(n+'.cover')).read_text().splitlines()[1:]}
 aa,bb=blocks(a),blocks(b)
 coverage[a]={'subsumer':b,'exclusive_blocks':[k for k,v in aa.items() if v>0 and bb.get(k,0)==0]}
(root/'coverage-differences.json').write_text(json.dumps(coverage,indent=2))
p='internal/lower/interface_cast.go'; s=pathlib.Path(p).read_text(); start=s.index('\t\tif part.Kind == ast.KindPropertyAccessExpression && fields[part.Name().Text()] && ast.IsAssignmentTarget(part) {'); end=s.index('\n\t\tif part.Kind == ast.KindGetAccessor',start)
plan=[('D1',p,s[start:end],'', 'Drop object-field assignment refusal block'),('D2','internal/lower/view_contracts.go','Name: l.checker.TypeToString(target), Of: of}','Name: "", Of: of}','Change completed scalar/object contract name to empty string'),('D3','internal/lower/view_unions_untagged.go','contract.Kind != ir.ViewObject || contract.Of != ir.Object','contract.Kind != ir.ViewObject && contract.Of != ir.Object','Flip member rejection connective from OR to AND')]
results=[]
for mid,p,old,new,change in plan:
 f=pathlib.Path(p); base=f.read_text(); assert base.count(old)==1,(mid,base.count(old)); line=base[:base.index(old)].count('\n')+1
 try:
  f.write_text(base.replace(old,new)); diff=subprocess.check_output(['git','diff','--',p]); (root/(mid+'.diff')).write_bytes(diff)
  t=time.monotonic()
  with (root/(mid+'-vet.log')).open('w') as log: vet=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
  assert vet.returncode==0,mid
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR=str(root/'cache'/mid)); cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.']
  with (root/(mid+'.log')).open('w') as log: run=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for ln in (root/(mid+'.log')).read_text().splitlines():
   try: events.append(json.loads(ln))
   except: pass
  failed=sorted({e['Test'] for e in events if e.get('Action')=='fail' and 'Test'in e and '/'not in e['Test']})
  passed=sorted({e['Test'] for e in events if e.get('Action')=='pass' and 'Test'in e and '/'not in e['Test']})
  result={'mutant':mid,'file_line':p+':'+str(line),'change':change,'rows_failed':failed,'rows_passed':passed,'command':'ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'exit_code':run.returncode,'wall_seconds':time.monotonic()-t,'failures':[e for e in events if e.get('Action')=='output' and any(x in e.get('Output','') for x in ['unchecked write','member was substituted','admitted without'])]}
  results.append(result);(root/'matrix.json').write_text(json.dumps(results,indent=2));print(mid,failed,run.returncode,flush=True)
 finally:f.write_text(base)
