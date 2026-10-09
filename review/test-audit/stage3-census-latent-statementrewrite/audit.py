from pathlib import Path
import subprocess,time,json,difflib,os
p=Path('/tmp/u162/evidence');f=Path('stage3/census/latent/statementrewrite/main.go');old=f.read_text();results=[]
plan=[dict(id='M1',old='latent statement rewrite: lowering.statement: expected exactly one function, found %d',new='latent statement rewrite: lowering.expression: expected exactly one function, found %d',menu='change constant'),dict(id='M2',old='if len(matches) != 1 {',new='if len(matches) == 1 {',menu='flip condition'),dict(id='M3',old='parser.ParseComments)',new='0)',menu='change option'),dict(id='P1',old='func main() {',new='func main() {\n if true { return }',menu='empty-entry probe')]
for x in plan:
 x['file']=str(f);x['line']=old[:old.index(x['old'])].count('\n')+1;new=old.replace(x['old'],x['new'],1);(p/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
(p/'plan.json').write_text(json.dumps(plan,indent=2));(p/'inventory.json').write_text(json.dumps({'direct_entry':'main','reached_functions':['main'],'other_package_functions':['printed: not reached by the missing-method fixture; reached by valid-method witness'],'oracle':'self-written diagnostic substring and no-output-file assertions'},indent=2))
def run(id,cmd,env=None):
 s=time.monotonic()
 with (p/(id+'.log')).open('w') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 result=dict(id=id,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-s);results.append(result);(p/'runs.json').write_text(json.dumps(results,indent=2));return r.returncode
pkg='./stage3/census/latent/statementrewrite/'
for i in range(1,4):assert run('timing-'+str(i),['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','^TestMissingStatementFailsWithMethodName$'])==0
switch=old.replace('func main() {','func main() {\n if os.Getenv("ADAMIC_MUTANT")=="P1" { return }',1)
switch=switch.replace('if len(matches) != 1 {','if (os.Getenv("ADAMIC_MUTANT")=="M2" && len(matches)==1) || (os.Getenv("ADAMIC_MUTANT")!="M2" && len(matches)!=1) {',1)
switch=switch.replace('"'+plan[0]['old']+'"','auditDiagnostic()',1).replace('parser.ParseComments)','auditParseMode())',1)
switch+='\nfunc auditDiagnostic() string { if os.Getenv("ADAMIC_MUTANT")=="M1" { return "'+plan[0]['new']+'" }; return "'+plan[0]['old']+'" }\nfunc auditParseMode() parser.Mode { if os.Getenv("ADAMIC_MUTANT")=="M3" {return 0};return parser.ParseComments }\n'
try:
 f.write_text(switch);(p/'switch.diff').write_text(''.join(difflib.unified_diff(old.splitlines(True),switch.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))));assert run('switch-vet',['go','vet',pkg])==0;assert run('switch-build',['go','build','-o','/tmp/u162/switched',pkg])==0
 for id in ['M0','M1','M2','M3','P1']:run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'],dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u162/cache/'+id))
 # Demonstrate the survivor on a valid method whose source contains comments.
 source=Path('/tmp/u162/comment-input.go');source.write_text('// package note\npackage lower\n// statement comment\nfunc (l *lowering) statement(node *ast.Node) ([]ir.Statement, error) { return nil, nil }\n')
 for id in ['M0','M3']:assert run('survivor-'+id,['/tmp/u162/switched',str(source),'/tmp/u162/comment-'+id+'.go'],dict(os.environ,ADAMIC_MUTANT=id))==0
 a=Path('/tmp/u162/comment-M0.go').read_text();b=Path('/tmp/u162/comment-M3.go').read_text();assert a!=b;(p/'M3-witness.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='clean-output.go',tofile='M3-output.go')));(p/'survivor-input.go.txt').write_text(source.read_text());(p/'survivor-clean.go.txt').write_text(a);(p/'survivor-mutant.go.txt').write_text(b)
finally:f.write_text(old)
for x in plan:
 subprocess.run(['git','apply','--check',str(p/(x['id']+'.diff'))],check=True);subprocess.run(['git','apply',str(p/(x['id']+'.diff'))],check=True)
 try:assert run(x['id']+'-standalone-vet',['go','vet',pkg])==0
 finally:subprocess.run(['git','apply','-R',str(p/(x['id']+'.diff'))],check=True)
assert subprocess.run(['git','diff','--exit-code']).returncode==0
assert run('restored',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])==0
print(json.dumps(results,indent=2))
