import pathlib,json,subprocess,statistics,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/internal-lower-library_language'
rows=['TestLibraryLanguageBoundaries','TestLibraryMapSetGapsStayRefused','TestLibraryMapSetIteratorCopyTypesRefused','TestLibraryMethodValues','TestLibraryMethodValueBoundaries','TestLibraryMethodValueSafety','TestNodeBufferRefusals']
timings={};start=time.monotonic()
for r in rows:
 vals=[]
 for n in range(3):
  f=out/(r+'-timing-'+str(n+1)+'.log')
  with f.open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+r+'$'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)
  for l in f.read_text().splitlines():
   try:e=json.loads(l)
   except:continue
   if e.get('Action')=='pass' and not e.get('Test'):vals.append(e['Elapsed'])
 timings[r]=dict(runs=vals,median=statistics.median(vals));print(r,timings[r],flush=True)
(out/'timings.json').write_text(json.dumps(timings,indent=2));(out/'timing-wall-seconds.json').write_text(json.dumps(time.monotonic()-start))
