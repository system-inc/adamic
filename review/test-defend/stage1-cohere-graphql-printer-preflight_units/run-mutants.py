import pathlib,json,subprocess,time
root=pathlib.Path('/workspace/adamic'); out=pathlib.Path('/tmp/def-graphql/evidence')
plans=[dict(mutant='D01',file='stage1/cohere/graphql/parser.ts',before='`Unexpected ${getTokenDesc(token)}.`',after='`Expected ${getTokenDesc(token)}.`',change='Change unexpected-token diagnostic constant; target EOF refusal absent from valid file-driver input'),dict(mutant='D02',file='stage1/cohere/graphql/lexer.ts',before='let line = 1;',after='let line = 0;',change='Off-by-one error-location starting line; target empty and whitespace-only refusal coordinates'),dict(mutant='D03',file='stage1/cohere/graphql/lexer.ts',before="for (let index = 0; index < body.length; index++) {\n\t\tconst code = body.charCodeAt(index);",after="for (let repeat = 0; repeat < 8; repeat++) {\n\tlastLineStart = 0;\n\tline = 1;\n\tfor (let index = 0; index < body.length; index++) {\n\t\tconst code = body.charCodeAt(index);",change='Repeat error-location scanning eight times with counters reset; answer-preserving work-growth attempt for the 30-second unit assertion',extra_before="\treturn { line, column: position + 1 - lastLineStart };",extra_after="\t}\n\treturn { line, column: position + 1 - lastLineStart };")]
for p in plans:
 s=(root/p['file']).read_text();assert s.count(p['before'])==1
 p['line']=s[:s.index(p['before'])].count('\n')+1
(out/'plan.json').write_text(json.dumps(plans,indent=2))
selector='^(TestPrinter(Whitespace.*|AsGoCohere.*|FileDriver|Shard.*|ConstructorGap))$'
runs=[]
for p in plans:
 path=root/p['file'];original=path.read_text()
 try:
  changed=original.replace(p['before'],p['after'])
  if 'extra_before' in p:
   assert changed.count(p['extra_before'])==1
   changed=changed.replace(p['extra_before'],p['extra_after'])
  path.write_text(changed)
  (out/(p['mutant']+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',p['file']],cwd=root))
  env=f"ADAMIC_GRAPHQL_PRETTIER=/tmp/def-graphql/prettier ADAMIC_BUILD_CACHE_DIR=/tmp/def-graphql/cache/{p['mutant']}"
  for phase,regex in [('build','^TestProduct_GraphQLPrinterSanitized$'),('matrix',selector)]:
   cmd=f"source /workspace/adamic-tools/env.sh\n{env} timeout 120 go test -json -count=1 -timeout 90s -parallel 2 ./stage1/cohere/graphql/printer/ -run '{regex}'"
   (out/(p['mutant']+'-'+phase+'-command.txt')).write_text(cmd)
   start=time.monotonic()
   with (out/(p['mutant']+'-'+phase+'.log')).open('w') as log:code=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
   runs.append(dict(mutant=p['mutant'],phase=phase,exit=code,wall_seconds=time.monotonic()-start))
   (out/'runs.json').write_text(json.dumps(runs,indent=2))
   if phase=='build' and code!=0:raise RuntimeError('native build did not pass')
 finally:path.write_text(original)
