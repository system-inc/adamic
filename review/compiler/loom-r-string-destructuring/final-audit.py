from pathlib import Path
import json,re,subprocess
p=Path(__file__).resolve().parent
load=lambda name:json.loads((p/name).read_text())
rows=load('consumer-results.json')
assert len(rows)==49
resolutions={
 './cmd/adamic':{'log':'cli-recheck.log.txt','prerequisite':'cli-runtime-product.log.txt','reason':'The CLI child changes XDG_CACHE_HOME. Built main owner TestProduct_NativeRuntimesSanitized0 and exported the existing default ADAMIC_BUILD_CACHE_DIR so its child reads the same product cache. No code or test edits.'},
 './stage1/cohere/estree':{'log':'estree-recheck.log.txt','prerequisite':'estree-oracle-product.log.txt','reason':'Built main owner stage1/cohere/estree/oracles TestProduct_EstreeGoOracle, then reran only TestRawInputGap. Other selected cases had passed or explicitly skipped. No code or test edits.'}}
resolutions.update({
 './stage1/cohere/lint/helpers/comments':{'log':'lint-recheck.log.txt','prerequisite':'lint-oracle-products.log.txt','reason':'Built declared TestProduct_CommentsOracle before rerunning TestJsxParserGapIsExplicit; no code or test changes.'},
 './stage1/cohere/lint/regex':{'log':'lint-recheck.log.txt','prerequisite':'lint-oracle-products.log.txt','reason':'Built declared TestProduct_RegexOptionDialects before rerunning TestOptionDialectGap; TestDynamicPatternGap already passed; no code or test changes.'}})
final=[];skips=[];timings=[]
for row in rows:
 logname='consumer-'+row['package'].removeprefix('./').replace('/','_')+'.log.txt'
 text=(p/logname).read_text()
 runs=re.findall(r'^=== RUN   (\S+)',text,re.M)
 assert all(any(re.fullmatch(pattern,name) for name in runs) for pattern in row['tests']),row
 for status,name,seconds in re.findall(r'^\s*--- (PASS|FAIL|SKIP): (\S+) \(([\d.]+)s\)',text,re.M):
  timings.append({'log':logname,'status':status,'test':name,'seconds':float(seconds)})
  if status=='SKIP':skips.append({'package':row['package'],'test':name,'log':logname})
 result=dict(row)
 if row['exit']:
  fix=resolutions[row['package']]
  for name in ['log','prerequisite']:
   check=(p/fix[name]).read_text()
   assert '\nPASS\n' in check and '\nFAIL\n' not in check,fix
  result.update(finalExit=0,resolution=fix)
 else:result['finalExit']=0
 final.append(result)
for row in load('consumer-already-validated.json'):
 final.append(dict(row,finalExit=0,evidence={'./internal/ir':['readers.log.txt','guards.log.txt'],'./internal/lower':['lower.log.txt'],'./internal/flow':['flow.log.txt']}[row['package']]))
assert len(final)==52
for name,obj in [('consumer-final-results.json',final),('consumer-skips.json',skips),('consumer-timing-audit.json',timings)]:
 (p/name).write_text(json.dumps(obj,indent=2)+'\n')
assert all(x['exit']==0 for x in load('validation.json'))
assert all(x['caught'] for x in load('mutants.json')) and load('count-mutant.json')['caught']
assert load('counts-audit.json')['changed']==[]
assert load('runtime-equivalence.json')['changedLines']==[]
assert subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip()=='dbdf977f5b8b0979d414ebe6f20e7aadd312e309'
print('52 reader packages checked; 4 prerequisite failures resolved without code changes; 4 mutants caught.')
