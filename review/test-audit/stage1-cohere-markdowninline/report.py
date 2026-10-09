import json,re,statistics,pathlib,shlex,subprocess
p=pathlib.Path('/tmp/u131/evidence'); repo=pathlib.Path('/workspace/adamic')
cat=json.loads((p/'catalog.json').read_text()); runs=json.loads((p/'runs.json').read_text()); run={x['id']:x for x in runs}
def events(id):
 out=[]
 for s in (p/(id+'.log')).read_text().splitlines():
  try: out.append(json.loads(s))
  except: pass
 return out
def fails(id): return [x['Test'] for x in events(id) if x.get('Action')=='fail' and x.get('Test') and '/' not in x['Test']]
def group(t):
 if t.startswith('TestProduct_'):return 'TestProduct_MarkdownInline family'
 if t.startswith('TestMarkdownInline_') or t=='TestMarkdownInlineUnion':return 'TestMarkdownInline family'
 return t
names=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]
groups=['TestDelimiterExpressionMatchesNode','TestProduct_MarkdownInline family','TestMarkdownInline family','TestMarkdownInlineShardUnion','TestMarkdownInlineShardSelector','TestMarkdownInlineProductSourceInputs','TestMarkdownInlineShardAssignmentStable']
members={g:[n for n in names if group(n)==g] for g in groups}; assert sum(map(len,members.values()))==2074
(p/'family-members.json').write_text(json.dumps(members,indent=2)+'\n')
matrix=[]
for c in cat:
 id=c['id']; f=fails(id); matrix.append({'id':id,'kind':c['kind'],'failed_tests':f,'failed_rows':sorted(set(map(group,f))), 'command':run[id]['command'],'environment':run[id]['environment'],'wall_seconds':run[id]['wall'],'exit':run[id]['exit']})
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
mt={x['id']:x for x in matrix}
def seconds(id):
 for e in events(id):
  s=e.get('Output',''); m=re.search(r'^(?:ok|FAIL)\s+\S+\s+([0-9.]+)s',s)
  if m:return float(m[1])
 raise Exception(id)
keys=['delimiter','products','inline','union','selector','sources','assignment']
timing={g:[seconds(f'timing-{k}-{i}') for i in [1,2,3]] for g,k in zip(groups,keys)}
(p/'timings.json').write_text(json.dumps(timing,indent=2)+'\n')
def fail_line(id,target):
 f=set(fails(id)); es=events(id)
 for e in es:
  if e.get('Test') in f and group(e['Test'])==target:
   s=e.get('Output','').strip()
   if '.go:' in s and ('disagreement:' in s or 'first byte difference' in s or 'invalid union accepted' in s or 'accepted "3/3"' in s or 'changed prerequisite' in s or 'insertion moved' in s or 'survived' in s or 'no such file' in s or 'clang:' in s):return s
 # find all source lines of failing tests and last
 ss=[e.get('Output','').strip() for e in es if e.get('Test') in f and group(e['Test'])==target and '.go:' in e.get('Output','')]
 return ss[-1] if ss else '--- FAIL: '+next(t for t in f if group(t)==target)
ids=['M4','S6','M2','S1','S2B','S3','S4']
oracles=['Node executes the protected delimiter expression; native bytes must match. A self-written Node output check also requires a\\*b plus newline.','Self-written product recipes require successful construction of Go, Node, lowered C, native and sanitized products; this is a build/setup check, not a formatting comparison.','Go cohere overlay, source Node, emitted JavaScript Node, and pinned Prettier 3.9.6 compare bytes. Self-written corpus coverage and planted-disagreement labels also apply. Under TS mutants Node shares the altered source; Go and Prettier remain independent.','Self-written missing/repeated/foreign shard union assertions.','Self-written selector bounds and selected-shard assertions.','Self-written requirement that changed source content changes the product key.','Self-written stable hash assignment and coverage under corpus insertion.']
rows=[]
for i,g in enumerate(groups):
 id=ids[i]; line=fail_line(id,g)
 # Normalize overlay diagnostics to origin line numbers, preserve raw in separate field.
 orig=line
 for a,b in ({'S1':{'go:645:':'go:667:'},'S3':{'go:704:':'go:711:'},'S6':{'go:315:':'go:318:'},'W1':{'go:637:':'go:649:'}}.get(id,{})).items():orig=orig.replace(a,b)
 kills=[m for m in ['M1','M2','M3','M4'] if g in mt[m]['failed_rows']]
 probes=[m for m in ['P1','P2'] if g in mt[m]['failed_rows']] if i in [0,2] else []
 cmd=' '.join(f'{k}={shlex.quote(v)}' for k,v in run[id]['environment'].items())+' '+shlex.join(run[id]['command'])+' > '+id+'.log 2>&1'
 r={'test':g,'package':'stage1/cohere/markdowninline','file':'stage1/cohere/markdowninline/'+('gaps_test.go' if i==0 else 'inline_build_products_test.go' if i==1 else 'inline_units_test.go; inline_products_shards_test.go' if i==2 else 'inline_products_shards_test.go'),'seconds':statistics.median(timing[g]),'oracle':oracles[i],'oracle_kind':['external-run','self'] if i in [0,2] else 'self','kills':kills,'unique_kills':kills if i in [0,2] else [],'last_proven_fail':id+': '+orig,'verdict':'sacred' if i in [0,2] else 'setup-check','subsumed_by':[],'mutants_in_matrix':4,'probe_kills':probes,'subsumer_seconds':None,'vacuous':False if i in [0,2] else None,'bounded':True,'matrix_rows':groups,'evidence':cmd+'; '+orig,'raw_failing_line':line,'members_file':'family-members.json','timing_samples':timing[g]}
 if i==2:r['witness_evidence']='W1: inline_products_shards_test.go:649: escaped delimiter parity survived (overlay raw line 637); comparison weakened to return nil; W1.log';r['timing_scope']='32 of 2048 shards plus corpus union; full-family timing unknown'
 if i not in [0,2]:r['construction_kills']=[id]
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Product rebuild details exactly as logged, no inferred build phases.
builds={}
for id in ['M1','M2','M3','M4','P1','P2']:
 entries=[]
 for e in events(id):
  m=re.search(r'build (\S+) ([0-9a-f]+) (miss|hit) ([0-9.]+)',e.get('Output',''))
  if m:entries.append(dict(product=m[1],key=m[2],status=m[3],seconds=float(m[4])))
 builds[id]={'command_wall_seconds':run[id]['wall'],'binary_seconds':seconds(id),'logged_miss_seconds':round(sum(x['seconds'] for x in entries if x['status']=='miss'),3),'products':entries}
(p/'builds.json').write_text(json.dumps(builds,indent=2)+'\n')
# Standalone diffs all apply on unchanged starting source.
for c in cat:
 subprocess.run(['git','apply','--check',str(p/(c['id']+'.diff'))],cwd=repo,check=True)
(p/'diff-validation.txt').write_text('All 14 standalone diffs passed git apply --check on d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb source. M1-M4 and P1-P2 compiled through native builds in their matrix runs. Every S/W Go overlay passed go vet; logs retained.\n')
friction='''The live scope is 2074 tests, not a file-sized unit. The entire clean package exhausted its 90-second binary budget at 90.049 seconds. Its timeout is not a red baseline. The narrowed, Prettier-enabled baseline passed in 6.853 seconds. The matrix runs all 26 non-shard tests and 32 shards spaced at ordinal multiples of 64, for 58 raw tests and seven grouped rows. The 2016 omitted shards are unknown, including their effects on uniqueness. Sacred here means unique within the bounded matrix only. Neither global package uniqueness nor repository uniqueness is proven.

The family rule combines all 2048 input wrappers with their corpus union. Whole generated wrapper bodies were mechanically checked against the shared-checker template; each differs only by ordinal. The checker also performs raw-input checks and built-in/planted disagreement witnesses at designated inputs. W1 disabled its comparison and made the embedded witness fail. I retain that evidence without turning this mixed family into a separate witness row. Full member lists are in family-members.json. There were no named slice rows to verify against the old 8de93800f4 inventory and no members vanished from the live list.

Timing the complete family three times would repeat the cooked run. Its reported median is explicitly for the bounded family subset plus corpus union. Product recipes were grouped because all call inlineBuild with different construction inputs. Successful compilation is their expected answer, so I judged that family as setup-check by omitting its required generated C file. That construction edit is separate from the four production mutants and does not establish formatting worthiness.

Warm tools did not provide the optional Prettier oracle. Stage3 API npm ci was run before baseline. Prettier 3.9.6 was installed in a dedicated directory using npm ci, then enabled for the bounded baseline, all timings and matrices. The first whole-package baseline had not enabled that optional oracle. No rows skipped in the accepted bounded runs. The TypeScript mutants also alter the source Node product, which would be a shared-error oracle alone; Go cohere and Prettier remain independent. The protected delimiter fixture and its Node execution were not edited. Its hand-written expected Node bytes were checked during the clean passing run, not assigned an external-authority label.

The brief asks for about three mutants per row but caps separately rebuilt port mutants at four. I used the four-rebuild cap, spread over title quoting, punctuation range lookup, pseudo-setext recognition and native RegExp replacement. Each has a private build cache. Construction edits, weakened comparisons and empty answers are labeled separately and contribute no production kills. Their inserted selector switches are absent because these are standalone rebuilds and private Go overlays.

S2 changed a selector zero-bound check but survived. It is an equivalent candidate: the existing index guard rejects every index when n=0. A predicate check over 121 pairs and the logical explanation are retained. S2B then changed the upper index bound and exposed acceptance of 3/3. M3 survives the bounded matrix, but a separately executed native === input differs from the clean native product and Go cohere. That is a bounded coverage gap; the omitted shards may catch it. No wider claim of unguarded package behavior is warranted.

Go overlays shorten function bodies, shifting reported line numbers. The report maps S1 raw line 645 to origin line 667, S3 raw 704 to origin 711, W1 raw 637 to origin 649, and S6 raw 315 to origin 318. Raw logs are retained. Native rebuild timings are product miss durations recorded by the build helper, not guessed clang-only durations; they may overlap due to parallel test execution. The matrix counts JSON fail actions, not intentional disagreement diagnostics printed by passing witnesses.

The tests generated an untracked go.work.sum, which was removed before committing evidence. The repository Markdown corpus is discovered dynamically, so evidence was copied into review only after all audit runs. Replaying after that copy adds audit Markdown to the corpus; the original corpus pin, counts and case IDs remain recorded in the logs. Full transitive runtime call coverage was not collected. The directly mutated functions and listed port functions were read, but no claim is made that every compiler/runtime helper was exhaustively inventoried. No other packages were run and no repository-wide replay was attempted.'''
(p/'friction.txt').write_text(friction+'\n')
summary=['Unit u131: origin/main d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb; nproc 5.','2074 top-level tests form seven rows; 58 raw tests ran in the bounded matrix.','Bounded verdicts: two sacred rows and five setup-check rows.','Four production mutants: three caught, M3 survived; both empty-entry probes caught.','Evidence: review/test-audit/stage1-cohere-markdowninline/ on the requested audit branch.']
lines='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2)+'\n```\n\n'
lines+='| ID | Origin file:line | Change | Failed rows |\n|---|---|---|---|\n'
for c in cat:
 before=c['before'];after=c['after']
 desc={'S1':'Return nil from union validation','S3':'Drop whole source-hashing loop','S4':'Use corpus index instead of name for key','S5':'Drop last mode by bound -1','W1':'Return nil from byte comparison','S6':'Drop whole required program.c write'}.get(c['id'],before+' -> '+after)
 desc=desc.replace('|','\\|').replace('\n',' ')
 lines+=f"| {c['id']} | {c['file']}:{c['line']} | {desc} | {', '.join(mt[c['id']]['failed_rows']) or 'none'} |\n"
lines+='\nSurvivors: M3 leaves === unescaped, whereas clean native and Go cohere emit two backslashes before ===; native execution commands and exact bytes are in survivor-witness.json. S2 is a construction equivalent candidate, with the redundant zero guard explanation in S2-equivalence.txt.\n\n'+friction+'\n\n'
lines+='Setup was skipped because env.sh worked. API npm ci reported 364 ms, retained in npm.log; Prettier npm ci reported 224 ms. The whole baseline binary cooked at 90.049 s; the accepted baseline took 6.853 s. All seven bounded rows ran alone three times. Sum of recorded audit command wall durations: '+str(round(sum(x.get('wall',0) for x in runs),3))+' s, excluding initial baseline/dependency setup and analysis.\n\n'
lines+='| Mutation/probe | Command wall s | Binary s | Logged product miss seconds |\n|---|---:|---:|---:|\n'
for id,b in builds.items():lines+=f"| {id} | {b['command_wall_seconds']:.3f} | {b['binary_seconds']} | {b['logged_miss_seconds']} |\n"
lines+='\nAll standalone diffs apply to the starting source and compiled with their relevant build tool. Detailed build products, times, commands, raw failures, family members, probes and construction edits are included. Not covered: full-family cost, omitted shards, full transitive call coverage, repository-wide uniqueness.\n'
(p/'REPORT.md').write_text(lines)
(p/'REPLAY.txt').write_text('Apply exactly one ID.diff to the starting commit d054e3578d4e53bef8cb8c9ab82d6ace69cdafeb. Source env.sh. For port/runtime mutants select a fresh ADAMIC_BUILD_CACHE_DIR and run the package with the command in runs.json. Enable ADAMIC_MARKDOWNINLINE_LIBRARY with installed Prettier 3.9.6. S/W edits are allowed construction/witness harness checks; their overlay JSON has session paths and should be regenerated or use standalone diff instead. P1/P2 are empty-answer probes, not production mutants. Matrix uniqueness is bounded. family-members.json lists every live Test; matrix-test-names.json lists the 58 executed tests.\n')
print(json.dumps({'seconds':{g:statistics.median(s) for g,s in timing.items()},'failures':{x['id']:x['failed_rows'] for x in matrix},'builds':{i:{k:v for k,v in b.items() if k!='products'} for i,b in builds.items()}},indent=2))
