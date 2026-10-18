import os,subprocess,json,time
from pathlib import Path
ev=Path('review/compiler/loom-slice-3-2');rows=[]
cases=[('native','./internal/native','^Test(TSGoCache(Release|Sanitized)|TSGoBuildSeesTheProgramsFeatures|CheckerBuildsNameNoPlaceOfTheArchive|NativeCachesWriteNothingUnderHome)$'),('optional','./internal/oracle','^TestOptionalAfterCall'),('spread','./internal/oracle','^Test(CallSpread|JSONStringifyShorthand)'),('function','./internal/oracle','^TestFunction(Adamic|TypeScript|CallSignature)'),('merged','./internal/oracle','^TestMergedField'),('assignment','./internal/oracle','^TestAssignmentProof'),('flow','./internal/flow','^TestOptionalAfterCallFlowEdges$'),('lower','./internal/lower','^TestOptionalAfterCallPreservesEarlierThrow$'),('readers','./internal/ir','^TestCallTargetReaders$')]
for label,package,selection in cases:
 command=['timeout','180','go','test',package,'-run',selection,'-count=1','-timeout','90s','-parallel','4','-json'];start=time.monotonic()
 with (ev/(label+'-green.jsonl')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'GOMAXPROCS':'4','ADAMIC_GATE_UNCACHED':'1'})
 rows.append(dict(label=label,command=command,exit=r.returncode,seconds=time.monotonic()-start));(ev/'focused-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(label,r.returncode,flush=True)
 if r.returncode:break
