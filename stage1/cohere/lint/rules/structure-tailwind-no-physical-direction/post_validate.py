#!/usr/bin/env python3
"""Extend validate.py outputs with full corpora, isolated mutants and startup-inclusive timings."""
import argparse,json,os,subprocess,shutil,time,statistics,hashlib
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);p.add_argument('--compiler',type=Path,required=True);p.add_argument('--corpus-only',action='store_true');a=p.parse_args();S=a.scratch.resolve();runner=ROOT/'oracle/node.mjs'
def run(label,cmd,cwd=ROOT,allow=False):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:
  start=time.perf_counter();r=subprocess.run(list(map(str,cmd)),cwd=cwd,stdout=out,stderr=err);seconds=time.perf_counter()-start
 stderr=(S/(label+'.stderr')).read_bytes()
 if r.returncode or stderr:
  print(label,'FAILED',r.returncode,stderr[:500],flush=True)
  if not allow:raise SystemExit(1)
 return (S/(label+'.log')).read_bytes(),seconds,r.returncode,stderr
base=json.loads((S/'upstream.json').read_text());names=list(dict.fromkeys(r['rule'] for r in base))
def commands(ast,source=HERE/'validation.ts',native=S/'native',emitted=S/'emitted.mjs',independent=False):
 suffix=['--parse'] if independent else []
 return [('Node',['node','--disable-warning=ExperimentalWarning',runner,source,ast,*suffix]),('emitted',['node','--disable-warning=ExperimentalWarning',runner,emitted,ast,*suffix]),('native',[native,ast,*suffix])]
def compare(label,rows,independent=False,allow=False):
 raw=S/(label+'.json');raw.write_text(json.dumps(rows));ast=S/(label+'-ast.log');want,_,_,_=run(label+'-go',[S/'oracle',raw]);run(label+'-ast',[S/'oracle',raw,'--ast'])
 observations=[]
 for name,cmd in commands(ast,independent=independent):
  got,seconds,code,stderr=run(label+'-'+name,cmd,allow=allow)
  match=not code and not stderr and got==want
  observations.append({'backend':name,'equal':match,'exit':code,'stderrBytes':len(stderr),'seconds':seconds})
  if not match and not allow:raise SystemExit(label+' '+name+' differs from Go')
 return observations
summary={'upstreamCases':len(base),'compilerPin':subprocess.check_output(['git','-C',str(a.compiler),'rev-parse','HEAD'],text=True).strip(),'corpora':[]}
assert summary['compilerPin']=='050880ce59e30b356b686bd3144efe24f875ebc8'
# Each file is read verbatim. No parse-diagnostic or gap exclusions.
for corpus,files in [('compiler',sorted((a.compiler/'src/compiler').rglob('*.ts'))),('stage1',sorted(f for f in (ROOT/'stage1').rglob('*') if f.suffix in ('.ts','.a') and '.generated' not in f.parts))]:
 print(corpus,'files',len(files),flush=True)
 census={'name':corpus,'files':len(files),'bytes':sum(f.stat().st_size for f in files),'batches':[]}
 for batch in range(0,len(files),4):
  rows=[{'name':'/'+str(f.relative_to(a.compiler if corpus=='compiler' else ROOT)), 'source':f.read_text(),'rule':name,'options':''} for f in files[batch:batch+4] for name in names]
  observed=compare(corpus+'-'+str(batch//4),rows)
  census['batches'].append(observed)
  if batch%40==0:print(corpus,'processed',min(batch+4,len(files)),flush=True)
 summary['corpora'].append(census)
 (S/'summary.json').write_text(json.dumps(summary,indent=2))
# The independent stage-1 parser is tested separately, with failures retained, not reclassified as parity.
plain=[r for r in base if not r['name'].endswith('.tsx')]
summary['independentParser']=compare('independent-upstream',plain,independent=True,allow=True)
(S/'summary.json').write_text(json.dumps(summary,indent=2))
if a.corpus_only:raise SystemExit(0)
summary['mutants']=[]
for directory in [HERE,HERE.parent/'eslint-comments-require-description',HERE.parent/'next-google-font-display',HERE.parent/'adamic-no-definite-assignment']:
 change=json.loads((directory/'mutant.json').read_text());scratch=S/('mutant-'+directory.name);shutil.copytree(ROOT/'stage1',scratch/'stage1',dirs_exist_ok=True)
 owned=scratch/'stage1/cohere/lint/rules'/directory.name/change['file'];text=owned.read_text();assert text.count(change['from'])==1;owned.write_text(text.replace(change['from'],change['to'],1))
 driver=scratch/'stage1/cohere/lint/rules'/HERE.name/'validation.ts';binary=scratch/'native';js=scratch/'emitted.mjs'
 run(change['name']+'-build',['go','run',HERE/'validation_build.go',driver,binary,js])
 ast=S/'ast.log';wanted=(S/'answer.log').read_bytes();observed=[]
 for backend,cmd in commands(ast,driver,binary,js):
  got,_,code,err=run(change['name']+'-'+backend,cmd)
  assert code==0 and not err and got!=wanted, 'mutant was not killed solely by byte comparison'
  observed.append(backend)
 summary['mutants'].append({'name':change['name'],'caughtBy':observed,'onlyByteComparison':True})
 print('mutant',change['name'],'caught by all three comparisons',flush=True)
 (S/'summary.json').write_text(json.dumps(summary,indent=2))
summary['benchmarks']=[]
for name in names:
 positives=[]
 for i,row in enumerate(base):
  if row['rule']!=name:continue
  raw=S/'positive.json';raw.write_text(json.dumps([row]));out,_,_,_=run('positive',[S/'oracle',raw,'--count'])
  if int(out)>0:positives.append(row);break
 assert positives
 rows=positives*200;raw=S/'benchmark.json';raw.write_text(json.dumps(rows));run('benchmark-ast',[S/'oracle',raw,'--ast']);ast=S/'benchmark-ast.log'
 record={'rule':name,'scope':'startup-inclusive; Go parses source; Adamic loads projected AST; sanitized native','samples':[]}
 for backend,cmd in [('Go',[S/'oracle',raw,'--count']),('Node',['node','--disable-warning=ExperimentalWarning',runner,HERE/'validation.ts',ast,'--count']),('native',[S/'native',ast,'--count'])]:
  times=[];counts=[]
  for i in range(3):
   output,seconds,_,_=run('benchmark-'+backend+'-'+str(i),cmd);times.append(seconds);counts.append(int(output))
  assert len(set(counts))==1 and counts[0]>0
  record['samples'].append({'backend':backend,'findings':counts[0],'seconds':times,'findingsPerSecond':counts[0]/statistics.median(times)})
 assert len({r['findings'] for r in record['samples']})==1
 summary['benchmarks'].append(record);print('benchmark',record,flush=True)
 (S/'summary.json').write_text(json.dumps(summary,indent=2))
print('post validation complete',flush=True)
