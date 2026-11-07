from pathlib import Path
import json,hashlib,tarfile,collections,math
s=Path('/workspace/scratch/gov-b');repo=Path('/workspace/adamic');u=repo/'stage3/adapt/32-indexed-reads-program';d=json.loads((u/'RECORD-TABLE-COUNT.json').read_text());assert d['complete']
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=[];totals=collections.Counter();oracle={side:collections.Counter() for side in ['before','after']};reports=set();valuesByUnit={};maxSeconds=0
canonical=lambda k:isinstance(k,str) and k.isascii() and k.isdigit() and str(int(k))==k and 0<=int(k)<4294967295
for piece in d['pieces']:
 n=piece['piece'];units.extend(piece['units']);assert not piece['failed'] and piece['seconds']<1500;maxSeconds=max(maxSeconds,piece['seconds']);archive=u/f'record-tables-measure-{n:02d}.tar.gz';assert hashlib.sha256(archive.read_bytes()).hexdigest()==(u/f'record-tables-measure-{n:02d}.sha256').read_text().split()[0]
 with tarfile.open(archive) as tf:
  prefix=f'piece-{n:02d}/';manifest=tf.extractfile(prefix+'SHA256SUMS').read().decode()
  for line in manifest.splitlines():
   h,name=line.split('  ',1);assert hashlib.sha256(tf.extractfile(prefix+name).read()).hexdigest()==h
  assert json.loads(tf.extractfile(prefix+'piece.json').read())==piece
  observations={}
  for line in tf.extractfile(prefix+'events.jsonl'):
   row=json.loads(line);key=(row['unit'],row['side']);assert key not in observations;observations[key]=row
   grouped=[]
   for event in row['rawRows']:
    if event['sequence']==1:grouped.append([])
    assert grouped and event['sequence']==len(grouped[-1])+1;grouped[-1].append(event)
    if event['event']=='write':
     kind,text=event['key'];number=float(text) if kind=='number' else float('nan');valid=kind=='number' and math.isfinite(number) and number.is_integer() and 0<=number<4294967295;assert event['canonical']==valid and valid
    else:
     assert event['event']=='values' and event['identityProjectionMatches'] and event['accessors']==0
     assert event['nonCanonicalKeys']==[k for k in event['keys'] if not canonical(k)]==[]
     assert len(event['keys'])==len(event['values'])
     if row['side']=='after':assert event['sourceKind']=='record'
   normal=[[{k:v for k,v in event.items() if k not in ['sourceKind','sparse','sequence']} for event in trace] for trace in grouped];normal.sort(key=lambda trace:json.dumps(trace,sort_keys=True));assert normal==row['traces']
  for unit in piece['units']:
   before=observations[(unit,'before')];after=observations[(unit,'after')];assert before['traces']==after['traces']
   returns=[x for x in before['rawRows'] if x['event']=='values'];stores=[x for x in before['rawRows'] if x['event']=='write'];totals['calls']+=len(returns);totals['beforeArrayCalls']+=sum(x['sourceKind']=='array' for x in returns);totals['afterArrayCalls']+=0;totals['baselineSparseCalls']+=sum(x['sparse'] for x in returns);totals['returnedElements']+=sum(len(x['values']) for x in returns);totals['idWrites']+=len(stores);valuesByUnit[unit]={'calls':len(returns),'sparse':sum(x['sparse'] for x in returns),'equal':True}
  for line in tf.extractfile(prefix+'shard-evidence.jsonl'):
   row=json.loads(line)
   if row['path']=='report.json':
    key=(row['unit'],row['side']);assert key not in reports;reports.add(key);report=json.loads(row['text']);assert report['status']=='pass' and not report['errors'];oracle[row['side']].update(report['counts'])
assert units==d['completedUnits']==[x['name'] for x in plan['units']];assert len(set(units))==398 and len(reports)==796
for key,value in totals.items():assert value==d[key]
assert totals['calls']==1109 and totals['beforeArrayCalls']==692 and totals['baselineSparseCalls']==445 and totals['afterArrayCalls']==0
assert all(dict(v)==plan['counts']==d['oracleCounts'][side] for side,v in oracle.items())
for key in ['nonCanonicalKeys','nonIntegerIds','valueDifferences','identityProjectionFailures']:assert d[key]==0
result={'wholeOraclePairPassed':True,'exactlyOnceUnitsPerSide':398,'callsPerSide':1109,'beforeArrays':692,'afterArrays':0,'previouslySparse':445,'returnedElementsPerSide':totals['returnedElements'],'idWritesPerSide':totals['idWrites'],'allIdsCanonicalIntegers':True,'allReturnKeysCanonical':True,'allOrderedValuesAndWritesEqual':True,'allWithinRunRecordProjectionsReferenceIdentical':True,'oracleCounts':{k:dict(v) for k,v in oracle.items()},'maxPieceSeconds':maxSeconds,'units':valuesByUnit}
(s/'raw/measure-reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps({k:v for k,v in result.items() if k!='units'}),flush=True)
