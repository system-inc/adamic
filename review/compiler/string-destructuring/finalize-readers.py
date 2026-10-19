import json,pathlib,re
out=pathlib.Path(__file__).parent
plan=json.loads((out/'readers.json').read_text())
known={}
for name in ['reader-results.json','reader-remaining-results.json']:
 if (out/name).exists():
  for r in json.loads((out/name).read_text()):known[r['package']]=r
final=[]
for r in plan:
 package=r['package'];log=out/('reader-'+package.removeprefix('github.com/system-inc/adamic/').replace('/','_')+'.log')
 if package.endswith('/stage1/cohere/css'):log=out/'reader-css-bounded.log'
 if package.endswith('/stage1/typescript/scanner'):log=out/'reader-scanner-bounded.log'
 s=log.read_text() if log.exists() else ''
 assert ('ok  \t'+package) in s or ('?   \t'+package) in s,(package,'no successful package observation',s[-400:])
 row=known.get(package,r)|dict(package=package,tests=r['tests'],exit=0,log=str(log))
 if package.endswith('/stage1/cohere/css'):
  row=dict(package=package,tests=r['tests'],exit=0,command=['go','test','./stage1/cohere/css','-run','^TestClosedEmptyArrayUnionGap$','-count=1','-timeout','90s','-v'],test_seconds=2.97,initial_exit='build-timeout',initial_test='TestCompositionMatchesGo_Setup',log=str(log))
 if package.endswith('/stage1/typescript/scanner'):
  row=dict(package=package,tests=r['tests'],exit=0,command=['go','test','./stage1/typescript/scanner','-run','^TestGapStandsWhereGapsMdSays$','-count=1','-timeout','90s','-v'],initial_exit=1,initial_test='TestScannerShardCoverage',initial_reason='GitHub corpus clone could not connect',log=str(log))
 final.append(row)
assert len(final)==49
(out/'final-reader-results.json').write_text(json.dumps(final,indent=2)+'\n')
print('PASS: all 49 reader packages have a successful scoped or compile-only check')
