import json,pathlib,subprocess,tempfile
root=pathlib.Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip());owned=root/'stage1/cohere/lint/helpers/from_wave1_02/forks';evidence=owned/'evidence';work=pathlib.Path(tempfile.mkdtemp(prefix='wave02-restore-'))
cfg=root/'cohere/internal/lint/ecmascript/control_flow_graph/cfg.go';s=cfg.read_text().replace('import (','import (\n "fmt"',1)
anchor='func (b *Builder[E]) restoreForks(snapshot []bool) {\n';assert s.count(anchor)==1
instrument=''' before, saved, identities := wave02FrameBits(b.tryStack),wave02BoolBits(snapshot),wave02FrameIDs(b.tryStack)
 defer func(){fmt.Printf("WAVE02_RESTORE:%s|%s|%s|%s|%s\\n",before,saved,wave02FrameBits(b.tryStack),identities,wave02BoolBits(snapshot))}()
'''
s=s.replace(anchor,anchor+instrument,1)
s+='''\nfunc wave02BoolBits(values []bool) string {s:="";for _,value:=range values {if value{s+="1"}else{s+="0"}};return s}
func wave02FrameBits[E any](frames []*tryFrame[E]) string {s:="";for _,frame:=range frames {if frame.thrownForked{s+="1"}else{s+="0"}};return s}
func wave02FrameIDs[E any](frames []*tryFrame[E]) string {s:="";ids:=make(map[*tryFrame[E]]int);for i,frame:=range frames {if i>0{s+=":"};id,exists:=ids[frame];if !exists{id=i;ids[frame]=id};s+=fmt.Sprint(id)};return s}\n'''
(work/'cfg.go').write_text(s)
replacements={str(cfg):str(work/'cfg.go'),str(cfg.parent/'wave02_restore_controls_test.go'):str(owned/'restore_controls.go.txt')}
for package in ['core','react']:replacements[str(root/'cohere/internal/lint/rules'/package/'wave02_restore_probes_test.go')]=str(owned/(package+'_probes.go.txt'))
overlay=work/'overlay.json';overlay.write_text(json.dumps({'Replace':replacements}))
consumers=[('array-callback-return','core','^TestArrayCallbackReturn','ArrayCallbackReturn'),('consistent-return','core','^TestConsistentReturn','ConsistentReturn'),('no-unreachable-loop','core','^TestNoUnreachableLoop','NoUnreachableLoop'),('react-hooks-rules-of-hooks','react','^TestRulesOfHooks','RulesOfHooks')]
rows=[];counts={}
def capture(name,path,pattern,require):
 with (evidence/(name+'-restore-go.log')).open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),path,'-run',pattern,'-count=1','-v','-timeout=15m'],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode==0,'Go suite failed: '+name
 found=[line.split('WAVE02_RESTORE:',1)[1] for line in (evidence/(name+'-restore-go.log')).read_text().splitlines() if 'WAVE02_RESTORE:' in line]
 if require:assert found,'no actual helper calls: '+name
 rows.extend(found);counts[name]=len(found);print(name,len(found),'actual calls',flush=True)
for name,package,pattern,probe in consumers:
 path='./internal/lint/rules/'+package;capture(name+'-original',path,pattern,False);capture(name+'-probe',path,'^TestWave02Snapshot'+probe+'$',True)
capture('controls','./internal/lint/ecmascript/control_flow_graph','^TestWave02RestoreControls$',True)
unique=[];aliased=[]
for row in rows:
 identities=row.split('|')[3].split(':') if row.split('|')[3] else []
 (unique if len(identities)==len(set(identities)) else aliased).append(row)
assert len(aliased)==252
trace=work/'unique.txt';trace.write_text('\n'.join(unique)+'\n');wanted=trace.read_bytes();aliases=work/'alias.txt';aliases.write_text('\n'.join(aliased)+'\n');alias_wanted=aliases.read_bytes()
(evidence/'restore-coverage.json').write_text(json.dumps(counts,indent=2)+'\n')
subprocess.run(['go','build','-o',str(work/'adamic'),'./cmd/adamic'],cwd=root,check=True);subprocess.run(['go','build','-o',str(work/'sanitized-build'),'./stage1/cohere/lint/helpers/from_wave1_02/final/build'],cwd=root,check=True)
def compare(label,inputs,expected):
 runner=owned/'restore_runner.a';node=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(runner),str(inputs)])
 (work/'runner.mjs').write_bytes(subprocess.check_output([str(work/'adamic'),'js',str(runner)]))
 emitted=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(work/'runner.mjs'),str(inputs)])
 (work/'runner.c').write_bytes(subprocess.check_output([str(work/'adamic'),'c',str(runner)]));subprocess.run([str(work/'sanitized-build'),str(work/'runner.c'),str(work/'runner')],check=True)
 native=subprocess.check_output([str(work/'runner'),str(inputs)]);out=[node,emitted,native];print(label,len(expected),'bytes',*[x==expected for x in out],flush=True);return out
assert all(out==wanted for out in compare('baseline',trace,wanted))
assert all(out==alias_wanted for out in compare('alias-baseline',aliases,alias_wanted))
helper=owned/'restore_forks.a';original=helper.read_text()
clobber=original.replace('index < snapshot.length','index < frames.length').replace("snapshot[index] ?? panic('missing snapshot bit')",'snapshot[index] ?? false')
reverse=original.replace('let index = 0; index < snapshot.length; index++','let index = snapshot.length - 1; index >= 0; index--')
mutants=[('invert-bit',original.replace('frame.thrownForked = forked;','frame.thrownForked = !forked;'),trace,wanted),('clobber-suffix',clobber,trace,wanted),('reverse-shared-writes',reverse,aliases,alias_wanted)]
for name,mutant,inputs,expected in mutants:
 assert mutant!=original
 try:
  helper.write_text(mutant)
  if name=='reverse-shared-writes':assert all(out==wanted for out in compare('reversed-unique-control',trace,wanted))
  assert all(out!=expected for out in compare('mutant-'+name,inputs,expected))
 finally:helper.write_text(original)
print('PASS: original suites and actual live probes for all consumers; three clean-executing semantic mutants caught.',flush=True);print('workdir',work,flush=True)
