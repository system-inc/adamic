import pathlib,os,json,subprocess,time,difflib
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-typescript-parser-jsx_rejection';(p/'diffs').mkdir(exist_ok=True)
plan=[
 ('D1','stage1/typescript/parser/jsx.ts','JSX expected name at','JSX invalid name at','member diagnostic class, contrasted with private-member boundary'),
 ('D2','stage1/typescript/scanner/scanner.ts','while(this.code() === 45 || isIdentifierPart(this.code()))','while(this.code() === 46 || isIdentifierPart(this.code()))','dash extension used by the permissive member variant versus legal names'),
 ('D3','stage1/typescript/parser/jsx.ts','if(tag) {','if(!tag) {','member descent versus plain tag and attribute names'),
 ('D4','stage1/typescript/parser/jsx.ts',"name = this.make('ThisKeyword', pos);","name = this.make('ThisKeyword', pos + 1);",'Node/native shared source, this-tag position boundary'),
 ('D5','stage1/typescript/parser/jsx.ts',"this.parser.node(id).end = this.parser.scanner.fullStart;","this.parser.node(id).end = this.parser.scanner.fullStart + 1;",'Node/native shared text-end byte boundary, including Unicode and CRLF'),
 ('D6','stage1/typescript/parser/jsx.ts',"const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [name, right]);","const right = this.identifier();\n            return this.make('JsxNamespacedName', pos, [right, name]);",'Node/native shared namespace input, child order')]
items=[]
for id,file,old,new,aim in plan:
 s=(root/file).read_text();assert s.count(old)==1,(id,s.count(old));line=s[:s.index(old)].count('\n')+1
 diff=''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file));(p/'diffs'/(id+'.diff')).write_text(diff);items.append(dict(id=id,file=file,line=line,old=old,new=new,aim=aim))
(p/'plan.json').write_text(json.dumps(items,indent=2))
env=os.environ.copy();env['TMPDIR']='/workspace/scratch/defend-jsx/tmp';env['ADAMIC_TYPESCRIPT_SOURCE']='/workspace/scratch/u146/pinned-typescript';env['ADAMIC_PARSER_BENCH']='1'
# All callers of jsxManifest plus rejection fixtures and their product wrappers.
rows=[l for l in (p/'list.log').read_text().splitlines() if l.startswith('TestJsx') or l.startswith('TestProduct_Jsx')];(p/'matrix-rows.json').write_text(json.dumps(rows,indent=2))
runs=[]
def run(id,label,pattern,cache):
 env['ADAMIC_BUILD_CACHE_DIR']=cache
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run',pattern];start=time.monotonic();log=id+'-'+label+'.log'
 with (p/log).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env,cwd=root)
 events=[]
 for l in (p/log).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 runs.append(dict(id=id,label=label,command=cmd,log=log,wall=time.monotonic()-start,exit=r.returncode,failed=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],passed=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],errors=[e.get('Output') for e in events if e.get('OutputType')=='error'],timeout=any('panic: test timed out' in e.get('Output','') for e in events)))
 (p/'runs.json').write_text(json.dumps(runs,indent=2));print(id,label,r.returncode,round(time.monotonic()-start,2),flush=True)
 return runs[-1]
# Clean matrix control, in separately bounded groups, before any production change.
for name in ['TestJsxNameBoundaryRejections','TestJsxScannerMutants']:
 r=run('clean',name,'^'+name+'$','/workspace/scratch/defend-jsx/cache/clean');assert r['exit']==0
r=run('clean','new-shards','^(TestJsxMutants.*|TestProduct_Jsx.*)$','/workspace/scratch/defend-jsx/cache/clean')
assert not r['failed'], 'red clean baseline'
for m in items:
 id=m['id'];diff=p/'diffs'/(id+'.diff');subprocess.run(['git','apply',str(diff)],cwd=root,check=True)
 try:
  cache='/workspace/scratch/defend-jsx/cache/'+id
  r=run(id,'matrix','^('+'|'.join(rows)+')$',cache)
  # Re-observe all defender rows if the aggregate run is cooked or incomplete.
  completed=set(r['failed']+r['passed'])
  for name in ['TestJsxMemberNameRejection','TestJsxNode','TestJsxNative']:
   if name not in completed:run(id,name,'^'+name+'$',cache)
 finally:subprocess.run(['git','apply','-R',str(diff)],cwd=root,check=True)
