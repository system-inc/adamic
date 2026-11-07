import json,os,signal,subprocess,time,shutil,hashlib,difflib
from pathlib import Path
s=Path('/workspace/scratch/core-compare');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-compare';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source',PYTHONDONTWRITEBYTECODE='1');phases={}
assert json.loads((s/'POST-MEASURE.json').read_text())['paired-replay']['exit']==0
scripts=['instrument.cjs','count-preload.cjs','count.py','reconcile.py','control.py','contract.cjs','witness.cjs','compare.cjs','compare-census.py','census.py','replay.py','construct.cjs','post-measure.py','close-proof.py','write-docs.py','check-types.py','prepare.sh','build-tools.sh']
for name in scripts:shutil.copyfile(s/name,out/name)
def run(name,args,cwd=repo):
 print('start '+name,flush=True);begin=time.monotonic()
 with (s/(name+'.log')).open('w') as log:
  p=subprocess.Popen(args,cwd=cwd,env=env,stdout=log,stderr=log,start_new_session=True)
  try:code=p.wait(timeout=1700)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[name]={'exit':code,'seconds':time.monotonic()-begin};(s/'FINAL-CHECKS.json').write_text(json.dumps(phases,indent=2)+'\n');assert code==0,(name,code);print('pass '+name,flush=True)
run('docs-pending',['python3',str(s/'write-docs.py')])
run('apply-final-after',['bash',str(repo/'stage3/apply.sh'),'--build-product',str(s/'final-after')])
def manifest(root):return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(root.rglob('*')) if p.is_file()}
expected=manifest(s/'after/src');actual=manifest(s/'final-after/src');assert expected==actual and len(actual)==746
(s/'full-apply-source-equality.json').write_text(json.dumps({'sourceFiles':len(actual),'byteIdenticalToProvedAfter':True,'manifest':actual},indent=2)+'\n')
run('types-final-after',['python3',str(s/'check-types.py'),'final-after'])
run('census-final-after',['python3',str(s/'census.py'),'final-after'])
run('census-final-comparison',['python3',str(s/'compare-census.py'),'final-after'])
assert (s/'final-after/package-lock.json').read_bytes()==(repo/'stage3/typescript/package-lock.json').read_bytes()
(s/'final-after/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
run('build-final-after',['npm','run','build:compiler'],cwd=s/'final-after')
assert manifest(s/'final-after/src')==actual
names=['_tsc.js','_tsserver.js','_typingsInstaller.js','tsc.js','tsserver.js','tsserverlibrary.js','typescript.js','typingsInstaller.js','watchGuard.js','typescript.d.ts','tsserverlibrary.d.ts']
for name in names:assert (s/'final-after/built/local'/name).read_bytes()==(s/'after/built/local'/name).read_bytes(),name
(s/'final-comparison').mkdir();run('final-comparison',['node',str(s/'compare.cjs'),str(s/'before'),str(s/'final-after'),str(s/'final-comparison')])
run('witness-final',['node',str(s/'witness.cjs'),str(s/'final-after')])
comparison=json.loads((s/'final-comparison.log').read_text());witness=json.loads((s/'witness-final.log').read_text());assert set(json.loads((s/'COUNT.json').read_text())['counts'])<={r['typeofA']+'/'+r['typeofB'] for r in witness['rows'] if r['same']}
raw=s/'final-comparison'
for remove in ['false','true']:
 a=(raw/f'core-{remove}-before.js').read_text();b=(raw/f'core-{remove}-after.js').read_text();(raw/f'core-{remove}.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/core.js',tofile='after/core.js')))
blocks=['# Every emitted JavaScript hunk\n\nAll nine emitted files were compared. Each nonempty hunk is below; the AST comparison also proves equality outside the complete compareComparableValues owner. Both standalone core.ts comment modes are included.\n']
for p in sorted(raw.glob('*.diff')):
 if p.name=='source.diff' or not p.read_text():continue
 blocks.append('\n## '+p.name+'\n\n```diff\n'+p.read_text()+'```\n')
(out/'JAVASCRIPT-DIFF.md').write_text(''.join(blocks))
run('docs-final',['python3',str(s/'write-docs.py'),'final'])
for source,dest in [('COUNT.json','COUNT.json'),('RECONCILED.json','RECONCILED.json'),('CONTROL.json','CONTROL.json'),('CENSUS.json','CENSUS.json'),('FINAL-CHECKS.json','FINAL-CHECKS.json'),('BINARY-SCHEMA.json','BINARY-SCHEMA.json'),('recorder-provenance.json','recorder-provenance.json'),('final-comparison.log','COMPARISON.json'),('contract.log','CONTRACT.json'),('witness-final.log','WITNESSES.json')]:shutil.copyfile(s/source,out/dest)
for p in sorted(s.glob('piece-[0-9][0-9].tar.gz')):shutil.copyfile(p,out/p.name)
for name in ['excluded-first-attempt.tar.gz','json-piece-01-passed.tar.gz','json-piece-02-failed.tar.gz','packed-piece-02-failed.tar.gz']:shutil.copyfile(s/name,out/name)
print('FRESH FULL APPLY AND CLOSING PROOFS COMPLETE',flush=True)
