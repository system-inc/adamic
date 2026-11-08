import sys
sys.dont_write_bytecode=True
from run import *
assert run('prepare-go-product',['go','test','./stage3/lane/adaptedtree','-count=1','-parallel=2','-timeout=25m','-run','^TestProduct_Stage3AdaptedTree$','-v'])==0
code=run('go-test-whole',['go','test','./stage3/lane/...','-count=1','-p=1','-parallel=2','-timeout=25m','-v'])
raise SystemExit(code)
