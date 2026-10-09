import pathlib, subprocess, json, time, difflib, os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/internal-ir'; os.chdir(root)
rows=[x for x in (out/'u022-list.log').read_text().splitlines() if x.startswith('Test')]
spec=[
('M01','call_targets.go','if call.Virtual != 0 {','if call.Virtual == 0 {'),
('M02','call_targets.go','return targets','return targets[:1]'),
('M03','call_targets.go','if p.Functions[target].MayThrow {','if !p.Functions[target].MayThrow {'),
('M04','call_targets.go','[]int{call.Direct - 1}','[]int{call.Direct}'),
('M05','call_targets.go','[]int{target - 1}','[]int{target}'),
('M06','call_targets.go','return FunctionTargets{Unknown: true}','return FunctionTargets{Unknown: false}'),
('M07','call_targets.go','if function.MayThrow {','if !function.MayThrow {'),
('M08','argument_slots.go','start := len(function.Parameters) - 1','start := len(function.Parameters)'),
('M09','argument_slots.go','fixed = max(fixed, n)','fixed = min(fixed, n)'),
('M10','argument_slots.go','} else if of.IsMaybe() {','} else if !of.IsMaybe() {'),
('M11','argument_slots.go','layout.Count = layout.Count || p.PackedCountNeeded(target)','layout.Count = layout.Count && p.PackedCountNeeded(target)'),
('M12','argument_slots.go','if f.ReadsArguments || f.RestElement != 0 && (f.Closure || f.Receiver) {','if f.ReadsArguments && f.RestElement != 0 && (f.Closure || f.Receiver) {'),
('M13','view_primitives.go','if contract.Unsupported != "" {','if contract.Unsupported == "" {'),
('M14','view_primitives.go','if contract.Undefined {','if !contract.Undefined {'),
('M15','view_primitives.go','contract.Of = Number','contract.Of = Boolean'),
('M16','view_unions_untagged.go','|| seen[literal]','|| !seen[literal]'),
('M17','view_unions_untagged.go','own.Name != field.Name || own.Optional ||','own.Name != field.Name || !own.Optional ||'),
('M18','view_unions_untagged.go','if valid {','if !valid {'),
]
probes=[
('P01','call_targets.go','func (p *Program) CallTargets(call Call) []int {','return nil'),
('P02','call_targets.go','func (p *Program) CallMayThrow(call Call) bool {','return false'),
('P03','call_targets.go','func (p *Program) ClosureTargets(call Expression) FunctionTargets {','return FunctionTargets{}'),
('P04','call_targets.go','func (p *Program) ClosureMayThrow(call Expression) bool {','return false'),
('P05','argument_slots.go','func (p *Program) ClosureArgumentLayout(call CallClosure) ArgumentLayout {','return ArgumentLayout{}'),
('P06','view_primitives.go','func PrimitiveViewMembers(program *Program, id ViewContractID) ([]ViewContract, bool) {','return nil, false'),
('P07','view_unions_untagged.go','func ViewUnionHasDiscriminant(contracts []ViewContract, root ViewContract) bool {','return false'),
]
files={f: subprocess.check_output(['git','show','origin/main:internal/ir/'+f]).decode() for _,f,_,_ in spec+probes}
for f,original in files.items(): (root/'internal/ir'/f).write_text(original)
menu=[]
for mid,f,a,b in spec:
 assert a in files[f],(mid,a)
 line=files[f][:files[f].index(a)].count('\n')+1
 menu.append(dict(id=mid,file='internal/ir/'+f,line=line,before=a,after=b,kind='mutant'))
for mid,f,a,b in probes:
 menu.append(dict(id=mid,file='internal/ir/'+f,line=files[f][:files[f].index(a)].count('\n')+1,before=a,after=b,kind='probe'))
(out/'menu.json').write_text(json.dumps(menu,indent=2)+'\n')
(out/'scope.json').write_text(json.dumps(rows,indent=2)+'\n')
def run(cmd,name,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  result=subprocess.run(cmd,shell=True,stdout=log,stderr=subprocess.STDOUT,env=env)
 return dict(command=cmd,status=result.returncode,wall=time.monotonic()-start)
meta=json.loads((out/'timing-meta.json').read_text())
# Switch only the relevant original expression, never the tests.
for f,original in files.items():
 changed=original
 for mid,mf,a,b in spec:
  if mf!=f: continue
  if a.startswith('if ') or a.startswith('} else if '):
   prefix='} else if ' if a.startswith('}') else 'if '
   expr=a[len(prefix):-2]; new=b[len(prefix):-2]
   switched=prefix+'((os.Getenv("ADAMIC_MUTANT") == "'+mid+'" && ('+new+')) || (os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && ('+expr+'))) {'
  elif a.startswith('|| '):
   switched='|| ((os.Getenv("ADAMIC_MUTANT") == "'+mid+'" && !seen[literal]) || (os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && seen[literal]))'
  elif a.startswith('own.Name'):
   switched='own.Name != field.Name || ((os.Getenv("ADAMIC_MUTANT") == "'+mid+'" && !own.Optional) || (os.Getenv("ADAMIC_MUTANT") != "'+mid+'" && own.Optional)) ||'
  elif a.startswith('start :='):
   switched='start := len(function.Parameters) - 1; if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { start = len(function.Parameters) }'
  elif a.startswith('[]int'):
   switched='[]int{func() int { if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { return '+b[6:-1]+' }; return '+a[6:-1]+' }()}'
  else:
   switched='if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+b+' } else { '+a+' }'
  assert a in changed,(mid,a)
  changed=changed.replace(a,switched,1)
 for mid,pf,a,b in probes:
  if pf==f: changed=changed.replace(a,a+'\n if os.Getenv("ADAMIC_MUTANT") == "'+mid+'" { '+b+' }',1)
 changed=changed.replace('package ir\n','package ir\n\nimport "os"\n',1)
 (root/'internal/ir'/f).write_text(changed)
meta['switch-vet']=run('gofmt -w internal/ir/argument_slots.go internal/ir/call_targets.go internal/ir/view_primitives.go internal/ir/view_unions_untagged.go && timeout 120 go vet ./internal/ir/','switch-vet')
if meta['switch-vet']['status']!=0: raise RuntimeError('switch vet failed')
(out/'switch.diff').write_text(subprocess.check_output(['git','diff','--','internal/ir']).decode())
for mid,_,_,_ in spec+probes:
 env=os.environ.copy();env['ADAMIC_MUTANT']=mid
 meta[mid]=run('timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .',mid,env)
 content=(out/(mid+'.log')).read_text()
 if 'panic:' in content:
  for row in rows: meta[mid+'-'+row]=run(f"timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run '^{row}$'",mid+'-'+row,env)
 (out/'run-meta.json').write_text(json.dumps(meta,indent=2)+'\n')
for f,original in files.items(): (root/'internal/ir'/f).write_text(original)
# Construction test: rename an allowlisted reader, making its construction stale.
f=root/'internal/ir/call_targets_guard_test.go'; original=f.read_text(); changed=original.replace('internal/native/arguments_length.go:spreadArguments:Call.Function','internal/native/arguments_length.go:missingReader:Call.Function',1)
(out/'S01.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/ir/call_targets_guard_test.go',tofile='b/internal/ir/call_targets_guard_test.go')))
f.write_text(changed)
meta['S01']=run("timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run '^TestCallTargetReaders$'",'S01')
f.write_text(original)
(out/'run-meta.json').write_text(json.dumps(meta,indent=2)+'\n')
print('audit runs complete')
