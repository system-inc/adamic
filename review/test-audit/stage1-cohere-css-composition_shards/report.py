import pathlib,json,re,statistics,subprocess
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-css-composition_shards');r=p.parents[2];groups=json.loads((p/'row-members.json').read_text());runs=json.loads((p/'audit-runs.json').read_text())+json.loads((p/'followup-runs.json').read_text());menu=json.loads((p/'menu.json').read_text());special=json.loads((p/'special-edits.json').read_text());probes=json.loads((p/'probes.json').read_text());maps=json.loads((p/'origin-line-maps.json').read_text());base=(p/'base.txt').read_text().strip()
def events(log):
 out=[]
 for line in (p/log).read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out

def failures(log):return [e['Test'] for e in events(log) if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']]
def line(log,members):
 failed={t.split('/')[0] for t in failures(log)}
 ev=events(log)
 preferred=[e for e in ev if e.get('OutputType')=='error' and e.get('Test','').split('/')[0] in failed and re.search(r'\w+_test.go:\d+:',e.get('Output',''))]
 for e in preferred+list(reversed(ev)):
  txt=e.get('Output','').strip();t=e.get('Test','').split('/')[0]
  if t in failed and re.search(r'\w+_test.go:\d+:',txt) and not re.search(r': (build |own work:|setup including products:|shard leaf:|union leaf:|only |=== RUN|.*byte-identical|.*elapsed )',txt):
   def cv(m):return m[1]+':'+str(maps.get(m[1],{}).get(m[2],m[2]))+':'
   return re.sub(r'(\w+_test.go):(\d+):',cv,txt)
 return 'No attributable assertion line; see log for timeout or build failure.'

oracles={
'TestCompositionMatchesGo setup family':('self','Constructed products must have nonempty identities; empty setup return bypasses this guard.'),
'TestCompositionMatchesGo family':('external-run','Independent Go cohere composition trees and errors, compared byte for byte.'),
'TestCSSPrinterAgreesWithGo family':('external-run','Independent Go cohere printer. Agreement leaves also compare Prettier and its Go fork. Mixed agreement and planted-mutant leaves share one checker; witness verdict rests only on weakened comparison leaves.'),
'TestCompositionMatchesGoUnion':(['external-run','self'],'Node port mutant comparisons plus own synthetic shard coverage and planted firstDifference checks.'),
'TestCSSPrinterAgreesWithGoUnion':('self','Handwritten shard count, mode coverage and partition union assertions.'),
'TestCSSThroughput':('external-run','PostCSS counts checked against native count/checksum, plus Node executes own source. Count agreement does not prove tree agreement.'),
'TestTheCanonicalRangeChecksCanFail':('self','Own planted IR range corruption must appear in rendered mismatch labels.'),
'TestEachGapStandsWhereGapsMdSaysItDoes':(['external-run','self'],'Node execution is checked against handwritten stdout; Adamic refusal text is handwritten and has no independent authority.'),
'TestClosedEmptyArrayUnionGap':(['external-run','self'],'Node, native and own JavaScript backend must print handwritten 0; sanitizer leak report must be empty.'),
'TestClosedParserRegexGap':(['external-run','self'],'Node, native and own JavaScript backend must print handwritten Parsed/Ok status labels. Status-only result is weaker than a tree comparison.'),
'TestClosedOptionalBooleanConditionGap':(['external-run','self'],'Node, native and own JavaScript backend must print handwritten important; sanitizer leak report must be empty.'),
'TestComposedMemoryChecksCanFail':(['external-run','self'],'Go printer bytes must survive ordinary runs; own planted C defects must trigger named ASan/UBSan/LSan reports.'),
'TestCSSPrinterOptimizedMatchesGo':('external-run','Independent Go cohere printer bytes, default and narrow modes, compared with optimized native artifact.'),
'TestCSSParserOptimizedMatchesNode':('external-run','Node executes the same port source. Port-source mutations change this oracle too, so no production kill is counted here.'),
'TestThePortParsesAsGoCohereDoesUnion':('self','Own shard assignments, fixture coverage and variant identities.'),
'TestCSSParserPlantedDisagreement':('self','Subprocess should report exactly the planted shard failure; its agreement guard must detect altered bytes.')}
rows=[];matrix=[]
for m in menu:
 for row in m['rows']:
  log=m['id']+'-'+row.replace(' ','_')+'.log'
  if not (p/log).exists():continue
  ev=events(log);term=[e['Test'] for e in ev if e.get('Action') in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']]
  matrix.append(dict(mutant=m['id'],row=row,log=log,failed=failures(log),observed=term,unknown=[t for t in groups[row] if t not in term],bounded=True))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
for row,members in groups.items():
 timing=[]
 for i in range(1,4):
  f='timing-'+row.replace(' ','_')+'-'+str(i)+'.log'
  if (p/f).exists():timing.extend([e['Elapsed'] for e in events(f) if e.get('Action')=='pass' and 'Test' not in e])
 kills=[m['id'] for m in menu if any(x['mutant']==m['id'] and x['row']==row and x['failed'] for x in matrix)]
 unique=[id for id in kills if len({x['row'] for x in matrix if x['mutant']==id and x['failed']})==1]
 sk=[s for s in special if row in s['rows']];pk=[];vacuous=None;proof=[]
 for x in probes:
  if row not in x['rows']:continue
  names=[x['id']+'-'+t+'.log' for t in members] if row=='TestCompositionMatchesGo setup family' else [x['id']+'-'+row.replace(' ','_')+'.log']
  if not all((p/n).exists() for n in names):continue
  if any(failures(n) for n in names):pk.append(x['id']);vacuous=False
  elif all(any(e.get('Action')=='pass' and 'Test' not in e for e in events(n)) for n in names):vacuous=True
 subs=[];verdict='cannot-judge';reason='No admissible production mutant of this compiler or dependency-port boundary was built within the four port-mutant menu.'
 if sk:
  s=sk[0];log=s['id']+'-'+row.replace(' ','_')+'.log'
  if (p/log).exists() and failures(log):verdict='setup-check' if s['id'].startswith('S') else 'witness';proof=[s['id'],log];reason='Weakened guard/construction failed in this session.'
  elif (p/log).exists():verdict='untrue';reason='Weakened guard run showed no attributable row failure; budget and unknown members are retained.'
 elif kills:
  if unique:verdict='sacred';reason='Unique only within the declared bounded matrix, not proven package unique.'
  else:
   candidates=[other for other in groups if other!=row and all(any(x['row']==other and x['mutant']==id and x['failed'] for x in matrix) for id in kills)]
   verdict='subsumed' if candidates else 'overlapping';subs=candidates[:1] if candidates else sorted({x['row'] for x in matrix if x['row']!=row and x['failed'] and x['mutant'] in kills});reason='Verdict rests on '+str(len(kills))+' caught production mutants; it is not deletion advice.'
  id=kills[-1];proof=[id,id+'-'+row.replace(' ','_')+'.log']
 elif row in [mrow for m in menu for mrow in m['rows']]:verdict='untrue';reason='No planted production mutant caused an attributable row failure.'
 ev='';last=None
 if row=='TestCSSPrinterAgreesWithGo family' and (p/'bounded-printer-063-W1.log').exists():proof=['W1','bounded-printer-063-W1.log']
 if proof:
  rr=next((x for x in runs if x['log']==proof[1]),{});fail=line(proof[1],members);last=proof[0]+': '+fail;ev='ADAMIC_MUTANT='+proof[0]+'; /tmp/u078-mutant='+proof[0]+'; '+' '.join(rr.get('command',[]))+'; '+fail+'; log='+proof[1]
 else:ev='Clean timing logs and entry probes only; no admissible production kill.'
 files=[]
 for f in ['composition_shards_test.go','css_printer_parallel_test.go','css_test.go','gaps_test.go','memory_checks_test.go','optimized_test.go','parser_shards_test.go']:
  content=subprocess.check_output(['git','show',base+':stage1/cohere/css/'+f],cwd='/workspace/adamic',text=True)
  for member in members:
   match=re.search(r'^func '+member+r'\(',content,re.M)
   if match:files.append('stage1/cohere/css/'+f+':'+str(content[:match.start()].count('\n')+1))
 kind,oracle=oracles[row]
 rows.append(dict(test=row,package='stage1/cohere/css',file=files[0] if files else None,files=files,members=members,seconds=statistics.median(timing) if len(timing)==3 else None,timing_samples=timing,oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=unique,last_proven_fail=last,verdict=verdict,subsumed_by=subs,mutants_in_matrix=[x['mutant'] for x in matrix if x['row']==row],probe_kills=pk,subsumer_seconds=None,vacuous=vacuous,bounded=True,matrix_rows=sorted({x['row'] for x in matrix}),evidence=ev,limitations=reason))
for row in rows:
 if row['verdict']=='subsumed':row['subsumer_seconds']=next(x['seconds'] for x in rows if x['test']==row['subsumed_by'][0])
(p/'rows.json').write_text(json.dumps(rows,indent=2));print([(x['test'],x['verdict'],x['kills'],x['probe_kills'],x['vacuous']) for x in rows])
