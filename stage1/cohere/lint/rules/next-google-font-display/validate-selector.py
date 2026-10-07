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
 target.mkdir(exist_ok=True);env=os.environ.copy();env['WAVE07_PROFILE']=str(entry);env['WAVE07_ARTIFACTS']=str(target)
 run(['go','test','./'+str(owned.relative_to(root)),'-run','^TestCompileProfiles$','-count=1','-v'],env=env)
virtual=root/'cohere/adamic_wave07_selector.go';adapter=root/'cohere/adamic_wave07_selector_adapter.go'
overlay=scratch/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(owned/'testdata/selector-oracle.go.txt'),str(adapter):str(owned/'oracle.go')}}))
oracle=scratch/'oracle';run(['go','build','-overlay='+str(overlay),'-o',str(oracle),str(virtual),str(adapter)],cwd=root/'cohere')
baseline=scratch/'baseline';build(owned/'selector-profile.a',baseline)
runner=root/'oracle/node.mjs'
fixtures=scratch/'fixtures';fixtures.mkdir(exist_ok=True)
upstream=(root/'cohere/internal/lint/rules/next/google_font_display_test.go').read_text()
sources=[value for value in re.findall(r'`([^`]+)`',upstream) if value.startswith('export const Test =')]
if len(sources)!=16:raise RuntimeError('expected all sixteen upstream source cases')
for href in ['https://fonts.googleapis.com/css2?display=block&display=swap','https://fonts.googleapis.com/css2?display=swap&display=block','https://fonts.googleapis.com/css2?display=','https://fonts.googleapis.com/css2?Display=block','https://fonts.googleapis.com/css2?display=block#fragment','https://fonts.googleapis.com/css2?d%69splay=block','https://fonts.googleapis.com/css2?display=a=block','https://fonts.googleapis.com/css2?display=block?ignored','https://fonts.googleapis.com/css2?x=1&#38;display=block','https://fonts.googleapis.com/css2?x=1&#x26;display=block','https://fonts.googleapis.com/css2?x=1&#X26;display=block']:
 sources.append('export const Test = () => <link HREF="'+href+'" />;')
manifest=scratch/'selector-manifest.txt';rows=[]
for index,source in enumerate(sources):
 path=fixtures/(str(index)+'.tsx');path.write_text(source)
 tag=re.search(r'<([a-zA-Z][\w.]*)',source).group(1)
 attribute=re.search(r'([\w:]+)="([^"\n]*)"',source)
 rows.append(str(path)+'\t'+tag+'\t'+(attribute.group(1) if attribute else '')+'\t'+('literal' if attribute else 'computed')+'\t'+(attribute.group(2) if attribute else ''))
manifest.write_text('\n'.join(rows)+'\n')
def commands(artifacts,entry,manifest,count=False):
 extra=['--count'] if count else []
 return {'Go':[str(oracle),str(manifest),*extra],'Node':['node','--disable-warning=ExperimentalWarning',str(runner),str(entry),str(manifest),*extra],'emitted JavaScript':['node','--disable-warning=ExperimentalWarning',str(runner),str(artifacts/'emitted.mjs'),str(manifest),*extra],'sanitized native':[str(artifacts/'native'),str(manifest),*extra]}
answers={side:run(command) for side,command in commands(baseline,owned/'selector-profile.a',manifest).items()}
for side,value in answers.items():
 if value!=answers['Go']:raise RuntimeError('selector bytes differ '+side)
log.write(f'PASS selector only: {len(sources)} rows including all 16 upstream sources; {len(answers["Go"])} identical bytes\n');log.flush()
# This is the rule decision on externally provided JSX attributes, not parser parity.
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
mutant=scratch/'mutant';build(copy/'selector-profile.a',mutant)
for side,command in commands(mutant,copy/'selector-profile.a',manifest).items():
 if side=='Go':continue
 answer=run(command)
 if answer==answers['Go']:raise RuntimeError('mutant survived '+side)
 log.write('PASS selector mutant block-display-accepted caught only by Go byte comparison on '+side+'\n');log.flush()
# Count rate includes startup and full Go parsing, while Adamic uses supplied attribute geometry.
# It is a partial selector measurement, not a whole-rule throughput comparison.
manifest=scratch/'throughput-manifest.txt';manifest.write_text((rows[3]+'\n')*1000)
cmd=commands(baseline,owned/'selector-profile.a',manifest,True);cmd['native']=[str(baseline/'release'),str(manifest),'--count']
for side in ['Go','native','Node']:
 best=1000000
 for repeat in range(5):
  started=time.perf_counter();answer=run(cmd[side]);best=min(best,time.perf_counter()-started)
  if answer!=b'1000\n':raise RuntimeError('bad throughput count')
 log.write(f'SELECTOR THROUGHPUT {side} {1000/best:.2f} findings/s best of 5 including startup; Go parses JSX, Adamic receives attributes\n');log.flush()
log.write('BLOCKED full parser, ranges, corpus and whole-rule throughput: current Parser cannot construct JSX trees\n');log.close()
print('PASS selector only '+str(scratch/'validation.txt'))
