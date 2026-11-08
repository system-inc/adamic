import sys
sys.dont_write_bytecode=True
from run import *
assert run('test-check-fixed',['go','test','./stage3/lane','-count=1','-parallel=2','-timeout=10m','-run','^TestLaneCheckPython$','-v'])==0
assert run('test-shards',['go','test','./stage3/lane','-count=1','-parallel=2','-timeout=10m','-run','^TestLaneShardsPython$','-v'])==0
