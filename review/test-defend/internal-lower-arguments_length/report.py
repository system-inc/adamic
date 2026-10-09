import pathlib,json,subprocess,shutil,gzip,re
p=pathlib.Path('review/test-defend/internal-lower-arguments_length');menus=json.loads((p/'menu.json').read_text())+json.loads((p/'extra-menu.json').read_text());runs=json.loads((p/'runs.json').read_text());md={m['id']:m for m in menus};rd={r['id']:r for r in runs};pairs=json.loads((p/'pairs.json').read_text());attempt_ids=[['D01','D02','D03'],['D02','D03','D04'],['D05','D07','D16'],['D05','D06','D07'],['D08','D09'],['D10'],['D07','D15'],['D12'],['D13','D18'],['D11'],['D14']];unique=[None,None,'D16',None,'D08','D10','D15','D12','D18','D11','D14'];rows=[]
cut=['argumentsRefusal','argumentsRefusal','arrayPredicateDomain and arrayPredicateMembers','predicateFlowProof.returned and proveBody, arrayPredicateMembers','arrayIsArray and arrayPredicateDomain','elementType and unknownView','inProperty, unknownView and array predicate flow','widened marker return relation','widened callable relation and censusCallableParameterType','shorthand and functionValue','censusOverload binder mapping and constraint/parameter/result checks']
notes=['No name/assertion mismatch: typed refusals and exact repairs are checked. D01 survived; D02 and D03 were caught by the fixture row too.','No name/assertion mismatch: real refusal fixtures are loaded in both .a and .ts and typed/text refusals checked. All three attempts also failed the direct refusal row.','Node compares arrayness and length; samples do not directly read declared number elements. D16 uniquely tests admission of the declared union, not all element-storage semantics.','No name/assertion mismatch found: unsound annotations must produce Refused with return-is-not-proven text. D05 and D06 did not fail this row; D07 did, but both acceptance rows caught it too.','Pins refusal type and representation reason; D08 admits an unsupported tuple domain and makes the row fail.','Any NotYet suffices, so a different refusal reason still passes. D10 enables erased storage and first subcase now returns nil error.','Node compares unknown reflection and wrapper length behavior. D15 changes the observed property key.','Only Refused type is pinned; a different refusal reason would pass. D12 disables the incompatible marker-result rejection.','Pins the strict optional number relation. D13 was additionally caught by TestProvenRelationsRefuse; D18 isolates optionality erasure.','Pins NotYet and overloaded-value diagnostic. D11 uniquely changes that boundary diagnostic.','Pins relation-specific diagnostic text. D14 is a diagnostic-role defense, not proof of all binder semantics.']
for i,(t,s) in enumerate(pairs.items()):
 attempts=[]
 for id in attempt_ids[i]:
  m=md[id];failed=set(rd[id]['rows_failed'])
  if id+'-replay' in rd:failed.update(rd[id+'-replay']['rows_failed'])
  attempts.append(dict(mutant=id,file_line=m['file']+':'+str(m['line']),change=m['old'].strip()+' -> '+m['new'].strip(),rows_failed=sorted(failed)))
 id=unique[i] or attempt_ids[i][-1];r=rd[id];err=next((e['Output'].strip() for e in r['errors'] if e.get('Test','').split('/')[0]==t),None);failing=err or 'This target passed '+id+'; failures were '+', '.join(r['rows_failed'])+'.';cmd='ADAMIC_BUILD_CACHE_DIR=/tmp/defend-arguments/cache/'+id+' '+' '.join(r['command'][:10])+' "<regex recorded in runs.json>" > '+id+'.log 2>&1'
 passed=set(r['rows_passed']);selected=set(r['rows_failed'])|passed
 if id+'-replay' in rd:passed.update(rd[id+'-replay']['rows_passed']);selected.update(rd[id+'-replay']['rows_passed']);selected.update(rd[id+'-replay']['rows_failed'])
 rows.append(dict(test=t,package='internal/lower',prior_verdict='untrue' if s is None else 'subsumed',subsumed_by=[s] if s else [],defense='defended' if unique[i] else 'not defended',unique_mutant=(id+' '+md[id]['file']+':'+str(md[id]['line'])) if unique[i] else None,attempts=attempts,evidence=cmd+'; '+failing,bounded=True,matrix_rows=sorted(selected),rows_passed=sorted(passed),code_under_test='Adamic Lower: '+cut[i],oracle='Node source execution versus lowered JavaScript stdout/stderr/exit' if i in [2,6] else 'Self-pinned lowering diagnostic type and/or text; no external authority run',oracle_kind='external-run' if i in [2,6] else 'self',owner_finding=notes[i]))
 forid=unique[i]
 if forid:assert rd[forid]['rows_failed']==[t],(t,forid)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
# Preserve complete observed pass/fail lists, including negative replay catchers.
for r in runs:
 expected=3 if r['id'].endswith('-replay') else 105 if r['id'] in ['D16','D18'] else 102
 assert len(r['rows_failed'])+len(r['rows_passed'])==expected,(r['id'],expected)
# Derive exclusive covered lines from Go's covered block ranges, retaining exact blocks too.
def lines(f):
 out=set()
 for l in f.read_text().splitlines()[1:]:
  if int(l.split()[-1])==0:continue
  loc=l.split()[0];file,rng=loc.rsplit(':',1);a,b=rng.split(',');out.update(file+':'+str(n) for n in range(int(a.split('.')[0]),int(b.split('.')[0])+1))
 return out
exclusive={}
for t,s in pairs.items():exclusive[t]=sorted(lines(p/(t+'.cover'))-lines(p/((s+'.cover') if s else (t+'-rest.cover'))))
(p/'exclusive-lines.json').write_text(json.dumps(exclusive,indent=2)+'\n')
for name in ['run.py','extra.py','report.py']:shutil.copy('/tmp/defend-arguments/'+name,p/name)
summary={'start_commit':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'nproc':int(subprocess.check_output(['nproc'],text=True)),'defended':8,'not_defended':3,'distinct_mutants':17,'setup_seconds':0,'matrix_and_replay_seconds':sum(r['seconds'] for r in runs),'solo_coverage_seconds':sum(r['seconds'] for r in json.loads((p/'coverage-runs.json').read_text()))};(p/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
report='''Eight rows defended within the observed bounded matrices; three not defended after three attempts each.
Starting origin/main76c59c81; whole package cooked at90s; clean102-row baseline26.463s.
No test changes; every standalone diff applies and passed go vet; complete pass lists are in runs.json and rows.json.

The package grew from239 to275 tests, with36 additions and no vanished requested rows. The current package was listed and attempted as a whole; after its budget timeout, the matrices used current reached-feature test files, including new tests there. They do not establish package-wide uniqueness. Targeted follow-ups expanded selected observations to105 rows. Rows outside the explicitly recorded sets remain unknown; no deletion follows from these findings. All matrix rows completed for every mutant; no mutant run cooked or aborted the binary.

CODE UNDER TEST and ORACLE: code-and-oracles.md and each row name the production lowering functions and exact current oracle. Node agreement is now part of the two array acceptance rows and the old arguments-neighbor subsumer. No oracle, preparation helper, test or harness was mutated. Every native build cache is isolated per trial.

Coverage: clean profiles for all11 targets and all named subsumers, plus each formerly untrue target against the other101 bounded rows. coverage-differences.json preserves exact exclusive covered Go blocks; exclusive-lines.json derives line numbers from these blocks, not instruction tracing. Semantic leads, including shared-line input differences, are recorded before outcomes in menu.json/extra-menu.json. All three nondefenses have three honest aimed attempts. D01 is an equivalent candidate on the observed corpus: dropping the parenthesis traversal changed no assertion outcome. D06 also survives this matrix because the remaining exact true-narrowing check rejects the unsound contracts; changed behavior outside this corpus is not established.

Results:
'''+json.dumps(rows,indent=2)+'''\n
Brief ambiguities and costs:
- The audit's starting commit was7b18d057; current main is76c59c81. Tests and oracle strength changed materially, so its kill sets are evidence of the old sample rather than current subsumption proofs.
- Whole package at90s without assertion failures is a budget timeout, not a red baseline. Clean bounded baseline and all additional row baselines passed. OptionalWideningCensus and OriginalCycleLedger lacked supplied opt-in inputs; the mixed-union subcase is explicitly deferred. Skips are retained in skips.json.
- Big-package narrowing permits a reached-row matrix, but the defended definition says no other package row. Here defended means unique in the explicitly observed bounded set, with outside catches unknown. The original102-row selection did not contain two diagnostic catchers discovered later; D13 replay failed TestProvenRelationsRefuse, so D13 is not used as a unique defense. D18 uniquely isolates optionality erasure in105 rows.
- The seven-mutant guidance refers to a near-minute whole-package matrix. The narrowed baseline took26.463s, so I used17 distinct mutants. Required isolated cold native caches raised many runs to36-56s, increasing cost beyond the nominal20-minute budget. These are total invocation durations, not individually instrumented clang durations.
- Diagnostic constant mutations count under the allowed menu. The shorthand and binder defenses prove named diagnostic boundaries; they do not claim runtime semantic execution for those refused programs.
- The initial npm log path used the wrong working directory and was corrected before the baseline. A transient exec transport disconnect prevented one preparation command; retry succeeded. No automatic approval rejection occurred.
- No name/assertion mismatch was found in the three nondefended rows: the arguments rows pin their stated refusal behavior, and the invented-contract row pins unproven-contract rejection. Their sampled failures overlap other rows. Other owner findings are attached to each row, including acceptance samples lacking direct number-element observations and broad any-NotYet/any-Refused oracles.

Timings and limits: warm toolchain setup skipped, nproc5. npm ci and all test output went to logs; npm duration was not separately instrumented. Whole clean binary90.214s (budget failure); bounded clean binary26.463s. Matrix/replay and solo coverage totals are in summary.json. No repository-wide replay, optional external census/cycle inputs, complete native behavioral validation of these rows or tests outside the bounded matrices were covered. Restored production sources and tests are unchanged.
''';(p/'report.md').write_text(report)
# Lossless compression preserves outputs while making the pushed evidence practical.
for f in p.glob('*.log'):
 with gzip.open(str(f)+'.gz','wb') as g:g.write(f.read_bytes())
 f.unlink()
print(json.dumps(summary));print([(r['test'],r['defense'],r['unique_mutant']) for r in rows])
