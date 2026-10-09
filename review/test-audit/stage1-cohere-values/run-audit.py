import pathlib, subprocess, os, time, json, difflib, re
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-values'; os.chdir(root)
base=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()
rows=['TestThePortParsesAsGoCohereDoes','TestEachGapStandsWhereGapsMdSaysItDoes']
menu=[('M1','stage1/cohere/values/tokenize.ts','let offset = -1;','let offset = 0;','change constant'),('M2','stage1/cohere/values/nodes.ts',"sourceIndex, unit, false, -1);","sourceIndex, '', false, -1);",'change option'),('M3','stage1/cohere/values/parser.ts',"pair === '/*' || pair === '*/'","pair === '/*' || pair === '**'",'change constant'),('M4','internal/native/emit_strings.go','"adamic_string_index_of_from(%s, %s, %s)", value, arguments[0], arguments[1]','"adamic_string_index_of_from(%s, %s, %s)", value, arguments[0], "0.0"','change option')]
original={f:(root/f).read_text() for _,f,*_ in menu}
plan=[]
for id,f,a,b,kind in menu:
 s=original[f]; assert s.count(a)==1,(id,s.count(a)); line=s[:s.index(a)].count('\n')+1
 plan.append(dict(id=id,file=f,line=line,change=a+' -> '+b,menu=kind))
 (out/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(a,b,1).splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
(out/'plan.json').write_text(json.dumps(dict(base=base,code='TypeScript values parser port; compiler lowering/native emission for gap fixtures',oracle='Go cohere, upstream postcss-values-parser 2.0.1, Node; hand-written gap refusal label',mutants=plan),indent=2))
functions=[]
for f in ['values.ts','tokenize.ts','parser.ts','nodes.ts','main.ts']:
 for n,line in enumerate((root/'stage1/cohere/values'/f).read_text().splitlines(),1):
  if re.search(r'function \w+|^\s*(?:#\w+|parse|constructor)\(',line):functions.append(f+':'+str(n)+' '+line.strip())
(out/'functions-static.txt').write_text('Static declaration inventory, not a proof of dynamic reach. Main count-only functions are not reached by default tests. Compiler entry Lower, native C/Build, emitter.stringCall.\n'+'\n'.join(functions))
env=dict(os.environ,ADAMIC_VALUES_LIBRARY='/tmp/u148-library'); records=[]
def run(name,cmd,extra={}):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(cmd,shell=True,stdout=log,stderr=subprocess.STDOUT,env=dict(env,**extra))
 events=[]
 for line in (out/(name+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 record=dict(name=name,command=cmd,wall=time.monotonic()-start,exit=r.returncode,fail=[e.get('Test') for e in events if e.get('Action')=='fail' and e.get('Test')],elapsed=next((e.get('Elapsed') for e in reversed(events) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None))
 records.append(record); (out/'runs.json').write_text(json.dumps(records,indent=2)); print(name,record['exit'],round(record['wall'],3),record['elapsed'],flush=True)
 return record
for row in rows:
 for i in range(3):run(row+'-'+str(i+1),f"timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run '^{row}$'")
for id,f,a,b,kind in menu:
 try:
  (root/f).write_text(original[f].replace(a,b,1))
  extra={'ADAMIC_BUILD_CACHE_DIR':'/tmp/u148/cache/'+id}
  if f.endswith('.ts'):run(id+'-build',f'timeout 90 go run ./cmd/adamic build stage1/cohere/values/main.ts -o /tmp/u148/{id} --sanitize',extra)
  else:run(id+'-vet','timeout 90 go vet ./internal/native/',extra)
  run(id,'timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .',extra)
 finally:(root/f).write_text(original[f])
# Valid empty parse result, no harness changes.
f='stage1/cohere/values/values.ts';s=(root/f).read_text(); start=s.index('\n\tconst tokenized'); p=s[:start]+"\n\treturn new Parser(loose, []).parse();\n}\n"
(out/'P1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),p.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
try:
 (root/f).write_text(p);run('P1-build','timeout 90 go run ./cmd/adamic build stage1/cohere/values/main.ts -o /tmp/u148/P1 --sanitize');run('P1','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .')
finally:(root/f).write_text(s)
# Witness-only harness exception.
f='stage1/cohere/values/values_test.go';s=(root/f).read_text();start=s.index('func firstDifference(');end=s.index('\n// lowered',start);p=s[:start]+'func firstDifference(got string, want string) string {\n\treturn ""\n}\n'+s[end:]
(out/'W1.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),p.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
try:
 (root/f).write_text(p);run('W1-vet','timeout 90 go vet ./stage1/cohere/values/');run('W1','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .')
finally:(root/f).write_text(s)
run('restored','timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/values/ -run .')
