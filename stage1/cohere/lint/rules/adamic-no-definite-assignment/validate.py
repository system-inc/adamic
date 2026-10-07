#!/usr/bin/env python3
"""Owned projected-AST parity runner. Production shared files remain untouched."""
import json, os, subprocess, shutil, argparse
from pathlib import Path
HERE=Path(__file__).resolve().parent
ROOT=HERE.parents[4]
parser=argparse.ArgumentParser();parser.add_argument('--scratch',type=Path,required=True);parser.add_argument('--prepare-only',action='store_true');parser.add_argument('--reuse',action='store_true');args=parser.parse_args()
S=args.scratch.resolve();S.mkdir(parents=True,exist_ok=True)
def run(label,cmd,cwd=ROOT,env=None):
 with (S/(label+'.log')).open('wb') as log:
  result=subprocess.run(cmd,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT)
 print(label,'exit',result.returncode,flush=True)
 if result.returncode: raise SystemExit(result.returncode)
 return S/(label+'.log')
virtual=ROOT/'cohere/adamic_wave11_oracle.go'
overlay=S/'oracle-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'validation_oracle.go.txt')}}))
run('oracle-build',['go','build','-overlay='+str(overlay),'-o',str(S/'oracle'),str(virtual)],ROOT/'cohere')
# Capture original Go fixture sources and decoded options, preserving filenames.
harness=ROOT/'cohere/internal/lint/testing/rule_testing.go'
src=harness.read_text();anchor='return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}'
assert src.count(anchor)==1
patched=S/'capture-harness.go';patched.write_text(src.replace(anchor,'result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result'))
overlay=S/'capture-overlay.json';overlay.write_text(json.dumps({'Replace':{str(harness):str(patched)}}))
env=os.environ|{'COHERE_DOCS_CAPTURE':str(S/'capture')}
for pkg,pattern in [('adamic','^TestNoDefiniteAssignment')]:
 if not args.reuse: run('capture-'+pkg,['go','test','-overlay='+str(overlay),'./internal/lint/rules/'+pkg,'-run',pattern,'-count=1','-v','-timeout=10m'],ROOT/'cohere',env)
names={'adamic/no-definite-assignment'}
rows={}
for file in sorted((S/'capture').glob('*.jsonl')):
 for line in file.read_text().split('\n'):
  if not line:continue
  row=json.loads(line)
  if row['rule'] not in names:continue
  options=json.dumps(row.get('options'),separators=(',',':')) if row.get('options') is not None else ''
  value={'name':row['file'],'source':row['source'],'rule':row['rule'],'options':options}
  rows[json.dumps(value,sort_keys=True)]=value
assert {r['rule'] for r in rows.values()}==names
corpus=S/'upstream.json';corpus.write_text(json.dumps(list(rows.values()),ensure_ascii=True))
print('captured',len(rows),{n:sum(r['rule']==n for r in rows.values()) for n in names},flush=True)
run('ast',[str(S/'oracle'),str(corpus),'--ast'])
run('answer',[str(S/'oracle'),str(corpus)])
run('types',['go','run','./cmd/adamic','types',str(HERE/'validation.a')])
run('build',['go','run',str(HERE/'validation_build.go'),str(HERE/'validation.a'),str(S/'native'),str(S/'emitted.mjs')])
if args.prepare_only:raise SystemExit(0)
runner=ROOT/'oracle/node.mjs'
for name,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',str(runner),str(HERE/'validation.a'),str(S/'ast.log')]),('emitted',['node','--disable-warning=ExperimentalWarning',str(runner),str(S/'emitted.mjs'),str(S/'ast.log')]),('native',[str(S/'native'),str(S/'ast.log')])]:
 path=run(name,cmd)
 assert path.read_bytes()==(S/'answer.log').read_bytes(),name+' differs from Go'
 print(name,'identical bytes',path.stat().st_size,flush=True)
