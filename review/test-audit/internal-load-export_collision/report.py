import pathlib,json,statistics,subprocess,re
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-load-export_collision'
rows=[l for l in (p/'u024-list-retry.log').read_text().splitlines() if l.startswith('Test')]; unit=rows[:15]
def events(f):
 a=[]
 for l in f.read_text().splitlines():
  try:a.append(json.loads(l))
  except:pass
 return a
def failed(f):return sorted({e['Test'].split('/')[0] for e in events(f) if e.get('Action')=='fail' and e.get('Test')})
menu=json.loads((p/'menu.json').read_text()); matrix={m['id']:failed(p/(m['id']+'.log')) for m in menu}
for row in rows:
 f=p/('P01-'+row+'.log');ev=events(f)
 if any(e.get('Action')=='fail' for e in ev):matrix['P01'].append(row)
matrix['P01']=sorted(set(matrix['P01']))
for id in ['P04','P05']:matrix[id]=failed(p/(id+'.log'))
(p/'matrix.json').write_text(json.dumps({'rows':rows,'kills':matrix,'bounded':False,'P01':'all rows rerun alone after panic; only those observed exits counted'},indent=2))
timing=json.loads((p/'timings.json').read_text()); prod=[m['id'] for m in menu if m['id'].startswith('M')];kills={r:[id for id in prod if r in matrix[id]] for r in rows}
oracles=[('self','Exact custom Adamic refusal string; both module names, name and position.'),('self','Acceptance only; no returned program or resolved binding assertion. Broad configuration errors also kill it.'),('self','Handwritten declaration types and positions, including expanded alias; no outside authority cited.'),('external-authority','tsc 6.0.3 diagnostic TS2322 text and position; independently checked this session in authority-ts2322.log.'),('external-authority','TypeScript diagnostic codes TS2322/TS2375/TS2584 and documented option behavior; no independent check of these exact three fixtures. Codes only, so a different error with the same code could pass.'),('self','Acceptance of isolated module scope; no program inspection. Broad configuration errors also kill it.'),('external-authority','TypeScript number inference and TS2322 codes/positions for imports; exact fixture values not independently checked. Negative assertions inspect codes and positions, not text.'),('self','Adamic prelude contract: never-return panic and one-string console; negative checks only TS2345/count.'),('self','Handwritten Adamic input refusal substrings, plus Load(nil) rejection.'),('external-authority','Test comment cites tsc 6.0.3 acceptance of spec programs; no spec fixture independently rerun through tsc this session. Acceptance only, no program inspection.'),('self','Overlay acceptance plus disk TS2322; no returned overlay program inspection.'),('external-authority','TypeScript acceptance and TS2366 implicit return rule; exact fixtures not independently checked. Negative checks only diagnostic code.'),('external-authority','Official @types/node 25.3.3 fs signatures; checked existsSync boolean at fs.d.ts:3762. Asserts declaration identity, then only TS2322 for boolean rejection.'),('self','Repository-owned pin 25.3.3; only substring rejection, so another error mentioning the pin could pass.'),('external-authority','Official @types/node 25.3.3 Console.log(...data:any[]):void at console.d.ts:107 checked this session. Acceptance only, no program inspection.')]
result=[]
for i,r in enumerate(unit):
 k=kills[r];unique=[id for id in k if matrix[id]==[r]]
 candidates=[s for s in unit if s!=r and k and set(k)<=set(kills[s])]
 if unique:verdict='sacred';sub=[];subseconds=None
 elif candidates:
  verdict='subsumed';sub=[min(candidates,key=lambda s:timing[s]['median'])];subseconds=timing[sub[0]]['median']
 elif k:
  verdict='overlapping';sub=[];remaining=set(k)
  while remaining:
   s=max([s for s in unit if s!=r],key=lambda s:len(remaining & set(kills[s])))
   if not (remaining & set(kills[s])):break
   sub.append(s);remaining-=set(kills[s])
  subseconds=None
 else:verdict='untrue';sub=[];subseconds=None
 chosen=unique[-1] if unique else k[-1];ev=events(p/(chosen+'.log'))
 assertion=next((e['Output'].strip() for e in ev if e.get('Test','').split('/')[0]==r and e.get('Action')=='output' and re.search(r'_test.go:\d+:',e.get('Output',''))),'')
 fail=next((e['Output'].strip() for e in ev if e.get('Test')==r and e.get('Output','').startswith('--- FAIL:')),'')
 own=['P01']
 if r=='TestOverlayPrecedesDiskForAdamicAliases':own=['P02']
 if r=='TestNodeLibraryRejectsDifferentVersion':own=['P03']
 if r=='TestDeclarationsCarryTheirProvenTypes':own=['P04']
 if r=='TestNodeLibraryUsesPinnedDeclarations':own=['P01','P05']
 probe_kills=[id for id in own if r in matrix[id]]
 passes=[]
 for id in own:
  f=p/(('P01-'+r+'.log') if id=='P01' else id+'.log')
  passes += [e['Test'] for e in events(f) if e.get('Action')=='pass' and e.get('Test','').startswith(r+'/')]
 file='internal/load/'+('export_collision_test.go' if i<2 else 'node_library_test.go' if i>=12 else 'load_test.go')
 result.append(dict(test=r,package='internal/load',file=file,seconds=timing[r]['median'],oracle=oracles[i][1],oracle_kind=oracles[i][0],kills=k,unique_kills=unique,last_proven_fail=chosen+' '+fail,verdict=verdict,subsumed_by=sub,mutants_in_matrix=20,probe_kills=probe_kills,subsumer_seconds=subseconds,vacuous=not bool(probe_kills),vacuous_subcases=passes if probe_kills else [],bounded=False,matrix_rows=rows,evidence='ADAMIC_MUTANT='+chosen+' timeout 120 go test -json -count=1 -timeout 90s ./internal/load/ -run . > '+chosen+'.log 2>&1; '+assertion))
(p/'results.json').write_text(json.dumps(result,indent=2))
summary={'verdicts':{v:sum(o['verdict']==v for o in result) for v in ['sacred','subsumed','overlapping']},'vacuous':[o['test'] for o in result if o['vacuous']],'baseline_seconds':1.225,'timing_binary_seconds':sum(sum(t['runs']) for t in timing.values()),'vet_seconds':sum(v['seconds'] for v in json.loads((p/'validation.json').read_text())),'matrix_binary_seconds':sum(e.get('Elapsed',0) for m in menu for e in events(p/(m['id']+'.log')) if e.get('Action') in ['pass','fail'] and not e.get('Test'))}
(p/'summary.json').write_text(json.dumps(summary,indent=2))
for m in menu:
 rc=subprocess.run(['git','apply','--check',str(p/(m['id']+'.diff'))],cwd=root,capture_output=True)
 assert rc.returncode==0,(m['id'],rc.stderr)
for id in ['P04','P05']:
 rc=subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],cwd=root,capture_output=True);assert rc.returncode==0,(id,rc.stderr)
print(json.dumps(summary,indent=2))
for o in result:print(o['test'],o['verdict'],o['kills'],o['subsumed_by'])
