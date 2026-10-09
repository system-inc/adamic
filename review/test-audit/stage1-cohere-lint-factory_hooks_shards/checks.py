from pathlib import Path
import subprocess,difflib,time,json,os
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-factory_hooks_shards';checks=[]
def edit(id,file,old,new,regex,kind):checks.append(dict(id=id,file=file,old=old,new=new,regex=regex,kind=kind))
f='stage1/cohere/lint/factory_hooks_shards_test.go';s=(root/f).read_text();line=next(x for x in s.splitlines(True) if 'return bytes.HasPrefix(output,' in x);edit('W01',f,line,'\treturn true\n','^TestFactoryHooksPlantedFailure$','witness')
f='stage1/cohere/lint/grain_dot_a_rename_test.go';s=(root/f).read_text();old='os.Rename(filepath.Join(dir, "rules/no-var/rule.a"), filepath.Join(dir, "rules/no-var/rule.ts"))';new='os.Rename(filepath.Join(dir, "rules/no-var/rule.ts"), filepath.Join(dir, "rules/no-var/rule.a"))';edit('K01',f,old,new,'^TestProduct_DotARenameSource$','construction')
old='args := append([]string{"build",';new='args := append([]string{"audit-invalid",';edit('K02',f,old,new,'^TestProduct_DotARenameOracle$','construction')
f='stage1/cohere/lint/grain_rules_agree_shards_test.go';edit('K03',f,old,new,'^TestProduct_RulesAgreeOracle$','construction');edit('K04',f,'strings.HasSuffix(name, "_test.go")','strings.HasSuffix(name, "_tesT.go")','^TestRulesAgreeLoweringCacheKey$','construction')
f='stage1/cohere/lint/main.ts';s=(root/f).read_text();old=s[s.index('const args = programArguments();'):];edit('P01',f,old,'','^TestFactoryHooks(Union|_[0-9]{3})$','empty-answer')
f='internal/lower/lower.go';s=(root/f).read_text();a=s.index('\n\tfiles := program.Files()');b=s.index('\n}\n\ntype lowering');t=s[:a]+'\n\treturn nil, nil'+s[b:];t=t.replace('\t"fmt"\n','').replace('\t"path/filepath"\n','');edit('P02',f,s,t,'^TestNestedConstructorGap$','empty-answer')
for id,file,name,nextname,test in [('P03','stage1/cohere/lint/grain_dot_a_rename_test.go','dotARenameSource','dotARenameOracle','TestProduct_DotARenameSource'),('P04','stage1/cohere/lint/grain_dot_a_rename_test.go','dotARenameOracle','dotARenameLowered','TestProduct_DotARenameOracle'),('P05','stage1/cohere/lint/grain_rules_agree_shards_test.go','rulesAgreeOracle','rulesAgreeCapture','TestProduct_RulesAgreeOracle')]:
 s=(root/file).read_text();a=s.index('func '+name+'(');b=s.index('func '+nextname+'(',a);old=s[a:b];signature=old[:old.index('{')+1];edit(id,file,old,signature+'\n\treturn ""\n}\n\n','^'+test+'$','empty-answer construction')
f='stage1/cohere/lint/grain_rules_agree_shards_test.go';s=(root/f).read_text();a=s.index('func rulesAgreeSourceInputsAt(');b=s.index('// Protect the warm-fetch contract',a);old=s[a:b];edit('P06',f,old,old[:old.index('{')+1]+'\n\treturn nil\n}\n\n','^TestRulesAgreeLoweringCacheKey$','empty-answer construction')
f='stage1/cohere/lint/lint_test.go';s=(root/f).read_text();a=s.index('func difference(');b=s.index('func manifest(',a);old=s[a:b];edit('W02',f,old,'func difference(got, want []byte) string { return "" }\n\n','^TestRulesAgreePlantedFailure$','witness')
metrics=[]
for c in checks:
 p=root/c['file'];s=p.read_text();assert s.count(c['old'])==1,(c['id'],s.count(c['old']));t=s.replace(c['old'],c['new']);p.write_text(t)
 (out/'diffs'/f"{c['id']}-{c['kind'].replace(' ','-')}.diff").write_text(''.join(difflib.unified_diff(s.splitlines(True),t.splitlines(True),fromfile='a/'+c['file'],tofile='b/'+c['file'])))
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u108/cache/'+c['id']
 try:
  if c['file'].endswith('.go'):
   with (out/'logs'/f"{c['id']}-vet.log").open('w') as log:subprocess.run(['go','vet','./internal/lower/' if c['id']=='P02' else './stage1/cohere/lint/'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
  start=time.monotonic()
  with (out/'logs'/f"{c['id']}.log").open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',c['regex']],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  metrics.append({**{k:c[k] for k in ['id','kind','file','regex']},'line':s[:s.index(c['old'])].count('\n')+1,'wall':time.monotonic()-start,'exit':r.returncode});(out/'check-times.json').write_text(json.dumps(metrics,indent=2))
 finally:p.write_text(s)
