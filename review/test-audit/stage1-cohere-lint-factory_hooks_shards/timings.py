from pathlib import Path
import subprocess,time,json,os
out=Path('/workspace/adamic/review/test-audit/stage1-cohere-lint-factory_hooks_shards');root=out.parents[2]
rows={'FactoryHooks family':'^TestFactoryHooks(Union|_[0-9]{3})$','TestFactoryHooksPlantedFailure':'^TestFactoryHooksPlantedFailure$','TestNestedConstructorGap':'^TestNestedConstructorGap$','TestProduct_DotARenameSource':'^TestProduct_DotARenameSource$','TestProduct_DotARenameOracle':'^TestProduct_DotARenameOracle$','TestProduct_RulesAgreeOracle':'^TestProduct_RulesAgreeOracle$','TestRulesAgreeLoweringCacheKey':'^TestRulesAgreeLoweringCacheKey$'}
metrics=[]
for row,regex in rows.items():
 for trial in [1,2,3]:
  label=row.replace(' ','_')+f'-{trial}';start=time.monotonic()
  with (out/'logs'/f'timing-{label}.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',regex],cwd='/workspace/adamic',stdout=log,stderr=subprocess.STDOUT)
  metrics.append(dict(row=row,trial=trial,regex=regex,wall=time.monotonic()-start,exit=r.returncode));(out/'timing-commands.json').write_text(json.dumps(metrics,indent=2))
