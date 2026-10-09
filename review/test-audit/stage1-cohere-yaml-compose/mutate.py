import pathlib,json,subprocess,time,os,re,difflib
p=pathlib.Path('/tmp/u150/evidence');root=pathlib.Path('stage1/cohere/yaml');groups=json.loads((p/'groups.json').read_text());filenames=['cstParser.ts','directives.ts','format.ts','layout.ts','compose_main.ts','cst_main.ts','main.ts','compose_test.go','cst_test.go','file_driver_shards_test.go'];original={f:(root/f).read_text() for f in filenames};switched=dict(original);plan=[]
def diff(id,f,new):
 old=original[f];(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+str(root/f),tofile='b/'+str(root/f))))
def change(id,f,old,new,switch,indices,kind='production'):
 text=original[f];assert old in text,(f,old);line=text[:text.index(old)].count('\n')+1;diff(id,f,text.replace(old,new,1));switched[f]=switched[f].replace(old,switch,1);plan.append(dict(id=id,kind=kind,file=str(root/f),line=line,old=old,new=new,indices=indices))
change('M1','cstParser.ts','this.lineStarts.push(0);','this.lineStarts.push(1);',"this.lineStarts.push(auditChoice === 'M1' ? 1 : 0);",[1,2,0,6])
change('M2','directives.ts',"version = '1.2';","version = '1.1';","version = auditChoice === 'M2' ? '1.1' : '1.2';",[1,0,6])
change('M3','format.ts',"result.text = (bom ? '\\ufeff' : '') + printer.format(root);","result.text = (bom ? '' : '') + printer.format(root);","result.text = (bom ? (auditChoice === 'M3' ? '' : '\\ufeff') : '') + printer.format(root);",[0,6])
change('M4','layout.ts','80 - this.column','40 - this.column',"(auditChoice === 'M4' ? 40 : 80) - this.column",[0,6])
change('S1','file_driver_shards_test.go','filepath.Join(directory, "format.c")','filepath.Join(directory, "missing.c")','filepath.Join(directory, yamlAuditProductName("format.c"))',[8],'construction')
for id,f,start,indices in [('P1','compose_main.ts','const path = programArguments()', [1]),('P2','cst_main.ts','const path = programArguments()',[2]),('P3','main.ts','const args = programArguments()',[0,6])]:
 text=original[f];pos=text.index(start);diff(id,f,text[:pos]);pos=switched[f].index(start);switched[f]=switched[f][:pos]+"if(auditChoice !== '"+id+"') {\n"+switched[f][pos:]+'\n}\n';plan.append(dict(id=id,kind='probe',file=str(root/f),line=text[:text.index(start)].count('\n')+1,change='Empty top-level port entry: skip its body and produce no stdout',indices=indices))
header='func fileDriverSetup(t *testing.T) *fileDriverState {'
change('P4','file_driver_shards_test.go',header,header+' if true { return nil };',header+' if os.Getenv("ADAMIC_MUTANT") == "P4" { return nil };',[8],'probe')
for id,f,indices in [('W1','compose_test.go',[3]),('W2','cst_test.go',[4])]:
 # Change only the comparison inside the witness function, preserving its assertions and argument references.
 text=original[f];fn='func TestComposeMutants' if id=='W1' else 'func TestCSTMutants';a=text.index(fn);old='bytes.Equal(side.out, expected)';pos=text.index(old,a);pure=text[:pos]+'(true || '+old+')'+text[pos+len(old):];diff(id,f,pure);switched[f]=switched[f][:pos]+'(os.Getenv("ADAMIC_MUTANT") == "'+id+'" || '+old+')'+switched[f][pos+len(old):];plan.append(dict(id=id,kind='witness',file=str(root/f),line=text[:pos].count('\n')+1,change='Weaken equality to always agree within the witness; assertions unchanged',indices=indices))
# Earlier replacements change character offsets in the file-driver file, so locate by function again.
f='file_driver_shards_test.go';text=original[f];old='bytes.Equal(output, wanted)';pos=text.index(old,text.index('func TestFileDriverUnion'));diff('W3',f,text[:pos]+'(true || '+old+')'+text[pos+len(old):]);pos=switched[f].index(old,switched[f].index('func TestFileDriverUnion'));switched[f]=switched[f][:pos]+'(os.Getenv("ADAMIC_MUTANT") == "W3" || '+old+')'+switched[f][pos+len(old):];plan.append(dict(id='W3',kind='witness',file=str(root/f),line=text[:text.index(old)].count('\n')+1,change='Always agree on planted output; native precondition comparison unchanged',indices=[5]))
for f in filenames:
 if not f.endswith('.ts'):continue
 text=switched[f]
 if "from 'adamic'" not in text:text="import { readTextFile } from 'adamic';\n"+text
 elif 'readTextFile' not in text.split("from 'adamic'")[0]:text=text.replace("import { panic } from 'adamic';","import { panic, readTextFile } from 'adamic';",1)
 lines=text.splitlines(True);last=max(i for i,line in enumerate(lines) if line.startswith('import '));helper="const auditRead = readTextFile('/tmp/u150/selector');\nconst auditChoice = auditRead.kind === 'Error' ? '' : auditRead.text.trim();\n";lines.insert(last+1,helper);switched[f]=''.join(lines)
switched['file_driver_shards_test.go']+='\nfunc yamlAuditProductName(name string) string { if os.Getenv("ADAMIC_MUTANT") == "S1" { return "missing.c" }; return name }\n'
(p/'plan.json').write_text(json.dumps(plan,indent=2));results=[]
def run(id,index,kind='matrix'):
 g,m=groups[index];path=p/f'{id}-{index}-{kind}.log';env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_YAML_LIBRARY='/tmp/u150/library',ADAMIC_BUILD_CACHE_DIR='/tmp/u150/cache/'+('switched' if id not in ['S1','P4'] else id),ADAMIC_YAML_ARTIFACTS='/tmp/u150/artifacts/'+id+'-'+str(index));pathlib.Path('/tmp/u150/selector').write_text(id);cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+'|'.join(m)+')$'];s=time.monotonic()
 with path.open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 result=dict(id=id,kind=kind,index=index,row=g,exit=r.returncode,seconds=time.monotonic()-s,command=cmd,log=path.name);results.append(result);(p/'matrix-timings.json').write_text(json.dumps(results,indent=2));print(result,flush=True)
 return r.returncode
try:
 for f,text in switched.items():(root/f).write_text(text)
 (p/'switch.diff').write_text(subprocess.check_output(['git','diff'],text=True))
 with (p/'switch-vet.log').open('w') as log:r=subprocess.run(['go','vet','./stage1/cohere/yaml/'],stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0,'switch vet failed'
 for index in [2,1,0,6]:assert run('M0',index,'clean-switch')==0,'instrumentation baseline failed'
 for item in plan:
  for index in item['indices']:run(item['id'],index)
finally:
 for f,text in original.items():(root/f).write_text(text)
