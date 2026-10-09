import pathlib,subprocess,json,time,re
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-lower-enum_flags');pkg='./internal/lower/';names=['TestFlagEnumsOpen','TestFlagEnumsDomain','TestEnumNeverDefault','TestFlagEnumLiteralSpellings','TestFlagEnumMemberAliases','TestEnumNameEnumeration','TestFlagEnumInlineIteration','TestFlagEnumAliasBoundaries','TestEnumInitializationGraphMemo','TestEnumNamespaceSharedCycle','TestStringEnumsStayClosed','TestNumericEnumLiteralPromises'];subs=['TestNumericEnumLiteralPromises','TestFlagEnumInlineIteration','TestNumericEnumNeverProof','TestFlagEnumInlineIteration','TestFlagEnumInlineIteration','TestFlagEnumInlineIteration','TestEnumNameEnumeration','TestNumericEnumLiteralPromises','TestEnumNamespaceSharedCycle','TestEnumInitializationGraphMemo',None,'TestFlagEnumAliasBoundaries'];allnames=list(dict.fromkeys(names+[n for n in subs if n]));costs={}
for f in ['audit-notes.txt','audit.json','plan.json','scope.json','matrix.json','reached-functions.txt','environment.json','row-aliases.json']:
 (p/('prior-'+f)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/internal-lower-enum_flags:review/test-audit/internal-lower-enum_flags/'+f]))
listed=re.findall(r'^Test\w+',(p/'list.log').read_text(),re.M);assert set(allnames)<=set(listed)
for n in allnames+['rest-of-package']:
 pattern='^'+n+'$' if n!='rest-of-package' else '^('+'|'.join(x for x in listed if x!='TestStringEnumsStayClosed')+')$'
 c=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/lower','-coverprofile='+str(p/(n+'.cover')),pkg,'-run',pattern];t=time.monotonic()
 with (p/(n+'-coverage.log')).open('w') as f:r=subprocess.run(c,stdout=f,stderr=subprocess.STDOUT)
 costs[n]={'exit':r.returncode,'seconds':time.monotonic()-t,'command':c};print(n,costs[n]['exit'],round(costs[n]['seconds'],3),flush=True)
 if r.returncode:break
(p/'coverage-runs.json').write_text(json.dumps(costs,indent=2))
def blocks(name):
 result={}
 for s in (p/(name+'.cover')).read_text().splitlines()[1:]:
  loc,stmts,count=s.split();result[loc]=int(count)
 return result
summary=[]
for n,s in zip(names,subs):
 if n not in costs or (s or 'rest-of-package')not in costs:continue
 own=blocks(n);other=blocks(s or 'rest-of-package');exclusive=[loc for loc,count in own.items() if count>0 and other.get(loc,0)==0]
 (p/(n+'-exclusive.txt')).write_text('\n'.join(exclusive)+'\n')
 summary.append({'test':n,'compared_with':s or 'rest of package','exclusive_blocks':len(exclusive),'enum_graph_cast_blocks':[x for x in exclusive if any(w in x for w in ['/enums.go:','/enum_','/namespaces_call_graph.go:','/cast_proof.go:','/assignments.go:'])]})
(p/'coverage-comparisons.json').write_text(json.dumps(summary,indent=2));print(json.dumps(summary,indent=2))
