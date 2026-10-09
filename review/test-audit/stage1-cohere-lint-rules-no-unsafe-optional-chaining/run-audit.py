import pathlib, subprocess, time, json, difflib, os
root=pathlib.Path('/workspace/adamic');p=pathlib.Path('/tmp/u123/evidence');rule=pathlib.Path('stage1/cohere/lint/rules/no-unsafe-optional-chaining/rule.a');original=rule.read_text();results=[]
plans=[('M1','if(!node.optional) { return; }','if(node.optional) { return; }',22),('M2',"read('disallowarithmeticoperators', 'false')","read('disallowarithmeticoperators', 'true')",32),('M3',"        this.context.report(index, 'no-unsafe-optional-chaining', arithmetic ? 'unsafeArithmetic' : 'unsafeOptionalChain', arithmetic ? arithmeticMessage : unsafe, '', '', '');",'',28)]
(p/'plan.json').write_text(json.dumps({'code_under_test':'Owned rule.a and profile.a compiled by lower.Lower, native.C, native.Build and javascript.JavaScript; products are not executed by the row','oracle':'self: successful compilation and successful emitted file write','port_functions':['Rule.constructor','Rule.pattern','Rule.reaching','Rule.visit','create','profile.ancestry','profile.visit','profile.run','profile top-level entry'],'mutants':[dict(id=i,file=str(rule),line=line,old=a,new=b) for i,a,b,line in plans]},indent=2))
w=pathlib.Path('/tmp/u123/input.ts');w.write_text('(obj?.foo).bar;\nobj?.foo + 1;\n');manifest=pathlib.Path('/tmp/u123/manifest.txt');manifest.write_text(str(w)+'\n')
def run(id,kind,cmd,env=None):
 start=time.monotonic()
 with (p/f'{id}-{kind}.log').open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 row=dict(id=id,kind=kind,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-start);results.append(row);(p/'mutation-timings.json').write_text(json.dumps(results,indent=2));print(row,flush=True);return r.returncode
run('clean','witness',['/tmp/u123/clean',str(manifest)])
try:
 for id,old,new,line in plans:
  changed=original.replace(old,new);assert changed!=original
  (p/f'{id}.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(rule),tofile='b/'+str(rule))))
  rule.write_text(changed)
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR=f'/tmp/u123/cache/{id}')
  run(id,'matrix',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/no-unsafe-optional-chaining/','-run','.'],env)
  if run(id,'build',['timeout','90','/tmp/u123/adamic','build','stage1/cohere/lint/rules/no-unsafe-optional-chaining/profile.a','-o',f'/tmp/u123/{id}'],env)==0:
   run(id,'witness',[f'/tmp/u123/{id}',str(manifest)],env)
  rule.write_text(original)
finally:rule.write_text(original)
