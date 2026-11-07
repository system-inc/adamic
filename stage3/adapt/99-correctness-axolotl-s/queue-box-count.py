import collections,hashlib,json,pathlib,sys
root=pathlib.Path(sys.argv[1]);counts=collections.Counter();processes=[]
for file in sorted((root/'count').glob('*.json')):
 row=json.loads(file.read_text());counts.update(row['counts']);processes.append({'file':file.name,'sha256':hashlib.sha256(file.read_bytes()).hexdigest(),'pid':row['pid']})
assert counts['counterexamples']==0,counts
operations={method:sum(n for key,n in counts.items() if key.startswith(method+':')) for method in ['init','enqueue','dequeue','isEmpty','postOperationIsEmpty']}
assert all(n>0 for n in operations.values()),operations
physical=(root/'embed.json').exists()
if physical:
 embed=json.loads((root/'embed.json').read_text());assert len(embed['records'])==3
 progress=json.loads((root/'lane-pieces/progress.json').read_text());assert progress['complete']
 assert len(progress['completedUnits'])==progress['units']
 assert {json.loads(p.read_text())['unit'] for p in (root/'count').glob('*.json')}==set(progress['completedUnits'])
else:assert any(k.startswith('instrumented:') for k in counts),counts
verdict=json.loads((root/'lane-count/verdict.json').read_text());assert verdict['status']=='pass',verdict
summary={'complete':True,'physicalOwners':physical,'oracleVerdict':verdict['status'],'operations':operations,'initialClasses':{key:counts['init:'+key] for key in ['undefined','empty','dense','sparse']},'ownIteratorInputs':counts['initOwnIterator:yes'],'dequeueValues':{key.removeprefix('dequeue:'):value for key,value in counts.items() if key.startswith('dequeue:')},'counterexamples':0,'seedCountsExcluded':True,'processes':processes,'rawCounts':dict(sorted(counts.items()))}
(root/'count-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
