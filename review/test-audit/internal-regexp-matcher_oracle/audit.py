import pathlib,json,re,subprocess,time,os,difflib
D=pathlib.Path('review/test-audit/internal-regexp-matcher_oracle'); P=pathlib.Path('internal/regexp')
original={str(p):p.read_text() for p in P.glob('*.go')}
rows=[s for s in (D/'test-list.log').read_text().splitlines() if s.startswith('Test')]
plan=[]
def add(id,file,old,new,menu):
 s=original['internal/regexp/'+file]; assert s.count(old)==1,(id,s.count(old)); plan.append(dict(id=id,file='internal/regexp/'+file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,menu=menu))
add('M01','matcher.go','b.steps >= b.limit','b.steps > b.limit','off-by-one bound')
add('M02','matcher.go','stateful := p.flags.Global || p.flags.Sticky','stateful := p.flags.Global && p.flags.Sticky','flip condition')
add('M03','matcher.go','r.LastIndex = uint64(state.pos)','r.LastIndex = uint64(state.pos + 1)','change constant')
add('M04','matcher.go','\t\tr.LastIndex = 0','\t\tr.LastIndex = 1','change constant')
add('M05','matcher.go','alternative.pc = i.y\n\t\t\tstack = append(stack, alternative)\n\t\t\ts.pc = i.x','alternative.pc = i.x\n\t\t\tstack = append(stack, alternative)\n\t\t\ts.pc = i.y','swap arguments')
add('M06','matcher.go','\t\t\t\tfor _, id := range i.clear {\n\t\t\t\t\tstate.caps[2*id] = -1\n\t\t\t\t\tstate.caps[2*id+1] = -1\n\t\t\t\t}','','drop whole loop')
add('M07','matcher.go','failed = ok == i.negative','failed = ok != i.negative','flip condition')
add('M08','matcher.go','if !f.DotAll {','if f.DotAll {','flip condition')
add('M09','matcher.go','if i.greedy {','if !i.greedy {','flip condition')
add('M10','matcher.go','return rune(c), pos + 1, true','return rune(c), pos + 2, true','off-by-one bound')
add('M11','canonicalize.go','if !f.IgnoreCase {','if f.IgnoreCase {','flip condition')
add('M12','sets.go','set.strings[index] = slices.Clone(text)','set.strings[index] = text','change option')
add('M13','sets.go','contains == intersection','contains != intersection','flip condition')
add('M14','sets.go','return i < len(s.ranges) && s.ranges[i].From <= c','return i < len(s.ranges) && s.ranges[i].From < c','off-by-one bound')
add('M15','parser.go','if f.Unicode && f.UnicodeSets {','if f.Unicode && !f.UnicodeSets {','flip condition')
add('M16','parser.go','limit = 2','limit = 3','change constant')
add('M17','native_search.go','return [2]uint64{a[0] | b[0], a[1] | b[1]}, ok && yes','return [2]uint64{a[0] & b[0], a[1] & b[1]}, ok && yes','change option') if False else None
add('M17','sets.go','b[k].From - 1','b[k].From','off-by-one bound')
add('M18','matcher.go','if p.anchored || p.flags.Sticky {','if p.anchored && p.flags.Sticky {','flip condition')
(D/'plan.json').write_text(json.dumps(plan,indent=2))
(D/'origin.txt').write_text(subprocess.check_output(['git','rev-parse','HEAD']).decode())
(D/'inventory.txt').write_text('CODE UNDER TEST: Adamic regexp Parse/ParseUTF16, compilation and RegExp.Exec/Program.run. ORACLES: live Node24, stored Node-generated test262 executions, self-written budget checks. Witness: Oct6 mutation checker.\nReached functions: see functions.txt, positive coverage entries. Complete conservative package function inventory:\n'+'\n'.join(f'{f}:{i}: {line}' for f,s in original.items() if not f.endswith('_test.go') for i,line in enumerate(s.splitlines(),1) if line.startswith('func ')))
commands=[]
def run(label,args,env=None):
 t=time.monotonic(); e=os.environ.copy(); e.update(env or {}); e['ADAMIC_BUILD_CACHE_DIR']='/tmp/u073/cache/'+label
 with (D/(label+'.log')).open('w') as out: rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT,env=e)
 commands.append(dict(label=label,command=' '.join(args),env=env or {},wall_seconds=round(time.monotonic()-t,3),exit=rc)); (D/'commands.json').write_text(json.dumps(commands,indent=2)); print(label,rc,flush=True)
for row in []:
 for i in range(3): run(f'time-{row}-{i+1}',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','^'+row+'$'])
for m in plan:
 f=m['file']; s=original[f]; new=s.replace(m['old'],m['new']); (D/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
 pathlib.Path(f).write_text(new); run('vet-'+m['id'],['timeout','90','go','vet','./internal/regexp/']); pathlib.Path(f).write_text(s)
# Selector instrumentation, not part of standalone mutants.
for f,s in original.items():
 for m in plan:
  if m['file']!=f:continue
  old,new=m['old'],m['new']
  if m['id']=='M06': replacement='if !auditSelected("M06") {'+old+'\n}'
  elif m['id']=='M17': replacement='auditRune("M17", '+old+', '+new+')'
  elif '\n' in old: replacement='if auditSelected("'+m['id']+'") {'+new+'\n} else {'+old+'\n}'
  elif old.startswith('stateful :='): replacement='stateful := auditBool("M02", p.flags.Global || p.flags.Sticky, p.flags.Global && p.flags.Sticky)'
  elif old.startswith('return '): replacement='if auditSelected("'+m['id']+'") { '+new+' }; '+old
  elif old.startswith('set.strings[index]') or old.startswith('r.LastIndex') or old.startswith('\t\tr.LastIndex') or old.startswith('limit =') or old.startswith('failed ='): replacement='if auditSelected("'+m['id']+'") {'+new+'} else {'+old+'}'
  else: replacement='auditBool("'+m['id']+'", '+old+', '+new+')'
  # whole if headers must keep syntax
  if old.startswith('if '): replacement='if auditBool("'+m['id']+'", '+old[3:-2]+', '+new[3:-2]+') {'
  s=s.replace(old,replacement)
 if f.endswith('/matcher.go'):
  s=s.replace('func (r *RegExp) Exec(input []uint16) (*Match, error) {','func (r *RegExp) Exec(input []uint16) (*Match, error) {\nif auditSelected("PExec") { return nil,nil }')
  s=s.replace('func (p *Program) run(input []uint16, pos int, caps []int, budget *executionBudget) (machineState, bool, error) {','func (p *Program) run(input []uint16, pos int, caps []int, budget *executionBudget) (machineState, bool, error) {\nif auditSelected("PRun") { return machineState{},false,nil }')
 if f.endswith('/parser.go'):s=s.replace('func Parse(pattern, flags string) (*Pattern, error) {','func Parse(pattern, flags string) (*Pattern, error) {\nif auditSelected("PParse") { return nil,nil }')
 pathlib.Path(f).write_text(s)
helper=P/'audit_selector.go';helper.write_text('package regexp\nimport "os"\nfunc auditSelected(id string) bool{return os.Getenv("ADAMIC_MUTANT")==id}\nfunc auditBool(id string, clean, mutant bool) bool{if auditSelected(id){return mutant};return clean}\nfunc auditRune(id string, clean, mutant rune) rune{if auditSelected(id){return mutant};return clean}\n')
run('switch-vet',['timeout','90','go','vet','./internal/regexp/'])
for m in plan:
 run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','.'],{'ADAMIC_MUTANT':m['id']})
 data=(D/(m['id']+'.log')).read_text()
 if 'panic:' in data:
  for row in rows: run(m['id']+'-alone-'+row,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','^'+row+'$'],{'ADAMIC_MUTANT':m['id']})
for id,regex in [('PExec','^TestMatcher'),('PRun','^TestMatcherStepLimitBoundary$'),('PParse','^TestNodeAgreement$')]:run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run',regex],{'ADAMIC_MUTANT':id})
for f,s in original.items():pathlib.Path(f).write_text(s)
helper.unlink()
# Witness check weakened to always report agreement, preserving the control.
f='internal/regexp/oct6_mutant_test.go';s=original[f];new=s.replace('if reflect.DeepEqual(got, expected[mutant.caseIndex]) {','if true {').replace('if !reflect.DeepEqual(got, want[mutant.index][step]) {','if false {')
(D/'W01.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),new.splitlines(True),fromfile='a/'+f,tofile='b/'+f)))
pathlib.Path(f).write_text(new);run('W01',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/regexp/','-run','^TestMatcherOct6Mutants$']);pathlib.Path(f).write_text(s)
