import pathlib,json,statistics,shlex,re,shutil
p=pathlib.Path('/tmp/u150/evidence');groups=json.loads((p/'groups.json').read_text());plan=json.loads((p/'plan.json').read_text());runs=json.loads((p/'matrix-timings.json').read_text())
def events(path):
 out=[]
 for s in path.read_text().splitlines():
  try:out.append(json.loads(s))
  except ValueError:pass
 return out
med=[]
for i in range(9):
 vals=[next(e['Elapsed'] for e in reversed(events(p/f'group-{i}-timing-{r}.log')) if e['Action']=='pass' and not e.get('Test')) for r in range(1,4)]
 med.append(statistics.median(vals))
observed={}
for r in runs:
 es=events(p/r['log']);fail=[e.get('Test') for e in es if e['Action']=='fail' and e.get('Test') and '/' not in e['Test']];lines=[e.get('Output','').strip() for e in es if e.get('Output') and re.search(r'\.go:\d+:',e['Output']) and ('build ' not in e['Output'] or 'no such file' in e['Output']) and not any(x in e['Output'] for x in ['corpus-', 'copied ', 'matching ', 'repository ', 'original Prettier', 'recorded original', 'shard case time', 'TestFileDriver (setup)'])]
 observed[(r['id'],r['index'])]=dict(failed=fail,lines=lines,command=shlex.join(r['command']),exit=r['exit'],log=r['log'])
(p/'matrix.json').write_text(json.dumps([dict(id=id,index=i,**v) for (id,i),v in observed.items()],indent=2))
killsets={i:{id for (id,j),v in observed.items() if j==i and id.startswith('M') and id!='M0' and v['failed']} for i in [0,1,2,6]}
files=['format_test.go','compose_test.go','cst_test.go','compose_test.go','cst_test.go','file_driver_shards_test.go','file_driver_shards_test.go','format_test.go','format_test.go']
oracles=['Go cohere formatter exact bytes; lowered Node and emitted JavaScript; original Prettier has 42 input-specific binary/minification exceptions plus count check.','Go cohere composition output compared byte-for-byte with native, Node, emitted JavaScript and original yaml library.','Go cohere CST serialization compared byte-for-byte with native, Node, emitted JavaScript and original yaml library.','Go cohere output and planted port mutants; judged by weakening only the agreement comparison.','Go cohere CST output and planted port mutants; judged by weakening only the agreement comparison.','Native control versus Go cohere, then planted stdout disagreements plus shard enumeration assertions; judged by weakening the planted comparison.','Go cohere file-mode formatter output compared byte-for-byte against native executable, emitted JavaScript and Node for each shard input.','Original yaml and Prettier actually run; output checked against a self-written six-line snapshot of known library differences. No Adamic port is executed.','Suite construction builds and publishes native/Node/Go products; construction break S1 is judged, not production behavior. Parent TestMain preflight child does the real work.']
rows=[]
for i,(name,members) in enumerate(groups):
 kills=sorted(killsets.get(i,set()));unique=[id for id in kills if sum(id in v for v in killsets.values())==1];sub=[];verdict='cannot-judge';subs=None;proof=None;probes=[];vacuous=None;witness=[]
 if i in killsets:
  if unique:verdict='sacred'
  elif kills:
   candidates=[j for j,v in killsets.items() if j!=i and set(kills)<=v]
   if candidates:
    j=min(candidates,key=lambda j:med[j]);sub=[groups[j][0]];subs=med[j];verdict='subsumed'
   else:verdict='overlapping';sub=[groups[j][0] for j,v in killsets.items() if j!=i and set(kills)&v]
  else:verdict='untrue'
  pid={0:'P3',1:'P1',2:'P2',6:'P3'}[i];v=observed.get((pid,i));vacuous=not bool(v['failed']) if v else None
  if v and v['failed']:probes=[pid]
  if kills:proof=(kills[-1],observed[(kills[-1],i)])
 elif i in [3,4,5]:
  id={3:'W1',4:'W2',5:'W3'}[i];v=observed.get((id,i));verdict='witness' if v and v['failed'] else 'untrue';witness=[id] if verdict=='witness' else [];proof=(id,v) if witness else None
 elif i==8:
  v=observed.get(('S1',i));verdict='setup-check' if v and v['exit']==1 else 'cannot-judge';proof=('S1',v) if verdict=='setup-check' else None
  v=observed.get(('P4',i));probes=['P4'] if v and v['exit']==1 else [];vacuous=False if probes else None
 line=proof[1]['lines'][-1] if proof and proof[1]['lines'] else ('setup process failed; see raw log' if proof else None)
 row=dict(test=name,package='stage1/cohere/yaml',file='stage1/cohere/yaml/'+files[i],seconds=med[i],oracle=oracles[i],oracle_kind=['external-run','self'] if i==7 else ('self' if i==8 else 'external-run'),kills=kills,unique_kills=unique,last_proven_fail=proof[0]+': '+line if proof else None,verdict=verdict,subsumed_by=sub,mutants_in_matrix=sum(1 for id,j in observed if j==i and id.startswith('M') and id!='M0'),probe_kills=probes,subsumer_seconds=subs,vacuous=vacuous,bounded=True,matrix_rows=([groups[j][0] for j in [0,1,2,6]] if i in [0,1,2,6] else [name] if i in [3,4,5,8] else []),evidence=(proof[1]['command']+' > '+proof[1]['log']+'; '+line) if proof else 'Three clean runs passed; no admissible Adamic mutation reaches this external-library-only row.')
 if len(members)>1:row['members']=members
 if i==6:row['vacuous_subcases']=['empty file: unchanged Go oracle ok\\t\\n; P3 native exit 0, empty stdout and stderr (empty-file-probe.json)']
 if witness:row['witness_kills']=witness
 if i==8:row['construction_kills']=['S1'] if proof else []
 rows.append(row)
(p/'rows.json').write_text(json.dumps(rows,indent=2))
summary=['Unit u150: all 16 requested names exist at '+ 'a7448d73cd17f16362b6cbc5c5c111080da64e43'+'.','Nine grouped rows: eight file shards form one family; Union has an independent planted-disagreement assertion.','Clean whole-package run hit the outer 120-second limit; all nine selected rows passed three narrowed clean runs.','Four production mutants, three port-entry probes, one setup probe, three witness checks and one construction break were run.','Evidence branch: test-audit/stage1-cohere-yaml-compose; directory: review/test-audit/stage1-cohere-yaml-compose/.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n| ID | Origin file:line | Change | Observed failed rows |\n|---|---|---|---|\n'
for item in plan:
 if item['kind']=='probe':continue
 id=item['id'];fails=[groups[j][0] for (rid,j),v in observed.items() if rid==id and (v['failed'] or (item['kind']=='construction' and v['exit']==1))];change=item.get('change',item.get('old','')+' -> '+item.get('new',''))
 text+='| '+id+' | '+item['file']+':'+str(item['line'])+' | '+change.replace('|','\\|')+' | '+', '.join(fails)+' |\n'
text+='\nSurvivors: '+('; '.join(id+' equivalent candidate: no differing output demonstrated' for id in ['M1','M2','M3','M4'] if not any(id in v for v in killsets.values())) or 'None among the four production mutants in the bounded matrix.')+'\n\n'
text+='The brief and measurement limits\n\n'
text+='The brief says eight rows, but its extra-assertion rule gives nine: TestFileDriverUnion checks a planted disagreement that the eight shards do not. It remains a separate witness. All names stayed in the listed files. The supplied historical hash was not current origin/main; the audit uses the fetched commit named above.\n\n'
text+='TestMain runs file-driver preparation in a child process before the parent test alarm starts. Its 38.76-second cold setup plus the package run exceeded the outer 120-second limit without an earlier failing test. This is a cooked baseline, not a red baseline. Matrices were narrowed to the selected rows that import each changed module, as recorded in plan.json. Unselected rows and all other packages remain unknown. Local subsumption rests on only four production mutants, not a deletion recommendation.\n\n'
text+='Timings use the Go test binary package pass.Elapsed line, not subprocess wall time. File-family timing runs all eight members as one row. The setup parent normally returns immediately because TestMain already prepared products, so the package line includes its child work. Raw command wall times are separate. No row skipped with the exact yaml@2.9.0 and prettier@3.9.6 libraries enabled.\n\n'
text+='TestBundledParserDifference executes only the external original libraries and checks their known six-line difference snapshot. An Adamic production mutant cannot reach it; mutating those libraries would mutate the oracle. Its verdict is cannot-judge and its probe status is null. The formatter oracle separately checks exact expected outputs and accepts 42 individually identified original-library differences, then checks the total; it does not accept arbitrary same-count failures.\n\n'
text+='Every switched matrix command sets ADAMIC_MUTANT to its ID, writes that ID to /tmp/u150/selector, enables ADAMIC_YAML_LIBRARY=/tmp/u150/library, and selects ADAMIC_BUILD_CACHE_DIR=/tmp/u150/cache/switched, except S1 and P4 use fresh ID-specific caches. Exact environment and command construction are in mutate.py. The function inventory is a conservative transitive module/function superset, not a dynamic proof of every call or every anonymous callback. Four production mutants were selected from that inventory before running any mutant. This follows the port-specific four-mutant cap rather than the general three-per-row aim. Runtime file selection allows one native build per product; M0 checks all four functional rows before the matrix. Compiler source was not changed, so the shared switched cache holds exactly the same executable across selections. Each standalone native validation has its own cache.\n\n'
text+='P1-P3 suppress the actual port entry stdout, rather than a preparation helper. P4 returns a nil setup product and may abort the child; its row was run alone. Probe failures are separate from production kills. The witness checks weaken only the comparison, leaving assertions and planted inputs intact. Setup S1 writes the generated C under the wrong name, leaving the later read intact. The first standalone empty-entry guard used if(false), which TypeScript rejected because unreachable-body union narrowing does not apply. The saved probes instead delete the entry body and successfully rebuild. All standalone diffs contain no runtime selector and apply independently to the starting commit.\n\n'
text+='The empty-file subcase was independently replayed and passes P3, as recorded in empty-file-probe.json; the family still rejects P3 on nonempty files. Other positive subcases after an early fatal probe assertion remain unknown. No claims are made about dynamic call coverage, package-wide uniqueness, or repository-wide uniqueness. No test or oracle was edited for production mutants. The construction and witness edits are the explicit exceptions in the brief.\n\n'
text+='Tool setup: warm env, no cloud/setup.sh run; nproc=5. npm ci and exact optional-library installation succeeded; their individual wall times were not instrumented. Whole-package backstop was 120 seconds. Three clean measurement rounds total '+str(round(sum(x['seconds'] for x in json.loads((p/'group-timings.json').read_text())),2))+' shell wall seconds. Matrix commands total '+str(round(sum(x['seconds'] for x in runs),2))+' shell wall seconds, including compilation and external runners. Per-run logs and build timings are in matrix-timings.json and standalone-validation.json. Final restoration is recorded in restored-baseline.log.\n'
if (p/'standalone-validation.json').exists():text+='Standalone build/vet total: '+str(round(sum(x['seconds'] for x in json.loads((p/'standalone-validation.json').read_text())),2))+' seconds.\n'
(p/'REPORT.md').write_text(text)
for f in ['measure.py','mutate.py','validate.py','report.py']:shutil.copy('/tmp/u150/'+f,p/f)
print(json.dumps([(r['test'],r['verdict'],r['kills'],r['last_proven_fail']) for r in rows],indent=2))
