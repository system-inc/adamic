#!/usr/bin/env python3
"""Validate reporting only and reproduce the native React substrate blockers."""
from pathlib import Path
import argparse,json,os,re,subprocess,time
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04/final/adamic');a=p.parse_args();o=a.directory.resolve();o.mkdir(parents=True,exist_ok=True);own=Path(__file__).resolve().parent;r=own.parents[3];records=[]
def run(label,command,expected=0,env=None,cwd=r):
 stem=o/label;started=time.perf_counter()
 with stem.with_suffix('.stdout').open('wb') as stdout,stem.with_suffix('.stderr').open('wb') as stderr:
  result=subprocess.run([str(x) for x in command],cwd=cwd,env=env,stdout=stdout,stderr=stderr)
 stdout=stem.with_suffix('.stdout').read_bytes();stderr=stem.with_suffix('.stderr').read_bytes()
 if result.returncode!=expected:raise RuntimeError(f'{label}: expected {expected}, got {result.returncode}')
 records.append(dict(label=label,exit=result.returncode,stdout_bytes=len(stdout),stderr_bytes=len(stderr),seconds=time.perf_counter()-started));print(records[-1],flush=True);return stdout,stderr
virtual=r/'cohere/internal/lint/rules/react/wave04_export_test.go';overlay=o/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'testdata/export_test.go')}}))
run('go-controls',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04ExportReactBlockers$','-count=1','-v'],env=dict(os.environ,WAVE04_REACT_RECORDS=str(o/'controls.json')),cwd=r/'cohere')
controls=json.loads((o/'controls.json').read_text());driver="import { PreserveManualMemoization } from './preserve_manual_memoization.a';\nimport { Purity } from './purity.a';\nimport { Refs } from './refs.a';\n";expected=[]
for c in controls:
 if c['Source']:(o/(c['Name']+'.tsx')).write_text(c['Source'])
 for d in c['Findings']:
  if c['Name']=='memo':expr="new PreserveManualMemoization().render("+str(d['Id'].endswith('DependencyMutable')).lower()+f", {d['Start']}, {d['End']})"
  elif c['Name']=='purity':expr='new Purity().render('+json.dumps(d['Message'].split('`')[1])+f", {d['Start']}, {d['End']})"
  else:expr='new Refs().render('+json.dumps(d['Id'])+f", {d['Start']}, {d['End']})"
  driver+='console.log('+expr+'.written());\n';expected.append(f"{d['Start']}\t{d['End']}\t{d['Rule']}\t{d['Id']}\t{d['Message']}\t0\t0\n")
# Driver generation is isolated; implementation modules stay in the owned directory.
driver=driver.replace("from './","from '"+str(own)+'/');(o/'reporting.a').write_text(driver);truth=''.join(expected).encode();(o/'reporting.expected').write_bytes(truth)
for sanitized in [False,True]:
 variant='asan' if sanitized else 'normal';command=[a.adamic,'build',o/'reporting.a','-o',o/('reporting-'+variant)]
 if sanitized:command.append('--sanitize')
 run('build-'+variant,command);actual,error=run('run-'+variant,[o/('reporting-'+variant)])
 if error or actual!=truth:raise RuntimeError('reporting '+variant+' differs from Go')
actual,error=run('run-source-node',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',o/'reporting.a'])
if error or actual!=truth:raise RuntimeError('source Node reporting differs from Go')
js,_=run('emit-js',[a.adamic,'js',o/'reporting.a']);(o/'reporting.js').write_bytes(js);actual,error=run('run-js',['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs',o/'reporting.js'])
if error or actual!=truth:raise RuntimeError('reporting emitted JS differs from Go')
run('parser-build',[a.adamic,'build',own/'parser_probe.a','-o',o/'parser-probe'])
for rule,status in [('memo',0),('purity',70),('refs',0)]:run('parser-'+rule,[o/'parser-probe',o/(rule+'.tsx')],expected=status)
run('analysis-build',[a.adamic,'build',own/'analysis_probe.a','-o',o/'analysis-probe'])
for rule in ['memo','purity','refs']:
 stdout,stderr=run('analysis-'+rule,[o/'analysis-probe',rule],expected=70)
 if stdout or b'NotYet:' not in stderr:raise RuntimeError('missing explicit HIR refusal')
mutants=[]
for filename,rule in [('preserve_manual_memoization.a','memo'),('purity.a','purity'),('refs.a','refs')]:
 for change in ['span','refusal']:
  directory=o/(rule+'-'+change+'-mutant');directory.mkdir(exist_ok=True)
  for path in own.glob('*.a'):
   source=path.read_text().replace("from '../","from '"+str(own.parent)+'/').replace("from '../../../typescript/","from '"+str(r/'stage1/typescript')+'/')
   if path.name==filename:
    if change=='span':
     if source.count(', start, end,')!=1:raise RuntimeError('nonunique span mutation')
     source=source.replace(', start, end,',', start, end + 1,')
    else:source=re.sub(r"analyze\(\): void \{ panic\('[^']*'\); \}","analyze(): void {}",source)
   (directory/path.name).write_text(source)
  if change=='span':(directory/'reporting_probe.a').write_text(driver.replace("from '"+str(own)+'/',"from './"))
  entry='reporting_probe.a' if change=='span' else 'analysis_probe.a'
  run(rule+'-'+change+'-build',[a.adamic,'build',directory/entry,'-o',directory/'native'])
  stdout,stderr=run(rule+'-'+change+'-run',[directory/'native',rule] if change=='refusal' else [directory/'native'])
  if stderr or change=='span' and (stdout==truth or len(stdout.splitlines())!=len(truth.splitlines())):raise RuntimeError('mutant survived or failed outside expected check')
  mutants.append(dict(rule=rule,change=change,exit=0,stderr_bytes=0,caught_by='reporting byte comparison' if change=='span' else 'expected NotYet panic assertion'))
(o/'runs.json').write_text(json.dumps(records,indent=2)+'\n');(o/'mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
print('PASS partial reporting only: Go/native/sanitized/emitted-JS bytes; six reporting/refusal mutants. Full rule analyses remain blocked.',flush=True)
