from pathlib import Path
import json,hashlib,tarfile,collections
s=Path('/workspace/scratch/gov-a');u=Path('/workspace/adamic/stage3/adapt/30-indexed-reads');repo=Path('/workspace/adamic')
d=json.loads((u/'GETOWNVALUES-COUNT.json').read_text());plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());assert d['complete'];assert len(d['pieces'])==20
seen=set();units=[];kinds=collections.Counter();lengths=collections.Counter();holes=collections.Counter();arraystats=collections.Counter();oracle=collections.Counter();count=0;reports={};maximum=0
for piece in d['pieces']:
 assert not piece['failed'] and piece['seconds']<1500;maximum=max(maximum,piece['seconds']);units.extend(piece['units']);n=piece['piece'];p=u/f'getownvalues-count-{n:02d}.tar.gz';assert hashlib.sha256(p.read_bytes()).hexdigest()==(u/f'getownvalues-count-{n:02d}.sha256').read_text().split()[0]
 with tarfile.open(p) as tf:
  prefix=f'piece-{n:02d}/';manifest=tf.extractfile(prefix+'SHA256SUMS').read().decode()
  for line in manifest.splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tf.extractfile(prefix+name).read()).hexdigest()==digest
  recorded=json.loads(tf.extractfile(prefix+'piece.json').read());assert recorded==piece
  perKinds=collections.Counter();perLengths=collections.Counter();perHoles=collections.Counter();perStats=collections.Counter();perCount=0
  for line in tf.extractfile(prefix+'calls.jsonl'):
   row=json.loads(line);assert row['event']=='call' and row['shard'] in piece['units'];identity=(row['shard'],row['pid'],row['sequence']);assert identity not in seen;seen.add(identity);count+=1;perCount+=1;kinds[row['kind']]+=1;perKinds[row['kind']]+=1
   if row['kind']=='array':
    assert 0<=row['present']<=row['length'] and row['holes']==row['length']-row['present'] and row['sparse']==(row['holes']>0)
    assert 0<=row['presentUndefined']<=row['present'] and 0<=row['accessorIndices']<=row['present']
    for aggregate in [lengths,perLengths]:aggregate[str(row['length'])]+=1
    for aggregate in [holes,perHoles]:aggregate[str(row['holes'])]+=1
    values={'sparse':row['holes']>0,'withNonIndexProperties':bool(row['nonIndexProperties']),'withPresentUndefined':row['presentUndefined']>0,'presentUndefinedElements':row['presentUndefined'],'accessorIndices':row['accessorIndices']}
    arraystats.update(values);perStats.update(values)
   elif row['kind']=='plain-record':assert row['prototype'] in ['Object','null']
  assert perCount==piece['calls'] and dict(perKinds)==piece['kinds']
  for key,value in perStats.items():assert value==piece['arrays'][key]
  assert dict(perLengths)==piece['arrays']['lengths'] and dict(perHoles)==piece['arrays']['holes']
  for line in tf.extractfile(prefix+'shard-evidence.jsonl'):
   row=json.loads(line)
   if row['path']=='report.json':
    report=json.loads(row['text']);assert report['status']=='pass' and not report['errors'];assert row['unit'] not in reports;reports[row['unit']]=report;oracle.update(report['counts'])
assert units==d['completedUnits']==[x['name'] for x in plan['units']];assert len(units)==len(set(units))==398
assert count==d['calls'] and {k:kinds[k] for k in ['array','plain-record','other']}==d['kinds']
assert dict(lengths)==d['arrays']['lengths'] and dict(holes)==d['arrays']['holes']
for key,value in arraystats.items():assert value==d['arrays'][key]
assert dict(oracle)==d['oracleCounts']==plan['counts']
for unit in plan['units']:assert reports[unit['name']]['counts']==unit['expect']
node=json.loads((s/'raw/node-witnesses.json').read_text());assert set(node['cases'])=={'sparse','presentUndefined','empty','record'}
nodeAgreement=all(v['original']==v['adapted'] and v['original']['exit']==0 and v['original']['stderr']=='' for v in node['cases'].values());assert nodeAgreement
backend=json.loads((s/'raw/backend-witnesses.json').read_text());native=backend['sparseWrite']['native'];nativeSparseMatch=native.get('runExit')==0 and native.get('stdout')==node['cases']['sparse']['adapted']['stdout'] and native.get('stderr')==''
holeRejected=native['compileExit']==0 and native['runExit']==70 and native['stdout']=='before-write-length=0\n' and native['stderr']=='adamic: panic: index 5 is outside an array of length 0\n';assert holeRejected and not nativeSparseMatch
for b in ['native','javascript']:
 for name in ['arrayArm','recordArm','arrayPresent','arrayEmpty']:
  entry=backend[name][b];assert entry['compileExit']==entry['runExit']==0 and entry['stdoutMatchesNode'] and entry['stderr']==''
 for name in ['sparse-adapted','presentUndefined-adapted','empty-adapted','record-adapted','recordBranch']:
  entry=backend[name][b];assert entry['compileExit']==1 and 'narrowed union member whose object tag cannot be checked' in entry['diagnostic']
result={'independentReconciliationPassed':True,'exactlyOneCompleteUnion':True,'units':len(units),'uniqueCalls':count,'kinds':{k:kinds[k] for k in ['array','plain-record','other']},'arrayStatistics':dict(arraystats),'lengths':dict(sorted(lengths.items(),key=lambda x:int(x[0]))),'holes':dict(sorted(holes.items(),key=lambda x:int(x[0]))),'oracleCounts':dict(oracle),'maxPieceSeconds':maximum,'pieces':len(d['pieces']),'nodeAgreementAllFour':nodeAgreement,'nativeSparseMatchesNode':nativeSparseMatch,'nativeRejectsHoleAtWrite':holeRejected,'stepBGate':{'allArgumentsArraysOrPlainRecords':kinds['other']==0,'noArrayOwnNonIndexProperty':arraystats['withNonIndexProperties']==0,'nativeSparseMatchOrNoSparse':nativeSparseMatch or (holeRejected and arraystats['sparse']==0)},'stepB':'not built'}
(s/'raw/reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)
