import pathlib,json,difflib,subprocess,time,os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/internal-lower-nested_functions'; os.chdir(root)
plans=[('M01','internal/lower/nested_functions.go','l.result.Locals[local].EnvironmentCell = true','l.result.Locals[local].EnvironmentCell = false','change constant'),('M02','internal/lower/nested_functions.go','function.Body = append([]ir.Statement{ir.AllocateEnvironment{Cells: slices.Clone(function.FrameEnvironment)}}, function.Body...)','','drop statement'),('M03','internal/lower/cycles.go','for _, local := range f.l.result.Functions[closure.function].Environment {\n\t\t\t\t\tqueue = append(queue, cycleNode{cell: local + 1})\n\t\t\t\t}','','drop whole loop'),('M04','internal/lower/nested_functions.go','if !implemented {','if implemented {','flip condition'),('W01','internal/lower/closed_frame_inputs.go','return safe && called','return safe || called','witness weakening')]
original={f:(root/f).read_text() for _,f,_,_,_ in plans}
for f in ['internal/lower/lower.go','internal/lower/modules.go','internal/lower/statements.go','internal/lower/functions.go']:original[f]=(root/f).read_text()
entries=[('PLower','internal/lower/lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil'),('PDeclareModule','internal/lower/modules.go','func (l *lowering) declareModule(statements []*ast.Node) error {','return nil'),('PStatements','internal/lower/statements.go','func (l *lowering) statements(nodes []*ast.Node) ([]ir.Statement, error) {','return nil, nil'),('PNestedDeclarations','internal/lower/nested_functions.go','func (l *lowering) nestedDeclarations(nodes []*ast.Node) ([]ir.Statement, error) {','return nil, nil'),('PLowerBody','internal/lower/functions.go','func (l *lowering) lowerBody(index int, declaration *ast.Node, this int, defaults []defaulted, patterns []patterned) error {','return nil'),('PClosedFrame','internal/lower/closed_frame_inputs.go','func (l *lowering) closedFrameInput(local int) bool {','return false')]
meta=[]
for ident,f,old,new,menu in plans:
 assert old in original[f]
 meta.append(dict(id=ident,file=f,line=original[f][:original[f].index(old)].count('\n')+1,old=old,new=new,menu=menu))
(out/'plan.json').write_text(json.dumps(meta,indent=2))
(out/'reached-functions.txt').write_text('\n'.join(x for x in (out/'coverage-functions.txt').read_text().splitlines() if not x.endswith('0.0%'))+'\n')
(out/'code-and-oracle.txt').write_text('Before mutations: code under test is internal/lower lowering and its reached functions inventoried in reached-functions.txt. Oracle is self-written diagnostic text and IR invariants. ClosedFrame is a built-in mutation witness, judged only by W01. Fixed menu and edits frozen in plan.json before outcomes. Four production mutants; W01 excluded from production matrix.\n')
(out/'diffs').mkdir(exist_ok=True)
validation=[]
for p in meta:
 f=p['file']; changed=original[f].replace(p['old'],p['new'],1)
 diff=''.join(difflib.unified_diff(original[f].splitlines(True),changed.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
 (out/'diffs'/f"{p['id']}.diff").write_text(diff)
 (root/f).write_text(changed)
 start=time.monotonic()
 with (out/f"vet-{p['id']}.log").open('w') as log:r=subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT)
 validation.append(dict(id=p['id'],status=r.returncode,seconds=time.monotonic()-start))
 (root/f).write_text(original[f])
 if r.returncode:raise RuntimeError('vet '+p['id'])
(out/'validation.json').write_text(json.dumps(validation,indent=2))
sw=dict(original)
for p in meta:
 old=p['old']; new=p['new']; ident=p['id']; f=p['file']
 if old.startswith('if '):replacement='if auditPick("'+ident+'", '+old[3:-2]+', '+new[3:-2]+') {'
 elif ident=='W01':replacement='return auditPick("W01", safe && called, safe || called)'
 else:replacement='if auditSelected("'+ident+'") { '+new+' } else { '+old+' }'
 sw[f]=sw[f].replace(old,replacement,1)
for ident,f,sig,ret in entries:sw[f]=sw[f].replace(sig,sig+'\n if auditSelected("'+ident+'") { '+ret+' }',1)
for f,s in sw.items():(root/f).write_text(s)
(root/'internal/lower/audit_switch.go').write_text('package lower\nimport "os"\nfunc auditSelected(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc auditPick(id string, normal, changed bool) bool { if auditSelected(id) {return changed}; return normal }\n')
subprocess.run(['gofmt','-w',*sw.keys(),'internal/lower/audit_switch.go'],check=True)
with (out/'switch-vet.log').open('w') as log:subprocess.run(['go','vet','./internal/lower/'],stdout=log,stderr=subprocess.STDOUT,check=True)
rows=json.loads((out/'rows.json').read_text()); runs=[]
for ident in [p['id'] for p in meta]+[e[0] for e in entries]:
 env=dict(os.environ,ADAMIC_MUTANT=ident,ADAMIC_BUILD_CACHE_DIR='/tmp/u039/cache/'+ident)
 regex='.' if ident.startswith('M') else '^('+ '|'.join(rows)+')$'
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex]
 start=time.monotonic(); path=out/(ident+'.log')
 with path.open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 runs.append(dict(id=ident,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd),log=path.name))
 if 'panic:' in path.read_text() or r.returncode==124:
  for row in rows:
   cmd[-1]='^'+row+'$';path=out/(ident+'-'+row+'.log');start=time.monotonic()
   with path.open('w') as log:rr=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
   runs.append(dict(id=ident,test=row,status=rr.returncode,wall=time.monotonic()-start,command=' '.join(cmd),log=path.name))
 (out/'runs.json').write_text(json.dumps(runs,indent=2))
for f,s in original.items():(root/f).write_text(s)
(root/'internal/lower/audit_switch.go').unlink()
print('complete',len(runs),'runs')
