import json,os,subprocess,time,sys
from pathlib import Path
out=Path('review/compiler/loom-slice-2-2')
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
env=dict(os.environ,GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1')
checks={
 'reader-products':(['go','test','./internal/ir','-run','^Test(CallTargetExportInputs|Product_CallTargetReaderArchives)','-json','-count=1','-timeout=90s'],240),
 'readers':(['go','test','./internal/ir','-run','^TestCallTargetReaders','-json','-count=1','-timeout=90s'],180),
 'oracle':(['go','test','./internal/oracle','-run','^Test(CallbackWidening|WeakDifferential|Miscompile2A|SelfCompare)','-json','-count=1','-timeout=90s'],300),
 'fixtures':(['go','test','./stage3/fixtures','-run','^Test(FixturesSelfCompare|FixtureDirectoriesHaveTopLevelTests)','-json','-count=1','-timeout=90s'],240),
 'support':(['go','test','./internal/native','./internal/ir','./internal/flow','./internal/fresh','-run','^Test(CallbackWidening|CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues|ErrorUndefinedMessageHasNoOwnProperty|SplitSelfCompareAgreesWithNode)$|^TestCallbackWidening','-json','-count=1','-timeout=90s'],300),
 'vet':(['go','vet','./internal/lower','./internal/native','./internal/ir','./internal/flow','./internal/fresh','./internal/oracle','./stage3/fixtures'],240),
 'counts':(['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-v','-timeout=600s','-args','-update-counts'],660),
}
for name in sys.argv[1:]:
 command,limit=checks[name];start=time.monotonic()
 with (out/('finish-'+name+'.jsonl' if name not in ['vet','counts'] else 'finish-'+name+'.log')).open('w') as f:
  ran=subprocess.run(['timeout',str(limit)]+command,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=limit+10)
 row=dict(name=name,command=['timeout',str(limit)]+command,exit=ran.returncode,seconds=round(time.monotonic()-start,3))
 with (out/'finish-checks.jsonl').open('a') as f:f.write(json.dumps(row)+'\n')
 print(json.dumps(row),flush=True)
 assert ran.returncode==0
