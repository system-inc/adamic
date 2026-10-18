from pathlib import Path
import subprocess,json,time,sys
out=Path('review/compiler/loom-r-fx3-refusals')
results=[]
def run(label,args,limit=180):
    began=time.monotonic()
    with (out/(label+'.log')).open('w') as log:
        try: code=subprocess.run(args,stdout=log,stderr=subprocess.STDOUT,timeout=limit).returncode
        except subprocess.TimeoutExpired: code=124
    row=dict(label=label,command=args,exit=code,seconds=round(time.monotonic()-began,3))
    results.append(row);(out/'validation.json').write_text(json.dumps(results,indent=2)+'\n')
    print(json.dumps(row),flush=True)
    if code: sys.exit(code)
def test(label,pkg,pattern,extra=[]):
    run(label,['go','test',pkg,'-run',pattern,'-count=1','-v','-timeout','90s','-parallel','4']+extra)
test('lower-focus','./internal/lower','^(TestArrayNarrowing.*|TestFX5.*|TestNativeAgreement.*|TestAgreement.*|TestProduct_LowerAgreementNative)$')
test('ir-products','./internal/ir','^TestProduct_CallTargetExports$')
test('ir-guards','./internal/ir','^(TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted)$')
test('oracle-predicates','./internal/oracle','^(TestPredicateMiscompileRefusals|TestPredicateDirectionCountsAreRecorded|TestCountsAreRecordedTable|TestRuntimeSetsAreEachUsed)$')
test('gate-guards','./cmd/adamic-gate','^(TestEveryTestIsParallelOrSaysWhy|TestEveryDuplicatedProductDeclarationAgrees)$')
run('vet',['go','vet','./internal/lower'])
run('packages',['go','list','-json','./...'])
run('cli-build',['go','build','-o',str(out/'adamic'),'./cmd/adamic'])
