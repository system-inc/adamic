import pathlib,subprocess,tempfile,json
root=pathlib.Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip());owned=root/'stage1/cohere/lint/helpers/from_wave1_02';work=pathlib.Path(tempfile.mkdtemp(prefix='wave02-joint-'))
virtual=root/'cohere/internal/lint/ecmascript/control_flow_graph/wave02_joint_test.go';overlay=work/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'joint.go.txt')}}))
with (owned/'evidence/joint-go.log').open('wb') as log:
 result=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/lint/ecmascript/control_flow_graph','-run','^TestWave02JointMarkers$','-count=1','-v'],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
assert result.returncode==0
rows=[line.split('WAVE02_JOINT:',1)[1] for line in (owned/'evidence/joint-go.log').read_text().splitlines() if 'WAVE02_JOINT:' in line];assert len(rows)==8
wanted=('\n'.join(rows)+'\n').encode()
subprocess.run(['go','build','-o',str(work/'adamic'),'./cmd/adamic'],cwd=root,check=True)
subprocess.run(['go','build','-o',str(work/'sanitized-build'),'./stage1/cohere/lint/helpers/from_wave1_02/final/build'],cwd=root,check=True)
def compare(label):
 runner=owned/'joint.a';node=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(runner)])
 (work/'joint.mjs').write_bytes(subprocess.check_output([str(work/'adamic'),'js',str(runner)]))
 emitted=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(work/'joint.mjs')])
 (work/'joint.c').write_bytes(subprocess.check_output([str(work/'adamic'),'c',str(runner)]))
 subprocess.run([str(work/'sanitized-build'),str(work/'joint.c'),str(work/'joint')],check=True)
 native=subprocess.check_output([str(work/'joint')]);print(label,len(wanted),node==wanted,emitted==wanted,native==wanted,flush=True);return [node,emitted,native]
assert all(out==wanted for out in compare('joint-baseline'))
helper=owned/'thrown/mark_thrown.a';original=helper.read_text();old='builder.thrown.push(handle);';assert original.count(old)==1
try:
 helper.write_text(original.replace(old,'builder.finals.push(handle);'))
 assert all(out!=wanted for out in compare('joint-mutant-wrong-list'))
finally:helper.write_text(original)
print('PASS: joint arena identity and list independence; compiling wrong-list mutant caught by all three comparisons.',flush=True)
