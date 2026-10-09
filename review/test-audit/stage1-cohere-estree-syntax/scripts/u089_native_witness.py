from pathlib import Path
import json,subprocess,shutil
root=Path('/workspace/adamic');E=root/'review/test-audit/stage1-cohere-estree-syntax';original=json.loads(Path('/tmp/u089/originals.json').read_text());scratch=Path('/tmp/u089/native-witness');source=scratch/'stage1/cohere/estree';source.mkdir(parents=True,exist_ok=True);dep=scratch/'stage1/typescript'
if not dep.exists():dep.symlink_to(root/'stage1/typescript',target_is_directory=True)
for f in(root/'stage1/cohere/estree').glob('*.ts'):
 if f.name=='auditSelector.ts':continue
 (source/f.name).write_text(original.get(str(f.relative_to(root)),f.read_text()))
result={}
for mid,text in [('M1','//\ufffd\nx;'),('M2',"x;\n/// <reference path='broken.ts />"),('M3','~x;'),('M4','x;'),('P1','x;')]:
 fixture=scratch/(mid+'.ts');fixture.write_text(text);commands={'before':['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(source/'main.ts'),str(fixture)],'after':[str(Path('/tmp/u089/replay')/mid/'port'),str(fixture)]};seen={}
 for phase,cmd in commands.items():
  p=subprocess.run(cmd,capture_output=True,text=True);(E/f'native-witness-{mid}-{phase}.stdout').write_text(p.stdout);(E/f'native-witness-{mid}-{phase}.stderr').write_text(p.stderr);seen[phase]=dict(exit=p.returncode,stdout=p.stdout,stderr=p.stderr,command=cmd)
 result[mid]=seen
(E/'native-behavior-witnesses.json').write_text(json.dumps(result,indent=2));print({k:(v['before']['exit'],v['after']['exit'],v['before']['stdout']!=v['after']['stdout'])for k,v in result.items()})
