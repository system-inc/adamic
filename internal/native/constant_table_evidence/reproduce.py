#!/usr/bin/env python3
"""Source the toolchain environment, then run from any directory. Logs stay in scratch."""
import gzip,json,os,pathlib,shutil,subprocess,tempfile
root=pathlib.Path(__file__).resolve().parents[3]
evidence=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path(tempfile.mkdtemp(prefix='adamic-constant-tables-'))
real=shutil.which('clang')
assert real,'source the cloud/setup.sh toolchain environment first'
for version in ('before','after'):
 (scratch/(version+'.c')).write_bytes(gzip.decompress((evidence/(version+'.c.gz')).read_bytes()))
folder=scratch/'full';folder.mkdir()
wrapper=folder/'clang'
wrapper.write_text((evidence/'clang.py').read_text().replace("real = '/workspace/adamic-tools/llvm/bin/clang'",'real = '+repr(real)))
wrapper.chmod(0o755)
env=os.environ.copy()
env.update(OUTLINE_MEASURE=str(scratch),OUTLINE_ROOT=str(root),ADAMIC_GATE_UNCACHED='1',ADAMIC_NATIVE_SPLIT='1',ADAMIC_NATIVE_JOBS='5')
def run(label,replacements,test):
 overlay=scratch/(label+'.json');overlay.write_text(json.dumps({'Replace':replacements}))
 with (scratch/(label+'.log')).open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-v','-count=1','-timeout=30m'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 print(label,'exit',result.returncode,'log',scratch/(label+'.log'),flush=True)
 assert result.returncode==0
run('measure',{str(root/'internal/native/constant_measure_test.go'):str(evidence/'measure.go.txt')},'TestMeasureOutlinedBuild')
baseline=scratch/'baseline-expressions.go'
baseline.write_bytes(subprocess.check_output(['git','show','a7ddba49:internal/native/emit_expressions.go'],cwd=root))
results=[]
for round in range(3):
 for version in (('after','before') if round==1 else ('before','after')):
  env['OUTLINE_VERSION']=version
  replacements={str(root/'internal/native/constant_end_to_end_test.go'):str(evidence/'end_to_end.go.txt')}
  if version=='before':replacements[str(root/'internal/native/emit_expressions.go')]=str(baseline)
  run('e2e-'+str(round)+'-'+version,replacements,'TestOutlineEndToEnd')
  row=json.loads((scratch/('end-to-end-'+version+'.json')).read_text())[0]
  row['Round']=round;results.append(row)
  (scratch/'end-to-end.json').write_text(json.dumps(results,indent=2)+'\n')
subprocess.run(['python3',str(evidence/'summarize.py'),str(scratch)],check=True)
print('Saved measurements:',scratch)
