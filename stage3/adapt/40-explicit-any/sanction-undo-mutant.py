#!/usr/bin/env python3
"""The real control API, with the retype undone, must fail the undefined sanction."""
import json,os,subprocess,sys
from pathlib import Path
sys.dont_write_bytecode=True
stage=Path(__file__).resolve().parents[2]
sys.path.insert(0,str(stage/'lane'))
from check import api_check
os.environ['TSC_ADAPT_TYPESCRIPT']=str(stage/'api/node_modules/typescript')
mode,control=sys.argv[1],Path(sys.argv[2])
sanction=json.loads((stage/'lane/sanctioned-api.json').read_text())
if mode=='check':
 errors,_=api_check(control,sanction)
 for error in errors:print(error)
 raise SystemExit(bool(errors))
assert mode=='prove' and len(sys.argv)==4
base=json.loads((Path(__file__).parent/'evidence/brands-undefined-2/adapt90-pin.json').read_text())['base']
baseline=json.loads(subprocess.check_output(['git','show',base+':stage3/lane/sanctioned-api.json'],cwd=stage.parent,text=True))
errors,_=api_check(control,baseline)
assert not errors,errors
mutant=subprocess.run([sys.executable,__file__,'check',str(control)],capture_output=True,text=True)
assert mutant.returncode==1,(mutant.returncode,mutant.stderr)
lines=mutant.stdout.splitlines()
changed=[line for line in lines if line.startswith('changed sanctioned API declaration:')]
assert len(changed)==27,lines
red=next(line for line in changed if '["ts","Path"]' in line)
output=Path(sys.argv[3]);output.mkdir(parents=True,exist_ok=True)
(output/'api-sanction-undo-mutant.json').write_text(json.dumps({'test':'test_undo_retype_rejected_by_lane_api_sanction','guard':'stage3/lane/check.py:api_check','mutation':'use the real main control API (all 27 public brands restored to void) against the undefined sanction','control_guard_errors':0,'mutant_exit':mutant.returncode,'changed_declarations':len(changed),'red_line':red,'errors':lines},indent=2)+'\n')
(output/'api-sanction-undo-mutant.txt').write_text(mutant.stdout)
print(red)
