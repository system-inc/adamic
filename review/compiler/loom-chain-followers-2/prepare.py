import difflib,json,subprocess
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-chain-followers-2';old=root/'review/compiler/loom-chain-followers';source='376febb44da1cc8018f1382f0bc488941375fcdc'
focus=json.loads((old/'focus.json').read_text())
focus['tests']['internal/lower'] += ['TestArrowFieldCaptureFrames','TestArrowSuperCaptureFrames','TestArrowStoredCaptureFrames','TestClassFeaturesThrowingAccessorSpread','TestNestedEmptyArrayElementKinds']
focus['tests']['internal/oracle'] += ['TestCheckedNonNullTypeScript','TestDebuggerNativeEmitsNothing']
focus['tests']['internal/ir']=['TestObjectSpreadAccessorEffects']
focus['main_witnesses']=['internal/oracle/testdata/accessor_spread_throw_'+p+'.a' for p in ['first','middle','last']]+['internal/oracle/testdata/plain_data_spread.a','internal/oracle/testdata/static_constructor_spread_gap.a','internal/oracle/testdata/static_constructor_spread_owned.a','internal/oracle/testdata/template_nullish.a']
(out/'focus.json').write_text(json.dumps(focus,indent=2))
(out/'main-base.txt').write_text(subprocess.check_output(['git','rev-parse','origin/main'],text=True))
(out/'counts-main.md.txt').write_text(Path('internal/oracle/counts.md').read_text())
(out/'exceptions-landed.patch').write_text(subprocess.check_output(['git','diff','c0fd8945','origin/main','--','internal/lower/exceptions.go'],text=True))
(out/'mutants').mkdir(exist_ok=True)
cases=json.loads((old/'mutants/cases.json').read_text())
for c in cases:
 target=c['file'];base=subprocess.check_output(['git','show',source+':'+target],text=True);mutant=(old/'mutants'/(c['topic']+'.go.txt')).read_text();current=(root/target).read_text();original=current
 for tag,a,b,x,y in difflib.SequenceMatcher(None,base.splitlines(True),mutant.splitlines(True),autojunk=False).get_opcodes():
  if tag=='equal':continue
  before=''.join(base.splitlines(True)[a:b]);after=''.join(mutant.splitlines(True)[x:y])
  if not before:
   anchor=''.join(base.splitlines(True)[max(0,a-1):a]);assert current.count(anchor)==1,(c['topic'],anchor);current=current.replace(anchor,anchor+after,1)
  else:
   assert current.count(before)==1,(c['topic'],before);current=current.replace(before,after,1)
 assert current!=original,c['topic']
 path=out/'mutants'/(c['topic']+'.go.txt');path.write_text(current)
 patch=out/'mutants'/(c['topic']+'.patch');patch.write_text(''.join(difflib.unified_diff(original.splitlines(True),current.splitlines(True),fromfile=target,tofile=target)))
 overlay=out/'mutants'/(c['topic']+'.json');overlay.write_text(json.dumps({'Replace':{str(root/target):str(path)}}));c['overlay']=str(overlay)
(out/'mutants/cases.json').write_text(json.dumps(cases,indent=2))
print('Prepared 14 source mutants as deltas on the reconciled tree; main changes retained.')
