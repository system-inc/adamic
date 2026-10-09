from pathlib import Path
import subprocess,os,time,json
p=Path('review/test-audit/stage1-cohere-estree-deep')
rows=['TestDeepGrammar','TestGeneratedAgreement','TestOriginalLibraries','TestDecoratedExports','TestDecoratedExportsPlantedDisagreement','TestDecoratedExportMutant','TestDecoratedExportMutantPlantedSurvivor','TestDecoratedExportLibraries','TestRecoveredExpressions','TestRecoveredExpressionMutant','TestUnattachedDecorator','TestUnattachedDecoratorControl']
results=[]
for row in rows:
 for rep in range(1,4):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^'+row+'$']
  start=time.monotonic()
  with (p/f'timing-{row}-{rep}.log').open('w') as out:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_ESTREE_LIBRARY='/tmp/u085/library'),stdout=out,stderr=subprocess.STDOUT)
  elapsed=time.monotonic()-start; results.append(dict(row=row,rep=rep,exit=r.returncode,wall=elapsed,command='ADAMIC_ESTREE_LIBRARY=/tmp/u085/library '+' '.join(cmd)))
  (p/'baseline-runs.json').write_text(json.dumps(results,indent=2)+'\n')
  print(row,rep,r.returncode,round(elapsed,3),flush=True)
  if r.returncode:
   print('STOP: clean isolated baseline nonzero; inspect log before proceeding',flush=True);raise SystemExit(r.returncode)
