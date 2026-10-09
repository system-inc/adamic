from pathlib import Path
import subprocess,json,difflib,os
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';f='stage1/cohere/estree/three_port_mutants_deadline_shards_test.go';s=(root/f).read_text();start=s.index('\tthreePortShared.once.Do(func() {');end=s.index('\n\tif threePortShared.prepared == nil',start);changed=s[:start]+s[end:];replacement=E/'broken-setup.go.txt';replacement.write_text(changed);overlay=E/'setup-overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/f):str(replacement)}}));(E/'witness-diffs/W2.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
env=os.environ.copy();env.update(ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4')
with(E/'W2-setup.log').open('w')as out:r=subprocess.run(['timeout','120','go','test','-overlay='+str(overlay),'-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^TestThreePortMutants_Setup$'],cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
print('exit',r.returncode)
