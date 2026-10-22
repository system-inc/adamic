import sys
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parent))
# run-suites only dispatches its CLI in all mode.
sys.argv=['verify','helper']
exec((Path(__file__).resolve().parent/'run-suites.py').read_text())
def check(name,cmd,limit=840):
 row=run(name,cmd,limit)
 if row['exit']:raise SystemExit(name+' failed')
for p in ['internal/native','internal/lower']:
 label=p.replace('/','-')
 check(label+'-products',['go','test','./'+p,'-run','^TestProduct_','-count=1','-json','-parallel=1','-timeout=13m','-p','1'])
 check(label,['go','test','./'+p+'/...','-skip','^TestProduct_','-count=1','-json','-parallel=4','-timeout=13m','-p','1'])
check('counts-products',['go','test','./internal/oracle','-run','^TestProduct_CountMeasurements','-count=1','-json','-parallel=4','-timeout=13m','-p','1'])
check('counts-update',['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-json','-parallel=4','-timeout=13m','-p','1','-args','-update-counts'])
check('reader-guard',['go','test','./internal/ir','-run','TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted','-count=1','-timeout=90s','-json'],120)
check('oracle-guards',['go','test','./internal/oracle','-run','TestRuntimeSetsAreEachUsed|TestOptionalAfterCall|TestMergedField|TestFunction(TypeScript|Adamic)|TestCallSpreadArray|TestJSONShorthand','-count=1','-timeout=6m','-parallel=4','-json'],390)
check('flow-guard',['go','test','./internal/flow','-run','TestOptionalAfterCallFlowEdges','-count=1','-timeout=90s','-json'],120)
check('vet',['go','vet','-p','1','./internal/native/...','./internal/lower/...','./internal/ir/...','./internal/flow/...','./internal/javascript/...','./internal/oracle/...'],180)
check('a-check',['python3',str(out/'scan.py'),'after'],180)
check('header-mutants',['python3',str(out/'scan.py'),'mutants'],180)
check('production-mutants',['python3',str(out/'run-mutants.py')],600)
