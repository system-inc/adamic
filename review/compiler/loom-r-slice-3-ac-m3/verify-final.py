import sys
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parent))
# run-suites only dispatches its CLI in all mode.
sys.argv=['verify','helper']
exec((Path(__file__).resolve().parent/'run-suites.py').read_text())
def check(name,cmd,limit=840):
 row=run(name,cmd,limit)
 if row['exit']:raise SystemExit(name+' failed')
for selected in ['Flow']:
 check('reader-product-repaired-'+selected,['go','test','./internal/ir','-run','^TestProduct_CallTargetReaderArchives'+selected+'$','-count=1','-timeout=90s','-json','-parallel=1'],120)
check('reader-guard-repaired',['go','test','./internal/ir','-run','TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted','-count=1','-timeout=90s','-json'],120)
check('oracle-guards',['go','test','./internal/oracle','-run','TestRuntimeSetsAreEachUsed|TestOptionalAfterCall|TestMergedField|TestFunction(TypeScript|Adamic)|TestCallSpreadArray|TestJSONShorthand','-count=1','-timeout=6m','-parallel=4','-json'],390)
check('flow-guard',['go','test','./internal/flow','-run','TestOptionalAfterCallFlowEdges','-count=1','-timeout=90s','-json'],120)
check('vet',['go','vet','-p','1','./internal/native/...','./internal/lower/...','./internal/ir/...','./internal/flow/...','./internal/javascript/...','./internal/oracle/...'],180)
check('a-check',['python3',str(out/'scan.py'),'after'],180)
check('header-mutants',['python3',str(out/'scan.py'),'mutants'],180)
check('production-mutants',['python3',str(out/'run-mutants.py')],600)
