#!/usr/bin/env python3
"""Retain compact, reproducible crypto-import evidence, following 92 and 93."""
from pathlib import Path
import gzip,json,hashlib,subprocess as s,os,re,ast
P=Path(__file__).resolve().parent;R=P.parents[3];D=Path(os.environ.get('CRYPTO_PROOF_SCRATCH','/workspace/scratch/wasm-crypto-proof'));logs=P/'logs'
base=json.loads((D/'base.json').read_text())['base'];name='99-wasm-crypto-import'
def logtext(name):
 path=logs/name
 if path.exists():return path.read_text()
 with gzip.open(Path(str(path)+'.gz'),'rt') as stream:return stream.read()
def h(b):return hashlib.sha256(b).hexdigest()
def write(name,value):(P/name).write_text(json.dumps(value,indent=2)+'\n')
def compressed(source,destination):
 with destination.open('wb') as raw:
  with gzip.GzipFile(filename='',mode='wb',fileobj=raw,mtime=0) as out:out.write(source.read_bytes())
comparison=json.loads((P/'oracle-comparison.json').read_text());assert comparison['status']=='pass'
checks=json.loads((P/'checks.json').read_text());bindings=json.loads((P/'bindings.json').read_text());witness=json.loads((P/'built-witness.json').read_text());preparation=json.loads((P/'preparation.json').read_text());assert preparation['status']=='pass'
assert preparation['apiSha256']==comparison['api']['sha256'] and preparation['apiBytes']==comparison['api']['bytes']
refs={x:s.check_output(['git','-C',str(D/'after'),'rev-parse',f'refs/stage3/{name}-{x}'],text=True).strip() for x in ['before','after']}
numstat=s.check_output(['git','-C',str(D/'after'),'diff','--numstat',refs['before'],refs['after']],text=True);assert numstat=='4\t11\tsrc/compiler/sys.ts\n',numstat
for phase in ['before','after']:
 b=s.check_output(['git','-C',str(D/'after'),'show',refs[phase]+':src/compiler/sys.ts'])
 scoped=json.loads(s.check_output(['node',str(P/'scoped-hashes.cjs'),'-'],input=b))
 pins=json.loads((P/'targets.json').read_text())[0]
 assert scoped['ownerSha256']==pins['owner'+phase.title()+'Sha256'] and scoped['importsSha256']==pins['imports'+phase.title()+'Sha256']
assert (D/'control/src/compiler/sys.ts').read_bytes()==s.check_output(['git','-C',str(D/'after'),'show',refs['before']+':src/compiler/sys.ts'])
assert (D/'after/src/compiler/sys.ts').read_bytes()==s.check_output(['git','-C',str(D/'after'),'show',refs['after']+':src/compiler/sys.ts'])
protected=['stage3/patch-set.md','stage3/lane','stage3/oracle','stage3/fixtures','stage3/meter']
assert not s.check_output(['git','diff',base,'--',*protected],cwd=R),'shared harness, fixtures or reports changed'
order=sorted(p.name for p in (R/'stage3/adapt').iterdir() if p.is_dir());assert order[-1]==name and order[-2]=='99-correctness-axolotl-i'
source=(R/'stage3/apply.py').read_text();parsed=ast.parse(source);adapt=next(n for n in parsed.body if isinstance(n,ast.FunctionDef) and n.name=='adapt')
loops=[n for n in ast.walk(adapt) if isinstance(n,ast.For) and isinstance(n.target,ast.Name) and n.target.id=='directory'];assert len(loops)==1
loop=loops[0];runs=[n for n in ast.walk(loop) if isinstance(n,ast.Call) and isinstance(n.func,ast.Name) and n.func.id=='run' and n.args and isinstance(n.args[0],ast.Constant) and n.args[0].value=='node'];assert len(runs)==1
sort=next(i for i,l in enumerate(source.splitlines(),1) if 'adaptations = sorted(' in l)
write('order.json',dict(order=order,last=True,sortLine=sort,loopLine=loop.lineno,adapterInvocationLine=runs[0].lineno,nonSetupSecondPass=False,setupGenerationBeforeSnapshot=True,earlierReader='stage3/adapt/98-axolotl-e/adapt.cjs:141',rerunOutput=logtext('earlier-reader.log')))
rows={label:[l for l in (D/label/'patch-set.md').read_text().splitlines() if l.startswith('| '+name+' |') or l.startswith('| **Total**')] for label in ['control','after']};write('table.json',rows)
control_rows=[l for l in (D/'control/patch-set.md').read_text().splitlines() if not l.startswith('| **Total**')]
after_rows=[l for l in (D/'after/patch-set.md').read_text().splitlines() if not l.startswith('| **Total**') and not l.startswith('| '+name+' |')]
assert control_rows==after_rows,'an earlier official table row changed'
write('provenance.json',dict(task='#redxes2 head 1',base=base,controlMain=base,controlReused=False,upstream='050880ce59e30b356b686bd3144efe24f875ebc8',adaptation=name,applyExit=dict(control=0,after=0),applySnapshots=refs,applyNumstat=numstat,sourceFiles=checks['scope']['files'],sourceScope=checks['scope']['changed'],oracleWorkers=8,sharedHarnessUnchanged=True,protectedPaths=protected,applyPySha256=h((R/'stage3/apply.py').read_bytes()),laneCheckSha256=h((R/'stage3/lane/check.py').read_bytes())))
write('results.json',dict(status='pass',preparation=preparation,scope={k:v for k,v in checks['scope'].items() if k!='inventory'},bindings=bindings,table=rows,oracles=[dict(label=r['label'],counts=r['counts'],seconds=r['seconds'],exit=r['exit'],lane=r['lane']['status']) for r in comparison['runs']],baselineDiff=comparison['baselineDiff'],api=comparison['api'],witnessOutput=witness['output'],tsbuildinfo=witness['tsbuildinfo'],moduleLoadWitness=True,memoryEventsUnchanged=all(r['memoryBefore']==r['memoryAfter'] for r in comparison['runs']),mutants=['module collision','enclosing parameter shadow','bare namespace value','moved target','extra erased type','extra runtime edit','leave crypto ternary','extra baseline line']))
# Exact red lines, with complete output compressed alongside them.
def redlines(name,contains):
 text=logtext(name);matches=[l for l in text.splitlines() if any(x in l for x in contains)];assert matches,(name,contains);return '\n'.join(matches)
def md(name,title,command,body):
 edits={
  'bindings.mutant.md':'Three separate virtual programs: append const _crypto = {}; add an _crypto parameter to createSHA256Hash; append void _crypto;. The scoped hash gate is bypassed so these exercise the semantic binding guard itself.',
  'scope.mutant.md':'Add | null to clearTimeout\'s private ambient parameter type. The command first passes both transpile comparisons, then fails exact source undo.',
  'runtime-scope.mutant.md':'Replace return process.stdout.columns ?? 0; with return process.stdout.columns ?? 1; before the check, as prepare.py does.',
  'moved-target.mutant.md':'Change darwin to darw1n inside getNodeSystem before invoking the adapter. prepare.py checks that both mutant bytes and nanosecond mtime stay unchanged until restoration.',
  'ternary.mutant.md':'Replace createHash: createSHA256Hash, with createHash: _crypto ? createSHA256Hash : generateDjb2Hash,. The command plants it on disk and calls the semantic guard directly, bypassing the hash gate.',
  'oracle.mutant.md':'Append an extra baseline mutant line to the completed oracle-after/baseline.diff, leaving both API snapshots unchanged. Both api_check and the real lane CLI must reject it.',
 }
 (P/name).write_text(title+'\n\nTask #redxes2, head 1.\n\nMutation: '+edits[name]+'\n\nCommand:\n\n```sh\n'+command+'\n```\n\nExact observed output:\n\n```text\n'+body+'\n```\n\nEvery on-disk mutant is restored in `finally`; virtual binding mutants write no files. Complete raw output is retained under `logs/`.\n')
prefix='stage3/adapt/'+name+'/proofs/';pair=str(D/'control')+' '+str(D/'after')
md('bindings.mutant.md','Three independently bound negative programs','node '+prefix+'bindings.cjs '+pair+' '+str(P/'bindings.json'),redlines('bindings.log',['CAUGHT ']))
md('scope.mutant.md','An erased type edit is caught only by the exact source proof','node '+prefix+'scope-mutant.cjs '+pair,redlines('scope-only-mutant.log',['PASS TestCryptoTranspileScope:','AssertionError [ERR_ASSERTION]: TestCryptoSourceScope:']))
md('runtime-scope.mutant.md','An unrelated terminal-width runtime edit is outside the adaptation','node '+prefix+'check.cjs '+pair,redlines('runtime-scope-mutant.log',['AssertionError [ERR_ASSERTION]: TestCryptoSourceScope:']))
md('moved-target.mutant.md','A changed owner byte is rejected before any write','node '+prefix+'../adapt.cjs '+str(D/'after'),redlines('moved-target-mutant.log',['AssertionError [ERR_ASSERTION]: uninventoried owner: getNodeSystem']))
md('ternary.mutant.md','Leaving the crypto conditional is a bare namespace value','node '+prefix+'ternary-mutant.cjs '+str(D/'after'),redlines('ternary-mutant.log',['AssertionError [ERR_ASSERTION]: collision or bare namespace use:']))
md('oracle.mutant.md','An extra baseline line is caught by the real lane byte checker','CRYPTO_PROOF_SCRATCH='+str(D)+' python3 '+prefix+'oracles.py',redlines('oracle-baseline-mutant.log',['baseline.diff does not exactly describe','FAIL stage3']))
(P/'earlier-reader.md').write_text('98-e is an earlier reader, not a later pass\n\nThe official apply runs 98-e before '+name+'. Its store guard at `stage3/adapt/98-axolotl-e/adapt.cjs:141` reads the conditional property that this adaptation removes. Running it alone on the final output is intentionally rejected. This is its input moving, not a defect in its guard.\n\n```sh\nnode stage3/adapt/98-axolotl-e/adapt.cjs '+str(D/'after')+'\n```\n\nExit 1. Exact red:\n\n```text\n'+redlines('earlier-reader.log',['the nodeSystem store: want one, got 0'])+'\n```\n\n`apply.py` sorts at line '+str(sort)+', then executes its sole non-setup adapter loop at lines '+str(loop.lineno)+'-'+str(runs[0].lineno)+'. There is no post-adaptation pass. The separate 00-setup generation occurs before the pristine snapshot. `order.json` records the complete order. No adapter sorts after this one, and no shared runner was changed. The failed standalone rerun writes no source bytes; the restored whole-source check passes immediately afterwards.\n')
for label in ['control','after']:
 for file in ['install.log','build.log','tests.log','report.json','failures.md','baseline.diff']:
  p=D/('oracle-'+label)/file
  if p.exists():compressed(p,logs/('oracle-'+label+'-'+file+'.gz'))
project_command=witness['cli'][0]['command']
project=Path(project_command[project_command.index('--project')+1]).parent
for label in ['control','after']:
 compressed(project/(label+'.tsbuildinfo'),logs/('incremental-'+label+'.tsbuildinfo.gz'))
for source in sorted(logs.glob('*.log')):compressed(source,Path(str(source)+'.gz'))
for source in sorted((D/'logs').glob('*.log')):
 if not source.name.startswith(('archive-driver','readme-driver')):compressed(source,logs/(source.name+'.gz'))
for source in sorted(logs.glob('*.log')):source.unlink()
# Human-readable reports and full raw logs remain separate; no accepted assertion is summarized away.
print('PASS archived last-adapter guards, scope, builds, serial oracles and incremental witness')
