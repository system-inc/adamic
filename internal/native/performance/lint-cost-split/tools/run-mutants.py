from pathlib import Path
import re,shutil,subprocess,json
src=Path('/workspace/lint-cost-prototype');s=(src/'main.c').read_text();mutations={}
v,n=re.subn(r'^\s*adamic_function_\d+_register\(&adamic_string_51, \(0x1p\+02\)\);$','',s,flags=re.M);assert n==1;mutations['missing-listener']=v
spans=[];total=0
for m in re.finditer(r'static [^\n]*adamic_function_\d+_ParseNode_new(?:_in)?\([^\n]+\) \{',s):
 end=s.index('\n}\n',m.end());chunk=s[m.start():end]
 changed,n=re.subn(r'(double (adamic_temporary_\d+) = adamic_function_\d+_numericKind\(adamic_local_\d+_kind\);)',r'\1\n\2 = 0.0;',chunk);total+=n;spans.append((m.start(),end,changed))
assert total==2,total
v=s
for start,end,chunk in reversed(spans):v=v[:start]+chunk+v[end:]
mutations['wrong-kind']=v
root=Path('/workspace/lint-cost-controls');records={}
for name,source in mutations.items():
 dst=Path('/workspace/lint-cost-mutant-'+name);dst.mkdir(exist_ok=True)
 for p in src.iterdir():
  if p.suffix in ['.c','.h']:shutil.copy2(p,dst/p.name)
 (dst/'main.c').write_text(source)
 with (dst/'harness-build.log').open('wb') as out:subprocess.run(['python3','/workspace/lint-cost-build.py','mutant-'+name],stdout=out,stderr=out,check=True)
 case='template_string' if name=='missing-listener' else 'private_class';manifest=root/(case+'.txt');outputs={}
 for label,cmd in [('Go',['/workspace/lint-cost-baseline/oracle']),('mutant',[str(dst/'scanner')])]:
  out=root/(name+'-'+label+'.stdout');err=root/(name+'-'+label+'.stderr')
  with out.open('wb') as o,err.open('wb') as e:r=subprocess.run(cmd+['--manifest',str(manifest),'--count'],stdout=o,stderr=e)
  assert r.returncode==0 and not err.read_bytes();outputs[label]=out.read_text()
 assert outputs['Go']=='1\n' and outputs['mutant']=='0\n',(name,outputs)
 records[name]={'Go':outputs['Go'],'mutant':outputs['mutant'],'exit':0,'caught':'semantic finding count disagrees with Go'};print(name,records[name],flush=True)
Path('/workspace/lint-cost-mutants.json').write_text(json.dumps(records,indent=2)+'\n')
