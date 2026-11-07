import json,sys
sys.dont_write_bytecode=True
from supervise import run,s,repo
assert json.loads((s/'raw/reuse-validation.json').read_text())['input_differences']==[]
assert json.loads((s/'raw/positive-controls.json').read_text())['passed']
assert json.loads((s/'raw/coverage-control.json').read_text())['passed']
number=0
while True:
 summary=s/'evidence/runtime-counts.json'
 if summary.exists() and json.loads(summary.read_text())['complete']:break
 number+=1;run('lane-batch-'+str(number),['python3',repo/'stage3/census/getfunctionname/lane.py','8'])
print('RUNTIME MEASUREMENT COMPLETE',flush=True)
