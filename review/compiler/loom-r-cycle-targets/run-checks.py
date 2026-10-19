import json, os, pathlib, shlex, subprocess, time, sys
root = pathlib.Path(__file__).resolve().parents[3]
evidence = pathlib.Path(__file__).resolve().parent
os.chdir(root)
records = json.loads((evidence/'commands.json').read_text()) if (evidence/'commands.json').exists() else []
def run(name, argv, seconds=180):
    command=['timeout',str(seconds),*argv]
    began=time.monotonic()
    with (evidence/(name+'.log')).open('w') as out:
        p=subprocess.run(command,stdout=out,stderr=subprocess.STDOUT,timeout=seconds+10)
    record=dict(name=name,command=shlex.join(command),exit=p.returncode,seconds=round(time.monotonic()-began,3))
    records.append(record)
    (evidence/'commands.json').write_text(json.dumps(records,indent=2)+'\n')
    print(json.dumps(record),flush=True)
    if p.returncode: sys.exit(p.returncode)
def test(name, packages, selector, limit=90, outer=180, extra=()):
    run(name,['go','test',*packages,'-run',selector,'-count=1','-parallel=4','-json','-timeout',str(limit)+'s',*extra],outer)
mode=sys.argv[1]
if mode=='focused':
    test('focused',['./internal/ir','./internal/lower','./internal/oracle','./cmd/adamic-gate'],r'^(TestEveryMapAProgramHoldsIsSorted|TestEveryTestIsParallelOrSaysWhy|TestNativeAgreementRowsNameRegisteredPrograms|TestRuntimeSetsAreEachUsed|TestClosureTargets.*|TestNamespaceAmbient.*|TestClosureCycle.*|TestGettersCensus.*)$',outer=300)
elif mode=='counts':
    test('counts-update',['./internal/oracle'],'TestCountsAreRecorded',570,600,('-args','-update-counts'))
elif mode=='readers':
    test('readers',['./internal/ir'],r'^(TestCallTargetReaders|TestCallTargetsIncludeEveryDescendant|TestEveryMapAProgramHoldsIsSorted|TestAProgramIsTheSameBytesEveryTimeGobWritesIt|TestASortedMapIsNilWhereAPlainMapIs)$')
    run('vet',['go','vet','./internal/ir','./internal/lower','./internal/oracle'],180)
    test('consumers',['./internal/native','./internal/flow','./internal/fresh','./internal/javascript'],r'^(TestCallTargetsElementBorrowPlan|TestBorrowChainTargets|TestLoopCallCoverage|TestDebuggerHasNoFlowInstruction|TestRegexOperandsStillJudgeCycleWrites|TestViewMixedUnionSelectsNumber)$')
    os.environ['ADAMIC_UNIT_BUDGET']='1'
    test('flow-corpus',['./internal/flow'],r'^(TestFlowCorpus.*|TestFlowProgram.*|TestFlowSingleAssignment.*|TestFlowMutationRanges.*|TestFlowGraphPaths.*|TestFlowLiveness.*)$',570,600)
    test('table-and-families',['./internal/oracle'],r'^(TestCountsAreRecordedTable|TestFamiliesOwnEveryCaseOnce|TestFamilyWrappersAreCurrent)$')
elif mode=='affected':
    os.environ['ADAMIC_GATE_UNCACHED']='1'
    test('affected-fixtures',['./internal/oracle'],r'^(TestNativeAgreesWithNode_[0-9]+|TestNativeAgreesWithNode_user_iterators)$/^internal$/^oracle$/^testdata$/^(route_targets_sort_values|borrow_chain_unknown|call_targets_element|generic_functions|instantiation_find|user_iterators)\.a$')
elif mode=='base-counts-fixture':
    os.environ['ADAMIC_GATE_UNCACHED']='1'
    test('instantiation-find',['./internal/oracle'],'^TestInstantiationFind$')
elif mode=='mutants':
    run('mutants',['python3',str(evidence/'run-mutants.py')],600)
    run('admission-mutant',['python3',str(evidence/'run-admission-mutant.py')],180)
elif mode=='lower':
    run('lower-list',['go','test','./internal/lower','-list','^Test','-timeout','90s'],120)
    tests=[line for line in (evidence/'lower-list.log').read_text().splitlines() if line.startswith('Test')]
    # Each top-level lower test gets its own Go timeout. This is the requested
    # complete lowering coverage without an unbounded whole-package invocation.
    (evidence/'lower-tests.json').write_text(json.dumps(tests,indent=2)+'\n')
    for offset in range(0,len(tests),20):
        test('lower-shard-'+str(offset//20),['./internal/lower'],'^('+'|'.join(tests[offset:offset+20])+')$',90,120)
    print('lower top-level tests:',len(tests),flush=True)
else: raise SystemExit('unknown mode')
