import pathlib,json,re,statistics,difflib,shlex,subprocess
p=pathlib.Path('/tmp/u151/evidence');root=pathlib.Path('/workspace/adamic');file='stage1/cohere/yaml/formatter_mutants_grain_test.go';base=(root/file).read_text();members=json.loads((p/'members.json').read_text());groups=list(members);run={x['id']:x for x in json.loads((p/'runs.json').read_text())};cat=json.loads((p/'catalog.json').read_text())
def ev(id):
 es=[]
 for line in (p/(id+'.log')).read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 return es
def grp(t):return next((g for g,mm in members.items() if t in mm),None)
def fails(id):return [x['Test'] for x in ev(id) if x.get('Action')=='fail' and x.get('Test')]
def sec(id):
 for e in ev(id):
  m=re.search(r'^(?:ok|FAIL)\s+\S+\s+([0-9.]+)s',e.get('Output',''))
  if m:return float(m[1])
 return None
matrix=[]
for c in cat:
 id=c['id']; ff=fails(id); rr=run[id];matrix.append({'id':id,'kind':c['kind'],'raw_failed_tests':ff,'failed_rows':sorted(set(g for t in ff if (g:=grp(t)))),'raw_passed_tests':[x['Test'] for x in ev(id) if x.get('Action')=='pass' and x.get('Test')],'command':rr['command'],'environment':rr['environment'],'wall_seconds':rr['wall'],'binary_seconds':sec(id),'exit':rr['exit']})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));m={x['id']:x for x in matrix}
def origin_line(id,s):
 changed=(p/(id+'.go.txt')).read_text();mp={}
 for block in difflib.SequenceMatcher(None,base.splitlines(),changed.splitlines()).get_matching_blocks():
  for k in range(block.size):mp[block.b+k+1]=block.a+k+1
 return re.sub(r'formatter_mutants_grain_test.go:(\d+):',lambda x:'formatter_mutants_grain_test.go:'+str(mp.get(int(x[1]),int(x[1])))+':',s)
def diagnostic(id,g):
 ff=set(fails(id));outputs=[e.get('Output','').strip() for e in ev(id) if e.get('Test') in ff and grp(e['Test'])==g]
 for s in outputs:
  if 'formatter_mutants_grain_test.go:' in s and ('no such file' in s or 'no file at' in s or 'planted survivor caught by' in s or 'cannot' in s or 'failed' in s):return origin_line(id,s),s
 return next((origin_line(id,s),s) for s in outputs if 'formatter_mutants_grain_test.go:' in s)
timings={g:[sec(f'timing-{i}-{j}') for j in [1,2,3]] for i,g in enumerate(groups)};(p/'timings.json').write_text(json.dumps(timings,indent=2))
rows=[]
oracles=['Go cohere generates the expected-output artifact, but this row checks self-written corpus counts, wrapper enumeration, union and required artifact availability, without comparing formatter answers.','Self-written successful product construction at source, lower and native levels, for six mutant recipes. Successful preparation is the whole assertion; the returned products are not checked by these wrappers.','Go cohere formatting supplies executed expected bytes. Native and source Node run six altered ports; each must differ somewhere. This aggregate inequality oracle accepts any wrong output and does not require the intended changed case.','Self-written synthetic oracle/wrong bytes require exactly shard 3 to reject the planted survivor. No external formatter executes in this row.']
for i,g in enumerate(groups):
 id=['S1','S2','W1','W1'][i]
 if i==2:diag='W1: all six witness members passed with formatterMutantSurvived returning nil; survivor detection disabled';raw=diag;last=None
 else:diag,raw=diagnostic(id,g);last=id+': '+diag
 rr=run[id];cmd=shlex.join(rr['command'])+' > '+id+'.log 2>&1';probe=['P2','P3','P1','P1'][i]
 rows.append({'test':g,'package':'stage1/cohere/yaml','file':file,'seconds':statistics.median(timings[g]),'oracle':oracles[i],'oracle_kind':'external-run' if i==2 else 'self','kills':[],'unique_kills':[],'last_proven_fail':last,'verdict':['setup-check','setup-check','untrue','witness'][i],'subsumed_by':[],'mutants_in_matrix':0,'probe_kills':[probe] if g in m[probe]['failed_rows'] else [],'subsumer_seconds':None,'vacuous':g not in m[probe]['failed_rows'],'bounded':True,'matrix_rows':groups,'evidence':cmd+'; '+diag,'raw_failing_line':raw,'members':members[g],'timing_samples':timings[g],'construction_kills':[id] if i in [0,1] else [],'weakened_check':'W1' if i in [2,3] else None})
(p/'rows.json').write_text(json.dumps(rows,indent=2))
# Native compilation timings are existing fixture construction, never audit production-mutant kills.
builds={}
for id in ['bounded-baseline','S1','S2']:
 products=[]
 for e in ev(id):
  mm=re.search(r'build (\S+) ([0-9a-f]+) (miss|hit) ([0-9.]+)',e.get('Output',''))
  if mm:products.append({'product':mm[1],'key':mm[2],'status':mm[3],'seconds':float(mm[4])})
 builds[id]={'binary_seconds':sec(id),'products':products,'logged_miss_seconds':round(sum(x['seconds'] for x in products if x['status']=='miss'),3)}
(p/'builds.json').write_text(json.dumps(builds,indent=2))
for c in cat:subprocess.run(['git','apply','--check',str(p/(c['id']+'.diff'))],cwd=root,check=True)
friction='''The brief calls this 21 rows but lists 26 Test functions. All 26 exist at current origin/main a7448d73cd17f16362b6cbc5c5c111080da64e43, in the named file; none moved or vanished. Reading the bodies and applying the family rule produces four rows. Eighteen product wrappers share one checker over mutant index and build level; six execution wrappers share one checker over mutant index. The oracle product and the planted survivor have different assertions and remain separate. Complete members are recorded in members.json.

The requested whole package has 72 tests. Its TestMain builds file-driver products in a separate setup process before m.Run. The outer 120-second timeout terminated the whole-package run after that preflight and later tests; the main binary had not yet emitted its own 90-second timeout. No semantic failure was recorded. This is an over-budget incomplete baseline, not a red assertion baseline. The separate full 26-test slice passed in 82.762 seconds. It overlapped the last part of the first baseline, which can affect cold-build timing, but no source edits existed during either baseline. Mutations began only after the slice was green. Other package rows are unknown. Every verdict and cost here is for the requested slice; no package uniqueness is claimed.

The package tools were warm, but optional YAML/Prettier libraries were not assumed warm. API npm ci ran first. yaml@2.9.0 and prettier@3.9.6 were installed through npm ci in a dedicated directory. Those libraries apply to other package tests, not this slice: these witnesses execute Go cohere and Node and have no library opt-in. The accepted slice did not skip. The whole run was terminated before its complete skip inventory could be obtained. No extra out-of-slice run was made just to obtain that inventory.

These rows are witnesses and construction checks, so the normal production-mutant matrix is inappropriate. Changing a working printer is not evidence that a disagreement witness can detect a disabled comparison. I used only the brief's allowed construction and weakened-check edits. Their kills do not count as production kills, and mutants_in_matrix is zero. The native programs compiled by these tests already include six built-in port mutants, but those are the tests' inputs, not production audit mutants. The optional three-mutants-per-row target therefore does not apply. Standalone diffs, Go vet outputs, exact commands and failures are retained for all six edits and probes.

The witness guard has inverted polarity: it returns an error when the deliberate wrong output equals the oracle, and nil when the wrong output differs. W1 makes it return nil unconditionally, removing detection of a surviving mutant. All six execution witnesses still pass. By the brief's witness rule, their family is untrue under this demonstrated weakening. This does not say their ports are correct or that their normal comparisons are useless. It shows they do not themselves reject removal of this guard. The separate planted-survivor witness does reject it, with caught [] instead of [3]. That row is witness, not sacred; the six other witness outputs cannot establish its production uniqueness.

The first overlay vet command placed -overlay before the Go subcommand and exited 2 without validating source. Its command and exit remain in runs.json. I corrected the order, reused the completed timing runs, and every final overlay validation passed. This command assembly error cost about a minute; it contributes no test evidence.

The execution witnesses compare an aggregate output stream with the aggregate oracle. Any difference is accepted, without requiring the intended mutant's affected case or checking that all unaffected cases match. Their command wrapper does require successful execution and rejects unexpected stderr. The expected bytes come from executed Go cohere; no external-authority label is used. The planted-survivor row uses synthetic bytes, so its oracle is self rather than external-run.

P2 initially failed vet because removing its preparation body left encoding/json unused. The standalone probe also removes that now-unused import. Its final compile check passed; the earlier unused-import failure is not a test kill.

The empty-return probes reveal a separate construction weakness. P2 returns an empty oracle product and its own top-level wrapper passes because it ignores the return. P3 returns two empty product paths and all 18 product wrappers pass for the same reason. These rows are vacuous under their own entries' probes even though construction defects S1 and S2 can make them fail elsewhere. P1 returns nil from the guard, so the six execution witnesses pass and are vacuous under this check-entry probe; the planted survivor fails and is not vacuous. These are probes of the actual check/construction entries, not claims about an empty native formatter's behavior. No production formatter probe was run because production formatting is not the object judged for witness rows.

S1 renames the required cases artifact, preserving Go cohere and its expected output. S2 renames copied source artifacts, so lowering cannot find the expected main.ts; the Go oracle remains untouched. Both use isolated build caches so stale successful construction cannot mask the break. Source-level product wrappers still pass S2, while lower/native wrappers fail, which is retained in the raw matrix. The family verdict rests on those observed failures, not on every member failing. Overlay edits shorten functions and shift diagnostics; reported lines are mapped back to origin with matching unchanged lines, and raw lines remain available.

Evidence was added only after tests because corpus enumeration discovers repository files. The tests' corpus pins, counts and build keys are recorded before the audit Markdown was added. Replay after publication may include additional review files. Build helper durations below are logged misses, may overlap, and are not clang-only times. No full transitive port/runtime coverage was measured, no unrelated packages were run, and no repository-wide uniqueness was attempted.'''
(p/'friction.txt').write_text(friction+'\n')
summary=['u151: starting origin/main a7448d73cd17f16362b6cbc5c5c111080da64e43; nproc 5.','All 26 named Tests exist; family grouping yields four rows.','Verdicts: two setup-check, one witness, one untrue witness family.','Three empty-entry probes: two product rows and the six-witness family pass their own probes.','Evidence pushed under review/test-audit/stage1-cohere-yaml-formatter_mutants_grain/.']
s='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for c in cat:s+=f"| {c['id']} | {c['file']}:{c['line']} | {c['kind']}: {c['after'].replace(chr(10),' ')} | {', '.join(m[c['id']]['failed_rows']) or 'none'} |\n"
s+='\nSurvivors: no production mutants were planted. W1 survives the six execution-witness members, while the planted-survivor test fails; this is a demonstrated weakened-check finding, not equivalent production behavior. P2/P3 surviving product wrappers are empty-answer findings, not mutant survivors.\n\n'+friction+'\n\n'
s+='Setup skipped: warm env.sh worked. API npm ci: 334 ms; library npm ci: 359 ms. Whole baseline hit outer timeout 120 s; slice baseline: 82.762 s. Each of four rows ran alone three times. Recorded subsequent command wall total: '+str(round(sum(x['wall'] for x in run.values()),3))+' s. Total elapsed approximately eleven minutes, including reading and analysis. Detailed build product times are in builds.json.\n\n'
for id,b in builds.items():s+=f"{id}: binary {b['binary_seconds']} s; sum of logged product misses {b['logged_miss_seconds']} s.\n\n"
s+='Not covered: the other 46 package tests, a completed whole-package baseline/skip inventory, production formatter mutation coverage, full transitive function coverage, repository-wide uniqueness. All standalone diffs apply to the starting source and their Go overlays passed go vet. Production source was left unchanged.\n';(p/'REPORT.md').write_text(s)
(p/'REPLAY.txt').write_text('Start from a7448d73cd17f16362b6cbc5c5c111080da64e43. Apply one ID.diff, or regenerate its Go overlay with current absolute paths. Run the corresponding runs.json command and environment. S1/S2 need separate fresh ADAMIC_BUILD_CACHE_DIR values. W1 is a weakened comparison; P1-P3 are empty-entry probes. None is a production mutant. The bounded matrix is all 26 names in matrix-test-names.json. members.json gives complete families. Read raw logs and JSON fail actions, not intentional wrong-byte diagnostics, to derive kills.\n')
print(json.dumps({'rows':[{k:r[k] for k in ['test','seconds','verdict','vacuous','last_proven_fail']} for r in rows],'matrix':{i:m[i]['failed_rows'] for i in m},'builds':{i:{k:v for k,v in b.items() if k!='products'} for i,b in builds.items()}},indent=2))
