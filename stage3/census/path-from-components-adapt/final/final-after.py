from pathlib import Path
import subprocess,sys,os,json,time,hashlib
repo=Path('/workspace/path-components-adapt');out=Path('/workspace/scratch/path-adapt-final');tree=Path('/workspace/scratch/path-adapt-final-tree');proof=Path('/workspace/scratch/path-adapt-proof')
# apply has its own bounded process; wait only for its completed status record.
start=time.monotonic()
while not (out/'apply-status.json').exists():
 if time.monotonic()-start>900:raise RuntimeError('apply status did not arrive within the bounded wait')
 time.sleep(1)
assert json.loads((out/'apply-status.json').read_text())['child_exit_code']==0
link=tree/'node_modules'
if not link.exists():link.symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
def run(label,args,cwd=repo):
 with (out/(label+'.log')).open('wb') as log:
  r=subprocess.run([sys.executable,str(out/'run.py'),str(out/label),*map(str,args)],cwd=cwd,stdout=log,stderr=subprocess.STDOUT)
 print(label+' supervisor_child_exit='+str(r.returncode),flush=True)
 if r.returncode:sys.exit(r.returncode)
run('replay',['node',out/'replay.cjs',tree,out/'replay.json'])
run('types-after-driver',[sys.executable,out/'check-types.py',tree,'types-after'])
run('stock-after',['node',proof/'stock-check.cjs',tree,out/'stock-after.json'])
run('build',['npm','run','build'],tree)
run('build-comparison',['node',proof/'compare-build.cjs','/workspace/scratch/path-adapt-before',tree,out/'build-comparison.json'])
run('census-after',['env','LATENT_FULL=1','LATENT_ONLY_FILE='+str(tree/'src/compiler/path.ts'),'LATENT_ASSERT_NO_OUTPUT=1',out/'latent',tree/'src/compiler',out/'census-after.jsonl'])
# Restore the already-proven original-owner mutant's exact baseline path source.
run('census-mutant',['env','LATENT_FULL=1','LATENT_ONLY_FILE=/workspace/scratch/path-adapt-mutant/src/compiler/path.ts','LATENT_ASSERT_NO_OUTPUT=1',out/'latent','/workspace/scratch/path-adapt-mutant/src/compiler',out/'census-mutant.jsonl'])
findings=[f for line in (out/'census-mutant.jsonl').read_text().splitlines() for f in json.loads(line).get('findings',[])]
assert any(f['kind']=='Refused' and f['where'].endswith('/path.ts:513:45') and f['reason']=="a cast the runtime can't check" for f in findings)
(out/'mutant-result.json').write_text(json.dumps({'restoredCast':'"" as T','refusalRestoredAt':'path.ts:513:45','reason':"a cast the runtime can't check"},indent=2)+'\n')
print('final after proofs passed',flush=True)
