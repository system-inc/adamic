import pathlib,json,difflib,re,subprocess,time,os,sys
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-lint-harness'; pkg='./stage1/cohere/lint/'
def read(p):return (root/p).read_text()
def body_drop(s,start,end):
 a=s.index(start); b=s.index(end,a); return s[:a]+s[b:]
jsx='stage1/cohere/lint/jsx_shards_test.go'; em='stage1/cohere/lint/emitted_javascript_shards_test.go'; cs='stage1/cohere/lint/complete_suggestion_serialization_shards_test.go'; sa='stage1/cohere/lint/suggestion_alongside_shards_test.go'
tree='^(TestJsxLintTrees_[0-9]{3}|TestJsxLintTreesUnion|TestJsxLintTrees_Setup|TestProduct_jsx_.*)$'
specs=[]
def add(id,file,old,new,kind,regex,description):
 s=read(file); assert s.count(old)==1,(id,s.count(old)); specs.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,kind=kind,regex=regex,change=description))
add('M1','stage1/typescript/parser/nodes.ts','code <= 126','code <= 125','production',tree,'off-by-one printable ASCII upper bound 126 -> 125')
add('M2','stage1/typescript/parser/nodes.ts','depth + 1, whole','depth + 2, whole','production',tree,'change recursive tree depth increment 1 -> 2')
add('M3','stage1/typescript/parser/jsx.ts',"semantic = whitespace ? '1' : '0'","semantic = whitespace ? '0' : '1'",'production',tree,'flip JSX text whitespace semantics')
add('H1',em,'const testEmittedJavaScriptMismatchShards = 1','const testEmittedJavaScriptMismatchShards = 2','setup','^TestEmittedJavaScriptMismatch$','change declared shard count 1 -> 2')
add('H2',sa,'const testSuggestionAlongsideAutomaticFixShards = 3','const testSuggestionAlongsideAutomaticFixShards = 4','setup','^TestSuggestionAlongsideAutomaticFix$','change declared shard count 3 -> 4')
add('W1',cs,'return caseID != plantedID || bytes.Equal(got, want)','return true','witness','^TestCompleteSuggestionSerialization$','return early with unconditional agreement')
add('W2',jsx,'\tif diff := difference(got, want); diff != "" {\n\t\tt.Fatalf("%s: %s", side, diff)\n\t}\n','', 'witness','^TestJsxLintTreesShardDisagreement$','drop the complete tree disagreement assertion')
add('H3',jsx,'if len(paths) == 0 {','if len(paths) < 0 {','setup','^TestJsxLintTreesShardCoverage$','flip empty-corpus rejection condition to impossible bound')
s=read(jsx); a=s.index('\tjsxPrepareOnce.Do(func() {',s.index('func jsxPrepareTrees')); b=s.index('\n\t// A failed builder',a)
add('H4',jsx,s[a:b],'','setup','^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesUnion|TestJsxLintTreesSetupIsolation)$','drop the complete shared preparation statement')
s=read(em); a=s.index('\temittedMismatchProducts.Do(func() {'); b=s.index('\n\tif emittedMismatchProducts.oracle ==',a)
add('H5',em,s[a:b],'','setup','^TestEmittedJavaScriptMismatch_Setup$','drop the complete mismatch product preparation statement')
add('H6',jsx,'\t\tif err := command.Run(); err != nil {\n\t\t\treturn fmt.Errorf("build %s: %w\\n%s\\n%s", name, err, &stdout, &stderr)\n\t\t}\n','','setup','^TestProduct_jsx_(membership|parser)$','drop the oracle construction command, preserving oracle source')
add('H7',jsx,'command.Dir = root','command.Dir = dir','setup','^TestProduct_jsx_(membership|parser)$','change the build working-directory option from cohere root to scratch output')
add('H8',jsx,'return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(lowered)), 0644)','return os.WriteFile(filepath.Join(dir, "empty.c"), []byte(native.C(lowered)), 0644)','setup','^TestProduct_jsx_tree_lowered$','change the construction output filename program.c -> empty.c')
add('H9',jsx,'return native.Build(string(data), filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})','return native.Build(string(data), filepath.Join(dir, "native"), native.Options{Sanitize: !sanitize})','setup','^TestProduct_jsx_tree_native$','flip the sanitizer construction option')
for id,file in [('P1','stage1/typescript/parser/main.ts')]:
 s=read(file); a=s.index('\nfunction run(')
 add(id,file,s[a:],'\n','probe',tree,'empty-output probe: drop all executable module code after imports')
(out/'diffs').mkdir(exist_ok=True)
for m in specs:
 s=read(m['file']); changed=s.replace(m['old'],m['new'],1)
 diff=''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (out/'diffs'/f"{m['id']}.diff").write_text(diff)
(out/'plan.json').write_text(json.dumps(specs,indent=2)+'\n')
# Conservative source inventory, no claim that all functions were dynamically reached.
files=list((root/'stage1/typescript/parser').glob('*.ts'))+list((root/'stage1/typescript/scanner').glob('*.ts'))+list((root/'stage1/cohere/lint').glob('*.ts'))+list((root/'stage1/cohere/lint').glob('*.a'))
inv=[]
for p in files:
 for i,line in enumerate(p.read_text().splitlines(),1):
  if re.match(r'(export )?function \w+|    (constructor|[A-Za-z]\w*)\([^;]*',line): inv.append(f'{p.relative_to(root)}:{i}: {line.strip()}')
(out/'source-inventory.txt').write_text('Conservative declaration inventory. Exact dynamic reach was not established.\n'+ '\n'.join(inv)+'\n')
if len(sys.argv)<2:sys.exit()
ids=sys.argv[1:]; results=[]
for m in specs:
 if m['id'] not in ids:continue
 path=root/m['file']; original=read(m['file']); path.write_text(original.replace(m['old'],m['new'],1))
 try:
  env=os.environ.copy(); env['ADAMIC_NATIVE_SPLIT']='1'
  # Separate persistent product cache for each construction break as well as each port mutation.
  env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u109/cache/'+m['id']; env['ADAMIC_BUILD_CACHE']='on'
  start=time.monotonic()
  if m['file'].endswith('.go'):
   with (out/f"{m['id']}-vet.log").open('w') as f:v=subprocess.run(['timeout','90','go','vet',pkg],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
   if v.returncode: raise RuntimeError('vet failed '+m['id'])
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run',m['regex']]
  with (out/f"{m['id']}.log").open('w') as f:p=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  if m['id']=='M1':
   with (out/'M1-after.log').open('w') as f: subprocess.run(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),'/tmp/u109/written-witness.ts'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  results.append(dict(id=m['id'],exit=p.returncode,wall=time.monotonic()-start,command=cmd))
  (out/'run-results.json').write_text(json.dumps(results,indent=2)+'\n')
  print(m['id'],p.returncode,results[-1]['wall'],flush=True)
 finally:path.write_text(original)
