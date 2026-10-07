# Capture actual runtime source inputs, including dynamically assembled fixture strings.
import json, pathlib, subprocess, tempfile, os, gzip, re
here=pathlib.Path(__file__).resolve().parent
root=here.parents[6]
cohere=root/'cohere'
ledger=json.loads((here.parents[2]/'readiness.json').read_text())['remaining']
symbols=['control_flow_graph.*Builder[E].pushJump', 'control_flow_graph.*Builder[E].popJump', 'control_flow_graph.*Builder[E].makeUnreachable']
selected=sorted({r['rule'] for r in ledger if any(h.endswith('/'+s) for h in r['remaining_helpers'] for s in symbols)})
locations=[json.loads(l) for l in (root/'stage1/cohere/lint/inventory/test-cases.jsonl').read_text().splitlines()]
packages=sorted({str(pathlib.PurePosixPath(r['location'].split(':')[0]).parent).replace('cohere/','./',1) for r in locations if r['rule'] in selected})
with tempfile.TemporaryDirectory(prefix='slot03-') as scratch:
 scratch=pathlib.Path(scratch)
 original=cohere/'internal/lint/testing/rule_testing.go'
 modified=scratch/'rule_testing.go'
 text=original.read_text();anchor='func RunWithOptions(t *testing.T, subject rule.Rule, fileName string, sourceText string, options any) Result {'
 assert text.count(anchor)==1
 modified.write_text(text.replace(anchor,anchor+'\n slot03Capture(subject.Name,fileName,sourceText)',1))
 replace={str(original):str(modified),str(cohere/'internal/lint/testing/adamic_slot03.go'):str(here.parents[1]/'testdata/capture.go')}

 # Typed fixtures and live-engine wrappers are captured before external-engine skips.
 original=cohere/'internal/lint/testing/program.go'
 text=original.read_text()
 match=re.search(r'func runTypedFiles\(.*?\) Result \{',text,re.S);assert match
 text=text[:match.end()]+"\n for file,source:=range files {slot03Capture(subject.Name,file,source)}"+text[match.end():]
 modified=scratch/'program.go';modified.write_text(text);replace[str(original)]=str(modified)
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':replace}))
 capture=scratch/'sources.jsonl';capture.write_text('')
 env=dict(os.environ,ADAMIC_SLOT03_RULES='\n'.join(selected),ADAMIC_SLOT03_CAPTURE=str(capture))
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
