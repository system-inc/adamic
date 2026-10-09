import pathlib,json,subprocess,os,time
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-defend/stage1-cohere-tsprinter-class_interface_gap'
plan=[('P1','internal/lower/expression.go',755,'append(body, ir.Return{Value: value})','append(body, ir.Return{Value: ir.NumberConstant{Value: 0}})'),('N1','internal/lower/library_math_number.go',43,'ast.SkipParentheses(node).Kind == ast.KindNullKeyword','ast.SkipParentheses(node).Kind == ast.KindStringLiteral'),('E1','stage1/cohere/tsprinter/expressions.ts',1802,"printer.node(printer.unwrapped(root)).kind === 'StringLiteral'","printer.node(printer.unwrapped(root)).kind === 'NoSubstitutionTemplateLiteral'")]
(out/'aimed-plan.json').write_text(json.dumps(plan,indent=2))
for ident,file,line,old,new in plan:
 p=root/file; original=p.read_text(); assert original.count(old)==1
 p.write_text(original.replace(old,new))
 with (out/(ident+'.diff')).open('w') as f: subprocess.run(['git','diff','--',file],cwd=root,stdout=f,check=True)
 env=os.environ.copy(); env.update(ADAMIC_BUILD_CACHE_DIR='/tmp/tsprinter-defense/cache/'+ident,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u156-typescript',ADAMIC_TS_PRETTIER='/tmp/tsprinter-defense/prettier/node_modules/prettier')
 groups=[('small','^(TestClosedPrefixUpdateValueGap|TestNumberConstructor|TestDocumentsAgainstGoAndPrettier|TestCompilerGaps)$'),('expressions','^TestExpressionsAgainstGoAndPrettier_000$'),('statements','^TestStatementsAgainstGoAndPrettier_000$'),('tsc','^TestTSCCorpusAgreement_[0-9]+$')]
 if ident=='E1': groups.append(('build','^TestProduct_TSPrinterExpressionsSanitized$'))
 try:
  if ident!='E1':
   with (out/(ident+'-vet.log')).open('w') as f: subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=90,check=True)
  for group,pattern in groups:
   command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',pattern]
   start=time.time()
   with (out/(ident+'-'+group+'.log')).open('w') as f: result=subprocess.run(command,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   with (out/'commands.jsonl').open('a') as f:f.write(json.dumps(dict(mutant=ident,group=group,command=command,seconds=time.time()-start,status=result.returncode,cache=env['ADAMIC_BUILD_CACHE_DIR']))+'\n')
   print(ident,group,result.returncode,round(time.time()-start,2),flush=True)
 finally:p.write_text(original)
