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
if mode == 'resolved-prove':
 import difflib, tempfile
 assert len(sys.argv) == 5, 'resolved-prove CONTROL_TREE TREATMENT_TREE EVIDENCE'
 treatment, output = Path(sys.argv[3]), Path(sys.argv[4])
 output.mkdir(parents=True, exist_ok=True)
 key = 'TypeAliasDeclaration:["ts","ResolvedConfigFileName"]#0'
 old = (control/'built/local/typescript.d.ts').read_text()
 new = (treatment/'built/local/typescript.d.ts').read_text()
 delta = list(difflib.unified_diff(old.splitlines(True), new.splitlines(True), fromfile='control/typescript.d.ts', tofile='treatment/typescript.d.ts'))
 removed = [line[1:].strip() for line in delta if line.startswith('-') and not line.startswith('---')]
 added = [line[1:].strip() for line in delta if line.startswith('+') and not line.startswith('+++')]
 assert removed == ['_isResolvedConfigFileName: never;'], removed
 assert added == ['_isResolvedConfigFileName: undefined;'], added
 baseline = dict(sanction, normalized_changes=[row for row in sanction['normalized_changes'] if row['declaration'] != key])
 assert len(baseline['normalized_changes']) + 1 == len(sanction['normalized_changes'])
 observations = {}
 for label, tree, api, manifest in [('control', control, old, baseline), ('treatment', treatment, new, sanction)]:
  local = tree/'tests/baselines/local/api/typescript.d.ts'
  previous = local.read_bytes() if local.exists() else None
  local.parent.mkdir(parents=True, exist_ok=True)
  try:
   with tempfile.TemporaryDirectory(prefix='resolved-api-guard-') as scratch:
    view = Path(scratch)
    (view/'adapted-tree').symlink_to(tree.resolve(), target_is_directory=True)
    (view/'oracle').mkdir()
    reference = (tree/'tests/baselines/reference/api/typescript.d.ts').read_text()
    def write_api(text):
     local.write_text(text)
     (view/'oracle/baseline.diff').write_text(''.join(difflib.unified_diff(reference.splitlines(True), text.splitlines(True), fromfile='reference/api/typescript.d.ts', tofile='local/api/typescript.d.ts')))
    write_api(api)
    errors, counts = api_check(view, manifest)
    assert not errors, errors
    observations[label] = dict(errors=errors, counts=counts)
    if label == 'treatment':
     assert api.count('_isResolvedConfigFileName: undefined;') == 1
     write_api(api.replace('_isResolvedConfigFileName: undefined;', '_isResolvedConfigFileName: never;'))
     mutant = subprocess.run([sys.executable, __file__, 'check', str(view)], capture_output=True, text=True)
     assert mutant.returncode == 1, (mutant.returncode, mutant.stderr)
     assert mutant.stdout.splitlines() == ['missing sanctioned API declaration: ' + key], mutant.stdout
     observations['undo'] = dict(exit=mutant.returncode, errors=mutant.stdout.splitlines())
  finally:
   if previous is None:
    local.unlink(missing_ok=True)
   else:
    local.write_bytes(previous)
 (output/'resolved-api.diff').write_text(''.join(delta))
 (output/'resolved-api-guard.json').write_text(json.dumps(dict(guard='stage3/lane/check.py:api_check', changedPublicLines=1, declaration=key, **observations), indent=2)+'\n')
 print(json.dumps(observations, indent=2))
 raise SystemExit(0)
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
