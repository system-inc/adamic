import pathlib,json,subprocess,re
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-graphql-printer-grain_mutant/session'
def events(name):
 return [json.loads(x) for x in (p/(name+'.log')).read_text().splitlines() if x.startswith('{')]
base=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip();matrix=json.loads((p/'matrix.json').read_text());assert len(matrix)==3
for m in matrix:
 assert m['runs'][-1]['exit']==0 and not m['runs'][-1]['rows_failed']
 subprocess.run(['git','apply','--check',str(p/(m['mutant']+'.diff'))],cwd=root,check=True)
 commands=[]
 for r in m['runs']:
  commands.append('ADAMIC_GRAPHQL_PRETTIER=/tmp/printer-defense-prettier ADAMIC_GRAPHQL_PRINTER_BENCH=1 ADAMIC_BUILD_CACHE_DIR='+r['cache']+' '+' '.join(r['command'])+' > '+r['log']+' 2>&1')
 m['commands']=commands
row=dict(test='TestPrinterThroughput',package='stage1/cohere/graphql/printer',prior_verdict='subsumed',subsumed_by=['TestPrinterAsGoCohere family'],defense='not defended',unique_mutant=None,attempts=[dict(mutant=m['mutant'],file_line=m['file_line'],change=m['change'],rows_failed=m['runs'][-1]['rows_failed']) for m in matrix],evidence='D1/D2/D3: ADAMIC_GRAPHQL_PRETTIER=/tmp/printer-defense-prettier ADAMIC_GRAPHQL_PRINTER_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/printer-defense/cache/<id> timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/graphql/printer/ -run . > <id>.log 2>&1; all passed, no failing line. Commands and passed rows: matrix.json.')
(p/'rows.json').write_text(json.dumps([row],indent=2));(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
meta={}
for name in ['baseline','baseline-warm','throughput-coverage','asgo-coverage','D1','D2','D3']:
 ev=events(name);meta[name]=dict(package_events=[e for e in ev if e['Action'] in ['pass','fail'] and not e.get('Test')],throughput_events=[e for e in ev if e['Action'] in ['pass','fail'] and e.get('Test')=='TestPrinterThroughput'],skips=[e['Test'] for e in ev if e['Action']=='skip'],build_lines=[e['Output'].strip() for e in ev if 'clang wall' in e.get('Output','') or 'C emission wall' in e.get('Output','') or 'lowering wall' in e.get('Output','')],timing_lines=[e['Output'].strip() for e in ev if 'texts/s' in e.get('Output','')])
(p/'validation-and-timing.json').write_text(json.dumps(dict(base=base,nproc=5,standalone_diffs_apply=True,source_restored=True,runs=meta),indent=2))
audit=json.loads((p/'audit-scope.json').read_text());current=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')];old=audit['discovered']
(p/'scope-difference.json').write_text(json.dumps(dict(current=current,added=sorted(set(current)-set(old)),removed=sorted(set(old)-set(current))),indent=2))
report=f'''TestPrinterThroughput was not defended by three output-preserving cost mutations.
All three whole-package matrices passed; no unique failure was observed.
No test was changed; evidence is based on origin/main {base}.

CODE UNDER TEST: GraphQL printer port main.ts, printer.ts, doc.ts and composed parser/lexer and Unicode-width ports. ORACLE: unchanged Go cohere expected bytes; the throughput row also executes Go cohere and Prettier 3.9.6 and checks their output. Node executes the port rather than furnishing an independent oracle.

Read audit rows.json, code-and-oracle.txt, scope, menu, matrix, experiments and replay validation. The audit's report.py assembles rows.json. There was no REPORT.md at the supplied path. Its four mutations changed formatted answers rather than testing cost. Archived report files are prefixed audit-.

Coverage: separate subject and whole subsumer-family profiles used -coverpkg on internal/lower and internal/native. Go cannot instrument TypeScript. Supplementary NODE_V8_COVERAGE records the actual port. coverage-difference.json lists the only exclusive Go block, native.go:106 (-O2 for unsanitized build). V8 finds no exclusive doc.ts or printer.ts bytes; main.ts's default-mode fallback is exclusive (origin line 43, transformed JS line 44), but picks settings equal to explicit defaults. Saved transformed sources define the V8 offset/line reference. The cost distinction is six source/native runs over successful defaults cases, using a release native product, versus the family's four option modes and sanitized execution.

Attempts aimed at that repeated workload:
D1 main.ts:7 changes the four escaping fast-path conditions to true. Three previously skipped split/join passes execute on the fixed probe, with unchanged stdout.
D2 main.ts:22 drops the no-backslash early return. A plain input now splits, constructs parts and joins instead of returning the string. stdout is unchanged.
D3 doc.ts:105 removes the cached width read at both fits subtraction sites and recomputes stringWidth(node.text). V8 confirms stringWidth calls rise from 105 to 121 on query{{hello(a:1,b:2)}} while stdout is identical. This uses the brief's explicit cost-mutant permission to drop a cache, a broader operation than its literal six-item expression menu. No oracle, harness or test was changed.

All native products rebuilt under each mutant's own /tmp/printer-defense/cache/<id>. The full matrix includes every currently discovered top-level test, including newer rows if any; passed rows are listed per mutant in matrix.json. The only skip is TestProduct_GraphQLPrinterRelease, explicitly Darwin-only. No package uniqueness is claimed from a skip.

Finding for owner: performance_test.go:86-94 checks exit status, stderr and exact output; :99-109 logs timings and best texts/s. It asserts no absolute throughput, relative speed or regression threshold. The name promises a measurement, which it performs, but it is not a performance regression gate. The generic 90s binary limit is not a declared benchmark threshold. The three mutants demonstrate preserved answers with additional work and no rejection. This bounded mutation search is not proof that every possible compiler or port mutation is covered elsewhere, nor a recommendation to delete the test.

Costs and limitations: warm env.sh worked; setup skipped. nproc=5. npm ci in stage3/api reported 464ms; scratch Prettier install reported 458ms. Cold full baseline timed out at 90.026s with no ordinary assertion error, largely after 45.71s oracle build; target had passed. Warm full baseline and both per-row/family coverage runs passed. The cold baseline is preserved, not counted as a mutant kill. Coverage ran concurrently with the warm baseline, so those times are not isolated benchmark comparisons. Mutant runs were sequential. Raw build and test timing evidence is in validation-and-timing.json. No broader packages or Darwin execution were tested.

Brief friction: Go -coverpkg alone measures the compiler, not this TS port, requiring supplementary V8 instrumentation and transformed-source mapping. The audit did not contain REPORT.md, and the supplied failing evidence was truncated. Cold prerequisite work consumed the baseline deadline despite warm tool binaries. The fixed menu omits the separately mandated cache-removal operation. The schema has no bounded or skipped-rows field, so those qualifications are recorded here and in matrix metadata. No rejection or user approval blocked the work.
'''
(p/'report.md').write_text(report)
print(json.dumps([row],indent=2))
