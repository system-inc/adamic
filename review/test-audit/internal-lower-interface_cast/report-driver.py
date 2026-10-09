import pathlib,json,re,statistics,subprocess,os,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-lower-interface_cast';rows=json.loads((p/'rows.json').read_text());menu=json.loads((p/'menu.json').read_text());ids=[x['id'] for x in menu if x['id'].startswith('M')]
pkgrows=[s for s in (p/'list.log').read_text().splitlines() if re.match(r'^Test\w+$',s)]
def events(file):
 es=[]
 for line in file.read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 return es
matrix={};fails={};fail_lines={};bounded=False
for rec in menu:
 mid=rec['id'];es=events(p/(mid+'.log'));states={}
 for e in es:
  row=e.get('Test','').split('/')[0]
  if row and e.get('Action') in ['pass','fail','skip']:states[row]=e['Action']
  if row and e.get('Action')=='output' and re.search(r'_test.go:\d+:',e.get('Output','')):fail_lines.setdefault((mid,row),e['Output'].strip())
 panic=any('panic:' in e.get('Output','') for e in es)
 if panic:
  bounded=True
  for row in rows:
   f=p/(mid+'.'+row+'.log')
   if not f.exists():states[row]='unknown';continue
   ies=events(f);outcomes=[e['Action'] for e in ies if e.get('Test')==row and e.get('Action') in ['pass','fail','skip']]
   states[row]=outcomes[-1] if outcomes else ('timeout' if any('panic: test timed out' in e.get('Output','') for e in ies) else 'panic')
   for e in ies:
    if re.search(r'_test.go:\d+:',e.get('Output','')):fail_lines[(mid,row)]=e['Output'].strip();break
   if states[row] in ['panic','timeout']:fail_lines[(mid,row)]=next((e['Output'].strip() for e in ies if 'panic:' in e.get('Output','')),'binary aborted')
 matrix[mid]={row:states.get(row,'unknown') for row in (pkgrows if mid.startswith('M') else rows)}
 fails[mid]=[r for r,state in matrix[mid].items() if state in ['fail','panic','timeout']]
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'matrix-rows.json').write_text(json.dumps(pkgrows,indent=2)+'\n');(p/'package-kills.json').write_text(json.dumps(fails,indent=2)+'\n')
seconds={};trials={}
for row in rows:
 vals=[]
 for i in range(1,4):
  s=(p/(row+f'.time{i}.log')).read_text();m=re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',s);assert m,s;vals.append(float(m.group(1)))
 seconds[row]=statistics.median(vals);trials[row]=vals
# Pick a common scoped subsumer where available. Otherwise use an observed package
# row, preferring a cheap clean-baseline elapsed value, then measure it alone.
base_elapsed={e['Test']:e.get('Elapsed',999) for e in events(p/'baseline.log') if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']}
choices={};needed=[]
for row in rows:
 kills=[mid for mid in ids if row in fails[mid]]
 unique=[mid for mid in kills if fails[mid]==[row] and not bounded]
 if unique or not kills:continue
 common=[r for r in pkgrows if r!=row and all(r in fails[mid] for mid in kills)]
 if common:
  known=[r for r in common if r in seconds]
  chosen=min(known,key=lambda r:seconds[r]) if known else min(common,key=lambda r:base_elapsed.get(r,999))
  choices[row]=chosen
  if chosen not in seconds and chosen not in needed:needed.append(chosen)
if os.environ.get('U033_MEASURE_SUBSUMERS')=='1':
 for row in needed:
  vals=[]
  for i in range(1,4):
   f=p/(row+f'.time{i}.log');env=os.environ.copy();env['ADAMIC_MUTANT']=''
   with f.open('w') as log:r=subprocess.run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
   assert not r.returncode,(row,r.returncode)
   m=re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',f.read_text());assert m,row;vals.append(float(m.group(1)))
  seconds[row]=statistics.median(vals);trials[row]=vals
else:
 for row in needed:
  vals=[]
  for i in range(1,4):
   f=p/(row+f'.time{i}.log')
   if f.exists():
    m=re.search(r'\bok\s+\S+\s+(\d+\.\d+)s',f.read_text());vals.append(float(m.group(1)))
  if len(vals)==3:seconds[row]=statistics.median(vals);trials[row]=vals
(p/'subsumer-candidates.json').write_text(json.dumps({'chosen':choices,'additional_rows_to_time':needed},indent=2)+'\n')
checks={
rows[0]:'Self-written admission for 12 sources and three diagnostic-substring refusals; accepted sources have no IR or runtime assertion.',
rows[1]:'Self-written nil-error admission; no IR, output, or environment-switch assertion.',
rows[2]:'Self-written NotYet error category only; accepts unrelated NotYet reasons, as M03 demonstrates.',
rows[3]:'Self-written NotYet plus hide-a-return substring.',
rows[4]:'Self-written NotYet plus receiver-convention substring.',
rows[5]:'Self-written Refused plus default substring in Fix.',
rows[6]:'Self-written Refused plus suspended-frames substring in Fix, for three generator syntaxes.',
rows[7]:'Self-written Refused plus cycle substring in What.',
rows[8]:'Self-written nil-error admission for two method views; no receiver IR or executed-output assertion.',
rows[9]:'Self-written NotYet plus seven descriptor-reason substrings.',
rows[10]:'Self-written NotYet plus represented-method-replacement substring from direct setProperty call.',
rows[11]:'Self-written acceptance of either Refused or NotYet for five inputs; reason and fix are unchecked.',
rows[12]:'Self-written Refused plus nominal-ancestry substring.',
rows[13]:'Self-written NotYet plus index-representation substring.',
rows[14]:'Self-written NotYet plus symbol-key-storage substring.'}
files={}
for name in ['interface_cast_test.go','iteration_test.go']:
 for n,line in enumerate((root/'internal/lower'/name).read_text().splitlines(),1):
  m=re.match(r'func (Test\w+)\(',line)
  if m:files[m.group(1)]='internal/lower/'+name+':'+str(n)
positive=['complete','missing payload','wrong payload','broad tag','alias write','unused bad factory','different tag','dynamic complete','optional target','spread','generic','staged class']
audit=[]
for row in rows:
 kills=[mid for mid in ids if row in fails[mid]];unique=[mid for mid in kills if fails[mid]==[row] and not bounded]
 subs=[]
 if unique:verdict='slow-worthy' if seconds[row]>60 else 'sacred'
 elif not kills:verdict='untrue'
 elif row in choices:verdict='subsumed';subs=[choices[row]]
 else:
  verdict='overlapping'
  for mid in kills:
   candidates=[r for r in fails[mid] if r!=row]
   if candidates:
    chosen=min(candidates,key=lambda r:seconds.get(r,base_elapsed.get(r,999)))
    if chosen not in subs:subs.append(chosen)
 chosen=(unique or kills)[-1] if kills else None;probe='P02' if row==rows[10] else 'P01';state=matrix[probe].get(row,'unknown')
 command=f'ADAMIC_MUTANT={chosen} ADAMIC_BUILD_CACHE_DIR=/tmp/u033/cache/{chosen} timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run . > {chosen}.log 2>&1'
 obj=dict(test=row,package='internal/lower',file=files[row],seconds=seconds[row],oracle=checks[row],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=chosen+' '+fail_lines.get((chosen,row),'observed failure') if chosen else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=12,probe_kills=[probe] if row in fails[probe] else [],subsumer_seconds=seconds.get(subs[0]) if verdict=='subsumed' else None,vacuous=state=='pass' if state in ['pass','fail','panic','timeout'] else None,bounded=bounded,evidence=command+'; '+fail_lines.get((chosen,row),'observed failure') if chosen else 'No production mutant killed this row.',vacuous_subcases=positive if row==rows[0] else [],subsumption_mutants=len(kills) if verdict=='subsumed' else None)
 if bounded:obj['matrix_rows']=rows
 audit.append(obj)
(p/'audit.json').write_text(json.dumps(audit,indent=2)+'\n');(p/'median-seconds.json').write_text(json.dumps({'medians':seconds,'trials':trials},indent=2)+'\n')
for obj in audit:print(obj['test'],obj['verdict'],obj['kills'],obj['unique_kills'],obj['subsumed_by'],obj['vacuous'])
print('additional subsumers',needed)
