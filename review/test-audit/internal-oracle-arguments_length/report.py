import pathlib,json,re,collections
p=pathlib.Path('review/test-audit/internal-oracle-arguments_length');rows=json.loads((p/'rows.json').read_text());matrix=json.loads((p/'matrix.json').read_text());checks=json.loads((p/'check-results.json').read_text());probes=json.loads((p/'probe-results.json').read_text());names=[r['test'] for r in rows];family='TestCallTarget family';members=names[7:9];groups=[[n] for n in names[:7]]+[members]+[[n] for n in names[9:]];groupnames=names[:7]+[family]+names[9:];timings={r['test']:r['seconds'] for r in rows};timings[family]=json.loads((p/'family-times.json').read_text())['seconds'];witnesses={names[1],names[13]};setup=set(names[2:7]);normal=set(names)-witnesses-setup
oracles=[('Original .ts source on Node decides stdout, stderr, and exit; native, release and JavaScript are compared; sanitizer/leak checks also run.','external-run'),('Original Node output versus a deliberately wrong native slot, plus self-written clean-run and stdout-diff expectations.',['external-run','self']),('Self-written expected puts output proves changed generated C invalidates cached evidence.','self'),('Self-written old/new values for mocked runtime and Node identities.','self'),('Self-written key inequality for source/JavaScript bytes and an edited imported module.','self'),('Self-written byte, exit, leak, exactly-once, corruption, and replay-comparison assertions.','self'),('Self-written fresh-versus-cached strings and preservation of old evidence after bypass.','self'),('Original source on Node compared byte-for-byte with both backends, plus self-written Node stdout preconditions.',['external-run','self']),('Self-written exit 70, exact stdout flush and panic text. Node executes generated JavaScript, rather than supplying an independent source answer.','self'),('Self-written Refused type and no-unchecked-cast label. Node is diagnostic only after erroneous admission; the test fails unconditionally then.','self'),('Self-written exit 70/stdout and absence of a handler-created marker. Node executes generated JavaScript.','self'),('Self-written exit 70, stdout and panic messages, plus original Node source must finish with exit zero.',['self','external-run']),('Built-in native mutants compared with original Node for successful casts and Adamic JavaScript for checked failures.',['external-run','self'])]
def grouped_kills(m):return sorted({family if n in members else n for n in m['kills'] if n in normal})
def failing(m,member):
 es=[e.get('Output','').strip() for e in m['events'] if e.get('Test','').split('/')[0]==member and e.get('Action')=='output' and 'gate cache' not in e.get('Output','') and re.search(r'_test.go:\d+:|^panic:',e.get('Output',''))];return es[-1] if es else 'see complete log'
report=[]
for index,group in enumerate(groups):
 name=groupnames[index];r=next(r for r in rows if r['test']==group[0]);kills=[m['id'] for m in matrix if name in grouped_kills(m)];unique=[m['id'] for m in matrix if grouped_kills(m)==[name]];sub=[];ss=None;last=None
 if name in witnesses:
  verdict='witness';last=next(m for m in checks if m['id']=='W01');checkids=['W01']
 elif name in setup:
  verdict='setup-check';matches=[m for m in checks if m['id'].startswith('S') and name in m['kills']];last=matches[-1];checkids=[m['id'] for m in matches]
 elif unique:verdict='sacred'
 elif kills:
  common=set.intersection(*(set(grouped_kills(m))-{name} for m in matrix if name in grouped_kills(m)))
  if common:verdict='subsumed';sub=[min(common,key=lambda n:(timings[n],n))];ss=timings[sub[0]]
  else:verdict='overlapping';sub=sorted(set.union(*(set(grouped_kills(m))-{name} for m in matrix if name in grouped_kills(m))))
 else:verdict='untrue'
 if name not in setup|witnesses:last=next((m for m in reversed(matrix) if name in grouped_kills(m)),None);checkids=[]
 failmember=next((n for n in group if last and n in last['kills']),group[0]);line=failing(last,failmember) if last else None
 own=[q for q in probes if q['test'] in group];pk=sorted(set(q['id'] for q in own if q['exit']));entry={q['id']:all(x['exit']==0 for x in own if x['id']==q['id']) for q in own};vacuous=all(entry.values()) if own else None
 result=dict(test=name,package='internal/oracle',file=r['file']+':'+str(r['line']),seconds=timings[name],oracle=oracles[index][0],oracle_kind=oracles[index][1],kills=kills,unique_kills=unique,last_proven_fail=last['id']+': '+line if last else None,verdict=verdict,subsumed_by=sub,mutants_in_matrix=4,probe_kills=pk,subsumer_seconds=ss,vacuous=vacuous,bounded=True,matrix_rows=groupnames,evidence=last['command']+'; '+line if last else 'All four bounded production runs completed; no failure in this row.',entry_probe_passes=entry)
 if len(group)>1:result['members']=group;result['member_files']=[q['file']+':'+str(q['line']) for q in rows if q['test'] in group]
 if name in setup:result['setup_kills']=checkids;result['setup_matrix_mutants']=5
 if name in witnesses:result['witness_kills']=checkids;result['ignored_production_failures']=[m['id'] for m in matrix if name in m['kills']]
 if verdict=='subsumed':result['subsumption_basis']=str(len(kills))+' production mutants; hint, not deletion'
 result['unique_kill_scope']='bounded selected rows only' if unique else None
 report.append(result)
(p/'report.json').write_text(json.dumps(report,indent=2));print(collections.Counter(q['verdict'] for q in report))
for r in report:print(r['test'],r['verdict'],r['kills'],r['last_proven_fail'])
