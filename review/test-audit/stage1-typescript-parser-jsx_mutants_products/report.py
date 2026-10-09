import pathlib,json,re,statistics,difflib,shlex,subprocess
p=pathlib.Path('/tmp/u155/evidence');root=pathlib.Path('/workspace/adamic');file='stage1/typescript/parser/jsx_mutants_products_test.go';base=(root/file).read_text();members=json.loads((p/'members.json').read_text());groups=list(members);runs=json.loads((p/'runs.json').read_text());run={x['id']:x for x in runs};cat=json.loads((p/'catalog.json').read_text())
def ev(id):
 es=[]
 for s in (p/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(s))
  except:pass
 return es
def grp(t):return next((g for g,mm in members.items() if t in mm),None)
def fails(id):return [e['Test'] for e in ev(id) if e.get('Action')=='fail' and e.get('Test')]
def sec(id):
 for e in ev(id):
  m=re.search(r'^(?:ok|FAIL)\s+\S+\s+([0-9.]+)s',e.get('Output',''))
  if m:return float(m[1])
 return None
matrix=[]
for c in cat:
 id=c['id'];rr=run[id];matrix.append({'id':id,'kind':c['kind'],'failed_tests':fails(id),'failed_rows':sorted(set(g for t in fails(id) if (g:=grp(t)))),'passed_tests':[e['Test'] for e in ev(id) if e.get('Action')=='pass' and e.get('Test')],'skipped_tests':[e['Test'] for e in ev(id) if e.get('Action')=='skip' and e.get('Test')],'command':rr['command'],'environment':rr['environment'],'wall_seconds':rr['wall'],'binary_seconds':sec(id),'exit':rr['exit']})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));m={x['id']:x for x in matrix}
def origin(id,s):
 changed=(p/(id+'.go.txt')).read_text();mp={}
 for a in difflib.SequenceMatcher(None,base.splitlines(),changed.splitlines()).get_matching_blocks():
  for k in range(a.size):mp[a.b+k+1]=a.a+k+1
 return re.sub(r'jsx_mutants_products_test.go:(\d+):',lambda x:'jsx_mutants_products_test.go:'+str(mp.get(int(x[1]),int(x[1])))+':',s)
def diagnostic(id,g,needle):
 ff=set(fails(id))
 for e in ev(id):
  s=e.get('Output','').strip()
  if e.get('Test') in ff and grp(e['Test'])==g and needle in s:return origin(id,s),s
 raise Exception((id,g))
timing={g:[sec(f'timing-{i}-{j}') for j in [1,2,3]] for i,g in enumerate(groups)};(p/'timings.json').write_text(json.dumps(timing,indent=2))
# Own entries are assigned per raw member. Failure in another entry's preparation is not a vacuity kill.
entry_map={}
for n in members[groups[0]]:entry_map[n]='P2' if n=='TestJsxMutantsUnion' else 'P1'
for n in members[groups[1]]:entry_map[n]='P3' if n=='TestProduct_JsxMutantsOracle' else 'P4' if 'Lower_' in n else 'P5'
(p/'probe-entry-map.json').write_text(json.dumps(entry_map,indent=2));assert all(n not in fails(probe) for n,probe in entry_map.items())
rows=[]
for i,g in enumerate(groups):
 id='W1' if i==0 else 'S1';diag,raw=diagnostic(id,g,'owner failed to catch planted survivor' if i==0 else 'program.c');rr=run[id]
 row={'test':g,'package':'stage1/typescript/parser','file':file,'seconds':statistics.median(timing[g]),'oracle':'Executed typescript-go AST bytes compared with an unchanged native control; deliberate Node/native mutant output must differ. Self-written child proof requires exactly owner 004 to reject its planted Node survivor; self-written union checks coverage.' if i==0 else 'Self-written build recipes and successful product preparation. Go builds the oracle, and lowering/native compilation build control and mutant products; wrappers ignore returned paths and do not compare parser answers.','oracle_kind':['external-run','self'] if i==0 else 'self','kills':[],'unique_kills':[],'last_proven_fail':id+': '+diag,'verdict':'witness' if i==0 else 'setup-check','subsumed_by':[],'mutants_in_matrix':0,'probe_kills':[],'subsumer_seconds':None,'vacuous':True,'bounded':True,'matrix_rows':groups,'evidence':shlex.join(rr['command'])+' > '+id+'.log 2>&1; '+diag,'members':members[g],'timing_samples':timing[g],'own_probes':['P1','P2'] if i==0 else ['P3','P4','P5'],'probe_entry_map':'probe-entry-map.json','raw_failing_line':raw,'construction_kills':['S2'] if i==0 else ['S1']}
 rows.append(row)
(p/'rows.json').write_text(json.dumps(rows,indent=2))
builds={}
for id in ['bounded-baseline','W1','S1']:
 products=[]
 for e in ev(id):
  mm=re.search(r'build (\S+) ([0-9a-f]+) (miss|hit) ([0-9.]+)',e.get('Output',''))
  if mm:products.append({'product':mm[1],'key':mm[2],'status':mm[3],'seconds':float(mm[4])})
 builds[id]={'binary_seconds':sec(id),'logged_miss_seconds':round(sum(x['seconds'] for x in products if x['status']=='miss'),3),'products':products}
(p/'builds.json').write_text(json.dumps(builds,indent=2))
for c in cat:subprocess.run(['git','apply','--check',str(p/(c['id']+'.diff'))],cwd=root,check=True)
initial_skips=[e['Test'] for e in ev('baseline') if e.get('Action')=='skip' and e.get('Test')];(p/'baseline-skips.json').write_text(json.dumps(initial_skips,indent=2))
friction='''The brief calls this 22 rows but supplies 31 Test functions. All 31 exist in the named file at current origin/main ef819b8e03c26ee3c5ac7bc1d8f067b77d553640; none moved or vanished. Applying the family rule produces two rows. The nine execution wrappers differ only by shard input to jsxMutantsRun; their coverage union joins that family. All 21 product wrappers reach jsxMutantsFetch and buildcache.Product with different recipes, covering an oracle, a control and nine mutants at lower/native levels. They assert successful construction and no parser answer. Full members and individual matrix outcomes are retained, so central review can split the family differently without losing observations.

The whole clean package timed out at the binary's 90.020-second line. No assertion failure was recorded before the timeout; later rows are unknown. The entire requested 31-test slice then passed in 64.336 seconds. All matrix runs use that same slice with no shard-selection variable and no skipped members. The initial baseline lacked the optional pinned TypeScript compiler corpus. I fetched its exact 050880ce59e30b356b686bd3144efe24f875ebc8 commit, then enabled ADAMIC_TYPESCRIPT_SOURCE for the timing and matrix runs. This slice does not call compilerManifest, so that corpus does not affect its oracle. The initial skip list is in baseline-skips.json; out-of-slice optional tests were not rerun because the whole package was already over budget.

This is a witness/product-construction unit. Production printer/parser defects cannot decide whether the witnesses detect a disabled agreement check. I used the brief's allowed weakened comparison and construction edits, and kept their failures out of production kills. mutants_in_matrix is zero. The nine port mutants the suite builds are pre-existing test inputs, not audit mutants chosen here. Their successful native construction and execution are baseline evidence; they do not establish production mutation coverage or uniqueness for this audit.

W1 changes the survivor condition to false && bytes.Equal. Each normal deliberate mutant still executes successfully, but its survivor guard cannot fail. The child proof plants the oracle answer as owner 004's Node output. That child now passes, and its parent reports owner failed to catch planted survivor. Exactly TestJsxMutants_004 fails; the other eight leaves pass. This is direct weakened-check evidence for the grouped witness verdict. The control parser comparison and external typescript-go oracle were left intact. A broken native/control precondition would not have been counted as witness proof.

The normal mutant check accepts any aggregate AST-output difference from typescript-go. It does not demand the intended changed node, nor prove unaffected inputs agree. The unaltered control comparison is stronger: every AST byte must match typescript-go. Commands require successful execution and no stderr. The planted child overrides Node output only; no separately planted native survivor was tested. The child proof additionally requires a specific error string, failed owner leaf and successful nonowner leaves; it therefore checks ownership rather than merely counting failures. The union supplies self-written enumeration and coverage checks, not an outside semantic oracle.

S1 misnames generated C as missing-program.c. Lowering still constructs its product, but native preparation cannot read program.c. It catches the product family's native members; other members remain green. A fresh ADAMIC_BUILD_CACHE_DIR isolates this construction defect from successful cached artifacts. S2 drops the seen[id] assignment from the union; the coverage member reports zero seen cases versus its complete enumeration. Both are suite-construction edits permitted by the brief, not production changes. Their member-level failures are retained instead of pretending every member of a family failed.

Empty-entry probes have different owners. P1 returns from jsxMutantsRun and P2 returns from jsxMutantsUnion; all members of their family pass their own entry probe, because P1 also skips its child proof. P3 returns an empty oracle path, P4 an empty lower-product path, and P5 an empty native path. The corresponding product wrappers ignore these returns and pass. P4 does fail native wrappers that read the absent lower-product prerequisite, but those wrappers' own entry is the native builder, and their own P5 passes. Counting P4 against native-wrapper vacuity would violate the own-entry rule. probe-entry-map.json records the mapping. Both grouped rows are vacuous under their own entries despite their demonstrated witness/setup verdicts.

The product family has several entry points, so vacuity cannot be inferred from a single whole-row probe run. I evaluated each constituent against its own entry and aggregated those observations. All constituents passed their own entry probes. No probe is used to award witness or setup-check. P1/P4/P5 remove imports made unused by their dropped function bodies, so each standalone diff compiles; those removals are documented in catalog.json. Private Go overlays avoid modifying tracked source, and origin line numbers are mapped from unchanged diagnostic lines while raw logs are retained.

The fixed production-mutant menu and separate native-rebuild cap are not applicable to these allowed harness/check edits. No selector switch was added: private overlays compile each edit, and a dedicated cache handles the one construction edit that must force a fresh build. Build product miss durations can overlap and do not measure clang alone. Child proof output is captured inside a parent's log, so the matrix counts only actual top-level JSON fail actions; a failed proof child expected by a passing parent is not a matrix kill.

Evidence was copied after all tests. It includes the list, full family membership, exact commands and environments, raw outputs, probe owners, standalone diffs and replay instructions. No unrelated packages were run. Full transitive compiler/runtime function coverage was not measured; the local check/construction functions and their preparation helpers were read. No production parser mutation coverage, repository-wide uniqueness or out-of-slice verdict is claimed. Warm env.sh avoided setup; npm ci in stage3/api still ran before baseline.'''
(p/'friction.txt').write_text(friction+'\n')
summary=['u155: origin/main ef819b8e03c26ee3c5ac7bc1d8f067b77d553640; nproc 5.','All 31 requested Tests exist; family grouping yields two rows.','Verdicts: JSX witness family and product setup-check family.','Both rows pass every member\'s own empty-entry probe and are vacuous.','Evidence: review/test-audit/stage1-typescript-parser-jsx_mutants_products/.']
s='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for c in cat:
 desc={'W1':'Disable survivor comparison','S1':'Rename required generated C artifact','S2':'Drop whole seen[id] assignment','P1':'Return at runner entry','P2':'Return at union entry','P3':'Return empty oracle path','P4':'Return empty lower path','P5':'Return empty native path'}[c['id']]
 s+=f"| {c['id']} | {c['file']}:{c['line']} | {desc} | {', '.join(m[c['id']]['failed_rows']) or 'none'} |\n"
s+='\nSurvivors: no production mutants were planted. All three check/construction edits are caught. Empty-entry survivors are probe findings, not equivalent production candidates.\n\n'+friction+'\n\n'
s+='Setup skipped because env.sh worked. API npm ci reported 348 ms. Whole baseline binary timed out at 90.020 s; slice baseline passed in 64.336 s. Both rows ran alone three times. Total elapsed approximately eleven minutes. Corpus fetch wall timing was not recorded. Subsequent command wall total: '+str(round(sum(x['wall'] for x in runs),3))+' s. Per-product rebuild times are in builds.json.\n\n'
for id,b in builds.items():s+=f"{id}: binary {b['binary_seconds']} s; sum of logged misses {b['logged_miss_seconds']} s.\n\n"
s+='Not covered: tests outside the requested slice, completed optional whole-package runs, production parser mutation coverage, full transitive function coverage and repository-wide uniqueness. All eight standalone diffs apply to the starting source and all Go overlays passed go vet. Production source is unchanged.\n';(p/'REPORT.md').write_text(s)
(p/'REPLAY.txt').write_text('Start from ef819b8e03c26ee3c5ac7bc1d8f067b77d553640. Apply exactly one ID.diff or regenerate its overlay with current absolute paths. Use the command and environment in runs.json. S1 requires a fresh cache. W1 is a weakened survivor comparison; S1/S2 are construction edits; P1-P5 are empty-entry probes. None is a production mutant. matrix-test-names.json gives all 31 selected Tests; members.json groups them; probe-entry-map.json assigns own-entry probes. Count top-level JSON failures, not child failures expected by successful proof parents.\n')
print(json.dumps({'rows':[{k:r[k] for k in ['test','seconds','verdict','vacuous','last_proven_fail']} for r in rows],'matrix':{i:m[i]['failed_tests'] for i in m},'builds':{i:{k:v for k,v in b.items() if k!='products'} for i,b in builds.items()},'initial_skips':initial_skips},indent=2))
