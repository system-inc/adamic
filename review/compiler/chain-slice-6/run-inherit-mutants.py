import json,os,subprocess
from pathlib import Path
root=Path.cwd(); out=root/'review/compiler/chain-slice-6/mutants'
mutants=[
 ('M02','internal/lower/class.go','metadata.Fields = append(metadata.Fields, l.result.Classes[base.class-1].Fields...)','// mutant: omitted inherited prefix','^TestInheritanceDerivedFieldOrder$','B.Fields'),
 ('M07','internal/lower/class_inheritance.go','return 0, true','return ir.Boolean, true','^TestInheritanceVoidBooleanOverrideReason$','want NotYet'),
 ('P_LOWER','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { if true { return nil, nil }','^TestInheritance(AllowsSoundOverrides|KeepsNominalTupleDestructuring|NativeSignatureNeighbors)$','accepted inheritance must produce')]
for name,path,old,new,selection,catcher in mutants:
 source=(root/path).read_text(); assert source.count(old)==1,(name,source.count(old))
 directory=out/name;directory.mkdir(exist_ok=True)
 changed=directory/((root/path).name+'.txt');changed.write_text(source.replace(old,new,1))
 overlay=directory/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/path):str(changed)}})+'\n')
 with (directory/'test.log').open('w') as log:
  result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run',selection,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 text=(directory/'test.log').read_text()
 assert result.returncode!=0 and catcher in text and '[build failed]' not in text,(name,text)
 print(name,'caught by',catcher,flush=True)
