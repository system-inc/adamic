"""Partial React pass kernels versus unmodified Go production kernels."""
import pathlib,json,argparse,subprocess
parser=argparse.ArgumentParser();parser.add_argument('--artifacts',required=True);parser.add_argument('--stage0',required=True);args=parser.parse_args();source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
def run(name,cmd,expected=0,cwd=repo):
 with open(root/(name+'.stdout'),'wb') as out,open(root/(name+'.stderr'),'wb') as err:p=subprocess.run([str(a) for a in cmd],stdout=out,stderr=err,cwd=cwd)
 output=(root/(name+'.stdout')).read_bytes();errors=(root/(name+'.stderr')).read_bytes();assert p.returncode==expected,(name,p.returncode,output.decode(),errors.decode());return output,errors
virtual=repo/'cohere/internal/lint/rules/react/adamic_wave08_core_test.go';(root/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/core_oracle_test.go')}}))
output,errors=run('go',['go','test','-overlay',root/'overlay.json','./internal/lint/rules/react','-run','^TestWave08ReactCoreOracle$','-count=1','-v'],cwd=repo/'cohere');truth=b''.join(line.removeprefix(b'wave08-core ')+b'\n' for line in output.splitlines() if line.startswith(b'wave08-core '));assert len(truth.splitlines())==52
run('native-build',[args.stage0,'build',source/'testdata/core_probe.a','-o',root/'native','--sanitize']);output,errors=run('native',[root/'native']);assert output==truth and errors==b''
print('native partial kernels equal unmodified Go: 36 lattice joins, six mutation decisions, ten derived effects; sanitized',flush=True)
for name,file,before,after in [('join','immutability.a','return 3;','return 1;'),('dep-count','no_deriving_state_in_effects.a','!== deps.length','=== deps.length')]:
 directory=root/name;directory.mkdir(exist_ok=True)
 for path in source.glob('*.a'):
  text=path.read_text().replace("'../","'"+str(source.parent)+"/")
  if path.name==file:
   assert before in text;text=text.replace(before,after,1)
  (directory/path.name).write_text(text)
 (directory/'probe.a').write_text((source/'testdata/core_probe.a').read_text().replace("'../immutability.a'","'./immutability.a'").replace("'../no_deriving_state_in_effects.a'","'./no_deriving_state_in_effects.a'"))
 run(name+'-build',[args.stage0,'build',directory/'probe.a','-o',root/name/'native']);output,errors=run(name,[root/name/'native']);assert output!=truth and errors==b''
 print(name,'kernel mutant compiles and exits 0; unmodified Go kernel bytes catch it',flush=True)
for option in ['--immutability','--derivation']:
 output,errors=run(option,[root/'native',option],70);assert output==b'' and b'blocked:' in errors
 print(option,'production entry refuses before findings, exit 70',flush=True)
print('PASS kernels only; neither HIR-dependent rule is a completed production port',flush=True)
