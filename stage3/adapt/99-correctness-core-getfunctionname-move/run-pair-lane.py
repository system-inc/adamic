import sys,json
sys.dont_write_bytecode=True
from run import run,s
assert json.loads((s/'raw/pair-positive-control.json').read_text())['passed']
assert json.loads((s/'raw/pair-negative-control.json').read_text())['passed']
for number in range(1,4):
 run('pair-lane-batch-'+str(number),['python3',s/'pair-lane.py','8'])
 if json.loads((s/'evidence/runtime-counts.json').read_text())['complete']:break
assert json.loads((s/'evidence/runtime-counts.json').read_text())['complete']
print('WHOLE PAIRED LANE COMPLETE',flush=True)
