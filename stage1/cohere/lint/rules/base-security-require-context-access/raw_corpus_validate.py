#!/usr/bin/env python3
"""Validate complete current source corpora with live requirements and independent parsing."""
import argparse,json,subprocess,hashlib
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--baseline',type=Path,required=True);p.add_argument('--scratch',type=Path,required=True);p.add_argument('--compiler',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();S.mkdir(parents=True,exist_ok=True);B=a.baseline.resolve()
base=json.loads((B/'upstream.json').read_text());options=next(row['options'] for row in base if 'AccountRequestContextKey' in row['options']);summary={'options':options,'scope':'configured rule; raw sources parsed independently; no Go AST projection','corpora':[]}
def run(label,cmd):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
assert subprocess.check_output(['git','-C',str(a.compiler),'rev-parse','HEAD'],text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
for corpus,files in [('compiler',sorted((a.compiler/'src/compiler').rglob('*.ts'))),('stage1',sorted(f for f in (ROOT/'stage1').rglob('*') if f.suffix in ('.ts','.a') and '.generated' not in f.parts))]:
 census={'name':corpus,'files':len(files),'sourceBytes':sum(f.stat().st_size for f in files),'inputs':[{'path':str(f.relative_to(a.compiler if corpus=='compiler' else ROOT)),'sha256':hashlib.sha256(f.read_bytes()).hexdigest()} for f in files],'batches':[]};print(corpus,'files',len(files),flush=True)
 for i in range(0,len(files),4):
  label=corpus+'-'+str(i//4);rows=[{'name':'/'+str(f.relative_to(a.compiler if corpus=='compiler' else ROOT)),'source':f.read_text(),'rule':'base/security-require-context-access','options':options} for f in files[i:i+4]];raw=S/(label+'.json');raw.write_text(json.dumps(rows));want=run(label+'-Go',[B/'oracle',raw]);observations=[]
  for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',HERE/'validation.ts',raw,'--parse']),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',B/'emitted.mjs',raw,'--parse']),('native',[B/'native',raw,'--parse'])]:
   got=run(label+'-'+name,cmd);assert got==want,label+' '+name;observations.append({'backend':name,'equal':True,'exit':0,'stderrBytes':0})
  census['batches'].append({'canonicalBytes':len(want),'sha256':hashlib.sha256(want).hexdigest(),'findings':sum(line.startswith(b'base/security-require-context-access\t') for line in want.splitlines()),'backends':observations})
  if i%40==0:print(corpus,'processed',min(i+4,len(files)),flush=True)
 summary['corpora'].append(census);(S/'summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print('configured raw-source parity complete',flush=True)
