import sys
sys.dont_write_bytecode=True
from run import *
assert run('go-test-whole-retry',['go','test','./stage3/lane/...','-count=1','-p=1','-parallel=3','-timeout=29m','-v'])==0
