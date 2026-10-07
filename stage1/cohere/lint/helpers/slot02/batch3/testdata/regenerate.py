# Capture workflow adapted from slot 03; overlays never edit the cohere worktree.
# Capture actual runtime source inputs, including dynamically assembled fixture strings.
import json, pathlib, subprocess, tempfile, os, gzip, re
here=pathlib.Path(__file__).resolve().parent
root=here.parents[6]
cohere=root/'cohere'
ledger=json.loads((here.parents[2]/'readiness.json').read_text())['remaining']
symbols=['react.IsEs5ComponentCall','react.isCreateClassName','tailwind/collapse.isMathFunctionName']
selected=sorted({r['rule'] for r in ledger if any(h.endswith('/'+s) for h in r['remaining_helpers'] for s in symbols)})
locations=[json.loads(l) for l in (root/'stage1/cohere/lint/inventory/test-cases.jsonl').read_text().splitlines()]
packages=sorted({str(pathlib.PurePosixPath(r['location'].split(':')[0]).parent).replace('cohere/','./',1) for r in locations if r['rule'] in selected})
with tempfile.TemporaryDirectory(prefix='slot02Batch3-') as scratch:
 scratch=pathlib.Path(scratch)
 original=cohere/'internal/lint/testing/rule_testing.go'
 modified=scratch/'rule_testing.go'
 text=original.read_text();anchor='func RunWithOptions(t *testing.T, subject rule.Rule, fileName string, sourceText string, options any) Result {'
 assert text.count(anchor)==1
 modified.write_text(text.replace(anchor,anchor+'\n slot02Batch3Capture(subject.Name,fileName,sourceText)',1))
 replace={str(original):str(modified),str(cohere/'internal/lint/testing/adamic_slot02Batch3.go'):str(here/'capture.go')}

 # Typed fixtures and live-engine wrappers are captured before external-engine skips.
 original=cohere/'internal/lint/testing/program.go'
 text=original.read_text()
 match=re.search(r'func runTypedFiles\(.*?\) Result \{',text,re.S);assert match
 text=text[:match.end()]+"\n for file,source:=range files {slot02Batch3Capture(subject.Name,file,source)}"+text[match.end():]
 modified=scratch/'program.go';modified.write_text(text);replace[str(original)]=str(modified)
 for file,functions in {
  'enforce_consistent_variant_order_test.go':[('runVariantOrderFixture','EnforceConsistentVariantOrder','fileName')],
  'no_unknown_classes_test.go':[('runUnknownFixtureWithOptions','NoUnknownClasses','fileName')],
  'no_conflicting_classes_test.go':[('runConflictFixtureWithOptions','NoConflictingClasses','fileName')],
  'enforce_canonical_classes_test.go':[('runCanonicalFixtureWithOptions','EnforceCanonicalClasses','fileName')],
  'enforce_consistent_class_order_test.go':[('runClassOrderFixture','EnforceConsistentClassOrder','fileName')],
  'enforce_consistent_class_order_options_test.go':[('runClassOrderFixtureWithOptions','EnforceConsistentClassOrder','"Component.tsx"')],
 }.items():
  original=cohere/'internal/lint/rules/tailwind'/file;text=original.read_text()
  for function,rule,file_expression in functions:
   match=re.search(r'func '+function+r'\(.*?\) rule_testing.Result \{',text,re.S);assert match,function
   text=text[:match.end()]+f'\n rule_testing.AdamicSlot02Batch3Capture({rule}.Name,{file_expression},source)'+text[match.end():]
  modified=scratch/file;modified.write_text(text);replace[str(original)]=str(modified)
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 capture=scratch/'sources.jsonl';capture.write_text('')
 env=dict(os.environ,ADAMIC_SLOT02_BATCH3_RULES='\n'.join(selected),ADAMIC_SLOT02_BATCH3_CAPTURE=str(capture))
 log=here.parent/'evidence/capture.log'
 with log.open('w') as output:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'-count=1','-timeout=20m',*packages],cwd=cohere,env=env,stdout=output,stderr=subprocess.STDOUT)
 if result.returncode:
  failures=set(re.findall(r'^--- FAIL: (\S+)',log.read_text(),re.M))
  known={'TestConflictFixturesActuallyRan','TestUnknownClassFixturesActuallyRan','TestCanonicalClassesPlacementIsAccountedFor','TestClassOrderFixturesActuallyRan','TestCanonicalFixturesActuallyRan','TestConflictingClassesPlacementIsAccountedFor','TestUnknownClassesPlacesEveryCorpusClass','TestClassOrderLiveMatchesTheEngineOverTheCorpus'}
  assert failures and failures<=known,log.read_text()
  print('Capture completed with known external Tailwind/corpus failures; this is not a passing Go rule-package gate.')
 rows=[json.loads(l) for l in capture.read_text().splitlines()]
 observed={r['rule'] for r in rows};assert observed==set(selected),(set(selected)-observed)
 rows=sorted({json.dumps(r,sort_keys=True,ensure_ascii=False) for r in rows})
 with gzip.GzipFile(filename='',mode='wb',fileobj=(here/'sources.jsonl.gz').open('wb'),mtime=0) as out:
  out.write(('\n'.join(rows)+'\n').encode())
 coverage={'symbols':symbols,'rules':selected,'sources':len(rows),'packages':packages,'capture_gate_exit':result.returncode,'cohere_commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=cohere,text=True).strip()}
 (here/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
 print(json.dumps(coverage,indent=2))
