from pathlib import Path
import json,subprocess,os,time,difflib,shutil
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';original=json.loads(Path('/tmp/u089/originals.json').read_text());f='stage1/cohere/estree/main.ts';s=original[f];header='function run(path: string): void {';start=s.index(header);end=s.index('\n}\nconst args',start)+2;changed=s[:start]+header+'\n    return;\n}'+s[end:];(E/'diffs/P2.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)));live=(root/f).read_text();live=live.replace(header,header+"\n    if(auditMutant === 'P2') { return; }",1);(root/f).write_text(live)
scratch=Path('/tmp/u089/replay/P2');source=scratch/'stage1/cohere/estree';source.mkdir(parents=True,exist_ok=True);dep=scratch/'stage1/typescript'
if not dep.exists():dep.symlink_to(root/'stage1/typescript',target_is_directory=True)
for p in(root/'stage1/cohere/estree').glob('*.ts'):
 if p.name=='auditSelector.ts':continue
 (source/p.name).write_text(original.get(str(p.relative_to(root)),p.read_text()))
subprocess.run(['git','apply','--check',str(E/'diffs/P2.diff')],cwd=scratch,check=True);subprocess.run(['git','apply',str(E/'diffs/P2.diff')],cwd=scratch,check=True);env=os.environ.copy();env.update(ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='4',ADAMIC_BUILD_CACHE_DIR='/tmp/u089/cache/replay-P2');cmd=['timeout','90','go','run','./cmd/adamic','build',str(source/'main.ts'),'-o',str(scratch/'port'),'--sanitize'];t=time.monotonic()
with(E/'P2-native-build.log').open('w')as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
p=E/'replay-checks.json';checks=json.loads(p.read_text());checks['P2']=dict(exit=r.returncode,seconds=round(time.monotonic()-t,3),command=cmd);p.write_text(json.dumps(checks,indent=2))
env.pop('ADAMIC_BUILD_CACHE_DIR');env.update(ADAMIC_ESTREE_LIBRARY='/tmp/u089/library',ADAMIC_ESTREE_BENCHMARK='1');Path('/tmp/u089/mutant').write_text('P2');p=E/'matrix.json';matrix=json.loads(p.read_text());matrix['P2']={}
for row in ['TestSyntaxGrammar','TestSyntaxRefusals','TestThroughput','TestCookedSurrogates','TestLossyInputRefusal']:
 log=E/f'P2-{row}.log';t=time.monotonic()
 with log.open('w')as out:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/estree/','-run','^'+row+'$'],env=env,stdout=out,stderr=subprocess.STDOUT)
 data=[]
 for l in log.read_text().splitlines():
  try:data.append(json.loads(l))
  except:pass
 terminal=[d for d in data if d.get('Test')==row and d.get('Action')in('pass','fail','skip')];cooked='panic: test timed out'in log.read_text()or r.returncode==124;matrix['P2'][row]=dict(result=terminal[-1]['Action']if terminal and not cooked else'unknown',seconds=round(time.monotonic()-t,3),exit=r.returncode,cooked=cooked);p.write_text(json.dumps(matrix,indent=2));print(row,matrix['P2'][row],flush=True)
