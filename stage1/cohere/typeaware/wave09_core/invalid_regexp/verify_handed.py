"""Held flag frontend: use the handed call, not its callback index or a refetch."""
import pathlib,subprocess,json
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
base=pathlib.Path('/workspace/wave-09-core');fixtures=base/'flags-frontend'
work=pathlib.Path('/workspace/wave-09-handed-flags');work.mkdir(exist_ok=True)
def run(name,args):
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=120)
 assert result.returncode==0,(name,(work/(name+'.stderr')).read_text())
 if name!='go':assert not (work/(name+'.stderr')).read_bytes(),name
 print(name,result.returncode,flush=True)
 return (work/(name+'.stdout')).read_bytes()
def imports(text):
 return text.replace("from '../../../../typescript/","from '"+str(repo/'stage1/typescript')+'/').replace("from '../../","from '"+str(repo/'stage1/cohere/typeaware')+'/')
main=imports((root/'main.a').read_text())
for filename in ['no_invalid_regexp.a','rune_quote.a']:main=main.replace("from './"+filename+"'","from '"+str(root/filename)+"'")
assert main.count('rule.visit(node, index)')==1
main=main.replace('rule.visit(node, index)','rule.visit(node, 0)')
source=work/'main.a';source.write_text(main)
args=[fixtures/'tsconfig.json',fixtures/'manifest']
truth=run('go',[fixtures/'oracle',*args])
for mode in ['normal','sanitized','mutant']:
 current=source
 if mode=='mutant':
  changed=imports((root/'no_invalid_regexp.a').read_text())
  anchor='const r = this.rules;\n        r.findings.push'
  assert changed.count(anchor)==1
  changed=changed.replace(anchor,'const r = this.rules; node = r.parser.node(0);\n        r.findings.push')
  module=work/'mutant.a';module.write_text(changed)
  current=work/'mutant-main.a';current.write_text(main.replace(str(root/'no_invalid_regexp.a'),str(module)))
 binary=work/mode
 build=[base/'adamic','build',current,'-o',binary,'--tsgo',base/('checker-asan.a' if mode=='sanitized' else 'checker.a')]
 if mode=='sanitized':build.append('--sanitize')
 run(mode+'-build',build);actual=run(mode,[binary,*args])
 if mode=='mutant':
  assert actual!=truth
  difference=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
 else:assert actual==truth
(work/'results.json').write_text(json.dumps({'bytes':len(truth),'mutant_difference':difference,'handed_index':0},indent=2)+'\n')
print('PASS flag callback index zero retains Go bytes; report-refetch mutant builds/exits zero and only comparison catches it',flush=True)
