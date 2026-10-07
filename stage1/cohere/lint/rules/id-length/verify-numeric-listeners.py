import pathlib,json,subprocess,tempfile
root=pathlib.Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip()); base=root/'stage1/cohere/lint/rules'
names=['array-callback-return','arrow-body-style','base-consistency-no-bare-throw','dot-notation','eslint-comments-require-description','grouped-accessor-pairs','id-length','structure-tailwind-no-physical-direction','typescript-no-non-null-asserted-optional-chain','typescript-no-non-null-assertion','typescript-no-restricted-types','typescript-no-this-alias']
work=pathlib.Path(tempfile.mkdtemp(prefix='wave02-listeners-'))
go='package main\nimport("fmt"; "strings"; ast "github.com/microsoft/TypeScript/tsc/shim/ast")\nfunc main(){for k:=ast.Kind(0); k<ast.KindCount; k++ {fmt.Printf("%s %d\\n",strings.TrimPrefix(k.String(),"Kind"),k)}}\n'
(work/'kinds.go').write_text(go)
values=dict(line.split() for line in subprocess.check_output(['go','run',str(work/'kinds.go')],cwd=root,text=True).splitlines())
runner=[];expected=[]
for i,name in enumerate(names):
 kinds=json.loads((base/name/'rule.json').read_text())['kinds']
 vals=[int(values[k]) for k in kinds]
 (base/name/'listeners.a').write_text('// Numeric parser kinds from pinned Go cohere AST.Kind.\n// The shared numeric driver must deliver the matching node directly.\nexport const syntaxKinds: readonly number[] = ['+', '.join(map(str,vals))+'];\n')
 entry=base/name/'rule.ts'; text=entry.read_text()
 prefix="import { syntaxKinds as listenerKinds } from './listeners.a';\nexport const syntaxKinds: readonly number[] = listenerKinds;\n"
 text=prefix+text.removeprefix("export { syntaxKinds } from './listeners.a';\n").removeprefix(prefix)
 entry.write_text(text)
 runner.append(f"import {{ syntaxKinds as kinds{i} }} from '{base/name}/listeners.a';")
 runner.append(f"for (const kind of kinds{i}) console.log(String(kind));")
 expected.extend(str(v) for v in vals)
 (work/f'{name}.json').write_text(json.dumps(dict(zip(kinds,vals))))
source=work/'check.a';source.write_text('\n'.join(runner)+'\n')
wanted=('\n'.join(expected)+'\n').encode()
subprocess.run(['go','build','-o',str(work/'adamic'),'./cmd/adamic'],cwd=root,check=True)
def verify(label):
 node=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(source)])
 js=subprocess.check_output([str(work/'adamic'),'js',str(source)])
 (work/'check.mjs').write_bytes(js)
 emitted=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(work/'check.mjs')])
 subprocess.run([str(work/'adamic'),'build',str(source),'-o',str(work/'check'),'--sanitize'],check=True)
 native=subprocess.check_output([str(work/'check')])
 print(label, 'bytes',len(node),'Go matches',node==wanted,emitted==wanted,native==wanted,flush=True)
 return node,emitted,native
assert all(out==wanted for out in verify('baseline'))
for name in names:
 path=base/name/'listeners.a';old=path.read_text();first=old.index('[' ,old.index('= '))+1;end=old.index(',',first) if ',' in old[first:] else old.index(']',first)
 mutant=old[:first]+str(int(old[first:end])+1)+old[end:]
 try:
  path.write_text(mutant)
  assert all(out!=wanted for out in verify('mutant '+name))
 finally:path.write_text(old)
print('PASS: twelve numeric declaration mutants compile and run cleanly; all three outputs disagree with Go.',flush=True)
print('workdir',work,flush=True)
