import pathlib,os,subprocess,time,json,re,difflib
root=pathlib.Path('/workspace/adamic');os.chdir(root);p=root/'review/test-audit/stage1-typescript-parser-setup_deadline_products';base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
rows={
'TestProduct_ParserOracle':['TestProduct_ParserOracle'],
'TestProduct_CompilerExpressionsLower':['TestProduct_CompilerExpressionsLower'],
'TestProduct_CompilerExpressionsNative':['TestProduct_CompilerExpressionsNative'],
'TestProduct_WholeMutantsLower family':['TestProduct_WholeMutantsLower_'+x for x in ['Control','000','001','002']],
'TestProduct_WholeMutantsNative family':['TestProduct_WholeMutantsNative_'+x for x in ['Control','000','001','002']],
'TestEveryTypeNodeKindAgrees':['TestEveryTypeNodeKindAgrees'],
'TestWholeGeneratedAgrees':['TestWholeGeneratedAgrees'],
'TestObsoleteImportAttributesAgrees':['TestObsoleteImportAttributesAgrees'],
'TestWholeMutants_Setup':['TestWholeMutants_Setup']}
names=sum(rows.values(),[]); listed=(p/'list.log').read_text().splitlines();assert all(n in listed for n in names)
pattern='^('+'|'.join(names)+')$'
env=dict(os.environ,ADAMIC_TYPESCRIPT_SOURCE='/tmp/u157-typescript');records=[]
def run(name,cmd,extra={}):
 started=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(cmd,shell=True,stdout=f,stderr=subprocess.STDOUT,env=dict(env,**extra))
 e=[]
 for line in (p/(name+'.log')).read_text().splitlines():
  try:
   event=json.loads(line)
   if isinstance(event,dict):e.append(event)
  except:pass
 d=dict(name=name,command=cmd,environment=extra,wall=time.monotonic()-started,exit=r.returncode,fail=[x.get('Test') for x in e if x.get('Action')=='fail' and x.get('Test')],elapsed=next((x.get('Elapsed') for x in reversed(e) if not x.get('Test') and x.get('Action') in ['pass','fail']),None),cooked=any('test timed out' in x.get('Output','') for x in e) or r.returncode==124)
 records.append(d);(p/'runs.json').write_text(json.dumps(records,indent=2));print(name,r.returncode,round(d['wall'],3),d['elapsed'],flush=True);return d
menu=[('M1','stage1/typescript/parser/nodes.ts','optional = false;','optional = true;','change constant','production'),('M2','stage1/typescript/parser/parser.ts',"this.make('JSDocVariadicType', pos, [type])","this.make('JSDocNullableType', pos, [type])",'change constant','production'),('M3','stage1/typescript/parser/statements.ts','node.operator = operator;',"node.operator = 'WithKeyword';",'change option','production'),('M4','stage1/typescript/parser/nodes.ts',"export function countTree(nodes: readonly ParseNode[], root: number): number {\n    const node = nodes[root] ?? panic('missing parse node');\n    let count = 1;","export function countTree(nodes: readonly ParseNode[], root: number): number {\n    const node = nodes[root] ?? panic('missing parse node');\n    let count = 0;",'change constant','production'),('S1','stage1/typescript/parser/whole_mutants_split_test.go','filepath.Join(dir, "oracle"), virtual','filepath.Join(dir, "absent/oracle"), virtual','change option','construction'),('S2','stage1/typescript/parser/compiler_expressions_shards_test.go','os.WriteFile(filepath.Join(dir, "main.c"),','os.WriteFile(filepath.Join(dir, "absent/main.c"),','change option','construction'),('S3','stage1/typescript/parser/compiler_expressions_shards_test.go','native.Build(string(source), filepath.Join(dir, "parser"),','native.Build(string(source), filepath.Join(dir, "absent/parser"),','change option','construction'),('S4','stage1/typescript/parser/whole_mutants_split_test.go','os.WriteFile(filepath.Join(dir, "main.c"),','os.WriteFile(filepath.Join(dir, "absent/main.c"),','change option','construction'),('S5','stage1/typescript/parser/whole_mutants_split_test.go','native.Build(string(source), filepath.Join(dir, "scanner"),','native.Build(string(source), filepath.Join(dir, "absent/scanner"),','change option','construction')]
f='stage1/typescript/parser/whole_mutants_setup_test.go';s=(root/f).read_text();a=s[s.index('\tfor i, source := range directories {\n\t\tif binaries'):s.index('\twholeMutantsPrepared = inputs')];menu.append(('S6',f,a,'','drop whole loop','construction'))
original={f:(root/f).read_text() for _,f,*_ in menu};plan=[]
for id,f,a,b,kind,role in menu:
 s=original[f];assert s.count(a)==1,(id,s.count(a));line=s[:s.index(a)].count('\n')+1
 plan.append(dict(id=id,file=f,line=line,from_text=a,to_text=b,menu=kind,role=role));(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(a,b,1).splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
(p/'plan.json').write_text(json.dumps(dict(base=base,rows=rows,matrix_pattern=pattern,mutants=plan),indent=2))
functions=[]
for f in (root/'stage1/typescript/parser').glob('*.ts'):
 for number,line in enumerate(f.read_text().splitlines(),1):
  if re.search(r'function \w+|^    \w+\([^;]*\).*\{',line):functions.append(f.name+':'+str(number)+' '+line.strip())
(p/'functions-static.txt').write_text('Static declaration inventory, not proof of runtime reach. Preparation functions and their callers were read in their complete test files.\n'+'\n'.join(functions))
# Timings grouped over the family, not summed member times.
for row,members in rows.items():
 pat='^('+'|'.join(members)+')$'
 for i in range(3):run(row.replace(' ','_')+'-'+str(i+1),f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '{pat}'")
# Witness for the count-only behavior, before mutations.
w=pathlib.Path('/tmp/u157');w.mkdir(exist_ok=True);(w/'witness.ts').write_text('const x = 1;\n');(w/'manifest.txt').write_text(str(w/'witness.ts')+'\n')
run('count-before','node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/parser/main.ts --manifest /tmp/u157/manifest.txt --whole --count')
for id,f,a,b,kind,role in menu:
 try:
  (root/f).write_text(original[f].replace(a,b,1));extra={'ADAMIC_BUILD_CACHE_DIR':'/tmp/u157/cache/'+id}
  if role=='production':
   build=run(id+'-build',f'timeout 90 go run ./cmd/adamic build stage1/typescript/parser/main.ts -o /tmp/u157/{id} --sanitize',extra)
   if build['exit']:continue
  else:
   vet=run(id+'-vet','timeout 90 go vet ./stage1/typescript/parser/',extra)
   if vet['exit']:continue
  d=run(id,f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '{pattern}'",extra)
  if d['cooked']:
   # Bound reruns to semantic rows for production; direct recipe consumers for construction.
   subset=names[11:14] if role=='production' else names
   for n in subset:run(id+'-'+n,f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^{n}$'",extra)
  if id=='M4':
   run('count-after','node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/typescript/parser/main.ts --manifest /tmp/u157/manifest.txt --whole --count')
   run('count-after-native','ASAN_OPTIONS=detect_leaks=0 /tmp/u157/M4 --manifest /tmp/u157/manifest.txt --whole --count')
 finally:(root/f).write_text(original[f])
# Parser entry probes: a valid empty file tree and no doc types.
f='stage1/typescript/parser/parser.ts';s=(root/f).read_text()
for id,begin,end,body in [('P1','    file(): number {','\n}',"    file(): number {\n        const root = this.make('SourceFile', 0, []);\n        this.node(root).end = this.scanner.text.length;\n        return root;\n    }"),('P2','    docTypes(): number[] {','    tupleNameAhead()',"    docTypes(): number[] {\n        return [];\n    }\n")]:
 start=s.index(begin);stop=s.index(end,start);changed=s[:start]+body+s[stop:]
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 try:
  (root/f).write_text(changed);run(id+'-build',f'timeout 90 go run ./cmd/adamic build stage1/typescript/parser/main.ts -o /tmp/u157/{id} --sanitize');run(id,f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '^(TestEveryTypeNodeKindAgrees|TestWholeGeneratedAgrees|TestObsoleteImportAttributesAgrees)$'",{'ADAMIC_BUILD_CACHE_DIR':'/tmp/u157/cache/'+id})
 finally:(root/f).write_text(s)
for diff in p.glob('*.diff'):run(diff.stem+'-apply',f'git read-tree {base} && git apply --cached --check {diff.relative_to(root)}',{'GIT_INDEX_FILE':'/tmp/u157-index'})
run('restored',f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/typescript/parser/ -run '{pattern}'")
