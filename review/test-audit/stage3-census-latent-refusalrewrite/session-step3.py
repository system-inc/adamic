import pathlib,json,subprocess,difflib,time,os
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-audit/stage3-census-latent-refusalrewrite';s=root/'stage3/census/latent/refusalrewrite/rewrite.go';base=s.read_text();plan=json.loads((p/'plan.json').read_text());runs=json.loads((p/'runs.json').read_text());pkg='./stage3/census/latent/refusalrewrite/'
def run(n,c,env=None):
 t=time.monotonic()
 with (p/(n+'.log')).open('w') as f:r=subprocess.run(c,stdout=f,stderr=subprocess.STDOUT,env=env)
 runs[n]={'exit':r.returncode,'wall_seconds':time.monotonic()-t};return r.returncode
for m in plan:
 if m['id']=='M2':
  change=base.replace('    defer func(){ latentFindingOwner = outerOwner }()','').replace('    outerOwner := latentFindingOwner','')
  m['new']='';m['compile_adjustment']='Also drop the now unused outerOwner declaration at line 159.'
 elif m['id']=='M3':
  change=base.replace('latentRecord(%s); %s = nil; node.ForEachChild(%s); latentStop = false','latentRecord(%s); node.ForEachChild(%s); latentStop = false').replace('binding, binding, name))','binding, name))')
  m['new']='latentRecord(%s); node.ForEachChild(%s); latentStop = false';m['compile_adjustment']='Remove the unused corresponding fmt.Sprintf argument at line 167.'
 elif m['id']=='M9':
  change=base.replace('\t\t\t\trewriteBlocks(child, rewrite)','').replace('if child, ok := node.(*ast.BlockStmt); ok {','if _, ok := node.(*ast.BlockStmt); ok {')
  m['new']='';m['compile_adjustment']='Discard the now unused BlockStmt binding at line 255.'
 else:continue
 (p/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),change.splitlines(True),fromfile='a/'+str(s.relative_to(root)),tofile='b/'+str(s.relative_to(root)))))
 try:
  assert run(m['id']+'-apply',['git','apply','--check',str(p/(m['id']+'.diff'))])==0
  s.write_text(change);assert run(m['id']+'-vet',['go','vet',pkg])==0
  run(m['id']+'-replay',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])
 finally:s.write_text(base)
(p/'plan.json').write_text(json.dumps(plan,indent=2))
try:
 subprocess.run(['git','apply',str(p/'switch.diff')],check=True)
 sw=s.read_text().replace('prefixText = strings.ReplaceAll(prefixText, "    defer func(){ latentFindingOwner = outerOwner }()", "    _ = outerOwner")','prefixText = strings.ReplaceAll(prefixText, "    defer func(){ latentFindingOwner = outerOwner }()", "")\n prefixText = strings.ReplaceAll(prefixText, "    outerOwner := latentFindingOwner", "")').replace('"; _ = "+binding+";"','";"')
 sw=sw.replace('if os.Getenv("ADAMIC_MUTANT") == "M9" {\n\t\t\t\t\t_ = child\n\t\t\t\t} else {','if os.Getenv("ADAMIC_MUTANT") != "M9" {')
 s.write_text(sw);assert run('switch-gofmt',['gofmt','-w',str(s)])==0
 (p/'switch.diff').write_text(subprocess.check_output(['git','diff','--',str(s.relative_to(root))],text=True))
 assert run('switch-build',['go','test','-c','-o','/tmp/u160-switch.test',pkg])==0
 assert run('switch-control',['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'])==0
 for m in plan:
  env=os.environ.copy();env['ADAMIC_MUTANT']=m['id'];run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','.'],env)
finally:s.write_text(base);(p/'runs.json').write_text(json.dumps(runs,indent=2))
print('pure drop-statement diffs and switch matrix verified')
