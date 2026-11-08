import sys
sys.dont_write_bytecode=True
from run import *
for name,pattern in [('test-generate-api','^TestLaneGenerateApi'),('test-check','^TestLaneCheckPython$'),('test-shards','^TestLaneShardsPython$')]:
 assert run(name,['go','test','./stage3/lane','-count=1','-parallel=2','-timeout=10m','-run',pattern,'-v'])==0
