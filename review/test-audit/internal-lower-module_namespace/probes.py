import pathlib,json,subprocess,os,time,difflib
os.chdir('/workspace/adamic');root=pathlib.Path('.');p=root/'review/test-audit/internal-lower-module_namespace';rows=json.loads((p/'rows.json').read_text());names=[r['test'] for r in rows]
plans=[('P01','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return &ir.Program{}, nil',names[:7]+[names[13]]),('P02','internal/lower/namespaces.go','func (l *lowering) namespaceInitialization(modules []*ast.SourceFile) error {','return nil',[names[6],names[10]]),('P03','internal/lower/namespaces.go','func namespaceRuntime(declaration *ast.Node) bool {','return false',[names[7]]),('P04','internal/lower/namespaces_call_graph.go','func (g *namespaceCallGraph) reach(function *ast.Node) map[*ast.Node]*ast.Node {','return nil',names[8:10]),('P05','internal/lower/modules.go','func (l *lowering) declareModule(statements []*ast.Node) error {','return nil',[names[11]]),('P06','internal/lower/namespaces.go','func (l *lowering) namespaceBody(node *ast.Node) ([]ir.Statement, error) {','return nil, nil',[names[11]]),('P07','internal/lower/locals.go','func (l *lowering) declareLocal(name *ast.Node) (int, error) {','return 0, nil',[names[12]])]
original={f:pathlib.Path(f).read_text() for _,f,_,_,_ in plans};switch=dict(original);results=[]
(p/'probes').mkdir(exist_ok=True)
try:
 for id,f,header,answer,tests in plans:
  before=original[f];assert before.count(header)==1
  after=before.replace(header,header+'\n\tif true { '+answer+' }')
  (p/'probes'/(id+'.diff')).write_text(''.join(difflib.unified_diff(before.splitlines(True),after.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
  condition='auditProbe("'+id+'")'
  if id in ['P05','P06']:condition+=' || auditProbe("P08")'
  switch[f]=switch[f].replace(header,header+'\n\tif '+condition+' { '+answer+' }')
  pathlib.Path(f).write_text(after)
  with (p/'logs'/('vet-'+id+'.log')).open('w') as out:r=subprocess.run(['timeout','90','go','vet','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT)
  pathlib.Path(f).write_text(before);assert r.returncode==0,id
 for f,t in switch.items():pathlib.Path(f).write_text(t)
 pathlib.Path('internal/lower/audit_probe.go').write_text('package lower\nimport "os"\nfunc auditProbe(id string)bool{return os.Getenv("ADAMIC_MUTANT")==id}\n')
 with (p/'logs'/'compile-probes.log').open('w') as out:subprocess.run(['go','test','-c','-o','/tmp/u037-probe.test','./internal/lower/'],stdout=out,stderr=subprocess.STDOUT,check=True)
 for id,f,header,answer,tests in plans+[('P08',None,None,None,[names[11]])]:
  for n in tests:
   cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+n+'$'];env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u037/cache/'+id);log=p/'logs'/(id+'-'+n+'.log');start=time.monotonic()
   with log.open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
   events=[]
   for s in log.read_text().splitlines():
    if s.startswith('{'):
     try:events.append(json.loads(s))
     except:pass
   results.append(dict(id=id,test=n,command=' '.join(cmd),exit=r.returncode,wall_seconds=time.monotonic()-start,events=events));(p/'probe-results.json').write_text(json.dumps(results,indent=2));print(id,n,r.returncode,flush=True)
finally:
 for f,t in original.items():pathlib.Path(f).write_text(t)
 pathlib.Path('internal/lower/audit_probe.go').unlink(missing_ok=True)
(p/'probe-plan.json').write_text(json.dumps([dict(id=id,file=f,entry=h,empty=a,rows=t) for id,f,h,a,t in plans],indent=2))
