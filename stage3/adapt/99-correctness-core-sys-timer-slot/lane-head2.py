import sys
sys.dont_write_bytecode=True
from run import *
assert run('head2-full-lane',['python3','-B',ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot/run-shards.py',S/'head2',S/'lane-head2'])==0
