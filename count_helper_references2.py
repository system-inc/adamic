import json,pathlib,subprocess
out=pathlib.Path(__file__).parent;target=out/'helper2-references';target.mkdir();rows=[];results=[]
roots=['/workspace/scratch/tracing-control2','/workspace/scratch/tracing-after2'];names=sorted(str(p.relative_to(pathlib.Path(roots[0])/'built/local')) for p in (pathlib.Path(roots[0])/'built/local').rglob('*.js'));assert len(names)==10
for label,root in zip(['before','after'],roots):
 for index,name in enumerate(names):
  file=target/(label+'-'+str(index)+'.json');args=['node','--max-old-space-size=3072',str(out/'helper_reference_census2.cjs'),root,name,str(file)]
  with (target/(label+'-'+str(index)+'.log')).open('w') as log:child=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
  rows.append(dict(label=label,file=name,argv=args,exit=child.returncode));(out/'helper2-reference-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(label,name,'child exit='+str(child.returncode),flush=True)
  if child.returncode:raise SystemExit(child.returncode)
  results.append(dict(label=label,**json.loads(file.read_text())))
summary={label:dict(references=sum(r['references'] for r in results if r['label']==label),catchReferences=sum(r['catchReferences'] for r in results if r['label']==label),otherReferences=sum(r['otherReferences'] for r in results if r['label']==label)) for label in ['before','after']}
(out/'helper2-reference-census.json').write_text(json.dumps(dict(summary=summary,allTenFiles=names,results=results,requiredClosureConditionHolds=summary['before']==dict(references=9,catchReferences=3,otherReferences=6) and summary['after']==dict(references=6,catchReferences=0,otherReferences=6)),indent=2)+'\n')

for name in names:
 a=next(r for r in results if r['label']=='before' and r['file']==name);b=next(r for r in results if r['label']=='after' and r['file']==name)
 assert [(s['sourceLine'],s['referenceExpression']) for s in a['sites'] if not s['inStartTracingCatch']]==[(s['sourceLine'],s['referenceExpression']) for s in b['sites'] if not s['inStartTracingCatch']],name
assert summary['before']==dict(references=9,catchReferences=3,otherReferences=6)
assert summary['after']==dict(references=6,catchReferences=0,otherReferences=6)
