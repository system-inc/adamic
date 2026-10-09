from pathlib import Path
import subprocess,os,json,difflib,time
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';(E/'witness-diffs').mkdir(exist_ok=True)
f='stage1/cohere/estree/estree_test.go';s=(root/f).read_text();header='func firstDifference(want, got []byte) string {';start=s.index(header);body=start+len(header);end=body;depth=1
while depth:
 if s[end]=='{':depth+=1
 elif s[end]=='}':depth-=1
 end+=1
changed=s[:body]+'\n\treturn ""\n}'+s[end:];replacement=E/'weakened-estree_test.go.txt';replacement.write_text(changed);overlay=E/'witness-overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/f):str(replacement)}}));(E/'witness-diffs/W1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/u089/library',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4',ADAMIC_THREE_PORT_PROOF='-1');runs=[('family-control',False,'^TestThreePortMutants_[0-9]{3}$'),('W1-family-and-proof',True,'^(TestThreePortMutants_[0-9]{3}|TestThreePortMutants_ShardProof)$'),('W1-cooked-witness',True,'^TestCookedSurrogateMutant$'),('W1-lossy-witness',True,'^TestLossyInputControl$')];results={}
for name,weaken,pattern in runs:
 cmd=['timeout','120','go','test'];cmd+=['-overlay='+str(overlay)]if weaken else[];cmd+=['-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run',pattern];t=time.monotonic()
 with(E/(name+'.log')).open('w')as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 results[name]=dict(command=cmd,exit=r.returncode,seconds=round(time.monotonic()-t,3));(E/'witness-results.json').write_text(json.dumps(results,indent=2));print(name,results[name],flush=True)
