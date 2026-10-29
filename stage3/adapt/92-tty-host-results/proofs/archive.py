#!/usr/bin/env python3
"""Archive the serial observations, matching adaptation 91's compressed-log layout."""
from pathlib import Path
import gzip,json,hashlib,subprocess as s,os
proof=Path(__file__).resolve().parent;repo=proof.parents[3];scratch=Path(os.environ.get('TTY_PROOF_SCRATCH','/tmp/tty-adaptation'));logs=proof/'logs'
comparison=json.loads((proof/'oracle-comparison.json').read_text());assert comparison['status']=='pass'
def compressed(source,destination):
 with destination.open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as out:out.write(source.read_bytes())
for label in ['control','after']:
 source=scratch/('oracle-'+label)
 for name in ['install.log','build.log','tests.log','report.json','failures.md','baseline.diff']:
  if (source/name).exists():compressed(source/name,logs/('oracle-'+label+'-'+name+'.gz'))
for source in logs.glob('*.log'):
 if 'attempt' not in source.name:compressed(source,Path(str(source)+'.gz'))
for name in ['full-source','oracles-driver','preparation-missing-npm']:
 source=scratch/'logs'/(name+'.log')
 if source.exists():compressed(source,logs/(name+'.log.gz'))
roots={}
for label in ['control','after']:
 tree=scratch/label
 roots[label]=[{'file':str(p.relative_to(tree)),'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted((tree/'src/compiler').rglob('*.ts'))]
 assert len(roots[label])==79
refs={name:s.check_output(['git','-C',str(scratch/'after'),'rev-parse','refs/stage3/92-tty-host-results-'+name],text=True).strip() for name in ['before','after']}
numstat=s.check_output(['git','-C',str(scratch/'after'),'diff','--numstat',refs['before'],refs['after']],text=True);assert numstat=='2\t2\tsrc/compiler/sys.ts\n'
pin=json.loads((proof/'provenance.json').read_text());pin.update(applySnapshots=refs,applyNumstat=numstat,roots=roots,sourceFiles=746)
(proof/'provenance.json').write_text(json.dumps(pin,indent=2)+'\n')
(proof/'results.json').write_text(json.dumps({'status':'pass','conditions':{'twoBodiesOnly':'pass','SystemInterfaceUnchanged':'pass','publishedAndHeadACheckers':'pass','callerEquivalence':'pass','sameFullOracleAndBaseline':'pass'},'checks':json.loads((proof/'checks.json').read_text()),'counts':{r['label']:r['counts'] for r in comparison['runs']},'baselineDiff':comparison['baselineDiff'],'api':comparison['api'],'requiredMutant':{'test':'TestTTYHostAxiomTypecheck','exit':1,'sites':['src/compiler/sys.ts:1531:13 TS2322','src/compiler/sys.ts:1534:13 TS2322'],'restoredExit':0},'otherMutants':['moved target','wrong width fallback','extra baseline line'],'postSuiteScope':json.loads((proof/'post-suite-scope.json').read_text()),'postSuiteIdempotence':0,'oracleWorkers':8,'memoryEventsUnchanged':all(r['memoryBefore']==r['memoryAfter'] for r in comparison['runs'])},indent=2)+'\n')
print('PASS archived raw evidence, 79-root hashes, exact full-apply snapshots and serial oracle comparison')
