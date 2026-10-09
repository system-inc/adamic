from pathlib import Path
import json,subprocess,shutil
p=Path('/tmp/defend-parser-namespaces');root=Path('/workspace/adamic');base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();dest=root/'review/test-defend/internal-oracle-parser_namespaces/sessions'/base[:8];dest.mkdir(parents=True,exist_ok=True)
for f in p.iterdir():
 if f.is_file():shutil.copy2(f,dest/f.name)
def ev(name):return [json.loads(l) for l in (p/name).read_text().splitlines() if l.startswith('{')]
matrix={};attempts=[];lines={}
for mid,file,_,__,line,change in json.loads((p/'plan.json').read_text())['mutants']:
 es=sum([ev(mid+'.'+label+'.log') for label in ['target','subsumer','expanded','review']],[])
 roots={e['Test']:e['Action'] for e in es if e['Action'] in ['pass','fail','skip'] and e.get('Test') and '/' not in e['Test']}
 fail=[k for k,v in roots.items() if v=='fail']
 grouped=[]
 if any(k in fail for k in ['TestParserNamespaceReceiver','TestParserNamespaceClass','TestParserCallableNamespace']):grouped.append('TestParserNamespace family')
 grouped+=['TestNativeAgreesWithNode family (six selected inputs)' if k=='TestNativeAgreesWithNode' else k for k in fail if not k.startswith('TestParser')]
 proofs=[e['Output'].strip() for e in es if '.go:' in e.get('Output','') and (e.get('Test','').split('/')[0] in fail) and ('Lower:' in e['Output'] or 'refuses this' in e['Output'] or 'stage 0' in e['Output'] or 'exit codes differ' in e['Output'])]
 lines[mid]=proofs
 matrix[mid]={'failed_rows':grouped,'top_level_results':roots,'passed_subtests':[e['Test'] for e in es if e['Action']=='pass' and '/' in e.get('Test','')],'failed_subtests':[e['Test'] for e in es if e['Action']=='fail' and '/' in e.get('Test','')],'failure_lines':proofs}
 attempts.append({'mutant':mid,'file_line':file+':'+str(line),'change':change,'rows_failed':grouped})
row={'test':'TestParserNamespace family','package':'internal/oracle','prior_verdict':'subsumed','subsumed_by':['TestNativeAgreesWithNode family (D in audit)'],'defense':'not defended','unique_mutant':None,'attempts':attempts,'evidence':'commands.json records timeout 120 go test -json -count=1 -timeout 90s ./internal/oracle/ -run <bounded selectors>; D1: '+lines['D1'][0]+'; D2: '+lines['D2'][0]+'; D3: '+lines['D3'][0],'bounded':True,'members':['TestParserNamespaceReceiver','TestParserNamespaceClass','TestParserCallableNamespace'],'base':base}
(dest/'rows.json').write_text(json.dumps([row],indent=2)+'\n');(dest/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
def cov(name):return {s[0]:int(s[2]) for l in (p/(name+'.cover')).read_text().splitlines()[1:] if len(s:=l.split())==3}
a,b=cov('target'),cov('subsumer');exclusive=[k for k,v in a.items() if v and not b.get(k)]
(dest/'coverage-diff.json').write_text(json.dumps({'coverpkg':'./internal/lower','target_covered_blocks':sum(bool(v) for v in a.values()),'subsumer_covered_blocks':sum(bool(v) for v in b.values()),'target_exclusive_blocks':exclusive,'subsumer_exclusive_blocks':[k for k,v in b.items() if v and not a.get(k)]},indent=2)+'\n')
old=subprocess.check_output(['git','show','origin/test-audit/internal-oracle-parser_namespaces:review/test-audit/internal-oracle-parser_namespaces/list.log'],cwd=root,text=True)
oldnames={l for l in old.splitlines() if l.startswith('Test')};names={l for l in (p/'list.log').read_text().splitlines() if l.startswith('Test')}
(dest/'scope.json').write_text(json.dumps({'base':base,'current_top_level_tests':sorted(names),'added_since_audit':sorted(names-oldnames),'removed_since_audit':sorted(oldnames-names),'bounded_matrix_rows':list(matrix['D1']['top_level_results']),'general_agreement_inputs':'six fixtures registered in parser_namespaces_test.go init','review_agreement_inputs':['smoke.a'],'outside_selected_inputs':'unknown','witness_results_not_used_for_production_subsumption':['TestReviewProgramsSelfTest']},indent=2)+'\n')
text='''# Parser namespace family defense

Result: not defended after three production-only attempts. This is a bounded finding and does not authorize a test deletion.

Code under test: Adamic namespace lowering in internal/lower, particularly callableNamespaceUse, namespaceOwnThis and namespaceReadyReads. Oracle: Node runs each original source; both lowered JavaScript and sanitized native output must agree on stdout, stderr and exit; successful originals also require clean leak checks. No oracle, harness or test was changed.

D means TestNativeAgreesWithNode family. Current parser family members are TestParserNamespaceReceiver, TestParserNamespaceClass and TestParserCallableNamespace. They call the same checker with pairs from the same six fixtures registered with D. D compares the same source to the same JavaScript and sanitized native products, checks the same leaks, and additionally compares a release product. The family calls Node before Lower; D calls Lower first. These six originals have no input history or filesystem side effects that give this ordering a distinct expected answer.

Both per-family compiler coverage profiles passed: target covered 2137 blocks, D covered 2137, no exclusive blocks either way. Coverage is only a lead; the whole test bodies and six original sources show no semantic input difference. Therefore the three attempts test different claimed corner cases rather than an invented exclusive input: D1 refuses typeof on callable namespaces; D2 stops exempting ordinary object methods from namespace receiver capture; D3 inverts a readiness guard so initialized namespace member reads panic. Each has a real compiler or runtime behavior change observed in the matrix. Every diff applies to this main base and passed go vet ./internal/lower/; D3 also successfully built native products before the observed runtime failure.

All three were caught by both families. D3 was also caught by TestNamespaceLiveExportBoundary. matrix.json lists all observed passes and failures, including subtests. There were no survivors among these three attempts. Witness/setup rows in the expanded run are recorded, but their production failures would not prove witness strength.

The whole clean package timed out at 90.073 binary seconds without an individual test failure before timeout. We narrowed to the three family members, D on six exact shared inputs, TestModuleNamespaceReadsMatchNode and TestNamespaceLiveExportBoundary. The package grew from 193 to 198 top-level tests; all five additions are in the expanded matrix: FractionalPowersReachRuntime, ReviewProgramsNoLooseFiles, ReviewProgramsRefuse, ReviewProgramsSelfTest, and ReviewProgramsAgreeWithNode (smoke.a only). There are 263 review .a inputs; none contains the word namespace. Remaining review-agreement inputs, other D inputs and other package rows are unknown. The bounded clean runs all passed before mutation; the restored family passes again.

Issues and time costs:
- The brief's unexplained D required resolving the audit's row names and matrix commands. It denotes a partially exercised family, not a test named D.
- The whole package exceeds 90 seconds, so package-wide uniqueness was unavailable. Another observed catcher is sufficient to reject a unique-catch claim for every attempted mutant; unknown rows cannot undo those observed catches.
- /tmp has only 8.8 GB total capacity. Disk-first cleanup removed only the preceding unit's /tmp/defend-unicode-alias scratch/cache. It had 8.8 GB free afterward; /workspace had 20 GB free. The requested 15 GB threshold cannot be met on /tmp, and there was no full-disk baseline failure.
- One review selector was over-escaped and matched no smoke subtest. The first runs are preserved as review-unmatched logs and are not evidence. Corrected replays assert the exact smoke subtest passed.
- Twins and cost exceptions do not apply: these rows are not separate executors or performance checks.

Owner finding: the family's name and assertions match its namespace correctness claim. It actually checks receiver, class and callable behavior against Node; no unasserted performance promise was found. The defense found duplicate inputs and covered behavior, not a vacuous assertion. Do not infer repository-wide redundancy from this bounded three-mutant result.

Warm tool setup was skipped; nproc=5. npm ci ran before baseline. Baseline binary 90.073 s; target coverage 0.909 s; D coverage 0.966 s; expanded clean 3.751 s; smoke clean 0.386 s. commands.json records each go-vet/rebuild/run wall time and cache. Each mutant used its own ADAMIC_BUILD_CACHE_DIR and ADAMIC_GATE_UNCACHED=1. No test was deleted, rewritten or weakened; all production sources were restored before committing evidence.
'''
(dest/'report.md').write_text(text)
print(dest)
