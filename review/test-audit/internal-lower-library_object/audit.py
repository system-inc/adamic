import pathlib,subprocess,json,time,difflib,os,re
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-lower-library_object';os.chdir(root)
rows=pathlib.Path('/tmp/u036-rows.txt').read_text().splitlines();scope=(out/'u036-list.log').read_text().splitlines();assert all(r in scope for r in rows)
spec=[
('M01','library_object.go','Fix: reason','Fix: ""'),
('M02','library_object.go','count := 1','count := 2'),
('M03','library_object.go','if depth > 16 {','if depth > 0 {'),
('M04','library_object.go','!l.hasProperty(written[0], key.Text())','l.hasProperty(written[0], key.Text())'),
('M05','library_object.go',"Object.groupBy's partial record", "Object.collectBy's partial record"),
('M06','library_object.go','declaration.Parent.Flags&ast.NodeFlagsConst == 0','declaration.Parent.Flags&ast.NodeFlagsConst != 0'),
('M07','library_regex_callback_shape.go','n.Kind != regex.Capturing','n.Kind == regex.Capturing'),
('M08','library_regex_callback_shape.go','if depth > 32 {','if depth > 0 {'),
('M09','library_regex_callback_shape.go','declaration.Parent.Flags&ast.NodeFlagsConst == 0','declaration.Parent.Flags&ast.NodeFlagsConst != 0'),
('M10','invariance.go','mutable := !l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet")','mutable := l.isLibraryType(to, "ReadonlyArray", "ReadonlyMap", "ReadonlySet")'),
('M11','invariance.go','if !fresh && !l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {','if !fresh && l.checker.IsReadonlySymbol(viewed) && l.checker.IsReadonlySymbol(inside) {'),
('M12','invariance.go','return l.checker.IsArrayType(l.checker.GetTypeAtLocation(access.Expression)) || l.isLibraryType(l.checker.GetTypeAtLocation(access.Expression), "ReadonlyArray")','return false'),
('M13','refusals.go','if node.Kind == ast.KindPropertyAccessExpression && !called(node) && !l.libraryNumberBoundMethod(node)','if node.Kind == ast.KindPropertyAccessExpression && called(node) && !l.libraryNumberBoundMethod(node)'),
('M14','prelude.go','return ir.Stderr, nil','return ir.Stdout, nil'),
('M15','diagnostics.go','%s: Adamic 0.1 refuses %s; %s','%s: Adamic 0.1 accepts %s; %s'),
('M16','diagnostics.go',"%s: stage 0 can't lower %s yet", "%s: stage 0 can't lower %s now"),
('M17','object.go','if value == nil || target == nil || value == target || depth > 4 {','if value == nil || target == nil || value == target || depth > 0 {'),
('M18','statements.go','return nil, l.notYet(node, "a class inside a function")','return nil, l.notYet(node, "a nested class")'),
('M19','library_object.go','call.Method == "freeze"','call.Method != "freeze"'),
('M20','prelude.go','if len(arguments) != 1 {','if len(arguments) == 1 {'),
]
# Each mutant is defined by a substitution in production code. Entry probes are separate.
probes=[('P01','lower.go','func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {','return nil, nil'),('P02','library_regex_callback_shape.go','func (l *lowering) regexCallbackNoCaptures(node *ast.Node, depth int) bool {','return false')]
files={f:subprocess.check_output(['git','show','origin/main:internal/lower/'+f]).decode() for _,f,_,_ in spec+probes}
menu=[]
for mid,f,a,b in spec+probes:
 assert a in files[f],(mid,a)
 menu.append(dict(id=mid,file='internal/lower/'+f,line=files[f][:files[f].index(a)].count('\n')+1,before=a,after=b,kind='probe' if mid.startswith('P') else 'mutant'))
(out/'menu.json').write_text(json.dumps(menu,indent=2)+'\n');(out/'scope.json').write_text(json.dumps(rows,indent=2)+'\n')
(out/'PRE-MUTATION.md').write_text('Starting commit: e2492670b06a4dce1deafe837158ce0366cf2bc4.\nCode under test: Adamic lowering, object/refusal/type-view logic, console IR production, diagnostic formatting, and the direct regex capture-free proof.\nOracle: handwritten error classes and diagnostic strings, handwritten IR expectations, or rejection/non-rejection predicates, all self. No outside authority was checked.\nComplete pre-mutation reached-function inventory: reached-functions.txt (447 coverage-observed functions in internal/lower). Anonymous visitors are included within their owning function coverage.\nFixed menu: menu.json, written before checking mutant failures. All M01-M20 are condition flips, changed constants, off-by-one bounds or early returns from production behavior. P01/P02 are empty entry probes, excluded from production kills.\nAll 11 names exist in the starting test list and remain in the four assigned files. Each is independent: shared lowerSource performs preparation, but their assertions and checks differ. No families, witnesses or helpers in the assigned set.\n')
meta={}
def run(cmd,name,env=None):
 t=time.monotonic()
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(cmd,shell=True,stdout=log,stderr=subprocess.STDOUT,env=env)
 v=dict(command=cmd,status=r.returncode,wall=time.monotonic()-t);meta[name]=v
 (out/'run-meta.json').write_text(json.dumps(meta,indent=2)+'\n')
 return v
for row in rows:
 for n in range(3):run(f"timeout 120 go test -count=1 -timeout 90s ./internal/lower/ -run '^{row}$'",f'timing-{row}-{n}')
# Vet every standalone substitution against the same origin/main source.
for mid,f,a,b in spec:
 changed=files[f].replace(a,b,1)
 (out/(mid+'.diff')).write_text(''.join(difflib.unified_diff(files[f].splitlines(True),changed.splitlines(True),fromfile='a/internal/lower/'+f,tofile='b/internal/lower/'+f)))
 (root/'internal/lower'/f).write_text(changed)
 v=run('timeout 120 go vet ./internal/lower/',mid+'-vet')
 (root/'internal/lower'/f).write_text(files[f])
 if v['status']!=0:raise RuntimeError(mid+' standalone vet failed')
# Generic switches operate on each original line, combining same-line edits.
for f,s in files.items():
 lines=s.splitlines(True); result=[]
 for line in lines:
  edits=[(mid,a,b) for mid,mf,a,b in spec if mf==f and a in line]
  if edits:
   # All relevant Go statements fit one line. For fragments within conditions or literals,
   # switch the complete statement line so no variable declarations change scope.
   stripped=line.strip();indent=line[:len(line)-len(line.lstrip())]
   if stripped.startswith('if ') and stripped.endswith('{'):
    expr=stripped[3:-2]
    initial=''
    if ';' in expr: initial,expr=expr.split(';',1); initial+='; '
    choices=[]
    for mid,a,b in edits:
     new=stripped.replace(a,b,1)[3:-2]
     if ';' in new: new=new.split(';',1)[1]
     choices.append(f'(os.Getenv("ADAMIC_MUTANT") == "{mid}" && ({new}))')
    normal=' && '.join(f'os.Getenv("ADAMIC_MUTANT") != "{mid}"' for mid,_,_ in edits)
    result.append(indent+'if '+initial+'('+' || '.join(choices+[f'(({normal}) && ({expr}))'])+') {\n')
   elif ':=' in stripped:
    # Stable variable scope for the two declaration mutations.
    mid,a,b=edits[0];left,right=stripped.split(':=',1);newright=stripped.replace(a,b,1).split(':=',1)[1]
    result.append(indent+left+':='+right+'\n'+indent+f'if os.Getenv("ADAMIC_MUTANT") == "{mid}" {{ '+left.strip()+' = '+newright.strip()+' }\n')
   elif 'return ' in stripped:
    for mid,a,b in edits:
     changed=stripped.replace(a,b,1);result.append(indent+f'if os.Getenv("ADAMIC_MUTANT") == "{mid}" {{ '+changed+' }\n')
    result.append(line)
   else:raise RuntimeError((f,line,edits))
  else:result.append(line)
 changed=''.join(result)
 for mid,pf,a,b in probes:
  if pf==f:changed=changed.replace(a,a+'\n if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+b+' }',1)
 if f=='lower.go':changed=changed.replace('"context"','"context"\n"os"',1)
 else:changed=changed.replace('import (','import (\n"os"',1)
 (root/'internal/lower'/f).write_text(changed)
v=run('gofmt -w '+ ' '.join('internal/lower/'+f for f in files)+' && timeout 120 go vet ./internal/lower/','switch-vet')
if v['status']!=0:raise RuntimeError('switch vet failed')
(out/'switch.diff').write_text(subprocess.check_output(['git','diff','--','internal/lower']).decode())
pattern='^('+'|'.join(rows)+')$'
for mid,_,_,_ in spec+probes:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u036/cache/'+mid
 v=run('timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .',mid,env)
 content=(out/(mid+'.log')).read_text()
 cooked=v['status']==124 or 'test timed out after' in content
 if cooked:
  run("timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '"+pattern+"'",mid+'-bounded',env)
  content=(out/(mid+'-bounded.log')).read_text()
 if 'panic:' in content or 'panic:' in (out/(mid+'.log')).read_text():
  for row in rows:run(f"timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '^{row}$'",mid+'-'+row,env)
 print(mid,v['status'],round(v['wall'],3),flush=True)
for f,s in files.items():(root/'internal/lower'/f).write_text(s)
print('restored',flush=True)
