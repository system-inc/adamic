import pathlib,subprocess,os,json,re,time,difflib
root=pathlib.Path.cwd(); p=root/'review/test-defend/internal-native-element_borrow/borrow-plans'; source=root/'internal/native/element_borrow.go'; original=source.read_text(); refs='origin/test-audit/internal-native-element_borrow:'; audit='review/test-audit/internal-native-element_borrow/'
current=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]; oldlog=subprocess.check_output(['git','show',refs+audit+'list.log']).decode();(p/'audit-list.log').write_text(oldlog);old=[x for x in oldlog.splitlines() if x.startswith('Test')];added=sorted(set(current)-set(old));(p/'scope.json').write_text(json.dumps({'current':current,'audit':old,'added':added,'vanished':sorted(set(old)-set(current))},indent=2)+'\n')
rows=set(added);callers={}
for f in (root/'internal/native').glob('*_test.go'):
 s=f.read_text()
 if re.search(r'\b(C|planElementBorrows|unchanging)\(',s):
  names=re.findall(r'^func (Test\w+)\(',s,re.M); rows.update(names);callers[str(f.relative_to(root))]=names
rows=sorted(rows); pattern='^('+ '|'.join(rows)+')$';(p/'matrix.regex').write_text(pattern+'\n');(p/'matrix-rows.json').write_text(json.dumps({'rows':rows,'callers':callers,'added':added,'unknown':sorted(set(current)-set(rows))},indent=2)+'\n')
targets=['TestNbodyIndexedElementsBorrow','TestCallTargetsElementBorrowPlan','TestDevirtualizeBorrowDocClaim']; profiles={}
for n in targets:
 c={}
 for l in (p/(n+'.cover')).read_text().splitlines()[1:]:
  b,s,count=l.split();c[b]=int(count)
 profiles[n]=c
exclusive=[]
for a,b in [(targets[0],targets[2]),(targets[1],targets[2]),(targets[2],targets[1])]:exclusive.append({'test':a,'subsumer':b,'exclusive':[k for k,v in profiles[a].items() if v>0 and profiles[b].get(k,0)==0]})
(p/'exclusive-coverage.json').write_text(json.dumps(exclusive,indent=2)+'\n')
(p/'CODE-AND-ORACLE.md').write_text('CODE UNDER TEST: native planElementBorrows, borrowable, changingFunctions, changes, unchanging and borrowElement C emission. ORACLE: self-written planner membership, five-declaration count, generated ownership and release assertions. No external authority or Node execution in these three rows. Node is used by some replay neighbors. No test, fixture or harness changes.\n')
d1old='e.line("%s %s = NULL;", cType(local.Type), owner)'; d1new='e.line("%s %s = adamic_retain(%s);", cType(local.Type), owner, name)'; assert d1old in original
D1=original.replace(d1old,d1new,1)
d2old='if targets.Unknown {\n\t\t\treturn false\n\t\t}'; d2new='if targets.Unknown {\n\t\t\treturn true\n\t\t}'; assert d2old in original
D2=original.replace(d2old,d2new,1)
d3old='case ir.Call:\n\t\tfor _, target := range program.CallTargets(expression) {'; d3new='case ir.Call:\n\t\tif expression.Virtual != 0 {\n\t\t\treturn false\n\t\t}\n\t\tfor _, target := range program.CallTargets(expression) {';assert d3old in original
D3=original.replace(d3old,d3new,1)
def line_of(s):return original[:original.index(s)].count('\n')+1
plan=[{'id':'D1','test':targets[0],'file_line':'internal/native/element_borrow.go:'+str(line_of(d1old)),'menu':'change a constant/ownership option','change':'Initialize the fallback owner by retaining the existing indexed element instead of NULL. Change existing emitted declaration format and supply its pointer operand; no extra emitted statement.','semantic_lead':'Extra retain/release preserves answers but loses the indexed borrow ownership cost property. Only nbody inspects emitted indexed ownership among the three rows.'},{'id':'D2','test':targets[1],'file_line':'internal/native/element_borrow.go:'+str(line_of(d2old)+1),'menu':'change a constant','change':'Unknown closure return false becomes true.','semantic_lead':'Unknown mutable callback can replace the element; doc claim has bounded harmless callbacks.'},{'id':'D3','test':targets[2],'file_line':'internal/native/element_borrow.go:'+str(line_of(d3old)),'menu':'return early','change':'Reject virtual calls at entry to the Call case before consulting all proven targets.','semantic_lead':'All virtual alternatives in the doc fixture only read length; the call-target negative fixture has a writing override. Conservative fallback preserves runtime answers while denying the positive borrow.'}]
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env.pop('ADAMIC_NATIVE_SPLIT',None);env.pop('ADAMIC_NATIVE_JOBS',None);results=[]
def run(label,cmd,cache):
 e=env.copy(); e['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-elements/cache/'+cache;started=time.monotonic()
 with (p/(label+'.log')).open('w') as log:r=subprocess.run(cmd,env=e,stdout=log,stderr=subprocess.STDOUT)
 events=[]
 for line in (p/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except ValueError:pass
 item={'label':label,'command':cmd,'cache':e['ADAMIC_BUILD_CACHE_DIR'],'wall_seconds':time.monotonic()-started,'exit':r.returncode,'failed':[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test')],'passed':[x['Test'] for x in events if x.get('Action')=='pass' and x.get('Test')],'skipped':[x['Test'] for x in events if x.get('Action')=='skip' and x.get('Test')],'timeout':any('panic: test timed out' in x.get('Output','') for x in events),'package_elapsed':[x.get('Elapsed') for x in events if x.get('Action') in ('pass','fail') and not x.get('Test')]}
 results.append(item);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(label,r.returncode,round(item['wall_seconds'],3),'fails',item['failed'],flush=True);return item
base=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run']
clean=run('bounded-baseline',base+[pattern],'clean');assert clean['exit']==0,'No mutants on a red or incomplete baseline'
try:
 for id,changed in [('D1',D1),('D2',D2),('D3',D3)]:
  source.write_text(changed);(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/native/element_borrow.go',tofile='b/internal/native/element_borrow.go')))
  vet=run(id+'-vet',['timeout','120','go','vet','./internal/native/'],id);assert vet['exit']==0
  matrix=run(id,base+[pattern],id)
  assert not matrix['timeout'] and matrix['exit'] in (0,1),'Mutant requires narrower replay'
finally:source.write_text(original)
run('restored',base+['^('+ '|'.join(targets)+')$'],'restored')
