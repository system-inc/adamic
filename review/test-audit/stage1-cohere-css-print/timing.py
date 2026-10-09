import json,pathlib,subprocess,time,re
out=pathlib.Path('review/test-audit/stage1-cohere-css-print'); out.mkdir(parents=True,exist_ok=True)
names=['TestCSSProfileArtifacts','TestClosedPrinterRegexGap','TestCSSPrinterThroughput','TestCSSPrinterBoundaryProofs','TestOptionalBooleanPrinterMatchesGo','TestProduct_CSSPrinterOracle family','TestProduct_CSSPrinterNative family','TestProduct_CSSPrinterDarwinLeaks','TestCSSPrinterShardingCatchesDisagreement','TestCSSProfileSnapshotsAgree','TestSharedSliceAppendAgreesWithNode']
patterns={n:'^'+n+'$' for n in names}
patterns[names[5]]='^TestProduct_CSSPrinter(ParserOracle|Oracle)$'
patterns[names[6]]='^TestProduct_CSSPrinter(SanitizedAndLowered|SemicolonMutant|IndentMutant|WidthMutant)$'
results=[]
for i,n in enumerate(names):
 for k in range(3):
  p=out/f'timing-{i:02d}-{k+1}.log'; start=time.time()
  with p.open('w') as f: code=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run',patterns[n]],stdout=f,stderr=subprocess.STDOUT).returncode
  events=[]
  for line in p.read_text().splitlines():
   try: events.append(json.loads(line))
   except: pass
  secs=next((e.get('Elapsed') for e in reversed(events) if e.get('Action')=='pass' and 'Test' not in e),None)
  skipped=[e.get('Test') for e in events if e.get('Action')=='skip']
  red=[e.get('Test') for e in events if e.get('Action')=='fail' and e.get('Test')]
  results.append(dict(test=n,attempt=k+1,seconds=secs,wall=time.time()-start,code=code,skipped=skipped,red=red,log=str(p),pattern=patterns[n]))
  (out/'timings.json').write_text(json.dumps(results,indent=2))
  if red:
   (out/'STOP-RED.txt').write_text(n+' failed its clean individual baseline; no mutations permitted.\n'); raise SystemExit(1)
