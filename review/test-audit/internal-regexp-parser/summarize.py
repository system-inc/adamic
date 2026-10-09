from pathlib import Path
import json,re,statistics,subprocess,csv
p=Path('review/test-audit/internal-regexp-parser')
scoped=['TestParse','TestFlags','TestQuantifierBounds','TestSharedCanonicalize']
raw=[s for s in (p/'list.log').read_text().splitlines() if s.startswith('Test')]
family=['TestMatcherNodeControls','TestMatcherRandomNode','TestMatcherTest262Executions','TestMatcherCanonicalizeNode','TestMatcherUTF16PatternsNode','TestMatcherOct6Node']
fname='TestMatcherAgreement family'; groups=[fname]+[r for r in raw if r not in family]
(p/'families.json').write_text(json.dumps({fname:family,'separate': [r for r in raw if r not in family],'reason':'Six corpus builders feed the same compareExecutionCases checker. Test262 selects stored expected results, the others run Node. Distinct custom checks, the stateful loop checker and the witness remain separate.'},indent=2)+'\n')
def events(path):
 out=[]
 for line in path.read_text().splitlines():
  try:out.append(json.loads(line))
  except:pass
 return out
med={}
for row in scoped+['TestNodeAgreement','TestMatcherOct6LoopsNode','TestMatcherAgreementFamily']:
 vals=[float(re.search(r'\t([\d.]+)s',f.read_text())[1]) for f in sorted(p.glob('timing-'+row+'-*.log'))]
 assert len(vals)==3,(row,vals)
 med[fname if row=='TestMatcherAgreementFamily' else row]=statistics.median(vals)
(p/'medians.json').write_text(json.dumps(med,indent=2)+'\n')
menu=json.loads((p/'menu.json').read_text()); matrix=[]
for m in menu:
 es=events(p/(m['id']+'.log')); final={e['Test']:e['Action'] for e in es if e.get('Test') in raw and e['Action'] in ['pass','fail','skip']}
 assert set(final)==set(raw),(m['id'],set(raw)-set(final))
 failed=[r for r in raw if final[r]=='fail']
 grouped=([fname] if any(r in failed for r in family) else [])+[r for r in failed if r not in family and r!='TestMatcherOct6Mutants']
 matrix.append(dict(id=m['id'],eligible=not m.get('supplemental',False),raw_results=final,failed_rows=grouped,witness_preconditions=['TestMatcherOct6Mutants'] if 'TestMatcherOct6Mutants' in failed else []))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
with (p/'matrix.csv').open('w') as f:
 w=csv.writer(f);w.writerow(['mutant','eligible']+groups)
 for m in matrix:w.writerow([m['id'],m['eligible']]+[('witness-precondition-fail' if r=='TestMatcherOct6Mutants' and m['witness_preconditions'] else ('fail' if r in m['failed_rows'] else 'pass')) for r in groups])
def fail_line(mid,row):
 es=events(p/(mid+'.log')); candidates=[e['Output'].strip() for e in es if e.get('Test')==row and e.get('Action')=='output' and ('parser_test.go:' in e.get('Output','') or 'properties_test.go:' in e.get('Output',''))]
 return candidates[0]
positive=[]
s=Path('internal/regexp/parser_test.go').read_text()
for a,b in re.findall(r'\{"((?:\\.|[^"\\])*)", "((?:\\.|[^"\\])*)", true\}',s):positive.append({'pattern':json.loads('"'+a+'"'),'flags':json.loads('"'+b+'"')})
assert len(positive)==13,len(positive)
rows=[]
for row in scoped:
 kills=[m['id'] for m in matrix if m['eligible'] and row in m['failed_rows']]
 unique=[m['id'] for m in matrix if m['eligible'] and m['failed_rows']==[row]]
 last=kills[-1];sub=[] if unique else ([fname] if row=='TestParse' else ['TestNodeAgreement'] if row=='TestFlags' else ['TestMatcherOct6LoopsNode'])
 oracle={
 'TestParse':'Handwritten valid/invalid cases. Only error presence is checked, not AST or diagnostic reason; 13 positive cases pass the empty Parse answer.',
 'TestFlags':'Handwritten invalid flags uu, uv and z. Only error presence is checked, not diagnostic reason.',
 'TestQuantifierBounds':'Handwritten AST fields: Min=2, Max=4, Greedy=false for a{2,4}?.',
 'TestSharedCanonicalize':'Internal generated regexp simpleCaseFold and legacyUppercase tables, derived from Unicode 17, decide expected values. Exact numeric agreement across all code points and BMP units; no independent authority value checked. Shared data provenance can conceal common errors.'}[row]
 command=f'ADAMIC_MUTANT={last} ADAMIC_BUILD_CACHE_DIR=/tmp/u074/cache/{last} timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run .'
 r=dict(test=row,package='internal/regexp',file='internal/regexp/'+('properties_test.go' if row=='TestSharedCanonicalize' else 'parser_test.go'),seconds=med[row],oracle=oracle,oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=last+': '+fail_line(last,row),verdict='sacred' if unique else 'subsumed',subsumed_by=sub,mutants_in_matrix=12,probe_kills=['P2','P3'] if row=='TestSharedCanonicalize' else ['P1'],subsumer_seconds=med[sub[0]] if sub else None,vacuous=False,bounded=False,matrix_rows=groups,evidence=command+' > '+str(p/(last+'.log'))+' 2>&1; '+fail_line(last,row),eligible_mutants=11,subsumption_kills=len(kills))
 if row=='TestParse':r.update(vacuous_subcases=positive,supplemental_kills=['M6'])
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2,ensure_ascii=False)+'\n')
phases=dict(setup_seconds=0,nproc=5,npm_ci_seconds=1.089,switch_build_seconds=1.120,matrix_command_wall_seconds=sum(m['wall_seconds'] for m in json.loads((p/'run-times.json').read_text())),matrix_binary_seconds=sum(next(e['Elapsed'] for e in reversed(events(p/(m['id']+'.log'))) if e['Action'] in ['pass','fail'] and not e.get('Test')) for m in matrix),standalone_validation_seconds=sum(m.get('seconds',0) for m in json.loads((p/'diff-checks.json').read_text())),isolated_binary_seconds=sum(float(re.search(r'\t([\d.]+)s',f.read_text())[1]) for f in p.glob('timing-*.log')),clean_baseline_binary_seconds=1.881,final_clean_binary_seconds=1.654)
(p/'phases.json').write_text(json.dumps(phases,indent=2)+'\n')
summary=['u074 completed at origin/main 09fe4b54913753188a9357982bfd47cdf36ef97c; all four names remain in their listed files.','Clean baseline: 1.881 s; final restored run: 1.654 s; nproc: 5; no skips.','Verdicts: one sacred, three subsumed; no whole-row vacuity, but 13 vacuous positive parser cases.','Matrix: complete package, 17 top-level names grouped into 12 rows; 11 menu mutants plus supplemental M6; three probes.','Evidence: test-audit/internal-regexp-parser, review/test-audit/internal-regexp-parser/.']
text='\n'.join(summary)+'\n\n```json\n'+json.dumps(rows,indent=2,ensure_ascii=False)+'\n```\n\nMutants, all locations at starting origin/main. Abbreviations: F = TestMatcherAgreement family, N = TestNodeAgreement, P = TestParse, G = TestFlags, Q = TestQuantifierBounds, S = TestSharedCanonicalize, L = TestMatcherOct6LoopsNode. Full names and raw member results are in matrix.json.\n\n| id | file:line | change | failed production rows |\n|---|---|---|---|\n'
short={fname:'F','TestNodeAgreement':'N','TestParse':'P','TestFlags':'G','TestQuantifierBounds':'Q','TestSharedCanonicalize':'S','TestMatcherOct6LoopsNode':'L'}
for m,x in zip(menu,matrix):
 text+='| '+m['id']+(' supplemental' if m.get('supplemental') else '')+' | '+m['file']+':'+str(m['line'])+' | '+m['kind']+' | '+', '.join(short.get(r,r) for r in x['failed_rows'])+' |\n'
text+='\nM4 also breaks a control precondition in TestMatcherOct6Mutants. This witness failure is excluded from kills and subsumption. Scoped rows do not form a family: their assertions differ. The six matcher corpus builders listed in families.json form one comparison family, including its stored-expectation test262 member. Family timing runs all six together, three times. Subsumption is a hint based on P: three kills, G: two kills and Q: one kill, not a deletion recommendation.\n\nSurvivor M7: '+(p/'survivor-control.log').read_text().strip()+'; mutant: '+(p/'survivor-M7.log').read_text().strip()+'. Every package row passes. This is changed, unguarded admission behavior in this matrix, not an equivalent candidate. Witness command: ADAMIC_MUTANT=M7 go run ./review/test-audit/internal-regexp-parser/survivor-witness.go (copy survivor-witness.go.txt first).\n\nProbes: P1 replaces Parse with return nil,nil; P2 and P3 replace the two canonicalization entries with return 0. All intended rows fail. P1 TestQuantifierBounds panics at parser_test.go:47; each intended row was already run separately, so no other result is inferred from the aborted binary. TestParse has 13 positive cases that pass P1 and 15 negative cases that fail; its vacuous_subcases lists every positive case. Probe failures are excluded from mutant kills.\n\nBrief ambiguities, errors and costs:\n\n- The historical commit 8de93800f4 is not current origin/main. Fetch selected 09fe4b5491; no scoped name moved or vanished. Every reported production line and diff uses that starting commit.\n- npm ci is required in stage3/api even though these tests execute plain Node and load no node_modules. It took 1.089 s. The first attempt failed because /usr/bin/time was missing, so I used the shell timer and repeated the baseline after successful installation.\n- The first planting script stopped before mutation because a guard substring occurs in two functions. I restricted the edit to the first function. A subsequent gofmt lookup failed because env.sh was sourced too late; sourcing it before the script corrected this. Neither attempt supplied a mutant result.\n- TestSharedCanonicalize compares two internal implementations. I explicitly chose the called unicodeproperties functions as code under test and left the generated regexp tables unchanged as the self oracle. This avoids mutating the expected-value source. Unicode provenance does not independently prove this comparison correct.\n- The menu says flip a condition or drop a statement, but M6 removes two condition operands. I conservatively labeled it supplemental after checking menu compliance, excluded it from kills, unique kills and verdicts, and retained its log and diff. Its initial choice was made before outcomes.\n- M2 uses a false-condition selector in scratch; its standalone diff drops the whole exclusion if statement. They are behaviorally equivalent. The standalone diff passes go vet.\n- Family boundaries are broader than filenames or top-level names. Six tests build different inputs for compareExecutionCases, so uniqueness and subsumption use their family row. Tests with additional custom checks, stateful loops and the witness remain separate.\n- The full-package run is only about two seconds, so narrowing was unnecessary. bounded=false means all package rows were observed; it does not claim repo-wide uniqueness.\n- A production failure of a built-in mutant witness can be a broken control precondition. M4 does that; its witness row is excluded rather than treated as another production kill. The witness itself is outside the requested four-row audit, so I did not weaken its checker.\n- Nil is the empty Parse answer, which crashes the AST-inspecting row. Probes were run per intended row to retain complete observations. Standalone probe diffs replace full bodies so go vet sees no unreachable statements.\n- The budget language about four compiler mutants concerns native rebuilds. These rows run Go parser/folding code; no native product is built. Twelve selector choices need one Go test-binary build, with separate ADAMIC_BUILD_CACHE_DIR values on all matrix runs.\n- Three isolated measurements were taken for each scoped row and all candidate subsumers. The family needed an additional three combined runs; reporting a fast member as the family cost would understate the cost.\n\nTiming, seconds: '+json.dumps(phases)+'. See phase and per-mutant timing JSON for separation of command wall time and binary time. Full-session wall time was not separately instrumented. No step exceeded 90 s. No runtime C/native build, opt-in dependency or skipped row occurred.\n\nLimits: four rows receive verdicts; all package rows supply matrix context. No other package was tested, no repo-wide uniqueness was assessed, no independent Unicode/spec value was checked, and no new mutant was added after outcomes. Production sources were restored and the package passed. Standalone Go mutant/probe diffs all apply and pass go vet. No PR or main push.\n'
(p/'report.md').write_text(text)
