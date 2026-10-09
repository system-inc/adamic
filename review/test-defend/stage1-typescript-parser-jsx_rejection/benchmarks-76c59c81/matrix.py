import os,pathlib,subprocess,time,json,difflib
R=pathlib.Path('/workspace/adamic');E=pathlib.Path('/tmp/d159/evidence');P='stage1/typescript/parser';base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=R,text=True).strip();meta=[]
countrows=['TestPerformance','TestWholePerformance','TestNodeCountCheckCatchesMutant','TestWholeCountCheckCatchesMutant']
setuprows=['TestCompilerExpressionsAgree_Setup','TestProduct_CompilerExpressionsNative']+['TestCompilerExpressionsAgree_%03d'%i for i in range(16)]
mutants=[dict(id='D1',test='TestPerformance',path=P+'/main.ts',old='    let count = 0;\n    for(const root',new='    let count = 0;\n    if(countOnly && !whole && parser.roots.length > 3) { return 0; }\n    for(const root',menu='return early',rows=countrows),dict(id='D2',test='TestWholePerformance',path=P+'/nodes.ts',old='    let count = 1;\n    for(const child of node.children) {\n        count += countTree',new='    let count = 1;\n    if(node.kind === \'InterfaceDeclaration\') { return 0; }\n    for(const child of node.children) {\n        count += countTree',menu='return early',rows=countrows),dict(id='D3',test='TestCompilerExpressionsAgree_Setup',path='internal/native/native.go',old='func Build(source string, output string, options Options) error {',new='func Build(source string, output string, options Options) error {\n\tif options.Sanitize { return fmt.Errorf("native: sanitized build unavailable") }',menu='return early',rows=setuprows),dict(id='D4',test='TestCompilerExpressionsAgree_Setup',path='internal/native/native.go',old='if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err != nil {',new='if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0o644); err == nil {',menu='flip condition',rows=setuprows),dict(id='D5',test='TestCompilerExpressionsAgree_Setup',path='internal/native/native.go',old='func Build(source string, output string, options Options) error {',new='func Build(source string, output string, options Options) error {\n\tif output != "" { return nil }',menu='return early',rows=setuprows)]
(E/'diffs').mkdir(exist_ok=True);(E/'menu.json').write_text(json.dumps(mutants,indent=2))
for log in ["baseline-count", "baseline-setup"]:
 events=[json.loads(l) for l in (E/(log+".log")).read_text().splitlines() if l.startswith("{")]
 assert any(e["Action"]=="pass" and "Test" not in e for e in events), "baseline not green: "+log
for m in mutants:
 id=m['id'];f=R/m['path'];original=f.read_text();assert original.count(m['old'])==1;changed=original.replace(m['old'],m['new']);m['line']=original[:original.index(m['old'])].count('\n')+1
 (E/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),'a/'+m['path'],'b/'+m['path'])));f.write_text(changed)
 try:
  env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u159/typescript',ADAMIC_PARSER_BENCH='1',ADAMIC_BUILD_CACHE_DIR='/tmp/d159/cache/'+id)
  if m['path'].endswith('.go'):
   with open(E/(id+'-vet.log'),'w') as out: subprocess.run(['go','vet','./internal/native/'],cwd=R,env=env,stdout=out,stderr=subprocess.STDOUT,check=True)
  pattern='^('+'|'.join(m['rows'])+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+P+'/', '-run',pattern];t=time.monotonic()
  with open(E/(id+'.log'),'w') as out:res=subprocess.run(cmd,cwd=R,env=env,stdout=out,stderr=subprocess.STDOUT)
  m.update(command=cmd,exit=res.returncode,wall=time.monotonic()-t,cache=env['ADAMIC_BUILD_CACHE_DIR']);meta.append(m);(E/'runs.json').write_text(json.dumps(meta,indent=2));print(id,res.returncode,round(m['wall'],3),flush=True)
 finally:f.write_text(original)
print('RESTORED',flush=True)
