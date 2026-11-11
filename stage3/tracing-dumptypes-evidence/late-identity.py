from pathlib import Path
import hashlib,json,subprocess,shutil
S=Path(__file__).resolve().parent;rows=json.loads((S/'stock-late-projects.json').read_text());results=[]
lib=S/'stock-late';instrumented=(lib/'_tsc.js').read_bytes();shutil.copyfile('/workspace/adamic-tools/typescript/lib/_tsc.js',lib/'_tsc.js')
try:
 for i,row in enumerate(rows):
  trace=S/('stock-late-trace-'+str(i))/'types.json';a=trace.read_bytes()
  with (S/('late-identity-project-'+str(i)+'.log')).open('w') as log:p=subprocess.run(row['command'],stdout=log,stderr=subprocess.STDOUT)
  b=trace.read_bytes();r=dict(project=row['project'],command=row['command'],exit=p.returncode,instrumentedExit=row['exit'],equal=a==b,bytes=len(a),sha256=hashlib.sha256(a).hexdigest());results.append(r);print(json.dumps(r),flush=True)
  assert a==b and p.returncode==row['exit']
finally:(lib/'_tsc.js').write_bytes(instrumented)
(S/'late-instrumentation-identity-results.json').write_text(json.dumps(results,indent=2)+'\n')
