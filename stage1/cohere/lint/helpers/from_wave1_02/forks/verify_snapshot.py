import json,pathlib,subprocess,tempfile
root=pathlib.Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip());owned=root/'stage1/cohere/lint/helpers/from_wave1_02/forks';evidence=owned/'evidence';work=pathlib.Path(tempfile.mkdtemp(prefix='wave02-snapshot-'))
cfg=root/'cohere/internal/lint/ecmascript/control_flow_graph/cfg.go';s=cfg.read_text().replace('import (','import (\n "fmt"',1)
anchor='func (b *Builder[E]) snapshotForks() []bool {\n';assert s.count(anchor)==1
s=s.replace(anchor,anchor+' before := wave02FrameBits(b.tryStack)\n',1)
anchor='\treturn snapshot\n}';assert s.count(anchor)==1
s=s.replace(anchor,' fmt.Printf("WAVE02_SNAPSHOT:%s|%s|%s\\n",before,wave02BoolBits(snapshot),wave02FrameBits(b.tryStack))\n'+anchor,1)
s+='''\nfunc wave02BoolBits(values []bool) string {s:="";for _,value:=range values {if value{s+="1"}else{s+="0"}};return s}
func wave02FrameBits[E any](frames []*tryFrame[E]) string {s:="";for _,frame:=range frames {if frame.thrownForked{s+="1"}else{s+="0"}};return s}\n'''
(work/'cfg.go').write_text(s)
replacements={str(cfg):str(work/'cfg.go'),str(cfg.parent/'wave02_snapshot_controls_test.go'):str(owned/'controls.go.txt')}
for package in ['core','react']:replacements[str(root/'cohere/internal/lint/rules'/package/'wave02_snapshot_probes_test.go')]=str(owned/(package+'_probes.go.txt'))
overlay=work/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
consumers=[('array-callback-return','core','^TestArrayCallbackReturn','ArrayCallbackReturn'),('consistent-return','core','^TestConsistentReturn','ConsistentReturn'),('no-unreachable-loop','core','^TestNoUnreachableLoop','NoUnreachableLoop'),('react-hooks-rules-of-hooks','react','^TestRulesOfHooks','RulesOfHooks')]
rows=[];counts={};alias=[]
def capture(name,path,pattern,require):
 with (evidence/(name+'-go.log')).open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),path,'-run',pattern,'-count=1','-v','-timeout=15m'],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode==0,'Go suite failed: '+name
 lines=(evidence/(name+'-go.log')).read_text().splitlines();found=[line.split('WAVE02_SNAPSHOT:',1)[1] for line in lines if 'WAVE02_SNAPSHOT:' in line]
 if require:assert found,'no actual helper calls: '+name
 counts[name]=len(found);rows.extend(found)
 alias.extend(line.split('WAVE02_SNAPSHOT_ALIAS:',1)[1] for line in lines if 'WAVE02_SNAPSHOT_ALIAS:' in line)
 print(name,len(found),'actual calls',flush=True)
for name,package,pattern,probe in consumers:
 path='./internal/lint/rules/'+package
 capture(name+'-original',path,pattern,False)
 capture(name+'-probe',path,'^TestWave02Snapshot'+probe+'$',True)
capture('controls','./internal/lint/ecmascript/control_flow_graph','^TestWave02SnapshotControls$',True)
assert len(alias)==31
trace=work/'trace.txt';trace.write_text('\n'.join(rows)+'\n');wanted=trace.read_bytes();aliases=work/'alias.txt';aliases.write_text('\n'.join(alias)+'\n');alias_wanted=aliases.read_bytes()
(evidence/'snapshot-coverage.json').write_text(json.dumps(counts,indent=2)+'\n')
subprocess.run(['go','build','-o',str(work/'adamic'),'./cmd/adamic'],cwd=root,check=True);subprocess.run(['go','build','-o',str(work/'sanitized-build'),'./stage1/cohere/lint/helpers/from_wave1_02/final/build'],cwd=root,check=True)
def compare(label,entry,inputs,expected):
 runner=owned/entry
 node=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(runner),str(inputs)])
 (work/'runner.mjs').write_bytes(subprocess.check_output([str(work/'adamic'),'js',str(runner)]))
 emitted=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(work/'runner.mjs'),str(inputs)])
 (work/'runner.c').write_bytes(subprocess.check_output([str(work/'adamic'),'c',str(runner)]));subprocess.run([str(work/'sanitized-build'),str(work/'runner.c'),str(work/'runner')],check=True)
 native=subprocess.check_output([str(work/'runner'),str(inputs)]);out=[node,emitted,native];print(label,len(expected),'bytes',*[x==expected for x in out],flush=True);return out
assert all(out==wanted for out in compare('baseline','runner.a',trace,wanted))
assert all(out==alias_wanted for out in compare('alias-baseline','alias.a',aliases,alias_wanted))
helper=owned/'snapshot_forks.a';original=helper.read_text()
mutants=[('invert-bit',original.replace('snapshot.push(frame.thrownForked);','snapshot.push(!frame.thrownForked);'),'runner.a',trace,wanted),('reverse-order',original.replace('let index = 0; index < frames.length; index++','let index = frames.length - 1; index >= 0; index--'),'runner.a',trace,wanted),('reuse-buffer','const reusedSnapshot: boolean[] = [];\n'+original.replace('const snapshot: boolean[] = [];','const snapshot = reusedSnapshot;\n    while (snapshot.length > 0) snapshot.pop();'),'alias.a',aliases,alias_wanted)]
for name,mutant,entry,inputs,expected in mutants:
 assert mutant!=original
 try:
  helper.write_text(mutant)
  if name=='reuse-buffer':
   assert all(out==wanted for out in compare('reused-buffer-pointwise-control','runner.a',trace,wanted))
  assert all(out!=expected for out in compare('mutant-'+name,entry,inputs,expected))
 finally:helper.write_text(original)
print('PASS: original suites and live probes for every consumer; three clean-executing semantic mutants caught.',flush=True);print('workdir',work,flush=True)
