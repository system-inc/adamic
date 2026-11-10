from pathlib import Path
import shutil,json,difflib
before=Path('/workspace/scratch/debug-display-before');after=Path('/workspace/scratch/debug-display-candidate');unit=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-t');pins=json.loads((unit/'pins.json').read_text());legacy=json.loads((unit/'debug-display-previous-enable.json').read_text())
shutil.copytree(before,after,ignore=lambda d,n:[x for x in ['.git','built','node_modules','tests'] if Path(d)==before and x in n])
(after/'node_modules').symlink_to('/workspace/adamic/stage3/typescript/node_modules',target_is_directory=True);(after/'tests').symlink_to(before/'tests',target_is_directory=True);(after/'.git').symlink_to(before/'.git',target_is_directory=True)
p=after/'src/compiler/debug.ts';text=p.read_bytes().decode();original=text
for pin in pins:
 if pin['file']!='debug.ts' or pin['owner'] not in ['Debug.attachFlowNodeDebugInfoWorker','Debug.attachFlowNodeDebugInfo','Debug.attachNodeArrayDebugInfoWorker','Debug.attachNodeArrayDebugInfo','Debug.enableDebugInfo']:continue
 old=legacy['after'] if pin['owner']=='Debug.enableDebugInfo' else pin['before'];assert text.count(old)==1,pin['owner'];text=text.replace(old,pin['after'])
p.write_bytes(text.encode());(unit/'debug-display-source.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),text.splitlines(True),fromfile='before/src/compiler/debug.ts',tofile='after/src/compiler/debug.ts')))
mutant=Path('/workspace/scratch/debug-display-mutant');mutant.mkdir();shutil.copytree(after/'src',mutant/'src')
flow=next(p for p in pins if p['owner']=='Debug.attachFlowNodeDebugInfo');left=flow['before'].split('\r\n');right=flow['after'].split('\r\n');i=next(i for i,s in enumerate(left) if 'flowNodeProto = Object.create(Object.prototype) as FlowNode;' in s);assert right[i]=='';right[i]=left[i];changed=text.replace(flow['after'],'\r\n'.join(right));assert changed!=text
(mutant/'src/compiler/debug.ts').write_bytes(changed.encode());lines=changed.splitlines();assert 'Object.create(Object.prototype) as FlowNode' in lines[524];assert lines[524].index('Object.create')+1==37
(unit/'debug-display-mutant.patch').write_text(''.join(difflib.unified_diff(text.splitlines(True),changed.splitlines(True),fromfile='candidate/src/compiler/debug.ts',tofile='mutant/src/compiler/debug.ts')))
print('candidate and one-line mutant prepared; mutant cast at 525:37')
