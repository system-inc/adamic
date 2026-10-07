"""Actual parsed expressions against Go reference.ConstantString."""
import pathlib,json,subprocess,time,itertools,hashlib
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-constants');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  r=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 print(name,r.returncode,round(time.monotonic()-start,6),flush=True)
 assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000])
 return (work/(name+'.stdout')).read_bytes()
atoms=['""','"[Á]"',r'"\u0301"','`u`','1','0x10','1e3','true','false','null','name','/[👍]/','-1','undefined','call()','String.raw`u`']
rows=atoms+[f'({a})+({b})' for a,b in itertools.product(atoms,repeat=2)]
rows += ['`head${'+a+'}tail`' for a in atoms]
rows += ['`x${"a"}${true}${null}${0x10}z`','("a"+"b")','1+2','"a"-1','"u" as string','("u" as string)!','`u${""}`','`[${"👍"}]`']
fixture=write('rows.json',json.dumps(rows));virtual=repo/'cohere/wave09_constant_oracle.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/constant_oracle.go')}}))
go=work/'go';run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere');truth=run('go',[go,fixture])
source='import {Parser} from '+json.dumps(str(repo/'stage1/typescript/parser/parser.ts'))+';\nimport {written} from '+json.dumps(str(repo/'stage1/typescript/parser/nodes.ts'))+';\nimport {constantString} from '+json.dumps(str(root/'constant_value.a'))+';\n'
source+='const rows:readonly string[]='+json.dumps(rows)+';for(const expression of rows){const parser=new Parser("const value="+expression+";");parser.file();for(const node of parser.nodes){if(node.kind===\'VariableDeclaration\'){const value=constantString(parser.nodes,node.children[node.children.length-1]??-1);console.log(`${value.known?\'true\':\'false\'}\\t${written(value.text)}`);break;}}}\n'
path=write('main.a',source)
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=(root/'constant_value.a').read_text().replace("'../../../../typescript/parser/nodes.ts'",json.dumps(str(repo/'stage1/typescript/parser/nodes.ts')))
  anchor='(!left.string&&!right.string)';assert text.count(anchor)==1
  mutant=write('mutated.a',text.replace(anchor,'(!left.string||!right.string)'));current=write('mutant-main.a',source.replace(str(root/'constant_value.a'),str(mutant)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':assert actual!=truth;diff=next(i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b)
 elif actual!=truth:
  diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)));raise AssertionError((mode,diff,actual[max(0,diff-80):diff+100],truth[max(0,diff-80):diff+100]))
emitted=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('main.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:assert run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p])==truth
write('results.json',json.dumps({'inputs':len(rows),'bytes':len(truth),'sha256':hashlib.sha256(truth).hexdigest(),'mutant_difference':diff},indent=2)+'\n')
print('PASS actual source constant-expression folding, Go bytes, native, sanitizers, both Node modes and coercion mutant; binding resolution is not exercised',flush=True)
