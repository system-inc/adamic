import pathlib,json,subprocess
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix');branch='origin/test-audit/internal-native-radix:review/test-audit/internal-native-radix/'
rows=json.loads(subprocess.check_output(['git','show',branch+'rows.json']))
wanted=['TestToStringWithARadixOutOfRangePanics','TestRecordBenchmark','TestRegExpSearchNode','TestRegExpLintPatternsNode','TestRegExpBytecodeTest262','TestRegExpNativeStepLimit','TestRegExpBytecodePatternUnits','TestRuntimeReleasePaths','TestRuntimeStringEquality','TestRegExpBytecodeRandomNode family']
rows=[r for r in rows if r['test'] in wanted];(p/'prior-rows.json').write_text(json.dumps(rows,indent=2))
for f in ['REPORT.md','friction-and-limits.md','mutant-table.md','scope.json','mutant-plan.json']:(p/('prior-'+f)).write_bytes(subprocess.check_output(['git','show',branch+f]))
def covered(row):
 f=p/(row.replace(' ','_')+'.cover');out=set()
 for l in f.read_text().splitlines()[1:]:
  block,_,count=l.rsplit(' ',2)
  if int(count)>0:
   file,span=block.rsplit(':',1);start,end=span.split(',');a=int(start.split('.')[0]);b=int(end.split('.')[0]);out.update((file,n) for n in range(a,b+1))
 return out
out=[]
for r in rows:
 a=covered(r['test']);b=covered(r['subsumed_by'][0]);out.append({'test':r['test'],'subsumer':r['subsumed_by'][0],'exclusive_go_lines':[f+':'+str(n) for f,n in sorted(a-b)],'shared_go_lines':len(a&b),'limitation':'Go coverage cannot instrument runtime C compiled in child clang processes. Exclusive Go lines are compiler/build leads, not runtime execution proof.'})
(p/'coverage-differences.json').write_text(json.dumps(out,indent=2));print(json.dumps(out,indent=2))
notes='''CODE UNDER TEST and ORACLE, before mutation:
- Radix row: runtime/radix.c adamic_number_to_radix, with live Node output, RangeError classification, native exit70 and diagnostic prefix.
- Benchmark: runtime record/map/string/heap behavior at sizes1000,10000,100000,1000000, unsanitized, five rounds. Live Node work checksum; hand-written timing field validation. There is no performance threshold.
- Regexp Search, Lint and Random family: runtime regexp VM and internal/regexp compilation/native bytecode, with live Node capture spans, groups and lastIndex. Node is the oracle, never mutated.
- BytecodeTest262: same compiler/runtime with recorded outside-authority test262 observations; no live test262 runner.
- StepLimit: runtime VM step budget, hand-written exit70 and diagnostic requirement. New boundary test checks exact instruction budgets independently.
- PatternUnits: regexp UTF16 pattern compiler and native VM, hand-written exact [0,1] spans for two different lone surrogates; does not use JSON-decoded pattern text as its authority.
- ReleasePaths: heap release/draining on NULL, immortal, shared strings and 100000 linked objects; live allocation counts and fixed output plus sanitizers.
- StringEquality: runtime equal for same header, separate equal headers, unequal strings and undefined; live Node strict equality.
No oracle, harness or test source was changed. Construction coverage is measured with Go coverpkg; C coverage remains unavailable. Semantic input differences are documented per mutation in mutants.json.
'''
(p/'code-and-oracle.md').write_text(notes)
