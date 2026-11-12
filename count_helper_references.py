import json,pathlib,subprocess
out=pathlib.Path(__file__).parent;target=out/'helper-references';target.mkdir();rows=[];results=[]
roots=['/workspace/scratch/tracing-control','/workspace/scratch/tracing-after-final'];names=sorted(str(p.relative_to(pathlib.Path(roots[0])/'built/local')) for p in (pathlib.Path(roots[0])/'built/local').rglob('*.js'));assert len(names)==10
for label,root in zip(['before','after'],roots):
 for index,name in enumerate(names):
  file=target/(label+'-'+str(index)+'.json');args=['node','--max-old-space-size=3072',str(out/'helper_reference_census.cjs'),root,name,str(file)]
  with (target/(label+'-'+str(index)+'.log')).open('w') as log:child=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT)
  rows.append(dict(label=label,file=name,argv=args,exit=child.returncode));(out/'helper-reference-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(label,name,'child exit='+str(child.returncode),flush=True)
  if child.returncode:raise SystemExit(child.returncode)
  results.append(dict(label=label,**json.loads(file.read_text())))
summary={label:dict(references=sum(r['references'] for r in results if r['label']==label),catchReferences=sum(r['catchReferences'] for r in results if r['label']==label),otherReferences=sum(r['otherReferences'] for r in results if r['label']==label)) for label in ['before','after']}
(out/'helper-reference-census.json').write_text(json.dumps(dict(summary=summary,allTenFiles=names,results=results,requiredClosureConditionHolds=summary['before']['otherReferences']==0 and summary['after']['references']==0),indent=2)+'\n')
