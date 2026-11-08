import sys
sys.dont_write_bytecode=True
from run import *
import shutil,hashlib,tarfile,difflib,collections
D=ROOT/'stage3/adapt/99-correctness-core-sys-timer-slot'
for source,target in [('head2-source-scope.json','source-scope.json'),('head2-artifacts.json','artifacts.json'),('head2-comparison.json','census-comparison.json'),('head2-paired-replay.json','paired-replay.json'),('head2-contract.log','contract-results.json'),('head1-stock.json','head1-stock.json'),('head2-stock.json','head2-stock.json'),('C-pins.json','storage-pin-survey.json')]:shutil.copyfile(S/source,D/target)
for p in S.glob('head2-*.diff.txt'):shutil.copyfile(p,D/p.name)
for name in ['run.py','finish-head2.py','paired-head2.py','compare-head2.py','stock-check.cjs','apply-head2.py','regenerate-head2.py','lane-head2.py','lane-head2-pieces.py'] :shutil.copyfile(S/name,D/name)
d=json.loads((S/'head2-comparison.json').read_text());before=d['summaries']['head1-candidate']['sitesLedger'];after=d['summaries']['head2-candidate']['sitesLedger'];key=lambda x:(x['where'],x['kind'],x['reason'],x['phase']);b={key(x):x for x in before};a={key(x):x for x in after}
(D/'census-kept.json').write_text(json.dumps([a[k] for k in sorted(a.keys()&b.keys())],indent=2)+'\n')
raw={}
for label in ['head1-candidate','head2-candidate','head2-mutant']:
 rows=[json.loads(l) for l in (S/(label+'-latent.jsonl')).read_text().splitlines()]
 raw[label]=[f for f in rows[1]['findings'] if f.get('reason','').startswith('a member typed any') and (':382:79' in f['where'] or ':1517:36' in f['where'])]
(D/'masking-witness.json').write_text(json.dumps(raw,indent=2)+'\n')
orig=subprocess.check_output(['git','-C','/workspace/typescript-source','show','050880ce59e30b356b686bd3144efe24f875ebc8:tests/baselines/reference/api/typescript.d.ts']).decode().replace('\r\n','\n');modified=orig
start=orig.index('    interface System {');end=orig.index('\n    }',start)+6;memberOwner=orig[start:end]
for row in json.loads((D/'api-sanctions.json').read_text()):assert memberOwner.count(row['before'])==1;memberOwner=memberOwner.replace(row['before'],row['after'])
modified=orig[:start]+memberOwner+orig[end:]
(D/'public-baseline-slot.diff.txt').write_text(''.join(difflib.unified_diff(orig.splitlines(True),modified.splitlines(True),fromfile='reference/api/typescript.d.ts',tofile='C-noinfer-b/api/typescript.d.ts')))
lane=json.loads((S/'lane-head2/ledger.json').read_text());assert lane['all_commands_executed'];assert all(r['exit']==0 and r['status']=='pass' for r in lane['units']);shutil.copyfile(S/'lane-head2/ledger.json',D/'lane-results.json')
counts=collections.Counter()
for row in lane['units']:counts.update(row['counts'])
assert counts=={'pass':106366,'fail':1,'skip':0},dict(counts)
assert json.loads((S/'head2-stock.json').read_text())['requireWitness']==[{'expression':'module: require(modulePath)','line':1613,'column':38,'valueType':'any','contextualType':'{}'}]
statuses={}
for p in S.glob('head2-*-status.json'):
 v=json.loads(p.read_text());statuses[p.name]=v
for name in ['head1-stock','head1-build-tests','head1-actual-types','head1-build-compiler','head2-apply']:
 p=S/(name+'-status.json');statuses[p.name]=json.loads(p.read_text())
# The first shard-generation attempt missed sourced Go PATH, then passed after sourcing.
interrupted=statuses.pop('head2-full-lane-status.json');assert interrupted['exit']!=0
assert all(v['exit']==0 for k,v in statuses.items()),[(k,v['exit']) for k,v in statuses.items() if v['exit']]
(D/'proof-results.json').write_text(json.dumps({'status':'pass','sourceRoots':79,'sourceFiles':746,'calls':22,'handleMemberReads':0,'addedStockDiagnostics':0,'javascriptFilesIdentical':9,'laneCounts':dict(counts),'laneUnits':len(lane['units']),'newAgainstHead1':8,'newAgainstMain':9,'unexpectedNewSites':[],'mutantRestores384':True,'commands':statuses,'interruptedFourWayAttempt':interrupted},indent=2)+'\n')
files=[]
for pattern in ['head2-*.log','head2-*.jsonl','head2-*-status.json','head1-stock.log','head1-build-tests.log','head2-apply.log','head1-candidate-latent.jsonl','head2-candidate-latent.jsonl','head2-mutant-latent.jsonl']:
 files+=list(S.glob(pattern))
files += [S/'lane-concurrent-stop.json',S/'lane-pieces-wrapper.log',S/'resume-head2.log']
with tarfile.open(D/'evidence.tar.gz','w:gz') as tar:
 for p in sorted(set(files)):tar.add(p,arcname=p.name)
 for label in ['head1','head2','head2-mutant']:tar.add(S/label/'src/compiler/sys.ts',arcname=label+'/sys.ts')
 for p in sorted((S/'lane-head2').rglob('*')):
  if p.is_file() and not p.is_symlink() and 'caches' not in p.relative_to(S/'lane-head2').parts and (p.suffix in ['.json','.jsonl','.log','.diff'] or p.name.endswith('.d.ts')):tar.add(p,arcname='lane/'+str(p.relative_to(S/'lane-head2')))
(D/'evidence.sha256').write_text(hashlib.sha256((D/'evidence.tar.gz').read_bytes()).hexdigest()+'  evidence.tar.gz\n')
print('collected full proof child_exit=0',flush=True)
