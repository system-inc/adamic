"""Capture every nextjs helper invocation from all consuming upstream tests.
Overlay renames bodies but does not alter their decisions. No saved Go answers are
fed to production helpers. Controls cover branches absent from upstream cases.
"""
import json, os, subprocess, tempfile, sys
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[5]
COHERE=ROOT/'cohere'
OUTPUT=Path(sys.argv[1]) if len(sys.argv)>1 else HERE
OUTPUT.mkdir(parents=True,exist_ok=True)
NAMES=['IsDocumentFile','lastSeparator','splitPath','IsDocumentPage','IsInApplicationDirectory','IsInPagesDirectory','splitSegments','RouteContractExports','buildRouteFileContracts']
with tempfile.TemporaryDirectory() as temp:
 temp=Path(temp); replacements={}
 for file in ['paths.go','contract.go']:
  original=COHERE/'internal/lint/ecmascript/nextjs'/file
  source=original.read_text()
  for name in NAMES:
   source=source.replace('func '+name+'(', 'func adamicRaw'+name+'(')
  side=temp/file;side.write_text(source);replacements[str(original)]=str(side)
 virtual=COHERE/'internal/lint/ecmascript/nextjs/adamic_capture.go'
 replacements[str(virtual)]=str(HERE/'capture.go')
 virtualTest=COHERE/'internal/lint/ecmascript/nextjs/adamic_controls_test.go'
 replacements[str(virtualTest)]=str(HERE/'controls.go')
 harness=COHERE/'internal/lint/testing/rule_testing.go'
 original='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
 source=harness.read_text();assert source.count(original)==1
 side=temp/'rule_testing.go';side.write_text(source.replace(original,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
 replacements[str(harness)]=str(side)
 overlay=temp/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
 cases=temp/'cases.jsonl';env=os.environ|{'ADAMIC_NEXTJS_CAPTURE':str(cases),'COHERE_DOCS_CAPTURE':str(temp/'asserted')}
 patterns={'next':'NoBeforeInteractiveScriptOutsideDocument|NoDocumentImportInPage|NoHeadElement|NoHeadImportInDocument|NoPageCustomFont|NoStyledJsxInDocument|NoTypos','structure':'NextNoNearMissRouteExport','nexus':'ConsistencyNoAbbreviatedIdentifier|ConsistencyRequireConstantCasing'}
 for family,pattern in patterns.items():
  subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+family,'-run','^Test('+pattern+')','-count=1','-timeout=10m'],cwd=COHERE,env=env,stderr=subprocess.STDOUT,check=True)
 consumerRows=[json.loads(x) for x in cases.read_text().split('\n') if x]
 ledger=json.loads((HERE.parents[1]/'comments/readiness.json').read_text())
 consumers={r['rule'] for r in ledger['remaining'] if any('/nextjs.' in h for h in r['remaining_helpers'])}
 consumers.update(['nexus/consistency-no-abbreviated-identifier','nexus/consistency-require-constant-casing'])
 asserted={name:set() for name in consumers}
 for file in (temp/'asserted').glob('*.jsonl'):
  for line in file.read_text().split('\n'):
   if not line: continue
   row=json.loads(line)
   if row['rule'] in asserted: asserted[row['rule']].add(json.dumps([row['file'],row['source'],row.get('options')],sort_keys=True))
 upstreamCounts={name:len(rows) for name,rows in asserted.items()}
 assert all(upstreamCounts.values()),upstreamCounts
 subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/ecmascript/nextjs','-count=1'],cwd=COHERE,env=env,stderr=subprocess.STDOUT,check=True)
 rows=[json.loads(x) for x in cases.read_text().split('\n') if x]
 consumerCounts={name:sum(r['Kind']==name for r in consumerRows) for name in NAMES}
 counts={name:sum(r['Kind']==name for r in rows) for name in NAMES}
 assert all(counts.values()),counts
 (OUTPUT/'cases.json').write_text(json.dumps(rows,ensure_ascii=True)+'\n')
 (OUTPUT/'counts.json').write_text(json.dumps({'consumerCalls':consumerCounts,'allCalls':counts,'upstreamCases':upstreamCounts},indent=2)+'\n')
 print(json.dumps({'consumerCalls':consumerCounts,'allCalls':counts,'upstreamCases':upstreamCounts}))
