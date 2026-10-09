import pathlib,json,re,subprocess,collections
D=pathlib.Path('review/test-audit/stage1-cohere-estree-gaps');plan=json.loads((D/'plan.json').read_text());commands=json.loads((D/'mutant-commands.json').read_text());timings=json.loads((D/'timings.json').read_text());family='TestJSXAgreement family'
def row(t):return family if re.fullmatch(r'TestJSXAgreement(?:_\d{3})?',t) else t.split('/')[0]
def events(id):
 out=[]
 for line in (D/(id+'.log')).read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
matrix=[]
for m in plan:
 ev=events(m['id']);failed=sorted(set(row(e['Test']) for e in ev if e.get('Action')=='fail' and e.get('Test')));status={r:('fail' if r in failed else 'pass' if any(e.get('Action')=='pass' and e.get('Test') and row(e['Test'])==r for e in ev) else 'unknown') for r in m['matrix_rows']}
 matrix.append(dict(id=m['id'],kind=m['kind'],matrix_rows=m['matrix_rows'],results=status,failing_members=sorted(set(e['Test'] for e in ev if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test'])),command=next(c for c in commands if c['label']==m['id'])))
(D/'matrix.json').write_text(json.dumps(matrix,indent=2))
(D/'matrix.csv').write_text('id,kind,row,result\n'+''.join(f'{m["id"]},{m["kind"]},{r},{s}\n' for m in matrix for r,s in m['results'].items()))
names=['TestPostfixValueGap','TestMethodReplacementGap','TestRawInputGap','TestParserRecoveryGap','TestInterfaceDefaultGap','TestInterfaceTypeMethodGap',family,'TestJSXAgreementPlantedDisagreement','TestJSXOriginalLibraries','TestJSXMutant','TestJSXMutantPlantedSurvivor','TestMiscShardUnionRejectsInvalidEnumeration','TestMiscShardGrowthKeepsAssignments','TestRecoveryCacheInputKeys','TestRecoveryNativeRecipe']
files={};members={}
for p in pathlib.Path('stage1/cohere/estree').glob('*_test.go'):
 s=subprocess.check_output(['git','show','HEAD:'+str(p)]).decode()
 for t in re.findall(r'^func (Test\w+)\(',s,re.M):
  if row(t) in names:files.setdefault(row(t),str(p));members.setdefault(row(t),[]).append(t)
# Use original test locations; all harness changes retain line counts.
def proof(id,r):
 candidates=[]
 for e in events(id):
  text=e.get('Output','').strip()
  if row(e.get('Test',''))!=r or not re.search(r'_test.go:\d+:',text):continue
  if any(x in text for x in ['union:', 'build ', 'yields 1','Node prints','Node 7','Node replacement', 'setup wall:', 'setup wall=', 'BUILD ']):continue
  candidates.append(text)
 return candidates[0] if candidates else 'No assertion line recorded'
oracles={
 names[0]:('Node runs fixture and verifies 7,1; self-written NotYet substring. M05 proves article wording, not postfix implementation.',['external-run','self']),
 names[1]:('Node runs replacement/original fixture; self-written Refused substring requires arrow guidance. M07 proves diagnostic guidance.',['external-run','self']),
 names[2]:('Node source runtime is the expected reader/UTF-8 output; native and emitted JS must match byte-for-byte. Go cohere must distinguish malformed bytes.','external-run'),
 names[3]:('Self-written requirement: source Node and native parser must each exceed 1s. Only timeout checked; any unrelated hang could pass.','self'),
 names[4]:('Node actually prints 5; exact self-written NotYet path:10:12 and prototype-erasure label.',['external-run','self']),
 names[5]:('Node actually prints 1; exact self-written NotYet path:10:12 and label; default-free callback control must lower and run as 1.',['external-run','self']),
 family:('Go cohere ESTree canonical bytes compared with source Node, sanitized native and emitted JS port output.','external-run'),
 names[7]:('Self synthetic agree/disagree values; exactly one actual top-level shard must reject the planted disagreement.','self'),
 names[8]:('Go cohere compared with installed @typescript-eslint/typescript-estree 8.65.0, TypeScript 6.0.3 and Prettier 3.9.6 through Node; exact known differences are self-written. This row runs no Adamic product.',['external-run','self']),
 names[9]:('Go cohere canonical bytes; built-in port mutant must disagree with Node/native/emitted JS. Only W01 weakening decides witness verdict.','external-run'),
 names[10]:('Self synthetic agreement/disagreement; exactly one planned survivor must fail its leaf.','self'),
 names[11]:('Self invalid-enumeration rejection cases.','self'),
 names[12]:('Self shard assignment equality before/after insertion and reordering.','self'),
 names[13]:('Self cache key equal for copied unchanged source, different for changed mutable source.','self'),
 names[14]:('Independent product bytes must agree; native fixtures must trigger Clang undefined/address/leak sanitizer diagnostics.',['self','external-run'])}
probeMap={names[0]:'PLower',names[1]:'PLower',names[2]:'PUtf8',names[3]:'PParser',names[4]:'PLower',names[5]:'PLower',family:'PMain'}
subs={names[2]:family,names[4]:names[5],names[5]:names[4]};special={names[7]:'W01',names[9]:'W01',names[10]:'W01',names[11]:'H01',names[12]:'H02',names[13]:'H03',names[14]:'H04'}
report=[]
for r in names:
 relevant=[m for m in matrix if m['kind']=='production' and r in m['matrix_rows']];kills=[m['id'] for m in relevant if m['results'][r]=='fail'];unique=[m['id'] for m in relevant if m['results'][r]=='fail' and sum(v=='fail' for v in m['results'].values())==1]
 id=special.get(r) or (kills[-1] if kills else None);line=proof(id,r) if id else None;command=(' '.join(k+'='+v for k,v in next(c for c in commands if c['label']==id)['env'].items())+' ADAMIC_ESTREE_LIBRARY=/tmp/u086/library '+next(c for c in commands if c['label']==id)['command'] + (' (selector file /tmp/u086/selector contains '+id+')' if id.startswith('M') and id in ['M01','M02','M03','M04'] else '')) if id else 'Clean baseline and three measurements; row does not invoke Adamic, so oracle mutations forbidden'
 probe=probeMap.get(r);pk=[];v=None
 if probe:
  ev=events(probe);pk=[probe] if any(e.get('Action')=='fail' and e.get('Test') and row(e['Test'])==r for e in ev) else [];v=not bool(pk)
 verdict='witness' if special.get(r)=='W01' else 'setup-check' if r in special else 'cannot-judge' if r==names[8] else 'sacred' if unique else 'subsumed' if r in subs else 'untrue'
 o=dict(test=r,package='stage1/cohere/estree',file=files[r],seconds=timings['TestJSXAgreement' if r==family else r]['median'],oracle=oracles[r][0],oracle_kind=oracles[r][1],kills=kills,unique_kills=unique,last_proven_fail=(id+': '+line) if id else None,verdict=verdict,subsumed_by=[subs[r]] if r in subs else [],mutants_in_matrix=len(relevant),probe_kills=pk,subsumer_seconds=timings['TestJSXAgreement' if subs.get(r)==family else subs[r]]['median'] if r in subs else None,vacuous=v,bounded=True,matrix_rows=sorted(set(x for m in relevant for x in m['matrix_rows'])) if relevant else next(m['matrix_rows'] for m in matrix if m['id']==special[r]) if r in special else [],evidence=command+' > '+(id+'.log; '+line if id else 'baseline-small.log'),members=members[r])
 if r in subs:o['subsumption_basis_mutants']=len(kills)
 if r in special:o['construction_or_witness_kills']=[special[r]]
 if r==family:o['vacuous_subcases']=[e['Test'] for e in events('PMain') if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']];o['vacuous_subcases_note']='Union and empty assigned shards pass; all nine positive cases belong to seven failing nonempty shards.'
 report.append(o)
(D/'rows.json').write_text(json.dumps(report,indent=2))
(D/'family-members.json').write_text(json.dumps(members,indent=2))
probeRows=[]
for id in ['PMain','PParser','PUtf8','PLower']:
 ev=events(id);probeRows.append(dict(id=id,failed_rows=sorted(set(row(e['Test']) for e in ev if e.get('Action')=='fail' and e.get('Test'))),entry={'PMain':'run(path) in main.ts','PParser':'Parser.file()','PUtf8':'native adamic_utf8_length(text)','PLower':'lower.Lower(ctx,program)'}[id],command=next(c for c in commands if c['label']==id)))
(D/'probes.json').write_text(json.dumps(probeRows,indent=2))
for p in D.glob('*.diff'):
 if p.name=='switch.diff':continue
 subprocess.run(['git','apply','--check',str(p)],check=True)
print(collections.Counter(r['verdict'] for r in report));print('all standalone diffs apply')
