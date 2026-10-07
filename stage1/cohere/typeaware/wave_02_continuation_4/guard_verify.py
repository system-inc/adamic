#!/usr/bin/env python3
"""Independent byte check for refusal of the shared parser's silent JSX misread."""
import argparse,json,subprocess
from pathlib import Path
parser=argparse.ArgumentParser()
parser.add_argument('--scratch',required=True)
parser.add_argument('--stage0',required=True)
parser.add_argument('--archive',required=True)
args=parser.parse_args()
own=Path(__file__).resolve().parent
root=own.parents[3]
scratch=Path(args.scratch).resolve()
commands=[]
def run(name,command,expected=0):
 with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:
  result=subprocess.run([str(x) for x in command],cwd=root,stdout=out,stderr=err)
 commands.append({'name':name,'args':[str(x) for x in command],'exit':result.returncode})
 (scratch/'guard-commands.json').write_text(json.dumps(commands,indent=2)+'\n')
 assert result.returncode==expected,(name,result.returncode)
 return (scratch/(name+'.stdout')).read_bytes()
source=scratch/'guard.tsx'
source.write_text('const foo = <button>Hi!</button>;\nexport {};\n')
manifest=scratch/'guard.manifest'
manifest.write_text(str(source)+'\n')
run('jsx-refusal',[scratch/'native',scratch/'tsconfig.json',manifest],70)
assert 'shared parser cannot represent JSX' in (scratch/'jsx-refusal.stderr').read_text()
truth=run('jsx-go',[scratch/'oracle',scratch/'tsconfig.json',manifest])
assert b'\treact/button-has-type\tmissingType\t' in truth
folder=scratch/'jsx-source'
folder.mkdir(exist_ok=True)
for file in own.glob('*.a'):
 text=file.read_text()
 if file.name=='suite.a':
  assert text.count("path.endsWith('.tsx')")==1
  text=text.replace("path.endsWith('.tsx')","false && path.endsWith('.tsx')",1)
 text=text.replace("from '../","from '"+str(own.parent)+'/').replace("from '../../../typescript/","from '"+str(root/'stage1/typescript')+'/')
 (folder/file.name).write_text(text)
run('jsx-mutant-build',[args.stage0,'build',folder/'suite.a','-o',scratch/'jsx-mutant','--tsgo',args.archive])
actual=run('jsx-mutant-run',[scratch/'jsx-mutant',scratch/'tsconfig.json',manifest])
assert not (scratch/'jsx-mutant-run.stderr').read_bytes()
assert truth!=actual
print('PASS: guard refuses JSX with panic 70; disabling it compiles and exits 0, but loses the independent Go button finding')
