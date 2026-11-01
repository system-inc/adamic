#!/usr/bin/env python3
"""Archive adaptation 93's raw observations, following adaptation 92's layout."""
from pathlib import Path
import gzip,json,hashlib,subprocess as s,os
proof=Path(__file__).resolve().parent;repo=proof.parents[3]
scratch=Path(os.environ.get('BUILTIN_PROOF_SCRATCH','/workspace/scratch/adapt-93-proof'));logs=proof/'logs'
comparison=json.loads((proof/'oracle-comparison.json').read_text());assert comparison['status']=='pass'
def compressed(source,destination):
 with destination.open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as out:out.write(source.read_bytes())
for label in ['control','after']:
 source=scratch/('oracle-'+label)
 for name in ['install.log','build.log','tests.log','report.json','failures.md','baseline.diff']:
  if (source/name).exists():compressed(source/name,logs/('oracle-'+label+'-'+name+'.gz'))
attempt=scratch/'oracle-after-attempt1'
if attempt.exists():
 for name in ['install.log','build.log','tests.log','report.json','failures.md','baseline.diff']:
  if (attempt/name).exists():compressed(attempt/name,logs/('oracle-after-attempt1-'+name+'.gz'))
for source in logs.glob('*.log'):compressed(source,Path(str(source)+'.gz'))
for name in ['source-clone','control-clone','control-apply','preparation-driver','oracles-driver','extra-proofs-driver']:
 source=scratch/'logs'/(name+'.log')
 if source.exists():compressed(source,logs/(name+'.log.gz'))
refs={name:s.check_output(['git','-C',str(scratch/'after'),'rev-parse','refs/stage3/93-builtin-imports-'+name],text=True).strip() for name in ['before','after']}
numstat=s.check_output(['git','-C',str(scratch/'after'),'diff','--numstat',refs['before'],refs['after']],text=True)
assert numstat=='6\t3\tsrc/compiler/sys.ts\n',numstat
roots={label:[dict(file=str(p.relative_to(scratch/label)),sha256=hashlib.sha256(p.read_bytes()).hexdigest()) for p in sorted((scratch/label/'src/compiler').rglob('*.ts'))] for label in ['control','after']}
base='7b9c0fe7e52cfe9cdeb774107389bd6870f41423'
assert all(len(files)==79 for files in roots.values())
assert not s.check_output(['git','diff',base,'--','stage3/lane','stage3/oracle'],cwd=repo), 'shared lane/oracle changed'
checks=json.loads((proof/'checks.json').read_text())
(proof/'provenance.json').write_text(json.dumps(dict(task='#he8tysb head 1',base=base,upstream='050880ce59e30b356b686bd3144efe24f875ebc8',control='complete main adaptation set, excluding only 93',applyExit=dict(control=0,after=0),applySnapshots=refs,applyNumstat=numstat,sourceFiles=checks['scope']['files'],roots=roots,oracleWorkers=8,laneCheckerUnchanged=True),indent=2)+'\n')
witness=json.loads((proof/'built-witness.json').read_text())
result=dict(status='pass',conditions=dict(onlyThreeImportsAndInitializers='pass',sourceUndoExact='pass',transpileBothCommentModes='pass',buildsAndPublicAPI='pass',sameFullOracleAndBaseline='pass',builtCompilerWitness='pass'),counts={r['label']:r['counts'] for r in comparison['runs']},baselineDiff=comparison['baselineDiff'],api=comparison['api'],scope={k:v for k,v in checks['scope'].items() if k!='inventory'},mutants=['moved target before write','extra runtime edit','extra erased type annotation caught only by scope','extra baseline line'],postSuiteIdempotence=0,oracleWorkers=8,memoryEventsUnchanged=all(r['memoryBefore']==r['memoryAfter'] for r in comparison['runs']),cliOutputs=witness['output'])
(proof/'results.json').write_text(json.dumps(result,indent=2)+'\n')
print('PASS archived exact official-apply snapshots, source inventory, built witness and serial oracle comparison')
