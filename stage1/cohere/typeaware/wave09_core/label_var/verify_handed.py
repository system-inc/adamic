"""Prove label visits use the supplied node rather than the callback index."""
import pathlib,subprocess,json
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-core');probe=work/'handed-probe';probe.mkdir(exist_ok=True)
def run(name,args):
 with (probe/(name+'.stdout')).open('wb') as out,(probe/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=repo,stdout=out,stderr=err,timeout=120)
 assert result.returncode==0,(name,(probe/(name+'.stderr')).read_text())
 stderr=(probe/(name+'.stderr')).read_bytes()
 if name=='go':assert stderr.startswith(b'cohere: load_ns=')
 else:assert not stderr,name
 print(name,result.returncode,flush=True)
 return (probe/(name+'.stdout')).read_bytes()
def imports(text):
 return text.replace("from '../../../../typescript/","from '"+str(repo/'stage1/typescript')+'/').replace("from '../../","from '"+str(repo/'stage1/cohere/typeaware')+'/')
main=imports((root/'main.a').read_text()).replace("from './no_label_var.a'","from '"+str(root/'no_label_var.a')+"'")
assert main.count('rule.visit(node, index)')==1
main=main.replace('rule.visit(node, index)','rule.visit(node, 0)')
source=probe/'main.a';source.write_text(main)
args=[work/'tsconfig.json',work/'controls.manifest']
truth=run('go',[work/'oracle',*args])
for mode in ['normal','sanitized','mutant']:
 current=source
 if mode=='mutant':
  changed=imports((root/'no_label_var.a').read_text()).replace("from './scope_value_symbols.a'","from '"+str(root/'scope_value_symbols.a')+"'")
  assert changed.count('const r = this.rules;')==1
  changed=changed.replace('const r = this.rules;','const r = this.rules; node = r.parser.node(_index);')
  module=probe/'mutant.a';module.write_text(changed)
  current=probe/'mutant-main.a';current.write_text(main.replace(str(root/'no_label_var.a'),str(module)))
 binary=probe/mode
 build=[work/'adamic','build',current,'-o',binary,'--tsgo',work/('checker-asan.a' if mode=='sanitized' else 'checker.a')]
 if mode=='sanitized':build.append('--sanitize')
 run(mode+'-build',build);actual=run(mode,[binary,*args])
 if mode=='mutant':
  assert actual!=truth
  difference=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)))
 else:assert actual==truth
(probe/'results.json').write_text(json.dumps({'bytes':len(truth),'mutant_difference':difference,'handed_index':0},indent=2)+'\n')
print('PASS callback index zero preserves Go bytes; refetch mutant builds/exits zero and only Go comparison catches it',flush=True)
