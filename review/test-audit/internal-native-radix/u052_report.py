import pathlib,json,re,statistics,gzip
p=pathlib.Path('review/test-audit/internal-native-radix');scope=json.loads((p/'scope.json').read_text());members={r:ms for r,ms in scope};mapping={n:r for r,ms in scope for n in ms};rows=list(members)
def events(id):
 ret=[]
 log=p/(id+'.log')
 for line in (log.read_text() if log.exists() else gzip.open(str(log)+'.gz','rt').read()).splitlines():
  try:ret.append(json.loads(line))
  except:pass
 return ret
mutants=[f'M{i:02}' for i in range(1,13)];probes=[f'P{i:02}' for i in range(1,9)];witnesses=['TestRecordReadMutants','TestRecordMutants family']
fails={};passes={};evidence={};status={}
for id in mutants+probes+['witness']:
 ev=events(id);f={mapping[e['Test'].split('/')[0]] for e in ev if e.get('Action')=='fail' and e.get('Test','').split('/')[0] in mapping};fails[id]=f
 passes[id]={mapping[e['Test']] for e in ev if e.get('Action')=='pass' and e.get('Test') in mapping}
 status[id]=bool(ev and any(e.get('Action') in ['pass','fail'] and 'Test' not in e for e in ev))
 for row in f:
  outs=[e.get('Output','').strip() for e in ev if mapping.get(e.get('Test','').split('/')[0])==row and e.get('Action')=='output']
  typed=[e.get('Output','').strip() for e in ev if mapping.get(e.get('Test','').split('/')[0])==row and e.get('OutputType')=='error']
  errors=typed or [s for s in outs if re.search(r'\w+_test\.go:\d+:',s) and not any(t in s for t in ['compiling ','caught by ','size=','native=','answers (','round='])]
  evidence[id,row]=(errors[0].splitlines()[0][:350] if errors else next((s for s in outs if '--- FAIL:' in s),'failure'))
seconds={r:statistics.median([x['seconds'] for x in json.loads((p/('timing-'+r.replace(' ','-')+'.json')).read_text())]) for r in rows}
killset={r:{m for m in mutants if r in fails[m]} if r not in witnesses else set() for r in rows}
oracles={
'TestToStringWithARadixMatchesNode':('Live Node Number.toString, exact text for every input.','external-run'),
'TestToStringWithARadixOutOfRangePanics':('Live Node output and exception classification; native exit70 and exact diagnostic prefix.','external-run'),
'TestRecordsAgainstNode':('Live Node fixture stdout and prototype list; self runtime own-key stop policy and balanced allocation counts.',['external-run','self']),
'TestRecordReadMutants':('Witness of exact stop contract and own-hit comparison, plus sanitizer exclusion.','self'),
'TestRecordBenchmark':('Live Node work checksum; self five timing-field validity checks, no performance threshold.','external-run'),
'TestRecordMutants family':('Witness of Node comparison and sanitizer detection.',['external-run','self']),
'TestRegExpSearchNode':('Live Node captures, groups, indices and lastIndex.','external-run'),
'TestRegExpLintPatternsNode':('Live Node captures, groups, indices and lastIndex on lint patterns.','external-run'),
'TestRegExpBytecodeTest262':('Recorded test262 RegExp observations in matches.json.gz; first digit-class result checked against live Node in authority-check.json.','external-authority'),
'TestRegExpNativeStepLimit':('Self exit70 and instruction-step-limit diagnostic; does not verify exact consumed instruction count, M12 survives.','self'),
'TestRegExpIteratorResultShape':('Self optional done-field expectations; checks native process exit status.','self'),
'TestRegExpBytecodePatternUnits':('Self two lone-surrogate capture spans [0,1].','self'),
'TestRegExpBytecodeRandomNode family':('Live Node captures, groups, indices and lastIndex; forty wrappers share runRegexCases.','external-run'),
'TestRuntimeReleasePaths':('Self live allocation counts 1 then0, fixed stdout; ASan and UBSan.','self'),
'TestRuntimeStringEquality':('Live Node strict string equality including undefined, exact six booleans.','external-run')}
probe_map={'TestToStringWithARadixMatchesNode':['P01'],'TestToStringWithARadixOutOfRangePanics':['P01'],'TestRecordsAgainstNode':['P07','P08'],'TestRecordBenchmark':['P07','P08'],'TestRegExpIteratorResultShape':['P04'],'TestRuntimeReleasePaths':['P06'],'TestRuntimeStringEquality':['P05','P06']}
for r in rows:
 if r.startswith('TestRegExp') and r!='TestRegExpIteratorResultShape':probe_map[r]=['P02'] if r=='TestRegExpNativeStepLimit' else ['P02','P03']
report=[]
for row in rows:
 kills=sorted(killset[row]);unique=[m for m in kills if len(fails[m]-set(witnesses))==1];subs=[];subs_seconds=None
 if row in witnesses:verdict='witness' if row in fails['witness'] else 'untrue';lastid='W01' if row=='TestRecordReadMutants' else 'W02';last=evidence.get(('witness',row));command=json.loads((p/'witness-run.json').read_text())['command']
 else:
  if unique:verdict='slow-worthy' if seconds[row]>60 else 'sacred'
  elif not kills:verdict='untrue'
  else:
   supers=[other for other in rows if other!=row and killset[row]<=killset[other]]
   if supers:verdict='subsumed';subs=[min(supers,key=lambda r:seconds[r])];subs_seconds=seconds[subs[0]]
   else:verdict='overlapping';subs=sorted({other for other in rows if other!=row and killset[row]&killset[other]})
  lastid=kills[-1] if kills else None;last=evidence.get((lastid,row));command=json.loads((p/(lastid+'-run.json')).read_text())['command'] if lastid else None
 own=probe_map.get(row,[]);pk=[q for q in own if row in fails[q]]
 # Runtime rows have multiple APIs. Overall value is limited to listed entry probes, not untested APIs.
 vacuous=False if pk else (True if own and all(row in passes[q] for q in own) else None)
 file=next('internal/native/'+f+'_test.go' for f in ['radix','record','regexp_search','regexp','runtime_profile'] if any(n in pathlib.Path('internal/native/'+f+'_test.go').read_text() for n in members[row]))
 oracle,kind=oracles[row]
 obj={'test':row,'package':'internal/native','file':file,'seconds':seconds[row],'oracle':oracle,'oracle_kind':kind,'kills':kills,'unique_kills':unique,'last_proven_fail':(lastid+': '+last) if last else None,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':12 if row not in witnesses else 0,'probe_kills':pk,'subsumer_seconds':subs_seconds,'vacuous':vacuous,'bounded':True,'matrix_rows':rows,'evidence':(command+' => '+last) if last else 'No observed production kill; inspect per-selector logs.','members':members[row],'entry_probes':{q:('fail' if row in fails[q] else 'pass' if row in passes[q] else 'unknown') for q in own}}
 if verdict=='subsumed':obj['subsumption_mutants']=len(kills)
 if row=='TestRecordsAgainstNode':
  obj['vacuous_subcases']={q:[e['Test'].split('/',1)[1] for e in events(q) if e.get('Action')=='pass' and e.get('Test','').startswith('TestRecordsAgainstNode/') and e['Test'].count('/')==1] for q in own}

 report.append(obj)
(p/'rows.json').write_text(json.dumps(report,indent=2))
(p/'matrix.json').write_text(json.dumps({m:{'failed_rows':sorted(fails[m]),'passed_rows':sorted(passes[m]),'terminal':status[m]} for m in mutants+probes},indent=2))
for r in report:print(r['test'],r['seconds'],r['verdict'],r['kills'],r['probe_kills'])
