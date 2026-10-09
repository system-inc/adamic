import pathlib,subprocess,time,json,difflib,os,re
root=pathlib.Path('/workspace/adamic');os.chdir(root);out=root/'review/test-audit/stage3-census-latent-refusalrewrite';src=root/'stage3/census/latent/refusalrewrite/rewrite.go';original=src.read_text();pkg='./stage3/census/latent/refusalrewrite/'
plan=json.loads((out/'plan.json').read_text());runs=json.loads((out/'runs.json').read_text())
def run(name,cmd):
 t=time.monotonic()
 with (out/(name+'.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 runs[name]={'exit':r.returncode,'wall_seconds':time.monotonic()-t};return r.returncode
for m in plan:
 changed=original.replace(m['old'],m['new'])
 if m['id']=='M8':
  changed=changed.replace('if ret, ok := statement.(*ast.ReturnStmt); ok {','if _, ok := statement.(*ast.ReturnStmt); ok {')
  m['compile_adjustment']='Discard the now unused ReturnStmt binding on line 240.'
 if m['id']=='P1':
  a=original.index('func Rewrite(source []byte) ([]byte, error) {');b=original.index('\nfunc eraseWalks',a)
  changed=original[:a]+'func Rewrite(source []byte) ([]byte, error) { return nil, nil }\n'+original[b:]
  m['compile_adjustment']='Replace whole Rewrite body with empty return to avoid unreachable code.'
 (out/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(src.relative_to(root)),tofile='b/'+str(src.relative_to(root)))))
 try:
  assert run(m['id']+'-apply',['git','apply','--check',str(out/(m['id']+'.diff'))])==0
  src.write_text(changed)
  assert run(m['id']+'-vet',['go','vet',pkg])==0
  run(m['id']+'-replay',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])
 finally:src.write_text(original)
(out/'plan.json').write_text(json.dumps(plan,indent=2))
# Survivor witnesses call public Rewrite through the repository's command, without test edits.
fixture=(root/'stage3/census/latent/refusalrewrite/testdata/refusals-41231d51.go.txt')
test=(root/'stage3/census/latent/refusalrewrite/rewrite_test.go').read_text();synthetic=re.search(r'source := `(.+?)`',test,re.S).group(1)
synthetic=synthetic.replace(' for _,pragma:=range module.Pragmas {if pragma.Name!="" {return &Refused{What:pragma.Name}}}\n','').replace(' if contractError!=nil{return contractError}\n','')
(out/'two-outer-returns.go.txt').write_text(synthetic)
for ident,inputfile in [('M10',fixture),('M11',out/'two-outer-returns.go.txt')]:
 for mode in ['before','after']:
  try:
   if mode=='after':
    m=next(x for x in plan if x['id']==ident);src.write_text(original.replace(m['old'],m['new']))
   run(ident+'-witness-'+mode,['go','run','./stage3/census/latent/refusalrewrite/cmd','-input',str(inputfile),'-output',str(out/(ident+'-witness-'+mode+'.go.txt'))])
  finally:src.write_text(original)
run('clean-build',['go','test','-c','-o','/tmp/u160-clean.test',pkg])
run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])
(out/'runs.json').write_text(json.dumps(runs,indent=2))
print('validation and survivor witnesses complete')
