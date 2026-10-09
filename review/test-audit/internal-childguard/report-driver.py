import pathlib,json,re,statistics
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-childguard')
rows=['TestChild','TestProgress','TestStalled','TestCeiling','TestExitIsNotGuardError','TestKillsProcessGroup','TestNoFirstOutput']
menu=json.loads((p/'menu.json').read_text()); matrix={}; evidence={}
for rec in menu:
 mid=rec['id']; matrix[mid]={}
 for row in rows:
  f=p/(mid+'.'+row+'.log' if mid=='M09' else mid+'.log')
  es=[]
  for l in f.read_text().splitlines():
   try: es.append(json.loads(l))
   except: pass
  states=[e['Action'] for e in es if e.get('Test')==row and e.get('Action') in ['pass','fail','skip']]
  timeout=any('panic: test timed out' in e.get('Output','') for e in es)
  result=states[-1] if states else ('timeout' if mid=='M09' and timeout else 'unknown')
  matrix[mid][row]=result
  lines=[e.get('Output','').strip() for e in es if e.get('Test')==row and re.search(r'childguard_test.go:\d+:',e.get('Output',''))]
  if result=='timeout': lines=['panic: test timed out after 1m30s']
  evidence[(mid,row)]=lines[0] if lines else None
# M13 is conservatively supplemental: error construction is broader than a literal option change.
for rec in menu:
 rec.pop('switch',None)
 rec['supplemental']=rec['id']=='M13'
 rec['failed_rows']=[r for r in rows if matrix[rec['id']][r] in ['fail','timeout']]
 rec['timed_out_rows']=[r for r in rows if matrix[rec['id']][r]=='timeout']
(p/'mutants.json').write_text(json.dumps(menu,indent=2)+'\n');(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
seconds={}
for row in rows:
 vals=[]
 for i in range(1,4):
  text=(p/(row+f'.time{i}.log')).read_text(); match=re.search(r'\bok\s+\S+\s+([\d.]+)s',text); assert match,text; vals.append(float(match.group(1)))
 seconds[row]=statistics.median(vals)
checks={
'TestProgress':'Self-written success and count of five progress newline occurrences; extra unrelated output is not rejected.',
'TestStalled':'Self-written stalled error fields and text, 200 ms to 1200 ms timing bound, and exact stderr once newline.',
'TestCeiling':'Self-written ceiling error reason and prefix, with upper timing bound only; no lower bound.',
'TestExitIsNotGuardError':'Self-written concrete exec.ExitError and exit status 7, excluding guard errors; exit code alone does not establish cause.',
'TestKillsProcessGroup':'Self-written stalled guard result and 2 s pipe-closure deadline. Descendant survival is inferred from inherited pipes, not independently checked by PID.',
'TestNoFirstOutput':'Self-written first-output flag, 300 ms window, diagnostic text, and 300 ms to 1300 ms timing bound.',
'TestChild':'Subprocess entry, activated only by CHILDGUARD_TEST; no independent oracle.'}
audit=[]
valid=[m['id'] for m in menu if m['id'].startswith('M') and not m['supplemental']]
for row in rows:
 kills=[m for m in valid if matrix[m][row] in ['fail','timeout']]
 unique=[m for m in kills if sum(matrix[m][r] in ['fail','timeout'] for r in rows)==1 and all(matrix[m][r]!='unknown' for r in rows)]
 chosen=(unique or kills)[-1] if kills else None
 subs=[]
 if row=='TestChild': verdict='helper'
 elif unique: verdict='sacred' if seconds[row]<=60 else 'slow-worthy'
 elif not kills: verdict='untrue'
 else:
  subs=[r for r in rows if r!=row and all(matrix[m][r] in ['fail','timeout'] for m in kills)]
  verdict='subsumed' if subs else 'overlapping'
  if subs: subs=[min(subs,key=lambda r:seconds[r])]
 command='ADAMIC_MUTANT='+str(chosen)+' timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run .'
 obj=dict(test=row,package='internal/childguard',file='internal/childguard/childguard_test.go',seconds=seconds[row],oracle=checks[row],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=(chosen+' '+str(evidence[(chosen,row)])) if chosen else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=18,probe_kills=['P01'] if matrix['P01'][row]=='fail' else [],subsumer_seconds=seconds[subs[0]] if subs else None,vacuous=(matrix['P01'][row]=='pass') if row!='TestChild' else None,bounded=False,matrix_rows=rows,evidence=(command+' > '+chosen+'.log 2>&1; '+str(evidence[(chosen,row)])) if chosen else 'go test -list . ./internal/childguard/; TestChild',supplemental_kills=['M13'] if matrix['M13'][row]=='fail' else [],vacuous_subcases=[])
 if row=='TestChild':obj['parents']=['TestProgress','TestStalled','TestCeiling','TestExitIsNotGuardError','TestNoFirstOutput']
 audit.append(obj)
(p/'audit.json').write_text(json.dumps(audit,indent=2)+'\n')
print(json.dumps(audit,indent=2))
