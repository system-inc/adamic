#!/usr/bin/env python3
"""Independent owned profile comparison, without modifying the shared harness."""
import argparse,json,os,re,subprocess,time,shutil
from pathlib import Path
owned=Path(__file__).resolve().parent
root=owned.parents[4]
parser=argparse.ArgumentParser();parser.add_argument('--scratch',type=Path,required=True);parser.add_argument('--compiler',type=Path,required=True);args=parser.parse_args()
scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
log=(scratch/'validation.txt').open('w')
sequence=0
def run(command,cwd=root,env=None):
 global sequence
 sequence+=1
 log.write('command '+repr(command)+'\n');log.flush()
 out=scratch/(f'{sequence:03d}.stdout.txt');err=scratch/(f'{sequence:03d}.stderr.txt')
 with out.open('wb') as stdout,err.open('wb') as stderr:
  result=subprocess.run(command,cwd=cwd,env=env,stdout=stdout,stderr=stderr)
 output=out.read_bytes();errors=err.read_bytes()
 log.write(f'exit {result.returncode}, stdout {len(output)} bytes, stderr {len(errors)} bytes\n');log.flush()
 if result.returncode or errors:
  log.write(errors.decode(errors='replace'));log.write(output.decode(errors='replace'));log.flush()
  raise RuntimeError('execution failed: '+repr(command))
 return output
# Go compiler commands can print benign build information; successful executions below require no stderr.
def build(entry,target):
 target.mkdir(exist_ok=True)
 compiler=scratch/'adamic-profile-compiler'
 if not compiler.exists():
  run(['go','build','-o',str(compiler),'./cmd/adamic'])
 run([str(compiler),'build',str(entry),'-o',str(target/'native'),'--sanitize'])
 run([str(compiler),'build',str(entry),'-o',str(target/'release')])
 (target/'emitted.mjs').write_bytes(run([str(compiler),'js',str(entry)]))

variant='NoThisAlias'
virtual=root/'cohere/adamic_wave07_owned.go';adapter=root/'cohere/adamic_wave07_adapter.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'testdata/oracle-driver.go.txt'),str(adapter):str(owned/'oracle.go')}}))
oracle=scratch/'oracle';run(['go','build','-overlay='+str(overlay),'-o',str(oracle),str(virtual),str(adapter)],cwd=root/'cohere')
baseline=scratch/'baseline';build(owned/'profile.a',baseline)
runner=root/'oracle/node.mjs'
def commands(artifacts,entry,manifest,count=False):
 extra=['--count'] if count else []
 return {'Go':[str(oracle),str(manifest),*extra],'Node':['node','--disable-warning=ExperimentalWarning',str(runner),str(entry),str(manifest),*extra],'emitted JavaScript':['node','--disable-warning=ExperimentalWarning',str(runner),str(artifacts/'emitted.mjs'),str(manifest),*extra],'sanitized native':[str(artifacts/'native'),str(manifest),*extra]}
def compare(label,paths):
 manifest=scratch/(label+'-manifest.txt');manifest.write_text(''.join(str(p.resolve())+'\n' for p in paths))
 answers={side:run(command) for side,command in commands(baseline,owned/'profile.a',manifest).items()}
 for side,value in answers.items():
  if value!=answers['Go']:
   for index,(left,right) in enumerate(zip(value.splitlines(),answers['Go'].splitlines())):
    if left!=right:log.write(f'difference {side} line {index}: {left!r} vs {right!r}\n');log.flush();break
   raise RuntimeError('byte difference '+side+' '+label)
 log.write(f'PASS {label}: {len(paths)} file rows, {len(answers["Go"])} identical bytes\n');log.flush()
 return manifest,answers['Go']
test=root/'cohere/internal/lint/rules/typescript'/'no_this_alias_test.go'
# Include each source literal and repaired-expression literal from the complete upstream test file.
sources=set()
for literal in re.findall(r'"(?:[^"\\]|\\.)*"',test.read_text()):
 value=json.loads(literal)
 if ';' in value or 'foo?.' in value:sources.add(value)
fixtures=scratch/'fixtures';fixtures.mkdir(exist_ok=True)
paths=[]
for index,text in enumerate(sorted(sources)):
 path=fixtures/(str(index)+'.ts');path.write_text(text);paths.append(path)
manifest,want=compare('upstream',paths)
pin=run(['git','-C',str(args.compiler),'rev-parse','HEAD']).decode().strip()
if pin!='050880ce59e30b356b686bd3144efe24f875ebc8':raise RuntimeError('wrong compiler pin')
paths=sorted((args.compiler/'src/compiler').rglob('*.ts'))+sorted((root/'stage1').rglob('*.ts'))+sorted((root/'stage1').rglob('*.a'))
compare('compiler-stage1',paths)
# Both semantic mutants compile and execute, then only the external Go comparison rejects them.
mutation=json.loads((owned/'mutant.json').read_text());copy=scratch/'mutant-source';copy.mkdir(exist_ok=True)
for path in owned.glob('*.a'):
 text=path.read_text()
 if path.name==mutation['file']:
  if text.count(mutation['from'])!=1:raise RuntimeError('mutant anchor not unique')
  text=text.replace(mutation['from'],mutation['to'])
 def rewrite(match):
  if match.group(1)=='adamic':return match.group(0)
  target=(owned/match.group(1)).resolve()
  return "from '"+str(copy/target.name if target.parent==owned else target)+"'"
 text=re.sub(r"from '([^']+)'",rewrite,text);(copy/path.name).write_text(text)
mutant=scratch/'mutant';build(copy/'profile.a',mutant)
for side,command in commands(mutant,copy/'profile.a',manifest).items():
 if side=='Go':continue
 answer=run(command)
 if answer==want:raise RuntimeError('mutant survived '+side)
 log.write('PASS mutant '+mutation['name']+' caught only by byte comparison on '+side+'\n');log.flush()
throughput=fixtures/'throughput.ts';throughput.write_text(''.join('const self'+str(i)+' = this;\n' for i in range(1000)))
manifest=scratch/'throughput-manifest.txt';manifest.write_text(str(throughput)+'\n')
cmd=commands(baseline,owned/'profile.a',manifest,True);cmd['native']=[str(baseline/'release'),str(manifest),'--count']
for side in ['Go','native','Node']:
 best=1000000
 for repeat in range(5):
  started=time.perf_counter();answer=run(cmd[side]);best=min(best,time.perf_counter()-started)
  if answer!=b'1000\n':raise RuntimeError('bad throughput count')
 log.write(f'THROUGHPUT {side} {1000/best:.2f} findings/s best of 5 including startup\n');log.flush()
log.write('PASS owned profile on the owner harness foundation\n');log.close()
print('PASS '+owned.name+' evidence '+str(scratch/'validation.txt'))
