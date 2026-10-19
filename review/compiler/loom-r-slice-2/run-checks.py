import json, os, subprocess, sys, time
from pathlib import Path
out=Path(__file__).parent
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:4])
env=dict(os.environ,GOMAXPROCS="4",GOPROXY="https://proxy.golang.org|direct")
checks={
 "parallel":("./cmd/adamic-gate","^TestEveryTestIsParallelOrSaysWhy$"),
 "maps":("./internal/ir","^TestEveryMapAProgramHoldsIsSorted$"),
 "runtime-sets":("./internal/oracle","^TestRuntimeSetsAreEachUsed$"),
 "agreement-registry":("./internal/lower","^TestNativeAgreementRowsNameRegisteredPrograms$"),
 "readers":("./internal/ir","^TestCallTargetReaders"),
 "reader-products":("./internal/ir","^Test(CallTargetExportInputs|Product_CallTargetReaderArchives)"),
 "oracle":("./internal/oracle","^Test(CallbackWidening|WeakDifferential|Miscompile2A|SelfCompare)"),
 "native":("./internal/native","^Test(ErrorUndefinedMessageHasNoOwnProperty|SplitSelfCompareAgreesWithNode)$"),
 "ir":("./internal/ir","^Test(CallbackAdapterTargets|ClosureTargetsBoundOnlyProvenValues)$"),
 "flow":("./internal/flow","^Test(CallbackWidening|Slice2)"),
 "fresh":("./internal/fresh","^TestCallbackWidening"),
 "fixtures":("./stage3/fixtures","^Test(FixturesSelfCompare|FixtureDirectoriesHaveTopLevelTests)"),
 "runtime-products":("./internal/oracle","^TestProduct_NativeRuntimes(Sanitized|Release|Counted)"),
}
for variant in ["Sanitized","Release","Counted"]:
 for group in range(3):
  checks["runtime-"+variant+str(group)]=("./internal/oracle","^TestProduct_NativeRuntimes"+variant+str(group)+"$")
for package in ["Native","Lower","Fresh","Flow"]:
 checks["reader-key-"+package]=("./internal/ir","^TestCallTargetExportInputs"+package+"$")
 checks["reader-archive-"+package]=("./internal/ir","^TestProduct_CallTargetReaderArchives"+package+"$")
checks["native-guards"]=("./internal/native","^Test(RuntimeFieldLayoutsAreIncluded|SplitBuildsNameNoPlaceOfTheirInputs)$")
checks["flow-corpus"]=("./internal/flow","^TestFlowCorpus")
checks["lower"]=("./internal/lower","^Test(LibraryRegexOffsetRequiresNoCaptures|ClassStaticInitializerCallIsEmitted|LibrarySpreadErrorView|NestedCallbackCycleIsRefused|ParameterPropertyCallbackReceiver|PredicateCallbackContracts|EveryNeedsCallbackEffects|ReadinessErrorIncludesReceiverExpression|InheritedLibraryReadsNeverLoadOwnFieldsError)$")
checks["primitive-main"]=("./internal/native","^TestOrdinaryToPrimitive$")
checks["slot-pairs-main"]=("./internal/native","^TestParallelSlotPairs$/^(asan|release)$")
checks["counts-table"]=("./internal/oracle","^TestCountsAreRecordedTable$")
for name in sys.argv[1:]:
 if name=="vet": command=["go","vet","./internal/lower","./internal/native","./internal/ir","./internal/flow","./internal/fresh","./internal/oracle","./stage3/fixtures"]
 elif name=="counts": command=["go","test","./internal/oracle","-run","TestCountsAreRecorded","-count=1","-v","-timeout=600s","-args","-update-counts"]
 else:
  package,pattern=checks[name]
  command=["go","test",package,"-run",pattern,"-count=1","-json","-timeout=90s","-parallel=4"]
 limit=660 if name=="counts" else 300
 start=time.monotonic()
 with (out/(name+".log")).open("w") as log:
  ran=subprocess.run(["timeout",str(limit)]+command,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=limit+10)
 row=dict(name=name,command=command,exit=ran.returncode,seconds=round(time.monotonic()-start,3))
 with (out/"checks.jsonl").open("a") as f: f.write(json.dumps(row)+"\n")
 print(json.dumps(row),flush=True)
 if ran.returncode: sys.exit(1)
 if name not in ["vet","counts"]:
  events=[json.loads(line) for line in (out/(name+".log")).read_text().splitlines() if line.startswith("{")]
  leaves=[dict(test=e["Test"],seconds=e.get("Elapsed",0),action=e["Action"]) for e in events if e.get("Test") and e["Action"] in ["pass","fail"]]
  (out/(name+"-seconds.json")).write_text(json.dumps(leaves,indent=2)+"\n")
  overdue=[e for e in leaves if e["seconds"]>=60]
  if overdue:
   print("OVER BUDGET: "+json.dumps(overdue),flush=True);sys.exit(2)
