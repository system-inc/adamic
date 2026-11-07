import os,sys,json,shutil,hashlib,subprocess,difflib
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
from child_runner import run,ChildExit
s=Path('/workspace/scratch/core-compare-amend');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-compare';old=Path('/workspace/scratch/core-compare');env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),PYTHONDONTWRITEBYTECODE='1',ADAMIC_TYPESCRIPT_SOURCE='/workspace/typescript-source')
def manifest(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(tree.rglob('*')) if p.is_file()}
def census(side):
 tree=s/side;dest=s/('census-'+side);dest.mkdir();before=manifest(tree/'src');(dest/'source-manifest.json').write_text(json.dumps(before,indent=2)+'\n');variables={k:v for k,v in env.items() if not k.startswith('LATENT_')};variables.update(LATENT_FULL='1',LATENT_ONLY_FILE=str(tree/'src/compiler/core.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(dest/'unit-times.jsonl'))
 result=run('census-'+side,[old/'latent',tree/'src/compiler',dest/'raw.jsonl'],env=variables);assert manifest(tree/'src')==before;return result

def main():
 # Canary is a validation subject: its deliberately nonzero own code must remain 7.
 run('exit-code-canary',['bash','-c','exec python3 /workspace/scratch/core-compare-amend/exit_canary.py'],env=env,accept=(7,));assert json.loads((s/'EXIT-CANARY.json').read_text())=={'childExitCode':7,'wrapperExitCode':7}
 run('contract',['node',s/'contract.cjs'],env=env)
 run('witness',['node',s/'witness.cjs'],env=env)
 for side in ['before','after']:
  roots=sorted((s/side/'src/compiler').rglob('*.ts'));assert len(roots)==79
  run('types-'+side,[old/'adamic','types',*roots],env=env)
  census(side)
  (s/side/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
  run('build-'+side,['npm','run','build:compiler'],cwd=s/side,env=env)
 raw=s/'comparison';raw.mkdir();run('comparison',['node',s/'compare.cjs',s/'before',s/'after',raw],env=env)
 # Restore only the complete original owner in a source-identical mutant.
 (s/'mutant').mkdir();shutil.copytree(s/'after/src',s/'mutant/src');(s/'mutant/node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 pin=json.loads((out/'owners.json').read_text())[0];p=s/'mutant/src/compiler/core.ts';text=p.read_bytes();nl=b'\r\n' if b'\r\n' in text else b'\n';before=pin['before'].encode().replace(b'\n',nl);after=pin['after'].encode().replace(b'\n',nl);assert text.count(after)==1;p.write_bytes(text.replace(after,before));assert manifest(s/'mutant/src')==manifest(s/'before/src')
 census('mutant');run('census-comparison',['python3',s/'compare-census.py'],env=env)
 # Paired replay records each Node child's own code; matched refusals are expected validation subjects.
 replay=s/'replay';replay.mkdir()
 for side in ['before','after']:
  shutil.copytree(s/side,s/('replay-'+side),symlinks=True,ignore=shutil.ignore_patterns('node_modules'));(s/('replay-'+side)/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
 rows=[]
 for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
  def invoke(side):
   tree=s/('replay-'+side);label='replay/'+directory.name+'-'+side;event=run(label,['node',directory/'adapt.cjs',tree],env=env,accept=None);return side,{'childExitCode':event['childExitCode'],'output':(s/(label+'.log')).read_text().replace(str(tree),'TREE')}
  with ThreadPoolExecutor(max_workers=2) as pool:row={'adapter':directory.name,**dict(pool.map(invoke,['before','after']))}
  rows.append(row);(replay/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
  if directory.name!='99-correctness-core-compare':assert row['before']==row['after'],row
  else:
   assert row['before']['childExitCode']==row['after']['childExitCode']==0
   assert json.loads(row['before']['output'])=={'files':1,'edits':1,'javascriptChanging':True}
   assert json.loads(row['after']['output'])=={'files':0,'edits':0,'javascriptChanging':True}
 assert manifest(s/'replay-before/src')==manifest(s/'replay-after/src')
 # Retain every emitted JavaScript hunk, including both standalone core comment modes.
 for mode in ['false','true']:
  a=(raw/f'core-{mode}-before.js').read_text();b=(raw/f'core-{mode}-after.js').read_text();(raw/f'core-{mode}.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/core.js',tofile='after/core.js')))
 report={'parentCommit':'3cf7ef3e29ab4ec52078d6f85e3545cbd0c6d45c','bothFullApplies':True,'replayAdapters':len(rows),'pairedReplayIdenticalApartFromOurs':True,'rootsEachSide':79,'diagnostics':0,'census':json.loads((s/'CENSUS.json').read_text()),'comparison':json.loads((s/'comparison.log').read_text()),'witnesses':json.loads((s/'witness.log').read_text()),'countReuse':json.loads((s/'COUNT-REUSE.json').read_text()),'childExitReporting':'Actual child returncodes are retained; launch wrappers propagate unexpected child failure codes; explicit validators record expected nonzero subjects.','exitCanary':json.loads((s/'EXIT-CANARY.json').read_text())}
 (s/'AMENDMENT.json').write_text(json.dumps(report,indent=2)+'\n');print('AMENDMENT PROOFS PASS',flush=True)
try:main()
except ChildExit as error:
 print(str(error),flush=True);sys.exit(error.code if error.code>=0 else 128-error.code)
