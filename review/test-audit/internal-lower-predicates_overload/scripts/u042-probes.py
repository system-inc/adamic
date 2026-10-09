import pathlib,json,time,subprocess,os,difflib
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';P=R/'internal/lower';base=json.load(open(E/'base.json'));meta=json.load(open(E/'function-offsets.json'))
probes=[('P01','lower.go','Lower','nil, nil',['TestPredicateOverloadRuntime','TestIndirectPredicateOverloadIsPending','TestPredicateCallbackContracts','TestPredicateOverloadCallback','TestUnprovenPredicateReturnsAreRefused','TestPredicateBodiesAreProven','TestPrimitiveAdmittingSlotsUseBoxes']),('P02','predicates_proof.go','provePredicate','predicateProof{}, nil',['TestPredicateBodyProof']),('P03','refusals.go','refuse','nil',['TestConditionAssertionAdmission','TestEveryNeedsCallbackEffects']),('P04','predicates_proof.go','predicateUseDirections','0',['TestPredicateUseRegions','TestPredicateUsesBelongToEachCall'])]
(E/'probes.json').write_text(json.dumps(probes,indent=2));(E/'probes').mkdir(exist_ok=True)
for id,f,name,value,rows in probes:
 assert f in meta,(id,f);fn=next(x for x in meta[f] if x['name']==name);s=base[f];pos=fn['body']+1;new=s[:pos]+'\nif auditSelector()=="'+id+'" {return '+value+'}\n'+s[pos:];(E/'probes'/(id+'.switched.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/internal/lower/'+f,tofile='b/internal/lower/'+f)))
while not (E/'matrix.done').exists():time.sleep(1)
for f in {p[1] for p in probes}:
 s=base[f];edits=[]
 for id,pf,name,value,rows in probes:
  if pf==f:
   fn=next(x for x in meta[f] if x['name']==name);edits.append((fn['body']+1,'\nif auditSelector()=="'+id+'" {return '+value+'}\n'))
 for pos,text in sorted(edits,reverse=True):s=s[:pos]+text+s[pos:]
 (P/f).write_text(s)
(P/'audit_selector.go').write_text('package lower\nimport "os"\nfunc auditSelector()string{return os.Getenv("ADAMIC_MUTANT")}\n')
records=[]
for id,f,name,value,rows in probes:
 for row in rows:
  start=time.monotonic();env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u042/cache/'+id)
  with (E/(id+'-'+row+'.log')).open('w') as log:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
  records.append(dict(id=id,row=row,exit=rc,seconds=time.monotonic()-start));(E/'probe-time.json').write_text(json.dumps(records,indent=2))
for f in {p[1] for p in probes}:(P/f).write_text(base[f])
(P/'audit_selector.go').unlink();(E/'probes.done').write_text('done')
