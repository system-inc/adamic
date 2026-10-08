import json,os,subprocess,tempfile,gzip,re
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[5];COHERE=ROOT/'cohere'
functions={
'enclosingFunctionOf':'node','isFunctionLike':'node','isReactCompiledFunction':'node','isTopLevelCompilationCandidate':'node','reactFunctionNameOf':'node','hasValidComponentParameters':'node','parametersOf':'node','isRestParameter':'node','returnsNonNode':'node','isNonNodeExpression':'node',
'enclosingClassOf':'node','enclosingComponentOf':'node','functionBodyBlock':'node','isComponentClass':'node','isCreateReactClassCall':'node','isPureComponentBase':'node','isReactComponentBase':'node','isThisExpression':'node','semanticParentOf':'node','sortDefaultPropsInitializerOf':'node','stylePropObjectUnwrapParentheses':'node','skipParenthesesOptional':'node','attributesOf':'node','sourceSliceOf':'ctx','reactPragmaFor':'ctx'}
with tempfile.TemporaryDirectory() as tmp:
 tmp=Path(tmp);replace={};wrappers='package react\nimport ("github.com/microsoft/TypeScript/tsc/shim/ast";"github.com/system-inc/cohere/internal/lint/rule")\n'
 for path in (COHERE/'internal/lint/rules/react').glob('*.go'):
  if path.name.endswith('_test.go'):continue
  source=path.read_text();changed=False
  for name,mode in functions.items():
   match=re.search(r'func '+name+r'\((.*?)\) ([^{]+)\{',source,re.S)
   if not match:continue
   result=match.group(2).strip();source=source[:match.start()]+source[match.start():].replace('func '+name+'(', 'func adamicOrig_'+name+'(',1);changed=True
   if mode=='ctx':
    args='ctx rule.Context'+(', node *ast.Node' if name=='sourceSliceOf' else '');call='ctx'+(', node' if name=='sourceSliceOf' else '');input='node' if name=='sourceSliceOf' else 'ctx.SourceFile.AsNode()'
   else:args='node *ast.Node';call='node';input='node'
   if result=='(string, bool)':body='a,b:=adamicOrig_'+name+'('+call+');recordAdamicAst("'+name+'",'+input+',adamicPair{a,b});return a,b'
   else:body='value:=adamicOrig_'+name+'('+call+');recordAdamicAst("'+name+'",'+input+',value);return value'
   wrappers+='func '+name+'('+args+') '+result+'{'+body+'}\n'
  if changed:
   side=tmp/path.name;side.write_text(source);replace[str(path)]=str(side)
 side=tmp/'wrappers.go';side.write_text(wrappers);replace[str(COHERE/'internal/lint/rules/react/adamic_ast_wrappers.go')]=str(side)
 recorder=HERE/'ast_recorder.go';replace[str(COHERE/'internal/lint/rules/react/adamic_ast_recorder.go')]=str(recorder)
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 records=tmp/'records.jsonl'
 with (HERE/'capture-ast.log').open('w') as log:
  subprocess.run(['go','test','-tags=lintoracle','-overlay='+str(overlay),'./internal/lint/rules/react','-count=1','-timeout=10m'],cwd=COHERE,env=os.environ|{'ADAMIC_AST_CAPTURE':str(records)},stdout=log,stderr=subprocess.STDOUT,check=True)
 arenas=[];calls=[]
 for line in records.read_text().splitlines():
  row=json.loads(line)
  if 'Nodes' in row:arenas.append(row)
  else:calls.append(row)
 assert {r['Name'] for r in calls}==set(functions),set(functions)-{r['Name'] for r in calls}
 (HERE/'ast.json.gz').write_bytes(gzip.compress(json.dumps({'Arenas':arenas,'Calls':calls},ensure_ascii=True).encode(),mtime=0))
 print('captured',len(arenas),'arenas',len(calls),'calls', {n:sum(r['Name']==n for r in calls) for n in functions})
