import json,pathlib,subprocess,tempfile
root=pathlib.Path(subprocess.check_output(['git','rev-parse','--show-toplevel'],text=True).strip())
owned=root/'stage1/cohere/lint/helpers/from_wave1_02/thrown';evidence=owned/'evidence';evidence.mkdir(exist_ok=True)
work=pathlib.Path(tempfile.mkdtemp(prefix='wave02-thrown-'))
cfg=root/'cohere/internal/lint/ecmascript/control_flow_graph/cfg.go';s=cfg.read_text().replace('import (','import (\n "fmt"',1)
anchor='func (b *Builder[E]) markThrown(blk *Block[E]) {'
assert s.count(anchor)==1
instrument='''
 id, reachable, wasThrown := int32(-1),false,false
 if blk != nil {id,reachable,wasThrown=blk.index,blk.Reachable,blk.thrown}
 before, current := wave02ThrownIDs(b.thrown), b.cur.index
 finalBefore, finalsBefore := false, wave02ThrownIDs(b.finals)
 if blk != nil {finalBefore=blk.final}
 defer func(){
  thrown, same, finalAfter := false,false,false
  if blk != nil { thrown=blk.thrown; finalAfter=blk.final; same=len(b.thrown)>0 && b.thrown[len(b.thrown)-1]==blk }
  fmt.Printf("WAVE02_THROWN:%d,%t,%t,%s,%t,%s,%t,%d,%d,%t,%t,%s,%s\\n",id,reachable,wasThrown,before,thrown,wave02ThrownIDs(b.thrown),same,current,b.cur.index,finalBefore,finalAfter,finalsBefore,wave02ThrownIDs(b.finals))
 }()
'''
s=s.replace(anchor,anchor+instrument)
s+='''\nfunc wave02ThrownIDs[E any](blocks []*Block[E]) string {s:="";for i,block:=range blocks {if i>0{s+=":"};s+=fmt.Sprint(block.index)};return s}\n'''
(work/'cfg.go').write_text(s);(work/'controls.go').write_text((owned/'controls.go.txt').read_text())
overlay=work/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(cfg):str(work/'cfg.go'),str(cfg.parent/'wave02_thrown_controls_test.go'):str(work/'controls.go')}}))
consumers=[('array-callback-return','core','^TestArrayCallbackReturn'),('consistent-return','core','^TestConsistentReturn'),('no-unreachable-loop','core','^TestNoUnreachableLoop'),('react-hooks-rules-of-hooks','react','^TestRulesOfHooks'),('controls',None,'^TestWave02ThrownControls$')]
rows=[];counts={}
for name,package,pattern in consumers:
 path='./internal/lint/rules/'+package if package else './internal/lint/ecmascript/control_flow_graph'
 with (evidence/(name+'-go.log')).open('wb') as log:
  result=subprocess.run(['go','test','-overlay='+str(overlay),path,'-run',pattern,'-count=1','-v','-timeout=15m'],cwd=root/'cohere',stdout=log,stderr=subprocess.STDOUT)
 assert result.returncode==0,'Go consumer failed: '+name
 lines=(evidence/(name+'-go.log')).read_text().splitlines()
 found=[line.split('WAVE02_THROWN:',1)[1] for line in lines if 'WAVE02_THROWN:' in line]
 assert found,'no actual helper calls: '+name
 counts[name]=len(found);rows+=found
 print(name,len(found),'actual calls',flush=True)
trace=work/'trace.txt';trace.write_text('\n'.join(rows)+'\n');wanted=trace.read_bytes();(evidence/'coverage.json').write_text(json.dumps(counts,indent=2)+'\n')
subprocess.run(['go','build','-o',str(work/'adamic'),'./cmd/adamic'],cwd=root,check=True)
subprocess.run(['go','build','-o',str(work/'sanitized-build'),'./stage1/cohere/lint/helpers/from_wave1_02/final/build'],cwd=root,check=True)
runner=owned/'runner.a'
def compare(label):
 node=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(runner),str(trace)])
 (work/'runner.mjs').write_bytes(subprocess.check_output([str(work/'adamic'),'js',str(runner)]))
 emitted=subprocess.check_output(['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(work/'runner.mjs'),str(trace)])
 (work/'runner.c').write_bytes(subprocess.check_output([str(work/'adamic'),'c',str(runner)]))
 subprocess.run([str(work/'sanitized-build'),str(work/'runner.c'),str(work/'runner')],check=True)
 native=subprocess.check_output([str(work/'runner'),str(trace)])
 print(label,len(wanted),'bytes',node==wanted,emitted==wanted,native==wanted,flush=True)
 return [node,emitted,native]
assert all(out==wanted for out in compare('baseline'))
helper=owned/'mark_thrown.a';original=helper.read_text()
for label,old,new in [('omit-unreachable-guard','!block.reachable || ',''),('omit-thrown-guard',' || block.thrown',''),('write-final-flag','block.thrown = true;','block.final = true;')]:
 assert original.count(old)==1
 try:
  helper.write_text(original.replace(old,new))
  assert all(out!=wanted for out in compare('mutant-'+label))
 finally:helper.write_text(original)
print('PASS: actual calls from all four consumer suites; three compiling semantic mutants caught only by byte comparison.',flush=True)
print('workdir',work,flush=True)
