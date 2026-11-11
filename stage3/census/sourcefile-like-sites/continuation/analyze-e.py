import pathlib,json,gzip,re,collections,hashlib
out=pathlib.Path(__file__).resolve().parents[1];p=out/'continuation'
base_root='/workspace/scratch/baobab-adapted';variant_root='/workspace/scratch/baobab-road2-e'
def load(file):
 if not file.exists():return json.loads(gzip.decompress(pathlib.Path(str(file)+'.gz').read_bytes()))
 return json.loads(file.read_text())
def stream(file):
 if not file.exists():file=pathlib.Path(str(file)+'.gz')
 if not file.exists():return []
 with (gzip.open(file,'rt') if file.suffix=='.gz' else file.open()) as src:
  result=[]
  for line in src:
   try:result.append(json.loads(line))
   except ValueError:pass
  return result
def back_line(file,line):
 if file.endswith('compiler/types.ts') and line>=4292:return line-6
 if file.endswith('compiler/scanner.ts') and line>=551:return line-3
 return line
def norm(value,variant):
 if isinstance(value,str):
  value=value.replace(variant_root if variant else base_root,'<tree>')
  if variant:
   value=re.sub(r'(<tree>/src/compiler/(?:types|scanner)\.ts:)(\d+)(:)',lambda m:m[1]+str(back_line(m[1][:-1],int(m[2])))+m[3],value)
  return value
 if isinstance(value,list):return [norm(v,variant) for v in value]
 if isinstance(value,dict):return {k:norm(v,variant) for k,v in value.items()}
 return value
def delta(old,new):
 a={json.dumps(r,sort_keys=True):r for r in old};b={json.dumps(r,sort_keys=True):r for r in new}
 return {'baseline_unique':len(a),'variant_unique':len(b),'added':[b[k] for k in sorted(b.keys()-a.keys())],'removed':[a[k] for k in sorted(a.keys()-b.keys())]}
def finding_key(f,variant,root=None):
 r=norm({k:f.get(k) for k in ['unit','phase','kind','where','site_where','reason','text']},variant)
 if root is not None:r['root']=root
 return r
def local(where):return where.split('/src/compiler/',1)[-1]
def within(where,site):
 m=re.fullmatch(r'(.+):(\d+):(\d+)',local(where))
 if not m or m[1]!=site['file']:return False
 span=site.get('_comparison_span',site['call_span']);pos=(int(m[2])-1,int(m[3])-1)
 return (span['start']['line'],span['start']['character'])<=pos<(span['end']['line'],span['end']['character'])
names=['scanner.ts']+load(p/'variant-files.json');coverage={r['file']:r for r in load(out/'coverage.json')}
old_all=[];new_all=[];all_findings=[];units=[];statuses=[];headers=[];checker_sites=[]
for name in names:
 for r in stream(out/coverage[name]['stream']):old_all.extend(finding_key(f,False,name) for f in r.get('findings') or [])
 stem=name.replace('/','__').removesuffix('.ts');s=load(p/'e/raw'/(stem+'.status.json'));rows=stream(p/'e/raw'/(stem+'.jsonl'))
 s['complete_stream']=s['child_exit']==0 and bool(rows) and rows[-1].get('status')=='relation_report';statuses.append(s)
 if rows:headers.append({'file':name,'checker_rejected':rows[0].get('checker_rejected'),'diagnostics':rows[0].get('diagnostics')})
 for r in rows:
  all_findings.extend(r.get('findings') or []);units.extend(r.get('units') or []);new_all.extend(finding_key(f,True,name) for f in r.get('findings') or [])
  for d in r.get('diagnostic_sites') or []:
   source=pathlib.Path(d['file']);prefix=source.read_bytes()[:d['start']]
   checker_sites.append(dict(d,where=str(source)+':'+str(prefix.count(b'\n')+1)+':'+str(len(prefix)-prefix.rfind(b'\n'))))
raw_sites=load(out/'sites.json');grouped={}
for s in raw_sites:grouped.setdefault(s['location'],dict(s,baseline_findings=[]))['baseline_findings'].append(s)
spans=collections.Counter((s['file'],json.dumps(s['call_span'],sort_keys=True)) for s in grouped.values())
for s in grouped.values():
 if spans[(s['file'],json.dumps(s['call_span'],sort_keys=True))]>1:s['_comparison_span']=s['argument_span']
status_by={s['file']:s for s in statuses};unique={json.dumps(f,sort_keys=True):f for f in all_findings};all_findings=list(unique.values())
diagnostics=load(p/'e/whole-src-diagnostics.json');comparison=[]
for location,s in grouped.items():
 findings=[f for f in all_findings if within(norm(f.get('where',''),True),s) or within(norm(f.get('site_where',''),True),s)]
 checker=[d for d in checker_sites if within(norm(d['where'],True),s)]
 stock=[d for d in diagnostics if within(d.get('file','').removeprefix('src/compiler/')+':'+str(d.get('line',0))+':'+str(d.get('column',0)),s)]
 own=[u for u in units if local(norm(u.get('where',''),True))==local(s['unit'])]
 complete=all(status_by[n]['complete_stream'] for n in s.get('observed_in') or [s['file']]);attempted=any(u.get('status') in ['attempted','instantiated'] and not u.get('checker_diagnostics') for u in own)
 old={(f['kind'],f['reason']) for f in s['baseline_findings']}
 verdict='incomplete' if not complete else 'checker diagnostic' if checker or stock else ('replacement finding' if any((f['kind'],f['reason']) not in old for f in findings) else 'unchanged finding') if findings else 'cleared at original view' if attempted else 'unit excluded or not attempted'
 comparison.append({'location':location,'callee':s['callee'],'enclosing_function':s['enclosing_function'],'baseline_findings':s['baseline_findings'],'verdict':verdict,'complete_stream':complete,'findings':findings,'checker_diagnostics':checker,'stock_diagnostics':stock,'unit_statuses':own})
(p/'e-site-comparison.json').write_text(json.dumps(comparison,indent=2)+'\n')
def coherent(rows,variant):
 result=[]
 for r in rows:
  if r.get('kind')!='finding':continue
  d=norm({k:r.get(k) for k in ['rule','path','line','column','severity','message']},variant)
  if variant:d['line']=back_line(d['path'],d['line'])
  result.append(d)
 return result
cohere_rows=stream(p/'e/cohere-rules.jsonl');cohere_delta=delta(coherent(stream(p/'original-rules/cohere-rules.jsonl'),False),coherent(cohere_rows,True))
def diag(rows,variant):
 result=[]
 for r in rows:
  d=norm({k:r.get(k) for k in ['file','line','column','code','category','message']},variant)
  if variant and d['file'] and d['line']:d['line']=back_line(d['file'],d['line'])
  result.append(d)
 return result
native=load(p/'e/whole-src-adamic-checker.json');native_old=load(p/'original/whole-src-adamic-checker.json')
exact_delta=delta(diag(load(p/'original-exact/whole-src-diagnostics.json'),False),diag(load(p/'e-exact/whole-src-diagnostics.json'),True))
native_delta=delta(norm(native_old['diagnostics'],False),norm(native['diagnostics'],True))
stock_calls=load(p/'e/calls.json');native_calls=load(p/'e/native-overload-resolution.json')['calls'];routing=[]
for site in sorted(grouped.values(),key=lambda s:s['location']):
 if site['callee']!='getLineAndCharacterOfPosition':continue
 native_matches=[c for c in native_calls if local(c['argument'])==site['location']]
 stock_matches=[c for c in stock_calls if c['callee']=='getLineAndCharacterOfPosition' and c['file']=='src/compiler/'+site['file'] and c['argument_line']==int(site['location'].rsplit(':',2)[1]) and c['argument_column']==int(site['location'].rsplit(':',2)[2])]
 assert len(native_matches)==len(stock_matches)==1,(site['location'],native_matches,stock_matches)
 routing.append({'location':site['location'],'enclosing_function':site['enclosing_function'],'native':native_matches[0],'stock':stock_matches[0]})
scanner=stream(p/'e/raw/scanner.jsonl');char_units=[u for r in scanner for u in r.get('units') or [] if u.get('name')=='getLineAndCharacterOfPosition']
char_locations={u['where'] for u in char_units}
char_findings=[f for r in scanner for f in r.get('findings') or [] if f.get('unit') in char_locations or 'getLineAndCharacterOfPosition' in f.get('reason','')]
scanner_overloads=[f for r in scanner for f in r.get('findings') or [] if 'overload' in f.get('reason','').lower()]
starts_units=[u for r in scanner for u in r.get('units') or [] if u.get('name')=='getLineStarts']
starts_findings=[f for r in scanner for f in r.get('findings') or [] if f.get('unit','').endswith('/scanner.ts:503:1')]
lower_delta=delta(old_all,new_all)
sdk_rows=stream(p/'sdk-e/raw/consumer.jsonl');sdk_old=stream(p/'sdk-original/raw/consumer.jsonl')
sdk_key=lambda f,new:{k:(v.replace('/workspace/scratch/baobab-sdk-e' if new else '/workspace/scratch/baobab-sdk-original','<sdk>') if isinstance(v,str) else v) for k,v in f.items() if k not in ['start','end']}
sdk_findings=[f for r in sdk_rows for f in r.get('findings') or []]
sdk_delta=delta([sdk_key(f,False) for r in sdk_old for f in r.get('findings') or []],[sdk_key(f,True) for f in sdk_findings])
result={'variant':'e','files':statuses,'headers':headers,'site_counts':dict(collections.Counter(r['verdict'] for r in comparison)),'cohere_summary':next(r for r in cohere_rows if r.get('kind')=='summary'),'cohere_delta':cohere_delta,'lowering_delta':lower_delta,'lowering_delta_by_phase':{phase:{key:sum(f['phase']==phase for f in lower_delta[key]) for key in ['added','removed']} for phase in sorted({f['phase'] for f in old_all+new_all})},'stock':load(p/'e/summary.json'),'stock_exact':load(p/'e-exact/summary.json'),'stock_exact_delta':exact_delta,'native_checker_errors':len(native['diagnostics']),'native_checker_delta':native_delta,'first_stop':stream(p/'e/core-first-stop.jsonl')[0],'routing':routing,'scanner_getLineAndCharacterOfPosition':{'units':char_units,'findings':char_findings},'scanner_getLineStarts':{'units':starts_units,'findings':starts_findings},'scanner_all_overload_findings':scanner_overloads,'new_overload_findings':[f for f in lower_delta['added'] if 'overload' in f['reason'].lower()],'sdk':{'checker':load(p/'e/sdk-adamic-checker.json'),'cohere_findings':[r for r in stream(p/'e/public-adamic-cohere.jsonl') if r.get('kind')=='finding'],'guarded_findings':sdk_findings,'guarded_delta':sdk_delta},'api':load(p/'e/api-analysis.json'),'lane':load(p/'e/lane-api-check.json')}
(p/'e-analysis.json').write_text(json.dumps(result,indent=2)+'\n')
(p/'e/overload-routing.json').write_text(json.dumps(routing,indent=2)+'\n')
(p/'e/scanner-overload-handling.json').write_text(json.dumps({'getLineAndCharacterOfPosition':result['scanner_getLineAndCharacterOfPosition'],'getLineStarts':result['scanner_getLineStarts'],'all_scanner_overload_findings':scanner_overloads,'new_overload_findings_across_ten_files':result['new_overload_findings']},indent=2)+'\n')
print(json.dumps({'sites':result['site_counts'],'cohere_added':len(cohere_delta['added']),'cohere_removed':len(cohere_delta['removed']),'lower_added':len(lower_delta['added']),'lower_removed':len(lower_delta['removed']),'internal_routing':dict(collections.Counter(r['native']['chosen_overload'] for r in routing)),'public_changes':result['api']['normalized_changes'],'lane_errors':len(result['lane']['errors']),'sdk_new_findings':len(sdk_delta['added'])}))
