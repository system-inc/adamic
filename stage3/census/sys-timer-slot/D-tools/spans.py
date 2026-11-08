import sys
sys.dont_write_bytecode=True
from run import *
for label in ['H','D']:assert run(label+'-spans',['node',S/'spans.cjs',S/label])==0
