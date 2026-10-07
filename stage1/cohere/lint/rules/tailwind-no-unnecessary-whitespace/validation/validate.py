#!/usr/bin/env python3
"""Rule-owned independent comparison; shared Adamic harness files stay untouched."""
from pathlib import Path
import json,subprocess,os,re,shutil,time
owned=Path(__file__).resolve().parent
repo=owned.parents[5]
roots=[owned.parent]
names=['better-tailwindcss/no-unnecessary-whitespace']
scratch=Path('/tmp/w05-fourth-whitespace');scratch.mkdir(exist_ok=True)
def command(args,label,cwd=repo,env=None):
 started=time.monotonic()
 with (scratch/(label+'.log')).open('wb') as out,(scratch/(label+'.stderr.log')).open('wb') as err:
  subprocess.run(args,cwd=cwd,env=env,stdout=out,stderr=err,check=True)
 if (scratch/(label+'.stderr.log')).stat().st_size:raise RuntimeError(label+' wrote stderr')
 return time.monotonic()-started
virtual=owned/'build.go'
(scratch/'build-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(owned/'build.go.txt')}}))
cohere=repo/'cohere'
replacements={str(cohere/'adamic_owned_oracle.go'):str(owned/'oracle.go.txt')}
for i,p in enumerate(roots):replacements[str(cohere/('adamic_owned_adapter'+str(i)+'.go'))]=str(p/'oracle.go')
(scratch/'oracle-overlay.json').write_text(json.dumps({'Replace':replacements}))
command(['go','build','-overlay='+str(scratch/'oracle-overlay.json'),'-o',str(scratch/'oracle'),*replacements.keys()],'oracle-build',cohere)
# Observe every successful upstream Run, including span/fix assertions without Expect helpers.
harness=cohere/'internal/lint/testing/rule_testing.go'
text=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
assert text.count(anchor)==1
(scratch/'capture.go').write_text(text.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t,result)\n return result'))
(scratch/'capture-overlay.json').write_text(json.dumps({'Replace':{str(harness):str(scratch/'capture.go')}}))
capture=scratch/'capture';capture.mkdir(exist_ok=True)
# Remove only this scratch run's prior capture records.
for path in capture.glob('*.jsonl'):path.unlink()
env=os.environ.copy();env['COHERE_DOCS_CAPTURE']=str(capture)
command(['go','test','-overlay='+str(scratch/'capture-overlay.json'),'./internal/lint/rules/tailwind','-run','^Test(NoUnnecessaryWhitespace|WhitespaceFix)','-count=1','-v'],'upstream-tests',cohere,env)
unique={}
for path in capture.glob('*.jsonl'):
 for line in path.read_text().splitlines():
  row=json.loads(line)
  if row['rule'] in names:
   key=json.dumps([row['rule'],row['file'],row.get('options'),row['source']],sort_keys=True)
   unique[key]=row
rows=[];counts={n:0 for n in names};excluded=[]
for i,key in enumerate(sorted(unique)):
 row=unique[key]
 if re.search(r'<(?:div|img)\b',row['source']):excluded.append(row);continue
 path=scratch/('case-'+str(i)+Path(row['file']).suffix);path.write_text(row['source']);counts[row['rule']]+=1
 rows.append(str(path)+'\t'+row['rule']+'\t'+json.dumps(row.get('options'),separators=(',',':')))
for i,p in enumerate(roots):
 for j,source in enumerate((p/'testdata').glob('*.ts.txt')):
  path=scratch/('witness-'+str(i)+'-'+str(j)+'.ts');path.write_bytes(source.read_bytes());rows.append(str(path)+'\t'+names[i]+'\t'+'null')
(scratch/'excluded.json').write_text(json.dumps(excluded,indent=2))
print('upstream cases',counts,'JSX exclusions',len(excluded),flush=True)
for i,text in enumerate(['  flex   items-center  gap-2 ', 'flex  \n    items-center', 'flex\t\titems-center','  ','','flex\u00a0  gap-2']):
 for multiline in [True,False]:
  path=scratch/('synthetic-'+str(i)+'-'+str(multiline)+'.ts');path.write_text('const buttonClassName = '+json.dumps(text)+';\n');rows.append(str(path)+'\t'+names[0]+'\t'+json.dumps({'allowMultiline':multiline}))
corpus=Path(os.environ.get('ADAMIC_TYPESCRIPT_SOURCE','/tmp/lint-wave1-05-typescript'))
files=sorted((corpus/'src/compiler').rglob('*.ts'))
files+=sorted(path for path in (repo/'stage1').rglob('*') if path.suffix in ['.ts','.a'] and '.generated' not in path.parts and 'testdata' not in path.parts)
# The driver is a valid stage1 source too; raw witness text remains outside the graph.
print('compiler/stage1 files',len(files),flush=True)
rows += [str(path)+'\t'+name+'\tnull' for path in files for name in names]
manifest=scratch/'manifest.txt';manifest.write_text('\n'.join(rows)+'\n')
command([str(scratch/'oracle'),'--manifest',str(manifest)],'Go')
want=(scratch/'Go.log').read_bytes()
def compare(module,label,mutant=False):
 prefix=scratch/(label+'-binary')
 command(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(virtual),str(module/'validation/driver.a'),str(prefix)],label+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(module/'validation/driver.a'),str(manifest)]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(prefix)+'.mjs',str(manifest)]),('native',[str(prefix),str(manifest)])]:
  command(args,label+'-'+side)
  got=(scratch/(label+'-'+side+'.log')).read_bytes()
  if mutant and got==want:raise RuntimeError(label+' survived on '+side)
  if not mutant and got!=want:
   a=got.decode().splitlines();b=want.decode().splitlines()
   for i,(left,right) in enumerate(zip(a,b)):
    if left!=right:raise RuntimeError(side+' line '+str(i+1)+' got '+repr(left)+' want '+repr(right)+' context '+repr(b[max(0,i-4):i+4]))
   raise RuntimeError(side+' length mismatch')
  print(label,side,'mutant caught' if mutant else 'identical',len(want),'bytes',flush=True)
compare(roots[0],'baseline')
for i,p in enumerate(roots):
 variant=scratch/('mutant-'+str(i));variant.mkdir(exist_ok=True)
 for original in roots:shutil.copytree(original,variant/original.name,dirs_exist_ok=True)
 # Keep sibling rule imports copied; resolve shared module imports from the source root.
 for original in roots:
  for path in (variant/original.name).rglob('*'):
   if path.suffix not in ['.a','.ts']:continue
   relative=path.relative_to(variant/original.name);origin=original/relative
   def rewrite(match):
    imported=match.group(2)
    if not imported.startswith('.'):return match.group(0)
    target=(origin.parent/imported).resolve()
    if any(target.is_relative_to(root) for root in roots):return match.group(0)
    return match.group(1)+json.dumps(str(target))+match.group(3)
   path.write_text(re.sub(r"(from\s+)[\"']([^\"']+)[\"'](;)",rewrite,path.read_text()))
 spec=json.loads((p/'mutant.json').read_text());target=variant/p.name/spec['file'];text=target.read_text();assert text.count(spec['from'])==1;target.write_text(text.replace(spec['from'],spec['to']))
 compare(variant/roots[0].name,'mutant-'+str(i),True)
# Throughput: compiler plus a nonzero 1000-statement stress file for each rule.
command(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(virtual),str(roots[0]/'validation/driver.a'),str(scratch/'release'),'release'],'release-build')
for i,name in enumerate(names):
 source="const button%dClassName = ' flex  gap-2 ';\n"
 stress=scratch/('stress-'+str(i)+'.ts');stress.write_text(''.join(source % j for j in range(1000)))
 performance=scratch/('perf-'+str(i)+'.txt');performance.write_text('\n'.join([str(p)+'\t'+name+'\tnull' for p in sorted((corpus/'src/compiler').rglob('*.ts'))]+[str(stress)+'\t'+name+'\tnull'])+'\n')
 best={};count=None
 for round in range(3):
  for side,args in [('Go',[str(scratch/'oracle'),'--manifest',str(performance),'--count']),('native',[str(scratch/'release'),str(performance),'--count']),('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),str(roots[0]/'validation/driver.a'),str(performance),'--count'])]:
   label='perf-'+str(i)+'-'+str(round)+'-'+side;elapsed=command(args,label);output=(scratch/(label+'.log')).read_bytes()
   if count is None:count=output
   assert output==count
   best[side]=min(best.get(side,float('inf')),elapsed)
 for side,elapsed in best.items():print(name,side,'findings',int(count),'seconds',round_value:=format(elapsed,'.6f'),'findings/s',format(int(count)/elapsed,'.2f'),flush=True)
