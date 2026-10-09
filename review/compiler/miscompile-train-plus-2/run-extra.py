from pathlib import Path
import subprocess,re,json,difflib
root=Path.cwd(); out=root/'review/compiler/miscompile-train-plus-2/resolved/extra-mutants';out.mkdir(exist_ok=True)
items=[('constructor-revert','review/compiler/fx6-candidates-3/constructor-revert.patch','oracle','TestNativeAgreesWithNode/internal/oracle/testdata/(nbody_static_collision|field_write_paths|class_features_static|class_features_static_private).a'),('reject-tuple-objects','review/compiler/fx7-tuple-recognition/reject-tuple-objects.diff','lower','^TestTupleRecognitionP68$'),('missing-tuple-marker','review/compiler/fx7-tuple-recognition/missing-tuple-marker.diff','lower','^TestTupleRecognitionP68$'),('admit-arbitrary-arrays','review/compiler/miscompile-train-plus-2/admit-arbitrary-arrays.patch','lower','^TestTupleRecognitionRejectsArrayObjectView$'),('clear-contract','review/compiler/fx7-callable-contract/clear-contract.diff','lower','^TestCallableContractChecksBeforeArguments$'),('source-argument-layout','review/compiler/fx7-callable-contract/source-argument-layout.diff','lower','^TestCallableContractWideProducer$'),('missing-producer-mask','review/compiler/fx7-callable-contract/missing-producer-mask.diff','lower','^TestCallableContractWideProducer$'),('spread-refusal','review/compiler/fx7-callable-contract/spread-refusal.diff','lower','^TestCallableContractSpreadRefused$')]
results=[]
for name,patch,pkg,test in items:
 text=Path(patch).read_text(); paths=re.findall(r'^\+\+\+ b/(.*)$',text,re.M);originals={p:Path(p).read_bytes() for p in paths}
 try:
  applied=subprocess.run(['git','apply','--unidiff-zero',patch],capture_output=True,text=True,timeout=10)
  if applied.returncode: raise RuntimeError(name+': patch failed '+applied.stderr)
  with (out/(name+'.log')).open('w') as log:r=subprocess.run(['go','test','./internal/'+pkg,'-run',test,'-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
  output=(out/(name+'.log')).read_text();wanted={'constructor-revert':'exit codes differ','reject-tuple-objects':'JavaScript backend stdout','missing-tuple-marker':'JavaScript backend stdout','admit-arbitrary-arrays':'ordinary arrays must fail','clear-contract':'callable contract must reject','source-argument-layout':'Native backend stdout','missing-producer-mask':'Native backend stdout','spread-refusal':'got <nil>'}[name];caught=r.returncode==1 and '[build failed]' not in output and wanted in output
  results.append({'name':name,'exit':r.returncode,'caught':caught});print(name,results[-1],flush=True)
 finally:
  for p,source in originals.items():Path(p).write_bytes(source)
 (out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
# 146 metadata revert is an independent behavioral check absent from its historical report.
p=Path('internal/native/emit_objects.go');original=p.read_text();needle='\t\tif cast, ok := expression.(ir.CheckedCast); ok && cast.CheckedFields {\n\t\t\tneeded = true\n\t\t}\n';assert needle in original
try:
 mutated=original.replace(needle,'')
 (out/(('checked-cast-metadata' if 'emit_objects' in str(p) else 'historical-single-filter')+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p))))
 p.write_text(mutated)
 with (out/'checked-cast-metadata.log').open('w') as log:r=subprocess.run(['go','test','./internal/lower','-run','^TestFX7CheckedCastFieldTypes$','-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 output=(out/'checked-cast-metadata.log').read_text();results.append({'name':'checked-cast-metadata','exit':r.returncode,'caught':r.returncode==1 and '[build failed]' not in output and ('stdout' in output and ('Node' in output or 'source' in output))});print(results[-1],flush=True)
finally:p.write_text(original)
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(r['caught'] for r in results if r['name']!='checked-cast-metadata'),results
# Candidate 3 explicitly recorded this masked historical single-filter mutant.
p=Path('internal/lower/view_unions_untagged.go');original=p.read_text()
needle='\t\t\tfunction := l.result.Functions[producer.function]\n\t\t\t// Named functions also have a direct-call record. Only their closure\n\t\t\t// thunk has the code convention registered by the callable adapter.\n\t\t\tif !function.Closure || function.Receiver {\n\t\t\t\tcontinue\n\t\t\t}\n'
assert original.count(needle)==1
try:
 mutated=original.replace(needle,'')
 (out/(('checked-cast-metadata' if 'emit_objects' in str(p) else 'historical-single-filter')+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile='a/'+str(p),tofile='b/'+str(p))))
 p.write_text(mutated)
 with (out/'historical-single-filter.log').open('w') as log:r=subprocess.run(['go','test','./internal/lower','-run','^TestCallableProducerCertificatesUseClosureThunks$','-count=1','-v','-timeout','90s'],stdout=log,stderr=subprocess.STDOUT,timeout=120)
 results.append({'name':'historical-single-filter','exit':r.returncode,'caught':r.returncode==1,'historically_masked':True});print(results[-1],flush=True)
finally:p.write_text(original)
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')

