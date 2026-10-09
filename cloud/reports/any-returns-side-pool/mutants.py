from pathlib import Path
import subprocess,json,os
root=Path('/workspace/adamic')
for relative in ['stage3/census/any-returns-concrete/mutants.py','stage3/census/any-returns-concrete/step05/mutants.py']:
 source=(root/relative).read_text()
 source=source.replace('from pathlib import Path','from pathlib import Path\nimport os,json')
 source=source.replace('root = Path(__file__).resolve().parents[3]',"root = Path('/workspace/adamic')").replace('root = Path(__file__).resolve().parents[4]',"root = Path('/workspace/adamic')")
 source=source.replace('path.write_text(original.replace(before, after))',"mutated = logs / (name + '.go')\n  mutated.write_text(original.replace(before, after))\n  overlay = logs / (name + '.overlay.json')\n  overlay.write_text(json.dumps({'Replace': {str(path):str(mutated)}}))")
 source=source.replace("'-count=1','-v'","'-count=1','-timeout','90s','-v'")
 source=source.replace('cwd=root, stdout=log, stderr=log)',"cwd=root, stdout=log, stderr=log, env=dict(os.environ, GOFLAGS='-overlay='+str(overlay)))")
 source=source.replace('path.write_text(original)','pass # Go overlays leave the checkout unchanged')
 script=Path('/tmp/stack-c2-'+('loader' if 'step05' in relative else 'generic')+'-mutants.py');script.write_text(source)
 subprocess.run(['python3',str(script)],cwd=root,check=True)
path=root/'internal/lower/generic.go';original=path.read_text()
needle='overload := resolved.Declaration() != nil && resolved.Declaration().Body() == nil && l.censusImplementation(resolved.Declaration()) == declaration'
assert original.count(needle)==1
mutated=Path('/tmp/stack-c2-any-drop-overload.go');mutated.write_text(original.replace(needle,'overload := false'))
overlay=Path('/tmp/stack-c2-any-drop-overload.json');overlay.write_text(json.dumps({'Replace':{str(path):str(mutated)}}))
log=Path('/tmp/stack-c2-any-drop-overload.log')
with log.open('w') as output: result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/oracle','-run','^TestNativeAgreesWithNode$/^internal/oracle/testdata/overload_sort_deduplicate.a$','-count=1','-timeout','90s','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
text=log.read_text();assert result.returncode!=0 and 'mapper the checker' in text and '[build failed]' not in text
print('drop-overload-inference: caught by the Node-held sort-and-deduplicate oracle')

path=root/'internal/lower/generic.go';original=path.read_text()
needle='if !proven {\n\t\t\t\t\t\t\t\trefused = l.notYet(target, "assigning a field of a "+typeName(held))'
assert original.count(needle)==1
mutated=Path('/workspace/scratch/any-classify/drop-generic-field-proof.go');mutated.write_text(original.replace(needle,needle.replace('if !proven {','if false && !proven {')))
overlay=Path('/workspace/scratch/any-classify/drop-generic-field-proof.json');overlay.write_text(json.dumps({'Replace':{str(path):str(mutated)}}))
log=Path('/workspace/scratch/any-classify/drop-generic-field-proof.log')
with log.open('w') as output: result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run','^TestHiddenTNodeConstraintMutationRemainsNotYet$','-count=1','-timeout','90s','-v'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
text=log.read_text();assert result.returncode!=0 and 'want constrained mutation refusal, got <nil>' in text and '[build failed]' not in text
print('drop-generic-field-proof: caught by the constrained-mutation refusal')
